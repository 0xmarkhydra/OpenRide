package marketplace

import (
	"testing"
	"time"
)

func TestValidRequestTransition(t *testing.T) {
	tests := []struct {
		name string
		from RequestStatus
		to   RequestStatus
		want bool
	}{
		{"draft opens", RequestDraft, RequestOpen, true},
		{"open receives quotes", RequestOpen, RequestReceivingQuotes, true},
		{"receiving quotes agrees", RequestReceivingQuotes, RequestAgreed, true},
		{"agreed closes", RequestAgreed, RequestClosed, true},
		{"closed cannot reopen", RequestClosed, RequestOpen, false},
		{"draft cannot jump to agreed", RequestDraft, RequestAgreed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidRequestTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf("ValidRequestTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestValidQuoteTransition(t *testing.T) {
	if !ValidQuoteTransition(QuotePending, QuoteAccepted) {
		t.Fatal("pending quote should be accept-able")
	}
	if ValidQuoteTransition(QuoteAccepted, QuoteWithdrawn) {
		t.Fatal("accepted quote must be terminal for quote state")
	}
}

func TestValidPassengerRideTransition(t *testing.T) {
	path := []RideStatus{
		RideAssigned,
		RideDriverEnRoute,
		RideDriverArrived,
		RidePassengerOnboard,
		RideInProgress,
		RideCompleted,
	}

	for i := 0; i < len(path)-1; i++ {
		if !ValidPassengerRideTransition(path[i], path[i+1]) {
			t.Fatalf("expected valid transition %q -> %q", path[i], path[i+1])
		}
	}

	if ValidPassengerRideTransition(RideAssigned, RideCompleted) {
		t.Fatal("ride must not jump from assigned to completed")
	}
}

func TestDriverTariffAllowsAutoQuote(t *testing.T) {
	tariff := DriverTariff{
		QuoteMode:         QuoteModeAuto,
		AutoQuoteMinMinor: 40_000,
		AutoQuoteMaxMinor: 80_000,
	}

	if !tariff.AllowsAutoQuote(52_000) {
		t.Fatal("quote inside driver bounds should be allowed")
	}
	if tariff.AllowsAutoQuote(30_000) {
		t.Fatal("quote below driver minimum must be rejected")
	}
	if tariff.AllowsAutoQuote(90_000) {
		t.Fatal("quote above driver maximum must be rejected")
	}

	manual := tariff
	manual.QuoteMode = QuoteModeManual
	if manual.AllowsAutoQuote(52_000) {
		t.Fatal("manual tariff must not authorize auto quote")
	}
}

func TestQuoteIsSelectable(t *testing.T) {
	now := time.Date(2026, time.September, 9, 10, 0, 0, 0, time.UTC)
	quote := Quote{
		Status:    QuotePending,
		ExpiresAt: now.Add(time.Minute),
	}
	if !quote.IsSelectable(now) {
		t.Fatal("pending unexpired quote should be selectable")
	}

	quote.ExpiresAt = now
	if quote.IsSelectable(now) {
		t.Fatal("quote at expiry time must not be selectable")
	}

	quote.ExpiresAt = now.Add(time.Minute)
	quote.Status = QuoteAccepted
	if quote.IsSelectable(now) {
		t.Fatal("non-pending quote must not be selectable")
	}
}
