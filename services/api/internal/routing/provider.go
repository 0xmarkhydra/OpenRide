package routing

import (
	"errors"
	"math"

	"flashx/services/api/internal/trips"
)

var (
	ErrInvalidRoute = errors.New("invalid route")
	ErrUnavailable  = errors.New("routing provider unavailable")
)

type Result struct {
	DistanceM int64
	DurationS int64
	Source    string
}

type Provider interface {
	Route(pickup, destination trips.Point, serviceType string) (Result, error)
}

// FallbackProvider is intentionally development-only. It provides deterministic
// estimates without pretending straight-line geometry is a production route.
type FallbackProvider struct{}

func NewFallbackProvider() *FallbackProvider { return &FallbackProvider{} }

func (p *FallbackProvider) Route(pickup, destination trips.Point, serviceType string) (Result, error) {
	if !validPoint(pickup) || !validPoint(destination) {
		return Result{}, ErrInvalidRoute
	}
	averageSpeedKMH := 25.0
	if serviceType == "car" {
		averageSpeedKMH = 22
	}
	straightM := haversineMeters(pickup, destination)
	distanceM := int64(math.Round(straightM * 1.25))
	if distanceM < 1 {
		distanceM = 1
	}
	durationS := int64(math.Ceil((float64(distanceM) / 1000 / averageSpeedKMH) * 3600))
	if durationS < 1 {
		durationS = 1
	}
	return Result{DistanceM: distanceM, DurationS: durationS, Source: "development_fallback"}, nil
}

func validPoint(p trips.Point) bool {
	return p.Lat >= -90 && p.Lat <= 90 && p.Lng >= -180 && p.Lng <= 180
}

func haversineMeters(a, b trips.Point) float64 {
	const earthRadiusM = 6_371_000.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLng := math.Sin(dLng / 2)
	h := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLng*sinLng
	return 2 * earthRadiusM * math.Asin(math.Sqrt(h))
}
