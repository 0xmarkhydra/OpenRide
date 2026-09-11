package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateRequestRejectsUnsafeNestedMinorBeforeDomainDecode(t *testing.T) {
	store := &v2FakeStore{}
	server := newV2TestServer(t, store)
	body := []byte(`{
		"service_type":"passenger.car",
		"pickup":{"lat":19.8067,"lng":105.7852},
		"destination":{"lat":19.7724,"lng":105.7762},
		"constraints":{"max_fare_minor":9007199254740992}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v2/requests", bytes.NewReader(body))
	req.Header.Set("X-OpenRide-Actor-ID", "rider_1")
	req.Header.Set("Idempotency-Key", "unsafe-money-request")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Body.String(); !bytes.Contains([]byte(got), []byte(`"code":"MONEY_MINOR_INVALID"`)) {
		t.Fatalf("expected MONEY_MINOR_INVALID, body=%s", got)
	}
}

func TestCreateRequestAcceptsMaximumSafeNestedMinor(t *testing.T) {
	store := &v2FakeStore{}
	server := newV2TestServer(t, store)
	body := []byte(`{
		"service_type":"passenger.car",
		"pickup":{"lat":19.8067,"lng":105.7852},
		"destination":{"lat":19.7724,"lng":105.7762},
		"constraints":{"max_fare_minor":9007199254740991}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v2/requests", bytes.NewReader(body))
	req.Header.Set("X-OpenRide-Actor-ID", "rider_1")
	req.Header.Set("Idempotency-Key", "safe-money-request")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestValidatePublicMinorFieldsRejectsFractionAndNegativeValues(t *testing.T) {
	for _, body := range []string{
		`{"fare_total_minor":1.5}`,
		`{"fare_total_minor":-1}`,
		`{"nested":{"pickup_fee_minor":9007199254740992}}`,
	} {
		var value any
		decoder := jsonDecoderUseNumber(body)
		if err := decoder.Decode(&value); err != nil {
			t.Fatal(err)
		}
		if err := validatePublicMinorFields(value); err == nil {
			t.Fatalf("expected rejection for %s", body)
		}
	}
}
