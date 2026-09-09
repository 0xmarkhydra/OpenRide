# OpenRide — Product Requirements / Source of Truth

> **Business rebaseline: September 2026**
>
> This document replaces the previous FlashX-only product definition as the primary product requirement source for OpenRide.
>
> Legacy designated-driver and vehicle-inspection requirements remain useful as future service vertical requirements, but they no longer define the OpenRide core business model.

## 1. Product definition

OpenRide is an **open-source mobility marketplace**.

Its job is to connect rider demand with driver/provider supply while keeping commercial choice transparent:

- drivers define their own tariffs/quote rules;
- riders compare offers and choose;
- the platform performs discovery, routing, ranking, realtime tracking, trust, payment and operations;
- the platform must not silently replace driver-owned pricing with one hidden central fare.

Core phrase:

> **Drivers set their terms. Riders choose. Algorithms connect.**

## 2. Core business flow

```text
Rider creates Mobility Request
-> OpenRide finds eligible nearby drivers
-> Drivers submit / authorize Quotes
-> OpenRide ranks and explains offers
-> Rider chooses OR uses Quick Match
-> OpenRide creates Agreement
-> Driver executes Ride
-> Payment / Rating / Settlement
```

The product must preserve the distinction between:
- Request;
- Quote;
- Agreement;
- Ride.

## 3. User groups

### 3.1 Rider

Needs to:
- sign in quickly;
- choose pickup and destination;
- choose service type;
- see route/distance context;
- receive multiple transparent offers when available;
- compare price, ETA, driver, vehicle and rating;
- choose manually or Quick Match;
- track driver realtime after agreement;
- contact/support/cancel according to policy;
- pay;
- review ride and driver;
- see history and receipts.

### 3.2 Driver

Needs to:
- register and pass KYC;
- declare vehicle/capabilities;
- go online/offline;
- configure own tariff/pricing policy;
- select manual/auto/hybrid quote mode;
- receive relevant requests;
- skip unsuitable work without arbitrary punishment;
- submit or confirm quotes;
- navigate to rider;
- execute ride lifecycle;
- view earnings, fees and settlement transparently;
- view history and ratings.

### 3.3 Operator

Needs to:
- review KYC;
- monitor marketplace health and active rides;
- handle safety, fraud and disputes;
- configure legal/service-area policies;
- configure providers and payments;
- inspect ranking/quote explanations when supporting users;
- manage transparent platform/operator fees;
- audit sensitive actions.

The operator is not the default owner of every driver's price.

## 4. Rider App requirements

### 4.1 Authentication

MVP target:
- phone + OTP;
- access/refresh session;
- basic profile;
- account status handling.

### 4.2 Home / create request

Primary flow should be mobility-first:

```text
Where are you going?
-> pickup
-> destination
-> service type
-> optional preferences
-> request offers
```

At minimum show:
- pickup;
- destination;
- route distance/duration when available;
- current service type;
- rider constraints/preferences where enabled.

### 4.3 Request states

Rider-friendly states:

```text
Preparing request
-> Finding drivers
-> Offers arriving
-> Choose an offer
-> Driver confirmed
-> Driver is coming
-> Driver arrived
-> Trip in progress
-> Completed
```

Technical states may differ but UI wording must remain understandable.

### 4.4 Offer list

This is a P0 differentiator for OpenRide.

Each offer card should show as available:
- total fare;
- pickup ETA;
- pickup distance;
- driver name/avatar;
- rating and ride count/reliability context;
- vehicle summary;
- expiry/countdown when useful;
- recommendation label/reason.

Possible labels:
- **Best overall**;
- **Cheapest**;
- **Fastest pickup**;
- **Highest rated**.

Do not make `Cheapest == Best overall` by default.

### 4.5 Rider choice

Marketplace Mode:
- rider taps one valid offer;
- sees final accepted terms;
- confirms;
- Agreement is created atomically.

The rider must not see a different fare immediately after confirmation unless an explicit, separately agreed change process exists.

### 4.6 Quick Match

For riders who do not want to compare manually.

Example constraints:
- fare <= 60,000 VND;
- ETA <= 8 min;
- rating >= 4.7.

Quick Match chooses from valid driver-authorized quotes using ranking rules.

It does not generate a new platform-owned fare.

### 4.7 Realtime tracking

After Agreement:
- driver live location;
- pickup marker;
- route driver -> rider when available;
- ETA to pickup;
- driver/vehicle information;
- ride state realtime;
- route during active ride;
- reconnect after network loss;
- REST snapshot sync after reconnect.

Realtime tracking is core functionality, not an optional visual extra.

### 4.8 Cancellation

Before Agreement:
- rider can cancel open request according to simple policy;
- pending quotes stop being selectable.

After Agreement:
- cancellation reason required when policy requires;
- any fee/driver compensation must be displayed before final cancellation where possible;
- no hidden fee calculation.

### 4.9 Payment

MVP should support a provider abstraction with at least:
- cash/offline settlement option if market requires;
- digital payment provider(s) by deployment;
- payment status independent from ride status;
- receipt/history;
- retry/recovery for failed payment.

### 4.10 Rating

After completed ride:
- rating score;
- optional tags/comment;
- support/report path for safety/serious incidents.

## 5. Driver App requirements

### 5.1 Onboarding / KYC

Minimum concepts:
- identity document;
- selfie/profile image;
- driving license;
- license class/expiry;
- bank/settlement information where required;
- driver vehicle documents for passenger ride services;
- service capabilities;
- review status.

Files should continue using direct client -> S3-compatible object storage with backend-issued presigned URLs; backend stores business metadata rather than proxying all binary bytes.

### 5.2 Driver Home

Must make these obvious:
- ONLINE/OFFLINE;
- current availability/reservation state;
- current location/GPS status;
- pending request/quote activity;
- today's earnings summary;
- KYC/document warnings;
- quick access to **My Price / Tariff**.

### 5.3 My Price / Tariff

This is a primary feature, not a hidden settings page.

Driver can configure, depending on deployment/service:
- base fare;
- minimum fare;
- per-km rate;
- per-minute rate;
- pickup fee;
- maximum pickup distance;
- night/time rule;
- long-distance rule;
- zone rule;
- automatic quote minimum/maximum;
- Manual / Auto / Hybrid mode.

The UI must explain how rules combine.

Example:

```text
Base distance fare      50,000
Night adjustment         5,000
Pickup distance          3,000
------------------------------
Quote                    58,000 VND
```

### 5.4 Manual quote mode

Driver sees a relevant request with:
- pickup/destination;
- trip route distance/time;
- distance/ETA from driver to pickup;
- service type;
- rider constraints where appropriate;
- suggested market range if enabled and clearly labeled as suggestion.

Actions:
- enter/send quote;
- skip.

Skip must not automatically reduce rating.

### 5.5 Auto quote mode

System generates a quote from the driver's active tariff inside explicit driver bounds.

Driver can review how auto quoting is configured.

### 5.6 Hybrid mode

System proposes the tariff-derived quote and driver can modify/approve according to configured response time.

### 5.7 Agreement / assigned ride

After rider chooses driver:
- notify driver immediately;
- freeze/display accepted fare and terms;
- transition driver to reserved/busy;
- show navigation to pickup;
- prevent incompatible double assignment.

### 5.8 Passenger ride lifecycle

```text
Accepted/Assigned
-> En route
-> Arrived
-> Passenger onboard
-> Start ride
-> In progress
-> Complete
```

State transitions must be backend-validated and idempotent.

### 5.9 Earnings

Driver should see:
- accepted fare;
- any rider-paid fee;
- operator/platform fee;
- other explicit deductions;
- net driver earning;
- payment/settlement state.

Transparency is a core product requirement.

## 6. Operator App requirements

### 6.1 Dashboard

Key marketplace health metrics:
- open requests;
- active drivers;
- quotes/request;
- time to first quote;
- time to agreement;
- no-candidate rate;
- no-quote rate;
- request expiry/cancellation;
- active rides;
- completion rate;
- payment failures;
- safety/dispute alerts.

### 6.2 Live marketplace

Operator can inspect:
- open requests;
- candidate/quote count;
- current offer distribution;
- agreement status;
- active ride map/timeline;
- degraded/stale-driver alerts.

Sensitive pricing/ranking debug details require appropriate RBAC/audit.

### 6.3 Driver/KYC

- pending/approved/rejected/more-info filters;
- document review via signed access;
- capability/vehicle review;
- approve/reject/request-more-info;
- audit every decision.

### 6.4 Tariff / pricing operations

Operator should not have a default page saying "set every driver's per-km price".

Operator may configure:
- legal min/max guardrails;
- service-area restrictions;
- allowed currency;
- allowed tariff rule types;
- market recommendation parameters;
- platform/operator fees;
- safety/fraud price anomaly alerts.

All price-related policies should be versioned/auditable.

### 6.5 Ranking operations

Operator may tune bounded ranking weights by instance/service, but changes should be:
- versioned;
- auditable;
- observable;
- explainable.

Metrics must make it possible to detect harmful effects such as always hiding one driver cohort.

## 7. Marketplace ranking requirements

Initial ranking should be deterministic.

Candidate factors:
- pickup ETA;
- fare fit;
- driver rating/quality;
- completion reliability;
- rider preferences;
- bounded fairness/exposure.

Hard requirements:
- cheapest is not automatically best;
- safety/eligibility overrides ranking;
- reason codes returned to client;
- ranking config version available for support/debug;
- no ML requirement for MVP.

## 8. Pricing requirements

### 8.1 Driver ownership

Commercial quote must originate from:
- driver's active tariff/rules; or
- driver's manual quote.

### 8.2 Platform recommendations

OpenRide may calculate:
- demand/supply context;
- suggested range;
- predicted acceptance context.

But recommendations must be clearly labeled and must not silently rewrite accepted driver bounds.

### 8.3 Agreement integrity

After accepted quote:
- fare is frozen in Agreement;
- pricing breakdown is snapshotted;
- policy/tariff version is auditable;
- subsequent surge/demand change does not mutate the agreement.

## 9. Safety and trust

Minimum architecture/product hooks:
- KYC;
- account suspension;
- document expiry;
- incident/report workflow;
- rider/driver support;
- audit log;
- payment fraud/risk hooks;
- location freshness;
- abnormal pricing/fraud monitoring where appropriate.

Open marketplace does not mean absence of safety rules.

## 10. Service types

### Phase A — primary marketplace vertical

Start with normal passenger ride semantics because this is the clearest validation of the OpenRide marketplace model.

At minimum support architecture for:
- car passenger ride;
- motorbike passenger ride when local deployment wants it.

### Phase B — reuse legacy FlashX capabilities

Add/retain as verticals:
- designated driver car;
- designated driver motorbike;
- vehicle inspection assistance.

Their execution states differ, but they still use:

```text
Request -> Quote -> Agreement -> Execution
```

### Future

- carpool;
- intercity;
- delivery;
- truck/logistics;
- other community-defined mobility services.

## 11. Self-host / instance requirements

The system should progressively support one deployment representing an `Instance`/operator/community.

Instance configuration can include:
- service area;
- timezone/currency;
- KYC requirements;
- payment providers;
- supported services;
- operator fee policy;
- legal price guardrails;
- ranking configuration.

Full federation is not MVP.

## 12. Non-functional requirements

### Reliability
- idempotent critical writes;
- atomic agreement creation;
- no compatible double assignment;
- graceful reconnect;
- durable source of truth in PostgreSQL.

### Performance
- Redis GEO candidate lookup;
- avoid DB full scans per request;
- websocket fan-out for realtime states;
- do not persist every GPS ping transactionally.

### Observability
Required metrics/logging for:
- request lifecycle;
- quote generation;
- ranking;
- agreement conflicts;
- location freshness;
- ride state;
- payments;
- operator actions.

### Security
- private KYC objects;
- presigned upload/download;
- JWT/session controls;
- RBAC for operator actions;
- audit sensitive changes;
- rate limit/idempotency on exposed critical endpoints.

## 13. Migration requirement

Do not perform a big-bang rewrite.

Current `trips`, `pricing` and `dispatch` remain compatibility paths until V2 marketplace equivalents are ready.

Required migration path:

```text
MobilityRequest
-> DriverTariff
-> Quote
-> Marketplace Ranking
-> Agreement
-> Ride
```

See [`OPENRIDE_MIGRATION_PLAN_V2.md`](./OPENRIDE_MIGRATION_PLAN_V2.md).

## 14. Definition of OpenRide MVP success

A meaningful OpenRide marketplace MVP is achieved when:

1. driver can configure own tariff;
2. rider can create mobility request;
3. multiple eligible drivers can produce valid quotes;
4. rider can compare and select an offer;
5. Quick Match can choose from valid quotes under rider constraints;
6. accepted fare is stored in immutable agreement snapshot;
7. realtime ride tracking works after agreement;
8. driver sees transparent gross/net earning;
9. operator can perform KYC/support/safety without being the hidden owner of every fare;
10. all critical marketplace flows are covered by automated tests.
