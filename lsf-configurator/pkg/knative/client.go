package knative

import (
	"context"
	"fmt"
	"lsf-configurator/pkg/bootstrapping"
	"lsf-configurator/pkg/config"
	"lsf-configurator/pkg/core"
	"lsf-configurator/pkg/filesystem"
	"lsf-configurator/pkg/uuid"
	"path"

	fn "knative.dev/func/pkg/functions"
	knativefunc "knative.dev/func/pkg/knative"
)

const CompositionTemplateName = "composition"

//var DefaultScaleMetric string = "concurrency"

type Client struct {
	fnClient      *fn.Client
	imageRegistry string
}

func NewClient(conf config.Configuration) *Client {
	o := []fn.Option{fn.WithRepository(conf.TemplatesPath)}

	o = append(o,
		fn.WithDeployer(knativefunc.NewDeployer(
			knativefunc.WithDeployerVerbose(conf.VerboseLogs))))

	fnClient := fn.New(o...)

	return &Client{
		fnClient:      fnClient,
		imageRegistry: conf.ImageRegistry + "/" + conf.ImageRepository,
	}
}

func (c *Client) Init(ctx context.Context, fc core.FunctionComposition, runtime, sourcePath string) (string, error) {
	buildDir, err := createBuildDir(sourcePath)
	if err != nil {
		return "", fmt.Errorf("could not create build directory: %v", err)
	}

	f := fn.Function{
		Name:     fc.Id,
		Runtime:  runtime,
		Registry: c.imageRegistry,
		Invoke:   "http",
		Root:     buildDir,
		Template: CompositionTemplateName,
	}

	_, err = c.fnClient.Init(f)
	if err != nil {
		return "", fmt.Errorf("could not initialize function based on config: %v", err)
	}

	err = copyNonSourceFiles(sourcePath, buildDir, fc.Files)
	if err != nil {
		return "", fmt.Errorf("failed to copy non-source files: %v", err)
	}

	bootstrapper, err := bootstrapping.NewBootstrapper(runtime, fc, buildDir, sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to create bootstrapper: %v", err)
	}

	err = bootstrapper.Setup()
	if err != nil {
		return "", fmt.Errorf("failed to setup bootstrapper: %v", err)
	}

	return buildDir, nil
}

func (c *Client) Deploy(ctx context.Context, deployment core.Deployment, image, appId string) error {
	//TODO calculate replica count based on edge invocation rate, and add it to deployment properties
	minReplicas := int64(1)
	f := fn.Function{
		Name:      deployment.Id,
		Namespace: deployment.Namespace,
		Image:     image,
		Deploy: fn.DeploySpec{
			Image: image,
			NodeAffinity: fn.NodeAffinity{
				RequiredNodes: []string{deployment.Node},
			},
			Namespace: deployment.Namespace,
			Options: fn.Options{
				Scale: &fn.ScaleOptions{
					Min: &minReplicas, 
				},
				//ScaleMetric: &DefaultScaleMetric,
			},
		},
		Run: fn.RunSpec{
			Envs: getDeployEnvs(appId, deployment.Id),
		},
	}

	if f.Image == "" {
		return fmt.Errorf("cannot deploy function deployment %v, because it has no corresponding image", deployment.Id)
	}

	_, err := c.fnClient.Deploy(ctx, f, fn.WithDeploySkipBuildCheck(true))
	return err
}

func (c *Client) Delete(ctx context.Context, deployment core.Deployment) error {
	f := fn.Function{
		Name:      deployment.Id,
		Namespace: deployment.Namespace,
		Deploy: fn.DeploySpec{
			Namespace: deployment.Namespace,
		},
	}

	err := c.fnClient.Remove(ctx, deployment.Id, deployment.Namespace, f, true)
	if err != nil {
		return err
	}
	return nil
}

func createBuildDir(sourcePath string) (string, error) {
	tempDir := path.Join(sourcePath, "temp", uuid.New())
	err := filesystem.CreateDir(tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to create build directory: %v", err)
	}
	return tempDir, nil
}

func copyNonSourceFiles(sourcePath, buildDir string, fileNames []string) error {
	_, err := filesystem.CopyFilesByNames(sourcePath, buildDir, fileNames, false)
	return err
}

func getDeployEnvs(appId, deploymentId string) []fn.Env {
	envFuncName := "FUNCTION_NAME"
	envFuncNameValue := deploymentId

	envAppName := "APP_NAME"
	envAppNameValue := appId

	envs := make([]fn.Env, 0)
	envs = append(envs, fn.Env{Name: &envFuncName, Value: &envFuncNameValue})
	envs = append(envs, fn.Env{Name: &envAppName, Value: &envAppNameValue})

	return envs
}
