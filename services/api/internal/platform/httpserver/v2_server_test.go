package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
)

func TestV2ServiceCatalog(t *testing.T) {
	server := NewV2(":0", Dependencies{}, []extension.Manifest{
		{ID: "passenger.car", Version: "1.0.0", DisplayName: "Passenger Car", Category: "passenger"},
		{ID: "carpool.intercity", Version: "1.0.0", DisplayName: "Intercity Carpool", Category: "carpool"},
	})

	req := httptest.NewRequest(http.MethodGet, "/v2/services", nil)
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var envelope struct {
		Data []extension.Manifest `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(envelope.Data) != 2 || envelope.Data[0].ID != "passenger.car" {
		t.Fatalf("unexpected catalog: %+v", envelope.Data)
	}
}

func TestV2WrapperKeepsV1Health(t *testing.T) {
	server := NewV2(":0", Dependencies{Persistence: "memory"}, nil)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("legacy health status = %d", rec.Code)
	}
}
