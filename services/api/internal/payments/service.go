package payments

import (
	"errors"
	"time"

	"flashx/services/api/internal/platform/ids"
	"flashx/services/api/internal/trips"
)

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) EnsureCash(trip trips.Trip) (Payment, error) {
	if existing, err := s.store.GetByTrip(trip.ID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return Payment{}, err
	}
	amount := trip.EstimatedFareMinor
	if trip.FinalFareMinor > 0 {
		amount = trip.FinalFareMinor
	}
	now := s.now()
	payment := Payment{
		ID:          ids.New("pay"),
		TripID:      trip.ID,
		Provider:    "cash",
		Method:      "cash",
		Status:      StatusPending,
		AmountMinor: amount,
		Currency:    trip.Currency,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.store.Create(payment); err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return s.store.GetByTrip(trip.ID)
		}
		return Payment{}, err
	}
	return payment, nil
}

func (s *Service) GetForTrip(trip trips.Trip) (Payment, error) {
	return s.store.GetByTrip(trip.ID)
}

// ReconcileCash makes the persisted cash ledger reflect the authoritative trip
// lifecycle. It is intentionally limited to customer payment state; driver
// payout/commission is a separate marketplace concern and is not inferred here.
func (s *Service) ReconcileCash(trip trips.Trip) (Payment, error) {
	payment, err := s.EnsureCash(trip)
	if err != nil {
		return Payment{}, err
	}
	switch trip.Status {
	case trips.StatusCompleted:
		return s.MarkCashCollected(trip)
	case trips.StatusCancelled:
		if payment.Status == StatusCancelled {
			return payment, nil
		}
		return s.CancelCash(trip)
	default:
		return payment, nil
	}
}

func (s *Service) MarkCashCollected(trip trips.Trip) (Payment, error) {
	if trip.Status != trips.StatusCompleted {
		return Payment{}, ErrInvalidState
	}
	payment, err := s.EnsureCash(trip)
	if err != nil {
		return Payment{}, err
	}
	if payment.Status == StatusPaid {
		return payment, nil
	}
	if payment.Provider != "cash" || payment.Method != "cash" {
		return Payment{}, ErrInvalidState
	}
	amount := trip.FinalFareMinor
	if amount <= 0 {
		amount = trip.EstimatedFareMinor
	}
	payment.AmountMinor = amount
	payment.Status = StatusPaid
	payment.UpdatedAt = s.now()
	if err := s.store.Save(payment); err != nil {
		return Payment{}, err
	}
	return payment, nil
}

func (s *Service) CancelCash(trip trips.Trip) (Payment, error) {
	payment, err := s.EnsureCash(trip)
	if err != nil {
		return Payment{}, err
	}
	if payment.Status == StatusPaid {
		return Payment{}, ErrInvalidState
	}
	payment.Status = StatusCancelled
	payment.UpdatedAt = s.now()
	if err := s.store.Save(payment); err != nil {
		return Payment{}, err
	}
	return payment, nil
}
