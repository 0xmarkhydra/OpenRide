package users

import (
	"errors"
	"sort"
	"sync"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)

type Store interface {
	Create(User) error
	Get(id string) (User, error)
	GetByPhone(phone string) (User, error)
	ListAll(limit int) ([]User, error)
	Save(User) error
}

type MemoryStore struct {
	mu      sync.RWMutex
	users   map[string]User
	byPhone map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]User), byPhone: make(map[string]string)}
}

func (s *MemoryStore) Create(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[user.ID]; ok {
		return ErrAlreadyExists
	}
	if _, ok := s.byPhone[user.Phone]; ok {
		return ErrAlreadyExists
	}
	s.users[user.ID] = user
	s.byPhone[user.Phone] = user.ID
	return nil
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

func (s *MemoryStore) ListAll(limit int) ([]User, error) {
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

func (s *MemoryStore) Save(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.users[user.ID]
	if !ok {
		return ErrNotFound
	}
	if current.Phone != user.Phone {
		if otherID, exists := s.byPhone[user.Phone]; exists && otherID != user.ID {
			return ErrAlreadyExists
		}
		delete(s.byPhone, current.Phone)
		s.byPhone[user.Phone] = user.ID
	}
	s.users[user.ID] = user
	return nil
}
