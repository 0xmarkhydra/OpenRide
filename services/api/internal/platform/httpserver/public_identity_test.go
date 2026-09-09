package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompatibilityAPIUsesOpenRidePublicIdentity(t *testing.T) {
	s := New(":0", Dependencies{Persistence: "memory"})

	health := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "/health", nil))
	if health.Code != http.StatusOK {
		t.Fatalf("health status = %d", health.Code)
	}
	var healthEnvelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(health.Body.Bytes(), &healthEnvelope); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if healthEnvelope.Data["service"] != "openride-compatibility-api" {
		t.Fatalf("unexpected public service identity: %v", healthEnvelope.Data["service"])
	}

	info := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(info, httptest.NewRequest(http.MethodGet, "/v1", nil))
	if info.Code != http.StatusOK {
		t.Fatalf("api info status = %d", info.Code)
	}
	var infoEnvelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(info.Body.Bytes(), &infoEnvelope); err != nil {
		t.Fatalf("decode api info: %v", err)
	}
	if infoEnvelope.Data["name"] != "OpenRide Compatibility API" {
		t.Fatalf("unexpected API name: %v", infoEnvelope.Data["name"])
	}
	if infoEnvelope.Data["lifecycle"] != "compatibility" {
		t.Fatalf("expected compatibility lifecycle, got %v", infoEnvelope.Data["lifecycle"])
	}
}
