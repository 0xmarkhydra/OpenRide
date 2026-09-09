package store

import (
	"context"
	"fmt"
)

type OutboxMessage struct {
	ID      string
	Subject string
	Payload []byte
}

// RelayOutboxBatch claims rows with SKIP LOCKED and keeps the row locks until
// publish results have been recorded. Delivery is intentionally at-least-once:
// if publishing succeeds and the DB commit later fails, the deterministic event
// ID lets JetStream/consumers deduplicate the replay.
func (s *Store) RelayOutboxBatch(ctx context.Context, limit int, publish func(OutboxMessage) error) (int, error) {
	if limit <= 0 { limit = 50 }
	tx, err := s.pool.Begin(ctx)
	if err != nil { return 0, err }
	defer func(){ _ = tx.Rollback(context.Background()) }()

	rows, err := tx.Query(ctx, `SELECT id,event_name,payload FROM outbox_events WHERE published_at IS NULL ORDER BY created_at FOR UPDATE SKIP LOCKED LIMIT $1`, limit)
	if err != nil { return 0, err }
	var messages []OutboxMessage
	for rows.Next() {
		var m OutboxMessage
		if err := rows.Scan(&m.ID,&m.Subject,&m.Payload); err != nil { rows.Close(); return 0, err }
		messages = append(messages,m)
	}
	if err := rows.Err(); err != nil { rows.Close(); return 0, err }
	rows.Close()

	published := 0
	for _, m := range messages {
		if err := publish(m); err != nil {
			if _, dbErr := tx.Exec(ctx, `UPDATE outbox_events SET attempts=attempts+1,last_error=$2 WHERE id=$1`,m.ID,err.Error()); dbErr != nil { return published, dbErr }
			continue
		}
		if _, err := tx.Exec(ctx, `UPDATE outbox_events SET published_at=now(),attempts=attempts+1,last_error=NULL WHERE id=$1`,m.ID); err != nil { return published,err }
		published++
	}
	if err := tx.Commit(ctx); err != nil { return published,fmt.Errorf("commit outbox relay: %w",err) }
	return published,nil
}
