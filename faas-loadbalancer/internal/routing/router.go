package routing

import (
	"bytes"
	"encoding/json"
	"faas-loadbalancer/internal/metrics"
	"faas-loadbalancer/internal/otel"
	"fmt"
	"log"
	"net/http"
)

type metricsBasedRouter struct {
	routingTable RoutingTable
	watcher      metrics.Watcher
}

func NewRouter(fl FunctionLayout, w metrics.Watcher) (Router, error) {
	w.Watch()
	rt := make(RoutingTable)
	for _, partition := range fl.FuncPartitions {
		for _, component := range partition.Components {
			val, ok := rt[component]
			if !ok {
				rt[component] = []Route{
					{partition, getPartitionUrl(partition)},
				}
			} else {
				rt[component] = append(val, Route{partition, getPartitionUrl(partition)})
			}
		}
	}

	return &metricsBasedRouter{
		routingTable: rt,
		watcher:      w,
	}, nil
}

func (mr *metricsBasedRouter) RouteRequest(req Request) (Route, error) {
	bestNodes := mr.watcher.GetNodesWithMostResources()
	routes := mr.routingTable[req.ToComponent]

	var sent *Route = nil
	var latestErr error
	for _, node := range bestNodes {
		for _, route := range routes {
			if route.Node == node {
				err := sendRequest(route.Url, req)
				// if there was an error we try the next best option
				if err != nil {
					log.Printf("tried sending request to partition: %v, on node: %v", route.Name, route.Node)
					latestErr = err
					continue
				}
				sent = &route
				latestErr = nil
				break
			}
		}
	}

	if sent == nil {
		return Route{}, fmt.Errorf("no matching route found for component: %v", req.ToComponent)
	}

	if latestErr != nil {
		return Route{}, latestErr
	}

	return *sent, nil
}

func sendRequest(url string, req Request) error {
	client := otel.WithOtelTransport(&http.Client{})

	httpReq, err := http.NewRequestWithContext(req.Context, "POST", url, bytes.NewReader(req.Body))
	if err != nil {
		return err
	}
	otel.InjectHeaders(httpReq, req.Context)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(ForwardToHeader, string(req.ToComponent))

	headers, _ := json.Marshal(httpReq.Header)
	log.Printf("http req headers: %s", headers)
	r, err := client.Do(httpReq)
	if err != nil {
		return err
	}

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status code: %v", r.StatusCode)
	}

	return nil
}

func getPartitionUrl(p FuncPartition) string {
	return fmt.Sprintf("http://%v.%v.svc.cluster.local", p.Name, p.Namespace)
}
