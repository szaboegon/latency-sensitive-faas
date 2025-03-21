package docker

import (
	"context"

	"github.com/docker/docker/api/types/image"
	dclient "github.com/docker/docker/client"
	knclient "knative.dev/func/pkg/docker"
)

type ImagePuller interface {
	PullImage(ctx context.Context, img string) error
}

type DockerClient struct {
	cli dclient.CommonAPIClient
}

func NewImagePuller() (ImagePuller, error) {
	cli, _, err := knclient.NewClient(dclient.DefaultDockerHost)
	if err != nil {
		return nil, err
	}

	return &DockerClient{
		cli: cli,
	}, nil
}

func (dc *DockerClient) PullImage(ctx context.Context, img string) error {
	readCloser, err := dc.cli.ImagePull(ctx, img, image.PullOptions{})
	if err != nil {
		return err
	}
	defer readCloser.Close()
	return nil
}
