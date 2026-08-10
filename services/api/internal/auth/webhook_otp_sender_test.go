package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookOTPSender(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("authorization=%q", got)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["phone"] != "+84912345678" || body["code"] != "123456" {
			t.Fatalf("unexpected payload: %+v", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sender := newWebhookOTPSenderForTest(server.URL, "secret", server.Client())
	if err := sender.Send(context.Background(), "+84912345678", "123456"); err != nil {
		t.Fatal(err)
	}
}

func TestWebhookOTPSenderRequiresCredentials(t *testing.T) {
	if _, err := NewWebhookOTPSender("", ""); err == nil {
		t.Fatal("expected unsafe config error")
	}
}
