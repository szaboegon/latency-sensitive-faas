package metrics

import (
	"faas-loadbalancer/internal/k8s"
)

type resourceBasedEvaluator struct {
	node k8s.Node
	bias float64
}

func NewResourceBasedEvaluator(node k8s.Node, bias float64) NodeEvaluator {
	return &resourceBasedEvaluator{
		node: node,
		bias: bias,
	}
}

// the bigger the weight the better
func (e resourceBasedEvaluator) CalculateWeight(nm NodeMetrics) float64 {
	weight := -nm.Cpu.Utilization
	if nm.Node == e.node {
		return weight + e.bias
	}
	return weight
}
