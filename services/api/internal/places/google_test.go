package places

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"flashx/services/api/internal/trips"
)

func TestGoogleTextSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s want POST", r.Method)
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-key" {
			t.Fatalf("missing api key header")
		}
		if !strings.Contains(r.Header.Get("X-Goog-FieldMask"), "places.location") {
			t.Fatalf("unexpected field mask: %s", r.Header.Get("X-Goog-FieldMask"))
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"textQuery":"Vincom Thanh Hóa"`) || !strings.Contains(string(body), `"radius":50000`) {
			t.Fatalf("unexpected request body: %s", body)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"places":[{"id":"place-1","displayName":{"text":"Vincom Plaza Thanh Hóa"},"formattedAddress":"27 Trần Phú, Thanh Hóa","location":{"latitude":19.8045,"longitude":105.7779}}]}`))
	}))
	defer server.Close()

	provider := newGoogleProviderForTest("test-key", server.URL, server.Client())
	bias := trips.Point{Lat: 19.8067, Lng: 105.7852}
	results, err := provider.Search(t.Context(), "Vincom Thanh Hóa", &bias)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].PlaceID != "place-1" || results[0].Location.Lat != 19.8045 || results[0].Source != "google_places" {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestGoogleTextSearchRejectsShortQuery(t *testing.T) {
	provider := newGoogleProviderForTest("test-key", "http://unused", http.DefaultClient)
	if _, err := provider.Search(t.Context(), "ab", nil); err != ErrInvalidQuery {
		t.Fatalf("err=%v want ErrInvalidQuery", err)
	}
}
