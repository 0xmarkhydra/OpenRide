package httpserver

import (
	"errors"
	"net/http"
	"strconv"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/trips"
)

type registerDevDriverRequest struct {
	ID          string `json:"id"`
	FullName    string `json:"full_name"`
	ServiceType string `json:"service_type"`
}

func (s *Server) registerDevDriver(w http.ResponseWriter, r *http.Request) {
	if s.deps.AppEnv == "production" {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found", nil)
		return
	}
	var req registerDevDriverRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	driver, err := s.deps.Drivers.RegisterApproved(req.ID, req.FullName, req.ServiceType)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: driver})
}

func (s *Server) driverMe(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	driver, err := s.deps.Drivers.Get(driverID)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: driver})
}

func (s *Server) driverTrips(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.deps.Trips.ListForDriver(driverID, limit)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

type availabilityRequest struct {
	Status drivers.AvailabilityStatus `json:"status"`
}

func (s *Server) driverAvailability(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	var req availabilityRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	driver, err := s.deps.Drivers.SetAvailability(driverID, req.Status)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	if req.Status == drivers.AvailabilityOnline && s.deps.Dispatch != nil {
		s.publishOffers(s.deps.Dispatch.DispatchWaiting(50))
	}
	s.publishActor(driverID, "driver.availability", "", driver)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: driver})
}

func (s *Server) driverLocation(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	var req drivers.Location
	if !decodeJSON(w, r, &req) {
		return
	}
	driver, err := s.deps.Drivers.UpdateLocation(driverID, req)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	if activeTrip, activeErr := s.deps.Trips.ActiveForDriver(driverID); activeErr == nil {
		s.publishActor(activeTrip.RiderID, "driver.location", activeTrip.ID, map[string]any{"driver_id": driverID, "location": req})
	}
	if s.deps.Dispatch != nil {
		s.publishOffers(s.deps.Dispatch.DispatchWaiting(50))
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: driver})
}

func (s *Server) currentDriverOffer(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	offer, err := s.deps.Dispatch.CurrentOfferForDriver(driverID)
	if errors.Is(err, dispatch.ErrOfferNotFound) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		s.writeDispatchError(w, err)
		return
	}
	trip, err := s.deps.Trips.Get(offer.TripID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{"offer": offer, "trip": trip}})
}

func (s *Server) acceptDriverOffer(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	offer, trip, err := s.deps.Dispatch.Accept(r.PathValue("id"), driverID)
	if err != nil {
		if errors.Is(err, trips.ErrInvalidState) {
			s.writeDomainError(w, err)
			return
		}
		s.writeDispatchError(w, err)
		return
	}
	s.publishTrip(trip, "trip.accepted", map[string]any{"offer": offer, "trip": trip})
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{"offer": offer, "trip": trip}})
}

func (s *Server) rejectDriverOffer(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	offer, err := s.deps.Dispatch.Reject(r.PathValue("id"), driverID)
	if err != nil {
		s.writeDispatchError(w, err)
		return
	}
	if nextOffer, nextErr := s.deps.Dispatch.CreateOffer(offer.TripID); nextErr == nil {
		s.publishOffers([]dispatch.Offer{nextOffer})
	}
	s.publishActor(driverID, "dispatch.offer_rejected", offer.TripID, offer)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: offer})
}

func (s *Server) driverTripArriving(w http.ResponseWriter, r *http.Request) {
	s.driverTripCommand(w, r, func(id, driverID string) (trips.Trip, error) {
		return s.deps.Ride.MarkArriving(id, driverID)
	})
}

func (s *Server) driverTripArrived(w http.ResponseWriter, r *http.Request) {
	s.driverTripCommand(w, r, func(id, driverID string) (trips.Trip, error) {
		return s.deps.Ride.MarkArrived(id, driverID)
	})
}

func (s *Server) driverTripStart(w http.ResponseWriter, r *http.Request) {
	s.driverTripCommand(w, r, func(id, driverID string) (trips.Trip, error) {
		return s.deps.Ride.Start(id, driverID)
	})
}

func (s *Server) driverTripComplete(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	trip, err := s.deps.Ride.Complete(r.PathValue("id"), driverID, 0)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	if s.deps.Payments != nil {
		_, _ = s.deps.Payments.MarkCashCollected(trip)
	}
	s.publishTrip(trip, "trip."+string(trip.Status), trip)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}

func (s *Server) driverTripCommand(w http.ResponseWriter, r *http.Request, command func(string, string) (trips.Trip, error)) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	trip, err := command(r.PathValue("id"), driverID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.publishTrip(trip, "trip."+string(trip.Status), trip)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}

func (s *Server) driverID(w http.ResponseWriter, r *http.Request) (string, bool) {
	return s.actorID(w, r, auth.RoleDriver, "X-Dev-Driver-ID")
}

func (s *Server) writeDriverError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, drivers.ErrNotFound):
		writeError(w, http.StatusNotFound, "DRIVER_NOT_FOUND", "Driver was not found", nil)
	case errors.Is(err, drivers.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "DRIVER_ALREADY_EXISTS", "Driver already exists", nil)
	case errors.Is(err, drivers.ErrApprovalRequired):
		writeError(w, http.StatusForbidden, "DRIVER_NOT_APPROVED", "Driver is not approved", nil)
	case errors.Is(err, drivers.ErrDriverUnavailable):
		writeError(w, http.StatusConflict, "DRIVER_UNAVAILABLE", "Driver is not available for this action", nil)
	case errors.Is(err, drivers.ErrInvalidLocation), errors.Is(err, drivers.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "DRIVER_INVALID", "Driver data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected server error", nil)
	}
}

func (s *Server) writeDispatchError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dispatch.ErrOfferNotFound):
		writeError(w, http.StatusNotFound, "OFFER_NOT_FOUND", "Offer was not found", nil)
	case errors.Is(err, dispatch.ErrOfferExpired):
		writeError(w, http.StatusConflict, "OFFER_EXPIRED", "Offer has expired", nil)
	case errors.Is(err, dispatch.ErrOfferForbidden):
		writeError(w, http.StatusForbidden, "OFFER_FORBIDDEN", "Offer does not belong to this driver", nil)
	case errors.Is(err, dispatch.ErrOfferUnavailable):
		writeError(w, http.StatusConflict, "OFFER_UNAVAILABLE", "Offer is no longer available", nil)
	case errors.Is(err, dispatch.ErrNoCandidate):
		writeError(w, http.StatusConflict, "NO_DRIVER_AVAILABLE", "No driver is currently available", nil)
	case errors.Is(err, drivers.ErrDriverUnavailable):
		s.writeDriverError(w, err)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected server error", nil)
	}
}
