# OpenRide API Contract V2

> Public V2 contract exposed through the Edge Gateway/BFF. Internal service topology is not a client concern.

## 1. API principles

- REST for durable public commands/queries.
- WebSocket/SSE for realtime delivery where appropriate.
- Critical create/accept/payment commands require idempotency keys.
- Clients resync durable snapshots after reconnect.
- Accepted commercial terms come from Agreement snapshots, never client-side recomputation.
- Canonical service type IDs use dotted names such as `passenger.car`.
- Public errors use stable codes; internal service errors are not leaked directly.

## 2. Service catalog

```text
GET /v2/services
```

Returns service manifests from Marketplace Service through the Edge.

Example:

```json
{
  "data": [
    {
      "id": "passenger.car",
      "version": "1.0.0",
      "display_name": "Passenger Car",
      "category": "passenger",
      "capabilities": ["passenger", "scheduled"]
    },
    {
      "id": "carpool.intercity",
      "version": "1.0.0",
      "display_name": "Intercity Carpool",
      "category": "carpool",
      "capabilities": ["passenger", "carpool", "scheduled"]
    }
  ]
}
```

## 3. Authentication

Representative public endpoints:

```text
POST /v2/auth/otp/request
POST /v2/auth/otp/verify
POST /v2/auth/refresh
POST /v2/auth/logout
GET  /v2/me
```

Identity Service owns authentication/session truth; Edge exposes the public contract.

## 4. Driver tariffs

```text
GET   /v2/drivers/me/tariffs
POST  /v2/drivers/me/tariffs
PATCH /v2/drivers/me/tariffs/{tariff_id}
GET   /v2/drivers/me/tariffs/{tariff_id}/rules
POST  /v2/drivers/me/tariffs/{tariff_id}/rules
```

Create example:

```json
{
  "service_type": "passenger.car",
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

Create/update commands require:

```text
Idempotency-Key: <key>
```

Updates should use optimistic versioning (`If-Match` or explicit version field).

Marketplace Service owns tariff data.

## 5. Route preview

```text
POST /v2/routes/preview
```

Route preview returns route facts without creating commercial terms.

Example request:

```json
{
  "service_type": "passenger.car",
  "pickup": {"lat": 19.8067, "lng": 105.7852},
  "destination": {"lat": 19.7724, "lng": 105.7762}
}
```

Route-provider failure must be explicit. Never fabricate distance for commercial agreement creation.

## 6. Mobility requests

### Create

```text
POST /v2/requests
Idempotency-Key: <key>
```

```json
{
  "service_type": "passenger.car",
  "pickup": {"lat": 19.8067, "lng": 105.7852},
  "destination": {"lat": 19.7724, "lng": 105.7762},
  "attributes": {},
  "constraints": {
    "max_fare_minor": 70000,
    "max_pickup_eta_s": 900
  }
}
```

Response:

```json
{
  "data": {
    "id": "req_...",
    "status": "open",
    "service_type": "passenger.car",
    "expires_at": "2026-09-09T12:00:00Z",
    "version": 1
  }
}
```

### Snapshot

```text
GET /v2/requests/{request_id}
```

Includes durable request state, selectable offer references and accepted Agreement reference when applicable.

### Cancel

```text
POST /v2/requests/{request_id}/cancel
Idempotency-Key: <key>
```

After Agreement creation, cancellation routes through execution/cancellation policy instead of freely rewriting the original request.

## 7. Driver request discovery

```text
GET /v2/drivers/me/requests
GET /v2/drivers/me/requests/{request_id}
```

For manual/hybrid quoting, driver request cards may expose only information necessary to decide whether to quote:
- service type;
- route summary;
- pickup distance/ETA;
- allowed quote bounds;
- transparent market suggestion if enabled.

Location-derived candidate facts come from Location Service through Marketplace orchestration.

## 8. Quotes

### Submit manual/hybrid quote

```text
POST /v2/requests/{request_id}/quotes
Idempotency-Key: <key>
```

```json
{
  "fare_total_minor": 52000,
  "currency": "VND"
}
```

Marketplace validates:
- driver/request eligibility;
- request state;
- tariff/quote mode;
- visible operator/legal guardrails;
- quote expiry;
- currency and amount invariants.

### Driver quote list

```text
GET /v2/drivers/me/quotes?status=pending
```

### Withdraw

```text
POST /v2/quotes/{quote_id}/withdraw
Idempotency-Key: <key>
```

### Rider offers

```text
GET /v2/requests/{request_id}/offers
```

Example:

```json
{
  "data": [
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
      "expires_at": "2026-09-09T12:00:00Z",
      "rank": 1,
      "reasons": ["BEST_OVERALL", "FAST_PICKUP"]
    }
  ]
}
```

Cheapest is not automatically `BEST_OVERALL`.

## 9. Agreement

### Accept quote

```text
POST /v2/quotes/{quote_id}/accept
Idempotency-Key: <key>
```

Agreement creation is a strong atomic command inside Marketplace Service.

```json
{
  "data": {
    "agreement": {
      "id": "agr_...",
      "request_id": "req_...",
      "quote_id": "quote_...",
      "driver_id": "drv_...",
      "rider_id": "usr_...",
      "service_type": "passenger.car",
      "fare_total_minor": 52000,
      "currency": "VND",
      "terms_snapshot": {},
      "created_at": "2026-09-09T12:00:00Z"
    }
  }
}
```

The command must be idempotent and protect against expired/double accepted quotes.

Marketplace publishes `openride.marketplace.agreement.created.v1` through Outbox after commit.

## 10. Ride execution

Representative public APIs:

```text
GET  /v2/rides/{ride_id}
POST /v2/rides/{ride_id}/commands/{command}
```

The exact command set depends on the registered service lifecycle.

Passenger examples may include:

```text
driver-en-route
arrived
passenger-onboard
start
complete
cancel
```

Carpool or designated-driver modules may expose different commands/states.

Ride Service owns execution state; Edge maps public commands to the correct service contract.

## 11. Realtime

Public realtime transport may expose:

```text
GET /v2/realtime
```

Client-visible event examples:

```text
request.updated
quote.created
quote.withdrawn
agreement.created
ride.status_changed
payment.updated
```

These are client transport events, not necessarily identical to internal NATS subjects.

After reconnect:
1. re-authenticate;
2. fetch durable snapshot via REST;
3. resume realtime stream.

## 12. Payments

Representative APIs:

```text
GET  /v2/rides/{ride_id}/payment
POST /v2/payments/{payment_id}/commands/{command}
```

Payment Service owns payment truth. Ride completion does not imply successful payment.

## 13. Error envelope

```json
{
  "error": {
    "code": "QUOTE_EXPIRED",
    "message": "The selected quote is no longer available",
    "details": {}
  }
}
```

Stable public error examples:

```text
INVALID_REQUEST
UNAUTHORIZED
FORBIDDEN
SERVICE_TYPE_UNSUPPORTED
REQUEST_NOT_FOUND
REQUEST_INVALID_STATE
QUOTE_NOT_FOUND
QUOTE_EXPIRED
QUOTE_INVALID_STATE
IDEMPOTENCY_KEY_REQUIRED
IDEMPOTENCY_CONFLICT
DEPENDENCY_UNAVAILABLE
RATE_LIMITED
INTERNAL_ERROR
```

Do not leak internal SQL/NATS/gRPC provider details to clients.

## 14. Correlation and tracing

Edge accepts or generates:

```text
X-Request-ID
```

Internal calls/events propagate correlation ID and causation ID where applicable.

## 15. Compatibility

Legacy V1 endpoints remain during extraction.

The Edge may proxy a V2 route to Marketplace/Location/Ride/etc. Clients must not depend on which internal service currently handles a route.

A route moves from compatibility runtime to its owning service without changing its public meaning unless the API contract itself is explicitly versioned.
