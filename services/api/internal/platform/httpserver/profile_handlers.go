package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/users"
)

func (s *Server) riderMe(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.actorID(w, r, auth.RoleRider, "X-Dev-Rider-ID")
	if !ok {
		return
	}
	if s.deps.Users == nil {
		writeError(w, http.StatusServiceUnavailable, "USER_DOMAIN_UNAVAILABLE", "User service is unavailable", nil)
		return
	}
	user, err := s.deps.Users.Get(riderID)
	if err != nil {
		s.writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: user})
}

type updateRiderProfileRequest struct {
	FullName string `json:"full_name"`
}

func (s *Server) updateRiderMe(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.actorID(w, r, auth.RoleRider, "X-Dev-Rider-ID")
	if !ok {
		return
	}
	var req updateRiderProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	user, err := s.deps.Users.UpdateProfile(riderID, req.FullName)
	if err != nil {
		s.writeUserError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: user})
}

type updateDriverProfileRequest struct {
	FullName    string `json:"full_name"`
	ServiceType string `json:"service_type"`
}

func (s *Server) updateDriverMe(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.actorID(w, r, auth.RoleDriver, "X-Dev-Driver-ID")
	if !ok {
		return
	}
	var req updateDriverProfileRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	driver, err := s.deps.Drivers.UpdateProfile(driverID, req.FullName, req.ServiceType)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: driver})
}

func (s *Server) writeUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, users.ErrNotFound):
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "User was not found", nil)
	case errors.Is(err, users.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "USER_ALREADY_EXISTS", "User already exists", nil)
	case errors.Is(err, users.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "USER_INVALID", "User data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected user error", nil)
	}
}

