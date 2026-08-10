package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"

	"flashx/services/api/internal/platform/ids"
)

var (
	ErrInvalidPhone  = errors.New("invalid phone")
	ErrInvalidRole   = errors.New("invalid auth role")
	ErrInvalidToken  = errors.New("invalid access token")
	ErrUnsafeConfig  = errors.New("unsafe auth configuration")
)

type ActorResolver interface {
	FindOrCreateActor(context.Context, string, Role) (Actor, error)
}

type Service struct {
	store       Store
	actors      ActorResolver
	sender      OTPSender
	secret      []byte
	appEnv      string
	otpMode     string
	now         func() time.Time
	otpTTL      time.Duration
	accessTTL   time.Duration
	refreshTTL  time.Duration
	maxAttempts int
}

func NewService(store Store, sender OTPSender, secret, appEnv, otpMode string) (*Service, error) {
	secret = strings.TrimSpace(secret)
	appEnv = strings.TrimSpace(appEnv)
	otpMode = strings.TrimSpace(otpMode)
	if store == nil || sender == nil || len(secret) < 16 {
		return nil, ErrUnsafeConfig
	}
	if appEnv == "production" && (otpMode == "development" || secret == "dev-only-change-me") {
		return nil, ErrUnsafeConfig
	}
	return &Service{
		store: store,
		sender: sender,
		secret: []byte(secret),
		appEnv: appEnv,
		otpMode: otpMode,
		now: func() time.Time { return time.Now().UTC() },
		otpTTL: 5 * time.Minute,
		accessTTL: 15 * time.Minute,
		refreshTTL: 30 * 24 * time.Hour,
		maxAttempts: 5,
	}, nil
}

func (s *Service) SetActorResolver(resolver ActorResolver) {
	s.actors = resolver
}

func (s *Service) ResolveActor(ctx context.Context, phone string, role Role) (Actor, error) {
	if s.actors == nil {
		return Actor{}, ErrUnsafeConfig
	}
	return s.actors.FindOrCreateActor(ctx, phone, role)
}

func (s *Service) RequestOTP(ctx context.Context, phone string, role Role) (OTPRequestResult, error) {
	if role != RoleRider && role != RoleDriver && role != RoleAdmin {
		return OTPRequestResult{}, ErrInvalidRole
	}
	normalized, err := NormalizeVietnamPhone(phone)
	if err != nil {
		return OTPRequestResult{}, err
	}
	challengeID := ids.New("otp")
	code, err := generateOTP()
	if err != nil {
		return OTPRequestResult{}, err
	}
	now := s.now()
	challenge := Challenge{
		ID: challengeID,
		Phone: normalized,
		Role: role,
		CodeHash: s.otpHash(challengeID, code),
		ExpiresAt: now.Add(s.otpTTL),
		CreatedAt: now,
	}
	if err := s.store.CreateChallenge(challenge); err != nil {
		return OTPRequestResult{}, err
	}
	if err := s.sender.Send(ctx, normalized, code); err != nil {
		return OTPRequestResult{}, err
	}
	result := OTPRequestResult{ChallengeID: challengeID, ExpiresAt: challenge.ExpiresAt}
	if s.appEnv != "production" && s.otpMode == "development" {
		result.DebugCode = code
	}
	return result, nil
}

func (s *Service) VerifyOTP(challengeID string, role Role, code string) (string, error) {
	code = strings.TrimSpace(code)
	if challengeID == "" || len(code) != 6 {
		return "", ErrOTPInvalid
	}
	challenge, err := s.store.VerifyChallenge(challengeID, role, s.otpHash(challengeID, code), s.now(), s.maxAttempts)
	if err != nil {
		return "", err
	}
	return challenge.Phone, nil
}

func (s *Service) IssueTokens(actorID string, role Role) (TokenPair, error) {
	if actorID == "" || (role != RoleRider && role != RoleDriver && role != RoleAdmin) {
		return TokenPair{}, ErrInvalidToken
	}
	now := s.now()
	claims := Claims{
		ActorID: actorID,
		Role: role,
		ExpiresAt: now.Add(s.accessTTL).Unix(),
		Nonce: ids.New("nonce"),
	}
	access, err := s.signAccess(claims)
	if err != nil {
		return TokenPair{}, err
	}
	refresh, err := randomToken(32)
	if err != nil {
		return TokenPair{}, err
	}
	session := RefreshSession{
		ID: ids.New("ses"),
		ActorID: actorID,
		Role: role,
		TokenHash: hashToken(refresh),
		ExpiresAt: now.Add(s.refreshTTL),
		CreatedAt: now,
	}
	if err := s.store.CreateRefreshSession(session); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken: access,
		RefreshToken: refresh,
		TokenType: "Bearer",
		ExpiresAt: time.Unix(claims.ExpiresAt, 0).UTC(),
	}, nil
}

func (s *Service) Authenticate(accessToken string) (Claims, error) {
	parts := strings.Split(strings.TrimSpace(accessToken), ".")
	if len(parts) != 3 || parts[0] != "v1" {
		return Claims{}, ErrInvalidToken
	}
	message := parts[0] + "." + parts[1]
	expected := s.sign([]byte(message))
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(expected, actual) {
		return Claims{}, ErrInvalidToken
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if claims.ActorID == "" || claims.ExpiresAt <= s.now().Unix() {
		return Claims{}, ErrInvalidToken
	}
	if claims.Role != RoleRider && claims.Role != RoleDriver && claims.Role != RoleAdmin {
		return Claims{}, ErrInvalidToken
	}
	return claims, nil
}

func (s *Service) Refresh(refreshToken string) (TokenPair, error) {
	session, err := s.store.GetRefreshByHash(hashToken(strings.TrimSpace(refreshToken)))
	if err != nil {
		return TokenPair{}, err
	}
	now := s.now()
	if session.RevokedAt != nil {
		return TokenPair{}, ErrSessionRevoked
	}
	if !session.ExpiresAt.After(now) {
		return TokenPair{}, ErrSessionExpired
	}
	if err := s.store.RevokeRefresh(session.ID, now); err != nil {
		return TokenPair{}, err
	}
	return s.IssueTokens(session.ActorID, session.Role)
}

func (s *Service) RevokeRefresh(refreshToken string) error {
	session, err := s.store.GetRefreshByHash(hashToken(strings.TrimSpace(refreshToken)))
	if err != nil {
		return err
	}
	return s.store.RevokeRefresh(session.ID, s.now())
}

func (s *Service) signAccess(claims Claims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	message := "v1." + encoded
	signature := base64.RawURLEncoding.EncodeToString(s.sign([]byte(message)))
	return message + "." + signature, nil
}

func (s *Service) sign(message []byte) []byte {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write(message)
	return mac.Sum(nil)
}

func (s *Service) otpHash(challengeID, code string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(challengeID + "|" + code))
	return hex.EncodeToString(mac.Sum(nil))
}

func NormalizeVietnamPhone(phone string) (string, error) {
	var b strings.Builder
	for _, r := range strings.TrimSpace(phone) {
		if unicode.IsDigit(r) || (r == '+' && b.Len() == 0) {
			b.WriteRune(r)
		}
	}
	value := b.String()
	if strings.HasPrefix(value, "0") && len(value) >= 9 && len(value) <= 11 {
		value = "+84" + value[1:]
	}
	if !strings.HasPrefix(value, "+") {
		return "", ErrInvalidPhone
	}
	digits := value[1:]
	if len(digits) < 8 || len(digits) > 15 {
		return "", ErrInvalidPhone
	}
	for _, r := range digits {
		if !unicode.IsDigit(r) {
			return "", ErrInvalidPhone
		}
	}
	return value, nil
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
