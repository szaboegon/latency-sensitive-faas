package api

import (
	"encoding/json"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"net/http"
)

const (
	FunctionCompositionsPath = "/function_compositions"
)

type HandlerFunctionCompositions struct {
	composer *core.Composer
	conf     config.Configuration
	mux      *http.ServeMux
}

func NewHandlerFunctionCompositions(composer *core.Composer, conf config.Configuration) *HandlerFunctionCompositions {
	h := &HandlerFunctionCompositions{
		composer: composer,
		conf:     conf,
		mux:      http.NewServeMux(),
	}

	h.mux.HandleFunc("PUT "+FunctionCompositionsPath+"/{fc_id}/routing_table", h.putRoutingTable)
	h.mux.HandleFunc("POST "+FunctionCompositionsPath+"/build_notify", h.buildNotify)
	h.mux.HandleFunc("POST "+FunctionCompositionsPath, h.create)

	return h
}

func (h *HandlerFunctionCompositions) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HandlerFunctionCompositions) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	appId := r.URL.Query().Get("app_id")
	if appId == "" {
		http.Error(w, "Missing app_id", http.StatusBadRequest)
		return
	}

	var fc core.FunctionComposition
	if err := json.NewDecoder(r.Body).Decode(&fc); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.composer.AddFunctionComposition(appId, &fc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *HandlerFunctionCompositions) putRoutingTable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	fcId := r.PathValue("fc_id")

	var rt core.RoutingTable
	if err := json.NewDecoder(r.Body).Decode(&rt); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err := h.composer.SetRoutingTable(fcId, rt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerFunctionCompositions) buildNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		FcId     string `json:"fc_id"`
		ImageURL string `json:"image_url"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	go h.composer.NotifyBuildReady(req.FcId, req.ImageURL, req.Status)
	w.WriteHeader(http.StatusOK)
}
