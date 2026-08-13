package trips

import (
	"errors"
	"testing"
	"time"
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
	trip, err = service.MarkVehicleReceived(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.Start(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.MarkHandover(trip.ID, "driver-1")
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

func TestInspectionLifecycleCanCompleteWithFailedResult(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, err := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: ServiceVehicleInspection,
		Pickup: Point{Lat: 19.806, Lng: 105.776}, Destination: Point{Lat: 19.815, Lng: 105.789},
		FareBreakdown: FareBreakdown{ServiceMinor: 299000, TotalMinor: 329000},
	})
	if err != nil {
		t.Fatal(err)
	}
	trip, err = service.AssignDriver(trip.ID, "driver-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []func() error{
		func() error { var e error; trip, e = service.MarkArriving(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.MarkArrived(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.MarkVehicleReceived(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.Start(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.ArriveInspectionCenter(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.StartInspection(trip.ID, "driver-1"); return e },
		func() error {
			var e error
			trip, e = service.CompleteInspection(trip.ID, "driver-1", "failed")
			return e
		},
		func() error { var e error; trip, e = service.ReturningVehicle(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.ArrivedForReturn(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.MarkHandover(trip.ID, "driver-1"); return e },
		func() error { var e error; trip, e = service.Complete(trip.ID, "driver-1", 0); return e },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	if trip.Status != StatusCompleted {
		t.Fatalf("status=%s, want completed", trip.Status)
	}
	if trip.InspectionResult != "failed" {
		t.Fatalf("inspection_result=%s, want failed", trip.InspectionResult)
	}
}

func TestScheduledTripActivatesWhenDue(t *testing.T) {
	service := NewService(NewMemoryStore())
	base := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return base }
	scheduledAt := base.Add(2 * time.Hour)
	trip, err := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: ServiceDesignatedDriverCar,
		BookingMode: BookingScheduled, ScheduledAt: &scheduledAt,
		Pickup: Point{Lat: 19.806, Lng: 105.776}, Destination: Point{Lat: 19.815, Lng: 105.789},
		FareBreakdown: FareBreakdown{ServiceMinor: 120000, TotalMinor: 120000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != StatusScheduled {
		t.Fatalf("status=%s, want scheduled", trip.Status)
	}
	if items, err := service.ActivateDueScheduled(10); err != nil || len(items) != 0 {
		t.Fatalf("activated early: items=%d err=%v", len(items), err)
	}

	service.now = func() time.Time { return scheduledAt }
	items, err := service.ActivateDueScheduled(10)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Status != StatusSearching || items[0].ID != trip.ID {
		t.Fatalf("unexpected activated jobs: %+v", items)
	}
}

func TestRiderCannotCancelAfterVehicleWasReceived(t *testing.T) {
	service := NewService(NewMemoryStore())
	trip, err := service.Create(CreateInput{
		RiderID: "rider-1", ServiceType: ServiceDesignatedDriverCar,
		Pickup: Point{Lat: 19.806, Lng: 105.776}, Destination: Point{Lat: 19.815, Lng: 105.789},
		FareBreakdown: FareBreakdown{TotalMinor: 120000},
	})
	if err != nil {
		t.Fatal(err)
	}
	trip, _ = service.AssignDriver(trip.ID, "driver-1")
	trip, _ = service.MarkArrived(trip.ID, "driver-1")
	trip, _ = service.MarkVehicleReceived(trip.ID, "driver-1")
	if _, err := service.Cancel(trip.ID, "rider-1", "changed_mind"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("cancel after custody error=%v, want ErrInvalidState", err)
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
	trip, _ = service.MarkVehicleReceived(trip.ID, "driver-1")
	trip, _ = service.Start(trip.ID, "driver-1")
	trip, _ = service.MarkHandover(trip.ID, "driver-1")
	trip, _ = service.Complete(trip.ID, "driver-1", 20000)

	if _, err := service.Cancel(trip.ID, "rider-1", "changed_mind"); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("cancel completed error = %v, want ErrInvalidState", err)
	}
}
