package core

type Component string

type FunctionApp struct {
	Id           string                 `json:"id"`
	Name         string                 `json:"name"`
	Runtime      string                 `json:"runtime"`
	Components   []Component            `json:"components"`
	Files        []string               `json:"files"`
	Compositions []*FunctionComposition `json:"compositions"`
	SourcePath   string                 `json:"source_path"`
}

type BuildStatus string

const (
	BuildStatusPending BuildStatus = "pending"
	BuildStatusBuilt   BuildStatus = "built"
	BuildStatusError   BuildStatus = "error"
)

type FunctionComposition struct {
	Id            string      `json:"id"`
	FunctionAppId string      `json:"function_app_id"`
	Components    []Component `json:"components"`
	Files         []string    `json:"files"`
	Status        BuildStatus `json:"status"`
	Build         `json:"build"`
	Deployments   []*Deployment `json:"deployments"`
}

type DeploymentStatus string

const (
	DeploymentStatusWaitingForBuild DeploymentStatus = "waiting_for_build"
	DeploymentStatusPending         DeploymentStatus = "pending"
	DeploymentStatusDeployed        DeploymentStatus = "deployed"
	DeploymentStatusError           DeploymentStatus = "error"
)

type Deployment struct {
	Id                    string           `json:"id"`
	FunctionCompositionId string           `json:"function_composition_id"`
	Node                  string           `json:"node"`
	Namespace             string           `json:"namespace"`
	RoutingTable          RoutingTable     `json:"routing_table"`
	Status                DeploymentStatus `json:"status"`
}

type Route struct {
	To       string `json:"to"`
	Function string `json:"function"`
}

type RoutingTable map[Component][]Route

type Build struct {
	Image     string `json:"image"`
	Timestamp string `json:"timestamp"`
}
