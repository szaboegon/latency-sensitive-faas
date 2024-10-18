package routing

import (
	"faas-loadbalancer/internal/k8s"
	"faas-loadbalancer/internal/metrics"
	"log"
)

type metricsBasedRouter struct {
	routingTable RoutingTable
	watcher      metrics.Watcher
}

func NewRouter(fl FunctionLayout, node k8s.Node) (Router, error) {
	watcher, err := metrics.NewMetricsWatcher(node)
	if err != nil {
		log.Fatal("Failed to create metrics watcher:", err)
		return nil, err
	}
	watcher.Watch()

	rt := make(RoutingTable)
	for _, partition := range fl.FuncPartitions {
		for _, component := range partition.Components {
			val, ok := rt[component]
			if !ok {
				rt[component] = []FuncPartition{
					partition,
				}
			} else {
				rt[component] = append(val, partition)
			}
		}
	}

	return &metricsBasedRouter{
		routingTable: rt,
		watcher:      watcher,
	}, nil
}

func (mr *metricsBasedRouter) RouteRequest(req Request) error {
	//partitions := mr.routingTable[req.ToComponent]

	return nil
}
