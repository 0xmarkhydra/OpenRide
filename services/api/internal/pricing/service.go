package pricing

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"flashx/services/api/internal/routing"
	"flashx/services/api/internal/trips"
)

var (
	ErrUnsupportedServiceType = errors.New("unsupported service type")
	ErrInvalidRoute           = errors.New("invalid route")
	ErrRouteUnavailable       = errors.New("route provider unavailable")
	ErrInvalidRule            = errors.New("invalid pricing rule")
)

const maxPricingAmountMinor int64 = 100_000_000

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
	version       string
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

type UpdateRuleInput struct {
	ServiceType   string `json:"service_type"`
	BaseFareMinor int64  `json:"base_fare_minor"`
	PerKMMinor    int64  `json:"per_km_minor"`
	ServiceMinor  int64  `json:"service_minor"`
	MinimumMinor  int64  `json:"minimum_minor"`
	ActorID       string `json:"-"`
}

type Service struct {
	mu      sync.RWMutex
	rules   map[string]rule
	routing routing.Provider
	store   Store
	now     func() time.Time
}

func NewService() *Service { return NewServiceWithRouting(routing.NewFallbackProvider()) }

func NewServiceWithRouting(provider routing.Provider) *Service {
	service, err := NewServiceWithStore(provider, NewMemoryStore())
	if err != nil {
		panic(err)
	}
	return service
}

func NewServiceWithStore(provider routing.Provider, store Store) (*Service, error) {
	if provider == nil {
		provider = routing.NewFallbackProvider()
	}
	if store == nil {
		store = NewMemoryStore()
	}
	s := &Service{
		routing: provider,
		store:   store,
		rules:   make(map[string]rule),
		now:     func() time.Time { return time.Now().UTC() },
	}
	for _, defaultRule := range defaultRules() {
		record, err := store.GetActive(defaultRule.ServiceType)
		if errors.Is(err, ErrRuleNotFound) {
			if err := store.ReplaceActive(defaultRule); err != nil {
				return nil, fmt.Errorf("seed pricing %s: %w", defaultRule.ServiceType, err)
			}
			record = defaultRule
		} else if err != nil {
			return nil, fmt.Errorf("load pricing %s: %w", defaultRule.ServiceType, err)
		}
		s.rules[record.ServiceType] = ruleFromRecord(record)
	}
	return s, nil
}

func defaultRules() []RuleRecord {
	now := time.Unix(0, 0).UTC()
	return []RuleRecord{
		{ID: "pricing_designated_car_default", ServiceType: trips.ServiceDesignatedDriverCar, BaseFareMinor: 90_000, PerKMMinor: 18_000, MinimumMinor: 120_000, Currency: "VND", PricingVersion: "marketplace-demo-v1", EffectiveAt: now},
		{ID: "pricing_designated_bike_default", ServiceType: trips.ServiceDesignatedDriverBike, BaseFareMinor: 45_000, PerKMMinor: 9_000, MinimumMinor: 60_000, Currency: "VND", PricingVersion: "marketplace-demo-v1", EffectiveAt: now},
		{ID: "pricing_inspection_default", ServiceType: trips.ServiceVehicleInspection, PerKMMinor: 5_000, ServiceMinor: 299_000, MinimumMinor: 299_000, Currency: "VND", PricingVersion: "marketplace-demo-v1", EffectiveAt: now},
	}
}

func (s *Service) Rules() []RuleSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order := []string{trips.ServiceDesignatedDriverCar, trips.ServiceDesignatedDriverBike, trips.ServiceVehicleInspection}
	items := make([]RuleSnapshot, 0, len(order))
	for _, serviceType := range order {
		current, ok := s.rules[serviceType]
		if !ok {
			continue
		}
		items = append(items, snapshot(serviceType, current))
	}
	return items
}

func (s *Service) UpdateRule(input UpdateRuleInput) (RuleSnapshot, error) {
	serviceType := trips.NormalizeServiceType(strings.TrimSpace(input.ServiceType))
	if !trips.IsSupportedService(serviceType) {
		return RuleSnapshot{}, ErrUnsupportedServiceType
	}
	amounts := []int64{input.BaseFareMinor, input.PerKMMinor, input.ServiceMinor, input.MinimumMinor}
	for _, amount := range amounts {
		if amount < 0 || amount > maxPricingAmountMinor {
			return RuleSnapshot{}, ErrInvalidRule
		}
	}
	if input.BaseFareMinor == 0 && input.PerKMMinor == 0 && input.ServiceMinor == 0 && input.MinimumMinor == 0 {
		return RuleSnapshot{}, ErrInvalidRule
	}

	now := s.now()
	version := fmt.Sprintf("admin-%s-%d", serviceType, now.UnixNano())
	record := RuleRecord{
		ID:             version,
		ServiceType:    serviceType,
		BaseFareMinor:  input.BaseFareMinor,
		PerKMMinor:     input.PerKMMinor,
		ServiceMinor:   input.ServiceMinor,
		MinimumMinor:   input.MinimumMinor,
		Currency:       "VND",
		PricingVersion: version,
		EffectiveAt:    now,
		UpdatedBy:      strings.TrimSpace(input.ActorID),
	}
	if err := s.store.ReplaceActive(record); err != nil {
		return RuleSnapshot{}, err
	}
	current := ruleFromRecord(record)
	s.mu.Lock()
	s.rules[serviceType] = current
	s.mu.Unlock()
	return snapshot(serviceType, current), nil
}

func (s *Service) Estimate(pickup, destination trips.Point, serviceType string) (Estimate, error) {
	serviceType = trips.NormalizeServiceType(serviceType)
	s.mu.RLock()
	current, ok := s.rules[serviceType]
	s.mu.RUnlock()
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

	distanceMinor := int64(math.Ceil(float64(route.DistanceM)/1000.0)) * current.perKMMinor
	total := current.baseFareMinor + current.serviceMinor + distanceMinor
	if total < current.minimumMinor {
		total = current.minimumMinor
	}
	return Estimate{
		ServiceType: serviceType,
		DistanceM:   route.DistanceM,
		DurationS:   route.DurationS,
		Fare: trips.FareBreakdown{
			BaseFareMinor: current.baseFareMinor,
			DistanceMinor: distanceMinor,
			ServiceMinor:  current.serviceMinor,
			TotalMinor:    total,
		},
		Currency: "VND", PricingVer: current.version, RouteSource: route.Source,
	}, nil
}

func ruleFromRecord(item RuleRecord) rule {
	return rule{
		baseFareMinor: item.BaseFareMinor,
		perKMMinor:    item.PerKMMinor,
		serviceMinor:  item.ServiceMinor,
		minimumMinor:  item.MinimumMinor,
		version:       item.PricingVersion,
	}
}

func snapshot(serviceType string, current rule) RuleSnapshot {
	return RuleSnapshot{
		ServiceType: serviceType, BaseFareMinor: current.baseFareMinor, PerKMMinor: current.perKMMinor,
		ServiceMinor: current.serviceMinor, MinimumMinor: current.minimumMinor, Currency: "VND",
		PricingVersion: current.version,
	}
}
