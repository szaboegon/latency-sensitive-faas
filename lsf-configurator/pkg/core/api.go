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
	Deploy(ctx context.Context, fc FunctionComposition) error
	Delete(ctx context.Context, fc FunctionComposition) error
}

type FunctionAppStore interface {
	Set(id string, app FunctionApp)
	Get(id string) (FunctionApp, error)
	Delete(id string)
}

type RoutingClient interface {
	SetRoutingTable(fc FunctionComposition) error
}

type Builder interface {
	Build(ctx context.Context, fc FunctionComposition, buildDir string) error
	NotifyBuildFinished()
}
