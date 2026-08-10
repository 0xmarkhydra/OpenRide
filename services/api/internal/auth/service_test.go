package auth

import (
	"context"
	"errors"
	"testing"
)

func newTestAuth(t *testing.T) *Service {
	t.Helper()
	service, err := NewService(NewMemoryStore(), DevelopmentOTPSender{}, "test-secret-123456789", "test", "development")
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func TestOTPAndTokenLifecycle(t *testing.T) {
	service := newTestAuth(t)
	request, err := service.RequestOTP(context.Background(), "0912 345 678", RoleRider)
	if err != nil {
		t.Fatal(err)
	}
	if request.DebugCode == "" {
		t.Fatal("debug code expected in test/development mode")
	}
	phone, err := service.VerifyOTP(request.ChallengeID, RoleRider, request.DebugCode)
	if err != nil {
		t.Fatal(err)
	}
	if phone != "+84912345678" {
		t.Fatalf("phone=%s", phone)
	}

	pair, err := service.IssueTokens("usr_1", RoleRider)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Authenticate(pair.AccessToken)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ActorID != "usr_1" || claims.Role != RoleRider {
		t.Fatalf("claims=%+v", claims)
	}

	rotated, err := service.Refresh(pair.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if rotated.RefreshToken == pair.RefreshToken {
		t.Fatal("refresh token must rotate")
	}
	if _, err := service.Refresh(pair.RefreshToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("old refresh err=%v want revoked", err)
	}
}

func TestOTPWrongCodeLocksAfterMaxAttempts(t *testing.T) {
	service := newTestAuth(t)
	request, err := service.RequestOTP(context.Background(), "+84912345678", RoleDriver)
	if err != nil {
		t.Fatal(err)
	}
	wrong := "000000"
	if wrong == request.DebugCode {
		wrong = "999999"
	}
	for i := 0; i < 4; i++ {
		if _, err := service.VerifyOTP(request.ChallengeID, RoleDriver, wrong); !errors.Is(err, ErrOTPInvalid) {
			t.Fatalf("attempt %d error=%v", i+1, err)
		}
	}
	if _, err := service.VerifyOTP(request.ChallengeID, RoleDriver, wrong); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("fifth error=%v want too many attempts", err)
	}
	if _, err := service.VerifyOTP(request.ChallengeID, RoleDriver, request.DebugCode); !errors.Is(err, ErrTooManyAttempts) {
		t.Fatalf("correct code after lock error=%v", err)
	}
}

func TestProductionRejectsDevelopmentOTP(t *testing.T) {
	_, err := NewService(NewMemoryStore(), DevelopmentOTPSender{}, "production-secret-123456789", "production", "development")
	if !errors.Is(err, ErrUnsafeConfig) {
		t.Fatalf("error=%v want ErrUnsafeConfig", err)
	}
}

func TestInvalidAccessToken(t *testing.T) {
	service := newTestAuth(t)
	if _, err := service.Authenticate("v1.bad.bad"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("error=%v", err)
	}
}
