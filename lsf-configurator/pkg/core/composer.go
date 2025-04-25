package core

import (
	"context"
	"fmt"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/uuid"
	"mime/multipart"
	"path/filepath"

	"github.com/apex/log"
)

const (
	MaxRetries     = 3
	WorkerPoolSize = 10
	QueueSize      = 30
)

type Composer struct {
	db            FunctionAppStore
	knClient      KnClient
	scheduler     Scheduler
	routingClient RoutingClient
}

func NewComposer(db FunctionAppStore, routingClient RoutingClient, knClient KnClient) *Composer {
	scheduler := NewScheduler(WorkerPoolSize, QueueSize)
	return &Composer{
		db:            db,
		knClient:      knClient,
		scheduler:     scheduler,
		routingClient: routingClient,
	}
}

func (c *Composer) CreateFunctionApp(uploadDir string, files []*multipart.FileHeader, fcs []FunctionComposition) (*FunctionApp, error) {
	id := uuid.New()
	fcApp := FunctionApp{
		Id:           id,
		Compositions: make(map[string]*FunctionComposition),
	}

	c.db.Set(id, fcApp)
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
		err := c.AddFunctionComposition(fcApp.Id, fc)

		if err != nil {
			return nil, fmt.Errorf("error while adding function compositions to app: %s", err.Error())
		}
	}

	return &fcApp, nil
}

func (c *Composer) AddFunctionComposition(appId string, fc FunctionComposition) error {
	app, err := c.db.Get(appId)
	if err != nil {
		return fmt.Errorf("app with id %s not found", appId)
	}

	id := UniqueFcId(app.Id, fc.Name)
	// function composition is already part of the application
	if _, ok := app.Compositions[id]; ok {
		return nil
	}

	fc.Id = id
	app.Compositions[id] = &fc
	c.db.Set(appId, app)

	// set the routing table in the cluster-wide store, so functions can read it on startup
	err = c.routingClient.SetRoutingTable(appId, fc)
	if err != nil {
		return fmt.Errorf("could not set routing table for function: %s, %s", fc.Id, err.Error())
	}

	go c.scheduleBuildAndDeploy(fc)
	return nil
}

func (c *Composer) DeleteFunctionApp(appId string) error {
	app, err := c.db.Get(appId)
	if err != nil {
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

	c.db.Delete(appId)
	return nil
}

func (c *Composer) SetRoutingTable(appId string, table RoutingTable) error {
	app, err := c.db.Get(appId)
	if err != nil {
		return fmt.Errorf("function app with id %s does not exist", appId)
	}

	//rewrite this, because routing table is part of function composition now, not function app
	//TODO implement routing table setting in the cluster-wide store
	//err = c.routingClient.SetRoutingTable(appId, table)
	c.db.Set(appId, app)
	return nil
}

func (c *Composer) scheduleBuildAndDeploy(fc FunctionComposition) {
	resultChan := c.scheduler.AddTask(c.buildTask(fc), MaxRetries)

	r := <-resultChan
	if r.Err != nil {
		log.Errorf("Build of function composition with id %v failed: %v", fc.Id, r.Err)
		return
	}
	fc = r.Value.(FunctionComposition)
	log.Infof("Successfully built function composition with id %v. Image: %v", fc.Id, fc.Build.Image)

	resultChan = c.scheduler.AddTask(c.deployTask(fc), MaxRetries)
	r = <-resultChan
	if r.Err != nil {
		log.Errorf("Deploying of function composition with id %v failed: %v", fc.Id, r.Err)
		return
	}
	log.Infof("Successfully deployed function composition with id %v", fc.Id)
}

func (c *Composer) buildTask(fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		fc, err := c.knClient.Build(context.TODO(), fc)
		if err != nil {
			return nil, err
		}
		return fc, err
	}
}

func (c *Composer) deployTask(fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Deploy(context.TODO(), fc)
	}
}

func (c *Composer) deleteTask(fc FunctionComposition) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Delete(context.TODO(), fc)
	}
}

func UniqueFcId(appId, funcName string) string {
	return "app-" + appId + "-" + funcName
}
