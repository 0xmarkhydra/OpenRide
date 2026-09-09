# OpenRide System Architecture V2

## 1. Direction

OpenRide uses **microservices at deployment boundaries** and clean/hexagonal architecture inside each service.

The goal is not to maximize service count. The goal is independent ownership, deployment, data isolation, failure containment and clear bounded contexts.

```text
Rider / Driver / Operator
          |
          v
   Edge Gateway / BFF
          |
  +-------+--------+----------------+
  |       |        |                |
Identity Market  Ride             Trust
          |
          +---- gRPC ----> Location

All domain services <----> NATS JetStream
```

The compatibility `services/api` runtime remains during migration but is not the target place for new Marketplace V2 ownership.

## 2. Hard architecture rules

1. Every domain service is independently deployable.
2. Every service owns its persistence model and migrations.
3. No service reads another service's database directly.
4. Cross-service IDs are references, not database foreign keys.
5. Public clients enter through an Edge Gateway/BFF.
6. Internal synchronous communication uses gRPC when an immediate answer is required.
7. Asynchronous integration uses NATS JetStream.
8. Cross-service workflows use Saga/process-manager patterns, never distributed DB transactions.
9. Transactional state + integration event uses an Outbox.
10. Event consumers use Inbox/idempotency before side effects.
11. Every service exposes health, readiness, structured logs, metrics and traces.
12. Shared packages may contain value objects/contracts/invariants but never shared repositories for service-owned state.

## 3. Bounded contexts

### Edge Gateway / BFF

Owns transport concerns:
- public REST/WebSocket edge;
- token verification;
- request correlation;
- API versioning;
- rate limits;
- BFF composition;
- migration/proxy compatibility.

It must not own marketplace pricing, agreements or ride state.

### Identity Service

Owns:
- authentication;
- accounts;
- sessions/tokens;
- identity roles;
- KYC identity status/references.

### Marketplace Service

Owns:
- MobilityRequest;
- DriverTariff;
- Quote;
- ranking orchestration;
- Agreement;
- service-module catalog and request validation.

This is the first extracted V2 service.

### Location Service

Owns:
- driver online/offline state;
- latest location;
- location freshness;
- nearby candidate discovery;
- Redis GEO/hot geo indexes;
- optional durable geo history.

Marketplace may synchronously ask Location for candidates. Marketplace must not read Location Redis keys directly.

### Ride Service

Owns execution after Agreement:
- ride/job lifecycle;
- state history;
- execution cancellation;
- incidents tied to execution;
- service-specific lifecycle state.

### Payment Service

Owns:
- payment intents;
- capture/refund state;
- driver earnings;
- operator/platform fee records;
- settlement ledger.

### Trust Service

Owns:
- ratings;
- reputation;
- safety reports;
- moderation/fraud signals;
- trust decisions.

### Realtime Service

Owns client realtime transport:
- WebSocket/SSE sessions;
- event fanout;
- reconnect/replay transport behavior;
- connected-client presence.

Durable business truth stays with domain services.

### Notification Service

Owns:
- push/SMS/email jobs;
- templates;
- provider retries;
- delivery receipts.

### Operator Service

Owns operator workflows and read models. It consumes events and invokes owning services rather than joining their databases.

## 4. Marketplace kernel

The cross-cutting marketplace model remains:

```text
MobilityRequest
    -> DriverTariff
    -> Quote
    -> Ranking
    -> Agreement
    -> Ride
```

`Request`, `Quote`, `Agreement` and `Ride` are separate concepts.

- Request is demand.
- Tariff is a driver's commercial policy.
- Quote is a driver-authorized offer for one request.
- Agreement is immutable accepted commercial truth.
- Ride is execution.

Stable invariants live in `packages/core-go`; service-owned persistence and orchestration live in the owning service.

## 5. Communication model

Use synchronous calls only when the caller cannot continue without the answer.

Examples:

```text
Gateway -> Marketplace: create request                HTTP/gRPC sync
Marketplace -> Location: discover nearby candidates   gRPC sync
Marketplace -> NATS: agreement.created.v1             async
Ride <- NATS: agreement.created.v1                    async
Payment <- NATS: ride.completed.v1                    async
Notification <- NATS: quote.created.v1                async
Operator <- NATS: domain events                       async projection
```

## 6. Event naming and envelope

Event names follow:

```text
openride.<context>.<aggregate>.<event>.v<major>
```

Examples:

```text
openride.marketplace.request.opened.v1
openride.marketplace.quote.created.v1
openride.marketplace.agreement.created.v1
openride.ride.ride.completed.v1
openride.payment.payment.captured.v1
```

Every integration event carries:
- event ID;
- event name/version;
- aggregate ID/type;
- instance/operator ID;
- occurred_at;
- correlation ID;
- causation ID when available;
- payload.

## 7. Data ownership

Development can use one PostgreSQL cluster, but logical ownership remains isolated:

```text
identity_db       -> identity-service
marketplace_db    -> marketplace-service
location_db       -> location-service
ride_db           -> ride-service
payment_db        -> payment-service
trust_db          -> trust-service
operator_db       -> operator projections
```

No service may perform SQL joins across these boundaries.

Redis namespaces/instances also have an owner. For example, hot driver GEO data belongs to Location Service, not Marketplace.

## 8. Reliability model

Critical command endpoints should be retry-safe and idempotent.

Transactional event publication:

```text
BEGIN DB TX
  mutate owned state
  insert outbox event
COMMIT
   |
   v
outbox relay
   |
   v
NATS JetStream
   |
   v
consumer inbox dedupe
   |
   v
consumer-owned state change
```

This prevents state being committed while its required integration event is lost.

## 9. Agreement consistency

Agreement acceptance is a strong consistency boundary **inside Marketplace Service**.

Marketplace must atomically verify:
- request is still open;
- quote belongs to request/driver;
- quote is pending and unexpired;
- quote has not already been accepted;
- accepted terms are snapshotted;
- idempotency constraints hold;
- outbox event is written in the same DB transaction.

Other services react to `agreement.created` asynchronously.

Do not attempt a distributed transaction across Marketplace, Ride and Payment.

## 10. Location path

```text
Driver GPS
  -> Edge/Realtime ingress
  -> Location Service
  -> validation/freshness
  -> Redis GEO + latest position
  -> candidate discovery API
  -> realtime fanout where required
```

Raw GPS pings should not flood transactional business tables.

## 11. Observability

Every service should emit:
- structured JSON logs;
- OpenTelemetry traces;
- correlation IDs;
- RED metrics (rate/errors/duration);
- event-consumer lag/retry/DLQ metrics;
- business metrics such as matching latency, quote conversion and agreement success.

## 12. Self-hosting

OpenRide must work for a community/operator running a small deployment as well as a larger operator.

A small deployment may run all services on one host and one Postgres cluster with separate logical databases. The service ownership rules do not disappear just because infrastructure is colocated.

## 13. Current extraction order

```text
1. Marketplace Service      ✅ first slice
2. Location Service
3. Ride Service
4. Identity Service
5. Payment Service
6. Trust Service
7. Realtime Service
8. Notification Service
9. Operator Service
10. retire compatibility API capability-by-capability
```

Extraction order can change when product priorities require it, but data ownership cannot become ambiguous.

## 14. Technology baseline

- Go domain services;
- Flutter Rider/Driver apps;
- Next.js Operator UI;
- PostgreSQL/PostGIS;
- Redis for owned hot-state use cases;
- NATS JetStream for durable async integration;
- gRPC for required synchronous service-to-service calls;
- REST/WebSocket at public edge;
- S3-compatible object storage;
- OpenTelemetry for observability.

Kubernetes/service mesh are deployment choices, not business-architecture requirements. Docker Compose remains a valid small/self-hosted deployment path.

## 15. Compatibility policy

Legacy `trips`, platform pricing and dispatch remain temporarily in `services/api`.

A legacy capability is deleted only when:
1. its owning V2 service exists;
2. data migration/projection is defined;
3. traffic is migrated;
4. compatibility tests pass;
5. rollback is understood.

See `MICROSERVICES_ARCHITECTURE.md` and `OPENRIDE_MIGRATION_PLAN_V2.md` for the migration source of truth.
