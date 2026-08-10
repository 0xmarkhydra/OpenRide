package payments

import (
	"errors"
	"sync"
)

var (
	ErrNotFound      = errors.New("payment not found")
	ErrAlreadyExists = errors.New("payment already exists")
	ErrInvalidState  = errors.New("invalid payment state")
)

type Store interface {
	Create(Payment) error
	GetByTrip(tripID string) (Payment, error)
	Save(Payment) error
}

type MemoryStore struct {
	mu       sync.RWMutex
	byTripID map[string]Payment
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byTripID: make(map[string]Payment)}
}

func (s *MemoryStore) Create(payment Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byTripID[payment.TripID]; exists {
		return ErrAlreadyExists
	}
	s.byTripID[payment.TripID] = payment
	return nil
}

func (s *MemoryStore) GetByTrip(tripID string) (Payment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	payment, ok := s.byTripID[tripID]
	if !ok {
		return Payment{}, ErrNotFound
	}
	return payment, nil
}

func (s *MemoryStore) Save(payment Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byTripID[payment.TripID]; !ok {
		return ErrNotFound
	}
	s.byTripID[payment.TripID] = payment
	return nil
}
