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

func NewHandlerApps(composer *core.Composer, conf config.Configuration) *HandlerApps {
	h := &HandlerApps{
		composer: composer,
		conf:     conf,
		mux:      http.NewServeMux(),
	}

	h.mux.HandleFunc("GET /", h.list)
	h.mux.HandleFunc("GET /{id}", h.get)
	h.mux.HandleFunc("POST /", h.create)
	h.mux.HandleFunc("POST /bulk", h.bulkCreate)
	h.mux.HandleFunc("DELETE /{id}", h.delete)

	return h
}

func (h *HandlerApps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	LoggingMiddleware(h.mux).ServeHTTP(w, r)
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
	r.ParseMultipartForm(10 << 20) // 10MB limit

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

	// Create the function app
	app, err := h.composer.CreateFunctionApp(h.conf.UploadDir, files, payload.FunctionApp.Name, payload.FunctionApp.Runtime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map temporary IDs to real IDs for function compositions
	compositionIdMap := make(map[string]string)

	for _, composition := range payload.FunctionCompositions {
		fc, err := h.composer.AddFunctionComposition(app.Id, composition.Components, composition.Files)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		compositionIdMap[composition.TempId] = fc.Id
	}

	// Map temporary deployment IDs to real deployment IDs
	deploymentIdMap := make(map[string]string)

	for _, deployment := range payload.Deployments {
		realCompositionId, exists := compositionIdMap[deployment.TempFunctionCompositionId]
		if !exists {
			http.Error(w, "Invalid FunctionCompositionId in deployment", http.StatusBadRequest)
			return
		}

		// Create deployment with an empty routing table
		dep, err := h.composer.CreateFcDeployment(realCompositionId, deployment.Namespace, deployment.Node, core.RoutingTable{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		deploymentIdMap[deployment.TempId] = dep.Id
	}

	// Translate routing tables and set them
	for _, deployment := range payload.Deployments {
		realDeploymentId, exists := deploymentIdMap[deployment.TempId]
		if !exists {
			http.Error(w, "Invalid deployment ID", http.StatusBadRequest)
			return
		}

		translatedRoutingTable := core.RoutingTable{}
		for component, routes := range deployment.RoutingTable {
			var translatedRoutes []core.Route
			for _, route := range routes {
				translatedDeploymentId, exists := deploymentIdMap[route.Function]
				if !exists {
					http.Error(w, "Invalid function reference in routing table", http.StatusBadRequest)
					return
				}
				translatedRoutes = append(translatedRoutes, core.Route{
					To:       route.To,
					Function: translatedDeploymentId,
				})
			}
			translatedRoutingTable[component] = translatedRoutes
		}

		if err := h.composer.SetRoutingTable(realDeploymentId, translatedRoutingTable); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
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
