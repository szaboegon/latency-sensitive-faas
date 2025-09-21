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
	nodesWithMostResources []k8s.Node
	metricsReader          MetricsReader
	nodeEvaluator          NodeEvaluator
	mu                     sync.RWMutex
}

func NewMetricsWatcher(evaluator NodeEvaluator, backendAddr string) (Watcher, error) {
	reader, err := NewMetricsReader(backendAddr)
	nodesWithMostResources := []k8s.Node{}

	if err != nil {
		return nil, err
	}

	return &metricsWatcher{
		metricsReader:          reader,
		nodeEvaluator:          evaluator,
		nodesWithMostResources: nodesWithMostResources,
		mu:                     sync.RWMutex{},
	}, nil
}

func (w *metricsWatcher) Watch() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			metrics, err := w.metricsReader.QueryNodeMetrics()
			if err != nil {
				log.Println("Failed to get metrics from metrics reader: ", err)
				continue
			}
			sort.Slice(metrics, func(i, j int) bool {
				return w.nodeEvaluator.CalculateWeight(metrics[i]) > w.nodeEvaluator.CalculateWeight(metrics[j])
			})
			// local function to make sure the mutex always unlocks
			func() {
				w.mu.Lock()
				defer w.mu.Unlock()
				w.nodesWithMostResources = []k8s.Node{}
				for _, val := range metrics {
					w.nodesWithMostResources = append(w.nodesWithMostResources, val.Node)
				}
			}()
		}
	}()
}

func (w *metricsWatcher) GetNodesWithMostResources() []k8s.Node {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.nodesWithMostResources
}
