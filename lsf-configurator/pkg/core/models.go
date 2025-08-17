package core

type Component string

type FunctionApp struct {
	Id           string                 `json:"id,omitempty"`
	Name         string                 `json:"name"`
	Runtime      string                 `json:"runtime"`
	Components   []Component            `json:"components,omitempty"`
	Files        []string               `json:"files,omitempty"`
	Compositions []*FunctionComposition `json:"compositions,omitempty"`
	SourcePath   string                 `json:"source_path,omitempty"`
}

type Status string

const (
	StatusPending  Status = "pending"
	StatusBuilt    Status = "built"
	StatusDeployed Status = "deployed"
	StatusError    Status = "error"
)

type FunctionComposition struct {
	Id            string      `json:"id,omitempty"`
	FunctionAppId string      `json:"function_app_id,omitempty"`
	Components    []Component `json:"components"`
	Files         []string    `json:"files"`
	Status        Status      `json:"status,omitempty"`
	Build         `json:"build,omitempty"`
	Deployments   []*Deployment `json:"deployments,omitempty"`
}

type Deployment struct {
	Id                    string       `json:"id,omitempty"`
	FunctionCompositionId string       `json:"function_composition_id,omitempty"`
	Node                  string       `json:"node,omitempty"`
	Namespace             string       `json:"namespace,omitempty"`
	RoutingTable          RoutingTable `json:"routing_table,omitempty"`
}

type Route struct {
	To       string `json:"to"`
	Function string `json:"function"`
}

type RoutingTable map[Component][]Route

type Build struct {
	Image     string `json:"image,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}
