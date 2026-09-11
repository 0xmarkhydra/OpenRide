package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	outboxLeaseDuration = 30 * time.Second
	outboxMaxAttempts   = 12
	outboxMaxBackoff    = 5 * time.Minute
)

type OutboxMessage struct {
	ID       string
	Subject  string
	Payload  []byte
	Attempts int
}

// RelayOutboxBatch leases ready rows in a short database transaction, publishes
// outside that transaction, then records success/failure using the lease token.
// If the process crashes after publish but before acknowledgement, the lease
// expires and the event may be replayed. Deterministic message IDs preserve
// at-least-once semantics while allowing JetStream/consumers to deduplicate.
func (s *Store) RelayOutboxBatch(ctx context.Context, limit int, publish func(OutboxMessage) error) (int, error) {
	if limit <= 0 {
		limit = 50
	}
	leaseToken, err := newOutboxLeaseToken()
	if err != nil {
		return 0, err
	}
	messages, err := s.claimOutboxBatch(ctx, limit, leaseToken)
	if err != nil {
		return 0, err
	}

	published := 0
	for _, m := range messages {
		if err := publish(m); err != nil {
			if dbErr := s.recordOutboxFailure(ctx, m, leaseToken, err); dbErr != nil {
				return published, dbErr
			}
			continue
		}
		if err := s.recordOutboxSuccess(ctx, m.ID, leaseToken); err != nil {
			return published, err
		}
		published++
	}
	return published, nil
}

func (s *Store) claimOutboxBatch(ctx context.Context, limit int, leaseToken string) ([]OutboxMessage, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()

	rows, err := tx.Query(ctx, `
WITH ready AS (
    SELECT id
    FROM outbox_events
    WHERE published_at IS NULL
      AND dead_lettered_at IS NULL
      AND next_attempt_at <= now()
      AND (locked_until IS NULL OR locked_until <= now())
    ORDER BY next_attempt_at, created_at
    FOR UPDATE SKIP LOCKED
    LIMIT $1
)
UPDATE outbox_events o
SET locked_until = now() + ($2 * interval '1 second'),
    lease_token = $3
FROM ready
WHERE o.id = ready.id
RETURNING o.id, o.event_name, o.payload, o.attempts`, limit, int64(outboxLeaseDuration/time.Second), leaseToken)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]OutboxMessage, 0, limit)
	for rows.Next() {
		var m OutboxMessage
		if err := rows.Scan(&m.ID, &m.Subject, &m.Payload, &m.Attempts); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit outbox lease: %w", err)
	}
	return messages, nil
}

func (s *Store) recordOutboxSuccess(ctx context.Context, id, leaseToken string) error {
	command, err := s.pool.Exec(ctx, `
UPDATE outbox_events
SET published_at = now(),
    attempts = attempts + 1,
    last_error = NULL,
    locked_until = NULL,
    lease_token = NULL
WHERE id = $1
  AND lease_token = $2
  AND published_at IS NULL`, id, leaseToken)
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("outbox acknowledgement lease lost for %s", id)
	}
	return nil
}

func (s *Store) recordOutboxFailure(ctx context.Context, m OutboxMessage, leaseToken string, publishErr error) error {
	nextAttempt := m.Attempts + 1
	backoff := outboxBackoff(nextAttempt)
	deadLetter := nextAttempt >= outboxMaxAttempts
	command, err := s.pool.Exec(ctx, `
UPDATE outbox_events
SET attempts = attempts + 1,
    last_error = $3,
    next_attempt_at = CASE WHEN $4 THEN next_attempt_at ELSE now() + ($5 * interval '1 second') END,
    dead_lettered_at = CASE WHEN $4 THEN now() ELSE dead_lettered_at END,
    locked_until = NULL,
    lease_token = NULL
WHERE id = $1
  AND lease_token = $2
  AND published_at IS NULL`, m.ID, leaseToken, publishErr.Error(), deadLetter, int64(backoff/time.Second))
	if err != nil {
		return err
	}
	if command.RowsAffected() != 1 {
		return fmt.Errorf("outbox failure lease lost for %s", m.ID)
	}
	return nil
}

func outboxBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 8 {
		shift = 8
	}
	backoff := time.Second * time.Duration(1<<shift)
	if backoff > outboxMaxBackoff {
		return outboxMaxBackoff
	}
	return backoff
}

func newOutboxLeaseToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate outbox lease token: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}
