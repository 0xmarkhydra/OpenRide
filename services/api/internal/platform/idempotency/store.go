package idempotency

import (
	"errors"
	"sync"
)

var ErrConflict = errors.New("idempotency key reused with different request")

type Record struct {
	Fingerprint string
	ResourceID  string
}

type Store interface {
	Get(scope, key string) (Record, bool, error)
	Put(scope, key string, record Record) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: make(map[string]Record)}
}

func (s *MemoryStore) Get(scope, key string) (Record, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[scope+":"+key]
	return record, ok, nil
}

func (s *MemoryStore) Put(scope, key string, record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	compound := scope + ":" + key
	if current, ok := s.records[compound]; ok {
		if current.Fingerprint != record.Fingerprint {
			return ErrConflict
		}
		return nil
	}
	s.records[compound] = record
	return nil
}
