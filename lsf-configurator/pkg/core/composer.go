package core

import (
	"context"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/google/uuid"
)

type KnClient interface {
	Build(ctx context.Context, fc FunctionComposition) (FunctionComposition, error)
	Deploy(ctx context.Context, fc FunctionComposition) error
}

type Composer struct {
	Apps      map[string]*FunctionApp
	knClient  KnClient
	scheduler Scheduler
}

func NewComposer(knClient KnClient) *Composer {
	scheduler := NewScheduler(100, 200)
	return &Composer{
		Apps:      make(map[string]*FunctionApp),
		knClient:  knClient,
		scheduler: scheduler,
	}
}

func (c *Composer) CreateFunctionApp() *FunctionApp {
	id := uuid.New().String()
	fcApp := FunctionApp{
		Id:           id,
		Compositions: make(map[string]*FunctionComposition),
	}

	c.Apps[id] = &fcApp
	return &fcApp
}

func (c *Composer) AddFunctionComposition(appId string, fc FunctionComposition) {
	app, ok := c.Apps[appId]
	if !ok {
		log.Error("App with id not found")
		return
	}

	var compNames []string
	for _, comp := range fc.Components {
		compNames = append(compNames, comp.Name)
	}
	id := uniqueId(app.Id, compNames)
	_, ok = app.Compositions[id]
	if ok {
		return
	}

	fc.Id = id
	app.Compositions[id] = &fc
	c.Apps[appId] = app

	go c.scheduleBuildAndDeploy(fc)
}

func (c *Composer) scheduleBuildAndDeploy(fc FunctionComposition) {
	resultChan := c.scheduler.AddTask(c.buildTask(fc))

	r := <-resultChan
	if r.Err != nil {
		log.Errorf("Build of function composition with id %v failed: %v", fc.Id, r.Err)
		return
	}
	fc = r.Value.(FunctionComposition)
	log.Infof("Successfully built function composition with id %v. Image: %v", fc.Id, fc.Build.Image)

	resultChan = c.scheduler.AddTask(c.deployTask(fc))
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

func uniqueId(appId string, compNames []string) string {
	sort.Strings(compNames)
	compId := strings.Join(compNames, "-")

	return compId + "-" + appId
}
