package passenger

import (
	"context"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

func TestCarRequiresDestination(t *testing.T) {
	request := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car",
		Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 19.8, Lng: 105.7}, RequestedAt: time.Now(), Version: 1,
	}
	if err := (Car{}).ValidateRequest(context.Background(), request); err == nil {
		t.Fatal("expected destination validation error")
	}
	destination := geo.Point{Lat: 19.7, Lng: 105.8}
	request.Destination = &destination
	if err := (Car{}).ValidateRequest(context.Background(), request); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
