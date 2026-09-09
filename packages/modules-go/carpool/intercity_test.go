package carpool

import (
	"context"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

func TestIntercityRequiresValidSeats(t *testing.T) {
	destination := geo.Point{Lat: 21.0, Lng: 105.8}
	request := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "carpool.intercity",
		Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 19.8, Lng: 105.7}, Destination: &destination,
		Attributes: map[string]any{"seats": 2.0}, RequestedAt: time.Now(), Version: 1,
	}
	if err := (Intercity{}).ValidateRequest(context.Background(), request); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	request.Attributes["seats"] = 9
	if err := (Intercity{}).ValidateRequest(context.Background(), request); err == nil {
		t.Fatal("expected invalid seats error")
	}
}

func TestIntercityLifecycleIsIndependentFromPassengerLifecycle(t *testing.T) {
	lifecycle := (Intercity{}).RideLifecycle()
	if lifecycle.Initial() != RidePooling {
		t.Fatalf("unexpected initial state: %s", lifecycle.Initial())
	}
	path := []marketplace.RideStatus{RidePooling, RidePickupWindow, RideIntercity, RideDropoffWindow, marketplace.RideCompleted}
	for i := 0; i < len(path)-1; i++ {
		if !lifecycle.ValidTransition(path[i], path[i+1]) {
			t.Fatalf("expected transition %s -> %s", path[i], path[i+1])
		}
	}
}
