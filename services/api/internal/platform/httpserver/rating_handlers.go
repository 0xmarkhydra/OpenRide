package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/ratings"
	"flashx/services/api/internal/trips"
)

type createRatingRequest struct {
	Stars   int16  `json:"stars"`
	Comment string `json:"comment"`
}

func (s *Server) createTripRating(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.Ratings == nil {
		writeError(w, http.StatusServiceUnavailable, "RATING_UNAVAILABLE", "Rating service is unavailable", nil)
		return
	}
	var req createRatingRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	rating, err := s.deps.Ratings.Create(riderID, r.PathValue("id"), req.Stars, req.Comment)
	if err != nil {
		s.writeRatingError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: rating})
}

func (s *Server) getTripRating(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.Ratings == nil {
		writeError(w, http.StatusServiceUnavailable, "RATING_UNAVAILABLE", "Rating service is unavailable", nil)
		return
	}
	rating, err := s.deps.Ratings.GetByTripForRider(riderID, r.PathValue("id"))
	if err != nil {
		s.writeRatingError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: rating})
}

func (s *Server) writeRatingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ratings.ErrNotFound):
		writeError(w, http.StatusNotFound, "RATING_NOT_FOUND", "Rating was not found", nil)
	case errors.Is(err, ratings.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "RATING_ALREADY_EXISTS", "Trip has already been rated", nil)
	case errors.Is(err, ratings.ErrForbidden), errors.Is(err, trips.ErrForbidden):
		writeError(w, http.StatusForbidden, "RATING_FORBIDDEN", "Trip cannot be rated by this rider", nil)
	case errors.Is(err, ratings.ErrInvalidInput), errors.Is(err, trips.ErrInvalidState):
		writeError(w, http.StatusUnprocessableEntity, "RATING_INVALID", "Rating is invalid for this trip", nil)
	case errors.Is(err, trips.ErrNotFound):
		writeError(w, http.StatusNotFound, "TRIP_NOT_FOUND", "Trip was not found", nil)
	default:
		writeError(w, http.StatusInternalServerError, "RATING_ERROR", "Unexpected rating error", nil)
	}
}
