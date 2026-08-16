package httpserver

import (
	"context"
	"errors"
	"net/http"
	"time"

	"flashx/services/api/internal/notifications"
)

func (s *Server) registerRiderPushDevice(w http.ResponseWriter, r *http.Request) {
	actorID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	s.registerPushDevice(w, r, actorID, notifications.RoleRider)
}

func (s *Server) disableRiderPushDevice(w http.ResponseWriter, r *http.Request) {
	actorID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	s.disablePushDevice(w, actorID, notifications.RoleRider, r.PathValue("id"))
}

func (s *Server) registerDriverPushDevice(w http.ResponseWriter, r *http.Request) {
	actorID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	s.registerPushDevice(w, r, actorID, notifications.RoleDriver)
}

func (s *Server) disableDriverPushDevice(w http.ResponseWriter, r *http.Request) {
	actorID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	s.disablePushDevice(w, actorID, notifications.RoleDriver, r.PathValue("id"))
}

func (s *Server) registerPushDevice(w http.ResponseWriter, r *http.Request, actorID string, role notifications.Role) {
	if s.deps.Notifications == nil {
		writeError(w, http.StatusServiceUnavailable, "PUSH_UNAVAILABLE", "Push notification service is unavailable", nil)
		return
	}
	var input notifications.RegisterDeviceInput
	if !decodeJSON(w, r, &input) {
		return
	}
	device, err := s.deps.Notifications.RegisterDevice(actorID, role, input)
	if err != nil {
		s.writeNotificationError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: device})
}

func (s *Server) disablePushDevice(w http.ResponseWriter, actorID string, role notifications.Role, deviceID string) {
	if s.deps.Notifications == nil {
		writeError(w, http.StatusServiceUnavailable, "PUSH_UNAVAILABLE", "Push notification service is unavailable", nil)
		return
	}
	if err := s.deps.Notifications.DisableDevice(actorID, role, deviceID); err != nil {
		s.writeNotificationError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) writeNotificationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, notifications.ErrNotFound):
		writeError(w, http.StatusNotFound, "PUSH_DEVICE_NOT_FOUND", "Push device was not found", nil)
	case errors.Is(err, notifications.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "PUSH_DEVICE_INVALID", "Push device data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "PUSH_ERROR", "Unexpected push notification error", nil)
	}
}

func (s *Server) RunNotificationDispatch(ctx context.Context, interval time.Duration) {
	if s.deps.Notifications == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Push delivery is best-effort. Durable status remains in the outbox,
			// while trip/dispatch state transitions stay independent of provider failures.
			_ = s.deps.Notifications.ProcessPending(ctx, 50)
		}
	}
}
