package pricing

import (
	"testing"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

func tariff() marketplace.DriverTariff {
	return marketplace.DriverTariff{
		ID: "tariff_1", InstanceID: "default", DriverID: "driver_1", ServiceType: "passenger.car",
		QuoteMode: marketplace.QuoteModeAuto,
		BaseFare: money.Must("VND", 10_000), MinimumFare: money.Must("VND", 20_000),
		PerKM: money.Must("VND", 5_000), PerMinute: money.Must("VND", 1_000), PickupFee: money.Must("VND", 2_000),
		AutoQuoteMinimum: money.Must("VND", 20_000), AutoQuoteMaximum: money.Must("VND", 100_000), Version: 1,
	}
}

func TestLinearCalculator(t *testing.T) {
	fare, err := (LinearCalculator{}).AutoFare(tariff(), Estimate{DistanceM: 10_000, DurationS: 600})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 10k base + 50k distance + 10k duration + 2k pickup.
	if fare.Minor != 72_000 {
		t.Fatalf("got %d, want 72000", fare.Minor)
	}
}

func TestLinearCalculatorHonorsDriverAutoBounds(t *testing.T) {
	tf := tariff()
	tf.AutoQuoteMaximum = money.Must("VND", 60_000)
	if _, err := (LinearCalculator{}).AutoFare(tf, Estimate{DistanceM: 10_000, DurationS: 600}); err != ErrOutsideAutoBounds {
		t.Fatalf("got %v, want ErrOutsideAutoBounds", err)
	}
}
