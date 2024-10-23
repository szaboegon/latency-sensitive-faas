package routing

import (
	"faas-loadbalancer/internal/metrics"
	"faas-loadbalancer/internal/otel"
	"fmt"
	"net/http"
	"sort"
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

	var r Route
	var latestErr error
	for _, n := range bestNodes {
		i := sort.Search(len(routes), func(i int) bool {
			return routes[i].Node == n
		})
		if i < len(routes) {
			r = routes[i]
			err := sendRequest(r.Url, req)
			// if there was an error we try the next best option
			if err != nil {
				latestErr = err
				continue
			}
			latestErr = nil
			break
		}
	}

	if latestErr != nil {
		return Route{}, latestErr
	}

	return r, nil
}

func sendRequest(url string, req Request) error {
	client := &http.Client{}
	client = otel.WithOtelTransport(client)

	httpReq, err := http.NewRequest("POST", url, req.BodyReader)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set(ForwardToHeader, string(req.ToComponent))
	if err != nil {
		return err
	}
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
