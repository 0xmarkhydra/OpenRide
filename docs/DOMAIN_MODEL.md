# OpenRide Domain Model V2

## 1. Domain rule

OpenRide is modeled as a marketplace first and a ride executor second.

The primary business chain is:

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

Legacy `Trip` semantics remain temporarily for compatibility but are no longer the target business model.

## 2. Identity

### User

Shared account identity.

Core attributes:
- id;
- phone/email identity;
- status;
- locale;
- created_at/updated_at.

A user may participate in more than one role over time.

### RiderProfile

Rider-specific preferences and trust context.

### DriverProfile

Driver identity, approval/KYC state, availability, rating aggregates and service capabilities.

Approval example:

```text
PENDING -> APPROVED
PENDING -> REJECTED
PENDING -> MORE_INFO_REQUIRED
MORE_INFO_REQUIRED -> PENDING / APPROVED / REJECTED
```

Availability example:

```text
OFFLINE -> ONLINE -> RESERVED/BUSY -> ONLINE -> OFFLINE
```

Only eligible/approved drivers can participate in live marketplace matching.

## 3. DriverVehicle and Capability

For passenger ride verticals, the provider vehicle belongs to the driver/provider side.

Suggested `DriverVehicle` attributes:
- id;
- driver_id;
- type;
- plate_number;
- brand/model;
- year/color;
- seats;
- service eligibility;
- document/inspection status;
- active status.

`DriverCapability` represents what work a driver is qualified/willing to perform.

Examples:
- passenger_car;
- passenger_motorbike;
- designated_driver_car;
- designated_driver_bike;
- vehicle_inspection_assist;
- delivery.

Legacy customer-owned vehicle support remains relevant for designated-driver/inspection verticals but must not define the entire OpenRide core.

## 4. DriverTariff

`DriverTariff` is a first-class aggregate and represents driver-owned commercial policy.

Core attributes:
- id;
- driver_id;
- instance_id;
- service_type;
- status;
- quote_mode: manual | auto | hybrid;
- currency;
- base_fare;
- minimum_fare;
- per_km;
- per_minute;
- pickup_fee/rules;
- auto_quote_min;
- auto_quote_max;
- version;
- effective time range.

Rules can cover:
- night/time windows;
- zones;
- holidays;
- long distance;
- pickup distance;
- service-specific conditions.

Important invariant:

> The platform may recommend a price but must not silently exceed the bounds accepted by the driver when auto quoting.

## 5. MobilityRequest

`MobilityRequest` captures rider demand before a driver is selected.

Core attributes:
- id;
- instance_id;
- rider_id;
- service_type;
- pickup;
- destination;
- optional stops;
- route facts;
- rider preferences;
- rider constraints;
- status;
- requested_at;
- expires_at;
- version.

Suggested states:

```text
DRAFT
  -> OPEN
  -> RECEIVING_QUOTES
  -> AGREED
  -> CLOSED

OPEN/RECEIVING_QUOTES -> CANCELLED
OPEN/RECEIVING_QUOTES -> EXPIRED
```

A request is demand, not an assigned trip.

## 6. Candidate

A candidate is an eligible driver discovered for a request.

Candidate discovery considers:
- geographic proximity/ETA;
- online/fresh location;
- capability/service type;
- vehicle eligibility;
- KYC/account status;
- service area;
- existing reservation/busy state;
- tariff/quote eligibility.

Candidate data may be ephemeral and does not always require a durable row unless audit/analytics require it.

## 7. Quote

`Quote` is a driver-authorized commercial offer for one request.

Core attributes:
- id;
- request_id;
- driver_id;
- driver_vehicle_id nullable;
- tariff_id/version;
- total fare;
- currency;
- pricing breakdown;
- pickup ETA/distance snapshot;
- status;
- expires_at;
- created_at;
- accepted_at.

Suggested states:

```text
PENDING -> ACCEPTED
PENDING -> REJECTED
PENDING -> WITHDRAWN
PENDING -> EXPIRED
```

Important invariants:
- quote belongs to exactly one request and one driver;
- accepted quote must not be expired;
- quote stores enough pricing context to explain the number shown to users;
- platform ranking does not mutate quote commercial terms.

## 8. CounterOffer

Counter-offer negotiation is optional and can be introduced after the basic marketplace is stable.

A future `CounterOffer` may support:
- rider proposes budget;
- driver counters;
- rider accepts/rejects;
- bounded negotiation rounds/TTL.

Do not block the initial marketplace on full negotiation support.

## 9. RankingResult

Ranking is not necessarily a durable aggregate, but it must produce explainable output.

Typical signals:
- pickup ETA;
- price fit;
- rating/quality;
- completion reliability;
- rider preferences;
- bounded fairness/exposure.

Typical reason codes:
- BEST_OVERALL;
- FAST_PICKUP;
- LOWEST_PRICE;
- HIGH_RATING;
- PREFERRED_VEHICLE.

Cheapest must not automatically equal best.

## 10. Agreement

`Agreement` is created when a rider accepts a valid quote or Quick Match selects one under rider-authorized constraints.

It is the commercial source of truth.

Snapshot fields should include:
- id;
- instance_id;
- request_id;
- quote_id;
- rider_id;
- driver_id;
- vehicle/service context;
- pickup/destination;
- fare total;
- currency;
- full pricing breakdown;
- accepted terms/policy version;
- created_at.

Key invariants:
- one active accepted agreement per normal single-driver request;
- immutable accepted fare snapshot;
- idempotent creation;
- driver cannot be assigned to incompatible concurrent rides;
- request and quote status transition atomically with agreement creation.

## 11. Ride

`Ride` represents actual execution after agreement.

Passenger ride example:

```text
ASSIGNED
 -> DRIVER_EN_ROUTE
 -> DRIVER_ARRIVED
 -> PASSENGER_ONBOARD
 -> IN_PROGRESS
 -> COMPLETED
```

Terminal/exception paths:
- CANCELLED;
- FAILED;
- incident/dispute lifecycle as separate records where appropriate.

A service vertical may define additional execution states without changing marketplace primitives.

## 12. Service Vertical

OpenRide should support service-specific behavior through policy/capability boundaries rather than putting every state into one universal Ride enum.

Examples:

### Passenger ride
Driver uses their eligible vehicle to transport rider.

### Designated driver
Driver travels to customer and drives customer-owned vehicle.

### Vehicle inspection assistance
Provider receives customer vehicle/documents, completes a workflow and returns vehicle.

All can share:

```text
Request -> Quote -> Agreement -> Execution
```

while their execution state machines differ.

## 13. Payment

Payment state is separate from Ride state.

Example:

```text
PENDING -> AUTHORIZED -> CAPTURED
PENDING/AUTHORIZED -> FAILED
CAPTURED -> REFUNDED
```

A completed ride may still have payment recovery work.

## 14. DriverEarning / PlatformFee

Earnings should be represented transparently.

Suggested concepts:
- gross agreed fare;
- rider-paid fees;
- operator/platform fee;
- driver earning;
- tax/withholding if applicable;
- settlement status.

OpenRide should avoid hiding deductions inside an opaque final number.

## 15. Rating and Trust

Trust domain includes:
- rider-to-driver rating;
- driver-to-rider rating where enabled;
- KYC state;
- incident reports;
- fraud/risk signals;
- moderation/suspension;
- reliability aggregates.

Trust signals can influence ranking only through documented/bounded rules.

## 16. Instance

`Instance` is the future self-host/operator boundary.

Possible attributes:
- id;
- name;
- service area;
- currency;
- supported service types;
- payment configuration references;
- KYC policy;
- operator fee policy;
- legal/price guardrails;
- ranking configuration/version.

Federation is not required initially, but avoiding one-global-operator assumptions now reduces later migration cost.

## 17. Operator / Audit

Sensitive actions require append-only audit records:
- KYC decisions;
- suspensions;
- policy/ranking config changes;
- legal price guardrail changes;
- manual agreement/ride intervention;
- payment/refund overrides;
- dispute resolution.

Audit records should capture actor, action, target, time, reason and metadata.

## 18. Compatibility mapping

During migration:

```text
legacy trips              -> compatibility execution/request projection
legacy pricing_rules      -> fallback/operator recommendation layer
legacy dispatch offers    -> precursor to marketplace quote delivery
legacy customer_vehicles  -> designated-driver vertical support
```

Do not remove these until replacement flows and migrations are tested.
