package marketplace

import "time"

type QuoteMode string

const (
	QuoteModeManual QuoteMode = "manual"
	QuoteModeAuto   QuoteMode = "auto"
	QuoteModeHybrid QuoteMode = "hybrid"
)

func (m QuoteMode) Valid() bool {
	switch m {
	case QuoteModeManual, QuoteModeAuto, QuoteModeHybrid:
		return true
	default:
		return false
	}
}

type RequestStatus string

const (
	RequestDraft           RequestStatus = "draft"
	RequestOpen            RequestStatus = "open"
	RequestReceivingQuotes RequestStatus = "receiving_quotes"
	RequestAgreed          RequestStatus = "agreed"
	RequestClosed          RequestStatus = "closed"
	RequestCancelled       RequestStatus = "cancelled"
	RequestExpired         RequestStatus = "expired"
)

func ValidRequestTransition(from, to RequestStatus) bool {
	switch from {
	case RequestDraft:
		return to == RequestOpen || to == RequestCancelled
	case RequestOpen:
		return to == RequestReceivingQuotes || to == RequestAgreed || to == RequestCancelled || to == RequestExpired
	case RequestReceivingQuotes:
		return to == RequestAgreed || to == RequestCancelled || to == RequestExpired
	case RequestAgreed:
		return to == RequestClosed
	default:
		return false
	}
}

type QuoteStatus string

const (
	QuotePending     QuoteStatus = "pending"
	QuoteAccepted    QuoteStatus = "accepted"
	QuoteRejected    QuoteStatus = "rejected"
	QuoteWithdrawn   QuoteStatus = "withdrawn"
	QuoteExpired     QuoteStatus = "expired"
	QuoteInvalidated QuoteStatus = "invalidated"
)

func ValidQuoteTransition(from, to QuoteStatus) bool {
	if from != QuotePending {
		return false
	}
	switch to {
	case QuoteAccepted, QuoteRejected, QuoteWithdrawn, QuoteExpired, QuoteInvalidated:
		return true
	default:
		return false
	}
}

type RideStatus string

const (
	RideAssigned         RideStatus = "assigned"
	RideDriverEnRoute    RideStatus = "driver_en_route"
	RideDriverArrived    RideStatus = "driver_arrived"
	RidePassengerOnboard RideStatus = "passenger_onboard"
	RideInProgress       RideStatus = "in_progress"
	RideCompleted        RideStatus = "completed"
	RideCancelled        RideStatus = "cancelled"
	RideFailed           RideStatus = "failed"
)

func ValidPassengerRideTransition(from, to RideStatus) bool {
	if to == RideCancelled || to == RideFailed {
		switch from {
		case RideAssigned, RideDriverEnRoute, RideDriverArrived, RidePassengerOnboard, RideInProgress:
			return true
		default:
			return false
		}
	}

	switch from {
	case RideAssigned:
		return to == RideDriverEnRoute
	case RideDriverEnRoute:
		return to == RideDriverArrived
	case RideDriverArrived:
		return to == RidePassengerOnboard
	case RidePassengerOnboard:
		return to == RideInProgress
	case RideInProgress:
		return to == RideCompleted
	default:
		return false
	}
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type DriverTariff struct {
	ID                string    `json:"id"`
	InstanceID        string    `json:"instance_id"`
	DriverID          string    `json:"driver_id"`
	ServiceType       string    `json:"service_type"`
	QuoteMode         QuoteMode `json:"quote_mode"`
	Currency          string    `json:"currency"`
	BaseFareMinor     int64     `json:"base_fare_minor"`
	MinimumFareMinor  int64     `json:"minimum_fare_minor"`
	PerKMMinor        int64     `json:"per_km_minor"`
	PerMinuteMinor    int64     `json:"per_minute_minor"`
	PickupFeeMinor    int64     `json:"pickup_fee_minor"`
	AutoQuoteMinMinor int64     `json:"auto_quote_min_minor"`
	AutoQuoteMaxMinor int64     `json:"auto_quote_max_minor"`
	Version           int64     `json:"version"`
}

func (t DriverTariff) AllowsAutoQuote(totalMinor int64) bool {
	if t.QuoteMode != QuoteModeAuto && t.QuoteMode != QuoteModeHybrid {
		return false
	}
	if totalMinor < 0 {
		return false
	}
	if t.AutoQuoteMinMinor > 0 && totalMinor < t.AutoQuoteMinMinor {
		return false
	}
	if t.AutoQuoteMaxMinor > 0 && totalMinor > t.AutoQuoteMaxMinor {
		return false
	}
	return true
}

type MobilityRequest struct {
	ID                 string        `json:"id"`
	InstanceID         string        `json:"instance_id"`
	RiderID            string        `json:"rider_id"`
	ServiceType        string        `json:"service_type"`
	Status             RequestStatus `json:"status"`
	Pickup             Point         `json:"pickup"`
	Destination        Point         `json:"destination"`
	EstimatedDistanceM int64         `json:"estimated_distance_m"`
	EstimatedDurationS int64         `json:"estimated_duration_s"`
	RequestedAt        time.Time     `json:"requested_at"`
	ExpiresAt          *time.Time    `json:"expires_at,omitempty"`
	Version            int64         `json:"version"`
}

type Quote struct {
	ID              string      `json:"id"`
	RequestID       string      `json:"request_id"`
	DriverID        string      `json:"driver_id"`
	DriverVehicleID string      `json:"driver_vehicle_id,omitempty"`
	TariffID        string      `json:"tariff_id,omitempty"`
	TariffVersion   int64       `json:"tariff_version,omitempty"`
	Status          QuoteStatus `json:"status"`
	FareTotalMinor  int64       `json:"fare_total_minor"`
	Currency        string      `json:"currency"`
	PickupDistanceM int64       `json:"pickup_distance_m"`
	PickupETAS      int64       `json:"pickup_eta_s"`
	ExpiresAt       time.Time   `json:"expires_at"`
	CreatedAt       time.Time   `json:"created_at"`
	AcceptedAt      *time.Time  `json:"accepted_at,omitempty"`
}

func (q Quote) IsSelectable(now time.Time) bool {
	return q.Status == QuotePending && q.ExpiresAt.After(now)
}

type Agreement struct {
	ID              string    `json:"id"`
	InstanceID      string    `json:"instance_id"`
	RequestID       string    `json:"request_id"`
	QuoteID         string    `json:"quote_id"`
	RiderID         string    `json:"rider_id"`
	DriverID        string    `json:"driver_id"`
	DriverVehicleID string    `json:"driver_vehicle_id,omitempty"`
	ServiceType     string    `json:"service_type"`
	FareTotalMinor  int64     `json:"fare_total_minor"`
	Currency        string    `json:"currency"`
	CreatedAt       time.Time `json:"created_at"`
}

type Ride struct {
	ID          string     `json:"id"`
	AgreementID string     `json:"agreement_id"`
	InstanceID  string     `json:"instance_id"`
	RiderID     string     `json:"rider_id"`
	DriverID    string     `json:"driver_id"`
	ServiceType string     `json:"service_type"`
	Status      RideStatus `json:"status"`
	Version     int64      `json:"version"`
}
