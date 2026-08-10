package httpserver

import (
	"net/http"
	"strings"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/drivers"
)

func (s *Server) adminMe(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return
	}
	if s.deps.Admin == nil {
		writeError(w, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "Admin service is unavailable", nil)
		return
	}
	user, err := s.deps.Admin.Get(adminID)
	if err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: user})
}

func (s *Server) adminDrivers(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return
	}
	if s.deps.Admin == nil {
		writeError(w, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "Admin service is unavailable", nil)
		return
	}
	if _, err := s.deps.Admin.Get(adminID); err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return
	}
	items, err := s.deps.Drivers.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list drivers", nil)
		return
	}
	filter := strings.TrimSpace(r.URL.Query().Get("approval"))
	if filter != "" {
		filtered := make([]drivers.Driver, 0)
		for _, driver := range items {
			if string(driver.Approval) == filter {
				filtered = append(filtered, driver)
			}
		}
		items = filtered
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

type driverApprovalRequest struct {
	Status drivers.ApprovalStatus `json:"status"`
	Reason string                 `json:"reason"`
}

func (s *Server) adminTrips(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok { return }
	if s.deps.Admin == nil { writeError(w,http.StatusServiceUnavailable,"ADMIN_UNAVAILABLE","Admin service is unavailable",nil); return }
	if _, err := s.deps.Admin.Get(adminID); err != nil { writeError(w,http.StatusForbidden,"ADMIN_FORBIDDEN","Admin account is not active",nil); return }
	items, err := s.deps.Trips.ListAll(100)
	if err != nil { writeError(w,http.StatusInternalServerError,"INTERNAL_ERROR","Unable to list trips",nil); return }
	writeJSON(w,http.StatusOK,dataEnvelope{Data:items,Meta:map[string]any{"count":len(items)}})
}

func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok { return }
	if s.deps.Admin == nil { writeError(w,http.StatusServiceUnavailable,"ADMIN_UNAVAILABLE","Admin service is unavailable",nil); return }
	if _, err := s.deps.Admin.Get(adminID); err != nil { writeError(w,http.StatusForbidden,"ADMIN_FORBIDDEN","Admin account is not active",nil); return }
	driversList, err := s.deps.Drivers.ListAll()
	if err != nil { writeError(w,http.StatusInternalServerError,"INTERNAL_ERROR","Unable to load dashboard",nil); return }
	tripList, err := s.deps.Trips.ListAll(500)
	if err != nil { writeError(w,http.StatusInternalServerError,"INTERNAL_ERROR","Unable to load dashboard",nil); return }
	metrics := map[string]int{"drivers_total":len(driversList),"drivers_online":0,"drivers_pending":0,"trips_total":len(tripList),"trips_searching":0,"trips_active":0,"trips_completed":0,"trips_cancelled":0}
	for _,d:=range driversList{if d.Availability==drivers.AvailabilityOnline{metrics["drivers_online"]++};if d.Approval==drivers.ApprovalPending{metrics["drivers_pending"]++}}
	for _,t:=range tripList{switch t.Status{case "searching":metrics["trips_searching"]++;case "accepted","arriving","arrived","in_progress":metrics["trips_active"]++;case "completed":metrics["trips_completed"]++;case "cancelled":metrics["trips_cancelled"]++}}
	writeJSON(w,http.StatusOK,dataEnvelope{Data:metrics})
}

func (s *Server) adminDriverApproval(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return
	}
	if s.deps.Admin == nil {
		writeError(w, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "Admin service is unavailable", nil)
		return
	}
	if _, err := s.deps.Admin.Get(adminID); err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return
	}
	var req driverApprovalRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Status != drivers.ApprovalApproved && req.Status != drivers.ApprovalRejected && req.Status != drivers.ApprovalSuspended {
		writeError(w, http.StatusUnprocessableEntity, "DRIVER_APPROVAL_INVALID", "Approval status must be approved, rejected, or suspended", nil)
		return
	}
	if (req.Status == drivers.ApprovalRejected || req.Status == drivers.ApprovalSuspended) && strings.TrimSpace(req.Reason) == "" {
		writeError(w, http.StatusUnprocessableEntity, "APPROVAL_REASON_REQUIRED", "A reason is required for reject or suspend", nil)
		return
	}

	driver, err := s.deps.Drivers.SetApproval(r.PathValue("id"), req.Status)
	if err != nil {
		s.writeDriverError(w, err)
		return
	}
	if err := s.deps.Admin.Audit(adminID, "driver.approval_changed", "driver", driver.ID, map[string]any{
		"status": req.Status,
		"reason": strings.TrimSpace(req.Reason),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Driver state changed but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: driver})
}
