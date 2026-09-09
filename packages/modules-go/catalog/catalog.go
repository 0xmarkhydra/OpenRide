package catalog

import (
	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/modules-go/carpool"
	"github.com/0xmarkhydra/OpenRide/packages/modules-go/passenger"
)

// DefaultRegistry returns the first-party service modules shipped with OpenRide.
// Operators can start here, then register or replace modules in their own runtime.
func DefaultRegistry() *extension.Registry {
	return extension.MustRegistry(
		passenger.Car{},
		carpool.Intercity{},
	)
}
