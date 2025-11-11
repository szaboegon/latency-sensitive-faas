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

// calculateRequiredReplicas computes the number of replicas needed to handle the given arrival rate
// while keeping utilization below targetUtilization to avoid queueing delays.
// targetUtilization should be between 0 and 1 (e.g., 0.7 for 70% max utilization)
func calculateRequiredReplicas(runtime, targetConcurrency int, arrivalRate float64, targetUtilization float64) int {
	// Capacity per replica: requests per second that one replica can handle
	capacity := (1000.0 / float64(runtime)) * float64(targetConcurrency)
	if capacity < 1e-6 {
		capacity = 1e-6
	}

	// To keep utilization below targetUtilization, we need:
	// arrivalRate / (replicas * capacity) <= targetUtilization
	// Therefore: replicas >= arrivalRate / (capacity * targetUtilization)
	if targetUtilization <= 0 || targetUtilization > 1 {
		targetUtilization = 0.7 // default to 70% max utilization
	}

	requiredReplicas := arrivalRate / (capacity * targetUtilization)
	return max(int(math.Ceil(requiredReplicas)), 1)
}
