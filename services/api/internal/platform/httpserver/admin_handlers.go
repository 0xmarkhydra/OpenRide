package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/pricing"
	"flashx/services/api/internal/trips"
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

func (s *Server) activeAdminID(w http.ResponseWriter, r *http.Request) (string, bool) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return "", false
	}
	if s.deps.Admin == nil {
		writeError(w, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "Admin service is unavailable", nil)
		return "", false
	}
	if _, err := s.deps.Admin.Get(adminID); err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return "", false
	}
	return adminID, true
}

type driverApprovalRequest struct {
	Status drivers.ApprovalStatus `json:"status"`
	Reason string                 `json:"reason"`
}

func (s *Server) adminTrips(w http.ResponseWriter, r *http.Request) {
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
	items, err := s.deps.Trips.ListAll(100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list trips", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminCustomers(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if s.deps.Users == nil {
		writeError(w, http.StatusServiceUnavailable, "CUSTOMERS_UNAVAILABLE", "Customer service is unavailable", nil)
		return
	}
	items, err := s.deps.Users.ListAll(300)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list customers", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminVehicles(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if s.deps.CustomerVehicles == nil {
		writeError(w, http.StatusServiceUnavailable, "VEHICLES_UNAVAILABLE", "Customer vehicle service is unavailable", nil)
		return
	}
	items, err := s.deps.CustomerVehicles.ListAll(300)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list customer vehicles", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminPricing(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if s.deps.Pricing == nil {
		writeError(w, http.StatusServiceUnavailable, "PRICING_UNAVAILABLE", "Pricing service is unavailable", nil)
		return
	}
	items := s.deps.Pricing.Rules()
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

type adminPricingUpdateRequest struct {
	BaseFareMinor int64 `json:"base_fare_minor"`
	PerKMMinor    int64 `json:"per_km_minor"`
	ServiceMinor  int64 `json:"service_minor"`
	MinimumMinor  int64 `json:"minimum_minor"`
}

func (s *Server) adminUpdatePricing(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.activeAdminID(w, r)
	if !ok {
		return
	}
	if s.deps.Pricing == nil {
		writeError(w, http.StatusServiceUnavailable, "PRICING_UNAVAILABLE", "Pricing service is unavailable", nil)
		return
	}
	var req adminPricingUpdateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	serviceType := trips.NormalizeServiceType(strings.TrimSpace(r.PathValue("serviceType")))
	before := s.deps.Pricing.Rules()
	updated, err := s.deps.Pricing.UpdateRule(pricing.UpdateRuleInput{
		ServiceType: serviceType, BaseFareMinor: req.BaseFareMinor, PerKMMinor: req.PerKMMinor,
		ServiceMinor: req.ServiceMinor, MinimumMinor: req.MinimumMinor, ActorID: adminID,
	})
	if err != nil {
		switch {
		case errors.Is(err, pricing.ErrUnsupportedServiceType):
			writeError(w, http.StatusUnprocessableEntity, "SERVICE_TYPE_UNSUPPORTED", "Service type is not supported", nil)
		case errors.Is(err, pricing.ErrInvalidRule):
			writeError(w, http.StatusUnprocessableEntity, "PRICING_INVALID", "Pricing values are invalid", nil)
		default:
			writeError(w, http.StatusInternalServerError, "PRICING_UPDATE_FAILED", "Unable to update pricing", nil)
		}
		return
	}
	var previous any
	for _, item := range before {
		if item.ServiceType == serviceType {
			previous = item
			break
		}
	}
	if err := s.deps.Admin.Audit(adminID, "pricing.rule_updated", "pricing_rule", serviceType, map[string]any{
		"previous": previous,
		"current":  updated,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Pricing changed but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: updated})
}

func (s *Server) adminAudit(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.deps.Admin.ListAudit(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list audit logs", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminSystem(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: map[string]any{
		"app_env":            s.deps.AppEnv,
		"persistence":        s.deps.Persistence,
		"allow_dev_identity": s.deps.AllowDevIdentity,
		"supported_services": []string{trips.ServiceDesignatedDriverCar, trips.ServiceDesignatedDriverBike, trips.ServiceVehicleInspection},
		"features": map[string]bool{
			"realtime":         s.deps.Realtime != nil,
			"driver_documents": s.deps.DriverDocuments != nil,
			"payments":         s.deps.Payments != nil,
			"ratings":          s.deps.Ratings != nil,
		},
	}})
}

type resolveIncidentRequest struct {
	Note string `json:"note"`
}

type adminAssignDriverRequest struct {
	DriverID string `json:"driver_id"`
	Reason   string `json:"reason"`
}

func (s *Server) adminTripDriverCandidates(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.activeAdminID(w, r); !ok {
		return
	}
	if s.deps.Dispatch == nil {
		writeError(w, http.StatusServiceUnavailable, "DISPATCH_UNAVAILABLE", "Dispatch service is unavailable", nil)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.deps.Dispatch.CandidatesForTrip(r.PathValue("id"), limit)
	if err != nil {
		if err == trips.ErrInvalidState {
			s.writeDomainError(w, err)
			return
		}
		s.writeDispatchError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminAssignTripDriver(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.activeAdminID(w, r)
	if !ok {
		return
	}
	if s.deps.Dispatch == nil {
		writeError(w, http.StatusServiceUnavailable, "DISPATCH_UNAVAILABLE", "Dispatch service is unavailable", nil)
		return
	}
	var req adminAssignDriverRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.DriverID = strings.TrimSpace(req.DriverID)
	req.Reason = strings.TrimSpace(req.Reason)
	if req.DriverID == "" {
		writeError(w, http.StatusUnprocessableEntity, "DRIVER_REQUIRED", "Driver is required", nil)
		return
	}
	before, err := s.deps.Trips.Get(r.PathValue("id"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	isReassign := before.DriverID != ""
	if isReassign && req.Reason == "" {
		writeError(w, http.StatusUnprocessableEntity, "REASSIGN_REASON_REQUIRED", "A reason is required when replacing a driver", nil)
		return
	}
	updated, err := s.deps.Dispatch.ManualAssign(before.ID, req.DriverID)
	if err != nil {
		if err == trips.ErrInvalidState {
			s.writeDomainError(w, err)
			return
		}
		s.writeDispatchError(w, err)
		return
	}
	action := "trip.driver_assigned"
	eventType := "trip.assigned"
	if isReassign {
		action = "trip.driver_reassigned"
		eventType = "trip.reassigned"
	}
	if err := s.deps.Admin.Audit(adminID, action, "trip", updated.ID, map[string]any{
		"previous_driver_id": before.DriverID,
		"driver_id":          updated.DriverID,
		"previous_status":    before.Status,
		"reason":             req.Reason,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Driver assignment changed but audit write failed", nil)
		return
	}
	if isReassign && before.DriverID != updated.DriverID {
		s.publishActor(before.DriverID, "trip.unassigned", updated.ID, map[string]any{"trip_id": updated.ID, "reason": req.Reason})
	}
	s.publishTrip(updated, eventType, updated)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: updated})
}

func (s *Server) adminResolveIncident(w http.ResponseWriter, r *http.Request) {
	adminID, ok := s.activeAdminID(w, r)
	if !ok {
		return
	}
	var req resolveIncidentRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	trip, err := s.deps.Trips.ResolveIncident(r.PathValue("id"))
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	if err := s.deps.Admin.Audit(adminID, "trip.incident_resolved", "trip", trip.ID, map[string]any{
		"incident_type": trip.IncidentType,
		"note":          strings.TrimSpace(req.Note),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Incident was resolved but audit write failed", nil)
		return
	}
	s.publishTrip(trip, "trip.incident_resolved", trip)
	writeJSON(w, http.StatusOK, dataEnvelope{Data: trip})
}

func (s *Server) adminDashboard(w http.ResponseWriter, r *http.Request) {
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
	driversList, err := s.deps.Drivers.ListAll()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load dashboard", nil)
		return
	}
	tripList, err := s.deps.Trips.ListAll(500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to load dashboard", nil)
		return
	}
	metrics := map[string]int{
		"drivers_total": len(driversList), "drivers_online": 0, "drivers_pending": 0,
		"trips_total": len(tripList), "trips_scheduled": 0, "trips_searching": 0,
		"trips_active": 0, "trips_completed": 0, "trips_cancelled": 0, "trips_incident": 0,
		"service_designated_car": 0, "service_designated_bike": 0, "service_inspection": 0,
	}
	for _, d := range driversList {
		if d.Availability == drivers.AvailabilityOnline {
			metrics["drivers_online"]++
		}
		if d.Approval == drivers.ApprovalPending {
			metrics["drivers_pending"]++
		}
	}
	for _, t := range tripList {
		if t.IncidentOpen {
			metrics["trips_incident"]++
		}
		switch trips.NormalizeServiceType(t.ServiceType) {
		case trips.ServiceDesignatedDriverCar:
			metrics["service_designated_car"]++
		case trips.ServiceDesignatedDriverBike:
			metrics["service_designated_bike"]++
		case trips.ServiceVehicleInspection:
			metrics["service_inspection"]++
		}
		switch t.Status {
		case trips.StatusScheduled:
			metrics["trips_scheduled"]++
		case trips.StatusSearching:
			metrics["trips_searching"]++
		case trips.StatusCompleted:
			metrics["trips_completed"]++
		case trips.StatusCancelled:
			metrics["trips_cancelled"]++
		default:
			if trips.IsDriverOccupiedStatus(t.Status) {
				metrics["trips_active"]++
			}
		}
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: metrics})
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
