package builder

import (
	"context"
	"fmt"
	"log"
	"lsf-configurator/pkg/core"
	"os"
	"os/user"
	"path/filepath"

	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type TektonBuilder struct {
	Namespace    string
	Pipeline     string
	NotifyURL    string
	WorkspacePVC string
	ImageRepo    string
}

func NewTektonBuilder() *TektonBuilder {
	// These values could be loaded from env/config as needed
	return &TektonBuilder{
		Namespace:    getEnv("TEKTON_NAMESPACE", "tekton"),
		Pipeline:     getEnv("TEKTON_PIPELINE", "function-build-pipeline"),
		NotifyURL:    getEnv("TEKTON_NOTIFY_URL", "http://lsf-configurator.lsf-configurator.svc.cluster.local:8080/apps/build_notify"),
		WorkspacePVC: getEnv("TEKTON_WORKSPACE_PVC", "tekton-pvc"),
		ImageRepo:    getEnv("TEKTON_IMAGE_REPO", "registry.hub.docker.com/szaboegon"),
	}
}

func (b *TektonBuilder) Build(ctx context.Context, fc *core.FunctionComposition) error {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		// Fallback to local kubeconfig if not running in cluster
		usr, _ := user.Current()
		kubeconfig := filepath.Join(usr.HomeDir, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			log.Printf("Error getting kube config: %v", err)
			return fmt.Errorf("failed to get kube config: %w", err)
		}
	}

	tektonClient, err := tektonclientset.NewForConfig(config)
	if err != nil {
		log.Printf("Error creating Tekton client: %v", err)
		return fmt.Errorf("failed to create tekton client: %w", err)
	}

	image := fmt.Sprintf("%s/%s:latest", b.ImageRepo, fc.Id)
	prName := fmt.Sprintf("build-%s", fc.Id)

	pr := &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      prName,
			Namespace: b.Namespace,
		},
		Spec: tektonv1.PipelineRunSpec{
			PipelineRef: &tektonv1.PipelineRef{
				Name: b.Pipeline,
			},
			Params: []tektonv1.Param{
				{Name: "IMAGE", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: image}},
				{Name: "CONTEXT_DIR", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: fc.SourcePath}},
				{Name: "NOTIFY_URL", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: b.NotifyURL}},
				{Name: "FC_ID", Value: tektonv1.ParamValue{Type: tektonv1.ParamTypeString, StringVal: fc.Id}},
			},
			Workspaces: []tektonv1.WorkspaceBinding{
				{
					Name: "source",
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: b.WorkspacePVC,
					},
				},
			},
		},
	}

	_, err = tektonClient.TektonV1().PipelineRuns(b.Namespace).Create(ctx, pr, metav1.CreateOptions{})
	if err != nil {
		log.Printf("Error creating PipelineRun: %v", err)
		return fmt.Errorf("failed to create PipelineRun: %w", err)
	}
	log.Printf("PipelineRun %s created successfully in namespace %s", prName, b.Namespace)
	return nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
