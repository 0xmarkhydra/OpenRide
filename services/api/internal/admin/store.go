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
	ListAudit(limit int) ([]AuditEntry, error)
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

func (s *MemoryStore) ListAudit(limit int) ([]AuditEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	count := len(s.audit)
	if count < limit {
		limit = count
	}
	items := make([]AuditEntry, 0, limit)
	for i := count - 1; i >= 0 && len(items) < limit; i-- {
		items = append(items, s.audit[i])
	}
	return items, nil
}
