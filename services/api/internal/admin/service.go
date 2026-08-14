package admin

import (
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/platform/ids"
)

var ErrInvalidInput = errors.New("invalid admin input")

type CreateAccountInput struct {
	Phone       string
	Email       string
	DisplayName string
	Role        string
}

type UpdateAccountInput struct {
	DisplayName string
	Role        string
	Status      string
}

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
	user, err := s.store.Bootstrap(User{
		ID:          ids.New("adm"),
		Phone:       phone,
		Email:       email,
		DisplayName: displayName,
		Role:        RoleSuperAdmin,
		Status:      StatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		return User{}, err
	}
	// The configured bootstrap account is the recovery/root identity. A deploy
	// must restore it if someone accidentally downgraded or disabled it.
	changed := false
	if user.Role != RoleSuperAdmin {
		user.Role = RoleSuperAdmin
		changed = true
	}
	if user.Status != StatusActive {
		user.Status = StatusActive
		changed = true
	}
	if changed {
		user.UpdatedAt = now
		return s.store.Save(user)
	}
	return user, nil
}

func (s *Service) Get(id string) (User, error) {
	user, err := s.store.Get(id)
	if err != nil {
		return User{}, err
	}
	if user.Status != StatusActive {
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
	if user.Status != StatusActive {
		return User{}, ErrDisabled
	}
	return user, nil
}

func (s *Service) ListAccounts(limit int) ([]User, error) {
	return s.store.List(limit)
}

func (s *Service) GetAccount(id string) (User, error) {
	return s.store.Get(id)
}

func (s *Service) CreateAccount(input CreateAccountInput) (User, error) {
	phone, err := auth.NormalizeVietnamPhone(input.Phone)
	if err != nil || !validRole(input.Role) {
		return User{}, ErrInvalidInput
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = phone
	}
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		digits := strings.NewReplacer("+", "", " ", "").Replace(phone)
		email = digits + "@admin.flashx.local"
	}
	now := s.now()
	return s.store.Create(User{
		ID: ids.New("adm"), Phone: phone, Email: email, DisplayName: displayName,
		Role: input.Role, Status: StatusActive, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) UpdateAccount(id string, input UpdateAccountInput) (User, error) {
	if strings.TrimSpace(id) == "" || !validRole(input.Role) || !validStatus(input.Status) {
		return User{}, ErrInvalidInput
	}
	current, err := s.store.Get(id)
	if err != nil {
		return User{}, err
	}
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = current.DisplayName
	}
	removingSuperAdmin := current.Role == RoleSuperAdmin && current.Status == StatusActive && (input.Role != RoleSuperAdmin || input.Status != StatusActive)
	if removingSuperAdmin {
		items, err := s.store.List(500)
		if err != nil {
			return User{}, err
		}
		otherActiveSuper := false
		for _, item := range items {
			if item.ID != current.ID && item.Role == RoleSuperAdmin && item.Status == StatusActive {
				otherActiveSuper = true
				break
			}
		}
		if !otherActiveSuper {
			return User{}, ErrLastSuperAdmin
		}
	}
	current.DisplayName = displayName
	current.Role = input.Role
	current.Status = input.Status
	current.UpdatedAt = s.now()
	return s.store.Save(current)
}

func (s *Service) RequireRole(id string, roles ...string) (User, error) {
	user, err := s.Get(id)
	if err != nil {
		return User{}, err
	}
	for _, role := range roles {
		if user.Role == role {
			return user, nil
		}
	}
	return User{}, ErrForbidden
}

func validRole(role string) bool {
	return role == RoleSuperAdmin || role == RoleOperations
}

func validStatus(status string) bool {
	return status == StatusActive || status == StatusDisabled
}

func (s *Service) Audit(actorID, action, resourceType, resourceID string, metadata map[string]any) error {
	if actorID == "" || action == "" || resourceType == "" {
		return ErrInvalidInput
	}
	return s.store.AppendAudit(AuditEntry{
		ActorType:    "admin",
		ActorID:      actorID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Metadata:     metadata,
		CreatedAt:    s.now(),
	})
}

func (s *Service) ListAudit(limit int) ([]AuditEntry, error) {
	return s.store.ListAudit(limit)
}
