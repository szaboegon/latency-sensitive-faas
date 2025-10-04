package core

import "math"

// calculateComponentMaxReplicas computes the number of replicas required for a given component
// based on its runtime, the total incoming invocation rate, and the target concurrency per replica.
//
// Parameters:
//   c              - The Component for which to calculate replicas. Its Runtime field is in milliseconds.
//   links          - A slice of ComponentLink representing all edges in the application graph.
//                    Only links where link.To == c.Name are considered for calculating the total invocation rate.
//   targetConcurrency - The target number of concurrent requests a single replica can handle.
//
// Returns:
//   int - The calculated number of replicas needed for this component, at least 1.
func calculateComponentMaxReplicas(c Component, links []ComponentLink, targetConcurrency int) int {
	totalInvocationRate := 0.0
	for _, link := range links {
		if link.To == c.Name {
			totalInvocationRate += link.InvocationRate
		}
	}

	if totalInvocationRate == 0 || c.Runtime <= 0 {
		return 1
	}

	// How many invocations can a single replica handle per second
	capacityPerReplica := (1000.0 / float64(c.Runtime)) * float64(targetConcurrency)

	// Calculate the maximum number of replicas
	maxReplicas := int(math.Ceil(totalInvocationRate / capacityPerReplica))
	if maxReplicas < 1 {
		maxReplicas = 1
	}
	return maxReplicas
}

func calculateDeploymentResources(comps []ComponentProfile) Resources {
	totalMemory := 0
	for _, comp := range comps {
		totalMemory += comp.TotalMemory()
	}
	return Resources{
		CPU:    1000, // Fixed at 1 core
		Memory: totalMemory,
	}
}

func calculateDeploymentMaxReplicas(comps []ComponentProfile, targetConcurrency int) Scale {
	// The deployment needs as many replicas as the component with the highest requirement
	maxReplicas := 1
	for _, comp := range comps {
		if comp.RequiredReplicas > maxReplicas {
			maxReplicas = comp.RequiredReplicas
		}
	}
	return Scale{MinReplicas: 1, MaxReplicas: maxReplicas, TargetConcurrency: targetConcurrency}
}
