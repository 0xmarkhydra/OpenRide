package persistence_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/platform/config"
	"flashx/services/api/internal/platform/persistence"
	"flashx/services/api/internal/ride"
	"flashx/services/api/internal/trips"
)

func TestPersistentRideLifecycle(t *testing.T) {
	if os.Getenv("FLASHX_INTEGRATION") != "1" {
		t.Skip("set FLASHX_INTEGRATION=1 to run persistent integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := config.Load()
	resources, err := persistence.Open(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer resources.Close()

	riderID := uuidString()
	driverID := uuidString()
	tripID := ""
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if tripID != "" {
			_, _ = resources.Postgres.Exec(cleanupCtx, `DELETE FROM trip_status_history WHERE trip_id=$1`, tripID)
			_, _ = resources.Postgres.Exec(cleanupCtx, `DELETE FROM trips WHERE id=$1`, tripID)
		}
		_, _ = resources.Postgres.Exec(cleanupCtx, `DELETE FROM drivers WHERE id=$1`, driverID)
		_, _ = resources.Postgres.Exec(cleanupCtx, `DELETE FROM users WHERE id=$1`, riderID)
	}()

	if _, err := resources.Postgres.Exec(ctx, `
		INSERT INTO users (id, phone, full_name) VALUES ($1, $2, 'Integration Rider')
	`, riderID, "+849"+hex.EncodeToString(randomBytes(5))); err != nil {
		t.Fatal(err)
	}

	tripService := trips.NewService(trips.NewPostgresStore(resources.Postgres))
	driverService := drivers.NewServiceWithLocationIndex(
		drivers.NewPostgresStore(resources.Postgres),
		drivers.NewRedisLocationIndex(resources.Redis),
	)
	dispatchNamespace := "flashx-it-" + driverID
	dispatchEngine := dispatch.NewEngineWithStore(
		driverService,
		tripService,
		dispatch.NewRedisOfferStore(resources.Redis, dispatchNamespace),
		dispatch.NewRedisLocker(resources.Redis, dispatchNamespace),
	)
	rideService := ride.NewService(tripService, driverService)

	if _, err := driverService.RegisterApproved(driverID, "Integration Driver", "bike"); err != nil {
		t.Fatal(err)
	}
	if _, err := driverService.SetAvailability(driverID, drivers.AvailabilityOnline); err != nil {
		t.Fatal(err)
	}
	if _, err := driverService.UpdateLocation(driverID, drivers.Location{
		Lat: 21.0290, Lng: 105.8540, AccuracyM: 5, CapturedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}

	trip, err := tripService.Create(trips.CreateInput{
		RiderID:            riderID,
		ServiceType:        "bike",
		Pickup:             trips.Point{Lat: 21.0285, Lng: 105.8542},
		Destination:        trips.Point{Lat: 21.0350, Lng: 105.8100},
		EstimatedDistanceM: 5200,
		EstimatedDurationS: 900,
		FareBreakdown:      trips.FareBreakdown{BaseFareMinor: 12000, DistanceMinor: 30000, TotalMinor: 42000},
	})
	if err != nil {
		t.Fatal(err)
	}
	tripID = trip.ID

	offer, err := dispatchEngine.CreateOffer(trip.ID)
	if err != nil {
		t.Fatal(err)
	}
	if offer.DriverID != driverID {
		t.Fatalf("offer driver=%s want=%s", offer.DriverID, driverID)
	}

	// Simulate an API process restart. Pending offers must remain visible because
	// production dispatch state lives in Redis rather than process memory.
	restartedDispatch := dispatch.NewEngineWithStore(
		driverService,
		tripService,
		dispatch.NewRedisOfferStore(resources.Redis, dispatchNamespace),
		dispatch.NewRedisLocker(resources.Redis, dispatchNamespace),
	)
	recoveredOffer, err := restartedDispatch.CurrentOfferForDriver(driverID)
	if err != nil {
		t.Fatal(err)
	}
	if recoveredOffer.ID != offer.ID {
		t.Fatalf("recovered offer=%s want=%s", recoveredOffer.ID, offer.ID)
	}

	_, trip, err = restartedDispatch.Accept(offer.ID, driverID)
	if err != nil {
		t.Fatal(err)
	}
	if trip.Status != trips.StatusAccepted {
		t.Fatalf("status=%s want accepted", trip.Status)
	}

	if _, err := rideService.MarkArriving(trip.ID, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err := rideService.MarkArrived(trip.ID, driverID); err != nil {
		t.Fatal(err)
	}
	if _, err := rideService.Start(trip.ID, driverID); err != nil {
		t.Fatal(err)
	}
	completed, err := rideService.Complete(trip.ID, driverID, 43000)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != trips.StatusCompleted || completed.FinalFareMinor != 43000 {
		t.Fatalf("unexpected completed trip: %+v", completed)
	}

	driver, err := driverService.Get(driverID)
	if err != nil {
		t.Fatal(err)
	}
	if driver.Availability != drivers.AvailabilityOnline {
		t.Fatalf("driver availability=%s want online", driver.Availability)
	}
}

func uuidString() string {
	b := randomBytes(16)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}
