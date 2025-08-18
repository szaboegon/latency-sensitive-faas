package api

import (
	"encoding/json"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"net/http"
)

const (
	DeploymentsPath = "/deployments"
)

type HandlerDeployments struct {
	composer *core.Composer
	conf     config.Configuration
	mux      *http.ServeMux
}

func NewHandlerDeployments(composer *core.Composer, conf config.Configuration, mux *http.ServeMux) *HandlerDeployments {
	h := &HandlerDeployments{
		composer: composer,
		conf:     conf,
		mux:      mux,
	}

	h.mux.HandleFunc("POST "+DeploymentsPath, h.create)
	h.mux.HandleFunc("PUT "+DeploymentsPath+"/{deployment_id}/routing_table", h.putRoutingTable)

	return h
}

func (h *HandlerDeployments) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FunctionCompositionId string            `json:"function_composition_id"`
		Node                  string            `json:"node"`
		Namespace             string            `json:"namespace"`
		RoutingTable          core.RoutingTable `json:"routing_table"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	deployment, err := h.composer.CreateFcDeployment(req.FunctionCompositionId, req.Node, req.Namespace, req.RoutingTable)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

func (h *HandlerDeployments) putRoutingTable(w http.ResponseWriter, r *http.Request) {
	deploymentId := r.PathValue("deployment_id")

	var rt core.RoutingTable
	if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.composer.SetRoutingTable(deploymentId, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
