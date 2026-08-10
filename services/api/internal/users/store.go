package users

import (
	"errors"
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
