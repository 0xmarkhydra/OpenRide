# OpenRide Marketplace Migration Plan V2

## Goal

Migrate the compatibility FlashX/ride-hailing runtime into an OpenRide microservices platform without a big-bang rewrite.

Target business chain:

```text
MobilityRequest -> DriverTariff -> Quote -> Marketplace -> Agreement -> Ride
```

Target deployment chain:

```text
Clients -> Edge Gateway -> independently deployable bounded-context services
                           <-> NATS JetStream
```

## Migration rules

1. Prefer additive migrations and additive traffic migration.
2. Do not edit already-applied production migrations in place.
3. Each extracted service owns its database and migrations.
4. No service reads another service database.
5. Public compatibility endpoints remain until V2 replacements are proven.
6. Critical state changes are idempotent.
7. State + event uses Outbox; consumers use Inbox/deduplication.
8. Cross-service workflows use Saga/process managers, not distributed transactions.
9. Synchronous internal calls use gRPC only when an immediate answer is required.
10. Async integration uses NATS JetStream.
11. Every extraction needs compatibility, failure and rollback tests.
12. Docker Compose remains a supported self-host topology; Kubernetes is optional.

## Phase 0 — Product and package rebaseline ✅

Delivered:
- OpenRide product identity and manifesto;
- Request/Quote/Agreement/Ride model;
- packageable `core-go`;
- versioned contracts;
- JS/TS SDK foundation;
- first-party service modules;
- public documentation and open-source governance.

## Phase 1 — Marketplace Service ✅ foundation

Extract Marketplace into `services/marketplace`.

Marketplace owns:
- MobilityRequest;
- DriverTariff;
- Quote;
- ranking orchestration;
- Agreement;
- module catalog/validation;
- Marketplace Outbox/Inbox.

Deliverables:
- independent Go module/entrypoint;
- service-owned migration directory;
- health/readiness endpoints;
- Docker image;
- separate `marketplace_db` in microservices Compose;
- NATS JetStream development dependency.

Compatibility API proxies V2 marketplace traffic instead of importing Marketplace ownership into the gateway.

## Phase 2 — Complete Marketplace persistence/events

Implement durable repositories only inside Marketplace Service.

Required commands:
- create/cancel MobilityRequest;
- create/update DriverTariff;
- create/withdraw Quote;
- accept Quote -> Agreement atomically.

Agreement acceptance transaction must:
1. verify request state;
2. verify quote state/expiry/ownership;
3. snapshot accepted commercial terms;
4. mark competing state appropriately;
5. write Outbox event in the same transaction.

Initial events:

```text
openride.marketplace.request.opened.v1
openride.marketplace.request.cancelled.v1
openride.marketplace.quote.created.v1
openride.marketplace.quote.withdrawn.v1
openride.marketplace.agreement.created.v1
```

## Phase 3 — Location Service

Extract location ownership from compatibility `drivers`/Redis code.

Location Service owns:
- driver online/offline state;
- latest coordinates;
- location freshness;
- GEO index;
- nearby candidate discovery.

Expose internal gRPC such as:

```text
DiscoverCandidates(request/service/area) -> candidate summaries
```

Marketplace must call Location Service, never read its Redis keys directly.

## Phase 4 — Marketplace matching pipeline

Wire Marketplace Core engine to:
- Location gRPC CandidateSource;
- Marketplace-owned tariff/quote repositories;
- explainable ranking policy;
- event publisher/outbox.

Target:

```text
Request
 -> Location candidates
 -> eligibility
 -> driver-authorized quotes
 -> rank many
 -> expose offers
```

Do not use ML initially. Deterministic policies and explanation codes are preferred.

## Phase 5 — Ride Service

Create execution service consuming:

```text
openride.marketplace.agreement.created.v1
```

Ride Service owns:
- Ride/Job execution aggregate;
- service lifecycle state;
- execution cancellation/incidents;
- ride state history;
- Ride Outbox/Inbox.

Passenger lifecycle is one module, not the universal lifecycle.

## Phase 6 — Realtime Service

Extract WebSocket/SSE client sessions and event fanout.

Realtime consumes domain events and streams client-safe views. It does not become durable business truth.

Clients reconnect then resync durable snapshots from owning services through the edge.

## Phase 7 — Identity Service

Extract:
- account authentication;
- OTP/session/token lifecycle;
- identity roles;
- KYC identity status/reference ownership.

The Edge validates/authenticates using Identity-issued credentials without reading Identity DB.

## Phase 8 — Payment Service

Consume Ride events and own:
- payment intents;
- cash/provider payment status;
- captures/refunds;
- earnings;
- operator fees;
- settlement ledger.

Ride and Payment remain independent state machines.

## Phase 9 — Trust Service

Extract:
- ratings;
- reputation;
- reports;
- moderation/safety/fraud signals.

Marketplace may consume trust summaries through API/events but does not own ratings tables.

## Phase 10 — Notification Service

Consume events and own push/SMS/email delivery, retries, templates and provider receipts.

Domain services publish business events; they do not directly embed every notification provider workflow.

## Phase 11 — Operator Service

Build operator read models from integration events and call owning services for commands.

Operator Service must not join Marketplace/Ride/Payment DBs.

## Phase 12 — Rider marketplace UI

Rider flow becomes:

```text
create request
 -> offers arrive
 -> compare price / ETA / trust
 -> select / Quick Match
 -> Agreement
 -> Ride tracking
```

Offer cards show explicit recommendation reasons and never call cheapest "best" by default.

## Phase 13 — Driver pricing UI

Add `My Price / My Tariff` as a primary Driver feature.

Controls include:
- minimum/base fare;
- per-km/per-minute;
- pickup rules;
- manual/auto/hybrid mode;
- automatic quote lower/upper bounds;
- service-specific rules.

## Phase 14 — Retire compatibility ownership

Capability by capability:
- stop writes to legacy tables;
- preserve read projection/history if required;
- remove legacy dispatch/platform fare ownership;
- retire compatibility handlers after client migration;
- finally simplify or replace `services/api` with a dedicated Edge Gateway.

## Data migration rule

Do not move data merely by allowing a new service to read the old database forever.

Use an explicit migration:

```text
legacy snapshot/export
 -> transform
 -> import into owning service DB
 -> dual-write/event bridge only for bounded migration window
 -> compare
 -> cut traffic
 -> remove compatibility write
```

## Rollout controls

Instance-level flags may temporarily control migration:

```text
marketplace_service_enabled
location_service_enabled
v2_quotes_enabled
rider_offer_selection_enabled
driver_tariffs_enabled
quick_match_enabled
```

Feature flags are migration tools, not permanent duplicate ownership.

## Test strategy

Every extracted service requires:
- domain unit tests;
- repository/integration tests against its own DB;
- contract tests;
- idempotency tests;
- Outbox/Inbox tests;
- duplicate event tests;
- timeout/retry tests;
- backward-compatibility tests where traffic is still proxied;
- Docker image/health tests.

Cross-service scenarios require tests for:
- event delay/duplication;
- dependency unavailable;
- Saga compensation/recovery;
- stale location;
- quote expiration races;
- double acceptance attempts;
- payment failure after ride completion.

## Definition of done

Migration is complete when normal passenger traffic no longer depends on the old platform-owned Trip/Pricing/Dispatch kernel and instead flows through independently owned services:

```text
Identity
   |
Edge -> Marketplace -> Location
             |
          Agreement
             |
            NATS
       +-----+-----+
       |           |
      Ride      Notification
       |
      NATS
       |
    Payment / Trust / Operator projections
```

with durable auditability, transparent pricing, independent data ownership and self-hostable deployment.
