package notifications

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRegisterDeviceDoesNotExposeTokenInJSON(t *testing.T) {
	service := NewService(NewMemoryStore(), DisabledProvider{})
	device, err := service.RegisterDevice("rider-1", RoleRider, RegisterDeviceInput{
		Platform: PlatformIOS,
		Token:    "0123456789abcdef0123456789abcdef",
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(device)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" || contains(string(payload), "0123456789abcdef") {
		t.Fatalf("push token leaked in JSON: %s", payload)
	}
}

func TestDisabledProviderMarksOutboxSkipped(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, DisabledProvider{})
	if _, err := service.RegisterDevice("driver-1", RoleDriver, RegisterDeviceInput{
		Platform: PlatformAndroid,
		Token:    "driver-token-0123456789abcdef",
	}); err != nil {
		t.Fatal(err)
	}
	if err := service.Enqueue("driver-1", RoleDriver, "dispatch.offer", "trip-1"); err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessPending(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if len(store.messages) != 1 {
		t.Fatalf("messages=%d want=1", len(store.messages))
	}
	for _, message := range store.messages {
		if message.Status != StatusSkipped || message.LastError != "provider_disabled" || message.Attempts != 1 {
			t.Fatalf("unexpected outbox state: %+v", message)
		}
	}
}

func TestNoActiveDeviceMarksOutboxSkipped(t *testing.T) {
	store := NewMemoryStore()
	service := NewService(store, DisabledProvider{})
	if err := service.Enqueue("rider-1", RoleRider, "trip.accepted", "trip-1"); err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessPending(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	for _, message := range store.messages {
		if message.Status != StatusSkipped || message.LastError != "no_active_device" {
			t.Fatalf("unexpected outbox state: %+v", message)
		}
	}
}

func contains(value, part string) bool {
	for i := 0; i+len(part) <= len(value); i++ {
		if value[i:i+len(part)] == part {
			return true
		}
	}
	return false
}
