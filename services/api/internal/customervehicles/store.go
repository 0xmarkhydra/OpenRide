package customervehicles

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound      = errors.New("customer vehicle not found")
	ErrAlreadyExists = errors.New("customer vehicle already exists")
)

type Store interface {
	Create(Vehicle) error
	Get(id string) (Vehicle, error)
	ListByOwner(ownerUserID string) ([]Vehicle, error)
}

type MemoryStore struct {
	mu       sync.RWMutex
	vehicles map[string]Vehicle
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{vehicles: make(map[string]Vehicle)} }

func (s *MemoryStore) Create(vehicle Vehicle) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.vehicles[vehicle.ID]; exists {
		return ErrAlreadyExists
	}
	for _, current := range s.vehicles {
		if current.OwnerUserID == vehicle.OwnerUserID && current.LicensePlate == vehicle.LicensePlate {
			return ErrAlreadyExists
		}
	}
	s.vehicles[vehicle.ID] = vehicle
	return nil
}

func (s *MemoryStore) Get(id string) (Vehicle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicle, ok := s.vehicles[id]
	if !ok {
		return Vehicle{}, ErrNotFound
	}
	return vehicle, nil
}

func (s *MemoryStore) ListByOwner(ownerUserID string) ([]Vehicle, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Vehicle, 0)
	for _, vehicle := range s.vehicles {
		if vehicle.OwnerUserID == ownerUserID {
			items = append(items, vehicle)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}
