package payments

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusPaid      Status = "paid"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Payment struct {
	ID                string    `json:"id"`
	TripID            string    `json:"trip_id"`
	Provider          string    `json:"provider"`
	Method            string    `json:"method"`
	Status            Status    `json:"status"`
	AmountMinor       int64     `json:"amount_minor"`
	Currency          string    `json:"currency"`
	ProviderReference string    `json:"provider_reference,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
