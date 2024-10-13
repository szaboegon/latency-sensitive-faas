package main

import (
	"context"
	"encoding/json"
	"errors"
	"faas-loadbalancer/internal/metrics"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var apiKeyConfig metrics.ApiKeyConfig

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Ok"))
}

func RouteRequestHandler(w http.ResponseWriter, r *http.Request) {

}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	reader, _ := metrics.NewMetricsReader(apiKeyConfig)
	json, _ := reader.Test()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(json))
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

	err = json.Unmarshal(byteValue, &apiKeyConfig)
	if err != nil {
		return metrics.ApiKeyConfig{}, err
	}

	return apiKeyConfig, nil
}

func main() {
	apiKey, err := readESCredentials()
	if err != nil {
		log.Fatal("Failed to get apikey from file:", err)
		panic(err)
	}

	_, err = metrics.NewMetricsWatcher(apiKey)
	if err != nil {
		log.Fatal("Failed to create metrics watcher:", err)
		panic(err)
	}
	//watcher.Watch()

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
