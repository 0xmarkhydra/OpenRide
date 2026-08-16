package httpserver

import (
	"context"
	"time"
)

// TickScheduledDispatch advances due scheduled jobs independently from driver
// heartbeats, then reuses the normal dispatch engine. Keeping this as a small
// one-shot method makes the scheduling behavior deterministic and testable.
func (s *Server) TickScheduledDispatch() error {
	if s.deps.Trips == nil {
		return nil
	}
	activated, err := s.deps.Trips.ActivateDueScheduled(100)
	if err != nil {
		return err
	}
	for _, trip := range activated {
		s.publishTrip(trip, "trip."+string(trip.Status), trip)
	}
	if s.deps.Dispatch != nil {
		s.publishOffers(s.deps.Dispatch.DispatchWaiting(100))
	}
	return nil
}

// RunScheduledDispatch keeps scheduled bookings moving even when no driver is
// currently sending availability/location heartbeats. It stops with the same
// application context used for graceful shutdown.
func (s *Server) RunScheduledDispatch(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// Run once at startup so jobs that became due while the API was restarting
	// do not need to wait for the first ticker interval.
	_ = s.TickScheduledDispatch()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.TickScheduledDispatch()
		}
	}
}
