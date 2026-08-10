package pricing

import (
	"errors"
	"testing"

	"flashx/services/api/internal/trips"
)

func TestEstimateBike(t *testing.T) {
	service := NewService()
	estimate, err := service.Estimate(
		trips.Point{Lat: 21.0285, Lng: 105.8542},
		trips.Point{Lat: 21.035, Lng: 105.81},
		"bike",
	)
	if err != nil {
		t.Fatal(err)
	}
	if estimate.DistanceM <= 0 || estimate.DurationS <= 0 {
		t.Fatalf("invalid route estimate: %+v", estimate)
	}
	if estimate.Fare.TotalMinor < estimate.Fare.BaseFareMinor {
		t.Fatalf("invalid fare: %+v", estimate.Fare)
	}
	if estimate.Currency != "VND" {
		t.Fatalf("currency = %s, want VND", estimate.Currency)
	}
}

func TestUnsupportedServiceType(t *testing.T) {
	service := NewService()
	_, err := service.Estimate(trips.Point{}, trips.Point{Lat: 1, Lng: 1}, "spaceship")
	if !errors.Is(err, ErrUnsupportedServiceType) {
		t.Fatalf("error = %v, want ErrUnsupportedServiceType", err)
	}
}

func TestInvalidRoute(t *testing.T) {
	service := NewService()
	_, err := service.Estimate(trips.Point{Lat: 100, Lng: 0}, trips.Point{Lat: 1, Lng: 1}, "bike")
	if !errors.Is(err, ErrInvalidRoute) {
		t.Fatalf("error = %v, want ErrInvalidRoute", err)
	}
}
