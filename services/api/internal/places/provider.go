package places

import (
	"context"
	"errors"

	"flashx/services/api/internal/trips"
)

var (
	ErrInvalidQuery = errors.New("invalid place query")
	ErrUnavailable  = errors.New("place search unavailable")
)

type Result struct {
	PlaceID  string      `json:"place_id"`
	Name     string      `json:"name"`
	Address  string      `json:"address"`
	Location trips.Point `json:"location"`
	Source   string      `json:"source"`
}

type Provider interface {
	Name() string
	Search(ctx context.Context, query string, bias *trips.Point) ([]Result, error)
}

type DisabledProvider struct{}

func (DisabledProvider) Name() string { return "disabled" }

func (DisabledProvider) Search(context.Context, string, *trips.Point) ([]Result, error) {
	return nil, ErrUnavailable
}
