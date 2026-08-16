package inspectionchecklist

import (
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

var (
	ErrInvalidInput = errors.New("inspection checklist invalid input")
	ErrForbidden    = errors.New("inspection checklist forbidden")
	ErrInvalidState = errors.New("inspection checklist invalid state")
	ErrNotReady     = errors.New("inspection checklist not ready")
)

var itemKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

type Service struct {
	store Store
	trips *trips.Service
	now   func() time.Time
}

func NewService(store Store, tripService *trips.Service) (*Service, error) {
	service := &Service{store: store, trips: tripService, now: func() time.Time { return time.Now().UTC() }}
	if err := store.EnsureDefaultTemplate(defaultTemplate(service.now())); err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) ActiveTemplate() (Template, error) { return s.store.GetActiveTemplate() }

func (s *Service) ReplaceTemplate(input ReplaceTemplateInput, actorID string) (Template, error) {
	input.Reason = strings.TrimSpace(input.Reason)
	items, err := normalizeTemplateItems(input.Items)
	if err != nil || input.Reason == "" || strings.TrimSpace(actorID) == "" {
		return Template{}, ErrInvalidInput
	}
	return s.store.CreateTemplate(items, actorID, input.Reason, s.now())
}

func (s *Service) ForRider(tripID, riderID string) (Snapshot, error) {
	trip, err := s.trips.GetForRider(tripID, riderID)
	if err != nil {
		return Snapshot{}, translateTripError(err)
	}
	return s.ensure(trip)
}

func (s *Service) ForDriver(tripID, driverID string) (Snapshot, error) {
	trip, err := s.trips.GetForDriver(tripID, driverID)
	if err != nil {
		return Snapshot{}, translateTripError(err)
	}
	return s.ensure(trip)
}

func (s *Service) ForAdmin(tripID string) (Snapshot, error) {
	trip, err := s.trips.Get(tripID)
	if err != nil {
		return Snapshot{}, translateTripError(err)
	}
	return s.ensure(trip)
}

func (s *Service) UpdateCustomer(tripID, riderID, itemKey string, input CustomerUpdate) (Snapshot, error) {
	trip, err := s.trips.GetForRider(tripID, riderID)
	if err != nil {
		return Snapshot{}, translateTripError(err)
	}
	if !canCustomerEdit(trip) {
		return Snapshot{}, ErrInvalidState
	}
	snapshot, err := s.ensure(trip)
	if err != nil {
		return Snapshot{}, err
	}
	input.Status = strings.TrimSpace(strings.ToLower(input.Status))
	input.Note = strings.TrimSpace(input.Note)
	if !validCustomerStatus(input.Status) || len([]rune(input.Note)) > 500 {
		return Snapshot{}, ErrInvalidInput
	}
	item, ok := findItem(snapshot.Items, itemKey)
	if !ok {
		return Snapshot{}, ErrNotFound
	}
	if item.Required && input.Status == CustomerNotApplicable {
		return Snapshot{}, ErrInvalidInput
	}
	now := s.now()
	changed := item.CustomerStatus != input.Status || item.CustomerNote != input.Note
	item.CustomerStatus = input.Status
	item.CustomerNote = input.Note
	item.CustomerUpdatedAt = &now
	if changed {
		// Driver verification refers to the customer's current declaration.
		item.DriverStatus = DriverPending
		item.DriverNote = ""
		item.DriverUpdatedAt = nil
	}
	if err := s.store.SaveItem(item); err != nil {
		return Snapshot{}, err
	}
	return s.load(trip.ID)
}

func (s *Service) UpdateDriver(tripID, driverID, itemKey string, input DriverUpdate) (Snapshot, error) {
	trip, err := s.trips.GetForDriver(tripID, driverID)
	if err != nil {
		return Snapshot{}, translateTripError(err)
	}
	if !canDriverVerify(trip) {
		return Snapshot{}, ErrInvalidState
	}
	snapshot, err := s.ensure(trip)
	if err != nil {
		return Snapshot{}, err
	}
	input.Status = strings.TrimSpace(strings.ToLower(input.Status))
	input.Note = strings.TrimSpace(input.Note)
	if !validDriverStatus(input.Status) || len([]rune(input.Note)) > 500 {
		return Snapshot{}, ErrInvalidInput
	}
	item, ok := findItem(snapshot.Items, itemKey)
	if !ok {
		return Snapshot{}, ErrNotFound
	}
	if item.Required && input.Status == DriverNotApplicable {
		return Snapshot{}, ErrInvalidInput
	}
	now := s.now()
	item.DriverStatus = input.Status
	item.DriverNote = input.Note
	item.DriverUpdatedAt = &now
	if err := s.store.SaveItem(item); err != nil {
		return Snapshot{}, err
	}
	return s.load(trip.ID)
}

func (s *Service) IsReady(tripID string) (bool, error) {
	snapshot, err := s.ForAdmin(tripID)
	if err != nil {
		return false, err
	}
	return snapshot.Ready, nil
}

func (s *Service) ensure(trip trips.Trip) (Snapshot, error) {
	if !trips.IsInspectionService(trip.ServiceType) {
		return Snapshot{}, ErrInvalidInput
	}
	snapshot, err := s.load(trip.ID)
	if err == nil {
		return snapshot, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Snapshot{}, err
	}
	template, err := s.store.GetActiveTemplate()
	if err != nil {
		return Snapshot{}, err
	}
	now := s.now()
	checklist := Checklist{TripID: trip.ID, TemplateID: template.ID, TemplateVersion: template.Version, CreatedAt: now, UpdatedAt: now}
	items := make([]Item, 0, len(template.Items))
	for _, templateItem := range template.Items {
		items = append(items, Item{
			ID: ids.New("inspection_item"), TripID: trip.ID, Key: templateItem.Key,
			Label: templateItem.Label, Required: templateItem.Required, SortOrder: templateItem.SortOrder,
			CustomerStatus: CustomerPending, DriverStatus: DriverPending,
		})
	}
	if err := s.store.CreateChecklist(checklist, items); err != nil {
		return Snapshot{}, err
	}
	return s.load(trip.ID)
}

func (s *Service) load(tripID string) (Snapshot, error) {
	checklist, items, err := s.store.GetChecklist(tripID)
	if err != nil {
		return Snapshot{}, err
	}
	customerComplete := true
	driverComplete := true
	ready := true
	for _, item := range items {
		if item.CustomerStatus == CustomerPending {
			customerComplete = false
		}
		if item.DriverStatus == DriverPending {
			driverComplete = false
		}
		if item.Required && (item.CustomerStatus != CustomerPresent || item.DriverStatus != DriverReceived) {
			ready = false
		}
	}
	return Snapshot{Checklist: checklist, Items: items, CustomerComplete: customerComplete, DriverComplete: driverComplete, Ready: ready}, nil
}

func normalizeTemplateItems(items []TemplateItem) ([]TemplateItem, error) {
	if len(items) == 0 || len(items) > 30 {
		return nil, ErrInvalidInput
	}
	seen := make(map[string]bool, len(items))
	result := make([]TemplateItem, 0, len(items))
	requiredCount := 0
	for index, item := range items {
		item.Key = strings.TrimSpace(strings.ToLower(item.Key))
		item.Label = strings.TrimSpace(item.Label)
		if !itemKeyPattern.MatchString(item.Key) || item.Label == "" || len([]rune(item.Label)) > 160 || seen[item.Key] {
			return nil, ErrInvalidInput
		}
		seen[item.Key] = true
		if item.SortOrder <= 0 {
			item.SortOrder = (index + 1) * 10
		}
		if item.Required {
			requiredCount++
		}
		result = append(result, item)
	}
	if requiredCount == 0 {
		return nil, ErrInvalidInput
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].SortOrder == result[j].SortOrder {
			return result[i].Key < result[j].Key
		}
		return result[i].SortOrder < result[j].SortOrder
	})
	return result, nil
}

func defaultTemplate(now time.Time) Template {
	return Template{
		ID: "inspection_checklist_v1", Version: 1, Active: true, Reason: "FlashX MVP default template", CreatedAt: now,
		Items: []TemplateItem{
			{Key: "vehicle_registration", Label: "Đăng ký xe / giấy tờ xe hợp lệ", Required: true, SortOrder: 10},
			{Key: "previous_inspection_docs", Label: "Giấy tờ đăng kiểm trước đây (nếu có)", SortOrder: 20},
			{Key: "authorization_docs", Label: "Giấy ủy quyền / giấy tờ đại diện (nếu áp dụng)", SortOrder: 30},
			{Key: "other_related_docs", Label: "Giấy tờ liên quan khác theo trường hợp", SortOrder: 40},
		},
	}
}

func canCustomerEdit(trip trips.Trip) bool {
	if !trips.IsInspectionService(trip.ServiceType) {
		return false
	}
	switch trip.Status {
	case trips.StatusScheduled, trips.StatusSearching, trips.StatusAccepted, trips.StatusArrivingForPickup, trips.StatusArrivedForPickup:
		return true
	default:
		return false
	}
}

func canDriverVerify(trip trips.Trip) bool {
	return trips.IsInspectionService(trip.ServiceType) && trip.Status == trips.StatusArrivedForPickup
}

func validCustomerStatus(status string) bool {
	switch status {
	case CustomerPending, CustomerPresent, CustomerNotAvailable, CustomerNotApplicable:
		return true
	default:
		return false
	}
}

func validDriverStatus(status string) bool {
	switch status {
	case DriverPending, DriverReceived, DriverMissing, DriverNotApplicable:
		return true
	default:
		return false
	}
}

func findItem(items []Item, key string) (Item, bool) {
	key = strings.TrimSpace(strings.ToLower(key))
	for _, item := range items {
		if item.Key == key {
			return item, true
		}
	}
	return Item{}, false
}

func translateTripError(err error) error {
	switch {
	case errors.Is(err, trips.ErrForbidden):
		return ErrForbidden
	case errors.Is(err, trips.ErrInvalidState):
		return ErrInvalidState
	case errors.Is(err, trips.ErrNotFound):
		return ErrNotFound
	default:
		return err
	}
}
