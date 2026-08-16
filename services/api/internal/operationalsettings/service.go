package operationalsettings

import (
	"errors"
	"strings"
	"sync"
	"time"

	"flashx/services/api/internal/trips"
)

var ErrInvalidConfig = errors.New("invalid operational settings")

type Service struct {
	mu    sync.RWMutex
	store Store
	cfg   Config
	apply func(Config)
	now   func() time.Time
}

func NewService(store Store) (*Service, error) {
	if store == nil {
		store = NewMemoryStore()
	}
	s := &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
	cfg, err := store.Get()
	if errors.Is(err, ErrNotFound) {
		cfg = DefaultConfig()
		cfg, err = store.Save(cfg)
	}
	if err != nil {
		return nil, err
	}
	if !validConfig(cfg) {
		return nil, ErrInvalidConfig
	}
	s.cfg = cfg
	return s, nil
}

func DefaultConfig() Config {
	return Config{
		DesignatedDriverCarEnabled: true, DesignatedDriverBikeEnabled: true, VehicleInspectionEnabled: true,
		DispatchMaxDistanceM: 5_000, DriverLocationMaxAgeSeconds: 20, PickupGracePeriodSeconds: 600, Version: 1,
		UpdatedAt: time.Unix(0, 0).UTC(),
	}
}

func (s *Service) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *Service) IsServiceEnabled(serviceType string) bool {
	serviceType = trips.NormalizeServiceType(serviceType)
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch serviceType {
	case trips.ServiceDesignatedDriverCar:
		return s.cfg.DesignatedDriverCarEnabled
	case trips.ServiceDesignatedDriverBike:
		return s.cfg.DesignatedDriverBikeEnabled
	case trips.ServiceVehicleInspection:
		return s.cfg.VehicleInspectionEnabled
	default:
		return false
	}
}

func (s *Service) Update(input Config, actorID string) (Config, error) {
	if input.PickupGracePeriodSeconds == 0 {
		input.PickupGracePeriodSeconds = s.Get().PickupGracePeriodSeconds
	}
	input.UpdatedBy = strings.TrimSpace(actorID)
	input.UpdatedAt = s.now()
	if !validConfig(input) {
		return Config{}, ErrInvalidConfig
	}
	stored, err := s.store.Save(input)
	if err != nil {
		return Config{}, err
	}
	s.mu.Lock()
	s.cfg = stored
	apply := s.apply
	s.mu.Unlock()
	if apply != nil {
		apply(stored)
	}
	return stored, nil
}

func (s *Service) SetApplyHook(apply func(Config)) {
	s.mu.Lock()
	s.apply = apply
	cfg := s.cfg
	s.mu.Unlock()
	if apply != nil {
		apply(cfg)
	}
}

func validConfig(cfg Config) bool {
	return cfg.DispatchMaxDistanceM >= 1_000 && cfg.DispatchMaxDistanceM <= 30_000 &&
		cfg.DriverLocationMaxAgeSeconds >= 5 && cfg.DriverLocationMaxAgeSeconds <= 120 &&
		cfg.PickupGracePeriodSeconds >= 60 && cfg.PickupGracePeriodSeconds <= 3_600
}
