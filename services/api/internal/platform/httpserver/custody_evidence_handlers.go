package httpserver

import (
	"errors"
	"net/http"

	"flashx/services/api/internal/custodyevidence"
)

func (s *Server) driverTripCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	items, err := s.deps.CustodyEvidence.ListForDriver(r.Context(), r.PathValue("id"), driverID)
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) driverUpdateCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	var req custodyevidence.EvidenceInput
	if !decodeJSON(w, r, &req) {
		return
	}
	snapshot, err := s.deps.CustodyEvidence.UpdateByDriver(r.PathValue("id"), driverID, custodyStage(r), req)
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	s.publishCustodyEvidence(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) driverPrepareCustodyPhotoUpload(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	var req custodyevidence.PhotoUploadInput
	if !decodeJSON(w, r, &req) {
		return
	}
	ticket, err := s.deps.CustodyEvidence.PreparePhotoUpload(r.Context(), r.PathValue("id"), driverID, custodyStage(r), req)
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: ticket})
}

func (s *Server) driverCompleteCustodyPhotoUpload(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	var req custodyevidence.CompletePhotoInput
	if !decodeJSON(w, r, &req) {
		return
	}
	snapshot, err := s.deps.CustodyEvidence.CompletePhotoUpload(r.PathValue("id"), driverID, custodyStage(r), req)
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	s.publishCustodyEvidence(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) driverConfirmCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	snapshot, err := s.deps.CustodyEvidence.ConfirmDriver(r.PathValue("id"), driverID, custodyStage(r))
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	s.publishCustodyEvidence(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) riderTripCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	items, err := s.deps.CustodyEvidence.ListForRider(r.Context(), r.PathValue("id"), riderID)
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) riderConfirmCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	snapshot, err := s.deps.CustodyEvidence.ConfirmRider(r.PathValue("id"), riderID, custodyStage(r))
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	s.publishCustodyEvidence(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) adminTripCustodyEvidence(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if !s.requireCustodyEvidence(w) {
		return
	}
	items, err := s.deps.CustodyEvidence.ListForAdmin(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeCustodyEvidenceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) requireCustodyEvidence(w http.ResponseWriter) bool {
	if s.deps.CustodyEvidence == nil {
		writeError(w, http.StatusServiceUnavailable, "CUSTODY_EVIDENCE_UNAVAILABLE", "Custody evidence service is unavailable", nil)
		return false
	}
	return true
}

func custodyStage(r *http.Request) custodyevidence.Stage {
	return custodyevidence.Stage(r.PathValue("stage"))
}

func (s *Server) publishCustodyEvidence(tripID string, data any) {
	trip, err := s.deps.Trips.Get(tripID)
	if err != nil {
		return
	}
	s.publishTrip(trip, "trip.custody_evidence_updated", data)
}

func (s *Server) writeCustodyEvidenceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, custodyevidence.ErrNotFound):
		writeError(w, http.StatusNotFound, "CUSTODY_EVIDENCE_NOT_FOUND", "Custody evidence was not found", nil)
	case errors.Is(err, custodyevidence.ErrForbidden):
		writeError(w, http.StatusForbidden, "CUSTODY_EVIDENCE_FORBIDDEN", "Custody evidence is not accessible by this actor", nil)
	case errors.Is(err, custodyevidence.ErrInvalidState):
		writeError(w, http.StatusConflict, "CUSTODY_EVIDENCE_INVALID_STATE", "Custody evidence cannot be changed from the current job state", nil)
	case errors.Is(err, custodyevidence.ErrNotReady):
		writeError(w, http.StatusConflict, "CUSTODY_EVIDENCE_NOT_READY", "Custody evidence requires a condition note, at least two photos and the required confirmation order", nil)
	case errors.Is(err, custodyevidence.ErrStorageUnavailable):
		writeError(w, http.StatusServiceUnavailable, "OBJECT_STORAGE_UNAVAILABLE", "Object storage is unavailable for custody evidence", nil)
	case errors.Is(err, custodyevidence.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "CUSTODY_EVIDENCE_CONFLICT", "Custody evidence photo already exists", nil)
	case errors.Is(err, custodyevidence.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "CUSTODY_EVIDENCE_INVALID", "Custody evidence data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected custody evidence error", nil)
	}
}
