package ratings

import (
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

type Service struct {
	store Store
	trips *trips.Service
	now   func() time.Time
}

func NewService(store Store, tripService *trips.Service) *Service {
	return &Service{store: store, trips: tripService, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Create(riderID, tripID string, stars int16, comment string) (Rating, error) {
	if riderID == "" || tripID == "" || stars < 1 || stars > 5 {
		return Rating{}, ErrInvalidInput
	}
	comment = strings.TrimSpace(comment)
	if len(comment) > 1000 {
		return Rating{}, ErrInvalidInput
	}
	trip, err := s.trips.GetForRider(tripID, riderID)
	if err != nil {
		if err == trips.ErrForbidden {
			return Rating{}, ErrForbidden
		}
		return Rating{}, err
	}
	if trip.Status != trips.StatusCompleted || trip.DriverID == "" {
		return Rating{}, ErrInvalidInput
	}
	rating := Rating{
		ID: ids.New("rating"), TripID: trip.ID, RiderID: riderID, DriverID: trip.DriverID,
		Stars: stars, Comment: comment, CreatedAt: s.now(),
	}
	if err := s.store.Create(rating); err != nil {
		return Rating{}, err
	}
	return rating, nil
}

func (s *Service) GetByTripForRider(riderID, tripID string) (Rating, error) {
	trip, err := s.trips.GetForRider(tripID, riderID)
	if err != nil {
		return Rating{}, err
	}
	return s.store.GetByTrip(trip.ID)
}
