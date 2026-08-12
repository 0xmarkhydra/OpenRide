package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/customervehicles"
)

type createVehicleRequest struct {
	Type           string `json:"type"`
	LicensePlate   string `json:"license_plate"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	Year           int    `json:"year"`
	Color          string `json:"color"`
	Transmission   string `json:"transmission"`
	Seats          int    `json:"seats"`
	Notes          string `json:"notes"`
	PhotoObjectKey string `json:"photo_object_key"`
}

func (s *Server) riderVehicles(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.CustomerVehicles == nil {
		writeError(w, http.StatusServiceUnavailable, "VEHICLE_DOMAIN_UNAVAILABLE", "Vehicle service is unavailable", nil)
		return
	}
	items, err := s.deps.CustomerVehicles.ListForOwner(riderID)
	if err != nil {
		s.writeVehicleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) createRiderVehicle(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.CustomerVehicles == nil {
		writeError(w, http.StatusServiceUnavailable, "VEHICLE_DOMAIN_UNAVAILABLE", "Vehicle service is unavailable", nil)
		return
	}
	var req createVehicleRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	vehicle, err := s.deps.CustomerVehicles.Create(customervehicles.CreateInput{
		OwnerUserID: riderID, Type: req.Type, LicensePlate: req.LicensePlate,
		Brand: req.Brand, Model: req.Model, Year: req.Year, Color: req.Color,
		Transmission: req.Transmission, Seats: req.Seats, Notes: req.Notes,
		PhotoObjectKey: req.PhotoObjectKey,
	})
	if err != nil {
		s.writeVehicleError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: vehicle})
}

func (s *Server) getRiderVehicle(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.CustomerVehicles == nil {
		writeError(w, http.StatusServiceUnavailable, "VEHICLE_DOMAIN_UNAVAILABLE", "Vehicle service is unavailable", nil)
		return
	}
	vehicle, err := s.deps.CustomerVehicles.GetForOwner(r.PathValue("id"), riderID)
	if err != nil {
		s.writeVehicleError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: vehicle})
}

func (s *Server) writeVehicleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, customervehicles.ErrNotFound):
		writeError(w, http.StatusNotFound, "VEHICLE_NOT_FOUND", "Vehicle was not found", nil)
	case errors.Is(err, customervehicles.ErrForbidden):
		writeError(w, http.StatusForbidden, "VEHICLE_FORBIDDEN", "Vehicle does not belong to this customer", nil)
	case errors.Is(err, customervehicles.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "VEHICLE_ALREADY_EXISTS", "This vehicle is already saved", nil)
	case errors.Is(err, customervehicles.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "VEHICLE_INVALID", "Vehicle data is invalid for this service", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected vehicle error", nil)
	}
}
