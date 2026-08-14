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

func TestUpdateRuleChangesEstimateAndPersistsVersion(t *testing.T) {
	store := NewMemoryStore()
	service, err := NewServiceWithStore(nil, store)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.UpdateRule(UpdateRuleInput{
		ServiceType:   trips.ServiceDesignatedDriverBike,
		BaseFareMinor: 50_000,
		PerKMMinor:    10_000,
		ServiceMinor:  5_000,
		MinimumMinor:  70_000,
		ActorID:       "admin-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.PricingVersion == "marketplace-demo-v1" || updated.BaseFareMinor != 50_000 || updated.ServiceMinor != 5_000 {
		t.Fatalf("unexpected updated rule: %+v", updated)
	}

	estimate, err := service.Estimate(
		trips.Point{Lat: 21.0285, Lng: 105.8542},
		trips.Point{Lat: 21.035, Lng: 105.81},
		trips.ServiceDesignatedDriverBike,
	)
	if err != nil {
		t.Fatal(err)
	}
	if estimate.PricingVer != updated.PricingVersion || estimate.Fare.BaseFareMinor != 50_000 || estimate.Fare.ServiceMinor != 5_000 {
		t.Fatalf("estimate did not use updated pricing: %+v", estimate)
	}

	restarted, err := NewServiceWithStore(nil, store)
	if err != nil {
		t.Fatal(err)
	}
	rules := restarted.Rules()
	found := false
	for _, item := range rules {
		if item.ServiceType == trips.ServiceDesignatedDriverBike {
			found = true
			if item.PricingVersion != updated.PricingVersion || item.PerKMMinor != 10_000 {
				t.Fatalf("restarted service lost updated rule: %+v", item)
			}
		}
	}
	if !found {
		t.Fatal("updated bike rule not found after restart")
	}
}

func TestUpdateRuleRejectsInvalidAmounts(t *testing.T) {
	service := NewService()
	_, err := service.UpdateRule(UpdateRuleInput{
		ServiceType:   trips.ServiceDesignatedDriverCar,
		BaseFareMinor: -1,
		PerKMMinor:    18_000,
		MinimumMinor:  120_000,
	})
	if !errors.Is(err, ErrInvalidRule) {
		t.Fatalf("error = %v, want ErrInvalidRule", err)
	}
}
