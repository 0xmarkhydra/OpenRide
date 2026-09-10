package engine

import (
	"context"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

type mutatingPersistedRanker struct{}

func (mutatingPersistedRanker) Rank(_ context.Context, _ marketplace.Request, quotes []marketplace.Quote) ([]RankedQuote, error) {
	quotes[0].Fare.Minor = 1
	quotes[0].Metadata["nested"].(map[string]any)["secret"] = "changed"
	return []RankedQuote{
		{Quote: quotes[0], Score: 1, Reasons: []string{"first"}, Recommended: true},
		{Quote: quotes[1], Score: 0.5, Reasons: []string{"second"}},
	}, nil
}

func TestRankQuotesRestoresCanonicalCommercialData(t *testing.T) {
	now := time.Now().UTC()
	req := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1",
		ServiceType: marketplace.ServiceType("passenger.car"), Status: marketplace.RequestOpen,
		Pickup: geo.Point{Lat: 10.7, Lng: 106.6}, RequestedAt: now, Version: 1,
	}
	quotes := []marketplace.Quote{
		{ID: "q1", RequestID: req.ID, DriverID: "d1", Status: marketplace.QuotePending, Fare: money.Must("VND", 50_000), Metadata: map[string]any{"nested": map[string]any{"secret": "original"}}, CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
		{ID: "q2", RequestID: req.ID, DriverID: "d2", Status: marketplace.QuotePending, Fare: money.Must("VND", 60_000), CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
	}

	ranked, err := RankQuotes(context.Background(), req, quotes, mutatingPersistedRanker{})
	if err != nil { t.Fatal(err) }
	if ranked[0].Quote.Fare.Minor != 50_000 { t.Fatalf("ranker mutated canonical fare: %d", ranked[0].Quote.Fare.Minor) }
	if got := ranked[0].Quote.Metadata["nested"].(map[string]any)["secret"]; got != "original" { t.Fatalf("ranker mutated nested canonical metadata: %v", got) }
	if got := quotes[0].Metadata["nested"].(map[string]any)["secret"]; got != "original" { t.Fatalf("caller quote mutated: %v", got) }
}
