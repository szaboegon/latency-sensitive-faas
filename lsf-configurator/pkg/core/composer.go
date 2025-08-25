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

// for now, only python is implemented
var runtimeExtensions = map[string]string{
	"python": ".py", // Python
}

type Composer struct {
	knClient        KnClient
	scheduler       Scheduler
	routingClient   RoutingClient
	builder         Builder
	functionAppRepo FunctionAppRepository
	fcRepo          FunctionCompositionRepository
	deploymentRepo  DeploymentRepository
}

func NewComposer(
	functionAppRepo FunctionAppRepository,
	fcRepo FunctionCompositionRepository,
	deploymentRepo DeploymentRepository,
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
		deploymentRepo:  deploymentRepo,
	}
}

func (c *Composer) GetFunctionApp(appId string) (*FunctionApp, error) {
	return c.functionAppRepo.GetByID(appId)
}

func (c *Composer) ListFunctionApps() ([]*FunctionApp, error) {
	return c.functionAppRepo.GetAll()
}

func (c *Composer) CreateFunctionApp(
	uploadDir string,
	files []*multipart.FileHeader,
	appName string,
	runtime string,
) (*FunctionApp, error) {
	id := uuid.New()
	fcApp := FunctionApp{
		Id:           id,
		Name:         appName,
		Compositions: make([]*FunctionComposition, 0),
		Files:        make([]string, 0),
		SourcePath:   "",
		Components:   make([]Component, 0),
		Runtime:      strings.ToLower(runtime),
	}

	appDir := filepath.Join(uploadDir, fcApp.Id)
	err := filesystem.CreateDir(appDir)
	if err != nil {
		return nil, fmt.Errorf("could not create directory for app files: %s", err.Error())
	}
	fcApp.SourcePath = appDir

	for _, fileHeader := range files {
		fileName := fileHeader.Filename
		if isComponent(fileName, fcApp.Runtime) {
			componentName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			fcApp.Components = append(fcApp.Components, Component(componentName))
		} else {
			fcApp.Files = append(fcApp.Files, fileName)
		}
		err := filesystem.SaveMultiPartFile(fileHeader, appDir)
		if err != nil {
			return nil, err
		}
	}

	if err := c.functionAppRepo.Save(&fcApp); err != nil {
		return nil, fmt.Errorf("could not persist function app: %w", err)
	}

	return &fcApp, nil
}

func isComponent(fileName string, runtime string) bool {
	extension := filepath.Ext(fileName)
	return runtimeExtensions[runtime] == extension
}

func (c *Composer) AddFunctionComposition(appId string, components []string, files []string) (*FunctionComposition, error) {
	fcApp, err := c.functionAppRepo.GetByID(appId)
	if err != nil || fcApp == nil {
		log.Errorf("function app with id %s does not exist", appId)
		return nil, fmt.Errorf("function app with id %s does not exist", appId)
	}

	id := uuid.New()
	convertedComps := make([]Component, 0)
	for _, comp := range components {
		c := Component(comp)
		convertedComps = append(convertedComps, c)
	}

	fc := &FunctionComposition{
		Id:            "fc-" + id,
		FunctionAppId: appId,
		Components:    convertedComps,
		Files:         files,
		Status:        BuildStatusPending,
	}

	if err := c.fcRepo.Save(fc); err != nil {
		return nil, fmt.Errorf("failed to update function app: %w", err)
	}

	c.scheduler.AddTask(c.buildTask(*fc, fcApp.Runtime, fcApp.SourcePath), MaxRetries)
	return fc, nil
}

func (c *Composer) CreateFcDeployment(fcId, namespace, node string, routingTable RoutingTable) (*Deployment, error) {
	fc, err := c.fcRepo.GetByID(fcId)
	if err != nil || fc == nil {
		return nil, fmt.Errorf("function composition with id %s does not exist", fcId)
	}

	deployment := Deployment{
		Id:                    "d-" + uuid.New(),
		FunctionCompositionId: fcId,
		Namespace:             namespace,
		Node:                  node,
		RoutingTable:          routingTable,
	}

	if fc.Status == BuildStatusBuilt {
		deployment.Status = DeploymentStatusPending
		c.startDeployment(&deployment, fc)
	} else {
		deployment.Status = DeploymentStatusWaitingForBuild
		log.Infof("Function composition with id %s is not built yet, deployment will be started after build is ready", fcId)
	}

	err = c.deploymentRepo.Save(&deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to save deployment: %w", err)
	}
	return &deployment, nil
}

func (c *Composer) DeleteFunctionApp(appId string) error {
	app, err := c.functionAppRepo.GetByID(appId)
	if err != nil || app == nil {
		return fmt.Errorf("function app with id %s does not exist", appId)
	}

	// delete all deployments of the app
	deployments, err := c.deploymentRepo.GetByFunctionAppID(appId)
	if err != nil {
		return fmt.Errorf("failed to get deployments for function app %s: %w", appId, err)
	}
	for _, d := range deployments {
		resultChan := c.scheduler.AddTask(c.deleteTask(*d), MaxRetries)

		r := <-resultChan
		if r.Err != nil {
			log.Errorf("Deleting deployment with id %v failed: %v", d.Id, r.Err)
			return r.Err
		}
		log.Infof("Successfully deleted deployment with id %v", d.Id)
		err = filesystem.DeleteDir(app.SourcePath)
		if err != nil {
			return fmt.Errorf("could not delete app source directory: %s", err.Error())
		}
	}

	if err := c.functionAppRepo.Delete(appId); err != nil {
		return fmt.Errorf("failed to delete function app: %w", err)
	}
	return nil
}

func (c *Composer) DeleteFunctionComposition(fcId string) error {
	fc, err := c.fcRepo.GetByID(fcId)
	if err != nil || fc == nil {
		return fmt.Errorf("function composition with id %s does not exist", fcId)
	}

	go func() {
		for _, deployment := range fc.Deployments {
			resultChan := c.scheduler.AddTask(c.deleteTask(*deployment), MaxRetries)
			r := <-resultChan
			if r.Err != nil {
				log.Errorf("Deleting of deployment with id %v failed: %v", deployment.Id, r.Err)
				fc.Status = BuildStatusError
				c.fcRepo.Save(fc)
				return
			}
			log.Infof("Successfully deleted deployment with id %v", deployment.Id)
		}
		log.Infof("Successfully deleted all deployments for function composition with id %v", fc.Id)
		if err := c.fcRepo.Delete(fcId); err != nil {
			log.Errorf("failed to delete function composition: %w", err)
		}
	}()

	return nil
}

func (c *Composer) SetRoutingTable(deploymentId string, table RoutingTable) error {
	deployment, err := c.deploymentRepo.GetByID(deploymentId)
	if err != nil || deployment == nil {
		return fmt.Errorf("deployment with id %s does not exist", deploymentId)
	}

	err = c.setRoutingTable(deployment, table)
	if err != nil {
		log.Errorf("failed to set routing table for deployment %s: %v", deploymentId, err)
		return err
	}
	return nil
}

func (c *Composer) NotifyBuildReady(fcId, imageURL string, status string) {
	c.builder.NotifyBuildFinished()
	fc, err := c.fcRepo.GetByID(fcId)
	if err != nil || fc == nil {
		log.Errorf("build notify failed, fc with id %s does not exist", fcId)
		return
	}

	fcApp, err := c.functionAppRepo.GetByID(fc.FunctionAppId)
	if err != nil || fcApp == nil {
		log.Errorf("function app with id %s does not exist", fc.FunctionAppId)
		return
	}

	// If the build failed, set status to failed
	if strings.ToLower(status) == "failed" {
		log.Errorf("Build for function composition %s failed", fcId)
		fc.Status = BuildStatusError
		if err := c.fcRepo.Save(fc); err != nil {
			log.Errorf("Failed to save function composition with id %s: %v", fc.Id, err)
		}
		return
	}

	fc.Build.Image = imageURL
	fc.Build.Timestamp = createBuildTimestamp()
	fc.Status = BuildStatusBuilt
	if err := c.fcRepo.Save(fc); err != nil {
		log.Errorf("Failed to save function composition with id %s: %v", fc.Id, err)
		return
	}
	log.Infof("Successfully built function composition with id %v. Image: %v", fc.Id, fc.Build.Image)

	// Trigger any pending deployments
	deployments, err := c.deploymentRepo.GetByFunctionCompositionID(fc.Id)
	if err != nil {
		log.Errorf("Failed to get deployments for function composition with id %s: %v", fc.Id, err)
		return
	}

	for _, deployment := range deployments {
		if deployment.Status == DeploymentStatusWaitingForBuild {
			c.startDeployment(deployment, fc)
		}
	}
}

func (c *Composer) RollbackBulk(
	app *FunctionApp,
	compositions []*FunctionComposition,
	deployments []*Deployment,
) {
	for _, dep := range deployments {
		if err := c.deploymentRepo.Delete(dep.Id); err != nil {
			log.Errorf("Failed to rollback deployment %s: %v", dep.Id, err)
		}
	}

	for _, fc := range compositions {
		if err := c.fcRepo.Delete(fc.Id); err != nil {
			log.Errorf("Failed to rollback function composition %s: %v", fc.Id, err)
		}
	}

	if app != nil {
		if err := c.functionAppRepo.Delete(app.Id); err != nil {
			log.Errorf("Failed to rollback function app %s: %v", app.Id, err)
		}
	}
}

func (c *Composer) setRoutingTable(deployment *Deployment, table RoutingTable) error {
	deployment.RoutingTable = table
	err := c.routingClient.SetRoutingTable(*deployment)
	if err != nil {
		return fmt.Errorf("could not set routing table for deployment: %s, %s", deployment.Id, err.Error())
	}

	if err := c.deploymentRepo.Save(deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

func (c *Composer) startDeployment(deployment *Deployment, fc *FunctionComposition) {
	resultChan := c.scheduler.AddTask(c.deployTask(*deployment, fc.Build.Image, fc.FunctionAppId), MaxRetries)
	r := <-resultChan
	if r.Err != nil {
		log.Errorf("Deploying of function composition with id %v and deploymentId %v failed: %v, ", fc.Id, deployment.Id, r.Err)
		deployment.Status = DeploymentStatusError
		c.deploymentRepo.Save(deployment)
		return
	}
	log.Infof("Successfully deployed function composition with id %v, deploymentId %v", fc.Id, deployment.Id)
	deployment.Status = DeploymentStatusDeployed
	c.deploymentRepo.Save(deployment)
}

func (c *Composer) buildTask(fc FunctionComposition, runtime, sourcePath string) func() (interface{}, error) {
	return func() (interface{}, error) {
		buildDir, err := c.knClient.Init(context.TODO(), fc, runtime, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to init build: %v", err)
		}
		err = c.builder.Build(context.TODO(), fc, buildDir)
		if err != nil {
			return nil, fmt.Errorf("failed to build image: %v", err)
		}
		return fc, nil
	}
}

func (c *Composer) deployTask(deployment Deployment, image, appId string) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Deploy(context.TODO(), deployment, image, appId)
	}
}

func (c *Composer) deleteTask(deployment Deployment) func() (interface{}, error) {
	return func() (interface{}, error) {
		return nil, c.knClient.Delete(context.TODO(), deployment)
	}
}

func createBuildTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
