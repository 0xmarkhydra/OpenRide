package customervehicles

import (
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

var (
	ErrInvalidInput = errors.New("invalid customer vehicle input")
	ErrForbidden    = errors.New("customer vehicle forbidden")
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Create(input CreateInput) (Vehicle, error) {
	input.OwnerUserID = strings.TrimSpace(input.OwnerUserID)
	input.Type = strings.TrimSpace(input.Type)
	input.LicensePlate = strings.ToUpper(strings.TrimSpace(input.LicensePlate))
	input.Transmission = strings.TrimSpace(input.Transmission)
	if input.OwnerUserID == "" || input.LicensePlate == "" || (input.Type != "car" && input.Type != "motorbike") {
		return Vehicle{}, ErrInvalidInput
	}
	if input.Transmission == "" {
		input.Transmission = "n/a"
	}
	if input.Transmission != "automatic" && input.Transmission != "manual" && input.Transmission != "n/a" {
		return Vehicle{}, ErrInvalidInput
	}
	if input.Type == "motorbike" {
		input.Seats = 0
	}
	if input.Year < 0 || input.Seats < 0 {
		return Vehicle{}, ErrInvalidInput
	}
	now := s.now()
	vehicle := Vehicle{
		ID: ids.New("veh"), OwnerUserID: input.OwnerUserID, Type: input.Type,
		LicensePlate: input.LicensePlate, Brand: strings.TrimSpace(input.Brand), Model: strings.TrimSpace(input.Model),
		Year: input.Year, Color: strings.TrimSpace(input.Color), Transmission: input.Transmission, Seats: input.Seats,
		Notes: strings.TrimSpace(input.Notes), PhotoObjectKey: strings.TrimSpace(input.PhotoObjectKey), Status: "active",
		CreatedAt: now, UpdatedAt: now,
	}
	if err := s.store.Create(vehicle); err != nil {
		return Vehicle{}, err
	}
	return vehicle, nil
}

func (s *Service) Get(id string) (Vehicle, error) {
	return s.store.Get(id)
}

func (s *Service) GetForOwner(id, ownerUserID string) (Vehicle, error) {
	vehicle, err := s.store.Get(id)
	if err != nil {
		return Vehicle{}, err
	}
	if vehicle.OwnerUserID != ownerUserID {
		return Vehicle{}, ErrForbidden
	}
	return vehicle, nil
}

func (s *Service) ListForOwner(ownerUserID string) ([]Vehicle, error) {
	if strings.TrimSpace(ownerUserID) == "" {
		return nil, ErrInvalidInput
	}
	return s.store.ListByOwner(ownerUserID)
}

func (s *Service) ValidateForService(id, ownerUserID, serviceType string) (Vehicle, error) {
	vehicle, err := s.GetForOwner(id, ownerUserID)
	if err != nil {
		return Vehicle{}, err
	}
	if vehicle.Status != "active" {
		return Vehicle{}, ErrInvalidInput
	}
	switch trips.NormalizeServiceType(serviceType) {
	case trips.ServiceDesignatedDriverCar, trips.ServiceVehicleInspection:
		if vehicle.Type != "car" {
			return Vehicle{}, ErrInvalidInput
		}
	case trips.ServiceDesignatedDriverBike:
		if vehicle.Type != "motorbike" {
			return Vehicle{}, ErrInvalidInput
		}
	default:
		return Vehicle{}, ErrInvalidInput
	}
	return vehicle, nil
}
