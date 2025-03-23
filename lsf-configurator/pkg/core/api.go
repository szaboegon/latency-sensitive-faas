package core

type FunctionApp struct {
	Id           string
	Compositions map[string]*FunctionComposition
	RoutingRules map[*Component]*RoutingRule
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
	Name string   `json:"name"`
	Next []string `json:"next"`
}

type Build struct {
	Image string `json:"-"`
	Stamp string `json:"-"`
}

type RuleType string

const (
	SingleForward RuleType = "single"
	MultiForward  RuleType = "multi"
)

type SingleRoute struct {
	Target string `json:"target"`
}

type MultiRoute struct {
	Targets []WeightedTarget `json:"targets"`
}

type WeightedTarget struct {
	Target string  `json:"target"`
	Weight float64 `json:"weight"` // Percentage (0-100)
}

type RoutingRule struct {
	Type   RuleType     `json:"type"`
	Single *SingleRoute `json:"single,omitempty"`
	Multi  *MultiRoute  `json:"multi,omitempty"`
}
