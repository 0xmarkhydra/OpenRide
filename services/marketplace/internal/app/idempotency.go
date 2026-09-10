package app

import "context"

type idempotencyContextKey struct{}

type IdempotencyContext struct {
	Key         string
	RequestHash string
}

func WithIdempotency(ctx context.Context, key, requestHash string) context.Context {
	return context.WithValue(ctx, idempotencyContextKey{}, IdempotencyContext{Key: key, RequestHash: requestHash})
}

func IdempotencyFromContext(ctx context.Context) (IdempotencyContext, bool) {
	value, ok := ctx.Value(idempotencyContextKey{}).(IdempotencyContext)
	return value, ok && value.Key != "" && value.RequestHash != ""
}
