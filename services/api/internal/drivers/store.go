package drivers

import (
	"errors"
	"sync"
)

var (
	ErrNotFound      = errors.New("driver not found")
	ErrAlreadyExists = errors.New("driver already exists")
)

type Store interface {
	Create(Driver) error
	Get(id string) (Driver, error)
	GetByPhone(phone string) (Driver, error)
	Save(Driver) error
	TryMarkBusy(id string) (Driver, error)
	List() ([]Driver, error)
}

type MemoryStore struct {
	mu      sync.RWMutex
	drivers map[string]Driver
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{drivers: make(map[string]Driver)} }

func (s *MemoryStore) Create(driver Driver) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.drivers[driver.ID]; exists { return ErrAlreadyExists }
	if driver.Phone != "" {
		for _, existing := range s.drivers {
			if existing.Phone == driver.Phone { return ErrAlreadyExists }
		}
	}
	s.drivers[driver.ID] = driver
	return nil
}

func (s *MemoryStore) Get(id string) (Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	driver, ok := s.drivers[id]
	if !ok { return Driver{}, ErrNotFound }
	return driver, nil
}

func (s *MemoryStore) GetByPhone(phone string) (Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, driver := range s.drivers {
		if driver.Phone == phone { return driver, nil }
	}
	return Driver{}, ErrNotFound
}

func (s *MemoryStore) Save(driver Driver) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.drivers[driver.ID]; !exists { return ErrNotFound }
	driver.Version++
	s.drivers[driver.ID] = driver
	return nil
}

func (s *MemoryStore) TryMarkBusy(id string) (Driver, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	driver, ok := s.drivers[id]
	if !ok { return Driver{}, ErrNotFound }
	if driver.Availability != AvailabilityOnline { return Driver{}, ErrDriverUnavailable }
	driver.Availability = AvailabilityBusy
	driver.Version++
	s.drivers[id] = driver
	return driver, nil
}

func (s *MemoryStore) List() ([]Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Driver, 0, len(s.drivers))
	for _, driver := range s.drivers { result = append(result, driver) }
	return result, nil
}
