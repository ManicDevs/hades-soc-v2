package modules

import (
	"fmt"

	"hades-v2/modules/auxiliary"
	"hades-v2/modules/payload"
	"hades-v2/pkg/sdk"
)

// LoadAllModules registers all available modules with the dispatcher
func LoadAllModules(dispatcher interface {
	RegisterModule(module sdk.Module) error
}) error {
	modules := []sdk.Module{
		// Payload modules
		payload.NewReverseShell(),

		// Auxiliary modules
		auxiliary.NewAPIServerFixed(),
		auxiliary.NewCacheManager(),
		auxiliary.NewDashboard(),
		auxiliary.NewResourceMonitor(),
		auxiliary.NewRiskScanner(),
		auxiliary.NewSIEMIntegration(),
		auxiliary.NewEventHandler(),
		auxiliary.NewTrendAnalyzer(),
		auxiliary.NewDistributedScanner(),
	}

	for _, module := range modules {
		if err := dispatcher.RegisterModule(module); err != nil {
			return fmt.Errorf("hades.modules.loader: failed to register %s: %w", module.Name(), err)
		}
	}

	return nil
}
