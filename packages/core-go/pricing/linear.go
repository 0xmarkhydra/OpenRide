package pricing

import (
	"errors"
	"math"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

var (
	ErrInvalidEstimate   = errors.New("pricing: distance and duration must not be negative")
	ErrOverflow          = errors.New("pricing: fare calculation overflow")
	ErrOutsideAutoBounds = errors.New("pricing: calculated fare is outside driver auto-quote bounds")
)

type Estimate struct {
	DistanceM int64 `json:"distance_m"`
	DurationS int64 `json:"duration_s"`
}

// LinearCalculator is a small default tariff implementation. Operators can
// replace it completely through engine.QuoteProvider when their market needs
// another pricing model.
type LinearCalculator struct{}

func (LinearCalculator) Fare(tariff marketplace.DriverTariff, estimate Estimate) (money.Amount, error) {
	if err := tariff.Validate(); err != nil {
		return money.Amount{}, err
	}
	if estimate.DistanceM < 0 || estimate.DurationS < 0 {
		return money.Amount{}, ErrInvalidEstimate
	}

	currency := tariff.BaseFare.Currency
	distance, err := scaled(tariff.PerKM.Minor, estimate.DistanceM, 1000)
	if err != nil {
		return money.Amount{}, err
	}
	duration, err := scaled(tariff.PerMinute.Minor, estimate.DurationS, 60)
	if err != nil {
		return money.Amount{}, err
	}

	total := tariff.BaseFare.Minor
	for _, part := range []int64{distance, duration, tariff.PickupFee.Minor} {
		if part > 0 && total > math.MaxInt64-part {
			return money.Amount{}, ErrOverflow
		}
		total += part
	}
	if total < tariff.MinimumFare.Minor {
		total = tariff.MinimumFare.Minor
	}
	return money.New(currency, total)
}

func (c LinearCalculator) AutoFare(tariff marketplace.DriverTariff, estimate Estimate) (money.Amount, error) {
	fare, err := c.Fare(tariff, estimate)
	if err != nil {
		return money.Amount{}, err
	}
	if !tariff.AllowsAutoQuote(fare) {
		return money.Amount{}, ErrOutsideAutoBounds
	}
	return fare, nil
}

func scaled(rate, units, denominator int64) (int64, error) {
	if rate == 0 || units == 0 {
		return 0, nil
	}
	if rate > math.MaxInt64/units {
		return 0, ErrOverflow
	}
	return rate * units / denominator, nil
}
