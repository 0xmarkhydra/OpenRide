# OpenRide Data Model V2

## 1. Storage principles

PostgreSQL + PostGIS is the durable source of truth for marketplace/business state.

Redis stores hot/ephemeral state such as:
- online driver geo indexes;
- latest locations;
- websocket presence;
- quote/offer TTL helpers;
- matching locks;
- short-lived idempotency/rate-limit state.

Accepted commercial terms must never exist only in Redis.

Money is stored as integer minor units / integer currency units according to one documented project convention. Never use float for fares.

## 2. Current compatibility schema

The repository currently contains legacy tables such as:
- users;
- drivers;
- vehicles/customer_vehicles;
- trips;
- trip_status_history;
- pricing_rules;
- payments;
- ratings;
- driver documents.

These remain operational while V2 marketplace tables are added through append-only migrations.

## 3. Target entity relationship

```text
users
  |
  +-- rider_profiles
  |
  +-- driver_profiles
          |
          +-- driver_vehicles
          +-- driver_capabilities
          +-- driver_tariffs
                  |
                  +-- driver_tariff_rules

mobility_requests
      |
      +-- quotes -------- driver_profiles
      |      |
      |      +-- quote_components / pricing snapshot
      |
      +-- agreements
              |
              +-- rides
                    |
                    +-- ride_status_history

agreements / rides
      |
      +-- payments
      +-- driver_earnings
      +-- platform_fees
      +-- ratings
      +-- safety_events / disputes

instances
  +-- policies / service config
```

## 4. instances

Prepare for self-host/operator boundaries before federation is needed.

Suggested fields:

```text
id UUID/TEXT PK
slug UNIQUE
name
status
currency
country_code
timezone
created_at
updated_at
```

Policy/config should not become an unstructured dumping ground. Keep sensitive business concepts in explicit tables where practical.

## 5. users

Suggested core:

```text
id
phone/email identity
full_name
status
phone_verified_at
avatar_object_key
last_login_at
created_at
updated_at
version
```

Roles should not require separate duplicated identities for rider and driver.

## 6. driver_profiles

Suggested fields:

```text
id
user_id UNIQUE
instance_id
approval_status
availability_status
rating_avg
rating_count
completion_rate
approved_at
suspended_at
last_idle_at
created_at
updated_at
version
```

Indexes should support instance/status lookups.

## 7. driver_vehicles

For passenger ride verticals:

```text
id
driver_id
instance_id
type
plate_number
brand
model
year
color
seats
status
document_status
created_at
updated_at
```

Unique constraints depend on local/operator rules, typically plate within instance.

Customer-owned vehicles remain a separate concept for designated-driver/inspection verticals.

## 8. driver_capabilities

Prefer normalized rows once capabilities become richer than a tiny array.

```text
id
driver_id
service_type
status
metadata JSONB
created_at
updated_at
```

Examples:
- passenger_car;
- passenger_motorbike;
- designated_driver_car;
- delivery;
- vehicle_inspection_assist.

## 9. driver_tariffs

First-class driver-owned commercial profile.

```text
id
driver_id
instance_id
service_type
quote_mode            # manual | auto | hybrid
currency
base_fare_minor
minimum_fare_minor
per_km_minor
per_minute_minor
pickup_fee_minor
auto_quote_min_minor
auto_quote_max_minor
status
version
effective_from
effective_to
created_at
updated_at
```

Important indexes:
- `(driver_id, service_type, status)`;
- `(instance_id, service_type, status)`.

A driver should not have ambiguous overlapping active tariffs for the same service unless product rules explicitly support them.

## 10. driver_tariff_rules

Optional richer rules:

```text
id
tariff_id
rule_type
priority
conditions JSONB
adjustment_type
adjustment_value
status
created_at
updated_at
```

Examples:
- night surcharge;
- holiday adjustment;
- long-distance discount;
- zone fee;
- pickup-distance rule.

Rule evaluation must be deterministic by tariff/rule version.

## 11. mobility_requests

```text
id
instance_id
rider_id
service_type
status
pickup GEOGRAPHY(POINT,4326)
destination GEOGRAPHY(POINT,4326)
stops JSONB or normalized table
estimated_distance_m
estimated_duration_s
preferences JSONB
constraints JSONB
requested_at
expires_at
agreed_at
cancelled_at
created_at
updated_at
version
```

Suggested indexes:
- GIST pickup/destination;
- `(instance_id, status, created_at DESC)`;
- expiry index for open requests.

State values initially:
- draft;
- open;
- receiving_quotes;
- agreed;
- closed;
- cancelled;
- expired.

## 12. quotes

```text
id
request_id
driver_id
driver_vehicle_id nullable
tariff_id nullable
tariff_version nullable
status
fare_total_minor
currency
pricing_snapshot JSONB
pickup_distance_m
pickup_eta_s
expires_at
created_at
updated_at
accepted_at
withdrawn_at
```

Important constraints:
- quote request/driver foreign keys;
- accepted quote must be durable;
- no float fare;
- request + driver may have one current active quote unless negotiation design says otherwise.

Indexes:
- `(request_id, status, created_at)`;
- `(driver_id, status, created_at DESC)`;
- `(expires_at)` for pending quote cleanup.

## 13. quote_components

May be JSONB in the first version or normalized later.

Expected explainable components:
- base;
- distance;
- duration;
- pickup;
- time rule;
- zone rule;
- service adjustment;
- discount;
- total.

The snapshot should carry the tariff/rule version and calculation reason codes.

## 14. agreements

This is the immutable commercial source of truth after selection.

```text
id
instance_id
request_id UNIQUE
quote_id UNIQUE
rider_id
driver_id
driver_vehicle_id nullable
service_type
pickup_snapshot JSONB
destination_snapshot JSONB
route_snapshot JSONB
fare_total_minor
currency
pricing_snapshot JSONB
terms_snapshot JSONB
policy_version
created_at
```

For normal single-provider rides, `request_id UNIQUE` guarantees only one accepted commercial agreement.

If future service types require multiple providers, model that explicitly instead of weakening this invariant globally.

Agreement rows should be immutable except for carefully scoped administrative metadata if ever necessary.

## 15. rides

```text
id
agreement_id UNIQUE
instance_id
rider_id
driver_id
service_type
status
started_at
completed_at
cancelled_at
created_at
updated_at
version
```

Do not duplicate commercial fare as mutable ride pricing. Read accepted terms from Agreement or store a read-only projection.

## 16. ride_status_history

Append-only:

```text
id
ride_id
sequence
status
actor_type
actor_id nullable
reason_code
metadata JSONB
created_at
```

Unique `(ride_id, sequence)`.

## 17. latest driver location

Redis hot state:

```text
driver:{instance}:{driver_id}:location
```

Fields:
- lat;
- lng;
- accuracy;
- heading;
- speed;
- captured_at;
- received_at.

GEO index:

```text
geo:drivers:{instance}:{city}:{service_type}
```

Stale-location policy is mandatory.

## 18. ride path persistence

Do not store every GPS ping in transactional ride rows.

Options when needed:
- time/distance sampling;
- partitioned location table;
- encoded polyline after ride;
- analytics/time-series store at larger scale.

Retention must be tied to support, safety and legal/privacy requirements.

## 19. payments

Suggested:

```text
id
agreement_id / ride_id
provider
external_reference
amount_minor
currency
status
idempotency_key
created_at
updated_at
```

Payment state remains separate from Ride.

## 20. driver_earnings and platform_fees

Make deductions transparent.

`driver_earnings`:
- agreement_id;
- gross_fare;
- deductions;
- net_driver_earning;
- settlement status.

`platform_fees`:
- agreement_id;
- fee type;
- amount;
- policy version.

Avoid opaque calculations that cannot be explained to drivers.

## 21. ratings / trust / safety

Suggested tables/concepts:
- ratings;
- reports;
- safety_events;
- disputes;
- moderation_actions;
- audit_logs.

Sensitive operator changes need actor/reason/time metadata.

## 22. Redis marketplace keys

Possible shapes:

```text
request:{request_id}:presence
quote:{quote_id}:ttl
lock:request:{request_id}:agreement
lock:driver:{driver_id}:assignment
idem:{scope}:{key}
```

Redis TTL expiration is not enough to update durable quote status; background reconciliation/query normalization may mark persisted pending quotes expired.

## 23. Migration rules

- migrations are append-only;
- do not rewrite migrations already applied in production;
- add nullable/backward-compatible columns/tables first;
- backfill separately;
- add strong constraints after data is ready;
- use concurrent indexes when production table size/locking requires it;
- keep compatibility reads/writes until old consumers migrate.

## 24. Proposed migration sequence

```text
007_instances.sql
008_mobility_requests.sql
009_driver_tariffs.sql
010_quotes.sql
011_agreements.sql
012_rides.sql
013_marketplace_indexes.sql
014_earnings_fees.sql
```

Exact numbering should be based on repository state when implementation begins; never renumber migrations already shipped.

## 25. Backup, retention and privacy

Production requires:
- automated PostgreSQL backups;
- tested restore procedure;
- documented KYC retention;
- location-history retention policy;
- audit retention;
- object-storage access policy;
- deletion/anonymization policy compatible with applicable law.
