package bootstrapping

import (
	"fmt"
	"lsf-configurator/pkg/core"
)

func NewBootstrapper(runtime string, fc core.FunctionComposition, buildDir string, sourcePath string) (Bootstrapper, error) {
	switch runtime {
	case "python":
		return &PythonBootstrapper{BaseBootstrapper{fc: fc, buildDir: buildDir, sourcePath: sourcePath}}, nil
	default:
		return nil, fmt.Errorf("no bootstrapper found for runtime: %v", runtime)
	}
}
