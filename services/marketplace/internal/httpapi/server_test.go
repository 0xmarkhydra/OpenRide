package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
)

func TestCatalogAndValidation(t *testing.T) {
	server, err := New(Config{Services: modulecatalog.DefaultRegistry()})
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/internal/v1/services", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("catalog status=%d body=%s", rec.Code, rec.Body.String())
	}

	destination := geo.Point{Lat: 19.77, Lng: 105.77}
	request := marketplace.Request{
		ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car",
		Status: marketplace.RequestOpen,
		Pickup: geo.Point{Lat: 19.80, Lng: 105.78}, Destination: &destination,
		RequestedAt: time.Now(), Version: 1,
	}
	payload, _ := json.Marshal(request)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/internal/v1/requests/validate", bytes.NewReader(payload)))
	if rec.Code != http.StatusOK {
		t.Fatalf("validate status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestUnsupportedServiceIsRejected(t *testing.T) {
	server, _ := New(Config{Services: modulecatalog.DefaultRegistry()})
	request := map[string]any{
		"id": "req_1", "instance_id": "default", "rider_id": "rider_1",
		"service_type": "unknown.service", "status": "open",
		"pickup": map[string]any{"lat": 19.8, "lng": 105.7},
		"requested_at": time.Now(), "version": 1,
	}
	payload, _ := json.Marshal(request)
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/internal/v1/requests/validate", bytes.NewReader(payload)))
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
