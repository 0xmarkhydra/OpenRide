package dispatch

import (
	"errors"
	"sync"
	"testing"
	"time"

	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/trips"
)

func setupDispatch(t *testing.T) (*Engine, *drivers.Service, *trips.Service, trips.Trip) {
	t.Helper()
	driverService := drivers.NewService(drivers.NewMemoryStore())
	tripService := trips.NewService(trips.NewMemoryStore())

	for _, item := range []struct {
		id  string
		lat float64
		lng float64
	}{
		{id: "driver-near", lat: 21.0290, lng: 105.8540},
		{id: "driver-far", lat: 21.0400, lng: 105.8400},
	} {
		_, err := driverService.RegisterApproved(item.id, item.id, "bike")
		if err != nil {
			t.Fatal(err)
		}
		_, err = driverService.SetAvailability(item.id, drivers.AvailabilityOnline)
		if err != nil {
			t.Fatal(err)
		}
		_, err = driverService.UpdateLocation(item.id, drivers.Location{
			Lat: item.lat, Lng: item.lng, AccuracyM: 5, CapturedAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	trip, err := tripService.Create(trips.CreateInput{
		RiderID: "rider-1", ServiceType: "bike",
		Pickup:        trips.Point{Lat: 21.0285, Lng: 105.8542},
		Destination:   trips.Point{Lat: 21.0350, Lng: 105.8100},
		FareBreakdown: trips.FareBreakdown{TotalMinor: 40000},
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngine(driverService, tripService), driverService, tripService, trip
}

func TestDispatchChoosesNearerDriver(t *testing.T) {
	engine, _, _, trip := setupDispatch(t)
	offer, err := engine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if offer.DriverID != "driver-near" {
		t.Fatalf("driver = %s, want driver-near", offer.DriverID)
	}
}

func TestRejectMovesToNextCandidate(t *testing.T) {
	engine, _, _, trip := setupDispatch(t)
	first, err := engine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Reject(first.ID, first.DriverID); err != nil {
		t.Fatal(err)
	}
	second, err := engine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if second.DriverID == first.DriverID {
		t.Fatalf("second offer reused rejected driver %s", first.DriverID)
	}
}

func TestAcceptAssignsExactlyOnce(t *testing.T) {
	engine, driverService, tripService, trip := setupDispatch(t)
	offer, err := engine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}

	type result struct{ err error }
	results := make(chan result, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := engine.Accept(offer.ID, offer.DriverID)
			results <- result{err: err}
		}()
	}
	wg.Wait()
	close(results)

	successes := 0
	failures := 0
	for r := range results {
		if r.err == nil {
			successes++
		} else {
			failures++
			if !errors.Is(r.err, ErrOfferUnavailable) {
				t.Fatalf("unexpected second accept error: %v", r.err)
			}
		}
	}
	if successes != 1 || failures != 1 {
		t.Fatalf("successes=%d failures=%d, want 1/1", successes, failures)
	}

	assigned, err := tripService.Get(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if assigned.DriverID != offer.DriverID || assigned.Status != trips.StatusAccepted {
		t.Fatalf("unexpected assigned trip: %+v", assigned)
	}
	driver, err := driverService.Get(offer.DriverID)
	if err != nil {
		t.Fatal(err)
	}
	if driver.Availability != drivers.AvailabilityBusy {
		t.Fatalf("driver availability = %s, want busy", driver.Availability)
	}
}

func TestNoCandidateWhenDriversOffline(t *testing.T) {
	driverService := drivers.NewService(drivers.NewMemoryStore())
	tripService := trips.NewService(trips.NewMemoryStore())
	trip, _ := tripService.Create(trips.CreateInput{
		RiderID: "rider-1", ServiceType: "bike",
		Pickup: trips.Point{Lat: 21, Lng: 105}, Destination: trips.Point{Lat: 21.1, Lng: 105.1},
		FareBreakdown: trips.FareBreakdown{TotalMinor: 30000},
	})
	engine := NewEngine(driverService, tripService)
	if _, err := engine.CreateOffer(trip.ID); !errors.Is(err, ErrNoCandidate) {
		t.Fatalf("error = %v, want ErrNoCandidate", err)
	}
}

func TestManualAssignUsesEligibleDriverAndInvalidatesOffer(t *testing.T) {
	engine, driverService, tripService, trip := setupDispatch(t)
	pending, err := engine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if pending.DriverID != "driver-near" {
		t.Fatalf("pending driver = %s, want driver-near", pending.DriverID)
	}

	assigned, err := engine.ManualAssign(trip.ID, "driver-far")
	if err != nil {
		t.Fatal(err)
	}
	if assigned.DriverID != "driver-far" || assigned.Status != trips.StatusAccepted {
		t.Fatalf("unexpected manually assigned trip: %+v", assigned)
	}
	far, err := driverService.Get("driver-far")
	if err != nil {
		t.Fatal(err)
	}
	if far.Availability != drivers.AvailabilityBusy {
		t.Fatalf("driver-far availability = %s, want busy", far.Availability)
	}
	storedOffer, err := engine.offers.Get(pending.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedOffer.Status != OfferInvalid {
		t.Fatalf("offer status = %s, want invalidated", storedOffer.Status)
	}
	storedTrip, err := tripService.Get(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if storedTrip.DriverID != "driver-far" {
		t.Fatalf("stored driver = %s, want driver-far", storedTrip.DriverID)
	}
}

func TestManualReassignBlockedAfterVehicleReceived(t *testing.T) {
	engine, _, tripService, trip := setupDispatch(t)
	assigned, err := engine.ManualAssign(trip.ID, "driver-near")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArriving(assigned.ID, "driver-near"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkArrived(assigned.ID, "driver-near"); err != nil {
		t.Fatal(err)
	}
	if _, err := tripService.MarkVehicleReceived(assigned.ID, "driver-near"); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.ManualAssign(assigned.ID, "driver-far"); !errors.Is(err, trips.ErrInvalidState) {
		t.Fatalf("error = %v, want trips.ErrInvalidState", err)
	}
}
