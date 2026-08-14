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
	serviceMinor  int64
	minimumMinor  int64
}

type RuleSnapshot struct {
	ServiceType    string `json:"service_type"`
	BaseFareMinor  int64  `json:"base_fare_minor"`
	PerKMMinor     int64  `json:"per_km_minor"`
	ServiceMinor   int64  `json:"service_minor"`
	MinimumMinor   int64  `json:"minimum_minor"`
	Currency       string `json:"currency"`
	PricingVersion string `json:"pricing_version"`
}

type Service struct {
	rules   map[string]rule
	routing routing.Provider
}

func NewService() *Service { return NewServiceWithRouting(routing.NewFallbackProvider()) }

func NewServiceWithRouting(provider routing.Provider) *Service {
	if provider == nil {
		provider = routing.NewFallbackProvider()
	}
	return &Service{
		routing: provider,
		rules: map[string]rule{
			trips.ServiceDesignatedDriverBike: {baseFareMinor: 45_000, perKMMinor: 9_000, minimumMinor: 60_000},
			trips.ServiceDesignatedDriverCar:  {baseFareMinor: 90_000, perKMMinor: 18_000, minimumMinor: 120_000},
			trips.ServiceVehicleInspection:    {baseFareMinor: 0, perKMMinor: 5_000, serviceMinor: 299_000, minimumMinor: 299_000},
		},
	}
}

func (s *Service) Rules() []RuleSnapshot {
	order := []string{trips.ServiceDesignatedDriverCar, trips.ServiceDesignatedDriverBike, trips.ServiceVehicleInspection}
	items := make([]RuleSnapshot, 0, len(order))
	for _, serviceType := range order {
		rule, ok := s.rules[serviceType]
		if !ok {
			continue
		}
		items = append(items, RuleSnapshot{
			ServiceType: serviceType, BaseFareMinor: rule.baseFareMinor, PerKMMinor: rule.perKMMinor,
			ServiceMinor: rule.serviceMinor, MinimumMinor: rule.minimumMinor, Currency: "VND",
			PricingVersion: "marketplace-demo-v1",
		})
	}
	return items
}

func (s *Service) Estimate(pickup, destination trips.Point, serviceType string) (Estimate, error) {
	serviceType = trips.NormalizeServiceType(serviceType)
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
	total := rule.baseFareMinor + rule.serviceMinor + distanceMinor
	if total < rule.minimumMinor {
		total = rule.minimumMinor
	}
	return Estimate{
		ServiceType: serviceType,
		DistanceM:   route.DistanceM,
		DurationS:   route.DurationS,
		Fare: trips.FareBreakdown{
			BaseFareMinor: rule.baseFareMinor,
			DistanceMinor: distanceMinor,
			ServiceMinor:  rule.serviceMinor,
			TotalMinor:    total,
		},
		Currency: "VND", PricingVer: "marketplace-demo-v1", RouteSource: route.Source,
	}, nil
}
