package core

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"sort"
	"strings"
	"time"
)

type latencyController struct {
	composer        *Composer
	metrics         MetricsReader
	layout          LayoutCalculator
	delay           time.Duration
	deployNamespace string
}

func NewController(composer *Composer, metrics MetricsReader, layout LayoutCalculator,
	delay time.Duration, deployNamespace string) Controller {
	return &latencyController{
		composer:        composer,
		metrics:         metrics,
		layout:          layout,
		delay:           delay,
		deployNamespace: deployNamespace,
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

	layouts, err := c.layout.CalculateLayouts(*app)
	if err != nil {
		log.Printf("Error calculating layout for app %s: %v", app.Id, err)
		return nil, err
	}

	for _, layout := range layouts {
		for _, components := range layout {
			_, err := c.composer.AddFunctionComposition(app.Id, components)
			if err != nil {
				log.Printf("Error adding function composition for app %s: %v", app.Id, err)
				return nil, err
			}
		}
	}

	err = c.deployLayout(app.Id, layouts[0]) // deploy the first layout by default
	if err != nil {
		log.Printf("Error deploying layout for app %s: %v", app.Id, err)
		return nil, err
	}

	log.Printf("Calculated layouts for app %s: %v", app.Id, layouts)
	return app, nil
}

func (c *latencyController) handleLatencyViolation(app *FunctionApp) error {
	log.Printf("App %s exceeds latency threshold (%d ms). Triggering reconfiguration.", app.Id, app.LatencyLimit)
	//TODO select a different layout candidate based on some criteria and then call deployLayout

	return nil
}

// //TODO https://chatgpt.com/share/68d6d4d1-5ba0-8009-9d2c-bc51efa84e16
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
	for node, components := range layout {
		// process each node+components in parallel
		go func(node string, components []string) {
			// check if function composition exists for this component set
			fcKey := componentsKey(components)
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
			emptyRT := make(RoutingTable)
			// maxReplicas=0 means no bounds, TODO minReplicas hardcoded to 1 for now
			newDep, depChan, err := c.composer.CreateFcDeployment(matchedFc.Id, node, c.deployNamespace, emptyRT, Scale{MinReplicas: 1, MaxReplicas: 0})
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
		fcKey := componentsKey(comps)
		matchedFc := fcByKey[fcKey]
		depKey := matchedFc.Id + "@" + node
		dep := activeDepsByKey[depKey]

		if dep == nil {
			panic(fmt.Sprintf("unexpected error: missing deployment for key %s", depKey))
		}

		rt := make(RoutingTable)
		for _, comp := range comps {
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

// calculates minimum required replicas based on edges and invocation rates
func calculateMinReplicas(fc *FunctionComposition, links []ComponentLink) int {
	compDemand := make(map[string]float64)
	for _, comp := range fc.Components {
		compDemand[comp] = 1 // default for components with no incoming link
	}

	for _, link := range links {
		if slices.Contains(fc.Components, link.To) {
			compDemand[link.To] += link.InvocationRate
		}
	}

	maxDemand := 0.0
	for _, comp := range fc.Components {
		if compDemand[comp] > maxDemand {
			maxDemand = compDemand[comp]
		}
	}

	return int(math.Ceil(maxDemand))
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
