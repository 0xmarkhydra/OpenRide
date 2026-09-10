package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	modulecatalog "github.com/0xmarkhydra/OpenRide/packages/modules-go/catalog"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

type v2FakeStore struct {
	tariff marketplace.DriverTariff
	request marketplace.Request
	offers []marketplace.Quote
}

func (f *v2FakeStore) CreateTariff(_ context.Context, t marketplace.DriverTariff) error { f.tariff = t; return nil }
func (f *v2FakeStore) ListDriverTariffs(context.Context, string) ([]marketplace.DriverTariff, error) { return nil, nil }
func (f *v2FakeStore) CreateRequest(context.Context, marketplace.Request) error { return nil }
func (f *v2FakeStore) GetRequest(context.Context, string) (marketplace.Request, error) { return f.request, nil }
func (f *v2FakeStore) CancelRequest(context.Context, string, string) error { return nil }
func (f *v2FakeStore) ListEligibleRequests(context.Context, string, marketplace.ServiceType) ([]marketplace.Request, error) { return nil, nil }
func (f *v2FakeStore) SubmitQuote(context.Context, marketplace.Quote) error { return nil }
func (f *v2FakeStore) ListOffers(context.Context, string) ([]marketplace.Quote, error) { return f.offers, nil }
func (f *v2FakeStore) WithdrawQuote(context.Context, string, string) error { return nil }

type noOpAcceptanceStore struct{}
func (noOpAcceptanceStore) WithinTx(context.Context, func(app.AcceptanceTx) error) error { return nil }

func newV2TestServer(t *testing.T, store V2Store) *Server {
	t.Helper()
	s, err := New(Config{
		Services: modulecatalog.DefaultRegistry(),
		V2Store: store,
		Acceptance: app.AcceptanceService{Store: noOpAcceptanceStore{}},
	})
	if err != nil { t.Fatalf("new server: %v", err) }
	return s
}

func TestCreateDriverTariffAcceptsFlatSDKContract(t *testing.T) {
	store := &v2FakeStore{}
	server := newV2TestServer(t, store)
	body := []byte(`{"service_type":"passenger.car","quote_mode":"manual","currency":"VND","per_km_minor":5000}`)
	req := httptest.NewRequest(http.MethodPost, "/v2/drivers/me/tariffs", bytes.NewReader(body))
	req.Header.Set("X-OpenRide-Actor-ID", "driver_1")
	req.Header.Set("Idempotency-Key", "tariff-1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	if store.tariff.DriverID != "driver_1" || store.tariff.PerKM.Currency != "VND" || store.tariff.PerKM.Minor != 5000 {
		t.Fatalf("unexpected tariff: %+v", store.tariff)
	}
}

func TestCreateDriverTariffRejectsInvalidCurrencyWithoutPanic(t *testing.T) {
	store := &v2FakeStore{}
	server := newV2TestServer(t, store)
	body := []byte(`{"service_type":"passenger.car","quote_mode":"manual","currency":"VN","per_km_minor":5000}`)
	req := httptest.NewRequest(http.MethodPost, "/v2/drivers/me/tariffs", bytes.NewReader(body))
	req.Header.Set("X-OpenRide-Actor-ID", "driver_1")
	req.Header.Set("Idempotency-Key", "tariff-2")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnprocessableEntity { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
}

func TestListOffersUsesFlatSDKContract(t *testing.T) {
	now := time.Now().UTC()
	store := &v2FakeStore{
		request: marketplace.Request{ID:"req_1", InstanceID:"default", RiderID:"rider_1", ServiceType:"passenger.car", Status:marketplace.RequestOpen, Pickup:geo.Point{Lat:19.8,Lng:105.7}, RequestedAt:now, Version:1},
		offers: []marketplace.Quote{{ID:"quote_1", RequestID:"req_1", DriverID:"driver_1", Status:marketplace.QuotePending, Fare:money.Must("VND",50000), PickupDistanceM:1200, PickupETAS:300, CreatedAt:now, ExpiresAt:now.Add(time.Minute), Explanation:[]string{"nearby"}}},
	}
	server := newV2TestServer(t, store)
	req := httptest.NewRequest(http.MethodGet, "/v2/requests/req_1/offers", nil)
	req.Header.Set("X-OpenRide-Actor-ID", "rider_1")
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK { t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String()) }
	var payload struct { Data []map[string]any `json:"data"` }
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil { t.Fatal(err) }
	if len(payload.Data) != 1 || payload.Data[0]["quote_id"] != "quote_1" || payload.Data[0]["fare_total_minor"] != float64(50000) {
		t.Fatalf("unexpected response: %s", rec.Body.String())
	}
	if _, exists := payload.Data[0]["fare"]; exists { t.Fatalf("domain money object leaked into V2 response: %s", rec.Body.String()) }
}
