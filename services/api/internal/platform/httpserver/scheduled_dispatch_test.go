package httpserver

import (
	"testing"
	"time"

	"flashx/services/api/internal/trips"
)

func TestTickScheduledDispatchActivatesDueJobWithoutDriverHeartbeat(t *testing.T) {
	store := trips.NewMemoryStore()
	now := time.Now().UTC()
	due := now.Add(-time.Minute)
	trip := trips.Trip{
		ID:          "trip_scheduled_due",
		RiderID:     "rider_scheduled",
		ServiceType: trips.ServiceDesignatedDriverCar,
		BookingMode: trips.BookingScheduled,
		ScheduledAt: &due,
		Status:      trips.StatusScheduled,
		Pickup:      trips.Point{Lat: 19.807, Lng: 105.776},
		Destination: trips.Point{Lat: 19.82, Lng: 105.79},
		Currency:    "VND",
		CreatedAt:   now.Add(-time.Hour),
		Version:     1,
	}
	if err := store.Create(trip); err != nil {
		t.Fatal(err)
	}

	service := trips.NewService(store)
	server := &Server{deps: Dependencies{Trips: service}}
	if err := server.TickScheduledDispatch(); err != nil {
		t.Fatal(err)
	}

	updated, err := service.Get(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != trips.StatusSearching {
		t.Fatalf("status=%s want %s", updated.Status, trips.StatusSearching)
	}
}
