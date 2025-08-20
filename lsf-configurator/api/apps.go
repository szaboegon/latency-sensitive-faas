package api

import (
	"encoding/json"
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

func NewHandlerApps(mux *http.ServeMux, composer *core.Composer, conf config.Configuration) *HandlerApps {
	h := &HandlerApps{
		composer: composer,
		conf:     conf,
		mux:      mux,
	}

	h.mux.HandleFunc("GET "+AppsPath, h.list)
	h.mux.HandleFunc("GET "+AppsPath+"/{id}", h.get)
	h.mux.HandleFunc("POST "+AppsPath, h.create)
	h.mux.HandleFunc("POST "+AppsPath+"/bulk_create", h.bulkCreate)
	h.mux.HandleFunc("DELETE "+AppsPath+"/{id}", h.delete)

	return h
}

func (h *HandlerApps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HandlerApps) list(w http.ResponseWriter, r *http.Request) {
	apps, err := h.composer.ListFunctionApps()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(apps)
}

func (h *HandlerApps) get(w http.ResponseWriter, r *http.Request) {
	appId := r.PathValue("id")
	app, err := h.composer.GetFunctionApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(app)
}

func (h *HandlerApps) create(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10MB limit

	jsonStr := r.FormValue("json")
	var payload FunctionAppCreateDto
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		http.Error(w, "No files were uploaded", http.StatusBadRequest)
		return
	}

	_, err := h.composer.CreateFunctionApp(h.conf.UploadDir, files, payload.Name, payload.Runtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerApps) bulkCreate(w http.ResponseWriter, r *http.Request) {
	jsonStr := r.FormValue("json")

	var payload BulkCreateRequest
	if err := json.Unmarshal([]byte(jsonStr), &payload); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		http.Error(w, "No files were uploaded", http.StatusBadRequest)
		return
	}

	// function app creation
	app, err := h.composer.CreateFunctionApp(h.conf.UploadDir, files,
		payload.FunctionApp.Name, payload.FunctionApp.Runtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// function composition creation
	for _, comp := range payload.FunctionCompositions {
		fc, err := h.composer.AddFunctionComposition(app.Id, comp.Components, comp.Files)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, dep := range comp.Deployments {
			_, err := h.composer.CreateFcDeployment(fc.Id,
				dep.Namespace, dep.Node, dep.RoutingTable)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

func (h *HandlerApps) delete(w http.ResponseWriter, r *http.Request) {
	appId := r.PathValue("id")
	err := h.composer.DeleteFunctionApp(appId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
