package api

import (
	"encoding/json"
	"log"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/metrics"
	"net/http"
)

const (
	MetricsPath = "/metrics/"
)

type HandlerMetrics struct {
	metricsClient metrics.MetricsReader
	conf          config.Configuration
	mux           *http.ServeMux
}

func NewHandlerMetrics(metricsClient metrics.MetricsReader, conf config.Configuration) *HandlerMetrics {
	h := &HandlerMetrics{
		metricsClient: metricsClient,
		conf:          conf,
		mux:           http.NewServeMux(),
	}

	h.mux.HandleFunc("GET "+MetricsPath+"apps/{app_id}", h.getAppMetrics)

	return h
}

func (h *HandlerMetrics) getAppMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	appId := r.PathValue("app_id")

	metrics, err := h.metricsClient.QueryAverageAppRuntime(appId)
	if err != nil {
		http.Error(w, "Failed to query metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
