package app

import (
	"context"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

type memoryAcceptanceStore struct {
	request      marketplace.Request
	quote        marketplace.Quote
	agreement    marketplace.Agreement
	idempotency  map[string]string
	outbox       []OutboxEvent
	acceptCount  int
	requestCount int
}

func (s *memoryAcceptanceStore) WithinTx(_ context.Context, fn func(AcceptanceTx) error) error {
	return fn(s)
}

func (s *memoryAcceptanceStore) FindIdempotentAgreement(_ context.Context, riderID, key string) (marketplace.Agreement, bool, error) {
	if agreementID := s.idempotency[riderID+":"+key]; agreementID != "" && agreementID == s.agreement.ID {
		return s.agreement, true, nil
	}
	return marketplace.Agreement{}, false, nil
}

func (s *memoryAcceptanceStore) GetQuoteForUpdate(context.Context, string) (marketplace.Quote, error) {
	return s.quote, nil
}

func (s *memoryAcceptanceStore) GetRequestForUpdate(context.Context, string) (marketplace.Request, error) {
	return s.request, nil
}

func (s *memoryAcceptanceStore) InsertAgreement(_ context.Context, agreement marketplace.Agreement) error {
	s.agreement = agreement
	return nil
}

func (s *memoryAcceptanceStore) MarkQuoteAccepted(_ context.Context, _ string, at time.Time) error {
	s.acceptCount++
	s.quote.Status = marketplace.QuoteAccepted
	s.quote.AcceptedAt = &at
	return nil
}

func (s *memoryAcceptanceStore) MarkRequestAgreed(context.Context, string) error {
	s.requestCount++
	s.request.Status = marketplace.RequestAgreed
	return nil
}

func (s *memoryAcceptanceStore) InvalidateOtherQuotes(context.Context, string, string) error { return nil }

func (s *memoryAcceptanceStore) SaveIdempotentAgreement(_ context.Context, riderID, key, agreementID string) error {
	s.idempotency[riderID+":"+key] = agreementID
	return nil
}

func (s *memoryAcceptanceStore) AppendOutbox(_ context.Context, event OutboxEvent) error {
	s.outbox = append(s.outbox, event)
	return nil
}

func TestAcceptQuoteCommitsAgreementAndOutboxOnce(t *testing.T) {
	now := time.Date(2026, 9, 9, 13, 0, 0, 0, time.UTC)
	destination := geo.Point{Lat: 21.0285, Lng: 105.8542}
	store := &memoryAcceptanceStore{
		request: marketplace.Request{
			ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car",
			Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 21.0278, Lng: 105.8342}, Destination: &destination,
			RequestedAt: now.Add(-time.Minute), Version: 1,
		},
		quote: marketplace.Quote{
			ID: "quote_1", RequestID: "req_1", DriverID: "driver_1", TariffID: "tariff_1", TariffVersion: 2,
			Status: marketplace.QuotePending, Fare: money.Must("VND", 52_000), PickupDistanceM: 900, PickupETAS: 180,
			Metadata: map[string]any{"source": "driver_tariff"}, CreatedAt: now.Add(-30 * time.Second), ExpiresAt: now.Add(time.Minute),
		},
		idempotency: map[string]string{},
	}
	service := AcceptanceService{Store: store}
	cmd := AcceptQuoteCommand{
		QuoteID: "quote_1", RiderID: "rider_1", IdempotencyKey: "idem_1",
		AgreementID: "agreement_1", EventID: "event_1", Now: now,
	}

	first, err := service.AcceptQuote(context.Background(), cmd)
	if err != nil {
		t.Fatalf("first acceptance failed: %v", err)
	}
	second, err := service.AcceptQuote(context.Background(), cmd)
	if err != nil {
		t.Fatalf("idempotent retry failed: %v", err)
	}
	if first.ID != second.ID || first.ID != "agreement_1" {
		t.Fatalf("expected same agreement on retry, first=%+v second=%+v", first, second)
	}
	if store.acceptCount != 1 || store.requestCount != 1 || len(store.outbox) != 1 {
		t.Fatalf("acceptance side effects must happen once: accepts=%d requests=%d outbox=%d", store.acceptCount, store.requestCount, len(store.outbox))
	}
	if store.outbox[0].Name != "marketplace.agreement.created.v1" {
		t.Fatalf("unexpected outbox event: %+v", store.outbox[0])
	}
}

func TestAcceptQuoteRejectsWrongRider(t *testing.T) {
	now := time.Date(2026, 9, 9, 13, 0, 0, 0, time.UTC)
	store := &memoryAcceptanceStore{
		request: marketplace.Request{ID: "req_1", InstanceID: "default", RiderID: "rider_1", ServiceType: "passenger.car", Status: marketplace.RequestOpen, Pickup: geo.Point{Lat: 21, Lng: 105}, RequestedAt: now.Add(-time.Minute), Version: 1},
		quote: marketplace.Quote{ID: "quote_1", RequestID: "req_1", DriverID: "driver_1", Status: marketplace.QuotePending, Fare: money.Must("VND", 52_000), CreatedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute)},
		idempotency: map[string]string{},
	}
	service := AcceptanceService{Store: store}
	_, err := service.AcceptQuote(context.Background(), AcceptQuoteCommand{QuoteID: "quote_1", RiderID: "rider_other", IdempotencyKey: "idem_1", AgreementID: "agreement_1", EventID: "event_1", Now: now})
	if err != ErrRiderMismatch {
		t.Fatalf("expected rider mismatch, got %v", err)
	}
	if store.agreement.ID != "" || len(store.outbox) != 0 {
		t.Fatal("wrong rider must not create agreement or outbox event")
	}
}
