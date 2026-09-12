# OpenRide UX / BA Roadmap

> **Baseline: 12/09/2026**  
> Purpose: turn the approved OpenRide mock direction into implementable product slices without letting visuals outrun domain truth.

## 1. Product statement

OpenRide is an open-source mobility marketplace where:

- riders create mobility demand;
- drivers/providers own their commercial terms;
- OpenRide discovers and ranks eligible offers;
- riders choose an offer or an allowed quick-match strategy chooses within rider constraints;
- accepted terms become an Agreement snapshot;
- execution happens after Agreement.

Core chain:

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

UI must preserve this chain. Do not collapse it back into one mutable `Trip` concept.

## 2. UX north star

The approved mock is the visual north star:

- premium dark navy canvas;
- cyan/electric-blue discovery/navigation accents;
- green accepted/completed state;
- strong card hierarchy;
- map as context, not decoration;
- price/ETA/reputation visible without extra taps;
- no platform-owned fare illusion.

The product should feel more like a transparent marketplace than a dispatch monopoly.

## 3. Non-negotiable marketplace invariants

### INV-01 — Driver owns quote terms

OpenRide may validate structure and eligibility. It must not silently replace a driver's commercial term.

### INV-02 — Ranking cannot mutate canonical quote data

Ranking may reorder eligible offers and explain the result. It cannot fabricate a quote, hide a quote outside explicit policy, or rewrite price/conditions.

### INV-03 — Rider acceptance is explicit

Selecting a card is not enough. Acceptance requires a confirmation step with the commercial terms visible.

### INV-04 — Agreement is immutable commercial evidence

After acceptance, the product renders from Agreement snapshot data. Do not recalculate and overwrite accepted terms.

### INV-05 — Money stays exact

Business values use integer/minor units. UI formatting never becomes the source of truth.

### INV-06 — Critical commands are retry-safe

Accept/confirm actions must be designed around idempotent server semantics. Repeated taps or network retries must not create multiple agreements.

### INV-07 — UI never claims planned backend capability as implemented

Fixtures must be clearly development/demo fixtures. Public product UI must not fabricate realtime, rating, KYC, payment or availability data.

## 4. Primary personas

### Rider

Needs to:

- define pickup/destination/service constraints;
- receive real alternatives;
- compare price, pickup ETA, reputation and vehicle;
- understand why an offer is recommended;
- choose safely;
- retain proof of the agreed terms;
- track execution and resolve problems.

### Driver

Needs to:

- become eligible to participate;
- see relevant demand;
- understand the request before pricing;
- send own commercial offer;
- manage pending/accepted/expired offers;
- execute accepted agreements;
- understand gross, fees and net settlement.

### Operator / community

Needs to:

- configure deployment policy without taking ownership of driver quotes;
- moderate eligibility/trust;
- support disputes;
- inspect marketplace health;
- explain ranking/policy decisions where appropriate.

## 5. Rider slices

### UX-R01 — Rider home / request entry

**Goal:** user can create demand with minimal friction.

**Screen content**

- pickup;
- destination;
- service/module selector only when multiple modules are available;
- recent/saved places;
- location permission state;
- CTA `Request offers`.

**Do not show**

- fake fare;
- fake nearby driver count;
- unverified ETA.

**Acceptance criteria**

- permission denial has manual address fallback;
- draft survives non-destructive navigation where feasible;
- route/service validation errors are readable;
- submit creates a MobilityRequest, not an Agreement.

### UX-R02 — Request review

**Goal:** ensure rider understands what is being sent to the market.

Show route, schedule when supported, service requirements, notes and edit affordances.

### UX-R03 — Waiting for offers

**Goal:** make waiting honest and understandable.

States:

- no offers yet;
- one or more offers arriving;
- request expired;
- request cancelled;
- offline/reconnecting.

Never show fabricated quote cards while waiting.

### UX-R04 — Offers marketplace

**Goal:** this is OpenRide's signature rider experience.

Offer card anatomy:

- driver identity / verification status when available;
- driver-set price;
- ETA to pickup;
- distance to pickup;
- reputation/rating when implemented;
- vehicle;
- expiry/conditions;
- recommendation reason when ranked as recommended.

Sort/filter baseline:

- Recommended;
- Lowest price;
- Fastest pickup;
- Best rated when rating data is trustworthy.

**BA rule:** the platform can rank but cannot silently mutate/hide/fabricate canonical quote terms.

### UX-R05 — Offer detail

**Goal:** show complete offer semantics before commitment.

Include:

- full commercial terms;
- vehicle/driver detail;
- quote expiry;
- cancellation policy if available;
- explanation of recommendation/rank factors at a human-readable level.

### UX-R06 — Select / confirm

**Goal:** prevent accidental acceptance and quote-race ambiguity.

Flow:

```text
Select offer
-> Confirmation sheet
-> server validates still-current quote
-> Agreement created
```

Race handling:

- quote expired -> block and refresh;
- quote withdrawn -> explain and return to marketplace;
- quote changed -> show new terms and require fresh confirmation;
- request already accepted -> open resulting Agreement;
- network timeout after tap -> query/replay idempotent result before inviting another acceptance.

### UX-R07 — Agreement locked

**Goal:** visibly transition from marketplace to contract/execution state.

Show:

- accepted price/terms;
- driver;
- route/request snapshot;
- acceptance timestamp;
- Agreement ID when useful;
- green success/locked treatment.

### UX-R08 — Active ride / execution

Design for:

- driver heading to pickup;
- arrived;
- in progress;
- completed;
- cancellation only where policy permits;
- contact/support/safety actions;
- offline/reconnect.

Agreement terms remain reachable throughout execution.

### UX-R09 — Completion / receipt

Show:

- completed state;
- agreed amount;
- fee/tax breakdown if applicable;
- payment/settlement state when implemented;
- rating/reputation flow;
- report/support.

### UX-R10 — History

History detail must clearly separate:

- request;
- quote;
- agreement;
- execution;
- receipt/settlement.

## 6. Driver slices

### UX-D01 — Readiness / eligibility

Show profile, verification, vehicle/capability requirements and reasons the driver cannot participate.

Do not use a vague disabled Online switch with no explanation.

### UX-D02 — Request discovery

Relevant request card may show:

- pickup area/distance;
- direction/destination according to policy;
- estimated distance/duration when available;
- service requirements;
- age/expiry.

### UX-D03 — Request detail

Driver reviews enough demand information to price responsibly.

Primary CTA: `Send my offer`.

### UX-D04 — Quote composer

**Signature driver experience.**

Headline must make ownership explicit: `Your price` / `Giá của bạn`.

Depending on module support:

- fixed total;
- tariff/per-km inputs;
- conditions/notes;
- expiry.

Before submit, render a plain-language quote summary.

### UX-D05 — Active quotes

States:

- Sent;
- Accepted;
- Expired;
- Withdrawn.

Only display `Viewed` if backend supports a real viewed event.

### UX-D06 — Accepted agreement

Driver sees the same accepted commercial semantics as rider. Marketplace actions disappear; execution actions replace them.

### UX-D07 — Execution

Critical state transitions should be explicit and server-confirmed:

```text
Heading to pickup
-> Arrived
-> In progress
-> Completed
```

Avoid optimistic success on actions with financial or safety consequence.

### UX-D08 — Earnings / history

If deployment charges fees, separate:

- gross agreed amount;
- operator/network/platform fee;
- net settlement.

## 7. Shared component backlog

### P0 primitives

- `OpenRideTheme`
- `AppScaffold`
- `PrimaryButton`
- `SecondaryButton`
- `DangerButton`
- `MoneyText`
- `StatusChip`
- `OfferCard`
- `AgreementCard`
- `DriverIdentityRow`
- `VehicleRow`
- `RouteSummary`
- `RecommendationReason`
- `LoadingSkeleton`
- `EmptyState`
- `ErrorState`
- `OfflineBanner`
- `ConfirmationSheet`

### OfferCard variants

- default;
- recommended;
- selected;
- stale;
- expired;
- withdrawn.

### AgreementCard rules

AgreementCard always uses accepted Agreement data and visually distinguishes locked terms from live market offers.

## 8. Ranking UX

Ranking must be explainable enough to avoid becoming a hidden price-control mechanism.

A recommended card may explain factors such as:

- balanced price and pickup time;
- strong reliability/reputation;
- closer pickup;
- rider constraints matched.

Do not say:

- `OpenRide chose the best price` if the system only ranked offers;
- `OpenRide optimized this driver's price`;
- `cheapest = recommended` unless that is genuinely the selected ranking mode.

## 9. Error and race matrix

| Case | Expected UX |
| --- | --- |
| Request has no offers | honest waiting/empty state + edit/cancel rules |
| Quote expires while open | disable acceptance + refresh marketplace |
| Driver withdraws quote | explain withdrawal, never substitute another quote |
| Driver changes quote | require rider to review the new terms |
| Rider double-taps accept | one Agreement result via idempotent/replay-safe server behavior |
| Network drops after accept | resolve server state before allowing another acceptance |
| Request already accepted elsewhere | open the resulting Agreement |
| Stale offer list | show stale/reconnecting state and refresh |
| Location denied | manual pickup input |
| Rating/KYC unavailable | omit or label unavailable; never fabricate |

## 10. BA implementation gate per screen

Before a screen moves from design to implementation, record:

1. user goal;
2. domain object/state;
3. owner package/service;
4. API/event contract;
5. money representation;
6. idempotency/race behavior;
7. loading/empty/error/offline states;
8. security/privacy implications;
9. analytics event if needed;
10. Definition of Done.

A mock alone is not implementation evidence.

## 11. Delivery phases

### Phase A — Design system foundation

Deliver:

- color/typography/spacing tokens;
- shared buttons/cards/status/empty/error components;
- dark map styling contract;
- accessibility baseline.

### Phase B — Prove marketplace loop

Deliver first end-to-end happy path:

```text
Rider request
-> Driver sees request
-> Driver sends quote
-> Rider sees multiple offers
-> Rider selects + confirms
-> Agreement shown to both sides
```

This phase is more important than polishing secondary tabs.

### Phase C — Execution

Add post-agreement ride states, contact, safety/support, reconnect behavior and history.

### Phase D — Trust / settlement

Add ratings, verification, payment/settlement presentation only as backend capabilities are proven.

### Phase E — Advanced marketplace UX

Add:

- compare 2–3 offers;
- richer ranking explanation;
- quick match with rider constraints;
- operator/community theming;
- multi-service module refinements.

## 12. Design deliverables sequence

Recommended screen-design order:

1. Rider Home / Request.
2. Driver Request Detail.
3. Driver Quote Composer.
4. Rider Offers Marketplace.
5. Offer Detail.
6. Accept Confirmation.
7. Rider Agreement Locked.
8. Driver Agreement Accepted.
9. Active Ride states.
10. Completion / Receipt.
11. Rider History.
12. Driver Earnings / History.
13. Account/KYC/Settings.
14. Operator surfaces.

## 13. Definition of Done for UX slices

A slice is Done only when:

- visual hierarchy follows `UX_UI_SYSTEM.md`;
- domain semantics match current OpenRide source of truth;
- no false backend capability is implied;
- loading/empty/error/race states are designed;
- quote ownership remains truthful;
- Agreement snapshot semantics are preserved;
- exact money representation is mapped correctly;
- critical actions have safe retry behavior;
- accessibility minimums are covered;
- implementation has a runnable/testable path appropriate to its status.

## 14. Current priority

The next design/implementation work should focus on the marketplace signature loop, not peripheral polish:

**Driver sets price -> Rider compares real offers -> Rider confirms -> Agreement locks accepted terms.**

If this loop is clear and trustworthy, OpenRide feels meaningfully different from a conventional centrally priced ride-hailing clone.
