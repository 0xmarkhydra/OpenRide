package engine

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

type testService struct{}

func (testService) Manifest() extension.Manifest {
	return extension.Manifest{ID: "passenger.car", Version: "1.0.0", DisplayName: "Car", Category: "passenger"}
}
func (testService) ValidateRequest(context.Context, marketplace.Request) error { return nil }

type testCandidates struct{}

func (testCandidates) Candidates(context.Context, marketplace.Request) ([]Candidate, error) {
	return []Candidate{
		{DriverID: "driver_a", PickupDistanceM: 800, PickupETAS: 180},
		{DriverID: "driver_b", PickupDistanceM: 500, PickupETAS: 120},
		{DriverID: "driver_bad", PickupDistanceM: 300, PickupETAS: 90},
	}, nil
}

type testQuoter struct{ now time.Time }

func (q testQuoter) Quote(_ context.Context, request marketplace.Request, candidate Candidate) (marketplace.Quote, error) {
	if candidate.DriverID == "driver_bad" {
		return marketplace.Quote{}, errors.New("driver tariff unavailable")
	}
	fare := int64(60_000)
	if candidate.DriverID == "driver_a" {
		fare = 52_000
	}
	return marketplace.Quote{
		ID: "quote_" + candidate.DriverID, RequestID: request.ID, DriverID: candidate.DriverID,
		Status: marketplace.QuotePending, Fare: money.Must("VND", fare),
		PickupDistanceM: candidate.PickupDistanceM, PickupETAS: candidate.PickupETAS,
		Metadata: map[string]any{"pricing": map[string]any{"source": "driver_tariff"}},
		CreatedAt: q.now, ExpiresAt: q.now.Add(time.Minute),
	}, nil
}

type testRanker struct{}

func (testRanker) Rank(_ context.Context, _ marketplace.Request, quotes []marketplace.Quote) ([]RankedQuote, error) {
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].Fare.Minor < quotes[j].Fare.Minor })
	out := make([]RankedQuote, 0, len(quotes))
	for i, q := range quotes {
		out = append(out, RankedQuote{Quote: q, Score: float64(len(quotes) - i), Reasons: []string{"demo: lower fare"}, Recommended: i == 0})
	}
	return out, nil
}

type mutatingRanker struct{}

func (mutatingRanker) Rank(_ context.Context, _ marketplace.Request, quotes []marketplace.Quote) ([]RankedQuote, error) {
	quotes[0].Fare = money.Must("VND", 1)
	quotes[0].Metadata["pricing"].(map[string]any)["source"] = "tampered"
	return []RankedQuote{
		{Quote: quotes[0], Score: 1, Reasons: []string{"tampered"}, Recommended: true},
		{Quote: quotes[1], Score: 0.5, Reasons: []string{"other"}},
	}, nil
}

func TestFindOffersKeepsHealthyDriversWhenOneQuoteFails(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	destination := geo.Point{Lat: 21.0285, Lng: 105.8542}
	request := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car",
		Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 21.0278, Lng: 105.8342}, Destination: &destination,
		RequestedAt: now, Version: 1,
	}
	registry := extension.MustRegistry(testService{})
	engine := Marketplace{Services: registry, Candidates: testCandidates{}, Quotes: testQuoter{now: now}, Ranker: testRanker{}, MaxParallelQuotes: 2}

	report, err := engine.FindOffers(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.CandidateCount != 3 || report.OfferCount != 2 || len(report.Failures) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	if !report.Offers[0].Recommended || report.Offers[0].Quote.DriverID != "driver_a" {
		t.Fatalf("unexpected recommendation: %+v", report.Offers)
	}
}

func TestFindOffersIsolatesRankerFromCanonicalQuotes(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	destination := geo.Point{Lat: 21.0285, Lng: 105.8542}
	request := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car",
		Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 21.0278, Lng: 105.8342}, Destination: &destination,
		RequestedAt: now, Version: 1,
	}
	registry := extension.MustRegistry(testService{})
	engine := Marketplace{Services: registry, Candidates: testCandidates{}, Quotes: testQuoter{now: now}, Ranker: mutatingRanker{}, MaxParallelQuotes: 2}

	report, err := engine.FindOffers(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Offers[0].Quote.Fare.Minor == 1 {
		t.Fatal("ranker must not be able to mutate canonical fare")
	}
	pricing := report.Offers[0].Quote.Metadata["pricing"].(map[string]any)
	if pricing["source"] != "driver_tariff" {
		t.Fatal("ranker must not be able to mutate nested canonical metadata")
	}
}

func TestCanonicalizeRankingProtectsCommercialQuotes(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	original := []marketplace.Quote{
		{ID: "q1", RequestID: "r1", DriverID: "d1", Status: marketplace.QuotePending, Fare: money.Must("VND", 50_000), CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
		{ID: "q2", RequestID: "r1", DriverID: "d2", Status: marketplace.QuotePending, Fare: money.Must("VND", 60_000), CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
	}
	mutated := original[0]
	mutated.Fare = money.Must("VND", 1)
	ranked := []RankedQuote{
		{Quote: mutated, Score: 0.9, Reasons: []string{"custom score"}, Recommended: true},
		{Quote: original[1], Score: 0.8, Reasons: []string{"custom score"}},
	}

	canonical, err := canonicalizeRanking(original, ranked)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canonical[0].Quote.Fare.Minor != 50_000 {
		t.Fatalf("ranker mutated canonical fare: %+v", canonical[0].Quote.Fare)
	}
}

func TestCanonicalizeRankingRejectsHiddenOrUnexplainedOffers(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	original := []marketplace.Quote{
		{ID: "q1", RequestID: "r1", DriverID: "d1", Status: marketplace.QuotePending, Fare: money.Must("VND", 50_000), CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
		{ID: "q2", RequestID: "r1", DriverID: "d2", Status: marketplace.QuotePending, Fare: money.Must("VND", 60_000), CreatedAt: now, ExpiresAt: now.Add(time.Minute)},
	}
	if _, err := canonicalizeRanking(original, []RankedQuote{{Quote: original[0], Score: 1, Reasons: []string{"only one returned"}}}); !errors.Is(err, ErrInvalidRanking) {
		t.Fatalf("expected hidden offer to be rejected, got %v", err)
	}
	unexplained := []RankedQuote{{Quote: original[0], Score: 1}, {Quote: original[1], Score: 0.5, Reasons: []string{"ok"}}}
	if _, err := canonicalizeRanking(original, unexplained); !errors.Is(err, ErrInvalidRanking) {
		t.Fatalf("expected unexplained ranking to be rejected, got %v", err)
	}
}
