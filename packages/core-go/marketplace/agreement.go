package marketplace

import (
	"errors"
	"time"
)

var ErrAgreementMismatch = errors.New("marketplace: request and quote cannot form an agreement")

// NewAgreement snapshots the commercial terms accepted by both sides.
// After this point, runtime pricing changes must not mutate the agreement.
func NewAgreement(id string, request Request, quote Quote, now time.Time) (Agreement, error) {
	if id == "" || request.Validate() != nil || quote.Validate() != nil {
		return Agreement{}, ErrAgreementMismatch
	}
	if request.ID != quote.RequestID || request.Status == RequestCancelled || request.Status == RequestExpired || !quote.IsSelectable(now) {
		return Agreement{}, ErrAgreementMismatch
	}
	return Agreement{
		ID:              id,
		InstanceID:      request.InstanceID,
		RequestID:       request.ID,
		QuoteID:         quote.ID,
		RiderID:         request.RiderID,
		DriverID:        quote.DriverID,
		DriverVehicleID: quote.DriverVehicleID,
		ServiceType:     request.ServiceType,
		Fare:            quote.Fare,
		TermsSnapshot: map[string]any{
			"tariff_id":      quote.TariffID,
			"tariff_version": quote.TariffVersion,
			"quote_metadata": quote.Metadata,
		},
		CreatedAt: now.UTC(),
	}, nil
}
