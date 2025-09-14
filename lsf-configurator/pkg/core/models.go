package core

type Component struct {
	Name    string `json:"name"`
	Memory  int    `json:"memory"`  // in MB
	Runtime string `json:"runtime"` // The execution time of the component in milliseconds
}

type ComponentLink struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	InvocationRate float64 `json:"invocation_rate"` // Relative invocation rate between components
}

type FunctionApp struct {
	Id           string                 `json:"id"`
	Name         string                 `json:"name"`
	Runtime      string                 `json:"runtime"`
	Components   []Component            `json:"components"`
	Links        []ComponentLink        `json:"links"`
	Files        []string               `json:"files"`
	Compositions []*FunctionComposition `json:"compositions"`
	SourcePath   string                 `json:"source_path"`
	LatencyLimit int                    `json:"latency_limit"` // in milliseconds
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
	Components    []string    `json:"components"`
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
	To       string `json:"to"`       // Component name
	Function string `json:"function"` // Deployment of a function composition
}

type RoutingTable map[string][]Route // Key: Component name

type Build struct {
	Image     string `json:"image"`
	Timestamp string `json:"timestamp"`
}
