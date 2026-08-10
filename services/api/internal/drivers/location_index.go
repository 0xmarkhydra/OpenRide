package drivers

import (
	"errors"
	"math"
	"sort"
	"sync"
	"time"
)

var ErrLocationNotFound = errors.New("driver location not found")

type NearbyLocation struct {
	DriverID  string
	Location  Location
	DistanceM float64
}

// LocationIndex separates latest-location storage from dispatch eligibility.
// Busy drivers keep updating their latest location for trip tracking while
// being removed from the eligible GEO set used by dispatch.
type LocationIndex interface {
	Upsert(driverID, serviceType string, location Location, eligible bool) error
	SetEligible(driverID, serviceType string, eligible bool) error
	Get(driverID string) (Location, error)
	Nearby(serviceType string, center Location, radiusM float64, limit int) ([]NearbyLocation, error)
}

type memoryLocationEntry struct {
	serviceType string
	location    Location
	eligible    bool
}

type MemoryLocationIndex struct {
	mu      sync.RWMutex
	entries map[string]memoryLocationEntry
}

func NewMemoryLocationIndex() *MemoryLocationIndex {
	return &MemoryLocationIndex{entries: make(map[string]memoryLocationEntry)}
}

func (m *MemoryLocationIndex) Upsert(driverID, serviceType string, location Location, eligible bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[driverID] = memoryLocationEntry{serviceType: serviceType, location: location, eligible: eligible}
	return nil
}

func (m *MemoryLocationIndex) SetEligible(driverID, serviceType string, eligible bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[driverID]
	if !ok {
		return ErrLocationNotFound
	}
	entry.serviceType = serviceType
	entry.eligible = eligible
	m.entries[driverID] = entry
	return nil
}

func (m *MemoryLocationIndex) Get(driverID string) (Location, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	entry, ok := m.entries[driverID]
	if !ok {
		return Location{}, ErrLocationNotFound
	}
	return entry.location, nil
}

func (m *MemoryLocationIndex) Nearby(serviceType string, center Location, radiusM float64, limit int) ([]NearbyLocation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	result := make([]NearbyLocation, 0)
	for driverID, entry := range m.entries {
		if !entry.eligible || entry.serviceType != serviceType {
			continue
		}
		distance := distanceMeters(center.Lat, center.Lng, entry.location.Lat, entry.location.Lng)
		if radiusM > 0 && distance > radiusM {
			continue
		}
		result = append(result, NearbyLocation{DriverID: driverID, Location: entry.location, DistanceM: distance})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DistanceM < result[j].DistanceM })
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func locationIsFresh(location Location, now time.Time, maxAge time.Duration) bool {
	if maxAge <= 0 {
		return true
	}
	return !location.CapturedAt.IsZero() && now.Sub(location.CapturedAt) <= maxAge && location.CapturedAt.Before(now.Add(30*time.Second))
}

func distanceMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusM = 6_371_000.0
	p1 := lat1 * math.Pi / 180
	p2 := lat2 * math.Pi / 180
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * earthRadiusM * math.Asin(math.Sqrt(a))
}
