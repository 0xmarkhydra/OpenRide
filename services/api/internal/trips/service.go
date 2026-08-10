package trips

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	if input.RiderID == "" || input.ServiceType == "" || !validPoint(input.Pickup) || !validPoint(input.Destination) {
		return Trip{}, ErrInvalidInput
	}
	if input.FareBreakdown.TotalMinor < 0 || input.EstimatedDistanceM < 0 || input.EstimatedDurationS < 0 {
		return Trip{}, ErrInvalidInput
	}

	trip := Trip{
		ID:                 newID("trip"),
		RiderID:            input.RiderID,
		ServiceType:        input.ServiceType,
		Status:             StatusSearching,
		Pickup:             input.Pickup,
		Destination:        input.Destination,
		EstimatedDistanceM: input.EstimatedDistanceM,
		EstimatedDurationS: input.EstimatedDurationS,
		EstimatedFareMinor: input.FareBreakdown.TotalMinor,
		FareBreakdown:      input.FareBreakdown,
		Currency:           "VND",
		CreatedAt:          s.now(),
		Version:            1,
	}
	if err := s.store.Create(trip); err != nil {
		return Trip{}, fmt.Errorf("create trip: %w", err)
	}
	return trip, nil
}

func (s *Service) Get(id string) (Trip, error) {
	return s.store.Get(id)
}

func (s *Service) GetForRider(id, riderID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.RiderID != riderID {
		return Trip{}, ErrForbidden
	}
	return trip, nil
}

func (s *Service) ListForRider(riderID string, limit int) ([]Trip, error) {
	if riderID == "" {
		return nil, ErrInvalidInput
	}
	return s.store.ListByRider(riderID, limit)
}

func (s *Service) ListForDriver(driverID string, limit int) ([]Trip, error) {
	if driverID == "" {
		return nil, ErrInvalidInput
	}
	return s.store.ListByDriver(driverID, limit)
}

func (s *Service) ListSearching(limit int) ([]Trip, error) {
	return s.store.ListSearching(limit)
}

func (s *Service) ListAll(limit int) ([]Trip, error) {
	return s.store.ListAll(limit)
}

func (s *Service) ActiveForDriver(driverID string) (Trip, error) {
	if driverID == "" { return Trip{}, ErrInvalidInput }
	return s.store.FindActiveByDriver(driverID)
}

func (s *Service) Cancel(id, riderID, reason string) (Trip, error) {
	trip, err := s.GetForRider(id, riderID)
	if err != nil {
		return Trip{}, err
	}
	if !trip.CanCancel() || !ValidTransition(trip.Status, StatusCancelled) {
		return Trip{}, ErrInvalidState
	}

	now := s.now()
	trip.Status = StatusCancelled
	trip.CancelledAt = &now
	trip.CancellationReason = reason
	return s.persist(trip)
}

func (s *Service) AssignDriver(id, driverID string) (Trip, error) {
	trip, err := s.store.Get(id)
	if err != nil {
		return Trip{}, err
	}
	if trip.Status != StatusSearching || trip.DriverID != "" || driverID == "" || !ValidTransition(trip.Status, StatusAccepted) {
		return Trip{}, ErrInvalidState
	}

	now := s.now()
	trip.DriverID = driverID
	trip.Status = StatusAccepted
	trip.AcceptedAt = &now
	return s.persist(trip)
}

func (s *Service) MarkArriving(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !ValidTransition(trip.Status, StatusArriving) {
		return Trip{}, ErrInvalidState
	}
	trip.Status = StatusArriving
	return s.persist(trip)
}

func (s *Service) MarkArrived(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !ValidTransition(trip.Status, StatusArrived) {
		return Trip{}, ErrInvalidState
	}

	now := s.now()
	trip.Status = StatusArrived
	trip.ArrivedAt = &now
	return s.persist(trip)
}

func (s *Service) Start(id, driverID string) (Trip, error) {
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !ValidTransition(trip.Status, StatusInProgress) {
		return Trip{}, ErrInvalidState
	}

	now := s.now()
	trip.Status = StatusInProgress
	trip.StartedAt = &now
	return s.persist(trip)
}

func (s *Service) Complete(id, driverID string, finalFareMinor int64) (Trip, error) {
	if finalFareMinor < 0 {
		return Trip{}, ErrInvalidInput
	}
	trip, err := s.driverTrip(id, driverID)
	if err != nil {
		return Trip{}, err
	}
	if !ValidTransition(trip.Status, StatusCompleted) {
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
	return s.store.Get(trip.ID)
}

func validPoint(p Point) bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lng >= -180 && p.Lng <= 180
}

func newID(prefix string) string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
