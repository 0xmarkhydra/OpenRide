package httpapi

import (
	"context"
	"errors"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
	marketplacestore "github.com/0xmarkhydra/OpenRide/services/marketplace/internal/store"
)

type durableIdempotentStore interface {
	CreateRequestIdempotent(context.Context, string, string, string, marketplace.Request) (marketplace.Request, bool, error)
	CreateTariffIdempotent(context.Context, string, string, string, marketplace.DriverTariff) (marketplace.DriverTariff, bool, error)
	SubmitQuoteIdempotent(context.Context, string, string, string, marketplace.Quote) (marketplace.Quote, bool, error)
	CancelRequestIdempotent(context.Context, string, string, string, string) (bool, error)
	WithdrawQuoteIdempotent(context.Context, string, string, string, string) (bool, error)
}

type idempotentV2Adapter struct {
	base V2Store
	idem durableIdempotentStore
}

func adaptV2Store(base V2Store) V2Store {
	if base == nil { return nil }
	idem, ok := base.(durableIdempotentStore)
	if !ok { return base }
	return &idempotentV2Adapter{base: base, idem: idem}
}

func commandIdempotency(ctx context.Context) (app.IdempotencyContext, error) {
	value, ok := app.IdempotencyFromContext(ctx)
	if !ok { return app.IdempotencyContext{}, errors.New("marketplace http: missing idempotency context") }
	return value, nil
}

func (a *idempotentV2Adapter) CreateTariff(ctx context.Context, t marketplace.DriverTariff) error {
	idem, err := commandIdempotency(ctx); if err != nil { return err }
	_, _, err = a.idem.CreateTariffIdempotent(ctx, t.DriverID, idem.Key, idem.RequestHash, t)
	return err
}
func (a *idempotentV2Adapter) ListDriverTariffs(ctx context.Context, driverID string) ([]marketplace.DriverTariff, error) { return a.base.ListDriverTariffs(ctx, driverID) }
func (a *idempotentV2Adapter) CreateRequest(ctx context.Context, r marketplace.Request) error {
	idem, err := commandIdempotency(ctx); if err != nil { return err }
	_, _, err = a.idem.CreateRequestIdempotent(ctx, r.RiderID, idem.Key, idem.RequestHash, r)
	return err
}
func (a *idempotentV2Adapter) GetRequest(ctx context.Context, id string) (marketplace.Request, error) { return a.base.GetRequest(ctx, id) }
func (a *idempotentV2Adapter) CancelRequest(ctx context.Context, requestID, riderID string) error {
	idem, err := commandIdempotency(ctx); if err != nil { return err }
	_, err = a.idem.CancelRequestIdempotent(ctx, riderID, idem.Key, idem.RequestHash, requestID)
	return err
}
func (a *idempotentV2Adapter) ListEligibleRequests(ctx context.Context, instanceID string, serviceType marketplace.ServiceType) ([]marketplace.Request, error) { return a.base.ListEligibleRequests(ctx, instanceID, serviceType) }
func (a *idempotentV2Adapter) SubmitQuote(ctx context.Context, q marketplace.Quote) error {
	idem, err := commandIdempotency(ctx); if err != nil { return err }
	_, _, err = a.idem.SubmitQuoteIdempotent(ctx, q.DriverID, idem.Key, idem.RequestHash, q)
	return err
}
func (a *idempotentV2Adapter) ListOffers(ctx context.Context, requestID string) ([]marketplace.Quote, error) { return a.base.ListOffers(ctx, requestID) }
func (a *idempotentV2Adapter) WithdrawQuote(ctx context.Context, quoteID, driverID string) error {
	idem, err := commandIdempotency(ctx); if err != nil { return err }
	_, err = a.idem.WithdrawQuoteIdempotent(ctx, driverID, idem.Key, idem.RequestHash, quoteID)
	return err
}

func isIdempotencyConflict(err error) bool { return errors.Is(err, marketplacestore.ErrIdempotencyConflict) }
