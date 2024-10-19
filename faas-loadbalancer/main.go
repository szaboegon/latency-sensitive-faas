package main

import (
	"context"
	"encoding/json"
	"errors"
	"faas-loadbalancer/internal/k8s"
	"faas-loadbalancer/internal/metrics"
	"faas-loadbalancer/internal/otel"
	"faas-loadbalancer/internal/routing"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var node k8s.Node
var metricsBackendAddr string

var router routing.Router

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func ForwardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[%p] %s %s", r, r.Method, r.URL)

	component := r.PathValue("component")
	forwardReq := routing.Request{
		ToComponent: routing.Component(component),
		BodyReader:  r.Body,
	}

	err := router.RouteRequest(forwardReq)
	if err != nil {
		log.Fatalf("failed to route request to component: %v, err: %v", forwardReq.ToComponent, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	// reader, _ := metrics.NewMetricsReader()
	// reader.Test()

	// w.WriteHeader(http.StatusOK)
	//w.Write([]byte(res))
}

func readESCredentials() (metrics.ApiKeyConfig, error) {
	apikeyFile, err := os.Open("apikey.json")
	if err != nil {
		return metrics.ApiKeyConfig{}, err
	}

	defer apikeyFile.Close()
	byteValue, err := io.ReadAll(apikeyFile)
	if err != nil {
		return metrics.ApiKeyConfig{}, err
	}

	var apiKeyConfig metrics.ApiKeyConfig
	err = json.Unmarshal(byteValue, &apiKeyConfig)
	if err != nil {
		return metrics.ApiKeyConfig{}, err
	}

	return apiKeyConfig, nil
}

func readFunctionLayout() (routing.FunctionLayout, error) {
	layoutFile, err := os.Open("function-layout.json")
	if err != nil {
		return routing.FunctionLayout{}, err
	}

	defer layoutFile.Close()
	byteValue, err := io.ReadAll(layoutFile)
	if err != nil {
		return routing.FunctionLayout{}, err
	}

	var funcLayout routing.FunctionLayout
	err = json.Unmarshal(byteValue, &funcLayout)
	if err != nil {
		return routing.FunctionLayout{}, err
	}

	return funcLayout, nil
}

func main() {
	node = "default"
	if os.Getenv(KUBERNETES_NODE_NAME) != "" {
		node = k8s.Node(os.Getenv(KUBERNETES_NODE_NAME))
	}

	metricsBackendAddr = "http://localhost:9200"
	if os.Getenv(METRICS_BACKEND_ADDRESS) != "" {
		metricsBackendAddr = os.Getenv(METRICS_BACKEND_ADDRESS)
	}

	// not using apikey as of now
	// apiKey, err := readESCredentials()
	// if err != nil {
	// 	log.Fatal("Failed to get apikey from file:", err)
	// 	panic(err)
	// }

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	otelShutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		log.Fatal("failed to initialize opentelemetry:", err)
		panic(err)
	}

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	funcLayout, err := readFunctionLayout()
	if err != nil {
		log.Fatal("failed to read function layout:", err)
		panic(err)
	}

	watcher, err := metrics.NewMetricsWatcher(node, metricsBackendAddr)
	if err != nil {
		log.Fatal("Failed to create metrics watcher:", err)
		panic(err)
	}

	router, err = routing.NewRouter(funcLayout, watcher)
	if err != nil {
		log.Fatal("failed to create router component:", err)
		panic(err)
	}

	s := &http.Server{
		Addr:        ":8080",
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     newHTTPHanlder(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- s.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return
	case <-ctx.Done():
		stop()
	}

	// graceful shutdown: wait for requests to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Minute*5))
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}

// needed for otel instrumentation
func newHTTPHanlder() http.Handler {
	mux := http.NewServeMux()

	// handleFunc is a replacement for mux.HandleFunc
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handleFunc := func(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
		// Configure the "http.route" for the HTTP instrumentation.
		handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
		mux.Handle(pattern, handler)
	}

	// Register handlers.
	handleFunc("/healthz", HealthCheckHandler)
	handleFunc("/forward/{component}", ForwardHandler)
	handleFunc("/test", TestHandler)

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
}
