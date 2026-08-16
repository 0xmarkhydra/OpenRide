package httpserver

import "net/http"

func (s *Server) riderTripIncident(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if s.deps.Ride == nil {
		writeError(w, http.StatusServiceUnavailable, "TRIP_INCIDENT_UNAVAILABLE", "Trip support is unavailable", nil)
		return
	}
	var req incidentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	trip, err := s.deps.Ride.ReportIncidentByRider(r.PathValue("id"), riderID, req.Type, req.Note)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.publishTrip(trip, "trip.incident", trip)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}
