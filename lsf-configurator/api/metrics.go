package api

import (
	"encoding/json"
	"log"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/metrics"
	"net/http"
)

const (
	MetricsPath = "/metrics"
)

type HandlerMetrics struct {
	metricsClient metrics.MetricsReader
	conf          config.Configuration
	mux           *http.ServeMux
}

func NewHandlerMetrics(mux *http.ServeMux, metricsClient metrics.MetricsReader, conf config.Configuration) *HandlerMetrics {
	h := &HandlerMetrics{
		metricsClient: metricsClient,
		conf:          conf,
		mux:           mux,
	}

	h.mux.HandleFunc("GET "+MetricsPath+"/apps/{app_id}", h.getAppMetrics)

	return h
}

func (h *HandlerMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HandlerMetrics) getAppMetrics(w http.ResponseWriter, r *http.Request) {
	appId := r.PathValue("app_id")

	metrics, err := h.metricsClient.QueryAverageAppRuntime(appId)
	if err != nil {
		log.Printf("Error querying metrics: %v", err)
		http.Error(w, "Failed to query metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
