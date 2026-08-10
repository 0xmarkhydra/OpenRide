package auth

import "context"

type OTPSender interface {
	Send(ctx context.Context, phone, code string) error
}

// DevelopmentOTPSender intentionally performs no external call. The HTTP API
// may expose DebugCode only outside production so local/staging flows are fully testable.
type DevelopmentOTPSender struct{}

func (DevelopmentOTPSender) Send(_ context.Context, _, _ string) error { return nil }
