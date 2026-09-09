# OpenRide Marketplace Migration Plan V2

## Goal

Migrate the current FlashX/ride-hailing-oriented business kernel into an OpenRide mobility marketplace without destroying the existing working runtime.

The current repository already has useful foundations for auth, KYC, drivers, realtime, PostgreSQL/PostGIS, Redis, payments, ratings and CI. Those foundations should be retained.

The migration changes the business semantics around demand, pricing, matching and assignment.

## Target kernel

```text
OLD
Trip -> Platform Pricing -> Dispatch -> Assignment

NEW
MobilityRequest -> DriverTariff -> Quote -> Marketplace -> Agreement -> Ride
```

## Migration rules

1. Prefer additive schema migrations.
2. Do not edit production-applied migrations in place.
3. Keep old endpoints available until replacement flows are tested.
4. New marketplace state must have a durable source of truth in PostgreSQL.
5. Redis remains hot/ephemeral state for geo, realtime, quote TTL and locks.
6. Do not introduce Kafka/Kubernetes/microservices unless measured production needs justify them.
7. Every phase should leave CI green before moving to the next phase.

## Phase 0 — Rebaseline documentation and naming

Deliverables:
- rewrite root README;
- publish product vision and manifesto;
- redefine architecture/domain/data/matching docs;
- mark FlashX-specific requirements as legacy vertical requirements;
- rename human-facing CI/project labels where low-risk;
- define licensing/governance decision.

Runtime behavior should remain unchanged.

## Phase 1 — MobilityRequest

Add a first-class demand aggregate.

Suggested fields:

```text
id
instance_id
rider_id
service_type
status
pickup
stops[]
destination
requested_at
expires_at
preferences
constraints
created_at
updated_at
version
```

Initial states:

```text
DRAFT -> OPEN -> RECEIVING_QUOTES -> AGREED -> CLOSED
OPEN -> EXPIRED
OPEN -> CANCELLED
```

Compatibility:
- current trip create flow can create both a legacy trip and a request behind a feature flag;
- no rider UI change required yet.

## Phase 2 — DriverTariff

Introduce driver-owned pricing configuration.

Suggested entities:
- `driver_tariffs`;
- `driver_tariff_rules`;
- `driver_quote_preferences`.

A driver can select:
- manual;
- auto;
- hybrid.

The current global pricing service remains a fallback only during migration.

## Phase 3 — Quote

Add durable marketplace quote semantics.

Suggested fields:

```text
id
request_id
driver_id
tariff_id
status
fare_total
currency
pricing_breakdown
pickup_eta_seconds
expires_at
created_at
accepted_at
withdrawn_at
```

States:

```text
PENDING -> ACCEPTED
PENDING -> REJECTED
PENDING -> EXPIRED
PENDING -> WITHDRAWN
```

Quote calculation must snapshot the tariff/rules used for auditability.

## Phase 4 — Marketplace candidate and ranking engine

Refactor current dispatch responsibilities:

```text
Current
nearby -> score -> pick one -> offer

Target
nearby -> eligibility -> quote generation -> rank many -> expose offers
```

Create modules/concepts:
- candidate discovery;
- eligibility;
- quote engine;
- ranking;
- fairness/exposure;
- explainability.

Do not use ML at first. Deterministic weights and transparent reason codes are preferable.

## Phase 5 — Agreement

When a rider selects an offer, create a durable agreement in one atomic flow.

Agreement snapshots:
- rider;
- driver;
- request;
- accepted quote;
- pickup/destination;
- price and breakdown;
- currency;
- service terms;
- accepted timestamp.

The agreement becomes the commercial source of truth.

Protect against:
- quote double-accept;
- driver double-assignment;
- expired quote acceptance;
- rider accepting two drivers;
- fare mutation after acceptance.

## Phase 6 — Ride execution

Introduce a cleaner execution aggregate separate from marketplace negotiation.

Passenger ride state proposal:

```text
ASSIGNED
-> DRIVER_EN_ROUTE
-> DRIVER_ARRIVED
-> PASSENGER_ONBOARD
-> IN_PROGRESS
-> COMPLETED
```

Service-specific workflows can extend execution behavior without changing request/quote/agreement semantics.

The legacy `trips` aggregate can remain as a compatibility projection until consumers move to `rides`.

## Phase 7 — Rider marketplace UI

Replace the single platform estimate + searching UX with:

```text
Request created
-> finding relevant drivers
-> offers arriving realtime
-> compare offers
-> choose / quick match
-> agreement
-> track ride
```

Offer cards should show:
- total price;
- pickup ETA;
- driver rating;
- vehicle summary;
- recommendation reason;
- meaningful labels such as cheapest / fastest / best overall.

Do not label the cheapest offer as best by default.

## Phase 8 — Driver pricing UI

Add a primary Driver App area: `My Price` / `My Tariff`.

Driver controls:
- per-km/per-minute values;
- minimum fare;
- pickup radius/fee;
- long-distance adjustments;
- time rules;
- manual/auto/hybrid mode;
- automatic quote lower/upper bounds.

The offer screen should show how the proposed quote was calculated.

## Phase 9 — Remove platform-owned fare as primary pricing

Once marketplace pricing is stable:
- deprecate global pricing as the commercial source of truth;
- retain operator guardrails and market recommendation tools;
- migrate old consumers to quote/agreement fare data;
- preserve historical fare records.

## Phase 10 — Instance / operator boundary

Add `instance_id` to marketplace-owned entities before federation becomes necessary.

An instance represents one deployment/operator/community context.

Possible instance configuration:
- service area;
- currency;
- payment providers;
- legal pricing bounds;
- KYC policy;
- fee policy;
- supported services;
- ranking configuration within project guardrails.

Do not implement federation yet unless there is a real cross-instance requirement.

## Phase 11 — Service plugin/vertical model

Move designated-driver and vehicle-inspection flows into service-specific execution policies.

Long-term examples:
- passenger ride;
- motorbike ride;
- carpool;
- intercity;
- delivery;
- designated driver;
- inspection assistance.

The marketplace primitives stay reusable:

```text
Request -> Quote -> Agreement -> Job/Ride
```

## Test strategy

Every marketplace phase needs tests for:
- invalid state transitions;
- quote expiration;
- idempotent create/accept;
- assignment race;
- pricing snapshot integrity;
- ranking determinism;
- driver tariff boundaries;
- request cancellation;
- Redis outage/degraded behavior;
- backward compatibility where active.

## Rollout strategy

Use feature flags or instance-level configuration:

```text
legacy_dispatch_enabled
marketplace_quotes_enabled
rider_offer_selection_enabled
driver_tariffs_enabled
quick_match_enabled
```

This permits production/staging comparison without a big-bang release.

## Definition of done

The migration is complete when the normal passenger ride path no longer depends on a platform-controlled fare plus one-driver dispatch flow, and instead uses:

```text
MobilityRequest
+ DriverTariff
+ Quote
+ Marketplace Ranking
+ Agreement
+ Ride
```

with durable auditability and transparent user-facing pricing semantics.
