package marketplace

func ValidRequestTransition(from, to RequestStatus) bool {
	switch from {
	case RequestDraft:
		return to == RequestOpen || to == RequestCancelled
	case RequestOpen:
		return to == RequestReceivingQuotes || to == RequestAgreed || to == RequestCancelled || to == RequestExpired
	case RequestReceivingQuotes:
		return to == RequestAgreed || to == RequestCancelled || to == RequestExpired
	case RequestAgreed:
		return to == RequestClosed
	default:
		return false
	}
}

func ValidQuoteTransition(from, to QuoteStatus) bool {
	if from != QuotePending {
		return false
	}
	switch to {
	case QuoteAccepted, QuoteRejected, QuoteWithdrawn, QuoteExpired, QuoteInvalidated:
		return true
	default:
		return false
	}
}

func ValidPassengerRideTransition(from, to RideStatus) bool {
	if to == RideCancelled || to == RideFailed {
		switch from {
		case RideAssigned, RideDriverEnRoute, RideDriverArrived, RidePassengerOnboard, RideInProgress:
			return true
		default:
			return false
		}
	}

	switch from {
	case RideAssigned:
		return to == RideDriverEnRoute
	case RideDriverEnRoute:
		return to == RideDriverArrived
	case RideDriverArrived:
		return to == RidePassengerOnboard
	case RidePassengerOnboard:
		return to == RideInProgress
	case RideInProgress:
		return to == RideCompleted
	default:
		return false
	}
}
