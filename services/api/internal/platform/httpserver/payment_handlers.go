package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/payments"
)

func (s *Server) getTripPayment(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.Payments == nil {
		writeError(w, http.StatusServiceUnavailable, "PAYMENT_UNAVAILABLE", "Payment service is unavailable", nil)
		return
	}
	trip, err := s.deps.Trips.GetForRider(r.PathValue("id"), riderID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	payment, err := s.deps.Payments.GetForTrip(trip)
	if errors.Is(err, payments.ErrNotFound) {
		payment, err = s.deps.Payments.EnsureCash(trip)
	}
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
