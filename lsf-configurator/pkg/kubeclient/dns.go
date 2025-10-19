package kubeclient

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *Client) EnsureDNSRecord(ctx context.Context, namespace, appName, targetServiceName string) error {
	serviceName := generateServiceName(appName)
	target := fmt.Sprintf("%s.%s.svc.cluster.local", targetServiceName, namespace)

	svc := &corev1.Service{}
	err := c.client.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: namespace}, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create new ExternalName service
			newSvc := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceName,
					Namespace: namespace,
					Labels: map[string]string{
						"app":  appName,
						"type": "faas-dns",
					},
				},
				Spec: corev1.ServiceSpec{
					Type:         corev1.ServiceTypeExternalName,
					ExternalName: target,
				},
			}
			if err := c.client.Create(ctx, newSvc); err != nil {
				return fmt.Errorf("failed to create persistent dns service: %w", err)
			}
			fmt.Printf("[dns] Created new persistent service %q → %q\n", serviceName, target)
			return nil
		}
		return fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	// Update existing one if target changed
	if svc.Spec.ExternalName != target {
		svc.Spec.ExternalName = target
		if err := c.client.Update(ctx, svc); err != nil {
			return fmt.Errorf("failed to update persistent dns service: %w", err)
		}
		fmt.Printf("[dns] Updated persistent service %q → %q\n", serviceName, target)
	}

	return nil
}

func (c *Client) DeleteDNSRecord(ctx context.Context, namespace, appName string) error {
	serviceName := generateServiceName(appName)

	svc := &corev1.Service{}
	err := c.client.Get(ctx, client.ObjectKey{Name: serviceName, Namespace: namespace}, svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get service %s: %w", serviceName, err)
	}

	if err := c.client.Delete(ctx, svc); err != nil {
		return fmt.Errorf("failed to delete persistent dns service: %w", err)
	}
	fmt.Printf("[dns] Deleted persistent service %q\n", serviceName)

	return nil
}

func generateServiceName(appName string) string {
	return fmt.Sprintf("%s-dns", appName)
}
