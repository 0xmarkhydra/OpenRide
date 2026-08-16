package notifications

import "time"

type Role string

const (
	RoleRider  Role = "rider"
	RoleDriver Role = "driver"
)

type Platform string

const (
	PlatformAndroid Platform = "android"
	PlatformIOS     Platform = "ios"
)

type Device struct {
	ID        string    `json:"id"`
	ActorID   string    `json:"actor_id"`
	ActorRole Role      `json:"actor_role"`
	Platform  Platform  `json:"platform"`
	Token     string    `json:"-"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusSkipped Status = "skipped"
	StatusFailed  Status = "failed"
)

type Message struct {
	ID         string         `json:"id"`
	ActorID    string         `json:"actor_id"`
	ActorRole  Role           `json:"actor_role"`
	EventType  string         `json:"event_type"`
	ResourceID string         `json:"resource_id,omitempty"`
	Title      string         `json:"title"`
	Body       string         `json:"body"`
	Data       map[string]any `json:"data"`
	Status     Status         `json:"status"`
	Attempts   int            `json:"attempts"`
	LastError  string         `json:"last_error,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type RegisterDeviceInput struct {
	Platform Platform `json:"platform"`
	Token    string   `json:"token"`
}
