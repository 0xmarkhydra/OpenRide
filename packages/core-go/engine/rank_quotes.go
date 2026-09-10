package engine

import (
	"context"
	"fmt"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

// RankQuotes applies an explainable Ranker to already-authorized, persisted
// quotes while preserving Core's canonical quote integrity guarantees. The
// ranker receives detached copies and cannot drop, fabricate, duplicate or
// mutate the commercial quote data returned to callers.
func RankQuotes(ctx context.Context, request marketplace.Request, quotes []marketplace.Quote, ranker Ranker) ([]RankedQuote, error) {
	if ranker == nil {
		return nil, ErrNotConfigured
	}
	if err := request.Validate(); err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return []RankedQuote{}, nil
	}
	for _, quote := range quotes {
		if quote.RequestID != request.ID {
			return nil, fmt.Errorf("engine: quote %s does not belong to request %s", quote.ID, request.ID)
		}
		if err := quote.Validate(); err != nil {
			return nil, fmt.Errorf("engine: invalid quote %s: %w", quote.ID, err)
		}
	}

	canonicalQuotes := cloneQuotes(quotes)
	rankInput := cloneQuotes(quotes)
	ranked, err := ranker.Rank(ctx, request, rankInput)
	if err != nil {
		return nil, fmt.Errorf("engine: rank quotes: %w", err)
	}
	return canonicalizeRanking(canonicalQuotes, ranked)
}
