package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"flashx/services/api/internal/auth"
)

func testAuthService(t *testing.T) *auth.Service {
	t.Helper()
	svc, err := auth.NewService(auth.NewMemoryStore(), auth.DevelopmentOTPSender{}, "test-secret-at-least-16-bytes", "test", "development")
	if err != nil {
		t.Fatalf("new auth service: %v", err)
	}
	return svc
}

func bearerFor(t *testing.T, svc *auth.Service, actorID string, role auth.Role) string {
	t.Helper()
	pair, err := svc.IssueTokens(actorID, role)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return "Bearer " + pair.AccessToken
}

func TestV2ServiceCatalogIsProxiedToMarketplaceService(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/services" {
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}
		if r.Header.Get("X-OpenRide-Actor-ID") != "" {
			t.Fatalf("public catalog must not synthesize an actor")
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

func TestV2MarketplaceRouteRequiresAuthentication(t *testing.T) {
	var upstreamCalls atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	server := NewV2(":0", Dependencies{Auth: testAuthService(t)}, upstream.URL)
	req := httptest.NewRequest(http.MethodGet, "/v2/requests/req_1", nil)
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if upstreamCalls.Load() != 0 {
		t.Fatalf("unauthenticated request reached Marketplace")
	}
}

func TestV2MarketplaceProxyRebuildsActorAndStripsClientTrustHeaders(t *testing.T) {
	authService := testAuthService(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/requests" {
			t.Fatalf("unexpected upstream path: %s", r.URL.Path)
		}
		if got := r.Header.Get("X-OpenRide-Actor-ID"); got != "rider_real" {
			t.Fatalf("trusted actor = %q, want rider_real", got)
		}
		if got := r.Header.Get("X-OpenRide-Role"); got != "" {
			t.Fatalf("client X-OpenRide-Role leaked upstream: %q", got)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("public bearer token leaked upstream")
		}
		if got := r.Header.Get("Idempotency-Key"); got != "create-request-1" {
			t.Fatalf("idempotency key = %q", got)
		}
		if got := r.Header.Get("X-Correlation-ID"); got != "corr-1" {
			t.Fatalf("correlation id = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"req_1"}}`))
	}))
	defer upstream.Close()

	server := NewV2(":0", Dependencies{Auth: authService}, upstream.URL)
	req := httptest.NewRequest(http.MethodPost, "/v2/requests", strings.NewReader(`{"service_type":"passenger.car"}`))
	req.Header.Set("Authorization", bearerFor(t, authService, "rider_real", auth.RoleRider))
	req.Header.Set("Idempotency-Key", "create-request-1")
	req.Header.Set("X-Correlation-ID", "corr-1")
	req.Header.Set("X-OpenRide-Actor-ID", "attacker")
	req.Header.Set("X-OpenRide-Role", "admin")
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestV2MarketplaceRouteEnforcesActorRole(t *testing.T) {
	var upstreamCalls atomic.Int32
	authService := testAuthService(t)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		upstreamCalls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	server := NewV2(":0", Dependencies{Auth: authService}, upstream.URL)
	req := httptest.NewRequest(http.MethodPost, "/v2/requests", strings.NewReader(`{}`))
	req.Header.Set("Authorization", bearerFor(t, authService, "driver_1", auth.RoleDriver))
	rec := httptest.NewRecorder()
	server.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if upstreamCalls.Load() != 0 {
		t.Fatalf("wrong-role request reached Marketplace")
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
