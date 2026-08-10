package persistence_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"flashx/services/api/internal/platform/config"
	"flashx/services/api/internal/platform/idempotency"
	"flashx/services/api/internal/platform/persistence"
)

func TestPersistentIdempotencyRoundTrip(t *testing.T) {
	if os.Getenv("FLASHX_INTEGRATION") != "1" { t.Skip("set FLASHX_INTEGRATION=1") }
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resources, err := persistence.Open(ctx, config.Load())
	if err != nil { t.Fatal(err) }
	defer resources.Close()
	store := idempotency.NewRedisStore(resources.Redis)
	key := fmt.Sprintf("integration-%d", time.Now().UnixNano())
	record := idempotency.Record{Fingerprint: "fingerprint", ResourceID: "resource"}
	if err := store.Put("integration", key, record); err != nil { t.Fatal(err) }
	got, ok, err := store.Get("integration", key)
	if err != nil || !ok { t.Fatalf("get: ok=%v err=%v", ok, err) }
	if got != record { t.Fatalf("got=%+v want=%+v", got, record) }
	if err := store.Put("integration", key, idempotency.Record{Fingerprint: "different", ResourceID: "resource"}); err != idempotency.ErrConflict {
		t.Fatalf("expected conflict, got %v", err)
	}
}
