package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"flashx/services/api/internal/auth"
)

type otpRequestBody struct {
	Phone string    `json:"phone"`
	Role  auth.Role `json:"role"`
}

func (s *Server) requestOTP(w http.ResponseWriter, r *http.Request) {
	if s.deps.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Authentication is unavailable", nil)
		return
	}
	var req otpRequestBody
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Role == auth.RoleAdmin {
		if s.deps.Admin == nil {
			writeError(w, http.StatusForbidden, "ADMIN_NOT_ALLOWED", "Admin access is not configured", nil)
			return
		}
		if _, err := s.deps.Admin.GetByPhone(req.Phone); err != nil {
			writeError(w, http.StatusForbidden, "ADMIN_NOT_ALLOWED", "This phone is not allowed to sign in as admin", nil)
			return
		}
	}
	result, err := s.deps.Auth.RequestOTP(r.Context(), req.Phone, req.Role)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, dataEnvelope{Data: result})
}

type otpVerifyBody struct {
	ChallengeID string    `json:"challenge_id"`
	Role        auth.Role `json:"role"`
	Code        string    `json:"code"`
}

func (s *Server) verifyOTP(w http.ResponseWriter, r *http.Request) {
	if s.deps.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Authentication is unavailable", nil)
		return
	}
	var req otpVerifyBody
	if !decodeJSON(w, r, &req) {
		return
	}
	phone, err := s.deps.Auth.VerifyOTP(req.ChallengeID, req.Role, req.Code)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}

	var (
		actor auth.Actor
		created bool
	)
	switch req.Role {
	case auth.RoleRider:
		if s.deps.Users == nil {
			writeError(w, http.StatusServiceUnavailable, "USER_DOMAIN_UNAVAILABLE", "User service is unavailable", nil)
			return
		}
		user, wasCreated, err := s.deps.Users.FindOrCreateByPhone(phone)
		if err != nil {
			s.writeUserError(w, err)
			return
		}
		actor = auth.Actor{ID: user.ID, Phone: user.Phone, Role: auth.RoleRider}
		created = wasCreated
	case auth.RoleDriver:
		if s.deps.Drivers == nil {
			writeError(w, http.StatusServiceUnavailable, "DRIVER_DOMAIN_UNAVAILABLE", "Driver service is unavailable", nil)
			return
		}
		driver, wasCreated, err := s.deps.Drivers.FindOrCreatePendingByPhone(phone)
		if err != nil {
			s.writeDriverError(w, err)
			return
		}
		actor = auth.Actor{ID: driver.ID, Phone: driver.Phone, Role: auth.RoleDriver}
		created = wasCreated
	case auth.RoleAdmin:
		if s.deps.Admin == nil {
			writeError(w, http.StatusForbidden, "ADMIN_NOT_ALLOWED", "Admin access is not configured", nil)
			return
		}
		adminUser, err := s.deps.Admin.GetByPhone(phone)
		if err != nil {
			writeError(w, http.StatusForbidden, "ADMIN_NOT_ALLOWED", "This phone is not allowed to sign in as admin", nil)
			return
		}
		actor = auth.Actor{ID: adminUser.ID, Phone: adminUser.Phone, Role: auth.RoleAdmin}
		created = false
	default:
		s.writeAuthError(w, auth.ErrInvalidRole)
		return
	}

	tokens, err := s.deps.Auth.IssueTokens(actor.ID, actor.Role)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{
		"actor": actor,
		"created": created,
		"tokens": tokens,
	}})
}

type refreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *Server) refreshAuth(w http.ResponseWriter, r *http.Request) {
	if s.deps.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Authentication is unavailable", nil)
		return
	}
	var req refreshRequestBody
	if !decodeJSON(w, r, &req) {
		return
	}
	tokens, err := s.deps.Auth.Refresh(req.RefreshToken)
	if err != nil {
		s.writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: tokens})
}

func (s *Server) logoutAuth(w http.ResponseWriter, r *http.Request) {
	if s.deps.Auth == nil {
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "Authentication is unavailable", nil)
		return
	}
	var req refreshRequestBody
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.deps.Auth.RevokeRefresh(req.RefreshToken); err != nil {
		s.writeAuthError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) actorID(w http.ResponseWriter, r *http.Request, expected auth.Role, devHeader string) (string, bool) {
	authorization := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(authorization), "bearer ") && s.deps.Auth != nil {
		token := strings.TrimSpace(authorization[len("Bearer "):])
		claims, err := s.deps.Auth.Authenticate(token)
		if err != nil {
			s.writeAuthError(w, err)
			return "", false
		}
		if claims.Role != expected {
			writeError(w, http.StatusForbidden, "AUTH_ROLE_FORBIDDEN", "The authenticated role cannot access this resource", nil)
			return "", false
		}
		return claims.ActorID, true
	}

	if s.deps.AllowDevIdentity && s.deps.AppEnv != "production" {
		if id := strings.TrimSpace(r.Header.Get(devHeader)); id != "" {
			return id, true
		}
	}
	writeError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "Authentication is required", nil)
	return "", false
}

func (s *Server) writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidPhone), errors.Is(err, auth.ErrInvalidRole):
		writeError(w, http.StatusUnprocessableEntity, "AUTH_INVALID_INPUT", "Authentication input is invalid", nil)
	case errors.Is(err, auth.ErrOTPInvalid), errors.Is(err, auth.ErrOTPExpired), errors.Is(err, auth.ErrOTPConsumed), errors.Is(err, auth.ErrChallengeNotFound):
		writeError(w, http.StatusUnauthorized, "OTP_INVALID", "OTP is invalid or expired", nil)
	case errors.Is(err, auth.ErrTooManyAttempts):
		writeError(w, http.StatusTooManyRequests, "OTP_LOCKED", "Too many OTP attempts", nil)
	case errors.Is(err, auth.ErrInvalidToken), errors.Is(err, auth.ErrSessionNotFound), errors.Is(err, auth.ErrSessionExpired), errors.Is(err, auth.ErrSessionRevoked):
		writeError(w, http.StatusUnauthorized, "AUTH_INVALID_TOKEN", "Authentication token is invalid or expired", nil)
	case errors.Is(err, auth.ErrUnsafeConfig):
		writeError(w, http.StatusServiceUnavailable, "AUTH_UNSAFE_CONFIG", "Authentication is not configured safely", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected authentication error", nil)
	}
}
