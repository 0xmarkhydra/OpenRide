package marketplace

// PassengerLifecycle is the reference execution lifecycle for a normal
// passenger ride. Service modules can expose another lifecycle through the
// extension.RideLifecycleProvider interface without changing Core enums.
type PassengerLifecycle struct{}

func (PassengerLifecycle) Initial() RideStatus { return RideAssigned }

func (PassengerLifecycle) ValidTransition(from, to RideStatus) bool {
	return ValidPassengerRideTransition(from, to)
}

func (PassengerLifecycle) Terminal(status RideStatus) bool {
	return status == RideCompleted || status == RideCancelled || status == RideFailed
}
