package httpserver

import (
	"net/http"
	"strings"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
)

// NewV2 wraps the existing V1 server with additive Marketplace V2 routes.
// V1 remains untouched while V2 gradually takes over marketplace behavior.
func NewV2(addr string, deps Dependencies, manifests []extension.Manifest) *Server {
	s := New(addr, deps)
	legacy := s.server.Handler

	v2Mux := http.NewServeMux()
	v2Mux.HandleFunc("GET /v2", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{
			"name":    "OpenRide API",
			"version": "v2",
			"stage":   "marketplace-foundation",
		}})
	})
	v2Mux.HandleFunc("GET /v2/services", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, dataEnvelope{Data: manifests})
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
