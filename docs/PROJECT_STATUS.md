# OpenRide Project Status

> Last reviewed: 2026-09-09

OpenRide is a **pre-1.0 open-source mobility marketplace** under active architectural migration. This document is intentionally conservative: it separates code that exists and is testable today from target architecture that is not implemented yet.

## Status legend

- ✅ **Implemented** — code exists in the repository and has an executable/testable path.
- 🟡 **Foundation** — contracts, schema, package, or service boundary exists but the end-to-end production path is incomplete.
- 🔁 **Compatibility** — legacy runtime still serves this capability while traffic is migrated.
- ⏳ **Planned** — architectural target only; do not assume it exists.

## Marketplace platform

| Capability | Status | Notes |
|---|---|---|
| OpenRide Core domain package | ✅ | Independent `packages/core-go` module with marketplace value objects, pricing/ranking ports and engine tests. |
| Service module registry | ✅ | `passenger.car` and `carpool.intercity` first-party modules. |
| Versioned JSON contracts | ✅ | Service manifest, marketplace event and request schemas. |
| JavaScript / TypeScript SDK | ✅ | Zero-runtime-dependency Fetch-based client package. |
| Marketplace Service process | ✅ | Independent Go service with health/readiness, catalog and request validation APIs. |
| Marketplace-owned PostgreSQL schema | 🟡 | Initial migration exists; full repositories/use cases are not yet wired. |
| Driver tariffs persisted through Marketplace Service | 🟡 | Domain/schema exists; complete public command/query API is not yet implemented. |
| Mobility request persistence | 🟡 | Domain/schema exists; end-to-end V2 request flow is not yet complete. |
| Quote persistence/acceptance | 🟡 | Domain/schema exists; full durable quote workflow is not yet complete. |
| Agreement persistence | 🟡 | Domain/schema exists; atomic cross-service production workflow is not yet complete. |
| Ranking engine | ✅ | Portable deterministic/explainable Core policy exists. Production data adapters are still being extracted. |
| Outbox / Inbox tables | 🟡 | Schema exists. Production relay/consumer implementation is not complete yet. |
| NATS JetStream topology | 🟡 | Development topology exists. Domain-event publisher/consumer coverage is incomplete. |

## Services

| Service | Status | Data ownership |
|---|---|---|
| Compatibility API / edge | 🔁 | Owns legacy V1 runtime data during migration. |
| Marketplace Service | 🟡 | Dedicated `marketplace` database/migrations in the microservices development topology. |
| Location Service | ⏳ | Will own online presence, current driver coordinates and candidate discovery. |
| Ride Service | ⏳ | Will own post-agreement execution state. |
| Identity Service | ⏳ | Auth/identity currently remains in compatibility runtime. |
| Payment Service | ⏳ | Payment logic currently remains in compatibility runtime. |
| Trust Service | ⏳ | Ratings/trust currently remains in compatibility runtime. |
| Realtime Service | ⏳ | Realtime transport currently remains in compatibility runtime. |
| Notification Service | ⏳ | Target service only. |
| Operator Service | ⏳ | Existing admin app consumes compatibility APIs; independent service not extracted. |

## Applications

| Application | Status | Notes |
|---|---|---|
| Rider Flutter app | 🔁 | Existing app foundation is tied primarily to V1 compatibility flow. |
| Driver Flutter app | 🔁 | Existing app foundation is tied primarily to V1 compatibility flow. |
| Operator/Admin Next.js app | 🔁 | Existing operations console uses compatibility API. |
| Landing page | ✅ | Public OpenRide positioning/branding exists. |

## Infrastructure

| Capability | Status | Notes |
|---|---|---|
| PostgreSQL/PostGIS compatibility stack | ✅ | Existing runtime persistence. |
| Redis compatibility stack | ✅ | Existing geo/hot-state implementation. |
| Marketplace PostgreSQL | ✅ | Dedicated development database in microservices compose topology. |
| NATS JetStream | ✅ | Development broker topology is defined. |
| Marketplace Docker image | ✅ | Independently buildable service image. |
| Service-level readiness | 🟡 | Marketplace currently checks dependency reachability; protocol-level DB/NATS health will mature with real clients. |
| OpenTelemetry | ⏳ | Architectural requirement, not yet wired across services. |
| Kubernetes / service mesh | ⏳ | Deliberately not required for the current pre-1.0 phase. |

## What OpenRide does **not** claim today

OpenRide does not currently claim:

- production readiness for carrying real passengers;
- completed multi-service ride booking from request through settlement;
- fully extracted microservices for every bounded context;
- production-grade fraud, insurance, KYC, incident response or local regulatory compliance;
- complete NATS outbox relay / saga implementation;
- multi-region/high-availability deployment;
- federation between independent OpenRide operators.

Those are explicit work items, not hidden gaps.

## Current migration path

```text
Legacy V1 runtime
      │
      │ compatibility edge
      ▼
Marketplace Service ✅/🟡
      │
      ├── Location Service ⏳
      ├── Ride Service ⏳
      ├── Identity Service ⏳
      ├── Payment Service ⏳
      └── Trust / Realtime / Notification / Operator ⏳
```

A capability moves out of the compatibility runtime only after the owning service has a tested data/API/event path and a rollback strategy.

## Definition of public credibility

Before marking a capability ✅, OpenRide expects at least:

1. an owning package/service;
2. explicit data ownership;
3. tests for core invariants;
4. a runnable local path;
5. documented API/event contracts when crossing a process boundary;
6. failure/idempotency behavior for critical commands.

If one of those is missing, this document should say so.
