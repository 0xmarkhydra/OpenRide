package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/engine"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/money"
)

type carService struct{}

func (carService) Manifest() extension.Manifest {
	return extension.Manifest{
		ID: "passenger.car", Version: "1.0.0", DisplayName: "Passenger Car", Category: "passenger",
		Capabilities: []extension.Capability{extension.CapabilityPassenger, extension.CapabilityScheduled},
	}
}
func (carService) ValidateRequest(_ context.Context, request marketplace.Request) error {
	if request.Destination == nil {
		return fmt.Errorf("passenger.car requires a destination")
	}
	return nil
}

type nearbyDrivers struct{}

func (nearbyDrivers) Candidates(context.Context, marketplace.Request) ([]engine.Candidate, error) {
	return []engine.Candidate{
		{DriverID: "driver-a", PickupDistanceM: 800, PickupETAS: 180},
		{DriverID: "driver-b", PickupDistanceM: 450, PickupETAS: 100},
		{DriverID: "driver-c", PickupDistanceM: 1200, PickupETAS: 250},
	}, nil
}

type driverQuotes struct{ now time.Time }

func (q driverQuotes) Quote(_ context.Context, request marketplace.Request, candidate engine.Candidate) (marketplace.Quote, error) {
	fares := map[string]int64{"driver-a": 50_000, "driver-b": 60_000, "driver-c": 55_000}
	return marketplace.Quote{
		ID: "quote-" + candidate.DriverID,
		RequestID: request.ID,
		DriverID: candidate.DriverID,
		Status: marketplace.QuotePending,
		Fare: money.Must("VND", fares[candidate.DriverID]),
		PickupDistanceM: candidate.PickupDistanceM,
		PickupETAS: candidate.PickupETAS,
		Explanation: []string{"price authorized by driver policy"},
		CreatedAt: q.now,
		ExpiresAt: q.now.Add(45 * time.Second),
	}, nil
}

type balancedRanker struct{}

func (balancedRanker) Rank(_ context.Context, _ marketplace.Request, quotes []marketplace.Quote) ([]engine.RankedQuote, error) {
	// Demo only: combine normalized price and pickup ETA instead of blindly picking the cheapest.
	sort.Slice(quotes, func(i, j int) bool {
		left := float64(quotes[i].Fare.Minor)/1000 + float64(quotes[i].PickupETAS)/20
		right := float64(quotes[j].Fare.Minor)/1000 + float64(quotes[j].PickupETAS)/20
		return left < right
	})
	out := make([]engine.RankedQuote, 0, len(quotes))
	for i, quote := range quotes {
		out = append(out, engine.RankedQuote{
			Quote: quote,
			Score: float64(len(quotes) - i),
			Recommended: i == 0,
			Reasons: []string{"balanced fare + pickup ETA"},
		})
	}
	return out, nil
}

func main() {
	now := time.Now().UTC()
	destination := geo.Point{Lat: 21.0285, Lng: 105.8542}
	request := marketplace.Request{
		ID: "request-demo", InstanceID: "community-hanoi", RiderID: "rider-1",
		ServiceType: "passenger.car", Status: marketplace.RequestOpen,
		Pickup: geo.Point{Lat: 21.0278, Lng: 105.8342}, Destination: &destination,
		RequestedAt: now, Version: 1,
	}

	core := engine.Marketplace{
		Services: extension.MustRegistry(carService{}),
		Candidates: nearbyDrivers{},
		Quotes: driverQuotes{now: now},
		Ranker: balancedRanker{},
	}

	report, err := core.FindOffers(context.Background(), request)
	if err != nil {
		panic(err)
	}
	for _, offer := range report.Offers {
		fmt.Printf("driver=%s fare=%d %s eta=%ds recommended=%v reasons=%v\n",
			offer.Quote.DriverID, offer.Quote.Fare.Minor, offer.Quote.Fare.Currency,
			offer.Quote.PickupETAS, offer.Recommended, offer.Reasons)
	}
}
