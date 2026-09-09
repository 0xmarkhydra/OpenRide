# ADR — OpenRide Marketplace V2 Rebaseline

Status: **Accepted for implementation on rebaseline branch**

Date: 2026-09-09

## Context

The repository evolved through multiple product definitions:

1. conventional ride-hailing / Grab-Uber style dispatch;
2. FlashX designated-driver and vehicle-inspection assistance;
3. OpenRide community/open-source mobility direction.

The technical foundation remains useful, but the business kernel still reflects platform-priced trips and one-driver dispatch.

OpenRide's target product requires driver-owned pricing, rider choice and a self-hostable community marketplace.

## Decision 1 — Marketplace, not Grab clone

OpenRide is defined as an open mobility marketplace.

Core chain:

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

Consequences:
- Trip is no longer the target root of every business concern;
- demand, offer, agreement and execution receive separate state models;
- legacy Trip remains temporarily for compatibility.

## Decision 2 — Driver-owned pricing

Introduce `DriverTariff` as first-class domain data.

The platform may provide route facts, market suggestions and transparent operator/legal guardrails.

It may not silently replace driver-owned pricing in the target model.

## Decision 3 — Multiple offers

The matching engine evolves from:

```text
find -> rank -> choose one driver -> send offer
```

to:

```text
find candidates -> obtain quotes -> rank offers -> expose choices
```

Quick Match remains possible but chooses from valid driver-authorized quotes under rider constraints.

## Decision 4 — Ranking is multi-objective

Lowest fare alone is not sufficient.

Initial ranking should be deterministic and may include:
- pickup ETA;
- price fit;
- rating/quality;
- reliability;
- preferences;
- bounded fairness/exposure.

Ranking results should include explainable reason codes.

## Decision 5 — Agreement is immutable commercial truth

When a quote is accepted, create an Agreement containing a snapshot of accepted fare, breakdown and relevant terms.

Subsequent market changes do not mutate the accepted commercial terms.

## Decision 6 — Keep modular monolith

Retain Go modular monolith for current phases.

Do not introduce microservices/Kubernetes/Kafka by default.

Potential future extraction order when justified:
1. Location;
2. Realtime;
3. Marketplace Matching/Ranking;
4. Notifications;
5. Payments/Settlement;
6. Analytics.

## Decision 7 — Keep PostgreSQL/PostGIS + Redis split

PostgreSQL/PostGIS remains durable source of truth.

Redis remains hot state for:
- GEO;
- latest locations;
- short locks;
- TTL;
- presence;
- cache/idempotency.

Accepted agreement data must not depend solely on Redis.

## Decision 8 — Preserve existing runtime through staged migration

No big-bang rewrite.

Migration sequence:

```text
MobilityRequest
-> DriverTariff
-> Quote
-> Marketplace Ranking
-> Agreement
-> Ride
```

Old paths remain until replacement flows are tested.

## Decision 9 — Service verticals sit above marketplace primitives

Passenger ride, designated driver, inspection assistance, delivery and future services may have different execution workflows.

They should reuse the marketplace chain where appropriate:

```text
Request -> Quote -> Agreement -> Execution
```

This prevents legacy FlashX workflows from defining the whole architecture.

## Decision 10 — Prepare for self-host instances

Introduce an `Instance`/operator boundary before federation is required.

Do not implement federation in the early MVP unless there is a real requirement.

## Consequences

Positive:
- aligns product and code with community marketplace mission;
- makes driver pricing explicit/auditable;
- supports rider choice and quick match together;
- enables future verticals;
- protects existing technical investment;
- creates a realistic self-host path.

Costs:
- new tables/APIs/state machines;
- dual compatibility period;
- rider/driver UX redesign;
- more explicit commercial audit data;
- migration/testing complexity.

These costs are accepted because layering negotiation onto the current Trip/Pricing/Dispatch model would create larger long-term complexity and preserve the wrong ownership semantics.
