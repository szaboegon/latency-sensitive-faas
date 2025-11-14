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

func calculateFanOutRatio(comps []core.ComponentProfile, links []core.ScenarioLink) float64 {
	groupNames := make(map[string]bool)
	for _, c := range comps {
		groupNames[c.Name] = true
	}

	// Find upstream components and their outgoing rates
	upstreamRates := make(map[string]float64)

	// First pass: find total RPS going INTO this group
	totalIncoming := 0.0
	upstreamComponents := make(map[string]bool)
	for _, l := range links {
		if groupNames[l.To] && !groupNames[l.From] {
			totalIncoming += l.InvocationRate
			upstreamComponents[l.From] = true
		}
	}

	// Second pass: find total RPS each upstream receives
	for upstreamName := range upstreamComponents {
		for _, l := range links {
			if l.To == upstreamName {
				upstreamRates[upstreamName] += l.InvocationRate
			}
		}
	}

	// Calculate max fan-out ratio
	maxFanOut := 1.0
	for upstreamName, upstreamIncoming := range upstreamRates {
		// Find this upstream's outgoing rate to our group
		outgoingToGroup := 0.0
		for _, l := range links {
			if l.From == upstreamName && groupNames[l.To] {
				outgoingToGroup += l.InvocationRate
			}
		}

		if upstreamIncoming > 0 {
			fanOut := outgoingToGroup / upstreamIncoming
			if fanOut > maxFanOut {
				maxFanOut = fanOut
			}
		}
	}

	return maxFanOut
}

// calculateRequiredReplicas computes the number of replicas needed to handle the given arrival rate
// while keeping utilization below targetUtilization to avoid queueing delays.
// targetUtilization should be between 0 and 1 (e.g., 0.7 for 70% max utilization)
func calculateRequiredReplicas(runtime, targetConcurrency int, arrivalRate, targetUtilization, fanOutRatio float64) int {
	// Capacity per replica: requests per second that one replica can handle
	burstAdjustedRate := arrivalRate * fanOutRatio

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

	requiredReplicas := burstAdjustedRate / (capacity * targetUtilization)
	return max(int(math.Ceil(requiredReplicas)), 1)
}
