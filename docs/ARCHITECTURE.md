# OpenRide System Architecture V2

## 1. Architectural direction

OpenRide keeps the current Go modular-monolith foundation and changes the business kernel from platform-priced dispatch into an open mobility marketplace.

The architecture target is:

```text
Rider Flutter -----------\
                         \
Driver Flutter ------------> OpenRide Go API + WebSocket
                         /           |
Operator Next.js --------/           +--> PostgreSQL + PostGIS
                                     +--> Redis GEO / Cache / Locks / TTL
                                     +--> Object Storage
                                     +--> Background workers
                                     |
                                     +--> Maps / SMS / Push / Payment providers
```

We do not split into microservices by default. Business boundaries must be clean enough to extract later when scale or ownership justifies it.

## 2. Business kernel

The target marketplace flow is:

```text
MobilityRequest
    -> Candidate Discovery
    -> DriverTariff
    -> Quote
    -> Marketplace Ranking
    -> Rider Selection / Quick Match
    -> Agreement
    -> Ride
    -> Payment / Rating
```

The key architectural rule is that `Request`, `Quote`, `Agreement` and `Ride` are different concepts and should not be collapsed into one giant Trip aggregate.

## 3. Target module boundaries

### identity / auth
Authentication, OTP, sessions and authorization primitives.

### users
User profile and account state.

### drivers
Driver profile, KYC status, capabilities, vehicle references and availability.

### tariffs
Driver-owned commercial rules:
- base/minimum fare;
- per-km/per-minute rates;
- pickup rule;
- time/zone/long-distance adjustments;
- manual/auto/hybrid mode;
- automatic quote bounds.

### requests
Mobility demand:
- pickup/destination/stops;
- service type;
- rider constraints/preferences;
- lifecycle and expiry.

### marketplace
Candidate discovery, eligibility and orchestration of request-to-quote discovery.

### quotes
Driver-authorized commercial offers, expiry, withdrawal and acceptance state.

### ranking
Multi-objective ordering of offers using explainable signals such as ETA, price fit, quality and reliability.

### fairness
Exposure/fairness signals used as a bounded ranking input. It must not force rider selection.

### agreements
Immutable accepted commercial snapshot connecting rider, driver, request and quote.

### rides
Execution state after agreement. Service-specific workflows can live behind policies/handlers.

### location
Realtime driver location ingestion, freshness, GEO indexing and ride fan-out.

### payments
Payment transaction state, provider abstraction, refund and settlement hooks.

### ratings / trust
Ratings, reputation, safety reports and trust signals.

### notifications
Push/SMS/email orchestration and delivery jobs.

### operator
Operator/admin use cases, KYC review, support, disputes, policy, RBAC and audit.

### instances
Future self-host/operator boundary. `instance_id` should be introduced before federation is required.

## 4. Data ownership

### PostgreSQL/PostGIS
Durable source of truth for:
- users and driver profiles;
- KYC metadata;
- driver tariffs;
- mobility requests;
- quotes and pricing snapshots;
- agreements;
- rides and status history;
- payment records;
- ratings/reports;
- operator policy and audit.

### Redis
Hot/ephemeral state only:
- online driver GEO index;
- latest locations;
- websocket presence;
- quote/offer TTL;
- matching/accept locks;
- idempotency/rate-limit cache;
- short-lived marketplace state.

Redis must not be the only durable source of truth for accepted commercial terms.

### Object Storage
Private files such as:
- KYC documents;
- avatars;
- driver vehicle images;
- incident/support attachments.

## 5. Request-to-agreement flow

```text
Rider
  -> create MobilityRequest
  -> route/distance enrichment
  -> request becomes OPEN

Marketplace
  -> Redis GEO candidate discovery
  -> eligibility filters
  -> load driver tariff/preferences
  -> generate or request Quote
  -> rank valid quotes
  -> stream offers to Rider

Rider
  -> selects Quote
  -> atomic agreement transaction
     * verify request open
     * verify quote valid/not expired
     * verify driver available
     * lock request/driver
     * create Agreement snapshot
     * mark request AGREED
     * reserve driver
  <- Agreement
```

Quick Match uses the same valid quote set but lets the ranking/constraint engine select on the rider's behalf.

## 6. Pricing architecture

The old global Pricing module becomes a compatibility/recommendation layer during migration.

Target pricing rule:

```text
DriverTariff
+ route facts
+ transparent contextual rules
= Quote
```

Platform/operator policy may enforce visible guardrails such as legal min/max or service-area rules, but it should not silently replace driver-owned pricing.

The quote stores a pricing breakdown and tariff/rule snapshot/version for auditability.

## 7. Ranking architecture

Initial ranking should remain deterministic and explainable.

Example normalized score:

```text
score =
    w_eta         * eta_score
  + w_price_fit   * price_fit_score
  + w_quality     * quality_score
  + w_reliability * reliability_score
  + w_preference  * preference_score
  + w_fairness    * bounded_fairness_score
```

Do not sort solely by fare.

Ranking output should include reason codes such as:
- `FAST_PICKUP`;
- `LOWEST_PRICE`;
- `BEST_OVERALL`;
- `HIGH_RATING`;
- `PREFERRED_VEHICLE`.

## 8. Agreement consistency

Agreement creation is one of the strongest consistency boundaries in the system.

Guards:
- request can only have one accepted agreement unless a service explicitly supports multi-provider jobs;
- quote must belong to request and driver;
- quote must be pending and unexpired;
- driver cannot be reserved for incompatible concurrent work;
- accepted price is snapshotted;
- endpoint is idempotent;
- database conditional updates are source of truth;
- Redis locks reduce races but do not replace database invariants.

## 9. Location flow

```text
Driver GPS
 -> app
 -> WebSocket/HTTP ingestion
 -> timestamp/accuracy validation
 -> Redis latest location + GEO index
 -> active ride fan-out to Rider room
 -> optional durable sampling for support/analytics
```

Do not write every raw GPS ping into primary transactional tables.

## 10. Background processing

Workers handle non-blocking tasks:
- push notifications;
- quote expiry cleanup/reconciliation;
- scheduled request activation;
- receipts;
- settlement/earning computation;
- analytics events;
- provider callback retries.

A Redis-backed queue is sufficient initially. NATS/Kafka should only be introduced after real event topology or throughput requires it.

## 11. Failure behavior

### Redis unavailable
- do not create unsafe new matches/agreements that depend on unavailable locks/geo state;
- durable requests, quotes, agreements and rides remain queryable from PostgreSQL;
- active rides degrade location freshness safely.

### Routing provider unavailable
- new route-dependent quote generation should fail clearly or use an explicitly configured fallback;
- never fabricate a commercial agreement from unknown distance data.

### WebSocket disconnect
- mobile reconnects with backoff;
- after reconnect, fetch durable snapshot over REST then resume realtime events.

### Payment failure
- Ride and Payment remain separate state machines;
- a completed ride may have pending/failed payment requiring recovery.

## 12. Instance/self-host path

OpenRide should prepare for self-hosted operator instances without implementing federation too early.

Target boundary:

```text
OpenRide Core
  ├── Instance A / local operator
  ├── Instance B / cooperative
  └── Instance C / community
```

Core marketplace-owned rows should eventually carry `instance_id` and provider configuration should remain abstracted.

## 13. Scale path

Possible extraction order when justified:
1. Location Service.
2. Realtime Gateway.
3. Marketplace Matching/Ranking Service.
4. Notification Worker.
5. Payment/Settlement Service.
6. Analytics pipeline.

Do not extract solely for architecture aesthetics.

## 14. Technology rules

Initial target remains:
- Go backend;
- Flutter rider/driver apps;
- Next.js operator web;
- PostgreSQL + PostGIS;
- Redis;
- WebSocket;
- S3-compatible object storage.

Avoid by default in early phases:
- Kubernetes;
- service mesh;
- global event sourcing;
- complex CQRS;
- multi-region active-active;
- ML-based ranking before trustworthy data exists.

## 15. Migration compatibility

The repository currently contains `trips`, `pricing`, `dispatch`, designated-driver and inspection semantics. These remain compatibility/runtime modules while marketplace replacements are introduced.

See [`OPENRIDE_MIGRATION_PLAN_V2.md`](./OPENRIDE_MIGRATION_PLAN_V2.md).
