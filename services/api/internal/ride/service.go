package ride

import (
	"sync"

	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/trips"
)

// Service orchestrates cross-domain trip/driver state changes. The in-memory
// MVP implementation serializes these commands. Production persistence must
// replace this process-local lock with a database transaction/outbox pattern.
type Service struct {
	mu      sync.Mutex
	trips   *trips.Service
	drivers *drivers.Service
}

func NewService(tripService *trips.Service, driverService *drivers.Service) *Service {
	return &Service{trips: tripService, drivers: driverService}
}

func (s *Service) CancelByRider(id, riderID, reason string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	before, err := s.trips.GetForRider(id, riderID)
	if err != nil {
		return trips.Trip{}, err
	}
	trip, err := s.trips.Cancel(id, riderID, reason)
	if err != nil {
		return trips.Trip{}, err
	}
	if before.DriverID != "" {
		if _, err := s.drivers.MarkAvailable(before.DriverID); err != nil {
			return trip, err
		}
	}
	return trip, nil
}

func (s *Service) MarkArriving(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.MarkArriving(id, driverID)
}

func (s *Service) MarkArrived(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.MarkArrived(id, driverID)
}

func (s *Service) Start(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.Start(id, driverID)
}

func (s *Service) Complete(id, driverID string, finalFareMinor int64) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trip, err := s.trips.Complete(id, driverID, finalFareMinor)
	if err != nil {
		return trips.Trip{}, err
	}
	if _, err := s.drivers.MarkAvailable(driverID); err != nil {
		return trip, err
	}
	return trip, nil
}
