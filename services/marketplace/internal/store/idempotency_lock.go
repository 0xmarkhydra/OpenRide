package store

import "context"

func (t *acceptanceTx) LockIdempotency(ctx context.Context, actorID, key string) error {
	_, err := t.tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, actorID+":"+key+":accept_quote")
	return err
}
