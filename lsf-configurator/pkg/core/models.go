package core

type Component string

type FunctionApp struct {
	Id           string                          `json:"id,omitempty"`
	Name         string                          `json:"name"`
	Compositions map[string]*FunctionComposition `json:"compositions,omitempty"`
	Components   []Component                     `json:"components,omitempty"`
}

type FunctionComposition struct {
	Id            string       `json:"id,omitempty"`
	FunctionAppId string       `json:"function_app_id,omitempty"`
	Node          string       `json:"node,omitempty"`
	Components    RoutingTable `json:"components"`
	NameSpace     string       `json:"namespace"`
	SourcePath    string       `json:"source_path,omitempty"`
	Runtime       string       `json:"runtime"`
	Files         []string     `json:"files"`
	Build
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
