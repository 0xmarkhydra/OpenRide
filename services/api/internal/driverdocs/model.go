package driverdocs

import "time"

type ReviewStatus string

const (
	ReviewPending  ReviewStatus = "pending"
	ReviewApproved ReviewStatus = "approved"
	ReviewRejected ReviewStatus = "rejected"
)

type Document struct {
	ID           string       `json:"id"`
	DriverID     string       `json:"driver_id"`
	DocumentType string       `json:"document_type"`
	ObjectKey    string       `json:"object_key"`
	Filename     string       `json:"filename"`
	ContentType  string       `json:"content_type"`
	SizeBytes    int64        `json:"size_bytes"`
	ReviewStatus ReviewStatus `json:"review_status"`
	ReviewNote   string       `json:"review_note"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	ReviewedAt   *time.Time   `json:"reviewed_at,omitempty"`
}

type UploadInput struct {
	DocumentType string `json:"document_type"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

type CompleteInput struct {
	DocumentID   string `json:"document_id"`
	DocumentType string `json:"document_type"`
	ObjectKey    string `json:"object_key"`
	Filename     string `json:"filename"`
	ContentType  string `json:"content_type"`
	SizeBytes    int64  `json:"size_bytes"`
}
