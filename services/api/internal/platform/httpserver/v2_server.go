package httpserver

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"flashx/services/api/internal/auth"
)

// NewV2 wraps the existing V1 server with additive V2 gateway routes.
// Marketplace business logic is owned by marketplace-service; this compatibility
// API authenticates public traffic and reconstructs trusted internal actor
// context before forwarding requests during the extraction period.
func NewV2(addr string, deps Dependencies, marketplaceServiceURL string) *Server {
	s := New(addr, deps)
	legacy := s.server.Handler
	marketplaceServiceURL = strings.TrimRight(strings.TrimSpace(marketplaceServiceURL), "/")
	client := &http.Client{Timeout: 5 * time.Second}

	v2Mux := http.NewServeMux()
	v2Mux.HandleFunc("GET /v2", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{
			"name":    "OpenRide Edge Compatibility API",
			"version": "v2",
			"stage":   "microservices-extraction",
		}})
	})
	v2Mux.HandleFunc("GET /v2/services", func(w http.ResponseWriter, r *http.Request) {
		s.proxyMarketplace(w, r, client, marketplaceServiceURL, "")
	})

	// Rider Marketplace commands and reads.
	v2Mux.HandleFunc("POST /v2/requests", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleRider))
	v2Mux.HandleFunc("GET /v2/requests/{requestID}", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleRider))
	v2Mux.HandleFunc("POST /v2/requests/{requestID}/cancel", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleRider))
	v2Mux.HandleFunc("GET /v2/requests/{requestID}/offers", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleRider))
	v2Mux.HandleFunc("POST /v2/quotes/{quoteID}/accept", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleRider))

	// Driver Marketplace commands and reads.
	v2Mux.HandleFunc("GET /v2/drivers/me/tariffs", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleDriver))
	v2Mux.HandleFunc("POST /v2/drivers/me/tariffs", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleDriver))
	v2Mux.HandleFunc("GET /v2/drivers/me/requests", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleDriver))
	v2Mux.HandleFunc("POST /v2/requests/{requestID}/quotes", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleDriver))
	v2Mux.HandleFunc("POST /v2/quotes/{quoteID}/withdraw", s.marketplaceActorProxy(client, marketplaceServiceURL, auth.RoleDriver))

	v2Handler := requestMiddleware(v2Mux)
	s.server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2" || strings.HasPrefix(r.URL.Path, "/v2/") {
			v2Handler.ServeHTTP(w, r)
			return
		}
		legacy.ServeHTTP(w, r)
	})
	return s
}

func (s *Server) marketplaceActorProxy(client *http.Client, marketplaceServiceURL string, role auth.Role) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		actorID, ok := s.actorID(w, r, role, marketplaceDevIdentityHeader(role))
		if !ok {
			return
		}
		s.proxyMarketplace(w, r, client, marketplaceServiceURL, actorID)
	}
}

func marketplaceDevIdentityHeader(role auth.Role) string {
	if role == auth.RoleDriver {
		return "X-Dev-Driver-ID"
	}
	return "X-Dev-Rider-ID"
}

func (s *Server) proxyMarketplace(w http.ResponseWriter, r *http.Request, client *http.Client, marketplaceServiceURL, actorID string) {
	if marketplaceServiceURL == "" {
		writeError(w, http.StatusServiceUnavailable, "MARKETPLACE_UNAVAILABLE", "Marketplace service is not configured", nil)
		return
	}
	base, err := url.Parse(marketplaceServiceURL)
	if err != nil || base.Scheme == "" || base.Host == "" {
		writeError(w, http.StatusBadGateway, "MARKETPLACE_UPSTREAM_INVALID", "Marketplace upstream URL is invalid", nil)
		return
	}

	upstreamURL := *base
	upstreamURL.Path = strings.TrimRight(base.Path, "/") + r.URL.Path
	upstreamURL.RawQuery = r.URL.RawQuery
	upstream, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), r.Body)
	if err != nil {
		writeError(w, http.StatusBadGateway, "MARKETPLACE_UPSTREAM_INVALID", "Marketplace upstream request could not be created", nil)
		return
	}

	// Deliberately allow-list public headers. Authorization and all client-supplied
	// X-OpenRide-* headers terminate at the edge and never cross the trust boundary.
	for _, header := range []string{"Accept", "Content-Type", "Idempotency-Key", "X-Correlation-ID"} {
		if value := strings.TrimSpace(r.Header.Get(header)); value != "" {
			upstream.Header.Set(header, value)
		}
	}
	if actorID != "" {
		upstream.Header.Set("X-OpenRide-Actor-ID", actorID)
	}

	response, err := client.Do(upstream)
	if err != nil {
		writeError(w, http.StatusBadGateway, "MARKETPLACE_UNAVAILABLE", "Marketplace service is unavailable", nil)
		return
	}
	defer response.Body.Close()

	if contentType := response.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	}
	if correlationID := response.Header.Get("X-Correlation-ID"); correlationID != "" {
		w.Header().Set("X-Correlation-ID", correlationID)
	}
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, io.LimitReader(response.Body, 2<<20))
}
