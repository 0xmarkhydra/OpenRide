package extension

import (
	"context"
	"testing"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

type testModule struct{ id marketplace.ServiceType }

func (m testModule) Manifest() Manifest {
	return Manifest{ID: m.id, Version: "1.0.0", DisplayName: "Test", Category: "test"}
}
func (m testModule) ValidateRequest(context.Context, marketplace.Request) error { return nil }

func TestRegistryRejectsDuplicateServiceIDs(t *testing.T) {
	r := MustRegistry(testModule{id: "passenger.car"})
	if err := r.Register(testModule{id: "passenger.car"}); err == nil {
		t.Fatal("expected duplicate service id to fail")
	}
}

func TestRegistryListsDeterministically(t *testing.T) {
	r := MustRegistry(testModule{id: "delivery.parcel"}, testModule{id: "passenger.car"})
	items := r.Manifests()
	if len(items) != 2 || items[0].ID != "delivery.parcel" || items[1].ID != "passenger.car" {
		t.Fatalf("unexpected manifest ordering: %+v", items)
	}
}
