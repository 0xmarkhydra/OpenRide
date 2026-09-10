package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/engine"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

type V2Store interface {
	CreateTariff(context.Context, marketplace.DriverTariff) error
	ListDriverTariffs(context.Context, string) ([]marketplace.DriverTariff, error)
	CreateRequest(context.Context, marketplace.Request) error
	GetRequest(context.Context, string) (marketplace.Request, error)
	CancelRequest(context.Context, string, string) error
	ListEligibleRequests(context.Context, string, marketplace.ServiceType) ([]marketplace.Request, error)
	SubmitQuote(context.Context, marketplace.Quote) error
	ListOffers(context.Context, string) ([]marketplace.Quote, error)
	WithdrawQuote(context.Context, string, string) error
}

type driverTariffInput struct {
	InstanceID        string                  `json:"instance_id,omitempty"`
	ServiceType       marketplace.ServiceType `json:"service_type"`
	QuoteMode         marketplace.QuoteMode   `json:"quote_mode"`
	Currency          string                  `json:"currency"`
	BaseFareMinor     int64                   `json:"base_fare_minor,omitempty"`
	MinimumFareMinor  int64                   `json:"minimum_fare_minor,omitempty"`
	PerKMMinor        int64                   `json:"per_km_minor,omitempty"`
	PerMinuteMinor    int64                   `json:"per_minute_minor,omitempty"`
	PickupFeeMinor    int64                   `json:"pickup_fee_minor,omitempty"`
	AutoQuoteMinMinor int64                   `json:"auto_quote_min_minor,omitempty"`
	AutoQuoteMaxMinor int64                   `json:"auto_quote_max_minor,omitempty"`
}

type quoteInput struct {
	FareTotalMinor int64  `json:"fare_total_minor"`
	Currency       string `json:"currency"`
	ExpiresAt      string `json:"expires_at,omitempty"`
}

type offerOutput struct {
	QuoteID         string   `json:"quote_id"`
	FareTotalMinor  int64    `json:"fare_total_minor"`
	Currency        string   `json:"currency"`
	PickupETAS      int64    `json:"pickup_eta_s"`
	PickupDistanceM int64    `json:"pickup_distance_m"`
	ExpiresAt       string   `json:"expires_at"`
	Rank            int      `json:"rank,omitempty"`
	Reasons         []string `json:"reasons,omitempty"`
	Recommended     bool     `json:"recommended,omitempty"`
}

type agreementOutput struct {
	ID             string                     `json:"id"`
	RequestID      string                     `json:"request_id"`
	QuoteID        string                     `json:"quote_id"`
	DriverID       string                     `json:"driver_id"`
	RiderID        string                     `json:"rider_id"`
	ServiceType    marketplace.ServiceType    `json:"service_type"`
	FareTotalMinor int64                      `json:"fare_total_minor"`
	Currency       string                     `json:"currency"`
	TermsSnapshot  marketplace.AgreementTerms `json:"terms_snapshot"`
	CreatedAt      string                     `json:"created_at"`
}

func amount(currency string, minor int64) (money.Amount, error) {
	return money.New(currency, minor)
}

func toAgreementOutput(a marketplace.Agreement) agreementOutput {
	return agreementOutput{
		ID:             a.ID,
		RequestID:      a.RequestID,
		QuoteID:        a.QuoteID,
		DriverID:       a.DriverID,
		RiderID:        a.RiderID,
		ServiceType:    a.ServiceType,
		FareTotalMinor: a.Fare.Minor,
		Currency:       a.Fare.Currency,
		TermsSnapshot:  a.TermsSnapshot,
		CreatedAt:      a.CreatedAt.UTC().Format(time.RFC3339Nano),
	}
}

func (s *Server) registerV2(mux *http.ServeMux) {
	mux.HandleFunc("GET /v2/services", s.listServices)
	if s.v2 == nil {
		return
	}
	mux.HandleFunc("POST /v2/requests", s.createRequest)
	mux.HandleFunc("GET /v2/requests/{requestID}", s.getRequest)
	mux.HandleFunc("POST /v2/requests/{requestID}/cancel", s.cancelRequest)
	mux.HandleFunc("GET /v2/requests/{requestID}/offers", s.listOffers)
	mux.HandleFunc("POST /v2/requests/{requestID}/quotes", s.submitQuote)
	mux.HandleFunc("POST /v2/quotes/{quoteID}/accept", s.acceptQuote)
	mux.HandleFunc("POST /v2/quotes/{quoteID}/withdraw", s.withdrawQuote)
	mux.HandleFunc("POST /v2/drivers/me/tariffs", s.createDriverTariff)
	mux.HandleFunc("GET /v2/drivers/me/tariffs", s.listDriverTariffs)
	mux.HandleFunc("GET /v2/drivers/me/requests", s.listEligibleRequests)
}

func actorID(r *http.Request) (string, bool) {
	id := strings.TrimSpace(r.Header.Get("X-OpenRide-Actor-ID"))
	return id, id != ""
}

func requireActor(w http.ResponseWriter, r *http.Request) (string, bool) {
	id, ok := actorID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "ACTOR_REQUIRED", "trusted gateway must provide X-OpenRide-Actor-ID")
		return "", false
	}
	return id, true
}

func idempotencyKey(w http.ResponseWriter, r *http.Request) (string, bool) {
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeError(w, http.StatusBadRequest, "IDEMPOTENCY_KEY_REQUIRED", "Idempotency-Key header is required")
		return "", false
	}
	return key, true
}

func newID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(raw[:]), nil
}

func (s *Server) createRequest(w http.ResponseWriter, r *http.Request) {
	rider, ok := requireActor(w, r)
	if !ok {
		return
	}
	if _, ok = idempotencyKey(w, r); !ok {
		return
	}
	var req marketplace.Request
	if !decodeJSON(w, r, &req) {
		return
	}
	id, err := newID("req")
	if err != nil {
		writeError(w, 500, "ID_GENERATION_FAILED", "could not create request id")
		return
	}
	now := time.Now().UTC()
	req.ID = id
	req.RiderID = rider
	req.Status = marketplace.RequestOpen
	req.RequestedAt = now
	req.Version = 1
	if req.InstanceID == "" {
		req.InstanceID = "default"
	}
	if err := req.Validate(); err != nil {
		writeError(w, 422, "REQUEST_INVALID", err.Error())
		return
	}
	module, err := s.services.Get(req.ServiceType)
	if err != nil {
		writeError(w, 422, "SERVICE_TYPE_UNSUPPORTED", err.Error())
		return
	}
	if err := module.ValidateRequest(r.Context(), req); err != nil {
		writeError(w, 422, "SERVICE_REQUEST_INVALID", err.Error())
		return
	}
	if err := s.v2.CreateRequest(r.Context(), req); err != nil {
		writeError(w, 409, "REQUEST_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, envelope{Data: req})
}

func (s *Server) getRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	req, err := s.v2.GetRequest(r.Context(), r.PathValue("requestID"))
	if err != nil {
		writeError(w, 404, "REQUEST_NOT_FOUND", "request not found")
		return
	}
	if req.RiderID != actor {
		writeError(w, 403, "REQUEST_FORBIDDEN", "request belongs to another rider")
		return
	}
	writeJSON(w, 200, envelope{Data: req})
}

func (s *Server) cancelRequest(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	if _, ok = idempotencyKey(w, r); !ok {
		return
	}
	if err := s.v2.CancelRequest(r.Context(), r.PathValue("requestID"), actor); err != nil {
		writeError(w, 409, "REQUEST_CANCEL_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]any{"status": "cancelled"}})
}

func (s *Server) listOffers(w http.ResponseWriter, r *http.Request) {
	actor, ok := requireActor(w, r)
	if !ok {
		return
	}
	req, err := s.v2.GetRequest(r.Context(), r.PathValue("requestID"))
	if err != nil || req.RiderID != actor {
		writeError(w, 404, "REQUEST_NOT_FOUND", "request not found")
		return
	}
	quotes, err := s.v2.ListOffers(r.Context(), req.ID)
	if err != nil {
		writeError(w, 500, "OFFERS_READ_FAILED", err.Error())
		return
	}
	ranked, err := engine.RankQuotes(r.Context(), req, quotes, s.ranker)
	if err != nil {
		writeError(w, 500, "OFFERS_RANK_FAILED", err.Error())
		return
	}
	out := make([]offerOutput, 0, len(ranked))
	for i, item := range ranked {
		q := item.Quote
		out = append(out, offerOutput{
			QuoteID:         q.ID,
			FareTotalMinor:  q.Fare.Minor,
			Currency:        q.Fare.Currency,
			PickupETAS:      q.PickupETAS,
			PickupDistanceM: q.PickupDistanceM,
			ExpiresAt:       q.ExpiresAt.UTC().Format(time.RFC3339Nano),
			Rank:            i + 1,
			Reasons:         item.Reasons,
			Recommended:     item.Recommended,
		})
	}
	writeJSON(w, 200, envelope{Data: out})
}

func (s *Server) createDriverTariff(w http.ResponseWriter, r *http.Request) {
	driver, ok := requireActor(w, r)
	if !ok {
		return
	}
	if _, ok = idempotencyKey(w, r); !ok {
		return
	}
	var input driverTariffInput
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := newID("tariff")
	if err != nil {
		writeError(w, 500, "ID_GENERATION_FAILED", "could not create tariff id")
		return
	}
	instanceID := input.InstanceID
	if instanceID == "" {
		instanceID = "default"
	}
	baseFare, err := amount(input.Currency, input.BaseFareMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	minimumFare, err := amount(input.Currency, input.MinimumFareMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	perKM, err := amount(input.Currency, input.PerKMMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	perMinute, err := amount(input.Currency, input.PerMinuteMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	pickupFee, err := amount(input.Currency, input.PickupFeeMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	autoMin, err := amount(input.Currency, input.AutoQuoteMinMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	autoMax, err := amount(input.Currency, input.AutoQuoteMaxMinor)
	if err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	t := marketplace.DriverTariff{
		ID:               id,
		InstanceID:       instanceID,
		DriverID:         driver,
		ServiceType:      input.ServiceType,
		QuoteMode:        input.QuoteMode,
		BaseFare:         baseFare,
		MinimumFare:      minimumFare,
		PerKM:            perKM,
		PerMinute:        perMinute,
		PickupFee:        pickupFee,
		AutoQuoteMinimum: autoMin,
		AutoQuoteMaximum: autoMax,
		Version:          1,
	}
	if err := t.Validate(); err != nil {
		writeError(w, 422, "TARIFF_INVALID", err.Error())
		return
	}
	if err := s.v2.CreateTariff(r.Context(), t); err != nil {
		writeError(w, 409, "TARIFF_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, envelope{Data: t})
}

func (s *Server) listDriverTariffs(w http.ResponseWriter, r *http.Request) {
	driver, ok := requireActor(w, r)
	if !ok {
		return
	}
	items, err := s.v2.ListDriverTariffs(r.Context(), driver)
	if err != nil {
		writeError(w, 500, "TARIFF_READ_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, envelope{Data: items})
}

func (s *Server) listEligibleRequests(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireActor(w, r); !ok {
		return
	}
	service := marketplace.ServiceType(strings.TrimSpace(r.URL.Query().Get("service_type")))
	if service == "" {
		service = "passenger.car"
	}
	items, err := s.v2.ListEligibleRequests(r.Context(), "default", service)
	if err != nil {
		writeError(w, 500, "REQUEST_READ_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, envelope{Data: items})
}

func (s *Server) submitQuote(w http.ResponseWriter, r *http.Request) {
	driver, ok := requireActor(w, r)
	if !ok {
		return
	}
	if _, ok = idempotencyKey(w, r); !ok {
		return
	}
	requestID := r.PathValue("requestID")
	req, err := s.v2.GetRequest(r.Context(), requestID)
	if err != nil {
		writeError(w, 404, "REQUEST_NOT_FOUND", "request not found")
		return
	}
	if req.Status != marketplace.RequestOpen && req.Status != marketplace.RequestReceivingQuotes {
		writeError(w, 409, "REQUEST_NOT_OPEN", "request is not accepting quotes")
		return
	}
	var input quoteInput
	if !decodeJSON(w, r, &input) {
		return
	}
	id, err := newID("quote")
	if err != nil {
		writeError(w, 500, "ID_GENERATION_FAILED", "could not create quote id")
		return
	}
	now := time.Now().UTC()
	expiresAt := now.Add(2 * time.Minute)
	if input.ExpiresAt != "" {
		parsed, parseErr := time.Parse(time.RFC3339Nano, input.ExpiresAt)
		if parseErr != nil {
			writeError(w, 422, "QUOTE_INVALID", "expires_at must be RFC3339")
			return
		}
		expiresAt = parsed.UTC()
	}
	fare, err := amount(input.Currency, input.FareTotalMinor)
	if err != nil {
		writeError(w, 422, "QUOTE_INVALID", err.Error())
		return
	}
	q := marketplace.Quote{
		ID:        id,
		RequestID: requestID,
		DriverID:  driver,
		Status:    marketplace.QuotePending,
		Fare:      fare,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}
	if err := q.Validate(); err != nil {
		writeError(w, 422, "QUOTE_INVALID", err.Error())
		return
	}
	if err := s.v2.SubmitQuote(r.Context(), q); err != nil {
		writeError(w, 409, "QUOTE_CREATE_FAILED", err.Error())
		return
	}
	writeJSON(w, 201, envelope{Data: offerOutput{
		QuoteID:         q.ID,
		FareTotalMinor:  q.Fare.Minor,
		Currency:        q.Fare.Currency,
		PickupETAS:      q.PickupETAS,
		PickupDistanceM: q.PickupDistanceM,
		ExpiresAt:       q.ExpiresAt.UTC().Format(time.RFC3339Nano),
		Reasons:         q.Explanation,
	}})
}

func (s *Server) withdrawQuote(w http.ResponseWriter, r *http.Request) {
	driver, ok := requireActor(w, r)
	if !ok {
		return
	}
	if _, ok = idempotencyKey(w, r); !ok {
		return
	}
	if err := s.v2.WithdrawQuote(r.Context(), r.PathValue("quoteID"), driver); err != nil {
		writeError(w, 409, "QUOTE_WITHDRAW_FAILED", err.Error())
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]any{"status": "withdrawn"}})
}

func (s *Server) acceptQuote(w http.ResponseWriter, r *http.Request) {
	rider, ok := requireActor(w, r)
	if !ok {
		return
	}
	key, ok := idempotencyKey(w, r)
	if !ok {
		return
	}
	agreementID, err := newID("agr")
	if err != nil {
		writeError(w, 500, "ID_GENERATION_FAILED", "could not create agreement id")
		return
	}
	eventID, err := newID("evt")
	if err != nil {
		writeError(w, 500, "ID_GENERATION_FAILED", "could not create event id")
		return
	}
	agreement, err := s.accept.AcceptQuote(r.Context(), app.AcceptQuoteCommand{
		QuoteID:        r.PathValue("quoteID"),
		RiderID:        rider,
		IdempotencyKey: key,
		AgreementID:    agreementID,
		EventID:        eventID,
		Now:            time.Now().UTC(),
	})
	if err != nil {
		status := http.StatusConflict
		code := "QUOTE_ACCEPT_FAILED"
		if errors.Is(err, app.ErrRiderMismatch) {
			status = http.StatusForbidden
			code = "RIDER_MISMATCH"
		}
		writeError(w, status, code, err.Error())
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]any{"agreement": toAgreementOutput(agreement)}})
}
