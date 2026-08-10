package auth

import "time"

type Role string

type Actor struct {
	ID    string `json:"id"`
	Phone string `json:"phone,omitempty"`
	Role  Role   `json:"role"`
}

const (
	RoleRider  Role = "rider"
	RoleDriver Role = "driver"
	RoleAdmin  Role = "admin"
)

type Challenge struct {
	ID        string
	Phone     string
	Role      Role
	CodeHash  string
	Attempts  int
	ExpiresAt time.Time
	ConsumedAt *time.Time
	CreatedAt time.Time
}

type RefreshSession struct {
	ID        string
	ActorID   string
	Role      Role
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type Claims struct {
	ActorID   string `json:"sub"`
	Role      Role   `json:"role"`
	ExpiresAt int64  `json:"exp"`
	Nonce     string `json:"nonce"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type OTPRequestResult struct {
	ChallengeID string    `json:"challenge_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	DebugCode   string    `json:"debug_code,omitempty"`
}
