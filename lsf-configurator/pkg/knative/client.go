package knative

import (
	"context"
	"fmt"
	"log"
	"lsf-configurator/pkg/bootstrapping"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/uuid"
	"path"

	http "net/http"

	builders "knative.dev/func/pkg/builders"
	pack "knative.dev/func/pkg/builders/buildpacks"
	docker "knative.dev/func/pkg/docker"
	fn "knative.dev/func/pkg/functions"
	knativefunc "knative.dev/func/pkg/knative"
)

const CompositionTemplateName = "composition"

//var DefaultScaleMetric string = "concurrency"

type Client struct {
	fnClient      *fn.Client
	imageRegistry string
	builderImages map[string]string
}

func NewClient(conf config.Configuration) *Client {
	o := []fn.Option{fn.WithRepository(conf.TemplatesPath)}
	c := NewCredentialsProvider(conf.RegistryUser, conf.RegistryPassword)

	o = append(o,
		fn.WithBuilder(pack.NewBuilder(
			pack.WithName(builders.Pack),
			pack.WithTimestamp(true),
			pack.WithVerbose(true))))

	o = append(o,
		fn.WithPusher(docker.NewPusher(
			docker.WithCredentialsProvider(c),
			docker.WithTransport(http.DefaultTransport),
			docker.WithVerbose(true))))

	o = append(o,
		fn.WithDeployer(knativefunc.NewDeployer(
			knativefunc.WithDeployerVerbose(true))))

	fnClient := fn.New(o...)

	return &Client{
		fnClient:      fnClient,
		imageRegistry: conf.ImageRegistry,
		builderImages: map[string]string{
			"pack": conf.BuilderImage,
		},
	}
}

func (c *Client) Build(ctx context.Context, fc core.FunctionComposition) (core.FunctionComposition, error) {
	buildDir := createBuildDir(fc.SourcePath)
	defer cleanUpBuildDir(buildDir)

	f := fn.Function{
		Name:      fc.Id,
		Namespace: fc.NameSpace,
		Runtime:   fc.Runtime,
		Registry:  c.imageRegistry,
		Invoke:    "http",
		Root:      buildDir,
		Template:  CompositionTemplateName,
		Build: fn.BuildSpec{
			Builder: "pack",
			BuilderImages: map[string]string{
				"pack": c.builderImages["pack"],
			},
		},
	}

	f, err := c.fnClient.Init(f)
	if err != nil {
		log.Fatalf("Could not initialize function based on config: %v", err)
	}

	bootstrapper, err := bootstrapping.NewBootstrapper(fc, buildDir)
	if err != nil {
		return core.FunctionComposition{}, err
	}

	err = bootstrapper.Setup()
	if err != nil {
		return core.FunctionComposition{}, err
	}

	f, err = c.fnClient.Build(ctx, f)
	if err != nil {
		return core.FunctionComposition{}, err
	}

	f, success, err := c.fnClient.Push(ctx, f) //TODO does not seem to push to my dockerhub :(
	if err != nil {
		return core.FunctionComposition{}, err
	}

	if !success {
		return core.FunctionComposition{}, fmt.Errorf("failed to push image %v to registry %v", f.Build.Image, f.Registry)
	}

	fc.Build.Image = f.Build.Image
	fc.Build.Stamp = f.BuildStamp()
	return fc, nil
}

func (c *Client) Deploy(ctx context.Context, fc core.FunctionComposition) error {
	f := fn.Function{ //TODO fix function is not built error
		Name:      fc.Id,
		Namespace: fc.NameSpace,
		Runtime:   fc.Runtime,
		Image:     fc.Image,
		Deploy: fn.DeploySpec{
			Image: fc.Image,
			NodeAffinity: fn.NodeAffinity{
				RequiredNodes: []string{fc.Node},
			},
			Namespace: fc.NameSpace,
		},
	}

	if fc.Image == "" {
		return fmt.Errorf("cannot deploy function %v, because it has no corresponding image", fc.Id)
	}

	_, err := c.fnClient.Deploy(ctx, f, fn.WithDeploySkipBuildCheck(true))
	return err
}

func createBuildDir(sourcePath string) string {
	tempDir := path.Join(sourcePath, "temp", uuid.New())
	filesystem.CreateDir(tempDir)

	return tempDir
}

func cleanUpBuildDir(path string) error {
	return filesystem.DeleteDir(path)
}
