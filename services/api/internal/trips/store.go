package trips

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound        = errors.New("trip not found")
	ErrVersionConflict = errors.New("trip version conflict")
)

type Store interface {
	Create(Trip) error
	Get(id string) (Trip, error)
	Save(Trip) error
	ListByRider(riderID string, limit int) ([]Trip, error)
	ListByDriver(driverID string, limit int) ([]Trip, error)
	ListSearching(limit int) ([]Trip, error)
	ListAll(limit int) ([]Trip, error)
	FindActiveByDriver(driverID string) (Trip, error)
}

type MemoryStore struct {
	mu    sync.RWMutex
	trips map[string]Trip
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{trips: make(map[string]Trip)}
}

func (s *MemoryStore) Create(trip Trip) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.trips[trip.ID]; exists {
		return errors.New("trip already exists")
	}
	s.trips[trip.ID] = trip
	return nil
}

func (s *MemoryStore) Get(id string) (Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	trip, ok := s.trips[id]
	if !ok {
		return Trip{}, ErrNotFound
	}
	return trip, nil
}

func (s *MemoryStore) Save(trip Trip) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.trips[trip.ID]; !ok {
		return ErrNotFound
	}
	trip.Version++
	s.trips[trip.ID] = trip
	return nil
}

func (s *MemoryStore) ListByRider(riderID string, limit int) ([]Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	result := make([]Trip, 0)
	for _, trip := range s.trips {
		if trip.RiderID == riderID {
			result = append(result, trip)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *MemoryStore) ListByDriver(driverID string, limit int) ([]Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	result := make([]Trip, 0)
	for _, trip := range s.trips {
		if trip.DriverID == driverID {
			result = append(result, trip)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *MemoryStore) ListSearching(limit int) ([]Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	result := make([]Trip, 0)
	for _, trip := range s.trips {
		if trip.Status == StatusSearching && trip.DriverID == "" {
			result = append(result, trip)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (s *MemoryStore) ListAll(limit int) ([]Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 500 { limit = 100 }
	result := make([]Trip, 0, len(s.trips))
	for _, trip := range s.trips { result = append(result, trip) }
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.After(result[j].CreatedAt) })
	if len(result) > limit { result = result[:limit] }
	return result, nil
}

func (s *MemoryStore) FindActiveByDriver(driverID string) (Trip, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, trip := range s.trips {
		if trip.DriverID == driverID && (trip.Status == StatusAccepted || trip.Status == StatusArriving || trip.Status == StatusArrived || trip.Status == StatusInProgress) {
			return trip, nil
		}
	}
	return Trip{}, ErrNotFound
}
