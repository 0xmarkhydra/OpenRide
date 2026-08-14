package admin

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound       = errors.New("admin not found")
	ErrDisabled       = errors.New("admin disabled")
	ErrConflict       = errors.New("admin conflict")
	ErrForbidden      = errors.New("admin forbidden")
	ErrLastSuperAdmin = errors.New("cannot remove last active super admin")
)

type Store interface {
	Bootstrap(User) (User, error)
	Create(User) (User, error)
	Save(User) (User, error)
	Get(id string) (User, error)
	GetByPhone(phone string) (User, error)
	List(limit int) ([]User, error)
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

func (s *MemoryStore) Create(user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byPhone[user.Phone]; ok {
		return User{}, ErrConflict
	}
	for _, existing := range s.users {
		if existing.Email == user.Email {
			return User{}, ErrConflict
		}
	}
	s.users[user.ID] = user
	s.byPhone[user.Phone] = user.ID
	return user, nil
}

func (s *MemoryStore) Save(user User) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.users[user.ID]
	if !ok {
		return User{}, ErrNotFound
	}
	if current.Phone != user.Phone {
		return User{}, ErrConflict
	}
	for id, existing := range s.users {
		if id != user.ID && existing.Email == user.Email {
			return User{}, ErrConflict
		}
	}
	s.users[user.ID] = user
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

func (s *MemoryStore) List(limit int) ([]User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	items := make([]User, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, user)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
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
