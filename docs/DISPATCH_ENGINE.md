# OpenRide Marketplace Matching Engine

> Historical filename retained for compatibility. This document supersedes the old one-driver dispatch model.

## 1. Goal

The matching engine does not exist to choose one driver and push a platform-owned fare.

Its target responsibilities are:

```text
discover candidates
-> check eligibility
-> obtain/generate driver-authorized quotes
-> rank valid offers
-> expose meaningful choices
-> create agreement safely when selected
```

## 2. Inputs

Typical request context:
- request_id;
- pickup/destination/stops;
- service_type;
- route distance/duration;
- city/zone/instance;
- rider preferences and constraints;
- candidate search policy.

## 3. Candidate discovery

Redis GEO remains the primary hot index for online driver discovery.

Suggested key shape:

```text
geo:drivers:{instance}:{city}:{service_type}
```

Flow:

```text
Request OPEN
 -> GEOSEARCH radius R1
 -> filter stale/ineligible drivers
 -> expand R2/R3 if needed
 -> produce candidate set
```

Do not scan PostgreSQL driver tables for every realtime request.

## 4. Eligibility filters

A driver is excluded when:
- KYC/approval invalid;
- account suspended;
- offline;
- location stale;
- wrong capability/service type;
- wrong service region/instance;
- required vehicle invalid;
- already reserved/busy for incompatible work;
- tariff/quote mode cannot serve the request;
- request violates driver-defined hard constraints.

## 5. Quote generation

Candidate discovery and pricing are separate concerns.

### Auto mode

System generates a quote using the driver's active tariff and explicit auto-quote bounds.

### Manual mode

The request is delivered to the driver. Driver enters/submits a quote.

### Hybrid mode

System proposes a quote from driver tariff. Driver may edit it before submission when policy/time permits.

A quote must retain enough information to explain:
- route facts;
- tariff version;
- rule adjustments;
- final total.

## 6. Ranking

Ranking is multi-objective and higher-is-better after normalization.

Example:

```text
score =
    w_eta         * eta_score
  + w_price_fit   * price_fit_score
  + w_quality     * quality_score
  + w_reliability * reliability_score
  + w_preference  * preference_score
  + w_fairness    * fairness_score
```

Initial implementation should be deterministic.

Do not use ML until there is trustworthy production data and clear observability.

## 7. Price fit is not cheapest wins

`price_fit_score` measures how suitable an offer is relative to rider constraints/market context, not simply whether it is the minimum fare.

UI/ranking should be able to label different offers:
- CHEAPEST;
- FASTEST_PICKUP;
- BEST_OVERALL;
- HIGH_RATING.

The cheapest offer must not automatically receive `BEST_OVERALL`.

## 8. Fairness

Fairness is a bounded exposure signal.

It may help avoid repeatedly hiding otherwise-qualified drivers when other scores are nearly equal.

It must not:
- force riders to choose a driver;
- override safety/eligibility;
- become a hidden quota system;
- overpower large ETA/quality differences.

Fairness logic and weight changes must be observable and auditable.

## 9. Explainability

Every ranked result should expose reason codes suitable for API/UI use.

Example:

```json
{
  "rank": 1,
  "score": 0.91,
  "reasons": ["BEST_OVERALL", "FAST_PICKUP", "HIGH_RATING"]
}
```

Internal debug output may include component scores and configuration version.

## 10. Offer delivery strategy

The old sequential single-offer model is replaced by configurable market exposure.

Possible strategies:

### Small marketplace batch
Expose request to top N eligible candidates and collect quotes for a short window.

### Progressive waves
Expose to a small nearby set first, then expand radius/driver count if too few usable quotes arrive.

### Manual-only driver pools
Requests may remain open longer to allow manual quotes.

The strategy can vary by density/service type but must preserve rider response-time goals.

## 11. Quote TTL

Every quote has an explicit expiry.

Expiry should account for:
- request freshness;
- driver location movement;
- service type;
- manual vs auto quote;
- market density.

Expired quotes cannot be accepted.

## 12. Rider selection

Marketplace mode:

```text
Rider receives ranked offers
-> selects one valid quote
-> agreement transaction
```

Quick Match:

```text
Rider supplies constraints
-> engine chooses highest-ranked quote satisfying constraints
-> agreement transaction
```

Quick Match is delegated choice, not platform-owned pricing.

## 13. Atomic agreement creation

Selection must be atomic.

Pseudo flow:

```text
BEGIN
  verify request is OPEN/RECEIVING_QUOTES
  verify quote belongs to request
  verify quote == PENDING and not expired
  lock/check driver availability
  verify driver still eligible
  create immutable agreement snapshot
  mark request AGREED
  mark quote ACCEPTED
  invalidate/close remaining quote paths
  reserve driver
COMMIT
```

Use database conditions/invariants as durable truth. Redis locks are only a race-reduction layer.

## 14. Double-assignment protection

Minimum protections:
- unique/conditional invariant for incompatible active work per driver;
- optimistic version/conditional request update;
- short Redis request/driver lock;
- idempotent accept endpoint;
- agreement unique key for normal single-provider request.

## 15. Cancellation interaction

If rider cancels before agreement:
- request -> CANCELLED;
- stop new quote generation;
- invalidate/expire outstanding quote visibility;
- notify relevant drivers as needed.

If agreement already exists, cancellation follows Ride/Agreement policy and may have explicit fees/compensation rules.

## 16. Driver goes stale/offline

Before agreement:
- remove/exclude from candidate discovery;
- invalidate new auto quoting;
- existing quote may be marked unavailable or allowed to expire according to policy.

After agreement:
- do not silently rematch without explicit ride recovery policy;
- surface degraded location/operator alerts.

## 17. Metrics

Required metrics:
- requests/sec;
- candidates/request;
- zero-candidate rate;
- quotes/request;
- time-to-first-quote;
- time-to-3-quotes or configured offer target;
- quote expiration rate;
- rider selection time;
- quick-match success rate;
- agreement conflict rate;
- pickup ETA distribution;
- price distribution by service/zone;
- driver exposure distribution;
- cancellation before agreement;
- request expiry rate;
- stale-driver exclusion count.

## 18. Migration from current engine

Current behavior roughly does:

```text
nearby -> score by distance/idle -> choose driver[0] -> offer -> accept -> trip assignment
```

Migration stages:
1. reuse candidate discovery;
2. introduce DriverTariff;
3. persist Quote;
4. return multiple quotes to new endpoints;
5. add ranking/reasons;
6. add Agreement transaction;
7. move rider UI to offer selection;
8. deprecate legacy one-driver offer flow.

The current engine remains a compatibility path until the marketplace flow is production-ready.
