package httpserver

import (
	"io"
	"net/http"
	"strings"
	"time"
)

// NewV2 wraps the existing V1 server with additive V2 gateway routes.
// Marketplace business logic is owned by marketplace-service; this compatibility
// API only proxies/composes public V2 traffic during the extraction period.
func NewV2(addr string, deps Dependencies, marketplaceServiceURL string) *Server {
	s := New(addr, deps)
	legacy := s.server.Handler
	marketplaceServiceURL = strings.TrimRight(strings.TrimSpace(marketplaceServiceURL), "/")
	client := &http.Client{Timeout: 3 * time.Second}

	v2Mux := http.NewServeMux()
	v2Mux.HandleFunc("GET /v2", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{
			"name":    "OpenRide Edge Compatibility API",
			"version": "v2",
			"stage":   "microservices-extraction",
		}})
	})
	v2Mux.HandleFunc("GET /v2/services", func(w http.ResponseWriter, r *http.Request) {
		if marketplaceServiceURL == "" {
			writeError(w, http.StatusServiceUnavailable, "MARKETPLACE_UNAVAILABLE", "Marketplace service is not configured", nil)
			return
		}
		upstream, err := http.NewRequestWithContext(r.Context(), http.MethodGet, marketplaceServiceURL+"/internal/v1/services", nil)
		if err != nil {
			writeError(w, http.StatusBadGateway, "MARKETPLACE_UPSTREAM_INVALID", "Marketplace upstream request could not be created", nil)
			return
		}
		if correlationID := r.Header.Get("X-Correlation-ID"); correlationID != "" {
			upstream.Header.Set("X-Correlation-ID", correlationID)
		}
		response, err := client.Do(upstream)
		if err != nil {
			writeError(w, http.StatusBadGateway, "MARKETPLACE_UNAVAILABLE", "Marketplace service is unavailable", nil)
			return
		}
		defer response.Body.Close()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, io.LimitReader(response.Body, 2<<20))
	})
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
