package main

import (
	"context"
	"errors"
	"io"
	"log"
	"lsf-configurator/api"
	"lsf-configurator/pkg/builder"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/data/db"
	"lsf-configurator/pkg/data/repos"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/knative"
	"lsf-configurator/pkg/metrics"
	"lsf-configurator/pkg/routing"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var composer *core.Composer
var conf config.Configuration
var metricsReader metrics.MetricsReader

func main() {
	logFile := configureLogging()
	defer logFile.Close()

	disableStdOut()
	conf = config.Init()

	knClient := knative.NewClient(conf)
	routingClient := routing.NewRouteConfigurator(conf.RedisUrl)
	tektonConf := builder.TektonConfig{
		Namespace:      conf.TektonNamespace,
		Pipeline:       conf.TektonPipeline,
		NotifyURL:      conf.TektonNotifyURL,
		WorkspacePVC:   conf.TektonWorkspacePVC,
		ImageRepo:      conf.ImageRepository,
		ServiceAccount: conf.TektonServiceAccount,
	}
	tektonBuilder := builder.NewTektonBuilder(tektonConf, conf.TektonConcurrencyLimit)

	db, err := db.InitDB(conf.DatabasePath)
	log.Printf("Database initialized successfully at path: %s", conf.DatabasePath)

	functionAppRepo := repos.NewFunctionAppRepository(db)
	fcRepo := repos.NewFunctionCompositionRepository(db)
	deploymentRepo := repos.NewDeploymentRepository(db)

	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	composer = core.NewComposer(functionAppRepo, fcRepo, deploymentRepo, routingClient, knClient, tektonBuilder)

	metricsReader, err = metrics.NewMetricsReader(conf.MetricsBackendAddress)
	if err != nil {
		log.Fatalf("failed to create metrics reader: %v", err)
	}
	err = filesystem.CreateDir(conf.UploadDir)
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
	mux := http.NewServeMux()
	s := &http.Server{Addr: "0.0.0.0:8080", Handler: loggingMiddleware(corsMiddleware(mux))}
	registerHandlers(mux)

	go func() {
		log.Printf("Server listening on port 8080")
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %s\n", err)
		}
	}()

	return s
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%p] %s %s", r, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func registerHandlers(mux *http.ServeMux) {
	http.HandleFunc(api.HealthzPath, api.HealthCheckHandler)
	http.Handle(api.AppsPath, api.NewHandlerApps(mux, composer, conf))
	http.Handle(api.FunctionCompositionsPath, api.NewHandlerFunctionCompositions(mux, composer, conf))
	http.Handle(api.MetricsPath, api.NewHandlerMetrics(mux, metricsReader, conf))

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)
}

func configureLogging() *os.File {
	f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	return f
}

// needed so knative library does not write stdout into docker logs
func disableStdOut() {
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		os.Stdout = null
		os.Stderr = null
	}
}
