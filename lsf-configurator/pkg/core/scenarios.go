package core

const (
	LayoutKeyMin = "min_rate"
	LayoutKeyAvg = "avg_rate"
	LayoutKeyMax = "max_rate"
)

var layoutUpgradePath = map[string]string{
	LayoutKeyMin: LayoutKeyAvg,
	LayoutKeyAvg: LayoutKeyMax,
	LayoutKeyMax: "", // No upgrade from Max
}

type scenarioManager struct {
	calculator LayoutCalculator
}

func NewScenarioManager(calculator LayoutCalculator) ScenarioManager {
	return &scenarioManager{calculator: calculator}
}

func (sm *scenarioManager) GenerateLayoutCandidates(
	components []Component,
	links []ComponentLink,
	appLatencyReq int,
	memoryAvailable int) (map[string]Layout, error) {
	rates := []struct {
		Name     string
		Key      string
		RateFunc func(min, max float64) float64
	}{
		{
			Name:     "Minimum Rate Layout",
			Key:      LayoutKeyMin,
			RateFunc: func(min, max float64) float64 { return min },
		},
		{
			Name:     "Average Rate Layout",
			Key:      LayoutKeyAvg,
			RateFunc: func(min, max float64) float64 { return (min + max) * 0.5 },
		},
		{
			Name:     "Maximum Rate Layout",
			Key:      LayoutKeyMax,
			RateFunc: func(min, max float64) float64 { return max },
		},
	}

	candidates := make(map[string]Layout, len(rates))

	compMap := make(map[string]Component)
	for _, c := range components {
		compMap[c.Name] = c
	}

	for _, r := range rates {
		layoutScenario := sm.buildLayoutScenario(compMap, links, r.RateFunc)
		layoutScenario.LatencyRequirement = appLatencyReq
		layoutScenario.AvailableNodeMemory = memoryAvailable

		layout, err := sm.calculator.CalculateLayout(*layoutScenario)
		if err != nil {
			return nil, err
		}
		candidates[r.Key] = layout
	}
	return candidates, nil
}

func (sm *scenarioManager) buildLayoutScenario(
	compMap map[string]Component,
	links []ComponentLink,
	rateFunc func(min, max float64) float64) *LayoutScenario {

	scenarioLinks := make([]ScenarioLink, 0, len(links))
	for _, link := range links {
		rate := rateFunc(link.InvocationRate.Min, link.InvocationRate.Max)
		scenarioLinks = append(scenarioLinks, ScenarioLink{
			From:           link.From,
			To:             link.To,
			InvocationRate: rate,
			DataDelay:      link.DataDelay,
		})
	}

	profiles := make([]ComponentProfile, 0, len(compMap))
	for _, comp := range compMap {
		// Calculate replicas using the provided helper function and the generated ScenarioLinks
		replicas := calculateComponentMaxReplicas(comp, scenarioLinks, targetConcurrency)
		profiles = append(profiles, ComponentProfile{
			Name:             comp.Name,
			Runtime:          comp.Runtime,
			Memory:           comp.Memory,
			RequiredReplicas: replicas,
		})
	}

	return &LayoutScenario{
		Profiles: profiles,
		Links:    scenarioLinks,
	}
}
