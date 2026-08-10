package ratings

import (
	"errors"
	"sync"
)

var (
	ErrNotFound      = errors.New("rating not found")
	ErrAlreadyExists = errors.New("rating already exists")
	ErrInvalidInput  = errors.New("invalid rating input")
	ErrForbidden     = errors.New("rating forbidden")
)

type Store interface {
	Create(Rating) error
	GetByTrip(tripID string) (Rating, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	byTrip map[string]Rating
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byTrip: make(map[string]Rating)}
}

func (s *MemoryStore) Create(rating Rating) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byTrip[rating.TripID]; exists {
		return ErrAlreadyExists
	}
	s.byTrip[rating.TripID] = rating
	return nil
}

func (s *MemoryStore) GetByTrip(tripID string) (Rating, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rating, ok := s.byTrip[tripID]
	if !ok {
		return Rating{}, ErrNotFound
	}
	return rating, nil
}
