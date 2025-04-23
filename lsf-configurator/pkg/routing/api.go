package routing

type RoutingTable map[string][]Route

type Route struct {
	Component string
	Url       string
}
