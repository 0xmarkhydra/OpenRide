package inspectionchecklist

import "time"

const (
	CustomerPending       = "pending"
	CustomerPresent       = "present"
	CustomerNotAvailable  = "not_available"
	CustomerNotApplicable = "not_applicable"

	DriverPending       = "pending"
	DriverReceived      = "received"
	DriverMissing       = "missing"
	DriverNotApplicable = "not_applicable"
)

type Template struct {
	ID        string         `json:"id"`
	Version   int64          `json:"version"`
	Active    bool           `json:"active"`
	Items     []TemplateItem `json:"items"`
	CreatedBy string         `json:"created_by,omitempty"`
	Reason    string         `json:"reason,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type TemplateItem struct {
	Key       string `json:"key"`
	Label     string `json:"label"`
	Required  bool   `json:"required"`
	SortOrder int    `json:"sort_order"`
}

type Checklist struct {
	TripID          string    `json:"trip_id"`
	TemplateID      string    `json:"template_id"`
	TemplateVersion int64     `json:"template_version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Item struct {
	ID                string     `json:"id"`
	TripID            string     `json:"trip_id"`
	Key               string     `json:"key"`
	Label             string     `json:"label"`
	Required          bool       `json:"required"`
	SortOrder         int        `json:"sort_order"`
	CustomerStatus    string     `json:"customer_status"`
	CustomerNote      string     `json:"customer_note,omitempty"`
	CustomerUpdatedAt *time.Time `json:"customer_updated_at,omitempty"`
	DriverStatus      string     `json:"driver_status"`
	DriverNote        string     `json:"driver_note,omitempty"`
	DriverUpdatedAt   *time.Time `json:"driver_updated_at,omitempty"`
}

type Snapshot struct {
	Checklist        Checklist `json:"checklist"`
	Items            []Item    `json:"items"`
	CustomerComplete bool      `json:"customer_complete"`
	DriverComplete   bool      `json:"driver_complete"`
	Ready            bool      `json:"ready"`
}

type CustomerUpdate struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

type DriverUpdate struct {
	Status string `json:"status"`
	Note   string `json:"note,omitempty"`
}

type ReplaceTemplateInput struct {
	Items  []TemplateItem `json:"items"`
	Reason string         `json:"reason"`
}
