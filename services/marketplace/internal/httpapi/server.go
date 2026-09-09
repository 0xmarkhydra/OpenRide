package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

type Server struct {
	http     *http.Server
	services *extension.Registry
	ready    func(context.Context) error
}

type Config struct {
	Addr     string
	Services *extension.Registry
	Ready    func(context.Context) error
}

type envelope struct {
	Data any `json:"data"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

func New(cfg Config) (*Server, error) {
	if cfg.Services == nil {
		return nil, errors.New("marketplace http: service registry is required")
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8090"
	}

	s := &Server{services: cfg.Services, ready: cfg.Ready}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /readyz", s.readiness)
	mux.HandleFunc("GET /internal/v1/services", s.listServices)
	mux.HandleFunc("POST /internal/v1/requests/validate", s.validateRequest)

	s.http = &http.Server{
		Addr:              cfg.Addr,
		Handler:           middleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s, nil
}

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }
func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }
func (s *Server) Handler() http.Handler { return s.http.Handler }

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{Data: map[string]any{
		"status":  "ok",
		"service": "marketplace-service",
	}})
}

func (s *Server) readiness(w http.ResponseWriter, r *http.Request) {
	if s.ready != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := s.ready(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "NOT_READY", "marketplace dependency unavailable")
			return
		}
	}
	writeJSON(w, http.StatusOK, envelope{Data: map[string]any{"status": "ready"}})
}

func (s *Server) listServices(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, envelope{Data: s.services.Manifests()})
}

func (s *Server) validateRequest(w http.ResponseWriter, r *http.Request) {
	var request marketplace.Request
	if !decodeJSON(w, r, &request) {
		return
	}
	if err := request.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "REQUEST_INVALID", err.Error())
		return
	}
	module, err := s.services.Get(request.ServiceType)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "SERVICE_TYPE_UNSUPPORTED", err.Error())
		return
	}
	if err := module.ValidateRequest(r.Context(), request); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "SERVICE_REQUEST_INVALID", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, envelope{Data: map[string]any{
		"valid":        true,
		"service_type": request.ServiceType,
		"module":       module.Manifest(),
	}})
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", err.Error())
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "request body must contain exactly one JSON value")
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{Error: apiError{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
