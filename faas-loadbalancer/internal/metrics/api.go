package metrics

type NodeMetrics struct {
	CpuUtilizationPercentage float64
}

type MetricsReader interface {
	GetNodeCpuUtilizations() (map[string]NodeMetrics, error)
	Test() (string, error)
}
