package extension

import "github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"

// RideLifecycle lets a service vertical define execution states without forcing
// every possible mobility workflow into one global enum.
type RideLifecycle interface {
	Initial() marketplace.RideStatus
	ValidTransition(from, to marketplace.RideStatus) bool
	Terminal(status marketplace.RideStatus) bool
}

// RideLifecycleProvider is optional. A ServiceModule implements it only when
// its execution lifecycle differs from the operator's default lifecycle.
type RideLifecycleProvider interface {
	RideLifecycle() RideLifecycle
}
