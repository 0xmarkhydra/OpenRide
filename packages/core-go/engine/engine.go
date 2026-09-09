package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var (
	ErrNotConfigured = errors.New("engine: marketplace is not fully configured")
	ErrNoOffers      = errors.New("engine: no valid offers")
	ErrInvalidRanking = errors.New("engine: ranker returned an invalid or non-transparent result")
)

type QuoteFailure struct {
	DriverID string `json:"driver_id"`
	Reason   string `json:"reason"`
}

type OfferReport struct {
	CandidateCount int            `json:"candidate_count"`
	OfferCount     int            `json:"offer_count"`
	Offers         []RankedQuote  `json:"offers"`
	Failures       []QuoteFailure `json:"failures,omitempty"`
}

type Marketplace struct {
	Services          *extension.Registry
	Candidates        CandidateSource
	Quotes            QuoteProvider
	Ranker            Ranker
	Events            EventPublisher
	Clock             Clock
	MaxParallelQuotes int
}

func (m Marketplace) validateConfiguration() error {
	if m.Services == nil || m.Candidates == nil || m.Quotes == nil || m.Ranker == nil {
		return ErrNotConfigured
	}
	return nil
}

// FindOffers executes the marketplace pipeline without coupling Core to storage,
// transport, routing, or a particular ranking algorithm.
func (m Marketplace) FindOffers(ctx context.Context, request marketplace.Request) (OfferReport, error) {
	if err := m.validateConfiguration(); err != nil {
		return OfferReport{}, err
	}
	if err := request.Validate(); err != nil {
		return OfferReport{}, err
	}
	module, err := m.Services.Get(request.ServiceType)
	if err != nil {
		return OfferReport{}, err
	}
	if err := module.ValidateRequest(ctx, request); err != nil {
		return OfferReport{}, fmt.Errorf("engine: service validation: %w", err)
	}

	candidates, err := m.Candidates.Candidates(ctx, request)
	if err != nil {
		return OfferReport{}, fmt.Errorf("engine: discover candidates: %w", err)
	}
	report := OfferReport{CandidateCount: len(candidates)}
	if len(candidates) == 0 {
		return report, ErrNoOffers
	}

	limit := m.MaxParallelQuotes
	if limit <= 0 {
		limit = 8
	}
	if limit > len(candidates) {
		limit = len(candidates)
	}

	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	var mu sync.Mutex
	validQuotes := make([]marketplace.Quote, 0, len(candidates))

	for _, candidate := range candidates {
		candidate := candidate
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				report.Failures = append(report.Failures, QuoteFailure{DriverID: candidate.DriverID, Reason: ctx.Err().Error()})
				mu.Unlock()
				return
			}

			quote, quoteErr := m.Quotes.Quote(ctx, request, candidate)
			if quoteErr == nil {
				if quote.RequestID != request.ID || quote.DriverID != candidate.DriverID {
					quoteErr = errors.New("quote identity does not match request/candidate")
				} else if err := quote.Validate(); err != nil {
					quoteErr = err
				}
			}

			mu.Lock()
			defer mu.Unlock()
			if quoteErr != nil {
				report.Failures = append(report.Failures, QuoteFailure{DriverID: candidate.DriverID, Reason: quoteErr.Error()})
				return
			}
			validQuotes = append(validQuotes, quote)
		}()
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return report, err
	}
	if len(validQuotes) == 0 {
		return report, ErrNoOffers
	}

	ranked, err := m.Ranker.Rank(ctx, request, validQuotes)
	if err != nil {
		return report, fmt.Errorf("engine: rank quotes: %w", err)
	}
	ranked, err = canonicalizeRanking(validQuotes, ranked)
	if err != nil {
		return report, err
	}
	report.Offers = ranked
	report.OfferCount = len(ranked)

	if m.Events != nil {
		clock := m.Clock
		if clock == nil {
			clock = SystemClock{}
		}
		_ = m.Events.Publish(ctx, Event{
			Name:       "marketplace.offers_ready.v1",
			OccurredAt: clock.Now(),
			Payload: map[string]any{
				"request_id":     request.ID,
				"service_type":   request.ServiceType,
				"candidate_count": report.CandidateCount,
				"offer_count":     report.OfferCount,
			},
		})
	}

	return report, nil
}

// canonicalizeRanking makes ranking an ordering/scoring concern only. A ranker
// cannot fabricate, drop, duplicate or mutate commercial quotes. It must also
// explain every score so operator/rider UIs can surface why an offer was ranked.
func canonicalizeRanking(quotes []marketplace.Quote, ranked []RankedQuote) ([]RankedQuote, error) {
	if len(ranked) != len(quotes) || len(ranked) == 0 {
		return nil, ErrInvalidRanking
	}
	canonical := make(map[string]marketplace.Quote, len(quotes))
	for _, quote := range quotes {
		canonical[quote.ID] = quote
	}

	seen := make(map[string]struct{}, len(ranked))
	recommended := 0
	for i := range ranked {
		quoteID := ranked[i].Quote.ID
		original, ok := canonical[quoteID]
		if !ok {
			return nil, ErrInvalidRanking
		}
		if _, duplicate := seen[quoteID]; duplicate {
			return nil, ErrInvalidRanking
		}
		if math.IsNaN(ranked[i].Score) || math.IsInf(ranked[i].Score, 0) || len(ranked[i].Reasons) == 0 {
			return nil, ErrInvalidRanking
		}
		if ranked[i].Recommended {
			recommended++
			if recommended > 1 {
				return nil, ErrInvalidRanking
			}
		}
		seen[quoteID] = struct{}{}
		// Always restore the canonical quote so a ranking plugin cannot mutate
		// fare, expiry, driver identity or any accepted commercial field.
		ranked[i].Quote = original
	}
	return ranked, nil
}
