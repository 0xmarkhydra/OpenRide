package httpserver

import (
	"errors"
	"net/http"
	"strings"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/driverdocs"
)

func (s *Server) prepareDriverDocumentUpload(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "OBJECT_STORAGE_UNAVAILABLE", "Object storage is unavailable", nil)
		return
	}
	var req driverdocs.UploadInput
	if !decodeJSON(w, r, &req) {
		return
	}
	ticket, err := s.deps.DriverDocuments.PrepareUpload(r.Context(), driverID, req)
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: ticket})
}

func (s *Server) completeDriverDocumentUpload(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "OBJECT_STORAGE_UNAVAILABLE", "Object storage is unavailable", nil)
		return
	}
	var req driverdocs.CompleteInput
	if !decodeJSON(w, r, &req) {
		return
	}
	doc, err := s.deps.DriverDocuments.CompleteUpload(driverID, req)
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: doc})
}

func (s *Server) listDriverDocuments(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "DOCUMENTS_UNAVAILABLE", "Driver documents are unavailable", nil)
		return
	}
	docs, err := s.deps.DriverDocuments.ListForDriver(driverID)
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: docs, Meta: map[string]any{"count": len(docs)}})
}

func (s *Server) viewDriverDocument(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	if s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "OBJECT_STORAGE_UNAVAILABLE", "Object storage is unavailable", nil)
		return
	}
	item, err := s.deps.DriverDocuments.ViewForDriver(r.Context(), driverID, r.PathValue("id"))
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: item})
}

func (s *Server) adminDriverDocuments(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return
	}
	if s.deps.Admin == nil || s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "DOCUMENTS_UNAVAILABLE", "Driver documents are unavailable", nil)
		return
	}
	if _, err := s.deps.Admin.Get(adminID); err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return
	}
	items, err := s.deps.DriverDocuments.ListForAdmin(r.Context(), r.PathValue("id"))
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

type reviewDriverDocumentRequest struct {
	Status driverdocs.ReviewStatus `json:"status"`
	Note   string                  `json:"note"`
}

func (s *Server) adminReviewDriverDocument(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return
	}
	if s.deps.Admin == nil || s.deps.DriverDocuments == nil {
		writeError(w, http.StatusServiceUnavailable, "DOCUMENTS_UNAVAILABLE", "Driver documents are unavailable", nil)
		return
	}
	if _, err := s.deps.Admin.Get(adminID); err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return
	}
	var req reviewDriverDocumentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	doc, err := s.deps.DriverDocuments.Review(r.PathValue("id"), r.PathValue("documentID"), req.Status, req.Note)
	if err != nil {
		s.writeDriverDocumentError(w, err)
		return
	}
	if err := s.deps.Admin.Audit(adminID, "driver.document_reviewed", "driver_document", doc.ID, map[string]any{
		"driver_id":     doc.DriverID,
		"document_type": doc.DocumentType,
		"status":        doc.ReviewStatus,
		"note":          strings.TrimSpace(req.Note),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Document state changed but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: doc})
}

func (s *Server) writeDriverDocumentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, driverdocs.ErrNotFound):
		writeError(w, http.StatusNotFound, "DRIVER_DOCUMENT_NOT_FOUND", "Driver document was not found", nil)
	case errors.Is(err, driverdocs.ErrAlreadyExists):
		writeError(w, http.StatusConflict, "DRIVER_DOCUMENT_EXISTS", "Driver document already exists", nil)
	case errors.Is(err, driverdocs.ErrForbidden):
		writeError(w, http.StatusForbidden, "DRIVER_DOCUMENT_FORBIDDEN", "Driver document is not accessible by this actor", nil)
	case errors.Is(err, driverdocs.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "DRIVER_DOCUMENT_INVALID", "Driver document data is invalid", nil)
	case errors.Is(err, driverdocs.ErrStorageUnavailable):
		writeError(w, http.StatusServiceUnavailable, "OBJECT_STORAGE_UNAVAILABLE", "Object storage signing is unavailable", nil)
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unexpected driver document error", nil)
	}
}
