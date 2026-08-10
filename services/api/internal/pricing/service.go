package pricing

import (
	"errors"
	"fmt"
	"math"

	"flashx/services/api/internal/routing"
	"flashx/services/api/internal/trips"
)

var (
	ErrUnsupportedServiceType = errors.New("unsupported service type")
	ErrInvalidRoute           = errors.New("invalid route")
	ErrRouteUnavailable       = errors.New("route provider unavailable")
)

type Estimate struct {
	ServiceType string              `json:"service_type"`
	DistanceM   int64               `json:"distance_m"`
	DurationS   int64               `json:"duration_s"`
	Fare        trips.FareBreakdown `json:"fare"`
	Currency    string              `json:"currency"`
	PricingVer  string              `json:"pricing_version"`
	RouteSource string              `json:"route_source"`
}

type rule struct {
	baseFareMinor int64
	perKMMinor    int64
}

type Service struct {
	rules   map[string]rule
	routing routing.Provider
}

func NewService() *Service {
	return NewServiceWithRouting(routing.NewFallbackProvider())
}

func NewServiceWithRouting(provider routing.Provider) *Service {
	if provider == nil {
		provider = routing.NewFallbackProvider()
	}
	return &Service{
		routing: provider,
		rules: map[string]rule{
			"bike": {baseFareMinor: 12_000, perKMMinor: 5_000},
			"car":  {baseFareMinor: 25_000, perKMMinor: 12_000},
		},
	}
}

func (s *Service) Estimate(pickup, destination trips.Point, serviceType string) (Estimate, error) {
	rule, ok := s.rules[serviceType]
	if !ok {
		return Estimate{}, ErrUnsupportedServiceType
	}

	route, err := s.routing.Route(pickup, destination, serviceType)
	if err != nil {
		switch {
		case errors.Is(err, routing.ErrInvalidRoute):
			return Estimate{}, ErrInvalidRoute
		case errors.Is(err, routing.ErrUnavailable):
			return Estimate{}, fmt.Errorf("%w: %v", ErrRouteUnavailable, err)
		default:
			return Estimate{}, fmt.Errorf("%w: %v", ErrRouteUnavailable, err)
		}
	}
	if route.DistanceM <= 0 || route.DurationS <= 0 {
		return Estimate{}, ErrInvalidRoute
	}

	distanceMinor := int64(math.Ceil(float64(route.DistanceM)/1000.0)) * rule.perKMMinor
	total := rule.baseFareMinor + distanceMinor
	return Estimate{
		ServiceType: serviceType,
		DistanceM:   route.DistanceM,
		DurationS:   route.DurationS,
		Fare: trips.FareBreakdown{
			BaseFareMinor: rule.baseFareMinor,
			DistanceMinor: distanceMinor,
			TotalMinor:    total,
		},
		Currency:    "VND",
		PricingVer:  "mvp-v1",
		RouteSource: route.Source,
	}, nil
}
