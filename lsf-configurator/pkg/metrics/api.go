package metrics

type Node string

type NodeMetrics struct {
	Node   Node
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
	QueryAverageAppRuntime(appId string) (float64, error)
}
