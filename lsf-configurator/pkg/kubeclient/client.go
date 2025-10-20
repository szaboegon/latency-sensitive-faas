package kubeclient

import (
	"fmt"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	"k8s.io/client-go/rest"
)

type Client struct {
	istio istioclient.Interface
}

func NewKubeclient() (*Client, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}

	ic, err := istioclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Istio client: %w", err)
	}

	return &Client{istio: ic}, nil
}
