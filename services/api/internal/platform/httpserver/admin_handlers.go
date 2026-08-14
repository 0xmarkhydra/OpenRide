package httpserver

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	admindomain "flashx/services/api/internal/admin"
	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/operationalsettings"
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

type adminAccountCreateRequest struct {
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

type adminAccountUpdateRequest struct {
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

func (s *Server) adminAccounts(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin); !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := s.deps.Admin.ListAccounts(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ADMIN_ACCOUNTS_FAILED", "Unable to list admin accounts", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: items, Meta: map[string]any{"count": len(items)}})
}

func (s *Server) adminCreateAccount(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin)
	if !ok {
		return
	}
	var req adminAccountCreateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	created, err := s.deps.Admin.CreateAccount(admindomain.CreateAccountInput{
		Phone: req.Phone, DisplayName: req.DisplayName, Role: strings.TrimSpace(req.Role),
	})
	if err != nil {
		s.writeAdminAccountError(w, err)
		return
	}
	if err := s.deps.Admin.Audit(actor.ID, "admin.account_created", "admin_user", created.ID, map[string]any{
		"phone": created.Phone, "role": created.Role, "status": created.Status,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Admin account created but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusCreated, dataEnvelope{Data: created})
}

func (s *Server) adminUpdateAccount(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin)
	if !ok {
		return
	}
	var req adminAccountUpdateRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	before, err := s.deps.Admin.GetAccount(r.PathValue("id"))
	if err != nil {
		s.writeAdminAccountError(w, err)
		return
	}
	req.Role = strings.TrimSpace(req.Role)
	req.Status = strings.TrimSpace(req.Status)
	req.Reason = strings.TrimSpace(req.Reason)
	if (req.Role != before.Role || req.Status != before.Status) && req.Reason == "" {
		writeError(w, http.StatusUnprocessableEntity, "ADMIN_CHANGE_REASON_REQUIRED", "A reason is required when changing role or status", nil)
		return
	}
	updated, err := s.deps.Admin.UpdateAccount(before.ID, admindomain.UpdateAccountInput{
		DisplayName: req.DisplayName, Role: req.Role, Status: req.Status,
	})
	if err != nil {
		s.writeAdminAccountError(w, err)
		return
	}
	if err := s.deps.Admin.Audit(actor.ID, "admin.account_updated", "admin_user", updated.ID, map[string]any{
		"previous_role": before.Role, "previous_status": before.Status,
		"role": updated.Role, "status": updated.Status, "reason": req.Reason,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Admin account changed but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: updated})
}

func (s *Server) writeAdminAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, admindomain.ErrInvalidInput):
		writeError(w, http.StatusUnprocessableEntity, "ADMIN_ACCOUNT_INVALID", "Admin account input is invalid", nil)
	case errors.Is(err, admindomain.ErrConflict):
		writeError(w, http.StatusConflict, "ADMIN_ACCOUNT_CONFLICT", "Phone or email already belongs to another admin", nil)
	case errors.Is(err, admindomain.ErrLastSuperAdmin):
		writeError(w, http.StatusConflict, "ADMIN_LAST_SUPER_ADMIN", "At least one active Super Admin must remain", nil)
	case errors.Is(err, admindomain.ErrNotFound):
		writeError(w, http.StatusNotFound, "ADMIN_ACCOUNT_NOT_FOUND", "Admin account was not found", nil)
	default:
		writeError(w, http.StatusInternalServerError, "ADMIN_ACCOUNT_FAILED", "Unable to update admin account", nil)
	}
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

func (s *Server) activeAdmin(w http.ResponseWriter, r *http.Request) (admindomain.User, bool) {
	adminID, ok := s.actorID(w, r, auth.RoleAdmin, "X-Dev-Admin-ID")
	if !ok {
		return admindomain.User{}, false
	}
	if s.deps.Admin == nil {
		writeError(w, http.StatusServiceUnavailable, "ADMIN_UNAVAILABLE", "Admin service is unavailable", nil)
		return admindomain.User{}, false
	}
	user, err := s.deps.Admin.Get(adminID)
	if err != nil {
		writeError(w, http.StatusForbidden, "ADMIN_FORBIDDEN", "Admin account is not active", nil)
		return admindomain.User{}, false
	}
	return user, true
}

func (s *Server) activeAdminID(w http.ResponseWriter, r *http.Request) (string, bool) {
	user, ok := s.activeAdmin(w, r)
	if !ok {
		return "", false
	}
	return user.ID, true
}

func (s *Server) requireAdminRole(w http.ResponseWriter, r *http.Request, roles ...string) (admindomain.User, bool) {
	user, ok := s.activeAdmin(w, r)
	if !ok {
		return admindomain.User{}, false
	}
	for _, role := range roles {
		if user.Role == role {
			return user, true
		}
	}
	writeError(w, http.StatusForbidden, "ADMIN_ROLE_FORBIDDEN", "Admin role cannot perform this action", nil)
	return admindomain.User{}, false
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

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	serviceType := trips.NormalizeServiceType(strings.TrimSpace(r.URL.Query().Get("service")))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	incident := strings.TrimSpace(r.URL.Query().Get("incident"))
	query := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	sortOrder := strings.TrimSpace(r.URL.Query().Get("sort"))

	// Read at most the latest 500 jobs, then apply admin-facing filters. This
	// bounds response cost today while keeping the endpoint compatible with a
	// future cursor-backed implementation once operations volume outgrows MVP.
	items, err := s.deps.Trips.ListAll(500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Unable to list trips", nil)
		return
	}
	filtered := make([]trips.Trip, 0, len(items))
	for _, trip := range items {
		if serviceType != "" && trip.ServiceType != serviceType {
			continue
		}
		if status != "" && string(trip.Status) != status {
			continue
		}
		if incident == "open" && !trip.IncidentOpen {
			continue
		}
		if incident == "closed" && trip.IncidentOpen {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(strings.Join([]string{
				trip.ID, trip.RiderID, trip.DriverID, trip.CustomerVehicleID,
				trip.ServiceType, string(trip.Status), trip.InspectionResult,
				trip.IncidentType, trip.IncidentNote,
			}, " "))
			if !strings.Contains(haystack, query) {
				continue
			}
		}
		filtered = append(filtered, trip)
	}
	if sortOrder == "oldest" {
		for left, right := 0, len(filtered)-1; left < right; left, right = left+1, right-1 {
			filtered[left], filtered[right] = filtered[right], filtered[left]
		}
	}
	total := len(filtered)
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: filtered, Meta: map[string]any{
		"count": len(filtered), "total": total, "limit": limit,
	}})
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
	adminUser, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin)
	if !ok {
		return
	}
	adminID := adminUser.ID
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
	data := map[string]any{
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
	}
	if s.deps.OperationalSettings != nil {
		data["operational_settings"] = s.deps.OperationalSettings.Get()
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: data})
}

type adminOperationalSettingsRequest struct {
	DesignatedDriverCarEnabled  bool   `json:"designated_driver_car_enabled"`
	DesignatedDriverBikeEnabled bool   `json:"designated_driver_bike_enabled"`
	VehicleInspectionEnabled    bool   `json:"vehicle_inspection_assist_enabled"`
	DispatchMaxDistanceM        int64  `json:"dispatch_max_distance_m"`
	DriverLocationMaxAgeSeconds int    `json:"driver_location_max_age_seconds"`
	Reason                      string `json:"reason"`
}

func (s *Server) adminUpdateOperationalSettings(w http.ResponseWriter, r *http.Request) {
	actor, ok := s.requireAdminRole(w, r, admindomain.RoleSuperAdmin)
	if !ok {
		return
	}
	if s.deps.OperationalSettings == nil {
		writeError(w, http.StatusServiceUnavailable, "OPERATIONAL_SETTINGS_UNAVAILABLE", "Operational settings are unavailable", nil)
		return
	}
	var req adminOperationalSettingsRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		writeError(w, http.StatusUnprocessableEntity, "SETTINGS_REASON_REQUIRED", "A reason is required when changing operational settings", nil)
		return
	}
	before := s.deps.OperationalSettings.Get()
	updated, err := s.deps.OperationalSettings.Update(operationalsettings.Config{
		DesignatedDriverCarEnabled:  req.DesignatedDriverCarEnabled,
		DesignatedDriverBikeEnabled: req.DesignatedDriverBikeEnabled,
		VehicleInspectionEnabled:    req.VehicleInspectionEnabled,
		DispatchMaxDistanceM:        req.DispatchMaxDistanceM,
		DriverLocationMaxAgeSeconds: req.DriverLocationMaxAgeSeconds,
	}, actor.ID)
	if err != nil {
		if errors.Is(err, operationalsettings.ErrInvalidConfig) {
			writeError(w, http.StatusUnprocessableEntity, "OPERATIONAL_SETTINGS_INVALID", "Operational settings are invalid", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "OPERATIONAL_SETTINGS_UPDATE_FAILED", "Unable to update operational settings", nil)
		return
	}
	if err := s.deps.Admin.Audit(actor.ID, "system.operational_settings_updated", "operational_settings", "default", map[string]any{
		"previous": before,
		"current":  updated,
		"reason":   req.Reason,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "AUDIT_FAILED", "Operational settings changed but audit write failed", nil)
		return
	}
	writeJSON(w, http.StatusOK, dataEnvelope{Data: updated})
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
