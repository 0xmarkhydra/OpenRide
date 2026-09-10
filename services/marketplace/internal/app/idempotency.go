package app

import "context"

type idempotencyContextKey struct{}

func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, idempotencyContextKey{}, key)
}

func IdempotencyKeyFromContext(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyContextKey{}).(string)
	return key, ok && key != ""
}
