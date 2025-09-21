package core

import (
	"context"
	"log"
	"time"
)

type latencyController struct {
	composer *Composer
	metrics  MetricsReader
	layout   LayoutCalculator
	delay    time.Duration
}

func NewController(composer *Composer, metrics MetricsReader, layout LayoutCalculator, delay time.Duration) Controller {
	return &latencyController{
		composer: composer,
		metrics:  metrics,
		layout:   layout,
		delay:    delay,
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
			runtimes , err := c.metrics.Query95thPercentileAppRuntimes()
			if err != nil {
				log.Printf("Error querying 95th percentile app runtimes: %v", err)
				continue
			}
			for appId, runtime := range runtimes {
				log.Printf("App %s 95th percentile runtime: %.2f ms", appId, runtime)
				app, err := c.composer.GetFunctionApp(appId)
				if err != nil {
					log.Printf("Error retrieving function app %s: %v", appId, err)
					continue
				}
				if app.LatencyLimit > 0 && runtime > float64(app.LatencyLimit) {
					go func (a *FunctionApp)  {
						if err := c.handleLatencyViolation(a); err != nil {
							log.Printf("Error handling latency violation for app %s: %v", a.Id, err)
						}
					}(app)
				}
			}
        }
    }
}

func (c *latencyController) handleLatencyViolation(app *FunctionApp) error {
	log.Printf("App %s exceeds latency threshold (%d ms). Triggering reconfiguration.", app.Id, app.LatencyLimit)
	// TODO either calculate new layout here, or use one of the pre-calculated layouts
	layout, err := c.layout.CalculateLayout(*app)
	if err != nil {
		log.Printf("Error calculating layout for app %s: %v", app.Id, err)
		return err
	}
	log.Printf("Calculated layout for app %s: %v", app.Id, layout)
	return nil
}