package places

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"flashx/services/api/internal/trips"
)

const googleTextSearchEndpoint = "https://places.googleapis.com/v1/places:searchText"

type GoogleProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewGoogleProvider(apiKey string) (*GoogleProvider, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("google places api key is required")
	}
	return &GoogleProvider{apiKey: apiKey, endpoint: googleTextSearchEndpoint, client: &http.Client{Timeout: 5 * time.Second}}, nil
}

func newGoogleProviderForTest(apiKey, endpoint string, client *http.Client) *GoogleProvider {
	return &GoogleProvider{apiKey: apiKey, endpoint: endpoint, client: client}
}

func (p *GoogleProvider) Name() string { return "google_places" }

type googleTextSearchRequest struct {
	TextQuery    string              `json:"textQuery"`
	LanguageCode string              `json:"languageCode,omitempty"`
	RegionCode   string              `json:"regionCode,omitempty"`
	PageSize     int                 `json:"pageSize,omitempty"`
	LocationBias *googleLocationBias `json:"locationBias,omitempty"`
}

type googleLocationBias struct {
	Circle googleCircle `json:"circle"`
}

type googleCircle struct {
	Center googleLatLng `json:"center"`
	Radius float64      `json:"radius"`
}

type googleLatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type googleTextSearchResponse struct {
	Places []struct {
		ID               string `json:"id"`
		FormattedAddress string `json:"formattedAddress"`
		DisplayName      struct {
			Text string `json:"text"`
		} `json:"displayName"`
		Location googleLatLng `json:"location"`
	} `json:"places"`
}

func (p *GoogleProvider) Search(ctx context.Context, query string, bias *trips.Point) ([]Result, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 3 || len([]rune(query)) > 160 {
		return nil, ErrInvalidQuery
	}
	input := googleTextSearchRequest{TextQuery: query, LanguageCode: "vi", RegionCode: "vn", PageSize: 5}
	if bias != nil && validPoint(*bias) {
		input.LocationBias = &googleLocationBias{Circle: googleCircle{Center: googleLatLng{Latitude: bias.Lat, Longitude: bias.Lng}, Radius: 50000}}
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set("X-Goog-FieldMask", "places.id,places.displayName,places.formattedAddress,places.location")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: google places status %d", ErrUnavailable, resp.StatusCode)
	}
	var decoded googleTextSearchResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("%w: invalid google places response", ErrUnavailable)
	}
	results := make([]Result, 0, len(decoded.Places))
	for _, place := range decoded.Places {
		point := trips.Point{Lat: place.Location.Latitude, Lng: place.Location.Longitude}
		if place.ID == "" || !validPoint(point) {
			continue
		}
		results = append(results, Result{PlaceID: place.ID, Name: strings.TrimSpace(place.DisplayName.Text), Address: strings.TrimSpace(place.FormattedAddress), Location: point, Source: "google_places"})
	}
	return results, nil
}

func validPoint(point trips.Point) bool {
	return point.Lat >= -90 && point.Lat <= 90 && point.Lng >= -180 && point.Lng <= 180 && (point.Lat != 0 || point.Lng != 0)
}
