package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"flashx/services/api/internal/places"
	"flashx/services/api/internal/trips"
)

func (s *Server) searchPlaces(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.riderID(w, r); !ok {
		return
	}
	if s.deps.Places == nil {
		writeError(w, http.StatusServiceUnavailable, "PLACE_SEARCH_UNAVAILABLE", "Place search is unavailable; choose a point on the map instead", nil)
		return
	}
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len([]rune(query)) < 3 || len([]rune(query)) > 160 {
		writeError(w, http.StatusUnprocessableEntity, "PLACE_QUERY_INVALID", "Place query must contain between 3 and 160 characters", nil)
		return
	}
	bias, ok := parsePlaceBias(r)
	if !ok {
		writeError(w, http.StatusUnprocessableEntity, "PLACE_BIAS_INVALID", "Location bias must contain valid lat and lng values", nil)
		return
	}
	results, err := s.deps.Places.Search(r.Context(), query, bias)
	if err != nil {
		switch {
		case errors.Is(err, places.ErrInvalidQuery):
			writeError(w, http.StatusUnprocessableEntity, "PLACE_QUERY_INVALID", "Place query is invalid", nil)
		case errors.Is(err, places.ErrUnavailable):
			writeError(w, http.StatusServiceUnavailable, "PLACE_SEARCH_UNAVAILABLE", "Place search is temporarily unavailable; choose a point on the map instead", nil)
		default:
			writeError(w, http.StatusInternalServerError, "PLACE_SEARCH_ERROR", "Unexpected place search error", nil)
		}
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: results})
}

func parsePlaceBias(r *http.Request) (*trips.Point, bool) {
	latRaw := strings.TrimSpace(r.URL.Query().Get("lat"))
	lngRaw := strings.TrimSpace(r.URL.Query().Get("lng"))
	if latRaw == "" && lngRaw == "" {
		return nil, true
	}
	if latRaw == "" || lngRaw == "" {
		return nil, false
	}
	lat, latErr := strconv.ParseFloat(latRaw, 64)
	lng, lngErr := strconv.ParseFloat(lngRaw, 64)
	if latErr != nil || lngErr != nil || lat < -90 || lat > 90 || lng < -180 || lng > 180 || (lat == 0 && lng == 0) {
		return nil, false
	}
	return &trips.Point{Lat: lat, Lng: lng}, true
}
