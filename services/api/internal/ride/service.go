package ride

import (
	"sync"
	"time"

	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/trips"
)

// Service orchestrates cross-domain job/driver changes. The process-local lock
// keeps the memory/demo runtime deterministic; persistent assignment still uses
// the dispatch lock + optimistic DB versioning already present in the project.
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

func (s *Service) CancelCustomerNoShow(id, driverID string, grace time.Duration) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	trip, err := s.trips.CancelCustomerNoShow(id, driverID, grace)
	if err != nil {
		return trips.Trip{}, err
	}
	if _, err := s.drivers.MarkAvailable(driverID); err != nil {
		return trip, err
	}
	return trip, nil
}
func (s *Service) MarkArrived(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.MarkArrived(id, driverID)
}
func (s *Service) MarkVehicleReceived(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.MarkVehicleReceived(id, driverID)
}
func (s *Service) Start(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.Start(id, driverID)
}
func (s *Service) ArriveInspectionCenter(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.ArriveInspectionCenter(id, driverID)
}
func (s *Service) StartInspection(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.StartInspection(id, driverID)
}
func (s *Service) CompleteInspection(id, driverID, result string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.CompleteInspection(id, driverID, result)
}
func (s *Service) ReturningVehicle(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.ReturningVehicle(id, driverID)
}
func (s *Service) ArrivedForReturn(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.ArrivedForReturn(id, driverID)
}
func (s *Service) MarkHandover(id, driverID string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.MarkHandover(id, driverID)
}
func (s *Service) ReportIncident(id, driverID, incidentType, note string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.ReportIncident(id, driverID, incidentType, note)
}

func (s *Service) ReportIncidentByRider(id, riderID, incidentType, note string) (trips.Trip, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trips.ReportIncidentForRider(id, riderID, incidentType, note)
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
