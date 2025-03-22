package core

type FunctionApp struct {
	Id           string
	Compositions map[string]*FunctionComposition
}

type FunctionComposition struct {
	Id         string      `json:"-"`
	Node       string      `json:"node,omitempty"`
	Components []Component `json:"components"`
	NameSpace  string      `json:"namespace"`
	SourcePath string      `json:"-"`
	Runtime    string      `json:"runtime"`
	Build
}

type Component struct {
	Name        string      `json:"name"`
	Next        []string    `json:"next"`
	RoutingRule RoutingRule `json:"-"`
}

type Build struct {
	Image string `json:"-"`
	Stamp string `json:"-"`
}

type RoutingRule struct {
}
