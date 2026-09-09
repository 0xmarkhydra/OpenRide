package ranking

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/engine"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var ErrInvalidWeights = errors.New("ranking: weights must be non-negative and at least one must be positive")

type Weights struct {
	Fare           float64 `json:"fare"`
	PickupETA      float64 `json:"pickup_eta"`
	PickupDistance float64 `json:"pickup_distance"`
}

func (w Weights) Validate() error {
	if w.Fare < 0 || w.PickupETA < 0 || w.PickupDistance < 0 || w.Fare+w.PickupETA+w.PickupDistance <= 0 {
		return ErrInvalidWeights
	}
	return nil
}

// Weighted is the default explainable ranker. It is intentionally small and
// replaceable; communities can supply a different engine.Ranker implementation.
type Weighted struct {
	Weights Weights
}

func Default() Weighted {
	return Weighted{Weights: Weights{Fare: 0.45, PickupETA: 0.45, PickupDistance: 0.10}}
}

func (r Weighted) Rank(_ context.Context, _ marketplace.Request, quotes []marketplace.Quote) ([]engine.RankedQuote, error) {
	if err := r.Weights.Validate(); err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return []engine.RankedQuote{}, nil
	}

	minFare, maxFare := quotes[0].Fare.Minor, quotes[0].Fare.Minor
	minETA, maxETA := quotes[0].PickupETAS, quotes[0].PickupETAS
	minDistance, maxDistance := quotes[0].PickupDistanceM, quotes[0].PickupDistanceM
	for _, q := range quotes[1:] {
		minFare, maxFare = minmax(q.Fare.Minor, minFare, maxFare)
		minETA, maxETA = minmax(q.PickupETAS, minETA, maxETA)
		minDistance, maxDistance = minmax(q.PickupDistanceM, minDistance, maxDistance)
	}

	totalWeight := r.Weights.Fare + r.Weights.PickupETA + r.Weights.PickupDistance
	out := make([]engine.RankedQuote, 0, len(quotes))
	for _, q := range quotes {
		fareComponent := lowerIsBetter(q.Fare.Minor, minFare, maxFare)
		etaComponent := lowerIsBetter(q.PickupETAS, minETA, maxETA)
		distanceComponent := lowerIsBetter(q.PickupDistanceM, minDistance, maxDistance)
		score := (r.Weights.Fare*fareComponent + r.Weights.PickupETA*etaComponent + r.Weights.PickupDistance*distanceComponent) / totalWeight
		out = append(out, engine.RankedQuote{
			Quote: q,
			Score: score,
			Reasons: []string{
				fmt.Sprintf("fare %d %s", q.Fare.Minor, q.Fare.Currency),
				fmt.Sprintf("pickup ETA %ds", q.PickupETAS),
				fmt.Sprintf("pickup distance %dm", q.PickupDistanceM),
			},
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return out[i].Quote.ID < out[j].Quote.ID
		}
		return out[i].Score > out[j].Score
	})
	out[0].Recommended = true
	return out, nil
}

func minmax(value, min, max int64) (int64, int64) {
	if value < min {
		min = value
	}
	if value > max {
		max = value
	}
	return min, max
}

func lowerIsBetter(value, min, max int64) float64 {
	if max == min {
		return 1
	}
	return 1 - float64(value-min)/float64(max-min)
}
