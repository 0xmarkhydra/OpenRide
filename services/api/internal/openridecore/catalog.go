package openridecore

import (
	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
)

// DefaultServiceRegistry is the runtime bridge from the current API server to
// packageable OpenRide service modules. New V2 handlers should depend on this
// registry instead of hardcoding service types in transport code.
func DefaultServiceRegistry() *extension.Registry {
	return modulecatalog.DefaultRegistry()
}

func DefaultServiceManifests() []extension.Manifest {
	return DefaultServiceRegistry().Manifests()
}
