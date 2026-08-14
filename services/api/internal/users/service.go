package users

import (
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
)

var ErrInvalidInput = errors.New("invalid user input")

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Get(id string) (User, error) {
	return s.store.Get(id)
}

func (s *Service) ListAll(limit int) ([]User, error) {
	return s.store.ListAll(limit)
}

func (s *Service) FindOrCreateByPhone(phone string) (User, bool, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return User{}, false, ErrInvalidInput
	}
	user, err := s.store.GetByPhone(phone)
	if err == nil {
		return user, false, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, false, err
	}
	now := s.now()
	user = User{
		ID:        ids.New("usr"),
		Phone:     phone,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.store.Create(user); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			existing, getErr := s.store.GetByPhone(phone)
			return existing, false, getErr
		}
		return User{}, false, err
	}
	return user, true, nil
}

func (s *Service) UpdateProfile(id, fullName string) (User, error) {
	user, err := s.store.Get(id)
	if err != nil {
		return User{}, err
	}
	fullName = strings.TrimSpace(fullName)
	if len(fullName) > 160 {
		return User{}, ErrInvalidInput
	}
	user.FullName = fullName
	user.UpdatedAt = s.now()
	if err := s.store.Save(user); err != nil {
		return User{}, err
	}
	return s.store.Get(id)
}
