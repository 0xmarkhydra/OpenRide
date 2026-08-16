package notifications

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound     = errors.New("notification resource not found")
	ErrInvalidInput = errors.New("invalid notification input")
)

type Store interface {
	UpsertDevice(Device) (Device, error)
	DisableDevice(actorID string, role Role, deviceID string) error
	ListDevices(actorID string, role Role) ([]Device, error)
	Enqueue(Message) error
	ListPending(limit int) ([]Message, error)
	SaveMessage(Message) error
}

type MemoryStore struct {
	mu       sync.RWMutex
	devices  map[string]Device
	messages map[string]Message
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{devices: make(map[string]Device), messages: make(map[string]Message)}
}

func (s *MemoryStore) UpsertDevice(device Device) (Device, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, current := range s.devices {
		if current.Token == device.Token {
			device.ID = id
			device.CreatedAt = current.CreatedAt
			s.devices[id] = device
			return device, nil
		}
	}
	s.devices[device.ID] = device
	return device, nil
}

func (s *MemoryStore) DisableDevice(actorID string, role Role, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	device, ok := s.devices[deviceID]
	if !ok || device.ActorID != actorID || device.ActorRole != role {
		return ErrNotFound
	}
	device.Enabled = false
	s.devices[deviceID] = device
	return nil
}

func (s *MemoryStore) ListDevices(actorID string, role Role) ([]Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]Device, 0)
	for _, device := range s.devices {
		if device.ActorID == actorID && device.ActorRole == role && device.Enabled {
			items = append(items, device)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
	return items, nil
}

func (s *MemoryStore) Enqueue(message Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages[message.ID] = message
	return nil
}

func (s *MemoryStore) ListPending(limit int) ([]Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	items := make([]Message, 0)
	for _, message := range s.messages {
		if message.Status == StatusPending {
			items = append(items, message)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func (s *MemoryStore) SaveMessage(message Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.messages[message.ID]; !ok {
		return ErrNotFound
	}
	s.messages[message.ID] = message
	return nil
}
