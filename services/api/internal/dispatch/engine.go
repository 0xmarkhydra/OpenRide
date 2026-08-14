package dispatch

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/trips"
)

var (
	ErrNoCandidate      = errors.New("no dispatch candidate")
	ErrOfferNotFound    = errors.New("offer not found")
	ErrOfferExpired     = errors.New("offer expired")
	ErrOfferForbidden   = errors.New("offer does not belong to driver")
	ErrOfferUnavailable = errors.New("offer unavailable")
)

type OfferStatus string

const (
	OfferPending  OfferStatus = "pending"
	OfferAccepted OfferStatus = "accepted"
	OfferRejected OfferStatus = "rejected"
	OfferExpired  OfferStatus = "expired"
	OfferInvalid  OfferStatus = "invalidated"
)

type Offer struct {
	ID         string      `json:"id"`
	TripID     string      `json:"trip_id"`
	DriverID   string      `json:"driver_id"`
	Status     OfferStatus `json:"status"`
	DistanceM  int64       `json:"distance_to_pickup_m"`
	CreatedAt  time.Time   `json:"created_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
	AcceptedAt *time.Time  `json:"accepted_at,omitempty"`
}

type Engine struct {
	// mu protects one Engine instance. Production also uses Locker so separate
	// API instances serialize mutations for the same trip.
	mu             sync.Mutex
	drivers        *drivers.Service
	trips          *trips.Service
	offers         OfferStore
	locker         Locker
	now            func() time.Time
	offerTTL       time.Duration
	lockTTL        time.Duration
	maxDistance    float64
	locationMaxAge time.Duration
}

func NewEngine(driverService *drivers.Service, tripService *trips.Service) *Engine {
	return NewEngineWithStore(driverService, tripService, NewMemoryOfferStore(), NewMemoryLocker())
}

func NewEngineWithStore(driverService *drivers.Service, tripService *trips.Service, offers OfferStore, locker Locker) *Engine {
	if offers == nil {
		offers = NewMemoryOfferStore()
	}
	if locker == nil {
		locker = NewMemoryLocker()
	}
	return &Engine{
		drivers:        driverService,
		trips:          tripService,
		offers:         offers,
		locker:         locker,
		now:            func() time.Time { return time.Now().UTC() },
		offerTTL:       12 * time.Second,
		lockTTL:        3 * time.Second,
		maxDistance:    5_000,
		locationMaxAge: 20 * time.Second,
	}
}

func (e *Engine) UpdatePolicy(maxDistanceM int64, locationMaxAgeSeconds int) {
	if maxDistanceM <= 0 || locationMaxAgeSeconds <= 0 {
		return
	}
	e.mu.Lock()
	e.maxDistance = float64(maxDistanceM)
	e.locationMaxAge = time.Duration(locationMaxAgeSeconds) * time.Second
	e.mu.Unlock()
}

func (e *Engine) CreateOffer(tripID string) (Offer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	unlock, acquired, err := e.locker.TryLock("trip:"+tripID, e.lockTTL)
	if err != nil {
		return Offer{}, err
	}
	if !acquired {
		if existing, ok, findErr := e.pendingOfferForTrip(tripID, e.now()); findErr == nil && ok {
			return existing, nil
		}
		return Offer{}, ErrOfferUnavailable
	}
	defer unlock()

	now := e.now()
	trip, err := e.trips.Get(tripID)
	if err != nil {
		return Offer{}, err
	}
	if trip.Status != trips.StatusSearching || trip.DriverID != "" {
		return Offer{}, trips.ErrInvalidState
	}

	existing, err := e.tripOffers(tripID, now)
	if err != nil {
		return Offer{}, err
	}
	for _, offer := range existing {
		if offer.Status == OfferPending && offer.ExpiresAt.After(now) {
			return offer, nil
		}
	}

	excluded := make(map[string]bool)
	for _, offer := range existing {
		if offer.Status != OfferAccepted {
			excluded[offer.DriverID] = true
		}
	}

	type candidate struct {
		driver   drivers.Driver
		distance float64
		score    float64
	}
	candidates := make([]candidate, 0)
	nearby, err := e.drivers.NearbyCandidates(
		trip.ServiceType,
		drivers.Location{Lat: trip.Pickup.Lat, Lng: trip.Pickup.Lng, CapturedAt: now},
		e.maxDistance,
		e.locationMaxAge,
		30,
	)
	if err != nil {
		return Offer{}, err
	}
	for _, nearbyDriver := range nearby {
		driver := nearbyDriver.Driver
		if excluded[driver.ID] {
			continue
		}
		distance := nearbyDriver.DistanceM
		idleSeconds := now.Sub(driver.LastIdleAt).Seconds()
		// Distance dominates MVP scoring; idle time only breaks close ties.
		score := distance - math.Min(idleSeconds, 1800)*0.5
		candidates = append(candidates, candidate{driver: driver, distance: distance, score: score})
	}
	if len(candidates) == 0 {
		return Offer{}, ErrNoCandidate
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].score < candidates[j].score })
	selected := candidates[0]

	offer := Offer{
		ID:        newID("offer"),
		TripID:    trip.ID,
		DriverID:  selected.driver.ID,
		Status:    OfferPending,
		DistanceM: int64(math.Round(selected.distance)),
		CreatedAt: now,
		ExpiresAt: now.Add(e.offerTTL),
	}
	if err := e.offers.Put(offer); err != nil {
		return Offer{}, err
	}
	return offer, nil
}

func (e *Engine) GetOffer(id, driverID string) (Offer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	offer, err := e.offers.Get(id)
	if err != nil {
		return Offer{}, err
	}
	if offer.DriverID != driverID {
		return Offer{}, ErrOfferForbidden
	}
	return e.normalizeOffer(offer, e.now())
}

func (e *Engine) Accept(offerID, driverID string) (Offer, trips.Trip, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	initial, err := e.offers.Get(offerID)
	if err != nil {
		return Offer{}, trips.Trip{}, err
	}
	if initial.DriverID != driverID {
		return Offer{}, trips.Trip{}, ErrOfferForbidden
	}
	unlock, acquired, err := e.locker.TryLock("trip:"+initial.TripID, e.lockTTL)
	if err != nil {
		return Offer{}, trips.Trip{}, err
	}
	if !acquired {
		return Offer{}, trips.Trip{}, ErrOfferUnavailable
	}
	defer unlock()

	now := e.now()
	offer, err := e.offers.Get(offerID)
	if err != nil {
		return Offer{}, trips.Trip{}, err
	}
	offer, err = e.normalizeOffer(offer, now)
	if err != nil {
		return Offer{}, trips.Trip{}, err
	}
	if offer.DriverID != driverID {
		return Offer{}, trips.Trip{}, ErrOfferForbidden
	}
	if offer.Status != OfferPending || !offer.ExpiresAt.After(now) {
		return Offer{}, trips.Trip{}, ErrOfferUnavailable
	}

	if _, err := e.drivers.TryMarkBusy(driverID); err != nil {
		return Offer{}, trips.Trip{}, err
	}
	trip, err := e.trips.AssignDriver(offer.TripID, driverID)
	if err != nil {
		_, _ = e.drivers.MarkAvailable(driverID)
		return Offer{}, trips.Trip{}, err
	}

	offer.Status = OfferAccepted
	offer.AcceptedAt = &now
	// The durable Trip/Driver records are the source of truth. If Redis becomes
	// unavailable after those writes, the ride must still be reported accepted;
	// readiness/monitoring can surface the degraded dispatch cache separately.
	_ = e.offers.Save(offer)
	others, _ := e.offers.ListByTrip(offer.TripID)
	for _, other := range others {
		if other.ID != offer.ID && other.Status == OfferPending {
			other.Status = OfferInvalid
			_ = e.offers.Save(other)
		}
	}
	return offer, trip, nil
}

func (e *Engine) Reject(offerID, driverID string) (Offer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	initial, err := e.offers.Get(offerID)
	if err != nil {
		return Offer{}, err
	}
	if initial.DriverID != driverID {
		return Offer{}, ErrOfferForbidden
	}
	unlock, acquired, err := e.locker.TryLock("trip:"+initial.TripID, e.lockTTL)
	if err != nil {
		return Offer{}, err
	}
	if !acquired {
		return Offer{}, ErrOfferUnavailable
	}
	defer unlock()

	now := e.now()
	offer, err := e.offers.Get(offerID)
	if err != nil {
		return Offer{}, err
	}
	offer, err = e.normalizeOffer(offer, now)
	if err != nil {
		return Offer{}, err
	}
	if offer.Status != OfferPending {
		return Offer{}, ErrOfferUnavailable
	}
	offer.Status = OfferRejected
	if err := e.offers.Save(offer); err != nil {
		return Offer{}, err
	}
	return offer, nil
}

func (e *Engine) CurrentOfferForDriver(driverID string) (Offer, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	items, err := e.offers.ListByDriver(driverID)
	if err != nil {
		return Offer{}, err
	}
	now := e.now()
	var current Offer
	found := false
	for _, item := range items {
		offer, normalizeErr := e.normalizeOffer(item, now)
		if normalizeErr != nil && !errors.Is(normalizeErr, ErrOfferExpired) {
			return Offer{}, normalizeErr
		}
		if offer.Status != OfferPending {
			continue
		}
		if !found || offer.CreatedAt.After(current.CreatedAt) {
			current = offer
			found = true
		}
	}
	if !found {
		return Offer{}, ErrOfferNotFound
	}
	return current, nil
}

func (e *Engine) DispatchWaiting(limit int) []Offer {
	waiting, err := e.trips.ListSearching(limit)
	if err != nil {
		return nil
	}
	created := make([]Offer, 0)
	for _, trip := range waiting {
		offer, err := e.CreateOffer(trip.ID)
		if err == nil {
			created = append(created, offer)
		}
	}
	return created
}

// CandidatesForTrip uses the same eligibility rules as automated dispatch:
// approved, online, capability-compatible and recently located near pickup.
func (e *Engine) CandidatesForTrip(tripID string, limit int) ([]drivers.NearbyDriver, error) {
	trip, err := e.trips.Get(tripID)
	if err != nil {
		return nil, err
	}
	if trip.IncidentOpen || (trip.Status != trips.StatusSearching && !trips.CanReassignBeforeCustody(trip.Status)) {
		return nil, trips.ErrInvalidState
	}
	if limit <= 0 || limit > 30 {
		limit = 15
	}
	e.mu.Lock()
	maxDistance := e.maxDistance
	locationMaxAge := e.locationMaxAge
	e.mu.Unlock()
	return e.drivers.NearbyCandidates(
		trip.ServiceType,
		drivers.Location{Lat: trip.Pickup.Lat, Lng: trip.Pickup.Lng, CapturedAt: e.now()},
		maxDistance,
		locationMaxAge,
		limit,
	)
}

// ManualAssign serializes with normal offer acceptance using the same trip
// lock. It supports first assignment while searching and safe reassignment
// before vehicle custody, then invalidates any stale pending offers.
func (e *Engine) ManualAssign(tripID, driverID string) (trips.Trip, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	unlock, acquired, err := e.locker.TryLock("trip:"+tripID, e.lockTTL)
	if err != nil {
		return trips.Trip{}, err
	}
	if !acquired {
		return trips.Trip{}, ErrOfferUnavailable
	}
	defer unlock()

	trip, err := e.trips.Get(tripID)
	if err != nil {
		return trips.Trip{}, err
	}
	if driverID == "" || trip.IncidentOpen || (trip.Status != trips.StatusSearching && !trips.CanReassignBeforeCustody(trip.Status)) || trip.DriverID == driverID {
		return trips.Trip{}, trips.ErrInvalidState
	}

	candidates, err := e.drivers.NearbyCandidates(
		trip.ServiceType,
		drivers.Location{Lat: trip.Pickup.Lat, Lng: trip.Pickup.Lng, CapturedAt: e.now()},
		e.maxDistance,
		e.locationMaxAge,
		30,
	)
	if err != nil {
		return trips.Trip{}, err
	}
	eligible := false
	for _, candidate := range candidates {
		if candidate.Driver.ID == driverID {
			eligible = true
			break
		}
	}
	if !eligible {
		return trips.Trip{}, drivers.ErrDriverUnavailable
	}

	if _, err := e.drivers.TryMarkBusy(driverID); err != nil {
		return trips.Trip{}, err
	}
	oldDriverID := trip.DriverID
	var updated trips.Trip
	if trip.Status == trips.StatusSearching {
		updated, err = e.trips.AssignDriver(tripID, driverID)
	} else {
		updated, err = e.trips.ReassignDriver(tripID, driverID)
	}
	if err != nil {
		_, _ = e.drivers.MarkAvailable(driverID)
		return trips.Trip{}, err
	}
	if oldDriverID != "" && oldDriverID != driverID {
		// Best effort: suspended/rejected drivers intentionally remain unavailable.
		_, _ = e.drivers.MarkAvailable(oldDriverID)
	}
	items, _ := e.offers.ListByTrip(tripID)
	for _, offer := range items {
		if offer.Status == OfferPending {
			offer.Status = OfferInvalid
			_ = e.offers.Save(offer)
		}
	}
	return updated, nil
}

func (e *Engine) pendingOfferForTrip(tripID string, now time.Time) (Offer, bool, error) {
	items, err := e.offers.ListByTrip(tripID)
	if err != nil {
		return Offer{}, false, err
	}
	for i := len(items) - 1; i >= 0; i-- {
		offer := items[i]
		if offer.Status == OfferPending && offer.ExpiresAt.After(now) {
			return offer, true, nil
		}
	}
	return Offer{}, false, nil
}

func (e *Engine) tripOffers(tripID string, now time.Time) ([]Offer, error) {
	items, err := e.offers.ListByTrip(tripID)
	if err != nil {
		return nil, err
	}
	for i := range items {
		normalized, normalizeErr := e.normalizeOffer(items[i], now)
		if normalizeErr != nil && !errors.Is(normalizeErr, ErrOfferExpired) {
			return nil, normalizeErr
		}
		items[i] = normalized
	}
	return items, nil
}

func (e *Engine) normalizeOffer(offer Offer, now time.Time) (Offer, error) {
	if offer.Status == OfferPending && !offer.ExpiresAt.After(now) {
		offer.Status = OfferExpired
		if err := e.offers.Save(offer); err != nil {
			return Offer{}, err
		}
		return offer, ErrOfferExpired
	}
	if offer.Status == OfferExpired {
		return offer, ErrOfferExpired
	}
	return offer, nil
}

func newID(prefix string) string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "_fallback"
	}
	return prefix + "_" + hex.EncodeToString(buf)
}
