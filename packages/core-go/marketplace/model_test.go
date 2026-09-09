package marketplace

import (
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

func validRequest(now time.Time) Request {
	destination := geo.Point{Lat: 21.0285, Lng: 105.8542}
	return Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1",
		ServiceType: "passenger.car", Status: RequestOpen,
		Pickup: geo.Point{Lat: 21.0278, Lng: 105.8342}, Destination: &destination,
		RequestedAt: now, Version: 1,
	}
}

func validQuote(now time.Time) Quote {
	return Quote{
		ID: "quote_1", RequestID: "req_1", DriverID: "driver_1",
		Status: QuotePending, Fare: money.Must("VND", 52_000),
		PickupDistanceM: 1200, PickupETAS: 240,
		CreatedAt: now, ExpiresAt: now.Add(time.Minute),
	}
}

func TestDriverTariffAllowsAutoQuote(t *testing.T) {
	t := DriverTariff{
		ID: "tariff_1", InstanceID: "default", DriverID: "driver_1", ServiceType: "passenger.car",
		QuoteMode: QuoteModeHybrid,
		BaseFare: money.Must("VND", 10_000), MinimumFare: money.Must("VND", 20_000),
		PerKM: money.Must("VND", 5_000), PerMinute: money.Must("VND", 0), PickupFee: money.Must("VND", 0),
		AutoQuoteMinimum: money.Must("VND", 40_000), AutoQuoteMaximum: money.Must("VND", 80_000), Version: 1,
	}
	if err := tt.Validate(); err != nil {
		t.Fatalf("tariff should be valid: %v", err)
	}
	if !tt.AllowsAutoQuote(money.Must("VND", 52_000)) {
		t.Fatal("expected quote inside driver bounds to be allowed")
	}
	if tt.AllowsAutoQuote(money.Must("VND", 90_000)) {
		t.Fatal("quote above driver maximum must be rejected")
	}
}

func TestNewAgreementSnapshotsAcceptedQuote(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	r := validRequest(now)
	q := validQuote(now)
	q.TariffID = "tariff_1"
	q.TariffVersion = 7

	agreement, err := NewAgreement("agreement_1", r, q, now.Add(5*time.Second))
	if err != nil {
		t.Fatalf("unexpected agreement error: %v", err)
	}
	if agreement.Fare.Minor != 52_000 || agreement.DriverID != "driver_1" || agreement.QuoteID != "quote_1" {
		t.Fatalf("agreement did not snapshot accepted terms: %+v", agreement)
	}
	if agreement.TermsSnapshot["tariff_version"] != int64(7) {
		t.Fatalf("expected tariff version in terms snapshot")
	}
}

func TestStateMachinesRejectUnsafeJumps(t *testing.T) {
	if ValidRequestTransition(RequestDraft, RequestAgreed) {
		t.Fatal("request must not jump from draft to agreed")
	}
	if ValidPassengerRideTransition(RideAssigned, RideCompleted) {
		t.Fatal("ride must not jump from assigned to completed")
	}
	if !ValidPassengerRideTransition(RideDriverArrived, RidePassengerOnboard) {
		t.Fatal("expected normal passenger ride transition")
	}
}
