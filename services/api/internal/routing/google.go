package routing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"flashx/services/api/internal/trips"
)

const googleRoutesEndpoint = "https://routes.googleapis.com/directions/v2:computeRoutes"

type GoogleProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewGoogleProvider(apiKey string) (*GoogleProvider, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("google routes api key is required")
	}
	return &GoogleProvider{
		apiKey:   apiKey,
		endpoint: googleRoutesEndpoint,
		client:   &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func newGoogleProviderForTest(apiKey, endpoint string, client *http.Client) *GoogleProvider {
	return &GoogleProvider{apiKey: apiKey, endpoint: endpoint, client: client}
}

type googleRouteRequest struct {
	Origin      googleWaypoint `json:"origin"`
	Destination googleWaypoint `json:"destination"`
	TravelMode  string         `json:"travelMode"`
	Language    string         `json:"languageCode,omitempty"`
	Units       string         `json:"units,omitempty"`
}

type googleWaypoint struct {
	Location googleLocation `json:"location"`
}

type googleLocation struct {
	LatLng googleLatLng `json:"latLng"`
}

type googleLatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type googleRouteResponse struct {
	Routes []struct {
		DistanceMeters int64  `json:"distanceMeters"`
		Duration       string `json:"duration"`
		Polyline       struct {
			EncodedPolyline string `json:"encodedPolyline"`
		} `json:"polyline"`
	} `json:"routes"`
}

func (p *GoogleProvider) Route(pickup, destination trips.Point, serviceType string) (Result, error) {
	if !validPoint(pickup) || !validPoint(destination) {
		return Result{}, ErrInvalidRoute
	}
	travelMode := "DRIVE"
	if serviceType == "bike" {
		travelMode = "TWO_WHEELER"
	}
	payload, err := json.Marshal(googleRouteRequest{
		Origin:      googleWaypoint{Location: googleLocation{LatLng: googleLatLng{Latitude: pickup.Lat, Longitude: pickup.Lng}}},
		Destination: googleWaypoint{Location: googleLocation{LatLng: googleLatLng{Latitude: destination.Lat, Longitude: destination.Lng}}},
		TravelMode:  travelMode,
		Language:    "vi-VN",
		Units:       "METRIC",
	})
	if err != nil {
		return Result{}, err
	}
	req, err := http.NewRequest(http.MethodPost, p.endpoint, bytes.NewReader(payload))
	if err != nil {
		return Result{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", p.apiKey)
	req.Header.Set("X-Goog-FieldMask", "routes.distanceMeters,routes.duration,routes.polyline.encodedPolyline")

	resp, err := p.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{}, fmt.Errorf("%w: google routes status %d", ErrUnavailable, resp.StatusCode)
	}
	var decoded googleRouteResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return Result{}, fmt.Errorf("%w: invalid google response", ErrUnavailable)
	}
	if len(decoded.Routes) == 0 || decoded.Routes[0].DistanceMeters <= 0 {
		return Result{}, ErrUnavailable
	}
	durationS, err := parseGoogleDuration(decoded.Routes[0].Duration)
	if err != nil || durationS <= 0 {
		return Result{}, ErrUnavailable
	}
	return Result{
		DistanceM:       decoded.Routes[0].DistanceMeters,
		DurationS:       durationS,
		Source:          "google_routes",
		EncodedPolyline: decoded.Routes[0].Polyline.EncodedPolyline,
	}, nil
}

func parseGoogleDuration(value string) (int64, error) {
	value = strings.TrimSuffix(strings.TrimSpace(value), "s")
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds < 0 {
		return 0, fmt.Errorf("invalid duration")
	}
	return int64(math.Ceil(seconds)), nil
}
