package api

import (
	"encoding/json"
	"log"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"net/http"
)

const (
	AppsPath = "/function_apps"
)

type HandlerApps struct {
	composer *core.Composer
	conf     config.Configuration
	mux      *http.ServeMux
}

func NewHandlerApps(composer *core.Composer, conf config.Configuration) *HandlerApps {
	h := &HandlerApps{
		composer: composer,
		conf:     conf,
		mux:      http.NewServeMux(),
	}

	h.mux.HandleFunc("GET "+AppsPath, h.list)
	h.mux.HandleFunc("GET "+AppsPath+"/{id}", h.get)
	h.mux.HandleFunc("POST "+AppsPath, h.create)
	h.mux.HandleFunc("DELETE "+AppsPath+"/{id}", h.delete)

	return h
}

func (h *HandlerApps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HandlerApps) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	apps, err := h.composer.ListFunctionApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(apps)
}

func (h *HandlerApps) get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	appId := r.PathValue("id")
	app, err := h.composer.GetFunctionApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func (h *HandlerApps) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	r.ParseMultipartForm(10 << 20) // 10MB limit

	jsonStr := r.FormValue("json")
	var payload struct {
		Name         string                     `json:"name"`
		Compositions []core.FunctionComposition `json:"compositions"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		http.Error(w, "No files were uploaded", http.StatusBadRequest)
		return
	}

	_, err := h.composer.CreateFunctionApp(h.conf.UploadDir, files, payload.Compositions, payload.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerApps) delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	appId := r.PathValue("id")
	err := h.composer.DeleteFunctionApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
