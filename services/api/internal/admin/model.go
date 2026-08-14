package admin

import "time"

const (
	RoleSuperAdmin = "super_admin"
	RoleOperations = "operations"
	StatusActive   = "active"
	StatusDisabled = "disabled"
)

type User struct {
	ID          string    `json:"id"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuditEntry struct {
	ActorType    string         `json:"actor_type"`
	ActorID      string         `json:"actor_id"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}
