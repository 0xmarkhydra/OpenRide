# OpenRide API Contract V2 — Marketplace Target

> Target contract for the marketplace migration. Existing V1 endpoints remain compatibility APIs until migration phases are implemented and tested.

## 1. API principles

- REST for durable command/query flows.
- WebSocket for realtime request/quote/ride events.
- Idempotency keys on critical create/accept/payment commands.
- Agreement creation must be atomic.
- Accepted commercial terms are returned from durable Agreement data, not recomputed client-side.
- Clients must be able to resync a durable snapshot after realtime disconnect.

## 2. Authentication

Existing auth/session infrastructure remains in place.

Representative endpoints:

```text
POST /v2/auth/otp/request
POST /v2/auth/otp/verify
POST /v2/auth/refresh
GET  /v2/me
```

## 3. Driver tariff APIs

### Get active tariffs

```text
GET /v2/drivers/me/tariffs
```

### Create tariff

```text
POST /v2/drivers/me/tariffs
Idempotency-Key: <key>
```

Example request:

```json
{
  "service_type": "passenger_car",
  "quote_mode": "auto",
  "currency": "VND",
  "base_fare_minor": 0,
  "minimum_fare_minor": 20000,
  "per_km_minor": 5000,
  "per_minute_minor": 0,
  "pickup_fee_minor": 0,
  "auto_quote_min_minor": 20000,
  "auto_quote_max_minor": 150000
}
```

### Update tariff

```text
PATCH /v2/drivers/me/tariffs/{tariff_id}
If-Match: <version>
```

### Tariff rules

```text
GET  /v2/drivers/me/tariffs/{tariff_id}/rules
POST /v2/drivers/me/tariffs/{tariff_id}/rules
```

Rules must be validated and deterministic.

## 4. Mobility request APIs

### Route preview

```text
POST /v2/routes/preview
```

Returns route distance/duration/provider metadata without creating commercial terms.

### Create request

```text
POST /v2/requests
Idempotency-Key: <key>
```

Example:

```json
{
  "service_type": "passenger_car",
  "pickup": {"lat": 19.8067, "lng": 105.7852},
  "destination": {"lat": 19.7724, "lng": 105.7762},
  "preferences": {},
  "constraints": {
    "max_fare_minor": 70000,
    "max_pickup_eta_s": 900
  }
}
```

Example response:

```json
{
  "id": "req_...",
  "status": "open",
  "service_type": "passenger_car",
  "route": {
    "distance_m": 10000,
    "duration_s": 1500
  },
  "expires_at": "...",
  "version": 1
}
```

### Get request snapshot

```text
GET /v2/requests/{request_id}
```

Snapshot includes request status, current selectable offers and accepted agreement reference when applicable.

### Cancel request

```text
POST /v2/requests/{request_id}/cancel
Idempotency-Key: <key>
```

Cancellation after agreement must route to ride/agreement cancellation policy instead of mutating the original request freely.

## 5. Driver request discovery

### Current eligible requests

```text
GET /v2/drivers/me/requests
```

For manual/hybrid quote workflows.

Returned request cards should include only data necessary for a driver to decide whether to quote, respecting privacy policy.

### Get request detail for driver

```text
GET /v2/drivers/me/requests/{request_id}
```

May include:
- service type;
- pickup/destination summary;
- route distance/duration;
- driver-to-pickup distance/ETA;
- allowed quote bounds;
- market suggestion when enabled and clearly identified as suggestion.

## 6. Quote APIs

### Submit manual/hybrid quote

```text
POST /v2/requests/{request_id}/quotes
Idempotency-Key: <key>
```

Example:

```json
{
  "fare_total_minor": 52000,
  "currency": "VND"
}
```

Backend validates:
- driver/request eligibility;
- request open state;
- driver tariff/quote mode policy;
- operator/legal guardrails;
- quote TTL.

### Driver current quotes

```text
GET /v2/drivers/me/quotes?status=pending
```

### Withdraw quote

```text
POST /v2/quotes/{quote_id}/withdraw
Idempotency-Key: <key>
```

Only the quote owner can withdraw a valid pending quote.

### Rider offers

```text
GET /v2/requests/{request_id}/offers
```

Example item:

```json
{
  "quote_id": "quote_...",
  "driver": {
    "id": "drv_...",
    "display_name": "...",
    "rating": 4.9
  },
  "vehicle": {
    "type": "car",
    "brand": "Toyota",
    "model": "Vios"
  },
  "fare_total_minor": 52000,
  "currency": "VND",
  "pickup_eta_s": 360,
  "pickup_distance_m": 2100,
  "expires_at": "...",
  "rank": 1,
  "reasons": ["BEST_OVERALL", "FAST_PICKUP"]
}
```

## 7. Agreement APIs

### Accept quote

```text
POST /v2/quotes/{quote_id}/accept
Idempotency-Key: <key>
```

This is a critical atomic command.

Success returns the created/reused Agreement:

```json
{
  "agreement": {
    "id": "agr_...",
    "request_id": "req_...",
    "quote_id": "quote_...",
    "driver_id": "drv_...",
    "rider_id": "usr_...",
    "fare_total_minor": 52000,
    "currency": "VND",
    "pricing_snapshot": {},
    "created_at": "..."
  },
  "ride_id": "ride_..."
}
```

Expected business errors include:
- REQUEST_NOT_OPEN;
- QUOTE_EXPIRED;
- QUOTE_UNAVAILABLE;
- DRIVER_UNAVAILABLE;
- AGREEMENT_ALREADY_EXISTS;
- VERSION_CONFLICT.

Idempotent retries must return the same successful agreement when the same command was already committed.

### Get agreement

```text
GET /v2/agreements/{agreement_id}
```

Agreement price/terms are immutable commercial snapshot data.

## 8. Quick Match API

```text
POST /v2/requests/{request_id}/quick-match
Idempotency-Key: <key>
```

Example constraints:

```json
{
  "max_fare_minor": 60000,
  "max_pickup_eta_s": 480,
  "minimum_rating": 4.7
}
```

Quick Match selects from current valid quotes satisfying rider constraints and performs the same Agreement transaction as explicit quote acceptance.

It must not generate a platform-owned replacement fare.

## 9. Ride APIs

### Get ride

```text
GET /v2/rides/{ride_id}
```

### Driver transitions

Representative commands:

```text
POST /v2/rides/{ride_id}/en-route
POST /v2/rides/{ride_id}/arrived
POST /v2/rides/{ride_id}/passenger-onboard
POST /v2/rides/{ride_id}/start
POST /v2/rides/{ride_id}/complete
```

Each command:
- checks authenticated actor;
- checks state transition;
- is idempotent where practical;
- records durable status history.

### Cancellation

```text
POST /v2/rides/{ride_id}/cancel
```

Cancellation policy may reference Agreement terms and operator policy version.

## 10. Location and realtime

### Driver location ingestion

```text
POST /v2/drivers/me/location
```

or authenticated WebSocket messages.

Payload includes:
- lat/lng;
- accuracy;
- heading/speed when available;
- captured_at.

Backend applies freshness/accuracy validation and updates Redis GEO/latest state.

## 11. WebSocket event model

Representative events:

```text
request.updated
quote.created
quote.updated
offers.snapshot
agreement.created
ride.updated
driver.location
payment.updated
```

Every event should carry:
- event type;
- entity id;
- timestamp;
- entity version/sequence when relevant.

Clients must tolerate duplicate/reordered realtime events by resyncing durable snapshots.

## 12. Operator APIs

Target groups:

```text
/v2/operator/marketplace/*
/v2/operator/drivers/*
/v2/operator/kyc/*
/v2/operator/rides/*
/v2/operator/disputes/*
/v2/operator/policies/*
/v2/operator/ranking/*
/v2/operator/payments/*
```

Sensitive actions require RBAC and audit.

Operator price policy APIs should manage transparent guardrails/recommendations/fees, not a mandatory hidden global driver fare.

## 13. Compatibility

Current V1 trip/pricing/dispatch endpoints remain available during migration.

Feature flags or instance configuration determine whether a client uses:
- legacy trip dispatch;
- V2 marketplace quotes;
- V2 rider offer selection;
- V2 quick match.

No legacy endpoint should be removed until the corresponding V2 client and tests are ready.
