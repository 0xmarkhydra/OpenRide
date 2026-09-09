package engine

import (
	"context"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

// Candidate is provider-neutral driver eligibility data returned by discovery.
type Candidate struct {
	DriverID        string         `json:"driver_id"`
	DriverVehicleID string         `json:"driver_vehicle_id,omitempty"`
	PickupDistanceM int64          `json:"pickup_distance_m"`
	PickupETAS      int64          `json:"pickup_eta_s"`
	Metadata        map[string]any `json:"metadata,omitempty"`
}

// CandidateSource can be backed by Redis GEO, PostGIS, an external dispatch
// provider, or a custom community implementation.
type CandidateSource interface {
	Candidates(ctx context.Context, request marketplace.Request) ([]Candidate, error)
}

// QuoteProvider turns an eligible candidate into a driver-authorized quote.
// Implementations may use manual, automatic, or hybrid driver tariff flows.
type QuoteProvider interface {
	Quote(ctx context.Context, request marketplace.Request, candidate Candidate) (marketplace.Quote, error)
}

type RankedQuote struct {
	Quote       marketplace.Quote `json:"quote"`
	Score       float64           `json:"score"`
	Reasons     []string          `json:"reasons,omitempty"`
	Recommended bool              `json:"recommended"`
}

// Ranker must return explainable ordering. Core does not mandate one formula.
type Ranker interface {
	Rank(ctx context.Context, request marketplace.Request, quotes []marketplace.Quote) ([]RankedQuote, error)
}

type Event struct {
	Name       string         `json:"name"`
	OccurredAt time.Time      `json:"occurred_at"`
	Payload    map[string]any `json:"payload,omitempty"`
}

type EventPublisher interface {
	Publish(ctx context.Context, event Event) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }
