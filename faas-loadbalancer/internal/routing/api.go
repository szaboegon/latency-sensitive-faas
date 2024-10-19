package routing

import (
	"faas-loadbalancer/internal/k8s"
	"io"
)

type Component string
type Request struct {
	ToComponent Component
	BodyReader  io.Reader
}

type FunctionLayout struct {
	ApplicationName string          `json:"application-name"`
	FuncPartitions  []FuncPartition `json:"func-partitions"`
}

type FuncPartition struct {
	Name       string      `json:"name"`
	Namespace  string      `json:"namespace"`
	Node       k8s.Node    `json:"edgenode"`
	Components []Component `json:"components"`
}

type Route struct {
	FuncPartition
	Url string
}

func (p1 *FuncPartition) Equals(p2 FuncPartition) bool {
	return p1.Name == p2.Name
}

type RoutingTable map[Component][]Route

type Router interface {
	RouteRequest(Request) error
}
