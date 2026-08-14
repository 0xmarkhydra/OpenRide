package custodyevidence

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound      = errors.New("custody evidence not found")
	ErrAlreadyExists = errors.New("custody evidence already exists")
)

type Store interface {
	UpsertEvidence(Evidence) (Evidence, error)
	GetByTripStage(tripID string, stage Stage) (Evidence, error)
	ListByTrip(tripID string) ([]Evidence, error)
	SaveEvidence(Evidence) (Evidence, error)
	CreatePhoto(Photo) error
	ListPhotos(evidenceID string) ([]Photo, error)
}

type MemoryStore struct {
	mu          sync.RWMutex
	evidence    map[string]Evidence
	byTripStage map[string]string
	photos      map[string]Photo
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		evidence:    make(map[string]Evidence),
		byTripStage: make(map[string]string),
		photos:      make(map[string]Photo),
	}
}

func evidenceKey(tripID string, stage Stage) string { return tripID + "|" + string(stage) }

func (s *MemoryStore) UpsertEvidence(item Evidence) (Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := evidenceKey(item.TripID, item.Stage)
	if id, ok := s.byTripStage[key]; ok {
		existing := s.evidence[id]
		item.ID = existing.ID
		item.CreatedAt = existing.CreatedAt
		s.evidence[id] = item
		return item, nil
	}
	if _, exists := s.evidence[item.ID]; exists {
		return Evidence{}, ErrAlreadyExists
	}
	s.evidence[item.ID] = item
	s.byTripStage[key] = item.ID
	return item, nil
}

func (s *MemoryStore) GetByTripStage(tripID string, stage Stage) (Evidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byTripStage[evidenceKey(tripID, stage)]
	if !ok {
		return Evidence{}, ErrNotFound
	}
	return s.evidence[id], nil
}

func (s *MemoryStore) ListByTrip(tripID string) ([]Evidence, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Evidence, 0, 2)
	for _, item := range s.evidence {
		if item.TripID == tripID {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func (s *MemoryStore) SaveEvidence(item Evidence) (Evidence, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.evidence[item.ID]; !ok {
		return Evidence{}, ErrNotFound
	}
	s.evidence[item.ID] = item
	return item, nil
}

func (s *MemoryStore) CreatePhoto(photo Photo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.photos[photo.ID]; exists {
		return ErrAlreadyExists
	}
	for _, existing := range s.photos {
		if existing.ObjectKey == photo.ObjectKey {
			return ErrAlreadyExists
		}
	}
	s.photos[photo.ID] = photo
	return nil
}

func (s *MemoryStore) ListPhotos(evidenceID string) ([]Photo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Photo, 0)
	for _, photo := range s.photos {
		if photo.EvidenceID == evidenceID {
			items = append(items, photo)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}
