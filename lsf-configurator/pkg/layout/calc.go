package layout

import (
	"lsf-configurator/pkg/core"
	"math"
)

// arrival rate is sum of incoming rates
// computes total arrival rate for a group of components
func calculateTotalArrivalRate(comps []core.ComponentProfile, links []core.ScenarioLink) float64 {
	groupNames := make(map[string]bool)
	for _, c := range comps {
		groupNames[c.Name] = true
	}
	totalRate := 0.0
	for _, l := range links {
		if groupNames[l.To] && !groupNames[l.From] {
			totalRate += l.InvocationRate
		}
	}

	if totalRate <= 0 {
		totalRate = 1.0
	}
	return totalRate
}

func calculateRequiredReplicas(runtime, targetConcurrency int, arrivalRate float64) int {
	capacity := (1000.0 / float64(runtime)) * float64(targetConcurrency)
	if capacity < 1e-6 {
		capacity = 1e-6
	}

	return max(int(math.Ceil(arrivalRate/capacity)), 1)
}
