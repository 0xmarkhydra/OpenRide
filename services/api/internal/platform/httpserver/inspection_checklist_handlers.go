package httpserver

import (
	"errors"
	"net/http"
	"strings"

	admindomain "flashx/services/api/internal/admin"
	"flashx/services/api/internal/inspectionchecklist"
)

func (s *Server) riderInspectionChecklist(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok || !s.requireInspectionChecklist(w) {
		return
	}
	snapshot, err := s.deps.InspectionChecklist.ForRider(r.PathValue("id"), riderID)
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) riderUpdateInspectionChecklist(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok || !s.requireInspectionChecklist(w) {
		return
	}
	var req inspectionchecklist.CustomerUpdate
	if !decodeJSON(w, r, &req) {
		return
	}
	snapshot, err := s.deps.InspectionChecklist.UpdateCustomer(r.PathValue("id"), riderID, r.PathValue("itemKey"), req)
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	s.publishInspectionChecklist(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) driverInspectionChecklist(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok || !s.requireInspectionChecklist(w) {
		return
	}
	snapshot, err := s.deps.InspectionChecklist.ForDriver(r.PathValue("id"), driverID)
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) driverUpdateInspectionChecklist(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok || !s.requireInspectionChecklist(w) {
		return
	}
	var req inspectionchecklist.DriverUpdate
	if !decodeJSON(w, r, &req) {
		return
	}
	snapshot, err := s.deps.InspectionChecklist.UpdateDriver(r.PathValue("id"), driverID, r.PathValue("itemKey"), req)
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	s.publishInspectionChecklist(r.PathValue("id"), snapshot)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) adminTripInspectionChecklist(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok || !s.requireInspectionChecklist(w) {
		return
	}
	snapshot, err := s.deps.InspectionChecklist.ForAdmin(r.PathValue("id"))
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func (s *Server) adminInspectionChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok || !s.requireInspectionChecklist(w) {
		return
	}
	template, err := s.deps.InspectionChecklist.ActiveTemplate()
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: template})
}

func (s *Server) adminReplaceInspectionChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin)
	if !ok || !s.requireInspectionChecklist(w) {
		return
	}
	var req inspectionchecklist.ReplaceTemplateInput
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	previous, _ := s.deps.InspectionChecklist.ActiveTemplate()
	template, err := s.deps.InspectionChecklist.ReplaceTemplate(req, actor.ID)
	if err != nil {
		s.writeInspectionChecklistError(w, err)
		return
	}
	if s.deps.Admin != nil {
		if err := s.deps.Admin.Audit(actor.ID, "inspection.checklist_template_updated", "inspection_checklist_template", template.ID, map[string]any{
			"previous_version": previous.Version,
			"version":          template.Version,
			"reason":           template.Reason,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Checklist template changed but audit write failed", nil)
			return
		}
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: template})
}

func (s *Server) requireInspectionChecklist(w http.ResponseWriter) bool {
	if s.deps.InspectionChecklist == nil {
		writeError(w, http.StatusServiceUnavailable, "INSPECTION_CHECKLIST_UNAVAILABLE", "Inspection checklist service is unavailable", nil)
		return false
	}
	return true
}

func (s *Server) writeInspectionChecklistError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, inspectionchecklist.ErrNotFound):
		writeError(w, http.StatusNotFound, "INSPECTION_CHECKLIST_NOT_FOUND", "Inspection checklist was not found", nil)
	case errors.Is(err, inspectionchecklist.ErrForbidden):
		writeError(w, http.StatusForbidden, "INSPECTION_CHECKLIST_FORBIDDEN", "Inspection checklist is not accessible by this actor", nil)
	case errors.Is(err, inspectionchecklist.ErrInvalidState):
		writeError(w, http.StatusConflict, "INSPECTION_CHECKLIST_INVALID_STATE", "Inspection checklist cannot be changed from the current job state", nil)
	case errors.Is(err, inspectionchecklist.ErrNotReady):
		writeError(w, http.StatusConflict, "INSPECTION_CHECKLIST_REQUIRED", "Required inspection documents are not fully verified", nil)
	case errors.Is(err, inspectionchecklist.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "INSPECTION_CHECKLIST_INVALID", "Inspection checklist data is invalid", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INSPECTION_CHECKLIST_FAILED", "Unable to process inspection checklist", nil)
	}
}

func (s *Server) publishInspectionChecklist(tripID string, snapshot inspectionchecklist.Snapshot) {
	trip, err := s.deps.Trips.Get(tripID)
	if err != nil {
		return
	}
	s.publishTrip(trip, "trip.inspection_checklist_updated", map[string]any{"trip_id": tripID, "checklist": snapshot})
}
