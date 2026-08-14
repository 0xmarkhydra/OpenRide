package pricing

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrRuleNotFound = errors.New("pricing rule not found")

type RuleRecord struct {
	ID             string
	ServiceType    string
	BaseFareMinor  int64
	PerKMMinor     int64
	ServiceMinor   int64
	MinimumMinor   int64
	Currency       string
	PricingVersion string
	EffectiveAt    time.Time
	UpdatedBy      string
}

type Store interface {
	GetActive(serviceType string) (RuleRecord, error)
	ListActive() ([]RuleRecord, error)
	ReplaceActive(RuleRecord) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	rules map[string]RuleRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{rules: make(map[string]RuleRecord)}
}

func (s *MemoryStore) GetActive(serviceType string) (RuleRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.rules[serviceType]
	if !ok {
		return RuleRecord{}, ErrRuleNotFound
	}
	return item, nil
}

func (s *MemoryStore) ListActive() ([]RuleRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]RuleRecord, 0, len(s.rules))
	for _, item := range s.rules {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ServiceType < items[j].ServiceType })
	return items, nil
}

func (s *MemoryStore) ReplaceActive(item RuleRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rules[item.ServiceType] = item
	return nil
}
