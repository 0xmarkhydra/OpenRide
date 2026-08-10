package drivers

import (
	"errors"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
)

var (
	ErrDriverUnavailable = errors.New("driver unavailable")
	ErrApprovalRequired  = errors.New("driver approval required")
	ErrInvalidLocation   = errors.New("invalid driver location")
	ErrInvalidInput      = errors.New("invalid driver input")
)

type NearbyDriver struct {
	Driver    Driver
	DistanceM float64
}

type Service struct {
	store     Store
	locations LocationIndex
	now       func() time.Time
}

func NewService(store Store) *Service {
	return NewServiceWithLocationIndex(store, NewMemoryLocationIndex())
}

func NewServiceWithLocationIndex(store Store, locations LocationIndex) *Service {
	if locations == nil {
		locations = NewMemoryLocationIndex()
	}
	return &Service{
		store:     store,
		locations: locations,
		now:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) RegisterApproved(id, fullName, serviceType string) (Driver, error) {
	if id == "" || serviceType == "" {
		return Driver{}, ErrInvalidInput
	}
	now := s.now()
	driver := Driver{
		ID:           id,
		FullName:     fullName,
		ServiceType:  serviceType,
		Approval:     ApprovalApproved,
		Availability: AvailabilityOffline,
		LastIdleAt:   now,
		CreatedAt:    now,
		UpdatedAt:    now,
		Version:      1,
	}
	if err := s.store.Create(driver); err != nil {
		return Driver{}, err
	}
	return driver, nil
}

func (s *Service) FindOrCreatePendingByPhone(phone string) (Driver, bool, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" { return Driver{}, false, ErrInvalidInput }
	if driver, err := s.store.GetByPhone(phone); err == nil {
		return driver, false, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Driver{}, false, err
	}
	now := s.now()
	driver := Driver{
		ID: ids.New("drv"), Phone: phone, ServiceType: "bike",
		Approval: ApprovalPending, Availability: AvailabilityOffline,
		LastIdleAt: now, CreatedAt: now, UpdatedAt: now, Version: 1,
	}
	if err := s.store.Create(driver); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			existing, getErr := s.store.GetByPhone(phone)
			return existing, false, getErr
		}
		return Driver{}, false, err
	}
	return driver, true, nil
}

func (s *Service) SetApproval(id string, status ApprovalStatus) (Driver, error) {
	if status != ApprovalApproved && status != ApprovalRejected && status != ApprovalSuspended && status != ApprovalPending {
		return Driver{}, ErrInvalidInput
	}
	driver, err := s.store.Get(id)
	if err != nil { return Driver{}, err }
	driver.Approval = status
	if status != ApprovalApproved {
		driver.Availability = AvailabilityOffline
	}
	driver.UpdatedAt = s.now()
	if err := s.store.Save(driver); err != nil { return Driver{}, err }
	if status != ApprovalApproved {
		_ = s.locations.SetEligible(id, driver.ServiceType, false)
	}
	return s.Get(id)
}

func (s *Service) UpdateProfile(id, fullName, serviceType string) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil { return Driver{}, err }
	fullName = strings.TrimSpace(fullName)
	serviceType = strings.TrimSpace(serviceType)
	if len(fullName) > 160 || (serviceType != "bike" && serviceType != "car") {
		return Driver{}, ErrInvalidInput
	}
	driver.FullName = fullName
	driver.ServiceType = serviceType
	driver.UpdatedAt = s.now()
	if err := s.store.Save(driver); err != nil { return Driver{}, err }
	return s.Get(id)
}

func (s *Service) ListAll() ([]Driver, error) {
	items, err := s.store.List()
	if err != nil {
		return nil, err
	}
	result := make([]Driver, 0, len(items))
	for _, driver := range items {
		if location, locationErr := s.locations.Get(driver.ID); locationErr == nil {
			driver.Location = &location
		}
		result = append(result, driver)
	}
	return result, nil
}

func (s *Service) Get(id string) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	if location, locationErr := s.locations.Get(id); locationErr == nil {
		driver.Location = &location
	}
	return driver, nil
}

func (s *Service) SetAvailability(id string, status AvailabilityStatus) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	if driver.Approval != ApprovalApproved {
		return Driver{}, ErrApprovalRequired
	}
	if status != AvailabilityOnline && status != AvailabilityOffline {
		return Driver{}, ErrInvalidInput
	}
	if driver.Availability == AvailabilityBusy {
		return Driver{}, ErrDriverUnavailable
	}

	now := s.now()
	driver.Availability = status
	if status == AvailabilityOnline {
		driver.LastIdleAt = now
	}
	driver.UpdatedAt = now
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}

	if location, locationErr := s.locations.Get(id); locationErr == nil {
		_ = s.locations.SetEligible(id, driver.ServiceType, status == AvailabilityOnline && locationIsFresh(location, now, 20*time.Second))
	}
	return s.Get(id)
}

func (s *Service) UpdateLocation(id string, location Location) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	if !validLocation(location) {
		return Driver{}, ErrInvalidLocation
	}
	now := s.now()
	if location.CapturedAt.IsZero() {
		location.CapturedAt = now
	}
	if location.CapturedAt.Before(now.Add(-2*time.Minute)) || location.CapturedAt.After(now.Add(30*time.Second)) {
		return Driver{}, ErrInvalidLocation
	}

	driver.Location = &location
	driver.UpdatedAt = now
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}
	eligible := driver.Approval == ApprovalApproved && driver.Availability == AvailabilityOnline
	if err := s.locations.Upsert(id, driver.ServiceType, location, eligible); err != nil {
		return Driver{}, err
	}
	return s.Get(id)
}

func (s *Service) NearbyCandidates(serviceType string, center Location, radiusM float64, maxLocationAge time.Duration, limit int) ([]NearbyDriver, error) {
	locations, err := s.locations.Nearby(serviceType, center, radiusM, limit)
	if err != nil {
		return nil, err
	}
	now := s.now()
	result := make([]NearbyDriver, 0, len(locations))
	for _, item := range locations {
		if !locationIsFresh(item.Location, now, maxLocationAge) {
			continue
		}
		driver, err := s.store.Get(item.DriverID)
		if err != nil {
			continue
		}
		if driver.Approval != ApprovalApproved || driver.Availability != AvailabilityOnline || driver.ServiceType != serviceType {
			continue
		}
		driver.Location = &item.Location
		result = append(result, NearbyDriver{Driver: driver, DistanceM: item.DistanceM})
	}
	return result, nil
}

// Candidates is retained for non-spatial administrative/debug use. Dispatch
// should use NearbyCandidates so production can leverage Redis GEO.
func (s *Service) Candidates(serviceType string, maxLocationAge time.Duration) []Driver {
	items, err := s.store.List()
	if err != nil {
		return nil
	}
	result := make([]Driver, 0)
	for _, driver := range items {
		if driver.Approval != ApprovalApproved || driver.Availability != AvailabilityOnline || driver.ServiceType != serviceType {
			continue
		}
		location, err := s.locations.Get(driver.ID)
		if err != nil || !locationIsFresh(location, s.now(), maxLocationAge) {
			continue
		}
		driver.Location = &location
		result = append(result, driver)
	}
	return result
}

func (s *Service) TryMarkBusy(id string) (Driver, error) {
	driver, err := s.store.TryMarkBusy(id)
	if err != nil {
		return Driver{}, err
	}
	if err := s.locations.SetEligible(id, driver.ServiceType, false); err != nil && !errors.Is(err, ErrLocationNotFound) {
		return Driver{}, err
	}
	return s.Get(id)
}

func (s *Service) MarkAvailable(id string) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	if driver.Approval != ApprovalApproved {
		return Driver{}, ErrApprovalRequired
	}
	now := s.now()
	driver.Availability = AvailabilityOnline
	driver.LastIdleAt = now
	driver.UpdatedAt = now
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}
	if location, locationErr := s.locations.Get(id); locationErr == nil {
		_ = s.locations.SetEligible(id, driver.ServiceType, locationIsFresh(location, now, 20*time.Second))
	}
	return s.Get(id)
}

func validLocation(location Location) bool {
	if location.Lat < -90 || location.Lat > 90 || location.Lng < -180 || location.Lng > 180 {
		return false
	}
	if location.AccuracyM < 0 || location.SpeedMPS < 0 || location.HeadingDeg < 0 || location.HeadingDeg > 360 {
		return false
	}
	return true
}
