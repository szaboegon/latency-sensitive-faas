package layout

import "lsf-configurator/pkg/core"

type slambucCalculator struct{}

func NewLayoutCalculator() core.LayoutCalculator {
	return &slambucCalculator{}
}

func (c *slambucCalculator) CalculateLayout(app core.FunctionApp) (map[string][]string, error) {
	return map[string][]string{}, nil
}