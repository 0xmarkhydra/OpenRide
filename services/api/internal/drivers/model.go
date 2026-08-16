package drivers

import "time"

type ApprovalStatus string
type AvailabilityStatus string

const (
	ApprovalPending   ApprovalStatus = "pending"
	ApprovalApproved  ApprovalStatus = "approved"
	ApprovalRejected  ApprovalStatus = "rejected"
	ApprovalSuspended ApprovalStatus = "suspended"

	AvailabilityOffline AvailabilityStatus = "offline"
	AvailabilityOnline  AvailabilityStatus = "online"
	AvailabilityBusy    AvailabilityStatus = "busy"
)

type Location struct {
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	AccuracyM  float64   `json:"accuracy_m,omitempty"`
	HeadingDeg float64   `json:"heading_deg,omitempty"`
	SpeedMPS   float64   `json:"speed_mps,omitempty"`
	CapturedAt time.Time `json:"captured_at"`
}

type Driver struct {
	ID             string             `json:"id"`
	Phone          string             `json:"phone"`
	FullName       string             `json:"full_name"`
	ServiceType    string             `json:"service_type"`
	Capabilities   []string           `json:"capabilities"`
	LicenseClass   string             `json:"license_class,omitempty"`
	LicenseExpiry  *time.Time         `json:"license_expiry,omitempty"`
	CanDriveManual bool               `json:"can_drive_manual"`
	Approval       ApprovalStatus     `json:"approval_status"`
	Availability   AvailabilityStatus `json:"availability_status"`
	Location       *Location          `json:"location,omitempty"`
	LastIdleAt     time.Time          `json:"last_idle_at"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	Version        int64              `json:"version"`
}
