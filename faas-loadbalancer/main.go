package main

import (
	"context"
	"encoding/json"
	"errors"
	"faas-loadbalancer/internal/k8s"
	"faas-loadbalancer/internal/metrics"
	"faas-loadbalancer/internal/otel"
	"faas-loadbalancer/internal/routing"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// envs
const (
	KUBERNETES_NODE_NAME    = "KUBERNETES_NODE_NAME"
	METRICS_BACKEND_ADDRESS = "METRICS_BACKEND_ADDRESS"
	NODE_EVALUATOR_BIAS     = "NODE_EVALUATOR_BIAS"
)

var node k8s.Node
var metricsBackendAddr string

var router routing.Router
var tracer trace.Tracer

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

	component := r.Header.Get(routing.ForwardToHeader)
	if component == "" {
		log.Printf("no value was provided for header %v. could not route request", routing.ForwardToHeader)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("no value was provied for header %v", routing.ForwardToHeader)))
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body, err: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	traceCtx := otel.ExtractTraceContext(r, context.Background())
	_, span := tracer.Start(traceCtx, "forward_request",
		trace.WithAttributes(attribute.String("to.component", component)),
		trace.WithAttributes(attribute.String("current.node", string(node))),
	)
	forwardReq := routing.Request{
		ToComponent: routing.Component(component),
		Body:        bodyBytes,
		Context:     traceCtx,
	}
	go func(r routing.Request, span trace.Span) {
		defer span.End()
		route, err := router.RouteRequest(r)
		span.AddEvent("routing_finished", trace.WithAttributes(
			attribute.String("to.func", route.Name),
			attribute.String("to.node", string(route.Node)),
		))
		if err != nil {
			log.Printf("failed to route request to component: %v, err: %v", r.ToComponent, err)
		} else {
			log.Printf("successfully routed request to component: %v, partition: %v, node: %v", r.ToComponent, route.Name, route.Node)
		}
	}(forwardReq, span)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	// reader, _ := metrics.NewMetricsReader()
	// reader.Test()

	// w.WriteHeader(http.StatusOK)
	//w.Write([]byte(res))
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

	evaluatorBias := 0.3
	if os.Getenv(NODE_EVALUATOR_BIAS) != "" {
		float, err := strconv.ParseFloat(os.Getenv(NODE_EVALUATOR_BIAS), 64)
		if err != nil {
			evaluatorBias = float
		}
	}

	log.Printf("Finished reading env variables. %v: %v, %v: %v", KUBERNETES_NODE_NAME, node, METRICS_BACKEND_ADDRESS, metricsBackendAddr)
	// not using apikey as of now
	// apiKey, err := readESCredentials()
	// if err != nil {
	// 	log.Fatal("Failed to get apikey from file:", err)
	// }

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	otelShutdown, err := otel.SetupOTelSDK(ctx)
	if err != nil {
		log.Fatal("failed to initialize opentelemetry:", err)
	}
	tracer = otel.CreateTracer()

	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	funcLayout, err := readFunctionLayout()
	if err != nil {
		log.Fatal("failed to read function layout:", err)
	}

	evaluator := metrics.NewResourceBasedEvaluator(node, evaluatorBias)
	watcher, err := metrics.NewMetricsWatcher(evaluator, metricsBackendAddr)
	if err != nil {
		log.Fatal("Failed to create metrics watcher:", err)
	}

	router, err = routing.NewRouter(funcLayout, watcher)
	if err != nil {
		log.Fatal("failed to create router component:", err)
	}

	s := &http.Server{Addr: ":8080"}
	http.HandleFunc("/", ForwardHandler)
	http.HandleFunc("/healthz", HealthCheckHandler)
	http.HandleFunc("/test", TestHandler)

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
	handleFunc("/", ForwardHandler)
	handleFunc("/healthz", HealthCheckHandler)
	handleFunc("/test", TestHandler)

	// Add HTTP instrumentation for the whole server.
	handler := otelhttp.NewHandler(mux, "/")
	return handler
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
