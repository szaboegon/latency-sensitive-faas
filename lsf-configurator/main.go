package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/knative"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	UploadDir = "uploads/apps"
)

var composer *core.Composer

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
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

	fcApp := composer.CreateFunctionApp()
	appDir := filepath.Join(UploadDir, fcApp.Id)

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
		composer.AddFunctionComposition(fcApp.Id, fc)
	}
}

func main() {
	knClient := knative.NewClient()
	composer = core.NewComposer(knClient)
	err := filesystem.CreateDir(UploadDir)
	if err != nil {
		log.Fatalf("Could not create uploads directory: %v", err)
	}

	s := startHttpServer()
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	<-signalCh
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}

func startHttpServer() *http.Server {
	s := &http.Server{Addr: ":8080"}
	registerHandlers()

	go func() {
		log.Printf("Server listening on port 8080")
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %s\n", err)
		}
	}()

	return s
}

func registerHandlers() {
	http.HandleFunc("/healthz", HealthCheckHandler)
	http.HandleFunc("/upload", UploadHandler)
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
}
