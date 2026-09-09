# OpenRide

> **Open-source mobility marketplace where drivers set their terms, riders choose, and algorithms connect.**

**OpenRide is not another Grab/Uber clone.** It is an open-source foundation for communities, cooperatives, local operators and startups to run fair mobility marketplaces without rebuilding the entire technical stack from zero.

Drivers can define their own pricing rules. Riders can compare transparent offers. OpenRide provides routing, discovery, matching, realtime location, trust, payments and operational tooling without silently taking ownership of the commercial terms between the two sides.

[![License: AGPL-3.0-or-later](https://img.shields.io/badge/License-AGPL--3.0--or--later-blue.svg)](LICENSE)
![Stage](https://img.shields.io/badge/stage-marketplace%20V2%20foundation-orange)

## Why OpenRide exists

Traditional ride-hailing platforms solve difficult technical problems, but the platform can also become the party that controls pricing, matching visibility and access to customers.

OpenRide explores a different model:

```text
Drivers set their terms.
Riders choose.
Algorithms connect.
Communities can self-host.
```

The goal is not to remove algorithms. The goal is to make them **serve a transparent marketplace instead of silently becoming the marketplace owner**.

## Core principles

1. **Drivers control their commercial terms.**
2. **Riders keep the final choice.**
3. **The marketplace must not rank only by lowest price.**
4. **Declining an unsuitable request is normal driver behavior, not an automatic punishment signal.**
5. **An accepted price is snapshotted into an agreement and cannot be silently changed.**
6. **Pricing and ranking decisions must be explainable.**
7. **OpenRide Core must remain self-hostable and operator-independent.**

See [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md).

## Marketplace model

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
Rider Choice / Quick Match
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

A ride is an execution object, not the entire marketplace.

## Simple pricing example

For the same 10 km request:

```text
Driver A: 5,000 VND/km -> 50,000 VND
Driver B: 6,000 VND/km -> 60,000 VND
Driver C: 5,500 VND/km -> 55,000 VND
```

OpenRide should not blindly pick Driver A because A is cheapest.

The rider can see useful trade-offs such as:

- **best overall match**;
- **cheapest offer**;
- **fastest pickup**;
- **highest-rated driver**.

Ranking can consider pickup ETA, price fit, reliability, quality, rider preferences and marketplace fairness. Recommendations should be explainable.

## Driver pricing

A driver tariff is more than one `price_per_km` field:

```text
Base / minimum fare
Per-km rate
Per-minute rate
Pickup radius
Pickup fee
Night / holiday adjustments
Long-distance rules
Minimum acceptable quote
Maximum automatic quote
Manual / Auto / Hybrid quote mode
```

Auto pricing is allowed only inside rules and bounds the driver has accepted. The system may recommend a market adjustment, but should not silently overwrite a driver's terms.

### Manual

The driver reviews a request and sends a price.

### Auto

OpenRide can generate a quote automatically, but only inside the driver's configured bounds.

### Hybrid

OpenRide prepares a suggested quote and the driver can adjust it before sending.

## Rider modes

### Marketplace mode

Multiple eligible drivers can return offers. The rider compares and chooses.

### Quick Match

The rider can ask OpenRide to choose automatically using constraints such as maximum fare, ETA and rating. Quick Match still chooses from valid driver-authorized quotes; it does not invent a platform-owned fare.

## Public-launch foundation

The repository now contains the first OpenRide V2 marketplace foundation:

- `DriverTariff`, `MobilityRequest`, `Quote`, `Agreement` and `Ride` domain primitives;
- additive V2 marketplace database migration;
- marketplace state-transition and quote-bound tests;
- rewritten architecture, data model, API and product documentation;
- Rider, Driver, Operator and Landing applications from the existing runtime foundation;
- PostgreSQL/PostGIS, Redis, realtime, auth, payments, ratings and object-storage foundations;
- community contribution, security and licensing policies.

This is a **pre-1.0 foundation**, not a claim that the project is ready to carry real passengers in production. Safety, legal, payment, KYC, incident response and marketplace runtime integration must be completed and validated by each operator before real-world deployment.

## Long-term service model

OpenRide is built around generic marketplace primitives so mobility verticals can plug in later:

- passenger car;
- motorbike;
- carpool;
- intercity;
- delivery;
- designated driver;
- vehicle assistance;
- vehicle inspection assistance.

Legacy FlashX designated-driver and inspection workflows in this repository are treated as service verticals and compatibility code, not as the OpenRide business kernel.

## Technology foundation

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

We do **not** plan to rewrite the system into microservices just to look enterprise. Modules should be extracted only when production load, ownership or deployment requirements justify it.

## Repository layout

```text
OpenRide/
├── apps/
│   ├── rider/
│   ├── driver/
│   ├── admin/          # operator console; directory rename is staged
│   └── landing/
├── services/
│   └── api/
│       └── internal/
│           └── marketplace/
├── infrastructure/
│   └── migrations/
├── docs/
├── CONTRIBUTING.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
└── LICENSE
```

## Read the design

- [`docs/PRODUCT_VISION.md`](docs/PRODUCT_VISION.md) — product vision and scope.
- [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md) — non-negotiable community principles.
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — target technical architecture.
- [`docs/DOMAIN_MODEL.md`](docs/DOMAIN_MODEL.md) — business domains and aggregates.
- [`docs/DATA_MODEL.md`](docs/DATA_MODEL.md) — persistence model.
- [`docs/DISPATCH_ENGINE.md`](docs/DISPATCH_ENGINE.md) — migration from dispatch to marketplace matching.
- [`docs/API_CONTRACT_V2.md`](docs/API_CONTRACT_V2.md) — V2 API direction.
- [`docs/CUSTOMER_REQUIREMENTS.md`](docs/CUSTOMER_REQUIREMENTS.md) — product requirements.
- [`docs/OPENRIDE_MIGRATION_PLAN_V2.md`](docs/OPENRIDE_MIGRATION_PLAN_V2.md) — staged migration plan.
- [`docs/ADR_OPENRIDE_V2.md`](docs/ADR_OPENRIDE_V2.md) — architecture decisions.
- [`docs/LEGACY_COMPATIBILITY.md`](docs/LEGACY_COMPATIBILITY.md) — why some `flashx` internal identifiers temporarily remain.

## Migration policy

We will not destroy the current runtime and replace everything in one rewrite.

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

Existing flows remain available until their marketplace replacements are tested and ready.

Some internal identifiers still use `flashx` for deployment compatibility. They are documented in [`docs/LEGACY_COMPATIBILITY.md`](docs/LEGACY_COMPATIBILITY.md) and should not be interpreted as the target product identity.

## Development

Existing development commands remain valid during the migration:

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

## Contributing

Start with [`CONTRIBUTING.md`](CONTRIBUTING.md). Product and architecture changes should preserve the manifesto and explain how they affect driver autonomy, rider choice, transparency, safety and self-hosting.

For security vulnerabilities, read [`SECURITY.md`](SECURITY.md) before opening a public issue. Community participation is governed by [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## License

OpenRide is licensed under **GNU AGPL-3.0-or-later**. See [`LICENSE`](LICENSE).

The copyleft choice is intentional: OpenRide is meant to remain a commons that communities can inspect, modify and self-host, including when modified versions are offered over a network.

## Community

The ambition is bigger than launching one ride-hailing company.

OpenRide should make it possible for a driver community, cooperative, startup or regional operator to run a fair mobility marketplace without becoming permanently dependent on a single closed platform.

> **Build infrastructure for mobility communities, not another closed platform that communities depend on.**
