package ranking

import (
	"context"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

func TestWeightedCanPreferFasterPickupOverCheapestOffer(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	quotes := []marketplace.Quote{
		{ID: "cheap-slow", RequestID: "req", DriverID: "a", Status: marketplace.QuotePending, Fare: money.Must("VND", 50_000), PickupETAS: 900, PickupDistanceM: 4_000, CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
		{ID: "mid-fast", RequestID: "req", DriverID: "b", Status: marketplace.QuotePending, Fare: money.Must("VND", 55_000), PickupETAS: 120, PickupDistanceM: 500, CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
		{ID: "expensive-mid", RequestID: "req", DriverID: "c", Status: marketplace.QuotePending, Fare: money.Must("VND", 65_000), PickupETAS: 300, PickupDistanceM: 1_200, CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
	}

	ranker := Weighted{Weights: Weights{Fare: 0.35, PickupETA: 0.55, PickupDistance: 0.10}}
	ranked, err := ranker.Rank(context.Background(), marketplace.Request{}, quotes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ranked[0].Quote.ID != "mid-fast" || !ranked[0].Recommended {
		t.Fatalf("expected mid-fast to win balanced ranking, got %+v", ranked)
	}
	if len(ranked[0].Reasons) < 3 {
		t.Fatalf("expected explainable reasons, got %+v", ranked[0].Reasons)
	}
}
