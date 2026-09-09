package passenger

import (
	"context"
	"errors"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var ErrDestinationRequired = errors.New("passenger.car: destination is required")

type Car struct{}

func (Car) Manifest() extension.Manifest {
	return extension.Manifest{
		ID:          "passenger.car",
		Version:     "1.0.0",
		DisplayName: "Passenger Car",
		Description: "Point-to-point passenger ride by car.",
		Category:    "passenger",
		Capabilities: []extension.Capability{
			extension.CapabilityPassenger,
			extension.CapabilityScheduled,
		},
		Contracts: map[string]string{
			"request_schema": "https://openride.dev/contracts/request.passenger-car.v1.schema.json",
		},
	}
}

func (Car) ValidateRequest(_ context.Context, request marketplace.Request) error {
	if request.Destination == nil {
		return ErrDestinationRequired
	}
	return nil
}

func (Car) RideLifecycle() extension.RideLifecycle {
	return marketplace.PassengerLifecycle{}
}
