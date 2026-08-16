package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/payments"
	"flashx/services/api/internal/trips"
)

func (s *Server) getTripPayment(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if !s.paymentAvailable(w) {
		return
	}
	trip, err := s.deps.Trips.GetForRider(r.PathValue("id"), riderID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.writeTripPayment(w, trip)
}

func (s *Server) driverTripPayment(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.paymentAvailable(w) {
		return
	}
	trip, err := s.deps.Trips.GetForDriver(r.PathValue("id"), driverID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.writeTripPayment(w, trip)
}

func (s *Server) adminTripPayment(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if !s.paymentAvailable(w) {
		return
	}
	trip, err := s.deps.Trips.Get(r.PathValue("id"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.writeTripPayment(w, trip)
}

func (s *Server) paymentAvailable(w http.ResponseWriter) bool {
	if s.deps.Payments != nil {
		return true
	}
	writeError(w, http.StatusServiceUnavailable, "PAYMENT_UNAVAILABLE", "Payment service is unavailable", nil)
	return false
}

func (s *Server) writeTripPayment(w http.ResponseWriter, trip trips.Trip) {
	payment, err := s.deps.Payments.ReconcileCash(trip)
	if err != nil {
		s.writePaymentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: payment})
}

func (s *Server) writePaymentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, payments.ErrNotFound):
		writeError(w, http.StatusNotFound, "PAYMENT_NOT_FOUND", "Payment was not found", nil)
	case errors.Is(err, payments.ErrInvalidState):
		writeError(w, http.StatusConflict, "PAYMENT_INVALID_STATE", "Payment cannot perform this action in its current state", nil)
	default:
		writeError(w, http.StatusInternalServerError, "PAYMENT_ERROR", "Unexpected payment error", nil)
	}
}
