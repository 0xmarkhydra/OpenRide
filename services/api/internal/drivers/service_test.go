package drivers

import (
	"errors"
	"testing"
	"time"
)

func TestApprovedDriverAvailabilityAndLocation(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, err := service.RegisterApproved("driver-1", "Driver One", "bike")
	if err != nil {
		t.Fatal(err)
	}
	if driver.Availability != AvailabilityOffline {
		t.Fatalf("availability = %s, want offline", driver.Availability)
	}

	driver, err = service.SetAvailability(driver.ID, AvailabilityOnline)
	if err != nil {
		t.Fatal(err)
	}
	driver, err = service.UpdateLocation(driver.ID, Location{Lat: 21.0285, Lng: 105.8542, AccuracyM: 5, CapturedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if driver.Location == nil {
		t.Fatal("location must be stored")
	}
	if got := service.Candidates("bike", time.Minute); len(got) != 1 {
		t.Fatalf("candidate count = %d, want 1", len(got))
	}
}

func TestBusyDriverCannotChangeAvailability(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, _ := service.RegisterApproved("driver-1", "Driver One", "bike")
	_, _ = service.SetAvailability(driver.ID, AvailabilityOnline)
	_, err := service.TryMarkBusy(driver.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetAvailability(driver.ID, AvailabilityOffline); !errors.Is(err, ErrDriverUnavailable) {
		t.Fatalf("error = %v, want ErrDriverUnavailable", err)
	}
}

func TestStaleLocationIsExcluded(t *testing.T) {
	service := NewService(NewMemoryStore())
	driver, _ := service.RegisterApproved("driver-1", "Driver One", "bike")
	_, _ = service.SetAvailability(driver.ID, AvailabilityOnline)
	_, err := service.UpdateLocation(driver.ID, Location{Lat: 21.0285, Lng: 105.8542, AccuracyM: 5, CapturedAt: time.Now().UTC().Add(-3 * time.Minute)})
	if !errors.Is(err, ErrInvalidLocation) {
		t.Fatalf("error = %v, want ErrInvalidLocation", err)
	}
}
