# ADR — OpenRide Marketplace V2 + Microservices Rebaseline

Status: **Accepted**

Date: 2026-09-09

## Context

The repository evolved through three product identities:

1. conventional ride-hailing / Grab-Uber style dispatch;
2. FlashX designated-driver and vehicle-inspection assistance;
3. OpenRide community/open-source mobility marketplace.

The technical foundation is still useful, but the target system must support driver-owned pricing, rider choice, self-hosting, multiple service verticals and independent capability scaling.

## Decision 1 — Marketplace, not Grab clone

OpenRide is an open mobility marketplace.

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

Demand, offer, accepted commercial truth and execution are separate aggregates.

## Decision 2 — Driver-owned pricing

`DriverTariff` is first-class domain data.

The system may provide route facts, market suggestions and transparent operator/legal guardrails. It may not silently replace a driver's authorized commercial terms.

## Decision 3 — Multiple offers + explainable ranking

Matching evolves from:

```text
find candidates -> pick one -> send platform offer
```

to:

```text
find candidates -> obtain driver-authorized quotes -> rank -> expose choices
```

Quick Match may choose on the rider's behalf only from valid quotes under rider constraints.

Ranking is multi-objective and must provide explanation reasons.

## Decision 4 — Agreement is immutable commercial truth

Accepted fare and relevant terms are snapshotted into Agreement and cannot be silently recomputed after acceptance.

## Decision 5 — Microservices are the target deployment architecture

OpenRide adopts microservices by bounded context.

Target services:

```text
edge-gateway
identity-service
marketplace-service
location-service
ride-service
payment-service
trust-service
realtime-service
notification-service
operator-service
```

`services/api` becomes a compatibility edge/runtime while capabilities are extracted.

## Decision 6 — Database ownership belongs to services

Each domain service owns its database/schema credentials and migrations.

Rules:
- no cross-service SQL;
- no cross-service database foreign keys;
- cross-service IDs are opaque references;
- projections are built from APIs/events, not direct joins.

Small deployments may colocate databases on one PostgreSQL cluster without weakening logical ownership.

## Decision 7 — NATS JetStream for asynchronous integration

Durable cross-service events use NATS JetStream.

Event publication follows transactional Outbox; consumers use Inbox/idempotency.

This avoids dual-write loss between database state and event publication.

## Decision 8 — gRPC only for truly synchronous internal dependencies

Use gRPC when the caller cannot proceed without an immediate response.

Example:

```text
Marketplace -> Location: nearby candidate discovery
```

Do not create synchronous call chains for work that can react to events.

## Decision 9 — Saga/process managers replace distributed transactions

Cross-service business workflows must not depend on 2PC/distributed DB transactions.

For example:

```text
Marketplace creates Agreement
 -> publishes agreement.created
 -> Ride creates execution
 -> Notification informs actors
```

Compensation/recovery is modeled explicitly when necessary.

## Decision 10 — Shared Core is allowed, shared persistence is not

`packages/core-go` may provide stable value objects, invariants and pure policy interfaces.

`packages/contracts` provides versioned wire contracts.

Shared packages must not contain repositories that allow one service to manipulate another service's database.

## Decision 11 — Service verticals remain plug-in modules

Passenger ride, carpool, delivery, designated driver and assistance workflows reuse generic marketplace primitives while owning vertical-specific validation/lifecycle behavior.

```text
passenger.car
carpool.intercity
parcel.instant
passenger.motorbike
designated-driver.car
...
```

## Decision 12 — Preserve compatibility through staged extraction

No big-bang rewrite.

Current order:

```text
Marketplace Service      first extraction
Location Service         next
Ride Service
Identity Service
Payment Service
Trust / Realtime / Notification / Operator
```

A legacy capability is removed only after the owning service is deployed and traffic/compatibility tests pass.

## Decision 13 — Self-hosting remains first-class

Microservices must not require Kubernetes.

A community should be able to run a small OpenRide deployment using Docker Compose while larger operators may choose Kubernetes or another orchestrator.

The same service/data boundaries apply in both modes.

## Consequences

### Positive

- clear service and data ownership;
- independent deployment and failure isolation;
- easier scaling of Location/Marketplace/Realtime independently;
- service-specific migrations and rollback;
- packageable Core without a giant application dependency;
- strong self-host path;
- multiple mobility verticals can evolve without redefining the kernel.

### Costs

- more deployment units;
- network/event failure modes;
- observability becomes mandatory;
- eventual consistency must be modeled explicitly;
- contract/version governance becomes more important;
- migration temporarily carries compatibility runtime plus new services.

These costs are accepted because OpenRide is intended to become reusable infrastructure rather than one closed application deployment.
