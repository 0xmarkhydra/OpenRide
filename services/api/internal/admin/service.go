package admin

import (
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/platform/ids"
)

var ErrInvalidInput = errors.New("invalid admin input")

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Bootstrap(phone, email, displayName string) (User, error) {
	phone, err := auth.NormalizeVietnamPhone(phone)
	if err != nil {
		return User{}, ErrInvalidInput
	}
	email = strings.TrimSpace(strings.ToLower(email))
	displayName = strings.TrimSpace(displayName)
	if email == "" {
		email = "bootstrap@flashx.local"
	}
	if displayName == "" {
		displayName = "FlashX Admin"
	}
	now := s.now()
	return s.store.Bootstrap(User{
		ID: ids.New("adm"),
		Phone: phone,
		Email: email,
		DisplayName: displayName,
		Role: "super_admin",
		Status: "active",
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) Get(id string) (User, error) {
	user, err := s.store.Get(id)
	if err != nil {
		return User{}, err
	}
	if user.Status != "active" {
		return User{}, ErrDisabled
	}
	return user, nil
}

func (s *Service) GetByPhone(phone string) (User, error) {
	normalized, err := auth.NormalizeVietnamPhone(phone)
	if err != nil {
		return User{}, ErrInvalidInput
	}
	user, err := s.store.GetByPhone(normalized)
	if err != nil {
		return User{}, err
	}
	if user.Status != "active" {
		return User{}, ErrDisabled
	}
	return user, nil
}

func (s *Service) Audit(actorID, action, resourceType, resourceID string, metadata map[string]any) error {
	if actorID == "" || action == "" || resourceType == "" {
		return ErrInvalidInput
	}
	return s.store.AppendAudit(AuditEntry{
		ActorType: "admin",
		ActorID: actorID,
		Action: action,
		ResourceType: resourceType,
		ResourceID: resourceID,
		Metadata: metadata,
		CreatedAt: s.now(),
	})
}
