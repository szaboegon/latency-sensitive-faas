package layout

import "lsf-configurator/pkg/core"

type slambucCalculator struct{}

func NewLayoutCalculator() core.LayoutCalculator {
	return &slambucCalculator{}
}

func (c *slambucCalculator) CalculateLayouts(app core.FunctionApp) ([]core.Layout, error) {
	var layouts []core.Layout

	layouts = append(layouts, core.Layout{
	"knative": {"resize", "grayscale"},
	"knative-m02": {"objectdetect", "cut"},
	"knative-m03": {"objectdetect2", "tag"},
	})
	
	return layouts, nil
}