package api

import (
	"encoding/json"
	"log"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"net/http"
	"path/filepath"
)

const (
	AppsPath = "/apps/"
)

type HandlerApps struct {
	composer core.Composer
	conf     config.Configuration
	mux      *http.ServeMux
}

func NewHandlerApps(composer core.Composer, conf config.Configuration) *HandlerApps {
	h := &HandlerApps{
		composer: composer,
		conf:     conf,
		mux:      http.NewServeMux(),
	}

	h.mux.HandleFunc(AppsPath+"create", h.create)

	return h
}

func (h *HandlerApps) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HandlerApps) create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	r.ParseMultipartForm(10 << 20) // 10MB limit

	jsonStr := r.FormValue("json")
	var fcs []core.FunctionComposition
	if err := json.Unmarshal([]byte(jsonStr), &fcs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		http.Error(w, "No files were uploaded", http.StatusBadRequest)
		return
	}

	fcApp := h.composer.CreateFunctionApp()
	appDir := filepath.Join(h.conf.UploadDir, fcApp.Id)

	err := filesystem.CreateDir(appDir)
	if err != nil {
		http.Error(w, "Could not create directory for app files", http.StatusInternalServerError)
		return
	}
	for _, fileHeader := range files {
		err := filesystem.SaveMultiPartFile(fileHeader, appDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	for _, fc := range fcs {
		fc.SourcePath = appDir
		h.composer.AddFunctionComposition(fcApp.Id, fc)
	}
}
