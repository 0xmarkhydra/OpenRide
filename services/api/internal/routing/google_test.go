package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"flashx/services/api/internal/trips"
)

func TestGoogleProviderRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Goog-Api-Key"); got != "test-key" {
			t.Fatalf("api key=%q", got)
		}
		if got := r.Header.Get("X-Goog-FieldMask"); got == "" {
			t.Fatal("missing field mask")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"routes":[{"distanceMeters":5234,"duration":"812.4s"}]}`))
	}))
	defer server.Close()

	provider := newGoogleProviderForTest("test-key", server.URL, server.Client())
	result, err := provider.Route(
		trips.Point{Lat: 21.0285, Lng: 105.8542},
		trips.Point{Lat: 21.0350, Lng: 105.8100},
		"bike",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.DistanceM != 5234 || result.DurationS != 813 || result.Source != "google_routes" {
		t.Fatalf("unexpected route: %+v", result)
	}
}

func TestGoogleProviderRejectsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	provider := newGoogleProviderForTest("test-key", server.URL, server.Client())
	if _, err := provider.Route(trips.Point{}, trips.Point{Lat: 1, Lng: 1}, "car"); err == nil {
		t.Fatal("expected provider error")
	}
}

func TestParseGoogleDuration(t *testing.T) {
	seconds, err := parseGoogleDuration("1.001s")
	if err != nil || seconds != 2 {
		t.Fatalf("duration=%d err=%v", seconds, err)
	}
}
