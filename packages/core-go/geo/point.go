package geo

import "errors"

var ErrInvalidPoint = errors.New("geo: latitude or longitude is out of range")

// Point is a WGS84 coordinate used by the domain layer.
// Routing/provider-specific geometry belongs in adapters, not in Core.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (p Point) Validate() error {
	if p.Lat < -90 || p.Lat > 90 || p.Lng < -180 || p.Lng > 180 {
		return ErrInvalidPoint
	}
	return nil
}
