# OpenRide Data Model V2

## 1. Principle: data belongs to a service

OpenRide does not have one global application database model.

Each bounded-context service owns its data and migrations:

```text
identity_db       -> Identity Service
marketplace_db    -> Marketplace Service
location_db       -> Location Service
ride_db           -> Ride Service
payment_db        -> Payment Service
trust_db          -> Trust Service
operator_db       -> Operator Service projections
```

A small deployment may run these logical databases on one PostgreSQL cluster. Services still use separate ownership boundaries and must not query each other's tables.

Cross-service identifiers are opaque references, **not SQL foreign keys across databases**.

## 2. Universal storage rules

- money uses integer minor units; never floating point;
- timestamps are UTC in storage;
- business aggregates use optimistic versioning where useful;
- critical command deduplication/idempotency is durable where required;
- service events use Outbox in the same transaction as state changes;
- consumers use Inbox/deduplication before side effects;
- Redis is owned hot state, not a shortcut to another service's durable truth;
- schema migrations live with the service that owns the schema.

## 3. Marketplace Service data

Marketplace currently owns the commercial marketplace chain:

```text
driver_tariffs
      |
mobility_requests
      |
marketplace_quotes
      |
marketplace_agreements

outbox_events
inbox_messages
```

### driver_tariffs

Suggested model:

```text
id
instance_id
driver_id
service_type
quote_mode              manual | auto | hybrid
currency
base_fare_minor
minimum_fare_minor
per_km_minor
per_minute_minor
pickup_fee_minor
auto_quote_min_minor
auto_quote_max_minor
rules_json
status
version
effective_from
effective_to
created_at
updated_at
```

Marketplace owns these commercial terms. Other services receive only the summaries/events they need.

### mobility_requests

```text
id
instance_id
rider_id
service_type
status
pickup_lat
pickup_lng
destination_lat
destination_lng
attributes_json
constraints_json
requested_at
expires_at
agreed_at
cancelled_at
version
created_at
updated_at
```

Marketplace does not store live driver GEO indexes. Candidate discovery belongs to Location Service.

### marketplace_quotes

```text
id
request_id
driver_id
driver_vehicle_id
tariff_id
tariff_version
status
fare_total_minor
currency
pricing_snapshot_json
pickup_eta_s
pickup_distance_m
explanation_json
expires_at
created_at
accepted_at
withdrawn_at
version
```

Quotes snapshot enough tariff/pricing information to audit why the offer existed.

### marketplace_agreements

```text
id
instance_id
request_id
quote_id
rider_id
driver_id
driver_vehicle_id
service_type
fare_total_minor
currency
terms_snapshot_json
created_at
```

Agreement is immutable accepted commercial truth.

It should have constraints preventing accidental double acceptance for service types where a request has only one provider.

### Marketplace Outbox

```text
id
event_name
aggregate_type
aggregate_id
instance_id
correlation_id
causation_id
payload_json
occurred_at
published_at
attempt_count
last_error
```

The Outbox row is inserted in the same DB transaction as the Marketplace state mutation.

### Marketplace Inbox

```text
message_id
consumer
received_at
processed_at
result_status
```

Used to safely process redelivered NATS messages.

## 4. Location Service data

Location owns realtime supply location.

Durable model may include:

```text
driver_presence
location_history_sample
service_area
```

Hot state may use Redis:

```text
online driver GEO index
latest driver position
presence TTL
location freshness
```

Only Location Service owns these Redis keys.

Marketplace obtains candidate summaries through Location gRPC, not Redis access.

## 5. Ride Service data

Ride owns execution after Agreement.

```text
rides
ride_status_history
ride_incidents
outbox_events
inbox_messages
```

Suggested `rides`:

```text
id
instance_id
agreement_id
marketplace_request_id
rider_id
driver_id
service_type
status
service_state_json
started_at
completed_at
cancelled_at
version
created_at
updated_at
```

`agreement_id` is a reference copied from the integration event, not a foreign key to Marketplace DB.

Service-specific lifecycle details can live in typed module contracts/state JSON until promoted into universal fields.

## 6. Identity Service data

Identity owns:

```text
accounts
identities
sessions
refresh_tokens
roles
kyc_identity_status
outbox_events
inbox_messages
```

Driver operational/commercial data should not be hidden inside authentication tables.

## 7. Payment Service data

Payment owns:

```text
payment_intents
payment_transactions
refunds
driver_earnings
operator_fees
settlement_ledger
outbox_events
inbox_messages
```

Payment references `ride_id`/`agreement_id` as external IDs, never through a DB FK into Ride/Marketplace databases.

A completed Ride can coexist with pending/failed Payment state.

## 8. Trust Service data

Trust owns:

```text
ratings
reputation_snapshots
reports
safety_events
fraud_signals
moderation_actions
outbox_events
inbox_messages
```

Marketplace consumes only required trust summaries through an API or projection/event contract.

## 9. Operator Service data

Operator Service primarily owns projections/read models and operator workflow state:

```text
operator_marketplace_projection
operator_ride_projection
operator_payment_projection
support_cases
operator_audit_log
inbox_messages
```

These projections are built from events/APIs.

The Operator Service must not join service databases directly to build a dashboard.

## 10. Instance/operator boundary

Every business service should propagate `instance_id` on owned aggregates/events where operator isolation is relevant.

An Instance can define:
- country/currency/timezone;
- service area references;
- enabled service types;
- legal/operator guardrails;
- payment/provider configuration references;
- ranking policy configuration.

Federation is not required for the initial architecture, but instance isolation must not require a future schema rewrite.

## 11. Service type identifiers

Canonical service IDs use dotted names:

```text
passenger.car
carpool.intercity
passenger.motorbike
parcel.instant
designated-driver.car
```

Stored/requested service identifiers should not alternate between `passenger_car`, `passenger-car` and `passenger.car`.

## 12. Legacy compatibility data

The existing compatibility runtime still owns old tables such as:
- users;
- drivers;
- customer vehicles;
- trips;
- trip status history;
- pricing rules;
- payments;
- ratings;
- driver documents.

These are not the target shared schema.

When a capability is extracted:

```text
legacy export/snapshot
 -> transform
 -> import into owning service DB
 -> bounded bridge/dual-write if required
 -> compare
 -> cut traffic
 -> stop legacy writes
```

A new service must not simply retain permanent read access to the legacy database.

## 13. Indexing examples

Marketplace:
- `driver_tariffs(instance_id, driver_id, service_type, status)`;
- `mobility_requests(instance_id, status, service_type, requested_at)`;
- `marketplace_quotes(request_id, status, expires_at)`;
- unique/conditional constraints for Agreement acceptance.

Location:
- geo indexes owned by Location;
- freshness/online lookups optimized independently.

Ride:
- `(instance_id, driver_id, status)`;
- `(instance_id, rider_id, created_at)`;
- `(agreement_id)` unique where one execution per agreement applies.

## 14. Auditability rule

Given an Agreement, OpenRide must be able to answer:

```text
what the rider requested
which driver quote was accepted
what price/currency was accepted
which tariff/version was involved
when the agreement happened
which service type governed execution
```

without recomputing commercial history from current pricing configuration.
