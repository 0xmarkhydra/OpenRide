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
		TermsSnapshot: AgreementTerms{
			TariffID:      quote.TariffID,
			TariffVersion: quote.TariffVersion,
			QuoteMetadata: cloneStringAnyMap(quote.Metadata),
		},
		CreatedAt: now.UTC(),
	}, nil
}

func cloneStringAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = cloneAny(value)
	}
	return out
}

func cloneAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneStringAnyMap(typed)
	case []any:
		out := make([]any, len(typed))
		for i := range typed {
			out[i] = cloneAny(typed[i])
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	default:
		return value
	}
}
