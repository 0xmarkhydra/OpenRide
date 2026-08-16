package operationalsettings

import (
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("operational settings not found")

type Config struct {
	DesignatedDriverCarEnabled  bool      `json:"designated_driver_car_enabled"`
	DesignatedDriverBikeEnabled bool      `json:"designated_driver_bike_enabled"`
	VehicleInspectionEnabled    bool      `json:"vehicle_inspection_assist_enabled"`
	DispatchMaxDistanceM        int64     `json:"dispatch_max_distance_m"`
	DriverLocationMaxAgeSeconds int       `json:"driver_location_max_age_seconds"`
	PickupGracePeriodSeconds    int       `json:"pickup_grace_period_seconds"`
	Version                     int64     `json:"version"`
	UpdatedBy                   string    `json:"updated_by,omitempty"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

type Store interface {
	Get() (Config, error)
	Save(Config) (Config, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	config *Config
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{} }

func (s *MemoryStore) Get() (Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.config == nil {
		return Config{}, ErrNotFound
	}
	return *s.config, nil
}

func (s *MemoryStore) Save(config Config) (Config, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.config != nil {
		config.Version = s.config.Version + 1
	} else if config.Version <= 0 {
		config.Version = 1
	}
	s.config = &config
	return config, nil
}
