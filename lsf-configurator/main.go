package main

import (
	"context"
	"errors"
	"io"
	"log"
	"lsf-configurator/api"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/docker"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/knative"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var composer *core.Composer
var conf config.Configuration
var puller docker.ImagePuller

func main() {
	logFile := configureLogging()
	defer logFile.Close()

	conf = config.Init()

	var err error
	puller, err = docker.NewImagePuller()
	if err != nil {
		log.Fatal("failed to create image puller: ", err)
	}
	puller.PullImage(context.Background(), conf.BuilderImage)

	knClient := knative.NewClient(conf)
	composer = core.NewComposer(knClient)
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
	s := &http.Server{Addr: "0.0.0.0:8080"}
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
	http.HandleFunc(api.HealthzPath, api.HealthCheckHandler)
	http.Handle(api.AppsPath, api.NewHandlerApps(*composer, conf))
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
