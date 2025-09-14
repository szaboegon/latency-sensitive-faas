//TODO introduce sqlite
// https://chatgpt.com/share/688a4a14-4744-8009-aaad-be0ba6e82700
// add method which allows updating the fc configurations in a function app
// in this case, newly added function compositions should be added to the existing ones
// if more function compositions use the same components, they should be merged/or the image should be reused
// the deployment state should always be tracked, to know which ones are currently in use

package core

import (
	"context"
)

type KnClient interface {
	Init(ctx context.Context, fc FunctionComposition, runtime, sourcePath string) (string, error)
	Deploy(ctx context.Context, deployment Deployment, image, appId string) error
	Delete(ctx context.Context, deployment Deployment) error
}

type RoutingClient interface {
	SetRoutingTable(deployment Deployment) error
}

type Builder interface {
	Build(ctx context.Context, fc FunctionComposition, buildDir string) error
	NotifyBuildFinished()
}

type AlertClient interface {
	EnsureAlertConnector(ctx context.Context, connectorName, indexName string) (string, error)
	CreateOrUpdateRule(ctx context.Context, serviceName string, latencyThresholdMs int) (string, error)
}

type MetricsReader interface {
	QueryNodeMetrics() ([]NodeMetrics, error)
	QueryAverageAppRuntime(appId string) (float64, error)
	EnsureIndex(ctx context.Context, indexName string) error
}
