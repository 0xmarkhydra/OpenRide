package dispatch

import (
	"sort"
	"sync"
)

// OfferStore keeps dispatch offers outside the Engine so production can use a
// shared Redis store while unit tests keep a fast in-memory implementation.
type OfferStore interface {
	Put(Offer) error
	Get(id string) (Offer, error)
	Save(Offer) error
	ListByTrip(tripID string) ([]Offer, error)
	ListByDriver(driverID string) ([]Offer, error)
}

type MemoryOfferStore struct {
	mu     sync.RWMutex
	offers map[string]Offer
}

func NewMemoryOfferStore() *MemoryOfferStore {
	return &MemoryOfferStore{offers: make(map[string]Offer)}
}

func (s *MemoryOfferStore) Put(offer Offer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.offers[offer.ID]; exists {
		return ErrOfferUnavailable
	}
	s.offers[offer.ID] = offer
	return nil
}

func (s *MemoryOfferStore) Get(id string) (Offer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	offer, ok := s.offers[id]
	if !ok {
		return Offer{}, ErrOfferNotFound
	}
	return offer, nil
}

func (s *MemoryOfferStore) Save(offer Offer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.offers[offer.ID]; !ok {
		return ErrOfferNotFound
	}
	s.offers[offer.ID] = offer
	return nil
}

func (s *MemoryOfferStore) ListByTrip(tripID string) ([]Offer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Offer, 0)
	for _, offer := range s.offers {
		if offer.TripID == tripID {
			items = append(items, offer)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func (s *MemoryOfferStore) ListByDriver(driverID string) ([]Offer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Offer, 0)
	for _, offer := range s.offers {
		if offer.DriverID == driverID {
			items = append(items, offer)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}
