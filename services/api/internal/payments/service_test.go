package payments

import (
	"testing"
	"time"

	"flashx/services/api/internal/trips"
)

func TestCashPaymentLifecycle(t *testing.T) {
	service := NewService(NewMemoryStore())
	service.now = func() time.Time { return time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC) }
	trip := trips.Trip{
		ID: "trip_1", Status: trips.StatusSearching, EstimatedFareMinor: 42000, Currency: "VND",
	}
	payment, err := service.EnsureCash(trip)
	if err != nil {
		t.Fatal(err)
	}
	if payment.Status != StatusPending || payment.AmountMinor != 42000 || payment.Provider != "cash" {
		t.Fatalf("unexpected pending payment: %+v", payment)
	}

	trip.Status = trips.StatusCompleted
	trip.FinalFareMinor = 45000
	paid, err := service.MarkCashCollected(trip)
	if err != nil {
		t.Fatal(err)
	}
	if paid.Status != StatusPaid || paid.AmountMinor != 45000 {
		t.Fatalf("unexpected paid payment: %+v", paid)
	}

	again, err := service.MarkCashCollected(trip)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != paid.ID || again.Status != StatusPaid {
		t.Fatalf("cash settlement is not idempotent: %+v", again)
	}
}

func TestCannotSettleCashBeforeTripCompletes(t *testing.T) {
	service := NewService(NewMemoryStore())
	_, err := service.MarkCashCollected(trips.Trip{ID: "trip_1", Status: trips.StatusInProgress, Currency: "VND"})
	if err != ErrInvalidState {
		t.Fatalf("err=%v want=%v", err, ErrInvalidState)
	}
}
