package trips

import "time"

type Status string

const (
	StatusSearching  Status = "searching"
	StatusAccepted   Status = "accepted"
	StatusArriving   Status = "arriving"
	StatusArrived    Status = "arrived"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type FareBreakdown struct {
	BaseFareMinor int64 `json:"base_fare_minor"`
	DistanceMinor int64 `json:"distance_minor"`
	DiscountMinor int64 `json:"discount_minor"`
	TotalMinor    int64 `json:"total_minor"`
}

type Trip struct {
	ID                 string        `json:"id"`
	RiderID            string        `json:"rider_id"`
	DriverID           string        `json:"driver_id,omitempty"`
	ServiceType        string        `json:"service_type"`
	Status             Status        `json:"status"`
	Pickup             Point         `json:"pickup"`
	Destination        Point         `json:"destination"`
	EstimatedDistanceM int64         `json:"estimated_distance_m"`
	EstimatedDurationS int64         `json:"estimated_duration_s"`
	EstimatedFareMinor int64         `json:"estimated_fare_minor"`
	FinalFareMinor     int64         `json:"final_fare_minor,omitempty"`
	FareBreakdown      FareBreakdown `json:"fare_breakdown"`
	Currency           string        `json:"currency"`
	CreatedAt          time.Time     `json:"created_at"`
	AcceptedAt         *time.Time    `json:"accepted_at,omitempty"`
	ArrivedAt          *time.Time    `json:"arrived_at,omitempty"`
	StartedAt          *time.Time    `json:"started_at,omitempty"`
	CompletedAt        *time.Time    `json:"completed_at,omitempty"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CancellationReason string        `json:"cancellation_reason,omitempty"`
	Version            int64         `json:"version"`
}

type CreateInput struct {
	RiderID            string
	ServiceType        string
	Pickup             Point
	Destination        Point
	EstimatedDistanceM int64
	EstimatedDurationS int64
	FareBreakdown      FareBreakdown
}

func (t Trip) CanCancel() bool {
	switch t.Status {
	case StatusSearching, StatusAccepted, StatusArriving, StatusArrived:
		return true
	default:
		return false
	}
}

func ValidTransition(from, to Status) bool {
	switch from {
	case StatusSearching:
		return to == StatusAccepted || to == StatusCancelled
	case StatusAccepted:
		return to == StatusArriving || to == StatusArrived || to == StatusCancelled
	case StatusArriving:
		return to == StatusArrived || to == StatusCancelled
	case StatusArrived:
		return to == StatusInProgress || to == StatusCancelled
	case StatusInProgress:
		return to == StatusCompleted
	default:
		return false
	}
}
