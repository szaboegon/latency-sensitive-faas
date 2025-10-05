package core

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// set target concurrency globally to 1 for now, but this can be different per function composition depending on how CPU-bound they are
// this should be measured and set accordingly for each function composition)
const targetConcurrency = 1

type latencyController struct {
	composer              *Composer
	metrics               MetricsReader
	scenarioManager       ScenarioManager
	delay                 time.Duration
	deployNamespace       string
	lastReconfigs         map[string]time.Time
	cooldownPeriod        time.Duration
	availableNodeMemoryGb int // same for all nodes for now, in GB
}

func NewController(composer *Composer, metrics MetricsReader, scenarioManager ScenarioManager,
	delay time.Duration, deployNamespace string, availableNodeMemoryGb int) Controller {
	return &latencyController{
		composer:              composer,
		metrics:               metrics,
		scenarioManager:       scenarioManager,
		delay:                 delay,
		deployNamespace:       deployNamespace,
		lastReconfigs:         make(map[string]time.Time),
		cooldownPeriod:        60 * time.Second,
		availableNodeMemoryGb: availableNodeMemoryGb,
	}
}

func (c *latencyController) Start(ctx context.Context) error {
	log.Println("Latency Controller started")
	ticker := time.NewTicker(c.delay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Latency Controller received cancellation signal")
			return nil
		case <-ticker.C:
			runtimes, err := c.metrics.Query95thPercentileAppRuntimes()
			if err != nil {
				log.Printf("Error querying 95th percentile app runtimes: %v", err)
				continue
			}
			for appId, runtime := range runtimes {
				log.Printf("App %s 95th percentile runtime: %.0f ms", appId, runtime)
				app, err := c.composer.GetFunctionApp(appId)
				if err != nil {
					log.Printf("Error retrieving function app %s: %v", appId, err)
					continue
				}
				if app.LatencyLimit > 0 && runtime > float64(app.LatencyLimit) {
					last, ok := c.lastReconfigs[app.Id]
					if ok && time.Since(last) < c.cooldownPeriod {
						log.Printf("Skipping reconfiguration for app %s due to cooldown (last at %v)", app.Id, last)
						continue
					}

					go func(a *FunctionApp) {
						if err := c.handleLatencyViolation(a); err != nil {
							log.Printf("Error handling latency violation for app %s: %v", a.Id, err)
						}
					}(app)
				}
			}
		}
	}
}

func (c *latencyController) RegisterFunctionApp(creationData FunctionAppCreationData) (*FunctionApp, error) {
	app, err := c.composer.CreateFunctionApp(creationData)
	if err != nil {
		log.Printf("Error creating function app: %v", err)
		return nil, err
	}

	candidates, err := c.scenarioManager.GenerateLayoutCandidates(
		app.Components,
		app.Links,
		app.LatencyLimit,
		c.availableNodeMemoryGb*1024)
	if err != nil {
		log.Printf("Error generating layout candidates for app %s: %v", app.Id, err)
		return nil, err
	}
	app.LayoutCandidates = candidates
	c.composer.functionAppRepo.Save(app)

	for _, layout := range app.LayoutCandidates {
		for _, components := range layout {
			componentNames := make([]string, len(components))
			for i, cp := range components {
				componentNames[i] = cp.Name
			}
			_, err := c.composer.AddFunctionComposition(app.Id, componentNames)
			if err != nil {
				log.Printf("Error adding function composition for app %s: %v", app.Id, err)
				return nil, err
			}
		}
	}

	app.ActiveLayoutKey = LayoutKeyMin
	go func(appId string, layout Layout) {
		err = c.deployLayout(appId, layout)
		if err != nil {
			log.Printf("Error deploying layout for app %s: %v", appId, err)
			return
		}
		log.Printf("Successfully deployed function app with layout %s: %v", appId, layout)
		// deploy the first layout by default
	}(app.Id, app.LayoutCandidates[LayoutKeyMin])

	return app, nil
}

// TODO think about a backwards direction: if the best layout is already active, and the app is still over the limit, should we try a less resource-intensive layout?
func (c *latencyController) handleLatencyViolation(app *FunctionApp) error {
	log.Printf("App %s exceeds latency threshold (%d ms). Triggering reconfiguration.", app.Id, app.LatencyLimit)

	nextLayoutKey := layoutUpgradePath[app.ActiveLayoutKey]
	if nextLayoutKey == "" {
		log.Printf("No further layout candidates available for app %s. Skipping reconfiguration.", app.Id)
		return nil
	}

	nextLayout, ok := app.LayoutCandidates[nextLayoutKey]
	if !ok {
		return fmt.Errorf("no layout candidate found for key %s in app %s", nextLayoutKey, app.Id)
	}

	app.ActiveLayoutKey = nextLayoutKey
	err := c.composer.functionAppRepo.Save(app)
	if err != nil {
		return fmt.Errorf("failed to update active layout key for app %s: %w", app.Id, err)
	}

	if err := c.deployLayout(app.Id, nextLayout); err != nil {
		return fmt.Errorf("failed to deploy new layout for app %s: %w", app.Id, err)
	}

	return nil
}

func (c *latencyController) deployLayout(appId string, layout Layout) error {
	log.Printf("Deploying layout for app %s: %v", appId, layout)

	app, err := c.composer.GetFunctionApp(appId)
	if err != nil {
		return err
	}

	// Build a fast lookup: compositionKey -> FunctionComposition
	fcByKey := make(map[string]*FunctionComposition)
	for _, fc := range app.Compositions {
		fcByKey[componentsKey(fc.Components)] = fc
	}

	// Build a map of currently active deployments keyed by fcId@node
	// We are going to modify this map as we reuse/create deployments
	activeDepsByKey := make(map[string]*Deployment)
	oldCompToDepID := make(map[string]string) // component -> deployment id mapping from old layout, used as fallback
	for _, fc := range app.Compositions {
		for _, d := range fc.Deployments {
			k := fc.Id + "@" + d.Node
			activeDepsByKey[k] = d
			for _, comp := range fc.Components {
				// Snapshot old component -> deployment id mapping (fallback)
				if _, ok := oldCompToDepID[comp]; !ok {
					oldCompToDepID[comp] = d.Id
				}
			}
		}
	}

	// single-pass creation/reuse and build comp -> dep mapping
	activeDepIDs := make(map[string]bool)  // set of active dep ids
	compToDepID := make(map[string]string) // component -> deployment id

	type depResult struct {
		key string
		dep *Deployment
		err error
	}

	resultChan := make(chan depResult, len(layout))

	compRuntimeMap := make(map[string]int) // component -> runtime in ms
	for _, comp := range app.Components {
		compRuntimeMap[comp.Name] = comp.Runtime
	}
	for node, components := range layout {
		// process each node+components in parallel
		go func(node string, components []ComponentProfile) {
			componentNames := make([]string, len(components))
			for i, cp := range components {
				componentNames[i] = cp.Name
			}
			fcKey := componentsKey(componentNames)
			matchedFc, ok := fcByKey[fcKey]
			if !ok {
				resultChan <- depResult{"", nil, fmt.Errorf("no matching function composition for components: %v", components)}
				return
			}

			depKey := matchedFc.Id + "@" + node
			if dep, ok := activeDepsByKey[depKey]; ok {
				// if a deployment already exists for this fc+node, reuse it
				log.Printf("Reusing existing deployment %s for node %s", dep.Id, node)
				resultChan <- depResult{depKey, dep, nil}
				return
			}

			// otherwise, create a new deployment, routing table will be set later once all deployments ids are known
			log.Printf("No existing deployment found for node %s, creating new", node)
			resources := calculateDeploymentResources(components)
			scale := calculateDeploymentMaxReplicas(components, targetConcurrency)

			emptyRT := make(RoutingTable)
			newDep, depChan, err := c.composer.CreateFcDeployment(matchedFc.Id, c.deployNamespace, node, emptyRT, scale, resources)
			if err != nil {
				resultChan <- depResult{"", nil, fmt.Errorf("failed to create deployment for fc %s on node %s: %w", matchedFc.Id, node, err)}
				return
			}

			r := <-depChan
			if r.Err != nil {
				resultChan <- depResult{"", nil, fmt.Errorf("deployment task failed for fc %s on node %s: %w", matchedFc.Id, node, r.Err)}
				return
			}

			matchedFc.Deployments = append(matchedFc.Deployments, newDep) // add to fc's deployments
			log.Printf("Created new deployment %s for fc %s on node %s", newDep.Id, matchedFc.Id, node)
			resultChan <- depResult{depKey, newDep, nil}
		}(node, components)
	}

	// collect results and update the active deployment mapping
	for i := 0; i < len(layout); i++ {
		res := <-resultChan
		if res.err != nil {
			return res.err
		}
		activeDepsByKey[res.key] = res.dep
		activeDepIDs[res.dep.Id] = true
		fc := getFunctionComposition(app, res.dep.FunctionCompositionId)
		for _, comp := range fc.Components {
			compToDepID[comp] = res.dep.Id
		}
	}

	// build and apply routing tables
	referencedDepIDs := make(map[string]bool)
	for node, comps := range layout {
		componentNames := make([]string, len(comps))
		for i, cp := range comps {
			componentNames[i] = cp.Name
		}

		fcKey := componentsKey(componentNames)
		matchedFc := fcByKey[fcKey]
		depKey := matchedFc.Id + "@" + node
		dep := activeDepsByKey[depKey]

		if dep == nil {
			panic(fmt.Sprintf("unexpected error: missing deployment for key %s", depKey))
		}

		rt := make(RoutingTable)
		for _, comp := range componentNames {
			var routes []Route
			for _, link := range app.Links {
				if link.From != comp {
					continue
				}
				if targetDepID, ok := compToDepID[link.To]; ok {
					routes = append(routes, Route{To: link.To, Function: targetDepID})
					referencedDepIDs[targetDepID] = true
					continue
				}
				// fallback: use old deployment if available
				if oldDepID, ok := oldCompToDepID[link.To]; ok {
					routes = append(routes, Route{To: link.To, Function: oldDepID})
					referencedDepIDs[oldDepID] = true
					log.Printf("Warning: using fallback routing for component %s to old deployment %s", link.To, oldDepID)
					continue
				}
				panic(fmt.Sprintf("no new or fallback deployment available for component %s", link.To))
			}
			rt[comp] = routes
		}

		err = c.composer.SetRoutingTable(dep.Id, rt)
		if err != nil {
			return fmt.Errorf("failed to set routing table for deployment %s: %w", dep.Id, err)
		}
		log.Printf("Set routing table for deployment %s: %v", dep.Id, rt)
	}

	// cleanup: remove unused deployments asynchronously
	go func() {
		time.Sleep(10 * time.Second)
		for _, fc := range app.Compositions {
			var kept []*Deployment
			for _, d := range fc.Deployments {
				if activeDepIDs[d.Id] || referencedDepIDs[d.Id] {
					kept = append(kept, d)
					continue
				}
				log.Printf("Deleting old deployment %s (fc %s on node %s)", d.Id, fc.Id, d.Node)
				delChan, err := c.composer.DeleteFcDeployment(d.Id)
				if err != nil {
					log.Printf("Failed to delete deployment %s: %v", d.Id, err)
					kept = append(kept, d)
					continue
				}
				r := <-delChan
				if r.Err != nil {
					log.Printf("Failed to delete deployment %s: %v", d.Id, r.Err)
					kept = append(kept, d)
				} else {
					log.Printf("Deleted deployment %s", d.Id)
				}
			}
			fc.Deployments = kept
		}
	}()

	return nil
}

func getFunctionComposition(app *FunctionApp, fcId string) *FunctionComposition {
	for _, fc := range app.Compositions {
		if fc.Id == fcId {
			return fc
		}
	}
	return nil
}

func componentsKey(comps []string) string {
	if len(comps) == 0 {
		return ""
	}
	cp := append([]string(nil), comps...)
	sort.Strings(cp)
	return strings.Join(cp, ",")
}
