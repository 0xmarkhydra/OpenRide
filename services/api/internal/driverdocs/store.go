package driverdocs

import (
	"errors"
	"sync"
)

var (
	ErrNotFound      = errors.New("driver document not found")
	ErrAlreadyExists = errors.New("driver document already exists")
)

type Store interface {
	Create(Document) error
	Get(id string) (Document, error)
	ListForDriver(driverID string) ([]Document, error)
	Review(id string, status ReviewStatus, note string) (Document, error)
}

type MemoryStore struct {
	mu   sync.RWMutex
	docs map[string]Document
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{docs: make(map[string]Document)}
}

func (s *MemoryStore) Create(doc Document) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.docs[doc.ID]; exists {
		return ErrAlreadyExists
	}
	for _, existing := range s.docs {
		if existing.ObjectKey == doc.ObjectKey {
			return ErrAlreadyExists
		}
	}
	s.docs[doc.ID] = doc
	return nil
}

func (s *MemoryStore) Get(id string) (Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doc, ok := s.docs[id]
	if !ok {
		return Document{}, ErrNotFound
	}
	return doc, nil
}

func (s *MemoryStore) ListForDriver(driverID string) ([]Document, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Document, 0)
	for _, doc := range s.docs {
		if doc.DriverID == driverID {
			result = append(result, doc)
		}
	}
	return result, nil
}

func (s *MemoryStore) Review(id string, status ReviewStatus, note string) (Document, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	doc, ok := s.docs[id]
	if !ok {
		return Document{}, ErrNotFound
	}
	doc.ReviewStatus = status
	doc.ReviewNote = note
	s.docs[id] = doc
	return doc, nil
}
