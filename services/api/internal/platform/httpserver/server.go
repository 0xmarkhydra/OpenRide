package httpserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"flashx/services/api/internal/admin"
	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/customervehicles"
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/driverdocs"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/payments"
	"flashx/services/api/internal/platform/idempotency"
	"flashx/services/api/internal/pricing"
	"flashx/services/api/internal/ratings"
	"flashx/services/api/internal/realtime"
	"flashx/services/api/internal/ride"
	"flashx/services/api/internal/trips"
	"flashx/services/api/internal/users"
)

type Dependencies struct {
	AppEnv           string
	Persistence      string
	Trips            *trips.Service
	Drivers          *drivers.Service
	CustomerVehicles *customervehicles.Service
	DriverDocuments  *driverdocs.Service
	Dispatch         *dispatch.Engine
	Ride             *ride.Service
	Pricing          *pricing.Service
	Payments         *payments.Service
	Ratings          *ratings.Service
	Idempotency      idempotency.Store
	Auth             *auth.Service
	Users            *users.Service
	Admin            *admin.Service
	Realtime         *realtime.Hub
	AllowDevIdentity bool
	ReadyCheck       func(context.Context) error
}

type Server struct {
	server       *http.Server
	deps         Dependencies
	createTripMu sync.Mutex
}

type apiError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type dataEnvelope struct {
	Data any `json:"data"`
	Meta any `json:"meta,omitempty"`
}

func New(addr string, deps Dependencies) *Server {
	s := &Server{deps: deps}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /ready", s.ready)
	mux.HandleFunc("GET /v1", s.apiInfo)
	mux.HandleFunc("POST /v1/auth/otp/request", s.requestOTP)
	mux.HandleFunc("POST /v1/auth/otp/verify", s.verifyOTP)
	mux.HandleFunc("POST /v1/auth/refresh", s.refreshAuth)
	mux.HandleFunc("POST /v1/auth/logout", s.logoutAuth)
	mux.HandleFunc("GET /v1/realtime", s.realtimeSocket)
	mux.HandleFunc("GET /v1/rider/me", s.riderMe)
	mux.HandleFunc("PATCH /v1/rider/me", s.updateRiderMe)
	mux.HandleFunc("GET /v1/rider/vehicles", s.riderVehicles)
	mux.HandleFunc("POST /v1/rider/vehicles", s.createRiderVehicle)
	mux.HandleFunc("GET /v1/rider/vehicles/{id}", s.getRiderVehicle)
	mux.HandleFunc("POST /v1/trips/estimate", s.estimateTrip)
	mux.HandleFunc("POST /v1/trips", s.createTrip)
	mux.HandleFunc("GET /v1/trips", s.listTrips)
	mux.HandleFunc("GET /v1/trips/{id}", s.getTrip)
	mux.HandleFunc("GET /v1/trips/{id}/payment", s.getTripPayment)
	mux.HandleFunc("GET /v1/trips/{id}/rating", s.getTripRating)
	mux.HandleFunc("POST /v1/trips/{id}/rating", s.createTripRating)
	mux.HandleFunc("POST /v1/trips/{id}/cancel", s.cancelTrip)

	mux.HandleFunc("POST /v1/dev/drivers", s.registerDevDriver)
	mux.HandleFunc("GET /v1/driver/me", s.driverMe)
	mux.HandleFunc("PATCH /v1/driver/me", s.updateDriverMe)
	mux.HandleFunc("POST /v1/driver/documents/upload-url", s.prepareDriverDocumentUpload)
	mux.HandleFunc("POST /v1/driver/documents/complete", s.completeDriverDocumentUpload)
	mux.HandleFunc("GET /v1/driver/documents", s.listDriverDocuments)
	mux.HandleFunc("GET /v1/driver/documents/{id}/view-url", s.viewDriverDocument)
	mux.HandleFunc("POST /v1/driver/availability", s.driverAvailability)
	mux.HandleFunc("POST /v1/driver/location", s.driverLocation)
	mux.HandleFunc("GET /v1/driver/offers/current", s.currentDriverOffer)
	mux.HandleFunc("GET /v1/driver/trips", s.driverTrips)
	mux.HandleFunc("GET /v1/driver/trips/{id}", s.driverTripDetail)
	mux.HandleFunc("POST /v1/driver/offers/{id}/accept", s.acceptDriverOffer)
	mux.HandleFunc("POST /v1/driver/offers/{id}/reject", s.rejectDriverOffer)
	mux.HandleFunc("POST /v1/driver/trips/{id}/arriving", s.driverTripArriving)
	mux.HandleFunc("POST /v1/driver/trips/{id}/arrived", s.driverTripArrived)
	mux.HandleFunc("POST /v1/driver/trips/{id}/vehicle-received", s.driverTripVehicleReceived)
	mux.HandleFunc("POST /v1/driver/trips/{id}/start", s.driverTripStart)
	mux.HandleFunc("POST /v1/driver/trips/{id}/arrive-inspection", s.driverTripArriveInspection)
	mux.HandleFunc("POST /v1/driver/trips/{id}/start-inspection", s.driverTripStartInspection)
	mux.HandleFunc("POST /v1/driver/trips/{id}/complete-inspection", s.driverTripCompleteInspection)
	mux.HandleFunc("POST /v1/driver/trips/{id}/returning", s.driverTripReturning)
	mux.HandleFunc("POST /v1/driver/trips/{id}/arrived-return", s.driverTripArrivedReturn)
	mux.HandleFunc("POST /v1/driver/trips/{id}/handover", s.driverTripHandover)
	mux.HandleFunc("POST /v1/driver/trips/{id}/incident", s.driverTripIncident)
	mux.HandleFunc("POST /v1/driver/trips/{id}/complete", s.driverTripComplete)

	mux.HandleFunc("GET /v1/admin/me", s.adminMe)
	mux.HandleFunc("GET /v1/admin/drivers", s.adminDrivers)
	mux.HandleFunc("POST /v1/admin/drivers/{id}/approval", s.adminDriverApproval)
	mux.HandleFunc("GET /v1/admin/drivers/{id}/documents", s.adminDriverDocuments)
	mux.HandleFunc("POST /v1/admin/drivers/{id}/documents/{documentID}/review", s.adminReviewDriverDocument)
	mux.HandleFunc("GET /v1/admin/trips", s.adminTrips)
	mux.HandleFunc("GET /v1/admin/dashboard", s.adminDashboard)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           requestMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s
}

func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{"status": "ok", "service": "flashx-api"}})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if s.deps.ReadyCheck != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := s.deps.ReadyCheck(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "NOT_READY", "A required dependency is unavailable", nil)
			return
		}
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{"status": "ready", "persistence": s.deps.Persistence}})
}

func (s *Server) apiInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{"name": "FlashX API", "version": "v1"}})
}

type estimateTripRequest struct {
	Pickup            trips.Point `json:"pickup"`
	Destination       trips.Point `json:"destination"`
	ServiceType       string      `json:"service_type"`
	CustomerVehicleID string      `json:"customer_vehicle_id,omitempty"`
	BookingMode       string      `json:"booking_mode,omitempty"`
	ScheduledAt       *time.Time  `json:"scheduled_at,omitempty"`
}

func (s *Server) estimateTrip(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.riderID(w, r); !ok {
		return
	}
	var req estimateTripRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	estimate, err := s.deps.Pricing.Estimate(req.Pickup, req.Destination, req.ServiceType)
	if err != nil {
		switch {
		case errors.Is(err, pricing.ErrUnsupportedServiceType):
			writeError(w, http.StatusUnprocessableEntity, "SERVICE_TYPE_UNSUPPORTED", "Service type is not supported", nil)
		case errors.Is(err, pricing.ErrRouteUnavailable):
			writeError(w, http.StatusServiceUnavailable, "ROUTING_UNAVAILABLE", "Routing provider is temporarily unavailable", nil)
		default:
			writeError(w, http.StatusBadRequest, "ESTIMATE_INVALID", "Unable to estimate this route", nil)
		}
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: estimate})
}

func (s *Server) createTrip(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required", nil)
		return
	}

	var req estimateTripRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.ServiceType = trips.NormalizeServiceType(req.ServiceType)
	if req.BookingMode == "" {
		req.BookingMode = trips.BookingImmediate
	}
	if req.CustomerVehicleID != "" && s.deps.CustomerVehicles != nil {
		if _, err := s.deps.CustomerVehicles.ValidateForService(req.CustomerVehicleID, riderID, req.ServiceType); err != nil {
			s.writeVehicleError(w, err)
			return
		}
	}
	fingerprint := requestFingerprint(riderID, req)
	scope := "create-trip:" + riderID

	// This lock makes the in-memory development implementation atomic. Production
	// persistence replaces it with a durable idempotency record/DB transaction.
	s.createTripMu.Lock()
	defer s.createTripMu.Unlock()

	record, exists, err := s.deps.Idempotency.Get(scope, key)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "IDEMPOTENCY_UNAVAILABLE", "Idempotency service is unavailable", nil)
		return
	}
	if exists {
		if record.Fingerprint != fingerprint {
			writeError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key was already used for a different request", nil)
			return
		}
		trip, err := s.deps.Trips.GetForRider(record.ResourceID, riderID)
		if err != nil {
			s.writeDomainError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
		return
	}

	estimate, err := s.deps.Pricing.Estimate(req.Pickup, req.Destination, req.ServiceType)
	if err != nil {
		switch {
		case errors.Is(err, pricing.ErrUnsupportedServiceType):
			writeError(w, http.StatusUnprocessableEntity, "SERVICE_TYPE_UNSUPPORTED", "Service type is not supported", nil)
		case errors.Is(err, pricing.ErrRouteUnavailable):
			writeError(w, http.StatusServiceUnavailable, "ROUTING_UNAVAILABLE", "Routing provider is temporarily unavailable", nil)
		default:
			writeError(w, http.StatusBadRequest, "TRIP_INVALID", "Trip request is invalid", nil)
		}
		return
	}

	trip, err := s.deps.Trips.Create(trips.CreateInput{
		RiderID:            riderID,
		CustomerVehicleID:  req.CustomerVehicleID,
		ServiceType:        req.ServiceType,
		BookingMode:        req.BookingMode,
		ScheduledAt:        req.ScheduledAt,
		Pickup:             req.Pickup,
		Destination:        req.Destination,
		EstimatedDistanceM: estimate.DistanceM,
		EstimatedDurationS: estimate.DurationS,
		FareBreakdown:      estimate.Fare,
	})
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	if err := s.deps.Idempotency.Put(scope, key, idempotency.Record{Fingerprint: fingerprint, ResourceID: trip.ID}); err != nil {
		writeError(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", "Idempotency key conflict", nil)
		return
	}
	// Cash is the default MVP payment method. Payment persistence is intentionally
	// best-effort here because the durable trip has already been created; GET
	// /payment lazily reconciles a missing cash ledger record instead of creating
	// a duplicate trip on mobile retry.
	if s.deps.Payments != nil {
		_, _ = s.deps.Payments.EnsureCash(trip)
	}
	eventType := "trip." + string(trip.Status)
	s.publishActor(trip.RiderID, eventType, trip.ID, trip)
	if s.deps.Dispatch != nil && trip.Status == trips.StatusSearching {
		if offer, offerErr := s.deps.Dispatch.CreateOffer(trip.ID); offerErr == nil {
			s.publishOffers([]dispatch.Offer{offer})
		}
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: trip})
}

func (s *Server) getTrip(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	trip, err := s.deps.Trips.GetForRider(r.PathValue("id"), riderID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}

func (s *Server) listTrips(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.deps.Trips.ListForRider(riderID, limit)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

type cancelTripRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) cancelTrip(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	var req cancelTripRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	trip, err := s.deps.Ride.CancelByRider(r.PathValue("id"), riderID, req.Reason)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	if s.deps.Payments != nil {
		_, _ = s.deps.Payments.CancelCash(trip)
	}
	s.publishTrip(trip, "trip.cancelled", trip)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}

func (s *Server) riderID(w http.ResponseWriter, r *http.Request) (string, bool) {
	return s.actorID(w, r, auth.RoleRider, "X-Dev-Rider-ID")
}

func (s *Server) writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, trips.ErrNotFound):
		writeError(w, http.StatusNotFound, "TRIP_NOT_FOUND", "Trip was not found", nil)
	case errors.Is(err, trips.ErrForbidden):
		writeError(w, http.StatusForbidden, "TRIP_FORBIDDEN", "Trip is not accessible by this actor", nil)
	case errors.Is(err, trips.ErrInvalidState):
		writeError(w, http.StatusConflict, "TRIP_INVALID_STATE", "Trip cannot perform this action from its current state", nil)
	case errors.Is(err, trips.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "TRIP_INVALID", "Trip data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected server error", nil)
	}
}

func requestFingerprint(riderID string, req estimateTripRequest) string {
	payload, _ := json.Marshal(struct {
		RiderID string              `json:"rider_id"`
		Request estimateTripRequest `json:"request"`
	}{RiderID: riderID, Request: req})
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Request body is invalid", map[string]any{"reason": err.Error()})
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Request body must contain exactly one JSON object", nil)
		return false
	}
	return true
}

func requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	writeJSON(w, status, errorEnvelope{Error: apiError{Code: code, Message: message, Details: details}})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
