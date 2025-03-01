package knative

import (
	"context"
	"log"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"path"

	"github.com/google/uuid"
	fn "knative.dev/func/pkg/functions"
)

type Client struct {
	fnClient fn.Client
}

func NewClient() *Client {
	return &Client{
		fnClient: *fn.New(),
	}
}

func (c *Client) Build(ctx context.Context, fc core.FunctionComposition) (string, error) {
	buildDir := createBuildDir(fc)

	f := fn.Function{
		Name:      fc.Id,
		Namespace: "application",
		Runtime:   "python",
		Registry:  "registry.hub.docker.com/szaboegon",
		Invoke:    "http",
		Build: fn.BuildSpec{
			Builder: "pack",
		},
		Root:     buildDir,
		Template: "http", //TODO change to own template called "composition"
	}

	f, err := c.fnClient.Init(f)
	if err != nil {
		log.Fatalf("Could not initialize function based on config: %v", err)
	}

	for _, comp := range fc.Components {
		filesystem.CopyFileToDstFolder(path.Join(fc.SourcePath, comp.Name+".py"), buildDir)
	}

	//TODO write a script that extracts handlers from the files and imports them into the main file

	//TODO the script should also discover all dependencies in these components and add them to the main requirements.txt file

	f, err = c.fnClient.Build(ctx, f)

	cleanUpBuildDir(buildDir)

	return f.Image, err
}

func (c *Client) Deploy(ctx context.Context, fc core.FunctionComposition) error {
	return nil
}

func createBuildDir(fc core.FunctionComposition) string {
	tempDir := path.Join(fc.SourcePath, "temp", uuid.New().String())
	filesystem.CreateDir(tempDir)

	return tempDir
}

func cleanUpBuildDir(path string) error {
	return filesystem.DeleteDir(path)
}
