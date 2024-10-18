package routing

type Component string
type Request struct {
	ToComponent Component
	Payload     any
}

type FunctionLayout struct {
	ApplicationName string          `json:"application-name"`
	FuncPartitions  []FuncPartition `json:"func-partitions"`
}

type FuncPartition struct {
	Name       string
	Node       string
	Components []Component
}

func (p1 *FuncPartition) Equals(p2 FuncPartition) bool {
	return p1.Name == p2.Name
}

type RoutingTable map[Component][]FuncPartition

type Router interface {
	RouteRequest(Request) error
}
