package httpserver

import (
	"net/http"

	"flashx/services/api/internal/routing"
	"flashx/services/api/internal/trips"
)

type liveRouteSnapshot struct {
	Phase              string          `json:"phase"`
	Primary            *routing.Result `json:"primary,omitempty"`
	Service            routing.Result  `json:"service"`
	FromDriverLocation bool            `json:"from_driver_location"`
}

func (s *Server) riderTripRoutes(w http.ResponseWriter, r *http.Request) {
	riderID, ok := s.riderID(w, r)
	if !ok {
		return
	}
	trip, err := s.deps.Trips.GetForRider(r.PathValue("id"), riderID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.writeTripRoutes(w, trip)
}

func (s *Server) driverTripRoutes(w http.ResponseWriter, r *http.Request) {
	driverID, ok := s.driverID(w, r)
	if !ok {
		return
	}
	trip, err := s.deps.Trips.GetForDriver(r.PathValue("id"), driverID)
	if err != nil {
		s.writeDomainError(w, err)
		return
	}
	s.writeTripRoutes(w, trip)
}

func (s *Server) writeTripRoutes(w http.ResponseWriter, trip trips.Trip) {
	if s.deps.Pricing == nil {
		writeError(w, http.StatusServiceUnavailable, "ROUTING_UNAVAILABLE", "Routing service is unavailable", nil)
		return
	}

	serviceRoute, err := s.deps.Pricing.Route(trip.Pickup, trip.Destination, trip.ServiceType)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "ROUTING_UNAVAILABLE", "Unable to calculate the service route", nil)
		return
	}

	phase, target := liveRoutePhase(trip)
	snapshot := liveRouteSnapshot{Phase: phase, Service: serviceRoute}

	origin := trip.Pickup
	if trip.DriverID != "" && s.deps.Drivers != nil {
		if driver, driverErr := s.deps.Drivers.Get(trip.DriverID); driverErr == nil && driver.Location != nil {
			origin = trips.Point{Lat: driver.Location.Lat, Lng: driver.Location.Lng}
			snapshot.FromDriverLocation = true
		}
	}

	// For the approach phase, a stale/missing driver location is less useful than
	// returning no primary route. The app can keep showing the service preview
	// until the first fresh driver location arrives.
	if phase != "approach" || snapshot.FromDriverLocation {
		if primary, primaryErr := s.deps.Pricing.Route(origin, target, trip.ServiceType); primaryErr == nil {
			snapshot.Primary = &primary
		}
	}

	writeJSON(w, http.StatusOK, dataEnvelope{Data: snapshot})
}

func liveRoutePhase(trip trips.Trip) (string, trips.Point) {
	switch trip.Status {
	case trips.StatusAccepted, trips.StatusArrivingForPickup, trips.StatusArrivedForPickup:
		return "approach", trip.Pickup
	}
	if trips.IsInspectionService(trip.ServiceType) {
		switch trip.Status {
		case trips.StatusReturningVehicle, trips.StatusArrivedForReturn, trips.StatusHandover:
			return "return", trip.Pickup
		}
	}
	return "service", trip.Destination
}
