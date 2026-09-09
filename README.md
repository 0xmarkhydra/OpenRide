# OpenRide

> **Open-source mobility marketplace where drivers set their terms, riders choose, and algorithms connect.**

OpenRide is not a Grab/Uber clone and it is not a platform whose core job is to impose one fare on every driver.

OpenRide is an open-source foundation for communities, cooperatives, local operators and startups to run fair mobility marketplaces. Drivers can define their own pricing rules. Riders can compare transparent offers. The platform provides routing, discovery, matching, realtime location, trust, payments and operational tooling without silently taking ownership of the commercial terms between the two sides.

## Core principles

1. **Drivers control their commercial terms.**
2. **Riders keep the final choice.**
3. **The marketplace must not rank only by lowest price.**
4. **Declining an unsuitable request is normal driver behavior, not an automatic punishment signal.**
5. **An accepted price is snapshotted into an agreement and cannot be silently changed.**
6. **Pricing and ranking decisions must be explainable.**
7. **OpenRide Core must remain self-hostable and operator-independent.**

See [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md).

## Product model

The core business flow is:

```text
Rider demand
    |
    v
Mobility Request
    |
    v
Candidate Discovery
    |
    v
Driver Tariffs / Quotes
    |
    v
Marketplace Ranking
    |
    v
Rider Choice
    |
    v
Agreement
    |
    v
Ride / Job
    |
    +--> Payment
    +--> Rating / Trust
```

The important separation is:

- **Request** — what the rider needs.
- **Quote** — what a driver is willing to do it for.
- **Agreement** — the terms both sides accepted.
- **Ride / Job** — execution of the agreed service.

A ride is therefore an execution object, not the entire marketplace.

## Example

For the same 10 km request:

```text
Driver A: 5,000 VND/km -> 50,000 VND
Driver B: 6,000 VND/km -> 60,000 VND
Driver C: 5,500 VND/km -> 55,000 VND
```

OpenRide should not blindly pick the cheapest driver.

The rider can instead see useful trade-offs such as:

- best overall match;
- cheapest offer;
- fastest pickup;
- highest-rated driver.

Ranking can consider pickup ETA, price fit, reliability, quality, rider preferences and marketplace fairness. The factors and resulting recommendation should be explainable.

## Driver pricing

A driver tariff can contain more than `price_per_km`:

```text
Base / minimum fare
Per-km rate
Per-minute rate
Pickup radius
Pickup fee
Night / holiday adjustments
Long-distance rule
Minimum acceptable quote
Maximum automatic quote
Manual / Auto / Hybrid quote mode
```

Auto pricing is allowed only within rules and bounds the driver has accepted. The system may recommend a market adjustment, but should not silently overwrite a driver's terms.

## Supported marketplace modes

### Marketplace mode

Multiple eligible drivers can return offers. The rider compares and chooses.

### Quick match

The rider can ask OpenRide to choose automatically using constraints such as maximum fare, ETA and rating. The system still chooses from valid driver-authorized quotes; it does not invent a platform-owned fare.

## Long-term service model

OpenRide should be built around generic marketplace primitives so additional mobility services can plug in later:

- passenger car;
- motorbike;
- carpool;
- intercity;
- delivery;
- designated driver;
- vehicle assistance;
- vehicle inspection assistance.

Legacy FlashX designated-driver and inspection workflows in this repository are treated as future service verticals, not as the OpenRide business kernel.

## Current technology foundation

The existing technical foundation is intentionally retained while the business kernel is migrated:

- Rider app: Flutter
- Driver app: Flutter
- Operator/Admin: Next.js + TypeScript
- Backend: Go modular monolith
- Primary database: PostgreSQL + PostGIS
- Realtime/cache/geo: Redis
- Realtime transport: WebSocket
- Routing/maps: provider abstraction
- Push: FCM/APNs-ready architecture
- Object storage: S3-compatible

We do **not** plan to rewrite the system into microservices just to look enterprise. Modules should be extracted only when production load, ownership or deployment needs justify it.

## Repository layout

```text
OpenRide/
├── apps/
│   ├── rider/
│   ├── driver/
│   ├── admin/          # target name: operator
│   └── landing/
├── services/
│   └── api/
├── infrastructure/
│   └── migrations/
├── docs/
└── .github/
```

The target architecture will gradually introduce marketplace-oriented modules such as `requests`, `tariffs`, `quotes`, `agreements`, `marketplace`, `fairness` and `rides` while preserving compatibility with the current runtime during migration.

## Architecture documents

Start here:

- [`docs/PRODUCT_VISION.md`](docs/PRODUCT_VISION.md) — product scope and user value.
- [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md) — non-negotiable community principles.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — target technical architecture.
- [`docs/DOMAIN_MODEL.md`](docs/DOMAIN_MODEL.md) — target business domains.
- [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) — target persistence model.
- [`docs/DISPATCH_ENGINE.md`](docs/DISPATCH_ENGINE.md) — migration from dispatch to marketplace matching.
- [`docs/CUSTOMER_REQUIREMENTS.md`](docs/CUSTOMER_REQUIREMENTS.md) — current product requirements.
- [`docs/OPENRIDE_MIGRATION_PLAN_V2.md`](docs/OPENRIDE_MIGRATION_PLAN_V2.md) — phased migration from the current codebase.

## Migration policy

The current codebase contains legacy ride-hailing and FlashX designated-driver concepts. We will not delete all of that at once.

Migration is additive and staged:

```text
Current Trip/Pricing/Dispatch
        |
        v
MobilityRequest
        |
        v
DriverTariff
        |
        v
Quote
        |
        v
Marketplace Matching
        |
        v
Agreement
        |
        v
Ride
```

Existing flows remain operational until their marketplace replacement is tested and ready.

## Development

The existing development commands remain valid during the rebaseline:

```bash
make stack-up
make stack-ps
make stack-logs
make stack-down
make infra-up
make api-run
make api-test
make docker-test
make docker-integration-test
```

API health in the current runtime:

```text
GET http://localhost:8080/health
```

## Project status

OpenRide is currently in a **business/domain rebaseline**. The repository already contains working foundations for auth, drivers, location, dispatch, trips, payments, ratings, realtime, PostgreSQL/PostGIS, Redis and CI. The next stage is to migrate the business kernel from platform-controlled trip pricing/dispatch toward request -> quote -> agreement -> ride marketplace semantics.

## Community

The goal is bigger than launching one ride-hailing company.

OpenRide should make it possible for a local driver community, cooperative, startup or regional operator to self-host a fair mobility marketplace without rebuilding the entire technical stack from zero.

> **Build infrastructure for mobility communities, not another closed platform that communities depend on.**
