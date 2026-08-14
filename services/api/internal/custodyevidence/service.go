package custodyevidence

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"flashx/services/api/internal/objectstorage"
	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

const maxPhotoSize = 15 * 1024 * 1024

var (
	ErrInvalidInput       = errors.New("invalid custody evidence input")
	ErrForbidden          = errors.New("custody evidence forbidden")
	ErrInvalidState       = errors.New("custody evidence invalid state")
	ErrStorageUnavailable = errors.New("custody evidence storage unavailable")
	ErrNotReady           = errors.New("custody evidence not ready")
)

var allowedPhotoTypes = map[string]struct{}{
	"front": {}, "rear": {}, "left": {}, "right": {}, "interior": {},
	"dashboard": {}, "document": {}, "other": {},
}

var allowedContentTypes = map[string]struct{}{
	"image/jpeg": {}, "image/png": {}, "image/webp": {},
}

type Service struct {
	store  Store
	signer objectstorage.Signer
	trips  *trips.Service
	now    func() time.Time
}

func NewService(store Store, signer objectstorage.Signer, tripService *trips.Service) *Service {
	return &Service{store: store, signer: signer, trips: tripService, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) UpdateByDriver(tripID, driverID string, stage Stage, input EvidenceInput) (Snapshot, error) {
	trip, err := s.driverTrip(tripID, driverID)
	if err != nil {
		return Snapshot{}, err
	}
	if !validStage(stage) || !canMutateStage(trip, stage) {
		return Snapshot{}, ErrInvalidState
	}
	input.ConditionNote = strings.TrimSpace(input.ConditionNote)
	if len([]rune(input.ConditionNote)) < 3 || !validMetrics(input) {
		return Snapshot{}, ErrInvalidInput
	}

	now := s.now()
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Snapshot{}, err
	}
	if errors.Is(err, ErrNotFound) {
		item = Evidence{ID: ids.New("custody"), TripID: tripID, Stage: stage, CreatedAt: now}
	}
	changed := item.ConditionNote != input.ConditionNote || !sameInt64Ptr(item.OdometerKm, input.OdometerKm) ||
		!sameIntPtr(item.FuelPercent, input.FuelPercent) || !sameIntPtr(item.BatteryPercent, input.BatteryPercent)
	item.ConditionNote = input.ConditionNote
	item.OdometerKm = cloneInt64Ptr(input.OdometerKm)
	item.FuelPercent = cloneIntPtr(input.FuelPercent)
	item.BatteryPercent = cloneIntPtr(input.BatteryPercent)
	if changed {
		item.DriverConfirmedAt = nil
		item.RiderConfirmedAt = nil
	}
	item.UpdatedAt = now
	item, err = s.store.UpsertEvidence(item)
	if err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(context.Background(), item)
}

func (s *Service) PreparePhotoUpload(ctx context.Context, tripID, driverID string, stage Stage, input PhotoUploadInput) (UploadTicket, error) {
	if s.signer == nil {
		return UploadTicket{}, ErrStorageUnavailable
	}
	trip, err := s.driverTrip(tripID, driverID)
	if err != nil {
		return UploadTicket{}, err
	}
	if !validStage(stage) || !canMutateStage(trip, stage) {
		return UploadTicket{}, ErrInvalidState
	}
	input = normalizePhotoInput(input)
	if !validPhotoInput(input) {
		return UploadTicket{}, ErrInvalidInput
	}
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil {
		return UploadTicket{}, err
	}
	photoID := ids.New("photo")
	key := buildObjectKey(tripID, stage, photoID, input.Filename)
	signed, err := s.signer.SignPut(ctx, key, input.ContentType, input.SizeBytes)
	if err != nil {
		return UploadTicket{}, fmt.Errorf("%w: %v", ErrStorageUnavailable, err)
	}
	return UploadTicket{EvidenceID: item.ID, PhotoID: photoID, PhotoType: input.PhotoType, ObjectKey: key, Upload: signed}, nil
}

func (s *Service) CompletePhotoUpload(tripID, driverID string, stage Stage, input CompletePhotoInput) (Snapshot, error) {
	trip, err := s.driverTrip(tripID, driverID)
	if err != nil {
		return Snapshot{}, err
	}
	if !validStage(stage) || !canMutateStage(trip, stage) {
		return Snapshot{}, ErrInvalidState
	}
	input.PhotoID = strings.TrimSpace(input.PhotoID)
	input.PhotoType = strings.TrimSpace(strings.ToLower(input.PhotoType))
	input.ObjectKey = strings.TrimSpace(input.ObjectKey)
	input.Filename = strings.TrimSpace(input.Filename)
	input.ContentType = strings.TrimSpace(strings.ToLower(input.ContentType))
	if input.PhotoID == "" || input.ObjectKey == "" || !validPhotoInput(PhotoUploadInput{
		PhotoType: input.PhotoType, Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes,
	}) || !objectKeyBelongsTo(tripID, stage, input.PhotoID, input.ObjectKey) {
		return Snapshot{}, ErrInvalidInput
	}
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil {
		return Snapshot{}, err
	}
	now := s.now()
	if err := s.store.CreatePhoto(Photo{
		ID: input.PhotoID, EvidenceID: item.ID, PhotoType: input.PhotoType, ObjectKey: input.ObjectKey,
		Filename: input.Filename, ContentType: input.ContentType, SizeBytes: input.SizeBytes, CreatedAt: now,
	}); err != nil {
		return Snapshot{}, err
	}
	// Adding a photo changes the evidence set. Any previous bilateral approval
	// must be collected again so confirmations always refer to the current set.
	item.DriverConfirmedAt = nil
	item.RiderConfirmedAt = nil
	item.UpdatedAt = now
	item, err = s.store.SaveEvidence(item)
	if err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(context.Background(), item)
}

func (s *Service) ConfirmDriver(tripID, driverID string, stage Stage) (Snapshot, error) {
	trip, err := s.driverTrip(tripID, driverID)
	if err != nil {
		return Snapshot{}, err
	}
	if !validStage(stage) || !canMutateStage(trip, stage) {
		return Snapshot{}, ErrInvalidState
	}
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil {
		return Snapshot{}, err
	}
	photos, err := s.store.ListPhotos(item.ID)
	if err != nil {
		return Snapshot{}, err
	}
	if !baseReady(item, photos) {
		return Snapshot{}, ErrNotReady
	}
	now := s.now()
	item.DriverConfirmedAt = &now
	item.UpdatedAt = now
	item, err = s.store.SaveEvidence(item)
	if err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(context.Background(), item)
}

func (s *Service) ConfirmRider(tripID, riderID string, stage Stage) (Snapshot, error) {
	trip, err := s.riderTrip(tripID, riderID)
	if err != nil {
		return Snapshot{}, err
	}
	if !validStage(stage) || !canMutateStage(trip, stage) {
		return Snapshot{}, ErrInvalidState
	}
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil {
		return Snapshot{}, err
	}
	photos, err := s.store.ListPhotos(item.ID)
	if err != nil {
		return Snapshot{}, err
	}
	if !baseReady(item, photos) || item.DriverConfirmedAt == nil {
		return Snapshot{}, ErrNotReady
	}
	now := s.now()
	item.RiderConfirmedAt = &now
	item.UpdatedAt = now
	item, err = s.store.SaveEvidence(item)
	if err != nil {
		return Snapshot{}, err
	}
	return s.snapshot(context.Background(), item)
}

func (s *Service) ListForDriver(ctx context.Context, tripID, driverID string) ([]Snapshot, error) {
	if _, err := s.driverTrip(tripID, driverID); err != nil {
		return nil, err
	}
	return s.list(ctx, tripID)
}

func (s *Service) ListForRider(ctx context.Context, tripID, riderID string) ([]Snapshot, error) {
	if _, err := s.riderTrip(tripID, riderID); err != nil {
		return nil, err
	}
	return s.list(ctx, tripID)
}

func (s *Service) ListForAdmin(ctx context.Context, tripID string) ([]Snapshot, error) {
	if s.trips == nil {
		return nil, ErrInvalidInput
	}
	if _, err := s.trips.Get(tripID); err != nil {
		return nil, translateTripError(err)
	}
	return s.list(ctx, tripID)
}

func (s *Service) IsReady(tripID string, stage Stage) (bool, error) {
	item, err := s.store.GetByTripStage(tripID, stage)
	if err != nil {
		return false, err
	}
	photos, err := s.store.ListPhotos(item.ID)
	if err != nil {
		return false, err
	}
	return baseReady(item, photos) && item.DriverConfirmedAt != nil && item.RiderConfirmedAt != nil, nil
}

func (s *Service) list(ctx context.Context, tripID string) ([]Snapshot, error) {
	items, err := s.store.ListByTrip(tripID)
	if err != nil {
		return nil, err
	}
	result := make([]Snapshot, 0, len(items))
	for _, item := range items {
		snapshot, err := s.snapshot(ctx, item)
		if err != nil {
			return nil, err
		}
		result = append(result, snapshot)
	}
	return result, nil
}

func (s *Service) snapshot(ctx context.Context, item Evidence) (Snapshot, error) {
	photos, err := s.store.ListPhotos(item.ID)
	if err != nil {
		return Snapshot{}, err
	}
	views := make([]PhotoWithURL, 0, len(photos))
	for _, photo := range photos {
		entry := PhotoWithURL{Photo: photo}
		if s.signer != nil {
			signed, signErr := s.signer.SignGet(ctx, photo.ObjectKey)
			if signErr != nil {
				return Snapshot{}, fmt.Errorf("%w: %v", ErrStorageUnavailable, signErr)
			}
			entry.View = &signed
		}
		views = append(views, entry)
	}
	return Snapshot{
		Evidence: item,
		Photos:   views,
		Ready:    baseReady(item, photos) && item.DriverConfirmedAt != nil && item.RiderConfirmedAt != nil,
	}, nil
}

func (s *Service) driverTrip(tripID, driverID string) (trips.Trip, error) {
	if s.trips == nil || strings.TrimSpace(driverID) == "" {
		return trips.Trip{}, ErrInvalidInput
	}
	trip, err := s.trips.GetForDriver(tripID, driverID)
	if err != nil {
		return trips.Trip{}, translateTripError(err)
	}
	return trip, nil
}

func (s *Service) riderTrip(tripID, riderID string) (trips.Trip, error) {
	if s.trips == nil || strings.TrimSpace(riderID) == "" {
		return trips.Trip{}, ErrInvalidInput
	}
	trip, err := s.trips.GetForRider(tripID, riderID)
	if err != nil {
		return trips.Trip{}, translateTripError(err)
	}
	return trip, nil
}

func translateTripError(err error) error {
	switch {
	case errors.Is(err, trips.ErrForbidden):
		return ErrForbidden
	case errors.Is(err, trips.ErrInvalidState):
		return ErrInvalidState
	case errors.Is(err, trips.ErrInvalidInput):
		return ErrInvalidInput
	case errors.Is(err, trips.ErrNotFound):
		return ErrNotFound
	default:
		return err
	}
}

func validStage(stage Stage) bool { return stage == StagePickup || stage == StageReturn }

func canMutateStage(trip trips.Trip, stage Stage) bool {
	if stage == StagePickup {
		switch trip.Status {
		case trips.StatusArrived, trips.StatusArrivedForPickup, trips.StatusVehicleReceived:
			return true
		default:
			return false
		}
	}
	if trips.IsInspectionService(trip.ServiceType) {
		switch trip.Status {
		case trips.StatusArrivedForReturn, trips.StatusHandover:
			return true
		default:
			return false
		}
	}
	switch trip.Status {
	case trips.StatusInProgress, trips.StatusHandover:
		return true
	default:
		return false
	}
}

func validMetrics(input EvidenceInput) bool {
	if input.OdometerKm != nil && *input.OdometerKm < 0 {
		return false
	}
	if input.FuelPercent != nil && (*input.FuelPercent < 0 || *input.FuelPercent > 100) {
		return false
	}
	if input.BatteryPercent != nil && (*input.BatteryPercent < 0 || *input.BatteryPercent > 100) {
		return false
	}
	return true
}

func baseReady(item Evidence, photos []Photo) bool {
	return len([]rune(strings.TrimSpace(item.ConditionNote))) >= 3 && len(photos) >= 2
}

func normalizePhotoInput(input PhotoUploadInput) PhotoUploadInput {
	input.PhotoType = strings.TrimSpace(strings.ToLower(input.PhotoType))
	input.Filename = strings.TrimSpace(input.Filename)
	input.ContentType = strings.TrimSpace(strings.ToLower(input.ContentType))
	return input
}

func validPhotoInput(input PhotoUploadInput) bool {
	if _, ok := allowedPhotoTypes[input.PhotoType]; !ok {
		return false
	}
	if _, ok := allowedContentTypes[input.ContentType]; !ok {
		return false
	}
	return input.Filename != "" && len(input.Filename) <= 180 && input.SizeBytes > 0 && input.SizeBytes <= maxPhotoSize
}

func buildObjectKey(tripID string, stage Stage, photoID, filename string) string {
	return fmt.Sprintf("flashx/custody/trips/%s/%s/%s-%s", safeSegment(tripID), stage, photoID, safeFilename(filename))
}

func objectKeyBelongsTo(tripID string, stage Stage, photoID, objectKey string) bool {
	prefix := fmt.Sprintf("flashx/custody/trips/%s/%s/", safeSegment(tripID), stage)
	return strings.HasPrefix(objectKey, prefix) && strings.Contains(filepath.Base(objectKey), photoID+"-")
}

func safeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	ext := strings.ToLower(filepath.Ext(name))
	base := strings.TrimSuffix(name, filepath.Ext(name))
	base = strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '-', r == '_':
			return r
		case unicode.IsSpace(r):
			return '-'
		default:
			return -1
		}
	}, base)
	base = strings.Trim(base, "-_")
	if base == "" {
		base = "photo"
	}
	if len(base) > 80 {
		base = base[:80]
	}
	if len(ext) > 10 {
		ext = ""
	}
	return base + ext
}

func safeSegment(value string) string {
	value = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, strings.TrimSpace(value))
	return strings.Trim(value, "-")
}

func sameIntPtr(a, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func sameInt64Ptr(a, b *int64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func cloneIntPtr(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneInt64Ptr(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
