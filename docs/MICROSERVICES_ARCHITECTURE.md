# OpenRide Microservices Architecture

OpenRide is moving from a compatibility modular runtime to independently deployable services.

## Service topology

```text
Clients
  |
  v
Edge Gateway / BFF
  |
  +--> Identity Service
  +--> Marketplace Service
  +--> Location Service
  +--> Ride Service
  +--> Payment Service
  +--> Trust Service
  +--> Realtime Service
  +--> Notification Service
  +--> Operator Service

                     NATS JetStream
                          |
        +-----------------+-----------------+
        |                 |                 |
   domain events     integration events   retries/DLQ
```

## Hard rules

1. Every service is independently deployable.
2. Every service owns its data model and migrations.
3. No service reads another service's database directly.
4. Cross-service identifiers are references, not foreign keys across databases.
5. External clients enter through the gateway/BFF.
6. Synchronous internal calls use gRPC only where the caller truly needs an immediate answer.
7. Domain/integration events use NATS JetStream.
8. Cross-service workflows use sagas/process managers, never distributed database transactions.
9. Event publishing from a transaction uses the transactional outbox pattern.
10. Consumers use inbox/idempotency records before applying side effects.
11. Every service exposes health/readiness and OpenTelemetry signals.
12. Shared packages may contain stable value objects/contracts, but never shared persistence repositories or service-owned business state.

## Initial bounded contexts

### edge-gateway

Responsibilities:
- public REST/WebSocket edge;
- auth token verification;
- API versioning;
- rate limiting;
- request correlation;
- routing/BFF composition.

Must not own marketplace pricing or ride state.

### identity-service

Owns:
- account/authentication;
- rider/driver/operator identity;
- sessions/tokens;
- KYC references and identity status.

### marketplace-service

Owns:
- MobilityRequest;
- DriverTariff;
- Quote;
- ranking policy orchestration;
- Agreement;
- service-module catalog/validation.

This is the first V2 service extracted from the compatibility runtime.

### location-service

Owns:
- online/offline presence;
- current driver position;
- nearby candidate discovery;
- geospatial indexes;
- location freshness.

PostGIS can hold durable geo history where required. Redis may be used for hot presence/location indexes, but only this service owns those keys.

### ride-service

Owns:
- execution state after Agreement;
- service-specific ride/job lifecycle;
- incident/cancellation execution state;
- ride state history.

### payment-service

Owns:
- payment intents;
- captures/refunds;
- driver earnings;
- platform/operator fees;
- settlement ledger.

### trust-service

Owns:
- ratings/reputation;
- reports;
- safety/fraud signals;
- moderation decisions.

### realtime-service

Owns:
- client WebSocket/SSE sessions;
- event fanout;
- presence of connected clients;
- replay/resync transport behavior.

Durable truth remains in owning domain services.

### notification-service

Owns:
- push/SMS/email jobs;
- templates;
- provider retries;
- delivery receipts.

### operator-service

Owns operator-facing read models and workflows. It consumes events and calls owning services; it does not join service databases.

## Communication policy

Use sync calls only when the user request cannot continue without the answer.

Examples:

```text
Gateway -> Marketplace: create mobility request       sync
Marketplace -> Location: discover candidates          gRPC sync
Marketplace -> NATS: agreement.created.v1             async
Ride <- NATS: agreement.created.v1                    async
Payment <- NATS: ride.completed.v1                    async
Notification <- NATS: quote.created.v1                async
Operator read model <- NATS: all relevant events      async
```

## Event naming

```text
openride.<context>.<aggregate>.<event>.v<major>
```

Examples:

```text
openride.marketplace.request.opened.v1
openride.marketplace.quote.created.v1
openride.marketplace.agreement.created.v1
openride.ride.ride.started.v1
openride.ride.ride.completed.v1
openride.payment.payment.captured.v1
```

Events must contain:
- event id;
- event name/version;
- aggregate id/type;
- instance/operator id;
- occurred_at;
- correlation id;
- causation id where available;
- payload.

## Data ownership

Development may use one PostgreSQL cluster for convenience, but services use separate databases/schemas and credentials.

Production target:

```text
identity_db       -> identity-service only
marketplace_db    -> marketplace-service only
location_db       -> location-service only
ride_db           -> ride-service only
payment_db        -> payment-service only
trust_db          -> trust-service only
operator_db       -> operator-service read models only
```

A shared PostgreSQL server is acceptable initially. A shared logical database with cross-service queries is not.

## Reliability

Every command endpoint must be retry-safe where practical. Critical commands require idempotency keys.

For event publication:

```text
DB transaction
  ├── domain state change
  └── outbox row
COMMIT
      ↓
outbox relay
      ↓
NATS JetStream
      ↓
consumer inbox dedupe
      ↓
consumer state change
```

This prevents the classic "DB committed but event was lost" failure.

## Observability

Every request/event carries a correlation id. Services should emit:
- structured JSON logs;
- OpenTelemetry traces;
- RED metrics (rate/errors/duration);
- consumer lag/retry/DLQ metrics;
- business metrics such as quote conversion and matching latency.

## Migration from the current runtime

`services/api` becomes a compatibility gateway/runtime during migration.

Extraction order:

1. marketplace-service;
2. location-service;
3. ride-service;
4. identity-service;
5. payment-service;
6. trust/realtime/notification/operator services.

A legacy capability is removed from `services/api` only after traffic has moved to the owning service and compatibility tests pass.

The goal is not "many services". The goal is **independent ownership, deployment, data boundaries and failure isolation**.
