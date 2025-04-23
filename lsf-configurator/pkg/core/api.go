package core

import "context"

type FunctionApp struct {
	Id           string
	Compositions map[string]*FunctionComposition
}

type Component string

type Route struct {
	To       string `json:"to"`
	Function string `json:"function"`
}

type RoutingTable map[Component][]Route

type FunctionComposition struct {
	Id         string       `json:"-"`
	Name       string       `json:"name"`
	Node       string       `json:"node,omitempty"`
	Components RoutingTable `json:"components"`
	NameSpace  string       `json:"namespace"`
	SourcePath string       `json:"-"`
	Runtime    string       `json:"runtime"`
	Build
}

type Build struct {
	Image string `json:"-"`
	Stamp string `json:"-"`
}

type KnClient interface {
	Build(ctx context.Context, fc FunctionComposition) (FunctionComposition, error)
	Deploy(ctx context.Context, fc FunctionComposition) error
}

type FunctionAppStore interface {
	Set(id string, app FunctionApp)
	Get(id string) (FunctionApp, error)
	Delete(id string)
}

type RoutingClient interface {
	SetRoutingTable(appId string, fc FunctionComposition) error
}
