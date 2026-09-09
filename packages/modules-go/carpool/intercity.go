package carpool

import (
	"context"
	"errors"
	"fmt"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var (
	ErrDestinationRequired = errors.New("carpool.intercity: destination is required")
	ErrSeatsRequired       = errors.New("carpool.intercity: seats must be between 1 and 8")
)

type Intercity struct{}

func (Intercity) Manifest() extension.Manifest {
	return extension.Manifest{
		ID:          "carpool.intercity",
		Version:     "1.0.0",
		DisplayName: "Intercity Carpool",
		Description: "Shared intercity mobility with seat-aware requests.",
		Category:    "carpool",
		Capabilities: []extension.Capability{
			extension.CapabilityPassenger,
			extension.CapabilityCarpool,
			extension.CapabilityScheduled,
		},
		Contracts: map[string]string{
			"request_schema": "https://openride.dev/contracts/request.carpool-intercity.v1.schema.json",
		},
	}
}

func (Intercity) ValidateRequest(_ context.Context, request marketplace.Request) error {
	if request.Destination == nil {
		return ErrDestinationRequired
	}
	seats, ok := integerAttribute(request.Attributes, "seats")
	if !ok || seats < 1 || seats > 8 {
		return ErrSeatsRequired
	}
	return nil
}

func integerAttribute(attributes map[string]any, key string) (int64, bool) {
	value, ok := attributes[key]
	if !ok {
		return 0, false
	}
	switch n := value.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case float64:
		if n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	default:
		_ = fmt.Sprintf("%v", value)
		return 0, false
	}
}

func (Intercity) RideLifecycle() extension.RideLifecycle { return intercityLifecycle{} }

type intercityLifecycle struct{}

const (
	RidePooling       marketplace.RideStatus = "pooling"
	RidePickupWindow  marketplace.RideStatus = "pickup_window"
	RideIntercity     marketplace.RideStatus = "intercity_in_progress"
	RideDropoffWindow marketplace.RideStatus = "dropoff_window"
)

func (intercityLifecycle) Initial() marketplace.RideStatus { return RidePooling }

func (intercityLifecycle) ValidTransition(from, to marketplace.RideStatus) bool {
	if to == marketplace.RideCancelled || to == marketplace.RideFailed {
		return !intercityLifecycle{}.Terminal(from)
	}
	switch from {
	case RidePooling:
		return to == RidePickupWindow
	case RidePickupWindow:
		return to == RideIntercity
	case RideIntercity:
		return to == RideDropoffWindow
	case RideDropoffWindow:
		return to == marketplace.RideCompleted
	default:
		return false
	}
}

func (intercityLifecycle) Terminal(status marketplace.RideStatus) bool {
	return status == marketplace.RideCompleted || status == marketplace.RideCancelled || status == marketplace.RideFailed
}
