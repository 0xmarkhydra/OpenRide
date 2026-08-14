package operationalsettings

import (
	"errors"
	"testing"

	"flashx/services/api/internal/trips"
)

func TestUpdatePersistsAndAppliesOperationalSettings(t *testing.T) {
	store := NewMemoryStore()
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	var applied Config
	service.SetApplyHook(func(config Config) { applied = config })

	updated, err := service.Update(Config{
		DesignatedDriverCarEnabled:  false,
		DesignatedDriverBikeEnabled: true,
		VehicleInspectionEnabled:    true,
		DispatchMaxDistanceM:        8000,
		DriverLocationMaxAgeSeconds: 30,
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version < 2 || updated.UpdatedBy != "admin-1" {
		t.Fatalf("unexpected update metadata: %+v", updated)
	}
	if service.IsServiceEnabled(trips.ServiceDesignatedDriverCar) {
		t.Fatal("car service should be disabled")
	}
	if !service.IsServiceEnabled(trips.ServiceDesignatedDriverBike) || !service.IsServiceEnabled(trips.ServiceVehicleInspection) {
		t.Fatal("bike and inspection services should remain enabled")
	}
	if applied.DispatchMaxDistanceM != 8000 || applied.DriverLocationMaxAgeSeconds != 30 {
		t.Fatalf("apply hook did not receive policy: %+v", applied)
	}

	restarted, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.Get().DispatchMaxDistanceM != 8000 || restarted.IsServiceEnabled(trips.ServiceDesignatedDriverCar) {
		t.Fatalf("persisted settings were not restored: %+v", restarted.Get())
	}
}

func TestUpdateRejectsUnsafeBounds(t *testing.T) {
	service, err := NewService(NewMemoryStore())
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Update(Config{
		DesignatedDriverCarEnabled:  true,
		DesignatedDriverBikeEnabled: true,
		VehicleInspectionEnabled:    true,
		DispatchMaxDistanceM:        999,
		DriverLocationMaxAgeSeconds: 20,
	}, "admin-1")
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
}
