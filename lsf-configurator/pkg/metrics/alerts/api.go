package alerts

import "context"

type AlertClient interface {
	EnsureAlertConnector(ctx context.Context, connectorName, indexName string) (string, error)
	CreateAlert(ctx context.Context, serviceName string, latencyThresholdMs int, connectorId string) (string, error)
}
