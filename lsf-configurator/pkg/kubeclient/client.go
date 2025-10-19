package kubeclient

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client struct {
	client client.Client
	scheme *runtime.Scheme
}

func NewKubeclient() (*Client, error) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)

	var cfg *rest.Config
	var err error

	if _, err = os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = config.GetConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load kube config: %w", err)
	}

	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client: %w", err)
	}

	return &Client{client: cl, scheme: scheme}, nil
}

func (c *Client) GetControllerClient() client.Client {
	return c.client
}
