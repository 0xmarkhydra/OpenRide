package trips

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidState = errors.New("invalid trip state")
	ErrForbidden    = errors.New("trip does not belong to actor")
	ErrInvalidInput = errors.New("invalid trip input")
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Create(input CreateInput) (Trip, error) {
	input.ServiceType = NormalizeServiceType(strings.TrimSpace(input.ServiceType))
	if input.RiderID == "" || !IsSupportedService(input.ServiceType) || !validPoint(input.Pickup) || !validPoint(input.Destination) {
		return Trip{}, ErrInvalidInput
	}
	if input.FareBreakdown.TotalMinor < 0 || input.EstimatedDistanceM < 0 || input.EstimatedDurationS < 0 {
		return Trip{}, ErrInvalidInput
	}
	if input.BookingMode == "" {
		input.BookingMode = BookingImmediate
	}
	if input.BookingMode != BookingImmediate && input.BookingMode != BookingScheduled {
		return Trip{}, ErrInvalidInput
	}
	if input.BookingMode == BookingScheduled && input.ScheduledAt == nil {
		return Trip{}, ErrInvalidInput
	}

	now := s.now()
	status := StatusSearching
	if input.BookingMode == BookingScheduled && input.ScheduledAt != nil && input.ScheduledAt.After(now.Add(30*time.Second)) {
		status = StatusScheduled
	}
	trip := Trip{
		ID:                 newID("trip"),
		RiderID:            input.RiderID,
		CustomerVehicleID:  input.CustomerVehicleID,
		ServiceType:        input.ServiceType,
		BookingMode:        input.BookingMode,
		ScheduledAt:        input.ScheduledAt,
		Status:             status,
		Pickup:             input.Pickup,
		Destination:        input.Destination,
		EstimatedDistanceM: input.EstimatedDistanceM,
		EstimatedDurationS: input.EstimatedDurationS,
		EstimatedFareMinor: input.FareBreakdown.TotalMinor,
		FareBreakdown:      input.FareBreakdown,
		Currency:           "VND",
		CreatedAt:          now,
		Version:            1,
	}
	if err := s.store.Create(trip); err != nil {
		return Trip{}, fmt.Errorf("create trip: %w", err)
	}
	return s.decorate(trip), nil
}

func (s *Service) Get(id string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	return s.decorate(trip), nil
}

func (s *Service) GetForDriver(id, driverID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.DriverID == "" || trip.DriverID != driverID {
		return Trip{}, ErrForbidden
	}
	return s.decorate(trip), nil
}

func (s *Service) GetForRider(id, riderID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.RiderID != riderID {
		return Trip{}, ErrForbidden
	}
	return s.decorate(trip), nil
}

func (s *Service) ListForRider(riderID string, limit int) ([]Trip, error) {
	if riderID == "" {
		return nil, ErrInvalidInput
	}
	items, err := s.store.ListByRider(riderID, limit)
	if err != nil {
		return nil, err
	}
	return s.decorateMany(items)
}

func (s *Service) ListForDriver(driverID string, limit int) ([]Trip, error) {
	if driverID == "" {
		return nil, ErrInvalidInput
	}
	items, err := s.store.ListByDriver(driverID, limit)
	if err != nil {
		return nil, err
	}
	return s.decorateMany(items)
}

func (s *Service) ListSearching(limit int) ([]Trip, error) {
	items, err := s.store.ListSearching(limit)
	if err != nil {
		return nil, err
	}
	return s.decorateMany(items)
}

func (s *Service) ListAll(limit int) ([]Trip, error) {
	items, err := s.store.ListAll(limit)
	if err != nil {
		return nil, err
	}
	return s.decorateMany(items)
}

func (s *Service) ActiveForDriver(driverID string) (Trip, error) {
	if driverID == "" {
		return Trip{}, ErrInvalidInput
	}
	trip, err := s.store.FindActiveByDriver(driverID)
	if err != nil {
		return Trip{}, err
	}
	return s.decorate(trip), nil
}

// ActivateDueScheduled is intentionally cheap and can be called from driver
// heartbeat/availability paths. A dedicated scheduler can replace this later.
func (s *Service) ActivateDueScheduled(limit int) ([]Trip, error) {
	items, err := s.store.ListScheduledDue(s.now(), limit)
	if err != nil {
		return nil, err
	}
	activated := make([]Trip, 0, len(items))
	for _, trip := range items {
		if !ValidTransitionForService(trip.ServiceType, trip.Status, StatusSearching) {
			continue
		}
		trip.Status = StatusSearching
		updated, saveErr := s.persist(trip)
		if saveErr == nil {
			activated = append(activated, updated)
		}
	}
	return activated, nil
}

func (s *Service) Cancel(id, riderID, reason string) (Trip, error) {
	trip, err := s.GetForRider(id, riderID)
	if err != nil {
		return Trip{}, err
	}
	if !trip.CanCancel() || !ValidTransitionForService(trip.ServiceType, trip.Status, StatusCancelled) {
		return Trip{}, ErrInvalidState
	}
	now := s.now()
	trip.Status = StatusCancelled
	trip.CancelledAt = &now
	trip.CancellationReason = strings.TrimSpace(reason)
	return s.persist(trip)
}

// CancelCustomerNoShow lets the assigned driver end a pickup only after the
// configured grace period has elapsed. This is enforced server-side so a
// modified client cannot bypass waiting policy.
func (s *Service) CancelCustomerNoShow(id, driverID string, grace time.Duration) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if trip.IncidentOpen || (trip.Status != StatusArrived && trip.Status != StatusArrivedForPickup) || trip.ArrivedAt == nil {
		return Trip{}, ErrInvalidState
	}
	if grace < time.Minute {
		grace = time.Minute
	}
	now := s.now()
	if now.Before(trip.ArrivedAt.Add(grace)) {
		return Trip{}, ErrInvalidState
	}
	if !ValidTransitionForService(trip.ServiceType, trip.Status, StatusCancelled) {
		return Trip{}, ErrInvalidState
	}
	trip.Status = StatusCancelled
	trip.CancelledAt = &now
	trip.CancellationReason = "customer_no_show"
	return s.persist(trip)
}

func (s *Service) AssignDriver(id, driverID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.Status != StatusSearching || trip.DriverID != "" || driverID == "" || !ValidTransitionForService(trip.ServiceType, trip.Status, StatusAccepted) {
		return Trip{}, ErrInvalidState
	}
	now := s.now()
	trip.DriverID = driverID
	trip.Status = StatusAccepted
	trip.AcceptedAt = &now
	return s.persist(trip)
}

// ReassignDriver is reserved for Operations. It only replaces a driver before
// vehicle custody and resets the lifecycle to accepted so the replacement
// driver must perform arrival and vehicle-receipt steps explicitly.
func (s *Service) ReassignDriver(id, driverID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if driverID == "" || trip.DriverID == "" || trip.DriverID == driverID || trip.IncidentOpen || !CanReassignBeforeCustody(trip.Status) {
		return Trip{}, ErrInvalidState
	}
	now := s.now()
	trip.DriverID = driverID
	trip.Status = StatusAccepted
	trip.AcceptedAt = &now
	trip.ArrivedAt = nil
	return s.persist(trip)
}

func (s *Service) MarkArriving(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	target := StatusArriving
	if IsInspectionService(trip.ServiceType) {
		target = StatusArrivingForPickup
	}
	return s.transition(trip, target)
}

func (s *Service) MarkArrived(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	target := StatusArrived
	if IsInspectionService(trip.ServiceType) {
		target = StatusArrivedForPickup
	}
	updated, err := s.transition(trip, target)
	if err != nil {
		return Trip{}, err
	}
	now := s.now()
	updated.ArrivedAt = &now
	return s.persistWithoutTransition(updated)
}

func (s *Service) MarkVehicleReceived(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	updated, err := s.transition(trip, StatusVehicleReceived)
	if err != nil {
		return Trip{}, err
	}
	now := s.now()
	updated.VehicleReceivedAt = &now
	return s.persistWithoutTransition(updated)
}

func (s *Service) Start(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	target := StatusInProgress
	if IsInspectionService(trip.ServiceType) {
		target = StatusEnRouteToInspection
	}
	updated, err := s.transition(trip, target)
	if err != nil {
		return Trip{}, err
	}
	now := s.now()
	updated.StartedAt = &now
	return s.persistWithoutTransition(updated)
}

func (s *Service) ArriveInspectionCenter(id, driverID string) (Trip, error) {
	return s.inspectionTransition(id, driverID, StatusArrivedAtInspectionCenter)
}
func (s *Service) StartInspection(id, driverID string) (Trip, error) {
	return s.inspectionTransition(id, driverID, StatusInspectionInProgress)
}
func (s *Service) CompleteInspection(id, driverID, result string) (Trip, error) {
	if result == "" {
		result = "passed"
	}
	if result != "passed" && result != "failed" && result != "deferred" && result != "unavailable" {
		return Trip{}, ErrInvalidInput
	}
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !IsInspectionService(trip.ServiceType) {
		return Trip{}, ErrInvalidState
	}
	updated, err := s.transition(trip, StatusInspectionCompleted)
	if err != nil {
		return Trip{}, err
	}
	updated.InspectionResult = result
	return s.persistWithoutTransition(updated)
}
func (s *Service) ReturningVehicle(id, driverID string) (Trip, error) {
	return s.inspectionTransition(id, driverID, StatusReturningVehicle)
}
func (s *Service) ArrivedForReturn(id, driverID string) (Trip, error) {
	return s.inspectionTransition(id, driverID, StatusArrivedForReturn)
}

func (s *Service) MarkHandover(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if trip.IncidentOpen {
		return Trip{}, ErrInvalidState
	}
	updated, err := s.transition(trip, StatusHandover)
	if err != nil {
		return Trip{}, err
	}
	now := s.now()
	updated.HandoverAt = &now
	return s.persistWithoutTransition(updated)
}

func (s *Service) Complete(id, driverID string, finalFareMinor int64) (Trip, error) {
	if finalFareMinor < 0 {
		return Trip{}, ErrInvalidInput
	}
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if trip.IncidentOpen || !ValidTransitionForService(trip.ServiceType, trip.Status, StatusCompleted) {
		return Trip{}, ErrInvalidState
	}
	if finalFareMinor == 0 {
		finalFareMinor = trip.EstimatedFareMinor
	}
	now := s.now()
	trip.Status = StatusCompleted
	trip.FinalFareMinor = finalFareMinor
	trip.CompletedAt = &now
	return s.persist(trip)
}

func (s *Service) ReportIncident(id, driverID, incidentType, note string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	return s.reportIncident(trip, incidentType, note)
}

func (s *Service) ReportIncidentForRider(id, riderID, incidentType, note string) (Trip, error) {
	trip, err := s.GetForRider(id, riderID)
	if err != nil {
		return Trip{}, err
	}
	return s.reportIncident(trip, incidentType, note)
}

func (s *Service) reportIncident(trip Trip, incidentType, note string) (Trip, error) {
	if !IsDriverOccupiedStatus(trip.Status) {
		return Trip{}, ErrInvalidState
	}
	incidentType = strings.TrimSpace(incidentType)
	note = strings.TrimSpace(note)
	if incidentType == "" || len(incidentType) > 80 || len(note) > 1000 {
		return Trip{}, ErrInvalidInput
	}
	trip.IncidentType = incidentType
	trip.IncidentNote = note
	trip.IncidentOpen = true
	return s.persistWithoutTransition(trip)
}

func (s *Service) ResolveIncident(id string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if !trip.IncidentOpen {
		return Trip{}, ErrInvalidState
	}
	trip.IncidentOpen = false
	return s.persistWithoutTransition(trip)
}

func (s *Service) inspectionTransition(id, driverID string, target Status) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !IsInspectionService(trip.ServiceType) {
		return Trip{}, ErrInvalidState
	}
	return s.transition(trip, target)
}

func (s *Service) transition(trip Trip, target Status) (Trip, error) {
	if trip.IncidentOpen || !ValidTransitionForService(trip.ServiceType, trip.Status, target) {
		return Trip{}, ErrInvalidState
	}
	trip.Status = target
	return s.persist(trip)
}

func (s *Service) driverTrip(id, driverID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.DriverID == "" || trip.DriverID != driverID {
		return Trip{}, ErrForbidden
	}
	return trip, nil
}

func (s *Service) persist(trip Trip) (Trip, error) {
	if err := s.store.Save(trip); err != nil {
		return Trip{}, err
	}
	return s.Get(trip.ID)
}

// persistWithoutTransition is used after transition() already persisted a new
// status and the same command still needs to stamp business metadata.
func (s *Service) persistWithoutTransition(trip Trip) (Trip, error) {
	if err := s.store.Save(trip); err != nil {
		return Trip{}, err
	}
	return s.Get(trip.ID)
}

func (s *Service) decorateMany(items []Trip) ([]Trip, error) {
	for i := range items {
		items[i] = s.decorate(items[i])
	}
	return items, nil
}

func (s *Service) decorate(trip Trip) Trip {
	trip.ServiceType = NormalizeServiceType(trip.ServiceType)
	if trip.BookingMode == "" {
		trip.BookingMode = BookingImmediate
	}
	trip.AllowedActions = allowedActions(trip)
	return trip
}

func allowedActions(trip Trip) []string {
	if trip.IncidentOpen {
		return []string{"contact_support"}
	}
	switch trip.Status {
	case StatusScheduled:
		return []string{"cancel"}
	case StatusSearching:
		return []string{"cancel"}
	case StatusAccepted:
		return []string{"arriving", "cancel"}
	case StatusArriving, StatusArrivingForPickup:
		return []string{"arrived", "cancel", "report_incident"}
	case StatusArrived, StatusArrivedForPickup:
		return []string{"vehicle_received", "cancel", "report_incident"}
	case StatusVehicleReceived:
		return []string{"start", "report_incident"}
	case StatusInProgress:
		return []string{"handover", "report_incident"}
	case StatusEnRouteToInspection:
		return []string{"arrive_inspection", "report_incident"}
	case StatusArrivedAtInspectionCenter:
		return []string{"start_inspection", "report_incident"}
	case StatusInspectionInProgress:
		return []string{"complete_inspection", "report_incident"}
	case StatusInspectionCompleted:
		return []string{"returning", "report_incident"}
	case StatusReturningVehicle:
		return []string{"arrived_return", "report_incident"}
	case StatusArrivedForReturn:
		return []string{"handover", "report_incident"}
	case StatusHandover:
		return []string{"complete"}
	default:
		return nil
	}
}

func validPoint(p Point) bool { return p.Lat >= -90 && p.Lat <= 90 && p.Lng >= -180 && p.Lng <= 180 }

func newID(prefix string) string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
