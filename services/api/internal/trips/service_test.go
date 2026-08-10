package trips

import (
	"errors"
	"testing"
)

func TestTripLifecycle(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, err := service.Create(CreateInput{
		RiderID:            "rider-1",
		ServiceType:        "bike",
		Pickup:             Point{Lat: 21.0285, Lng: 105.8542},
		Destination:        Point{Lat: 21.035, Lng: 105.81},
		EstimatedDistanceM: 5000,
		EstimatedDurationS: 900,
		FareBreakdown:      FareBreakdown{BaseFareMinor: 12000, DistanceMinor: 25000, TotalMinor: 37000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != StatusSearching {
		t.Fatalf("status = %s, want searching", trip.Status)
	}

	trip, err = service.AssignDriver(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != StatusAccepted || trip.DriverID != "driver-1" {
		t.Fatalf("unexpected accepted trip: %+v", trip)
	}

	trip, err = service.MarkArriving(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.MarkArrived(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.Start(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.Complete(trip.ID, "driver-1", 0)
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != StatusCompleted {
		t.Fatalf("status = %s, want completed", trip.Status)
	}
	if trip.FinalFareMinor != trip.EstimatedFareMinor {
		t.Fatalf("final fare = %d, want %d", trip.FinalFareMinor, trip.EstimatedFareMinor)
	}
	if trip.CompletedAt == nil {
		t.Fatal("completed_at must be set")
	}
}

func TestInvalidTransitionsAreRejected(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, err := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: "bike",
		Pickup: Point{Lat: 21, Lng: 105}, Destination: Point{Lat: 21.1, Lng: 105.1},
		FareBreakdown: FareBreakdown{TotalMinor: 20000},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := service.Start(trip.ID, "driver-1"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("start without assignment error = %v, want ErrForbidden", err)
	}

	trip, err = service.AssignDriver(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Start(trip.ID, "driver-1"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("start before arrived error = %v, want ErrInvalidState", err)
	}
}

func TestRiderCannotAccessAnotherRiderTrip(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, err := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: "bike",
		Pickup: Point{Lat: 21, Lng: 105}, Destination: Point{Lat: 21.1, Lng: 105.1},
		FareBreakdown: FareBreakdown{TotalMinor: 20000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetForRider(trip.ID, "rider-2"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}
}

func TestCompletedTripCannotBeCancelled(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, _ := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: "bike",
		Pickup: Point{Lat: 21, Lng: 105}, Destination: Point{Lat: 21.1, Lng: 105.1},
		FareBreakdown: FareBreakdown{TotalMinor: 20000},
	})
	trip, _ = service.AssignDriver(trip.ID, "driver-1")
	trip, _ = service.MarkArrived(trip.ID, "driver-1")
	trip, _ = service.Start(trip.ID, "driver-1")
	trip, _ = service.Complete(trip.ID, "driver-1", 20000)

	if _, err := service.Cancel(trip.ID, "rider-1", "changed_mind"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("cancel completed error = %v, want ErrInvalidState", err)
	}
}
