package store

import (
	"context"
	"encoding/json"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

func (s *Store) ListDriverTariffs(ctx context.Context, driverID string) ([]marketplace.DriverTariff, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,instance_id,driver_id,service_type,quote_mode,currency,base_fare_minor,minimum_fare_minor,per_km_minor,per_minute_minor,pickup_fee_minor,auto_quote_min_minor,auto_quote_max_minor,rules,version
	FROM driver_tariffs WHERE driver_id=$1 AND active=TRUE ORDER BY service_type,id`, driverID)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []marketplace.DriverTariff
	for rows.Next() {
		var t marketplace.DriverTariff
		var service, mode, currency string
		var base, minimum, perKM, perMinute, pickup, autoMin, autoMax int64
		var rules []byte
		if err := rows.Scan(&t.ID,&t.InstanceID,&t.DriverID,&service,&mode,&currency,&base,&minimum,&perKM,&perMinute,&pickup,&autoMin,&autoMax,&rules,&t.Version); err != nil { return nil, err }
		t.ServiceType = marketplace.ServiceType(service)
		t.QuoteMode = marketplace.QuoteMode(mode)
		t.BaseFare = money.Must(currency, base)
		t.MinimumFare = money.Must(currency, minimum)
		t.PerKM = money.Must(currency, perKM)
		t.PerMinute = money.Must(currency, perMinute)
		t.PickupFee = money.Must(currency, pickup)
		t.AutoQuoteMinimum = money.Must(currency, autoMin)
		t.AutoQuoteMaximum = money.Must(currency, autoMax)
		if len(rules) > 0 { if err := json.Unmarshal(rules, &t.Rules); err != nil { return nil, err } }
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) CancelRequest(ctx context.Context, requestID, riderID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE mobility_requests SET status='cancelled',version=version+1,updated_at=now() WHERE id=$1 AND rider_id=$2 AND status IN ('open','receiving_quotes')`, requestID, riderID)
	if err != nil { return err }
	if tag.RowsAffected() != 1 { return ErrNotFound }
	_, _ = s.pool.Exec(ctx, `UPDATE marketplace_quotes SET status='invalidated' WHERE request_id=$1 AND status='pending'`, requestID)
	return nil
}

func (s *Store) WithdrawQuote(ctx context.Context, quoteID, driverID string) error {
	tag, err := s.pool.Exec(ctx, `UPDATE marketplace_quotes SET status='withdrawn' WHERE id=$1 AND driver_id=$2 AND status='pending'`, quoteID, driverID)
	if err != nil { return err }
	if tag.RowsAffected() != 1 { return ErrNotFound }
	return nil
}
