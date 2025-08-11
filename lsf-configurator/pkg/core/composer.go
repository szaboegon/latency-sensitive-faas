package core

import (
	"context"
	"fmt"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/uuid"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/apex/log"
)

const (
	MaxRetries     = 3
	WorkerPoolSize = 10
	QueueSize      = 30
)

type Composer struct {
	knClient        KnClient
	scheduler       Scheduler
	routingClient   RoutingClient
	builder         Builder
	functionAppRepo FunctionAppRepository
	fcRepo          FunctionCompositionRepository
}

func NewComposer(
	functionAppRepo FunctionAppRepository,
	fcRepo FunctionCompositionRepository,
	routingClient RoutingClient,
	knClient KnClient,
	builder Builder,
) *Composer {
	scheduler := NewScheduler(WorkerPoolSize, QueueSize)
	return &Composer{
		knClient:        knClient,
		scheduler:       scheduler,
		routingClient:   routingClient,
		builder:         builder,
		functionAppRepo: functionAppRepo,
		fcRepo:          fcRepo,
	}
}

func (c *Composer) CreateFunctionApp(
	uploadDir string,
	files []*multipart.FileHeader,
	fcs []FunctionComposition,
	appName string,
) (*FunctionApp, error) {
	id := uuid.New()
	fcApp := FunctionApp{
		Id:           id,
		Name:         appName,
		Compositions: make([]*FunctionComposition, 0),
	}

	if err := c.functionAppRepo.Save(&fcApp); err != nil {
		return nil, fmt.Errorf("could not persist function app: %w", err)
	}

	appDir := filepath.Join(uploadDir, fcApp.Id)
	err := filesystem.CreateDir(appDir)
	if err != nil {
		return nil, fmt.Errorf("could not create directory for app files: %s", err.Error())
	}

	for _, fileHeader := range files {
		err := filesystem.SaveMultiPartFile(fileHeader, appDir)
		if err != nil {
			return nil, err
		}
	}

	for _, fc := range fcs {
		fc.SourcePath = appDir
		err := c.AddFunctionComposition(fcApp.Id, &fc)

		if err != nil {
			return nil, fmt.Errorf("error while adding function compositions to app: %s", err.Error())
		}
	}

	return &fcApp, nil
}

func (c *Composer) AddFunctionComposition(appId string, fc *FunctionComposition) error {
	id := uuid.New()
	fc.Id = "fc-" + id
	fc.FunctionAppId = appId

	if err := c.fcRepo.Save(fc); err != nil {
		return fmt.Errorf("failed to update function app: %w", err)
	}

	// set the routing table in the cluster-wide store, so functions can read it on startup
	err := c.routingClient.SetRoutingTable(*fc)
	if err != nil {
		return fmt.Errorf("could not set routing table for function: %s, %s", fc.Id, err.Error())
	}

	c.scheduler.AddTask(c.buildTask(*fc), MaxRetries)
	return nil
}

func (c *Composer) DeleteFunctionApp(appId string) error {
	app, err := c.functionAppRepo.GetByID(appId)
	if err != nil || app == nil {
		return fmt.Errorf("function app with id %s does not exist", appId)
	}

	// delete all function compositions of the app
	for _, fc := range app.Compositions {
		resultChan := c.scheduler.AddTask(c.deleteTask(*fc), MaxRetries)

		r := <-resultChan
		if r.Err != nil {
			log.Errorf("Deleting function composition with id %v failed: %v", fc.Id, r.Err)
			return r.Err
		}
		log.Infof("Successfully deleted function composition with id %v", fc.Id)
		err = filesystem.DeleteDir(fc.SourcePath)
		if err != nil {
			return fmt.Errorf("could not delete app source directory: %s", err.Error())
		}
	}

	if err := c.functionAppRepo.Delete(appId); err != nil {
		return fmt.Errorf("failed to delete function app: %w", err)
	}
	return nil
}

func (c *Composer) SetRoutingTable(fcId string, table RoutingTable) error {
	fc, err := c.fcRepo.GetByID(fcId)
	if err != nil || fc == nil {
		return fmt.Errorf("function composition with id %s does not exist", fcId)
	}

	return c.setRoutingTable(fc, table)
}

// Called by HTTP handler when CI/CD pipeline notifies build is ready
func (c *Composer) NotifyBuildReady(fcId, imageURL string, status string) {
	c.builder.NotifyBuildFinished()
	fc, err := c.fcRepo.GetByID(fcId)
	if err != nil || fc == nil {
		log.Errorf("build notify failed, fc with id %s does not exist", fcId)
		return
	}
	// If the build failed, requeue the build task, do not deploy
	if strings.ToLower(status) == "failed" {
		log.Errorf("Build for function composition %s failed, requeuing...", fcId)
		c.scheduler.AddTask(c.buildTask(*fc), MaxRetries)
		return
	}

	fc.Build.Image = imageURL
	fc.Build.Timestamp = createBuildTimestamp()
	if err := c.fcRepo.Save(fc); err != nil {
		log.Errorf("Failed to save function composition with id %s: %v", fc.Id, err)
		return
	}
	log.Infof("Successfully built function composition with id %v. Image: %v", fc.Id, fc.Build.Image)

	resultChan := c.scheduler.AddTask(c.deployTask(fc.FunctionAppId, *fc), MaxRetries)
	r := <-resultChan
	if r.Err != nil {
		log.Errorf("Deploying of function composition with id %v failed: %v", fc.Id, r.Err)
		return
	}
	log.Infof("Successfully deployed function composition with id %v", fc.Id)
}

func (c *Composer) setRoutingTable(fc *FunctionComposition, table RoutingTable) error {
	fc.Components = table
	err := c.routingClient.SetRoutingTable(*fc)
	if err != nil {
		return fmt.Errorf("could not set routing table for function: %s, %s", fc.Id, err.Error())
	}

	if err := c.fcRepo.Save(fc); err != nil {
		return fmt.Errorf("failed to update function composition: %w", err)
	}

	return nil
}

func (c *Composer) buildTask(fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		buildDir, err := c.knClient.Init(context.TODO(), fc)
		c.builder.Build(context.TODO(), fc, buildDir)
		if err != nil {
			return nil, fmt.Errorf("failed to build image: %v", err)
		}
		return fc, nil
	}
}

func (c *Composer) deployTask(appId string, fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Deploy(context.TODO(), appId, fc)
	}
}

func (c *Composer) deleteTask(fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Delete(context.TODO(), fc)
	}
}

func createBuildTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
