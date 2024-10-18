package metrics

import (
	"log"
	"sync"
	"time"
)

type MetricsWatcher interface {
	Watch()
	GetNodeMetrics() map[string]NodeMetrics
}

type metricsWatcher struct {
	nodeMetrics   map[string]NodeMetrics
	metricsReader MetricsReader
	mu            sync.RWMutex
}

func NewMetricsWatcher(apiKeyConfig ApiKeyConfig) (MetricsWatcher, error) {
	reader, err := NewMetricsReader(apiKeyConfig)
	nodeMetrics := make(map[string]NodeMetrics)

	if err != nil {
		return nil, err
	}

	return &metricsWatcher{
		metricsReader: reader,
		nodeMetrics:   nodeMetrics,
	}, nil
}

func (w *metricsWatcher) Watch() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			w.mu.Lock()
			metrics, err := w.metricsReader.GetNodeMetrics()
			if err != nil {
				log.Fatal("Failed to get metrics from metrics reader:", err)
			}
			w.nodeMetrics = metrics
			w.mu.Unlock()
		}
	}()
}

func (w *metricsWatcher) GetNodeMetrics() map[string]NodeMetrics {
	w.mu.RLock()
	defer w.mu.Unlock()
	return w.nodeMetrics
}
