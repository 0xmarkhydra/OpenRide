package httpserver

import (
	"flashx/services/api/internal/dispatch"
	"flashx/services/api/internal/realtime"
	"flashx/services/api/internal/trips"
)

func (s *Server) publishActor(actorID, eventType, tripID string, data any) {
	if s.deps.Realtime == nil || actorID == "" {
		return
	}
	s.deps.Realtime.Publish(actorID, realtime.Event{Type: eventType, TripID: tripID, Data: data})
}

func (s *Server) publishTrip(trip trips.Trip, eventType string, data any) {
	s.publishActor(trip.RiderID, eventType, trip.ID, data)
	if trip.DriverID != "" {
		s.publishActor(trip.DriverID, eventType, trip.ID, data)
	}
}

func (s *Server) publishOffers(offers []dispatch.Offer) {
	for _, offer := range offers {
		trip, err := s.deps.Trips.Get(offer.TripID)
		if err != nil {
			continue
		}
		payload := map[string]any{"offer": offer, "trip": trip}
		if s.deps.CustomerVehicles != nil && trip.CustomerVehicleID != "" {
			if vehicle, vehicleErr := s.deps.CustomerVehicles.Get(trip.CustomerVehicleID); vehicleErr == nil {
				payload["vehicle"] = vehicle
			}
		}
		s.publishActor(offer.DriverID, "dispatch.offer", offer.TripID, payload)
	}
}
