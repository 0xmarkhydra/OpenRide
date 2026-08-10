package admin

import (
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("admin not found")
	ErrDisabled = errors.New("admin disabled")
)

type Store interface {
	Bootstrap(User) (User, error)
	Get(id string) (User, error)
	GetByPhone(phone string) (User, error)
	AppendAudit(AuditEntry) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	users   map[string]User
	byPhone map[string]string
	audit   []AuditEntry
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User), byPhone: make(map[string]string)}
}

func (s *MemoryStore) Bootstrap(user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.byPhone[user.Phone]; ok {
		return s.users[id], nil
	}
	s.users[user.ID] = user
	s.byPhone[user.Phone] = user.ID
	return user, nil
}

func (s *MemoryStore) Get(id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return user, nil
}

func (s *MemoryStore) GetByPhone(phone string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byPhone[phone]
	if !ok {
		return User{}, ErrNotFound
	}
	return s.users[id], nil
}

func (s *MemoryStore) AppendAudit(entry AuditEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.audit = append(s.audit, entry)
	return nil
}
