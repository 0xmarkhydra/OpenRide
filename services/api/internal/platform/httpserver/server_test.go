package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"flashx/services/api/internal/admin"
	"flashx/services/api/internal/auth"
	"flashx/services/api/internal/customervehicles"
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/drivers"
	"flashx/services/api/internal/payments"
	"flashx/services/api/internal/platform/idempotency"
	"flashx/services/api/internal/pricing"
	"flashx/services/api/internal/ratings"
	"flashx/services/api/internal/ride"
	"flashx/services/api/internal/trips"
	"flashx/services/api/internal/users"
)

func newTestServer() *Server {
	tripService := trips.NewService(trips.NewMemoryStore())
	driverService := drivers.NewService(drivers.NewMemoryStore())
	vehicleService := customervehicles.NewService(customervehicles.NewMemoryStore())
	userService := users.NewService(users.NewMemoryStore())
	adminService := admin.NewService(admin.NewMemoryStore())
	if _, err := adminService.Bootstrap("0900000001", "admin@flashx.test", "Test Admin"); err != nil {
		panic(err)
	}
	authService, err := auth.NewService(auth.NewMemoryStore(), auth.DevelopmentOTPSender{}, "test-secret-123456789", "test", "development")
	if err != nil {
		panic(err)
	}
	return New(":0", Dependencies{
		AppEnv:           "test",
		Persistence:      "memory",
		AllowDevIdentity: true,
		Trips:            tripService,
		Drivers:          driverService,
		CustomerVehicles: vehicleService,
		Users:            userService,
		Admin:            adminService,
		Dispatch:         dispatch.NewEngine(driverService, tripService),
		Ride:             ride.NewService(tripService, driverService),
		Pricing:          pricing.NewService(),
		Payments:         payments.NewService(payments.NewMemoryStore()),
		Ratings:          ratings.NewService(ratings.NewMemoryStore(), tripService),
		Idempotency:      idempotency.NewMemoryStore(),
		Auth:             authService,
	})
}

func TestCreateTripIsIdempotent(t *testing.T) {
	s := newTestServer()
	body := []byte(`{"pickup":{"lat":21.0285,"lng":105.8542},"destination":{"lat":21.035,"lng":105.81},"service_type":"bike"}`)

	first := perform(t, s, http.MethodPost, "/v1/trips", body, map[string]string{
		"X-Dev-Rider-ID":  "rider-1",
		"Idempotency-Key": "booking-1",
	})
	if first.Code != http.StatusCreated {
		t.Fatalf("first status = %d body=%s", first.Code, first.Body.String())
	}
	firstID := responseTripID(t, first.Body.Bytes())

	second := perform(t, s, http.MethodPost, "/v1/trips", body, map[string]string{
		"X-Dev-Rider-ID":  "rider-1",
		"Idempotency-Key": "booking-1",
	})
	if second.Code != http.StatusOK {
		t.Fatalf("second status = %d body=%s", second.Code, second.Body.String())
	}
	secondID := responseTripID(t, second.Body.Bytes())
	if firstID != secondID {
		t.Fatalf("duplicate request created a different trip: %s != %s", firstID, secondID)
	}
}

func TestCreateTripRejectsIdempotencyKeyReuseWithDifferentBody(t *testing.T) {
	s := newTestServer()
	headers := map[string]string{"X-Dev-Rider-ID": "rider-1", "Idempotency-Key": "booking-1"}
	body1 := []byte(`{"pickup":{"lat":21.0285,"lng":105.8542},"destination":{"lat":21.035,"lng":105.81},"service_type":"bike"}`)
	body2 := []byte(`{"pickup":{"lat":21.0285,"lng":105.8542},"destination":{"lat":21.045,"lng":105.82},"service_type":"bike"}`)

	if rr := perform(t, s, http.MethodPost, "/v1/trips", body1, headers); rr.Code != http.StatusCreated {
		t.Fatalf("first status = %d body=%s", rr.Code, rr.Body.String())
	}
	rr := perform(t, s, http.MethodPost, "/v1/trips", body2, headers)
	if rr.Code != http.StatusConflict {
		t.Fatalf("reuse status = %d, want 409; body=%s", rr.Code, rr.Body.String())
	}
}

func TestProductionRejectsDevelopmentIdentity(t *testing.T) {
	s := New(":0", Dependencies{
		AppEnv:      "production",
		Trips:       trips.NewService(trips.NewMemoryStore()),
		Pricing:     pricing.NewService(),
		Idempotency: idempotency.NewMemoryStore(),
	})
	rr := perform(t, s, http.MethodPost, "/v1/trips/estimate", []byte(`{"pickup":{"lat":21,"lng":105},"destination":{"lat":21.1,"lng":105.1},"service_type":"bike"}`), map[string]string{"X-Dev-Rider-ID": "rider-1"})
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rr.Code)
	}
}

func TestFullRideFlow(t *testing.T) {
	s := newTestServer()

	register := perform(t, s, http.MethodPost, "/v1/dev/drivers", []byte(`{"id":"driver-1","full_name":"Driver One","service_type":"bike"}`), nil)
	if register.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", register.Code, register.Body.String())
	}
	driverHeaders := map[string]string{"X-Dev-Driver-ID": "driver-1"}

	online := perform(t, s, http.MethodPost, "/v1/driver/availability", []byte(`{"status":"online"}`), driverHeaders)
	if online.Code != http.StatusOK {
		t.Fatalf("online status=%d body=%s", online.Code, online.Body.String())
	}
	location := perform(t, s, http.MethodPost, "/v1/driver/location", []byte(`{"lat":21.029,"lng":105.854,"accuracy_m":5}`), driverHeaders)
	if location.Code != http.StatusOK {
		t.Fatalf("location status=%d body=%s", location.Code, location.Body.String())
	}

	riderHeaders := map[string]string{"X-Dev-Rider-ID": "rider-1", "Idempotency-Key": "ride-e2e-1"}
	create := perform(t, s, http.MethodPost, "/v1/trips", []byte(`{"pickup":{"lat":21.0285,"lng":105.8542},"destination":{"lat":21.035,"lng":105.81},"service_type":"bike"}`), riderHeaders)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	tripID := responseTripID(t, create.Body.Bytes())

	current := perform(t, s, http.MethodGet, "/v1/driver/offers/current", nil, driverHeaders)
	if current.Code != http.StatusOK {
		t.Fatalf("offer status=%d body=%s", current.Code, current.Body.String())
	}
	var offerPayload struct {
		Data struct {
			Offer struct {
				ID string `json:"id"`
			} `json:"offer"`
		} `json:"data"`
	}
	if err := json.Unmarshal(current.Body.Bytes(), &offerPayload); err != nil {
		t.Fatal(err)
	}
	if offerPayload.Data.Offer.ID == "" {
		t.Fatalf("missing offer id: %s", current.Body.String())
	}
	offerID := offerPayload.Data.Offer.ID

	accept := perform(t, s, http.MethodPost, "/v1/driver/offers/"+offerID+"/accept", []byte(`{}`), driverHeaders)
	if accept.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", accept.Code, accept.Body.String())
	}

	for _, action := range []string{"arriving", "arrived", "vehicle-received", "start", "handover", "complete"} {
		rr := perform(t, s, http.MethodPost, "/v1/driver/trips/"+tripID+"/"+action, []byte(`{}`), driverHeaders)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status=%d body=%s", action, rr.Code, rr.Body.String())
		}
	}

	get := perform(t, s, http.MethodGet, "/v1/trips/"+tripID, nil, map[string]string{"X-Dev-Rider-ID": "rider-1"})
	if get.Code != http.StatusOK {
		t.Fatalf("get status=%d body=%s", get.Code, get.Body.String())
	}
	var tripPayload struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &tripPayload); err != nil {
		t.Fatal(err)
	}
	if tripPayload.Data.Status != "completed" {
		t.Fatalf("trip status=%s, want completed", tripPayload.Data.Status)
	}

	me := perform(t, s, http.MethodGet, "/v1/driver/me", nil, driverHeaders)
	if me.Code != http.StatusOK {
		t.Fatalf("driver me status=%d body=%s", me.Code, me.Body.String())
	}
	var driverPayload struct {
		Data struct {
			Availability string `json:"availability_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(me.Body.Bytes(), &driverPayload); err != nil {
		t.Fatal(err)
	}
	if driverPayload.Data.Availability != "online" {
		t.Fatalf("driver availability=%s, want online", driverPayload.Data.Availability)
	}
}

func TestRiderOTPBearerFlow(t *testing.T) {
	s := newTestServer()
	request := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0912345678","role":"rider"}`), nil)
	if request.Code != http.StatusAccepted {
		t.Fatalf("otp request status=%d body=%s", request.Code, request.Body.String())
	}
	var requested struct {
		Data struct {
			ChallengeID string `json:"challenge_id"`
			DebugCode   string `json:"debug_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request.Body.Bytes(), &requested); err != nil {
		t.Fatal(err)
	}
	if requested.Data.ChallengeID == "" || requested.Data.DebugCode == "" {
		t.Fatalf("missing development OTP data: %s", request.Body.String())
	}

	verifyBody, _ := json.Marshal(map[string]any{
		"challenge_id": requested.Data.ChallengeID,
		"role":         "rider",
		"code":         requested.Data.DebugCode,
	})
	verified := perform(t, s, http.MethodPost, "/v1/auth/otp/verify", verifyBody, nil)
	if verified.Code != http.StatusOK {
		t.Fatalf("otp verify status=%d body=%s", verified.Code, verified.Body.String())
	}
	var result struct {
		Data struct {
			Actor struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"actor"`
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	if err := json.Unmarshal(verified.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.Actor.ID == "" || result.Data.Actor.Role != "rider" || result.Data.Tokens.AccessToken == "" {
		t.Fatalf("invalid auth response: %s", verified.Body.String())
	}

	me := perform(t, s, http.MethodGet, "/v1/rider/me", nil, map[string]string{
		"Authorization": "Bearer " + result.Data.Tokens.AccessToken,
	})
	if me.Code != http.StatusOK {
		t.Fatalf("rider me status=%d body=%s", me.Code, me.Body.String())
	}
}

func TestNewDriverIsPendingAndCannotGoOnline(t *testing.T) {
	s := newTestServer()
	request := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0987654321","role":"driver"}`), nil)
	if request.Code != http.StatusAccepted {
		t.Fatalf("otp request status=%d body=%s", request.Code, request.Body.String())
	}
	var requested struct {
		Data struct {
			ChallengeID string `json:"challenge_id"`
			DebugCode   string `json:"debug_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request.Body.Bytes(), &requested); err != nil {
		t.Fatal(err)
	}
	verifyBody, _ := json.Marshal(map[string]any{
		"challenge_id": requested.Data.ChallengeID,
		"role":         "driver",
		"code":         requested.Data.DebugCode,
	})
	verified := perform(t, s, http.MethodPost, "/v1/auth/otp/verify", verifyBody, nil)
	if verified.Code != http.StatusOK {
		t.Fatalf("otp verify status=%d body=%s", verified.Code, verified.Body.String())
	}
	var result struct {
		Data struct {
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	if err := json.Unmarshal(verified.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	headers := map[string]string{"Authorization": "Bearer " + result.Data.Tokens.AccessToken}
	me := perform(t, s, http.MethodGet, "/v1/driver/me", nil, headers)
	if me.Code != http.StatusOK {
		t.Fatalf("driver me status=%d body=%s", me.Code, me.Body.String())
	}
	if !bytes.Contains(me.Body.Bytes(), []byte(`"approval_status":"pending"`)) {
		t.Fatalf("new driver must be pending: %s", me.Body.String())
	}
	online := perform(t, s, http.MethodPost, "/v1/driver/availability", []byte(`{"status":"online"}`), headers)
	if online.Code != http.StatusForbidden {
		t.Fatalf("pending driver online status=%d body=%s", online.Code, online.Body.String())
	}
}

func TestAdminOTPApprovesDriverThenDriverCanGoOnline(t *testing.T) {
	s := newTestServer()

	driverOTP := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0977000001","role":"driver"}`), nil)
	if driverOTP.Code != http.StatusAccepted {
		t.Fatalf("driver otp status=%d body=%s", driverOTP.Code, driverOTP.Body.String())
	}
	var driverChallenge struct {
		Data struct {
			ChallengeID string `json:"challenge_id"`
			DebugCode   string `json:"debug_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(driverOTP.Body.Bytes(), &driverChallenge); err != nil {
		t.Fatal(err)
	}
	driverVerifyBody, _ := json.Marshal(map[string]any{
		"challenge_id": driverChallenge.Data.ChallengeID,
		"role":         "driver",
		"code":         driverChallenge.Data.DebugCode,
	})
	driverVerified := perform(t, s, http.MethodPost, "/v1/auth/otp/verify", driverVerifyBody, nil)
	if driverVerified.Code != http.StatusOK {
		t.Fatalf("driver verify status=%d body=%s", driverVerified.Code, driverVerified.Body.String())
	}
	var driverAuth struct {
		Data struct {
			Actor struct {
				ID string `json:"id"`
			} `json:"actor"`
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	if err := json.Unmarshal(driverVerified.Body.Bytes(), &driverAuth); err != nil {
		t.Fatal(err)
	}

	blocked := perform(t, s, http.MethodPost, "/v1/driver/availability", []byte(`{"status":"online"}`), map[string]string{
		"Authorization": "Bearer " + driverAuth.Data.Tokens.AccessToken,
	})
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("pending driver should be forbidden, status=%d body=%s", blocked.Code, blocked.Body.String())
	}

	unauthorizedAdmin := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0900000002","role":"admin"}`), nil)
	if unauthorizedAdmin.Code != http.StatusForbidden {
		t.Fatalf("unknown admin phone status=%d body=%s", unauthorizedAdmin.Code, unauthorizedAdmin.Body.String())
	}

	adminOTP := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0900000001","role":"admin"}`), nil)
	if adminOTP.Code != http.StatusAccepted {
		t.Fatalf("admin otp status=%d body=%s", adminOTP.Code, adminOTP.Body.String())
	}
	var adminChallenge struct {
		Data struct {
			ChallengeID string `json:"challenge_id"`
			DebugCode   string `json:"debug_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminOTP.Body.Bytes(), &adminChallenge); err != nil {
		t.Fatal(err)
	}
	adminVerifyBody, _ := json.Marshal(map[string]any{
		"challenge_id": adminChallenge.Data.ChallengeID,
		"role":         "admin",
		"code":         adminChallenge.Data.DebugCode,
	})
	adminVerified := perform(t, s, http.MethodPost, "/v1/auth/otp/verify", adminVerifyBody, nil)
	if adminVerified.Code != http.StatusOK {
		t.Fatalf("admin verify status=%d body=%s", adminVerified.Code, adminVerified.Body.String())
	}
	var adminAuth struct {
		Data struct {
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	if err := json.Unmarshal(adminVerified.Body.Bytes(), &adminAuth); err != nil {
		t.Fatal(err)
	}

	approve := perform(t, s, http.MethodPost, "/v1/admin/drivers/"+driverAuth.Data.Actor.ID+"/approval", []byte(`{"status":"approved","reason":"documents_verified"}`), map[string]string{
		"Authorization": "Bearer " + adminAuth.Data.Tokens.AccessToken,
	})
	if approve.Code != http.StatusOK {
		t.Fatalf("approve status=%d body=%s", approve.Code, approve.Body.String())
	}

	online := perform(t, s, http.MethodPost, "/v1/driver/availability", []byte(`{"status":"online"}`), map[string]string{
		"Authorization": "Bearer " + driverAuth.Data.Tokens.AccessToken,
	})
	if online.Code != http.StatusOK {
		t.Fatalf("approved driver online status=%d body=%s", online.Code, online.Body.String())
	}
}

func TestAdminOperationsEndpoints(t *testing.T) {
	s := newTestServer()
	adminID, adminHeaders := adminSession(t, s)

	customer, _, err := s.deps.Users.FindOrCreateByPhone("0911222333")
	if err != nil {
		t.Fatal(err)
	}
	customer, err = s.deps.Users.UpdateProfile(customer.ID, "Khách Test")
	if err != nil {
		t.Fatal(err)
	}
	vehicle, err := s.deps.CustomerVehicles.Create(customervehicles.CreateInput{
		OwnerUserID:  customer.ID,
		Type:         "car",
		LicensePlate: "36A-12345",
		Brand:        "Toyota",
		Model:        "Vios",
		Transmission: "automatic",
	})
	if err != nil {
		t.Fatal(err)
	}
	driver, err := s.deps.Drivers.RegisterApproved("driver-admin-test", "Driver Admin Test", trips.ServiceDesignatedDriverCar)
	if err != nil {
		t.Fatal(err)
	}
	trip, err := s.deps.Trips.Create(trips.CreateInput{
		RiderID:            customer.ID,
		CustomerVehicleID:  vehicle.ID,
		ServiceType:        trips.ServiceDesignatedDriverCar,
		Pickup:             trips.Point{Lat: 19.807, Lng: 105.776},
		Destination:        trips.Point{Lat: 19.82, Lng: 105.79},
		EstimatedDistanceM: 5000,
		EstimatedDurationS: 900,
		FareBreakdown:      trips.FareBreakdown{TotalMinor: 180000},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.AssignDriver(trip.ID, driver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.MarkArriving(trip.ID, driver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.MarkArrived(trip.ID, driver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.MarkVehicleReceived(trip.ID, driver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.Start(trip.ID, driver.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.deps.Trips.ReportIncident(trip.ID, driver.ID, "vehicle_issue", "Test incident"); err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		path     string
		contains string
	}{
		{"/v1/admin/customers", customer.ID},
		{"/v1/admin/vehicles", vehicle.ID},
		{"/v1/admin/pricing", trips.ServiceDesignatedDriverCar},
		{"/v1/admin/system", `"persistence":"memory"`},
	}
	for _, check := range checks {
		rr := perform(t, s, http.MethodGet, check.path, nil, adminHeaders)
		if rr.Code != http.StatusOK {
			t.Fatalf("GET %s status=%d body=%s", check.path, rr.Code, rr.Body.String())
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte(check.contains)) {
			t.Fatalf("GET %s missing %q: %s", check.path, check.contains, rr.Body.String())
		}
	}

	resolved := perform(t, s, http.MethodPost, "/v1/admin/trips/"+trip.ID+"/incident/resolve", []byte(`{"note":"Đã xác minh với khách và tài xế"}`), adminHeaders)
	if resolved.Code != http.StatusOK {
		t.Fatalf("resolve incident status=%d body=%s", resolved.Code, resolved.Body.String())
	}
	if !bytes.Contains(resolved.Body.Bytes(), []byte(`"incident_open":false`)) {
		t.Fatalf("incident must be closed: %s", resolved.Body.String())
	}

	audit := perform(t, s, http.MethodGet, "/v1/admin/audit?limit=20", nil, adminHeaders)
	if audit.Code != http.StatusOK {
		t.Fatalf("audit status=%d body=%s", audit.Code, audit.Body.String())
	}
	if !bytes.Contains(audit.Body.Bytes(), []byte("trip.incident_resolved")) || !bytes.Contains(audit.Body.Bytes(), []byte(adminID)) {
		t.Fatalf("audit missing incident resolution: %s", audit.Body.String())
	}
}

func adminSession(t *testing.T, s *Server) (string, map[string]string) {
	t.Helper()
	request := perform(t, s, http.MethodPost, "/v1/auth/otp/request", []byte(`{"phone":"0900000001","role":"admin"}`), nil)
	if request.Code != http.StatusAccepted {
		t.Fatalf("admin otp status=%d body=%s", request.Code, request.Body.String())
	}
	var challenge struct {
		Data struct {
			ChallengeID string `json:"challenge_id"`
			DebugCode   string `json:"debug_code"`
		} `json:"data"`
	}
	if err := json.Unmarshal(request.Body.Bytes(), &challenge); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(map[string]any{"challenge_id": challenge.Data.ChallengeID, "role": "admin", "code": challenge.Data.DebugCode})
	verified := perform(t, s, http.MethodPost, "/v1/auth/otp/verify", body, nil)
	if verified.Code != http.StatusOK {
		t.Fatalf("admin verify status=%d body=%s", verified.Code, verified.Body.String())
	}
	var session struct {
		Data struct {
			Actor struct {
				ID string `json:"id"`
			} `json:"actor"`
			Tokens struct {
				AccessToken string `json:"access_token"`
			} `json:"tokens"`
		} `json:"data"`
	}
	if err := json.Unmarshal(verified.Body.Bytes(), &session); err != nil {
		t.Fatal(err)
	}
	return session.Data.Actor.ID, map[string]string{"Authorization": "Bearer " + session.Data.Tokens.AccessToken}
}

func perform(t *testing.T, s *Server, method, path string, body []byte, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rr, req)
	return rr
}

func responseTripID(t *testing.T, body []byte) string {
	t.Helper()
	var payload struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Data.ID == "" {
		t.Fatalf("response has no trip id: %s", body)
	}
	return payload.Data.ID
}
