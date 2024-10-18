package main

import (
	"context"
	"encoding/json"
	"errors"
	"faas-loadbalancer/internal/k8s"
	"faas-loadbalancer/internal/metrics"
	"faas-loadbalancer/internal/routing"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var NODE k8s.Node
var router routing.Router

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func RouteRequestHandler(w http.ResponseWriter, r *http.Request) {

}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	reader, _ := metrics.NewMetricsReader()
	reader.Test()

	w.WriteHeader(http.StatusOK)
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
	NODE = "default"
	if os.Getenv("NODE_NAME") != "" {
		NODE = k8s.Node(os.Getenv("NODE_NAME"))
	}

	// not using apikey as of now
	// apiKey, err := readESCredentials()
	// if err != nil {
	// 	log.Fatal("Failed to get apikey from file:", err)
	// 	panic(err)
	// }

	funcLayout, err := readFunctionLayout()
	if err != nil {
		log.Fatal("failed to read function layout:", err)
		panic(err)
	}

	router, err = routing.NewRouter(funcLayout, NODE)
	if err != nil {
		log.Fatal("failed to create router component:", err)
		panic(err)
	}

	s := &http.Server{Addr: ":8080"}
	http.HandleFunc("/healthz", HealthCheckHandler)
	http.HandleFunc("/", RouteRequestHandler)
	http.HandleFunc("/test", TestHandler)

	go func() {
		log.Printf("Server listening on %v", s.Addr)
		if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	// handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // wait for termination signal
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Minute*5))
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}
