package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
)

func TestMarketplaceV2RejectsDirectRequestWithoutGatewayToken(t *testing.T) {
	server, err := New(Config{
		Services:     modulecatalog.DefaultRegistry(),
		GatewayToken: "test-marketplace-gateway-token-123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v2/services", nil)
	req.Header.Set("X-OpenRide-Actor-ID", "spoofed-actor")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMarketplaceV2AcceptsTrustedGatewayToken(t *testing.T) {
	server, err := New(Config{
		Services:     modulecatalog.DefaultRegistry(),
		GatewayToken: "test-marketplace-gateway-token-123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v2/services", nil)
	req.Header.Set(gatewayTokenHeader, "test-marketplace-gateway-token-123456")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestMarketplaceHealthDoesNotRequireGatewayToken(t *testing.T) {
	server, err := New(Config{
		Services:     modulecatalog.DefaultRegistry(),
		GatewayToken: "test-marketplace-gateway-token-123456",
	})
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("health status=%d body=%s", rec.Code, rec.Body.String())
	}
}
