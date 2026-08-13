package customervehicles

import "time"

type Vehicle struct {
	ID             string    `json:"id"`
	OwnerUserID    string    `json:"owner_user_id"`
	Type           string    `json:"type"`
	LicensePlate   string    `json:"license_plate"`
	Brand          string    `json:"brand"`
	Model          string    `json:"model"`
	Year           int       `json:"year,omitempty"`
	Color          string    `json:"color"`
	Transmission   string    `json:"transmission"`
	Seats          int       `json:"seats,omitempty"`
	Notes          string    `json:"notes,omitempty"`
	PhotoObjectKey string    `json:"photo_object_key,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateInput struct {
	OwnerUserID    string
	Type           string
	LicensePlate   string
	Brand          string
	Model          string
	Year           int
	Color          string
	Transmission   string
	Seats          int
	Notes          string
	PhotoObjectKey string
}
