package metrics

import (
	"faas-loadbalancer/internal/k8s"
	"log"
	"sort"
	"sync"
	"time"
)

type Watcher interface {
	Watch()
	GetNodesWithMostResources() []k8s.Node
}

type metricsWatcher struct {
	node                   k8s.Node
	nodesWithMostResources []k8s.Node
	metricsReader          MetricsReader
	mu                     sync.RWMutex
}

func NewMetricsWatcher(node k8s.Node) (Watcher, error) {
	reader, err := NewMetricsReader()
	nodesWithMostResources := []k8s.Node{}

	if err != nil {
		return nil, err
	}

	return &metricsWatcher{
		node:                   node,
		metricsReader:          reader,
		nodesWithMostResources: nodesWithMostResources,
		mu:                     sync.RWMutex{},
	}, nil
}

func (w *metricsWatcher) Watch() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			w.mu.Lock()
			metrics, err := w.metricsReader.QueryNodeMetrics()
			if err != nil {
				log.Fatal("Failed to get metrics from metrics reader:", err)
			}
			sort.Slice(metrics, func(i, j int) bool {
				return w.calculateWeight(metrics[i]) < w.calculateWeight(metrics[j])
			})
			w.nodesWithMostResources = []k8s.Node{}
			for _, val := range metrics {
				w.nodesWithMostResources = append(w.nodesWithMostResources, val.Node)
			}
			w.mu.Unlock()
		}
	}()
}

func (w *metricsWatcher) GetNodesWithMostResources() []k8s.Node {
	w.mu.RLock()
	defer w.mu.Unlock()
	return w.nodesWithMostResources
}

// the smaller the weight the better
func (w *metricsWatcher) calculateWeight(nm NodeMetrics) float64 {
	if nm.Node == w.node {
		return nm.Cpu.Utilization * 0.8
	}
	return nm.Cpu.Utilization
}
