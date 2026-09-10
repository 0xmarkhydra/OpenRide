package store

import (
	"errors"
	"testing"
)

func TestVerifyReplayAcceptsMatchingRecord(t *testing.T) {
	rec := idempotencyRecord{RequestHash: "hash-a", ResourceType: "request", ResourceID: "req_1"}
	if err := verifyReplay(rec, "hash-a", "request"); err != nil {
		t.Fatalf("expected matching replay to pass: %v", err)
	}
}

func TestVerifyReplayRejectsPayloadReuse(t *testing.T) {
	rec := idempotencyRecord{RequestHash: "hash-a", ResourceType: "request", ResourceID: "req_1"}
	if err := verifyReplay(rec, "hash-b", "request"); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}
