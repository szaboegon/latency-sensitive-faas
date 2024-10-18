package metrics

import "faas-loadbalancer/internal/k8s"

type NodeMetrics struct {
	Node   k8s.Node
	Cpu    cpu    `json:"cpu"`
	Memory memory `json:"memory"`
}

type cpu struct {
	Utilization float64 `json:"utilization"`
}

type memory struct {
	Usage float64 `json:"usage"`
}

type MetricsReader interface {
	QueryNodeMetrics() ([]NodeMetrics, error)
	Test() (string, error)
}
