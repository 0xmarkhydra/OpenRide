package trips

import "time"

type Status string

const (
	ServiceDesignatedDriverCar  = "designated_driver_car"
	ServiceDesignatedDriverBike = "designated_driver_bike"
	ServiceVehicleInspection    = "vehicle_inspection_assist"

	BookingImmediate = "immediate"
	BookingScheduled = "scheduled"

	StatusScheduled                 Status = "scheduled"
	StatusSearching                 Status = "searching"
	StatusAccepted                  Status = "accepted"
	StatusArriving                  Status = "arriving"
	StatusArrived                   Status = "arrived"
	StatusArrivingForPickup         Status = "arriving_for_pickup"
	StatusArrivedForPickup          Status = "arrived_for_pickup"
	StatusVehicleReceived           Status = "vehicle_received"
	StatusInProgress                Status = "in_progress"
	StatusEnRouteToInspection       Status = "en_route_to_inspection"
	StatusArrivedAtInspectionCenter Status = "arrived_at_inspection_center"
	StatusInspectionInProgress      Status = "inspection_in_progress"
	StatusInspectionCompleted       Status = "inspection_completed"
	StatusReturningVehicle          Status = "returning_vehicle"
	StatusArrivedForReturn          Status = "arrived_for_return"
	StatusHandover                  Status = "handover"
	StatusCompleted                 Status = "completed"
	StatusCancelled                 Status = "cancelled"
)

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type FareBreakdown struct {
	BaseFareMinor  int64 `json:"base_fare_minor"`
	DistanceMinor  int64 `json:"distance_minor"`
	ServiceMinor   int64 `json:"service_minor,omitempty"`
	ScheduleMinor  int64 `json:"schedule_minor,omitempty"`
	SurchargeMinor int64 `json:"surcharge_minor,omitempty"`
	DiscountMinor  int64 `json:"discount_minor"`
	TotalMinor     int64 `json:"total_minor"`
}

type Trip struct {
	ID                 string        `json:"id"`
	RiderID            string        `json:"rider_id"`
	DriverID           string        `json:"driver_id,omitempty"`
	CustomerVehicleID  string        `json:"customer_vehicle_id,omitempty"`
	ServiceType        string        `json:"service_type"`
	BookingMode        string        `json:"booking_mode"`
	ScheduledAt        *time.Time    `json:"scheduled_at,omitempty"`
	Status             Status        `json:"status"`
	Pickup             Point         `json:"pickup"`
	Destination        Point         `json:"destination"`
	EstimatedDistanceM int64         `json:"estimated_distance_m"`
	EstimatedDurationS int64         `json:"estimated_duration_s"`
	EstimatedFareMinor int64         `json:"estimated_fare_minor"`
	FinalFareMinor     int64         `json:"final_fare_minor,omitempty"`
	FareBreakdown      FareBreakdown `json:"fare_breakdown"`
	Currency           string        `json:"currency"`
	InspectionResult   string        `json:"inspection_result,omitempty"`
	IncidentType       string        `json:"incident_type,omitempty"`
	IncidentNote       string        `json:"incident_note,omitempty"`
	IncidentOpen       bool          `json:"incident_open"`
	AllowedActions     []string      `json:"allowed_actions,omitempty"`
	CreatedAt          time.Time     `json:"created_at"`
	AcceptedAt         *time.Time    `json:"accepted_at,omitempty"`
	ArrivedAt          *time.Time    `json:"arrived_at,omitempty"`
	VehicleReceivedAt  *time.Time    `json:"vehicle_received_at,omitempty"`
	StartedAt          *time.Time    `json:"started_at,omitempty"`
	HandoverAt         *time.Time    `json:"handover_at,omitempty"`
	CompletedAt        *time.Time    `json:"completed_at,omitempty"`
	CancelledAt        *time.Time    `json:"cancelled_at,omitempty"`
	CancellationReason string        `json:"cancellation_reason,omitempty"`
	Version            int64         `json:"version"`
}

type CreateInput struct {
	RiderID            string
	CustomerVehicleID  string
	ServiceType        string
	BookingMode        string
	ScheduledAt        *time.Time
	Pickup             Point
	Destination        Point
	EstimatedDistanceM int64
	EstimatedDurationS int64
	FareBreakdown      FareBreakdown
}

func NormalizeServiceType(value string) string {
	switch value {
	case "car":
		return ServiceDesignatedDriverCar
	case "bike":
		return ServiceDesignatedDriverBike
	default:
		return value
	}
}

func IsSupportedService(value string) bool {
	switch NormalizeServiceType(value) {
	case ServiceDesignatedDriverCar, ServiceDesignatedDriverBike, ServiceVehicleInspection:
		return true
	default:
		return false
	}
}

func IsInspectionService(value string) bool {
	return NormalizeServiceType(value) == ServiceVehicleInspection
}

func IsTerminalStatus(status Status) bool {
	return status == StatusCompleted || status == StatusCancelled
}

func IsDriverOccupiedStatus(status Status) bool {
	return status != StatusScheduled && status != StatusSearching && !IsTerminalStatus(status)
}

// CanReassignBeforeCustody defines the only states where Operations may swap
// an assigned driver. Once vehicle_received is reached, custody has changed
// hands and normal reassignment is intentionally forbidden.
func CanReassignBeforeCustody(status Status) bool {
	switch status {
	case StatusAccepted, StatusArriving, StatusArrived, StatusArrivingForPickup, StatusArrivedForPickup:
		return true
	default:
		return false
	}
}

func (t Trip) CanCancel() bool {
	if t.IncidentOpen {
		return false
	}
	switch t.Status {
	case StatusScheduled, StatusSearching, StatusAccepted, StatusArriving, StatusArrived, StatusArrivingForPickup, StatusArrivedForPickup:
		return true
	default:
		return false
	}
}

func ValidTransitionForService(service string, from, to Status) bool {
	inspection := IsInspectionService(service)
	if from == StatusScheduled {
		return to == StatusSearching || to == StatusCancelled
	}
	if from == StatusSearching {
		return to == StatusAccepted || to == StatusCancelled
	}
	if inspection {
		switch from {
		case StatusAccepted:
			return to == StatusArrivingForPickup || to == StatusCancelled
		case StatusArrivingForPickup:
			return to == StatusArrivedForPickup || to == StatusCancelled
		case StatusArrivedForPickup:
			return to == StatusVehicleReceived || to == StatusCancelled
		case StatusVehicleReceived:
			return to == StatusEnRouteToInspection
		case StatusEnRouteToInspection:
			return to == StatusArrivedAtInspectionCenter
		case StatusArrivedAtInspectionCenter:
			return to == StatusInspectionInProgress
		case StatusInspectionInProgress:
			return to == StatusInspectionCompleted
		case StatusInspectionCompleted:
			return to == StatusReturningVehicle
		case StatusReturningVehicle:
			return to == StatusArrivedForReturn
		case StatusArrivedForReturn:
			return to == StatusHandover
		case StatusHandover:
			return to == StatusCompleted
		default:
			return false
		}
	}
	switch from {
	case StatusAccepted:
		return to == StatusArriving || to == StatusArrived || to == StatusCancelled
	case StatusArriving:
		return to == StatusArrived || to == StatusCancelled
	case StatusArrived:
		return to == StatusVehicleReceived || to == StatusCancelled
	case StatusVehicleReceived:
		return to == StatusInProgress
	case StatusInProgress:
		return to == StatusHandover
	case StatusHandover:
		return to == StatusCompleted
	default:
		return false
	}
}

// ValidTransition is retained for legacy tests/callers and follows designated-driver semantics.
func ValidTransition(from, to Status) bool {
	return ValidTransitionForService(ServiceDesignatedDriverBike, from, to)
}
