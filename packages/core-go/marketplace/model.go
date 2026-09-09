package marketplace

import (
	"errors"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

type ServiceType string

type QuoteMode string

const (
	QuoteModeManual QuoteMode = "manual"
	QuoteModeAuto   QuoteMode = "auto"
	QuoteModeHybrid QuoteMode = "hybrid"
)

func (m QuoteMode) Valid() bool {
	return m == QuoteModeManual || m == QuoteModeAuto || m == QuoteModeHybrid
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

type QuoteStatus string

const (
	QuotePending     QuoteStatus = "pending"
	QuoteAccepted    QuoteStatus = "accepted"
	QuoteRejected    QuoteStatus = "rejected"
	QuoteWithdrawn   QuoteStatus = "withdrawn"
	QuoteExpired     QuoteStatus = "expired"
	QuoteInvalidated QuoteStatus = "invalidated"
)

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

var (
	ErrInvalidRequest = errors.New("marketplace: invalid request")
	ErrInvalidTariff  = errors.New("marketplace: invalid tariff")
	ErrInvalidQuote   = errors.New("marketplace: invalid quote")
)

// Request describes rider demand without assuming a specific service vertical.
// Vertical-specific fields belong in Attributes and are validated by a registered service module.
type Request struct {
	ID          string         `json:"id"`
	InstanceID  string         `json:"instance_id"`
	RiderID     string         `json:"rider_id"`
	ServiceType ServiceType    `json:"service_type"`
	Status      RequestStatus  `json:"status"`
	Pickup      geo.Point      `json:"pickup"`
	Destination *geo.Point     `json:"destination,omitempty"`
	Attributes  map[string]any `json:"attributes,omitempty"`
	Constraints map[string]any `json:"constraints,omitempty"`
	RequestedAt time.Time      `json:"requested_at"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	Version     int64          `json:"version"`
}

func (r Request) Validate() error {
	if r.ID == "" || r.InstanceID == "" || r.RiderID == "" || r.ServiceType == "" {
		return ErrInvalidRequest
	}
	if err := r.Pickup.Validate(); err != nil {
		return err
	}
	if r.Destination != nil {
		if err := r.Destination.Validate(); err != nil {
			return err
		}
	}
	if r.Version < 1 {
		return ErrInvalidRequest
	}
	if r.ExpiresAt != nil && !r.ExpiresAt.After(r.RequestedAt) {
		return ErrInvalidRequest
	}
	return nil
}

// DriverTariff is the driver's declared commercial policy for one service type.
// Core never silently overwrites these terms.
type DriverTariff struct {
	ID                string         `json:"id"`
	InstanceID        string         `json:"instance_id"`
	DriverID          string         `json:"driver_id"`
	ServiceType       ServiceType    `json:"service_type"`
	QuoteMode         QuoteMode      `json:"quote_mode"`
	BaseFare          money.Amount   `json:"base_fare"`
	MinimumFare       money.Amount   `json:"minimum_fare"`
	PerKM             money.Amount   `json:"per_km"`
	PerMinute         money.Amount   `json:"per_minute"`
	PickupFee         money.Amount   `json:"pickup_fee"`
	AutoQuoteMinimum  money.Amount   `json:"auto_quote_minimum"`
	AutoQuoteMaximum  money.Amount   `json:"auto_quote_maximum"`
	Rules             map[string]any `json:"rules,omitempty"`
	Version           int64          `json:"version"`
}

func (t DriverTariff) Validate() error {
	if t.ID == "" || t.DriverID == "" || t.InstanceID == "" || t.ServiceType == "" || !t.QuoteMode.Valid() || t.Version < 1 {
		return ErrInvalidTariff
	}
	amounts := []money.Amount{t.BaseFare, t.MinimumFare, t.PerKM, t.PerMinute, t.PickupFee, t.AutoQuoteMinimum, t.AutoQuoteMaximum}
	currency := ""
	for _, amount := range amounts {
		if err := amount.Validate(); err != nil {
			return err
		}
		if currency == "" {
			currency = amount.Currency
		} else if amount.Currency != currency {
			return money.ErrCurrencyMismatch
		}
	}
	if t.AutoQuoteMaximum.Minor > 0 && t.AutoQuoteMaximum.Minor < t.AutoQuoteMinimum.Minor {
		return ErrInvalidTariff
	}
	return nil
}

func (t DriverTariff) AllowsAutoQuote(total money.Amount) bool {
	if t.QuoteMode != QuoteModeAuto && t.QuoteMode != QuoteModeHybrid {
		return false
	}
	if err := t.Validate(); err != nil || total.Validate() != nil || total.Currency != t.BaseFare.Currency {
		return false
	}
	if t.AutoQuoteMinimum.Minor > 0 && total.Minor < t.AutoQuoteMinimum.Minor {
		return false
	}
	if t.AutoQuoteMaximum.Minor > 0 && total.Minor > t.AutoQuoteMaximum.Minor {
		return false
	}
	return true
}

type Quote struct {
	ID               string         `json:"id"`
	RequestID        string         `json:"request_id"`
	DriverID         string         `json:"driver_id"`
	DriverVehicleID  string         `json:"driver_vehicle_id,omitempty"`
	TariffID         string         `json:"tariff_id,omitempty"`
	TariffVersion    int64          `json:"tariff_version,omitempty"`
	Status           QuoteStatus    `json:"status"`
	Fare             money.Amount   `json:"fare"`
	PickupDistanceM  int64          `json:"pickup_distance_m"`
	PickupETAS       int64          `json:"pickup_eta_s"`
	Explanation      []string       `json:"explanation,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	ExpiresAt        time.Time      `json:"expires_at"`
	CreatedAt        time.Time      `json:"created_at"`
	AcceptedAt       *time.Time     `json:"accepted_at,omitempty"`
}

func (q Quote) Validate() error {
	if q.ID == "" || q.RequestID == "" || q.DriverID == "" || q.Status == "" || q.PickupDistanceM < 0 || q.PickupETAS < 0 {
		return ErrInvalidQuote
	}
	if err := q.Fare.Validate(); err != nil {
		return err
	}
	if !q.ExpiresAt.After(q.CreatedAt) {
		return ErrInvalidQuote
	}
	return nil
}

func (q Quote) IsSelectable(now time.Time) bool {
	return q.Validate() == nil && q.Status == QuotePending && q.ExpiresAt.After(now)
}

type Agreement struct {
	ID              string         `json:"id"`
	InstanceID      string         `json:"instance_id"`
	RequestID       string         `json:"request_id"`
	QuoteID         string         `json:"quote_id"`
	RiderID         string         `json:"rider_id"`
	DriverID        string         `json:"driver_id"`
	DriverVehicleID string         `json:"driver_vehicle_id,omitempty"`
	ServiceType     ServiceType    `json:"service_type"`
	Fare            money.Amount   `json:"fare"`
	TermsSnapshot   map[string]any `json:"terms_snapshot,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Ride struct {
	ID          string      `json:"id"`
	AgreementID string      `json:"agreement_id"`
	InstanceID  string      `json:"instance_id"`
	RiderID     string      `json:"rider_id"`
	DriverID    string      `json:"driver_id"`
	ServiceType ServiceType `json:"service_type"`
	Status      RideStatus  `json:"status"`
	Version     int64       `json:"version"`
}
