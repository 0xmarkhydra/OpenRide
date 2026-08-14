package custodyevidence

import (
	"time"

	"flashx/services/api/internal/objectstorage"
)

type Stage string

const (
	StagePickup Stage = "pickup"
	StageReturn Stage = "return"
)

type Evidence struct {
	ID                string     `json:"id"`
	TripID            string     `json:"trip_id"`
	Stage             Stage      `json:"stage"`
	ConditionNote     string     `json:"condition_note"`
	OdometerKm        *int64     `json:"odometer_km,omitempty"`
	FuelPercent       *int       `json:"fuel_percent,omitempty"`
	BatteryPercent    *int       `json:"battery_percent,omitempty"`
	DriverConfirmedAt *time.Time `json:"driver_confirmed_at,omitempty"`
	RiderConfirmedAt  *time.Time `json:"rider_confirmed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type Photo struct {
	ID          string    `json:"id"`
	EvidenceID  string    `json:"evidence_id"`
	PhotoType   string    `json:"photo_type"`
	ObjectKey   string    `json:"object_key"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	SizeBytes   int64     `json:"size_bytes"`
	CreatedAt   time.Time `json:"created_at"`
}

type PhotoWithURL struct {
	Photo Photo                        `json:"photo"`
	View  *objectstorage.SignedRequest `json:"view,omitempty"`
}

type Snapshot struct {
	Evidence Evidence       `json:"evidence"`
	Photos   []PhotoWithURL `json:"photos"`
	Ready    bool           `json:"ready"`
}

type EvidenceInput struct {
	ConditionNote  string `json:"condition_note"`
	OdometerKm     *int64 `json:"odometer_km,omitempty"`
	FuelPercent    *int   `json:"fuel_percent,omitempty"`
	BatteryPercent *int   `json:"battery_percent,omitempty"`
}

type PhotoUploadInput struct {
	PhotoType   string `json:"photo_type"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type CompletePhotoInput struct {
	PhotoID     string `json:"photo_id"`
	PhotoType   string `json:"photo_type"`
	ObjectKey   string `json:"object_key"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

type UploadTicket struct {
	EvidenceID string                      `json:"evidence_id"`
	PhotoID    string                      `json:"photo_id"`
	PhotoType  string                      `json:"photo_type"`
	ObjectKey  string                      `json:"object_key"`
	Upload     objectstorage.SignedRequest `json:"upload"`
}
