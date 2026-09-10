package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

func TestAcceptQuoteHighContentionPostgres(t *testing.T) {
	databaseURL := strings.TrimSpace(os.Getenv("MARKETPLACE_TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("MARKETPLACE_TEST_DATABASE_URL is not set")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse MARKETPLACE_TEST_DATABASE_URL: %v", err)
	}
	if !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
		t.Fatalf("refusing destructive integration test against non-test database %q", parsed.Path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	s, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open marketplace store: %v", err)
	}
	defer s.Close()

	truncate := func() {
		_, truncErr := s.pool.Exec(context.Background(), `TRUNCATE marketplace_idempotency, outbox_events, marketplace_agreements, marketplace_quotes, mobility_requests, driver_tariffs RESTART IDENTITY CASCADE`)
		if truncErr != nil {
			t.Logf("truncate integration database: %v", truncErr)
		}
	}
	truncate()
	t.Cleanup(truncate)

	now := time.Now().UTC().Truncate(time.Microsecond)
	prefix := fmt.Sprintf("concurrency_%d", now.UnixNano())
	requestID := prefix + "_request"
	riderID := prefix + "_rider"

	req := marketplace.Request{
		ID:          requestID,
		InstanceID:  "integration-test",
		RiderID:     riderID,
		ServiceType: marketplace.ServiceType("passenger.car"),
		Status:      marketplace.RequestOpen,
		Pickup:      geo.Point{Lat: 10.7769, Lng: 106.7009},
		RequestedAt: now,
		Version:     1,
	}
	if err := s.CreateRequest(ctx, req); err != nil {
		t.Fatalf("create request: %v", err)
	}

	const contenders = 100
	quoteIDs := make([]string, contenders)
	for i := 0; i < contenders; i++ {
		quoteIDs[i] = fmt.Sprintf("%s_quote_%03d", prefix, i)
		q := marketplace.Quote{
			ID:        quoteIDs[i],
			RequestID: requestID,
			DriverID:  fmt.Sprintf("%s_driver_%03d", prefix, i),
			Status:    marketplace.QuotePending,
			Fare:      money.Must("VND", 50_000+int64(i)),
			CreatedAt: now,
			ExpiresAt: now.Add(5 * time.Minute),
		}
		if err := s.SubmitQuote(ctx, q); err != nil {
			t.Fatalf("submit quote %d: %v", i, err)
		}
	}

	service := app.AcceptanceService{Store: s}
	start := make(chan struct{})
	var wg sync.WaitGroup
	var successes atomic.Int32
	errs := make(chan error, contenders)

	for i := 0; i < contenders; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := service.AcceptQuote(ctx, app.AcceptQuoteCommand{
				QuoteID:        quoteIDs[i],
				RiderID:        riderID,
				IdempotencyKey: fmt.Sprintf("%s_key_%03d", prefix, i),
				AgreementID:    fmt.Sprintf("%s_agreement_%03d", prefix, i),
				EventID:        fmt.Sprintf("%s_event_%03d", prefix, i),
				Now:            now.Add(time.Second),
			})
			if err == nil {
				successes.Add(1)
				return
			}
			errs <- err
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	if got := successes.Load(); got != 1 {
		t.Fatalf("expected exactly one successful acceptance, got %d", got)
	}

	for err := range errs {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01") {
			t.Fatalf("retryable PostgreSQL concurrency error escaped application boundary: %s: %v", pgErr.Code, err)
		}
		if !errors.Is(err, app.ErrRequestUnavailable) && !errors.Is(err, app.ErrQuoteUnavailable) {
			t.Fatalf("unexpected loser error: %T: %v", err, err)
		}
	}

	var agreements int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM marketplace_agreements WHERE request_id=$1`, requestID).Scan(&agreements); err != nil {
		t.Fatalf("count agreements: %v", err)
	}
	if agreements != 1 {
		t.Fatalf("expected one agreement row, got %d", agreements)
	}

	var accepted, pending, invalidated int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FILTER (WHERE status='accepted'), count(*) FILTER (WHERE status='pending'), count(*) FILTER (WHERE status='invalidated') FROM marketplace_quotes WHERE request_id=$1`, requestID).Scan(&accepted, &pending, &invalidated); err != nil {
		t.Fatalf("count quote states: %v", err)
	}
	if accepted != 1 || pending != 0 || invalidated != contenders-1 {
		t.Fatalf("unexpected quote states: accepted=%d pending=%d invalidated=%d", accepted, pending, invalidated)
	}

	var status string
	if err := s.pool.QueryRow(ctx, `SELECT status FROM mobility_requests WHERE id=$1`, requestID).Scan(&status); err != nil {
		t.Fatalf("read request status: %v", err)
	}
	if status != string(marketplace.RequestAgreed) {
		t.Fatalf("expected request agreed, got %q", status)
	}

	var outbox int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE event_name='openride.marketplace.agreement.created.v1' AND payload->>'request_id'=$1`, requestID).Scan(&outbox); err != nil {
		t.Fatalf("count outbox events: %v", err)
	}
	if outbox != 1 {
		t.Fatalf("expected one agreement-created outbox event, got %d", outbox)
	}

	var acceptanceIdempotency int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM marketplace_idempotency WHERE command_name='accept_quote' AND agreement_id IS NOT NULL AND actor_id=$1`, riderID).Scan(&acceptanceIdempotency); err != nil {
		t.Fatalf("count acceptance idempotency rows: %v", err)
	}
	if acceptanceIdempotency != 1 {
		t.Fatalf("expected one committed acceptance idempotency row, got %d", acceptanceIdempotency)
	}
}
