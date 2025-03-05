package knative

import (
	"context"
	"log"
	"lsf-configurator/pkg/bootstrapping"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"path"

	"github.com/google/uuid"
	fn "knative.dev/func/pkg/functions"
)

const CompositionTemplateName = "composition"

type Client struct {
	fnClient      *fn.Client
	imageRegistry string
}

func NewClient(templateRepo, imageRegistry string) *Client {
	fnClient := fn.New(fn.WithRepository(templateRepo))

	return &Client{
		fnClient:      fnClient,
		imageRegistry: imageRegistry,
	}
}

func (c *Client) Build(ctx context.Context, fc core.FunctionComposition) (string, error) {
	buildDir := createBuildDir(fc.SourcePath)

	f := fn.Function{
		Name:      fc.Id,
		Namespace: "application",
		Runtime:   fc.Runtime,
		Registry:  c.imageRegistry,
		Invoke:    "http",
		Build: fn.BuildSpec{
			Builder: "pack",
		},
		Root:     buildDir,
		Template: CompositionTemplateName,
	}

	_, err := c.fnClient.Init(f)
	if err != nil {
		log.Fatalf("Could not initialize function based on config: %v", err)
	}

	bootstrapper, err := bootstrapping.NewBootstrapper(fc, buildDir)
	if err != nil {
		return "", err
	}

	err = bootstrapper.Setup()
	if err != nil {
		return "", err
	}

	f, err = c.fnClient.Build(ctx, f)
	cleanUpBuildDir(buildDir)

	return f.Build.Image, err
}

func (c *Client) Deploy(ctx context.Context, fc core.FunctionComposition) error {
	return nil
}

func createBuildDir(sourcePath string) string {
	tempDir := path.Join(sourcePath, "temp", uuid.New().String())
	filesystem.CreateDir(tempDir)

	return tempDir
}

func cleanUpBuildDir(path string) error {
	return filesystem.DeleteDir(path)
}
