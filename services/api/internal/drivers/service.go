package drivers

import (
	"errors"
	"sort"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

var (
	ErrDriverUnavailable = errors.New("driver unavailable")
	ErrApprovalRequired  = errors.New("driver approval required")
	ErrInvalidLocation   = errors.New("invalid driver location")
	ErrInvalidInput      = errors.New("invalid driver input")
)

type NearbyDriver struct {
	Driver    Driver  `json:"driver"`
	DistanceM float64 `json:"distance_to_pickup_m"`
}

type QualificationInput struct {
	Capabilities   []string
	LicenseClass   string
	LicenseExpiry  *time.Time
	CanDriveManual bool
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
	var capabilities []string
	if strings.TrimSpace(serviceType) == "all" {
		serviceType = trips.ServiceDesignatedDriverCar
		capabilities = []string{trips.ServiceDesignatedDriverCar, trips.ServiceDesignatedDriverBike, trips.ServiceVehicleInspection}
	} else {
		serviceType = trips.NormalizeServiceType(serviceType)
		if !trips.IsSupportedService(serviceType) {
			return Driver{}, ErrInvalidInput
		}
		capabilities = capabilitiesForServiceType(serviceType)
	}
	driver := Driver{
		ID:           id,
		FullName:     fullName,
		ServiceType:  serviceType,
		Capabilities: capabilities,
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
	if phone == "" {
		return Driver{}, false, ErrInvalidInput
	}
	if driver, err := s.store.GetByPhone(phone); err == nil {
		return driver, false, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Driver{}, false, err
	}
	now := s.now()
	driver := Driver{
		ID: ids.New("drv"), Phone: phone, ServiceType: trips.ServiceDesignatedDriverCar,
		Capabilities: []string{},
		Approval:     ApprovalPending, Availability: AvailabilityOffline,
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
	if err != nil {
		return Driver{}, err
	}
	if status == ApprovalApproved && len(driver.Capabilities) == 0 {
		return Driver{}, ErrInvalidInput
	}
	driver.Approval = status
	if status != ApprovalApproved {
		driver.Availability = AvailabilityOffline
	}
	driver.UpdatedAt = s.now()
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}
	if status != ApprovalApproved {
		_ = s.locations.SetEligible(id, driver.ServiceType, false)
	}
	return s.Get(id)
}

func (s *Service) UpdateProfile(id, fullName, serviceType string) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	fullName = strings.TrimSpace(fullName)
	serviceType = strings.TrimSpace(serviceType)
	if len(fullName) > 160 {
		return Driver{}, ErrInvalidInput
	}
	if serviceType != "" {
		serviceType = trips.NormalizeServiceType(serviceType)
		if !trips.IsSupportedService(serviceType) {
			return Driver{}, ErrInvalidInput
		}
		driver.ServiceType = serviceType
	}
	// A driver may edit their public profile/service preference, but only
	// Operations may grant marketplace capabilities or manual-transmission
	// eligibility. Self-service profile edits must never widen dispatch access.
	driver.FullName = fullName
	driver.UpdatedAt = s.now()
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}
	return s.Get(id)
}

func (s *Service) UpdateQualifications(id string, input QualificationInput) (Driver, error) {
	driver, err := s.store.Get(id)
	if err != nil {
		return Driver{}, err
	}
	seen := make(map[string]bool)
	capabilities := make([]string, 0, len(input.Capabilities))
	for _, raw := range input.Capabilities {
		capability := trips.NormalizeServiceType(strings.TrimSpace(raw))
		if !trips.IsSupportedService(capability) {
			return Driver{}, ErrInvalidInput
		}
		if !seen[capability] {
			seen[capability] = true
			capabilities = append(capabilities, capability)
		}
	}
	licenseClass := strings.TrimSpace(input.LicenseClass)
	if len(licenseClass) > 32 {
		return Driver{}, ErrInvalidInput
	}
	if input.LicenseExpiry != nil {
		expiry := input.LicenseExpiry.UTC()
		input.LicenseExpiry = &expiry
	}
	driver.Capabilities = capabilities
	driver.LicenseClass = licenseClass
	driver.LicenseExpiry = input.LicenseExpiry
	driver.CanDriveManual = input.CanDriveManual
	if len(capabilities) > 0 {
		driver.ServiceType = capabilities[0]
	}
	if len(capabilities) == 0 && driver.Approval == ApprovalApproved {
		driver.Approval = ApprovalPending
		driver.Availability = AvailabilityOffline
		_ = s.locations.SetEligible(id, driver.ServiceType, false)
	}
	driver.UpdatedAt = s.now()
	if err := s.store.Save(driver); err != nil {
		return Driver{}, err
	}
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
	serviceType = trips.NormalizeServiceType(serviceType)
	locations, err := s.locations.Nearby(serviceType, center, radiusM, limit)
	if err != nil {
		return nil, err
	}
	now := s.now()
	result := make([]NearbyDriver, 0, len(locations))
	seen := map[string]bool{}
	for _, item := range locations {
		if !locationIsFresh(item.Location, now, maxLocationAge) {
			continue
		}
		driver, getErr := s.store.Get(item.DriverID)
		if getErr != nil || driver.Approval != ApprovalApproved || driver.Availability != AvailabilityOnline || !driverCanServe(driver, serviceType) {
			continue
		}
		driver.Location = &item.Location
		result = append(result, NearbyDriver{Driver: driver, DistanceM: item.DistanceM})
		seen[driver.ID] = true
	}
	// Compatibility path: a driver may be indexed under the legacy/primary service
	// while their approved capabilities include this service. This keeps the demo
	// and rolling migration operational without rewriting the GEO index first.
	all, listErr := s.store.List()
	if listErr != nil {
		return nil, listErr
	}
	for _, driver := range all {
		if seen[driver.ID] || driver.Approval != ApprovalApproved || driver.Availability != AvailabilityOnline || !driverCanServe(driver, serviceType) {
			continue
		}
		location, locationErr := s.locations.Get(driver.ID)
		if locationErr != nil || !locationIsFresh(location, now, maxLocationAge) {
			continue
		}
		distance := distanceMeters(center.Lat, center.Lng, location.Lat, location.Lng)
		if radiusM > 0 && distance > radiusM {
			continue
		}
		driver.Location = &location
		result = append(result, NearbyDriver{Driver: driver, DistanceM: distance})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DistanceM < result[j].DistanceM })
	if limit > 0 && len(result) > limit {
		result = result[:limit]
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
	serviceType = trips.NormalizeServiceType(serviceType)
	for _, driver := range items {
		if driver.Approval != ApprovalApproved || driver.Availability != AvailabilityOnline || !driverCanServe(driver, serviceType) {
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

func capabilitiesForServiceType(serviceType string) []string {
	serviceType = trips.NormalizeServiceType(serviceType)
	switch serviceType {
	case trips.ServiceVehicleInspection:
		return []string{trips.ServiceVehicleInspection, trips.ServiceDesignatedDriverCar}
	case trips.ServiceDesignatedDriverCar, trips.ServiceDesignatedDriverBike:
		return []string{serviceType}
	default:
		return nil
	}
}

func driverCanServe(driver Driver, serviceType string) bool {
	serviceType = trips.NormalizeServiceType(serviceType)
	capabilities := driver.Capabilities
	if len(capabilities) == 0 {
		return false
	}
	if driver.LicenseExpiry != nil && !driver.LicenseExpiry.After(time.Now().UTC()) {
		return false
	}
	has := func(target string) bool {
		for _, capability := range capabilities {
			if trips.NormalizeServiceType(capability) == target {
				return true
			}
		}
		return false
	}
	if serviceType == trips.ServiceVehicleInspection {
		return has(trips.ServiceVehicleInspection) && has(trips.ServiceDesignatedDriverCar)
	}
	return has(serviceType)
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
