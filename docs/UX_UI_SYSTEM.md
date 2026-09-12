# OpenRide UX/UI System

> **Rebaseline: 12/09/2026**
> This is the production UX contract for OpenRide. The approved dark concept mock is the visual north star, but domain truth always wins over decoration.

## 1. Product UX promise

OpenRide must make three things obvious on every marketplace screen:

1. **Drivers set their own commercial terms.**
2. **Riders compare and choose.**
3. **OpenRide connects, ranks and explains — it does not silently rewrite quotes.**

Once a rider accepts a quote, the product leaves **marketplace mode** and enters **agreement mode**. The accepted agreement must render the commercial snapshot that both sides accepted.

## 2. Visual direction

The approved mock defines the visual language:

- dark navy / near-black canvas;
- cyan / electric blue for navigation, discovery and platform actions;
- green for selected, accepted and completed states;
- white primary text, cool-gray secondary text;
- dark elevated cards with subtle borders;
- restrained glow only for active/selected states;
- dark map style with high-contrast route and driver markers;
- premium, calm, modern mobility feel;
- mobile-first, one-hand friendly.

### Avoid

- neon everywhere;
- fake urgency/countdowns unless the domain really has TTL;
- hiding fare/ETA/rating behind extra taps;
- UI that implies OpenRide owns or sets the driver's quote;
- oversized recommendation treatment that visually suppresses alternatives;
- marketing copy inside task-critical states.

## 3. Design tokens

### Colors

| Token | Role | Starting value |
| --- | --- | --- |
| `bg.canvas` | app background | `#030712` |
| `bg.surface` | cards/sheets | `#0B1220` |
| `bg.surfaceRaised` | elevated surfaces | `#111827` |
| `border.subtle` | default border | `rgba(148,163,184,0.18)` |
| `text.primary` | main text | `#F8FAFC` |
| `text.secondary` | secondary text | `#94A3B8` |
| `accent.primary` | cyan | `#22D3EE` |
| `accent.secondary` | blue | `#38BDF8` |
| `state.success` | green | `#34D399` |
| `state.warning` | warning | `#FBBF24` |
| `state.danger` | destructive | `#FB7185` |

Use semantic tokens in Flutter. Do not scatter literal colors through feature widgets.

### Typography

- Inter/system sans.
- Body: 15–16sp.
- Secondary: 13–14sp.
- Screen title: 24–32sp.
- Driver-set price: 24–32sp with tabular numerals where possible.
- Uppercase only for short status labels/badges.

### Spacing and geometry

- 4px base grid.
- Main spacing: 8 / 12 / 16 / 24 / 32.
- Screen horizontal padding: 16–20.
- Card radius: 16–20.
- Primary CTA height: 52–56.
- Minimum tap target: 44x44.
- Bottom-sheet radius: 24–28.

## 4. Core navigation

### Rider

1. Home
2. Trips
3. Messages
4. Account

Booking/request steps live in a focused flow, not extra permanent tabs.

### Driver

1. Work
2. Requests / Quotes
3. Trips
4. Earnings
5. Account

Online/offline state must always be visible on Work.

## 5. Rider UX flow

### R0 — Entry / permissions

- Ask location only when needed.
- Manual pickup is always available if permission is denied.
- Explain permission value in plain Vietnamese.
- Notification permission is not a hard blocker.

### R1 — Home / map

Primary elements:

- pickup;
- destination;
- nearby supply/map context;
- saved/recent places;
- primary CTA: **Request offers**.

Do not show a platform-owned final fare as if it were the driver's price.

### R2 — Request review

Show:

- route;
- service module/category;
- now/later if supported;
- rider notes/constraints;
- clear `Request offers` CTA.

This creates a **Mobility Request**, not a deal.

### R3 — Waiting for offers

- `Drivers are reviewing your request` state;
- real offer count;
- request expiry/cancel/edit rules if supported;
- honest empty/skeleton state;
- never render fake sample quotes as live data.

### R4 — Offers marketplace — signature screen

Each OfferCard must expose:

- driver identity/verification where available;
- driver-set price/terms;
- ETA to pickup;
- distance to pickup;
- rating/reputation;
- vehicle;
- quote expiry/conditions when relevant.

Recommended hierarchy:

1. Price.
2. ETA + distance.
3. Reputation.
4. Vehicle.
5. Conditions.

Sorting:

- Recommended
- Lowest price
- Fastest pickup
- Best rated

If `Recommended` is shown, explain the main reason, e.g. `Balanced for price, pickup time and reputation`.

**Invariant:** ranking may reorder offers; it must not mutate, hide or fabricate quote terms.

### R5 — Offer detail

Show complete terms, cancellation rule, driver/vehicle detail and ranking explanation.

Comparison of 2–3 offers is P1; do not delay MVP for a complex table.

### R6 — Select and confirm

Use a two-step mental model:

1. Select offer.
2. Confirm acceptance.

The confirmation sheet repeats the exact commercial terms. If the quote changed, force a fresh review.

### R7 — Agreement locked

After acceptance, visually leave marketplace mode.

Use green success treatment:

- `Trip agreed`
- `Price and terms confirmed`

Agreement screen renders:

- accepted price/terms;
- driver;
- route/request snapshot;
- acceptance timestamp;
- agreement identifier where useful.

**Invariant:** render from the Agreement snapshot, not a newly recalculated quote.

### R8 — Driver en route / active ride

Prioritize:

- map/route;
- ETA/status;
- contact/message;
- safety actions;
- agreed terms still accessible.

Do not keep promoting alternative marketplace offers after agreement.

### R9 — Completion

- completed state;
- receipt / amount explanation;
- rating/feedback;
- report/support entry.

### R10 — History

Trip detail must distinguish:

- request;
- quote;
- accepted agreement;
- execution;
- settlement/receipt.

## 6. Driver UX flow

### D0 — Readiness

Show:

- profile/verification status;
- capability/vehicle requirements where relevant;
- online/offline;
- reasons blocking quote participation.

### D1 — Work / request discovery

Relevant request cards may include:

- pickup area/distance;
- destination/direction if policy permits;
- approximate distance/duration when available;
- service requirements;
- request age/expiry.

### D2 — Request detail

Driver reviews request before pricing.

Primary CTA: **Send my offer**.

### D3 — Quote composer — signature driver screen

Make ownership explicit: **Your price**.

Depending on module contract, support:

- fixed total;
- per-km/formula pricing;
- notes/conditions;
- quote expiry.

Before submit, show a human-readable summary.

**Invariant:** validation may reject/normalize structure; OpenRide must not silently replace the driver's commercial term.

### D4 — Active quotes

States:

- Sent
- Accepted
- Expired
- Withdrawn

Only show `Viewed` if the backend truly supports it.

### D5 — Accepted agreement

Driver sees the same accepted commercial semantics as rider.

Primary actions switch from marketplace actions to execution actions.

### D6 — Ride execution

Status machine must be explicit and hard to trigger accidentally:

- heading to pickup;
- arrived;
- in progress;
- completed;
- cancel only where policy allows.

Critical transitions require confirmation and server-confirmed state.

### D7 — Earnings / history

If a deployment charges fees, separate:

- gross agreed amount;
- network/platform fee;
- net settlement.

Never collapse them into one opaque number.

## 7. Shared components

### OfferCard

Variants:

- default;
- recommended;
- selected;
- stale;
- expired;
- withdrawn.

Required anatomy:

- driver block;
- price;
- ETA + distance;
- rating;
- vehicle;
- conditions;
- status/recommendation reason;
- selection affordance.

### AgreementCard

- lock icon;
- agreement status;
- accepted terms;
- participants;
- timestamp;
- link to full detail.

### MoneyText

- consumes integer/minor-unit domain values;
- locale-formats for display;
- never feeds business logic from formatted strings.

### StatusChip

Semantic colors only. Status must also be readable in text.

### EmptyState / ErrorState / OfflineBanner

Every remote-data screen must cover honest no-data, error and reconnect states.

## 8. Required screen states

For all main screens design:

- loading;
- loaded;
- empty;
- error;
- offline;
- stale data;
- permission denied where relevant.

Quote-specific race states:

- quote expired while rider is viewing it;
- driver updated quote before acceptance;
- quote withdrawn;
- request already cancelled/accepted;
- network dropped immediately after accept tap.

Never resolve a race by silently substituting another quote.

## 9. Motion

- standard transitions: 150–220ms;
- sheets/dialogs: 250–350ms;
- selected offer: subtle border/glow;
- new offer: soft insertion, no flashing;
- agreement accepted: one clear success transition;
- respect reduced-motion preferences.

## 10. Accessibility

Minimum bar:

- WCAG-minded contrast;
- status not encoded by color alone;
- 44x44 minimum targets;
- semantic labels for icon-only controls;
- dynamic text does not break price/CTA;
- critical map information has text/list alternatives.

## 11. Content rules

Prefer plain contractual copy:

- `Tài xế đưa giá`
- `Bạn chọn tài xế`
- `Giá đã thỏa thuận`
- `Thỏa thuận đã được xác nhận`

Avoid misleading copy:

- `Giá của OpenRide`
- `OpenRide quyết định giá tốt nhất`
- `Chúng tôi đã tối ưu giá tài xế`

Recommendation copy explains *why*; it never implies OpenRide changed the quote.

## 12. Mock → production rule

The concept mock is a **visual north star**, not business truth.

A screen is only Done when:

1. domain state is identified;
2. owner package/service is known;
3. loading/empty/error/race states exist;
4. money maps to exact integer/minor-unit values;
5. quote ownership is visually truthful;
6. accepted terms render from Agreement snapshot data;
7. accessibility is covered;
8. real API/event contract or explicit fixture exists;
9. important commands have safe/idempotent UX behavior;
10. the screen does not claim backend capability that is still only planned.

## 13. Implementation order

1. Theme/tokens + shared primitives.
2. Rider Home + Request flow.
3. Driver Request + Quote Composer.
4. Rider Offers marketplace + OfferCard.
5. Sorting + recommendation explanation.
6. Select/confirm + AgreementCard.
7. Accepted agreement screens for both roles.
8. Ride execution states.
9. History/receipt/reputation.
10. Messaging/safety refinements.
11. Advanced comparison and operator theming.

This order proves OpenRide's marketplace contract before polishing peripheral surfaces.
