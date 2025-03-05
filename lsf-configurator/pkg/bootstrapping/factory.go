package bootstrapping

import (
	"fmt"
	"lsf-configurator/pkg/core"
)

func NewBootstrapper(fc core.FunctionComposition, buildDir string) (Bootstrapper, error) {
	switch fc.Runtime {
	case "python":
		return &PythonBootstrapper{BaseBootstrapper{fc: fc, buildDir: buildDir}}, nil
	default:
		return nil, fmt.Errorf("no bootstrapper found for runtime: %v", fc.Runtime)
	}
}
