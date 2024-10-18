package metrics

type NodeMetrics struct {
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
	GetNodeMetrics() (map[string]NodeMetrics, error)
	Test() (string, error)
}
