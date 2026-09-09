package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestV2ServiceCatalogIsProxiedToMarketplaceService(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/services" {
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"passenger.car","version":"1.0.0","display_name":"Passenger Car","category":"passenger"}]}`))
	}))
	defer upstream.Close()

	server := NewV2(":0", Dependencies{}, upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/v2/services", nil)
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "passenger.car") {
		t.Fatalf("unexpected catalog body: %s", rec.Body.String())
	}
}

func TestV2MarketplaceRouteFailsClosedWhenServiceIsNotConfigured(t *testing.T) {
	server := NewV2(":0", Dependencies{}, "")
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v2/services", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestV2WrapperKeepsV1Health(t *testing.T) {
	server := NewV2(":0", Dependencies{Persistence: "memory"}, "")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("legacy health status = %d", rec.Code)
	}
}
