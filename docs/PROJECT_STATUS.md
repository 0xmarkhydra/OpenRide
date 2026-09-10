# OpenRide Project Status

> Last reviewed: 2026-09-11

OpenRide is a **pre-1.0 open-source mobility marketplace** under active architectural migration. This document is intentionally conservative: it separates code that exists from capabilities that still lack end-to-end production proof.

## Status legend

- ✅ **Implemented** — code exists with an executable/testable path for the stated scope.
- 🟡 **Foundation** — meaningful implementation exists, but production integration, failure proof, or migration is incomplete.
- 🔁 **Compatibility** — legacy runtime still serves this capability while traffic is migrated.
- ⏳ **Planned** — architectural target only; do not assume it exists.

## Marketplace platform

| Capability | Status | Notes |
|---|---|---|
| OpenRide Core domain package | ✅ | Independent `packages/core-go` module with marketplace value objects, integer-minor-unit money, pricing/ranking ports and invariant tests. |
| Service module registry | ✅ | `passenger.car` and `carpool.intercity` first-party modules. |
| Marketplace Service process | ✅ | Independent Go process with real PostgreSQL and NATS clients, health/readiness and V2 routes. |
| Marketplace-owned PostgreSQL schema | ✅ | Dedicated migrations and runtime store own tariffs, requests, quotes, agreements, outbox and idempotency state. |
| Driver tariff persistence | ✅ | V2 create/list path persists driver-owned tariffs. Public wire DTO uses flat minor-unit fields. |
| Mobility request persistence | ✅ | V2 request creation/read/cancel path is persisted. Cancel is transactionally coupled with pending-quote invalidation. |
| Quote persistence | ✅ | Driver quote submission and withdrawal are persisted. Database guards reject pending quotes for requests that are no longer open. |
| Quote acceptance / Agreement | 🟡 | Atomic PostgreSQL transaction creates one Agreement, changes request/quote state, stores acceptance idempotency and appends an outbox event. Acceptance locks Request before Quote and retries PostgreSQL serialization/deadlock failures up to three attempts. A 100-way PostgreSQL integration test path exists, but hosted runner infrastructure has not executed it yet. |
| Agreement commercial snapshot | 🟡 | Accepted quote metadata is detached into a typed snapshot and the Agreement table is append-only at the database layer. Canonical content hashing/signing is not implemented. |
| Ranking engine | 🟡 | Core ranking is deterministic/explainable, isolates rankers from canonical quote data, and the V2 rider-offers path now ranks persisted selectable quotes through the hardened Core entrypoint. Executable service CI proof is still blocked by hosted runner availability. |
| Transactional outbox relay | 🟡 | PostgreSQL `SKIP LOCKED` relay publishes to JetStream with deterministic `Nats-Msg-Id`. Retry backoff/dead-letter handling and consumer inbox dedupe are still incomplete. |
| JavaScript / TypeScript SDK | 🟡 | Fetch-based SDK matches the current flat V2 tariff/quote/offer contract and rejects unsafe money inputs. Full SDK↔service E2E verification is still required. |
| Generic command idempotency | 🟡 | Create request, create tariff, submit quote, cancel request and withdraw quote persist payload hash, resource identity, response status and response snapshot in the same PostgreSQL transaction as the mutation. Same-key/same-payload retries replay the original response; same-key/different-payload returns `409 IDEMPOTENCY_CONFLICT`. Real PostgreSQL integration proof is still required. Acceptance retains its specialized durable Agreement replay path. |

## Services

| Service | Status | Data ownership |
|---|---|---|
| Compatibility API / edge | 🔁 | Owns legacy V1 runtime data during migration. |
| Marketplace Service | 🟡 | Owns Marketplace V2 state in its dedicated database. It must remain behind a trusted gateway because actor identity is currently supplied through an internal header. |
| Location Service | ⏳ | Will own online presence, current driver coordinates and candidate discovery. |
| Ride Service | ⏳ | Will own post-agreement execution state. |
| Identity Service | ⏳ | Auth/identity currently remains in compatibility runtime. |
| Payment Service | ⏳ | Payment logic currently remains in compatibility runtime. |
| Trust Service | ⏳ | Ratings/trust currently remains in compatibility runtime. |
| Realtime Service | ⏳ | Realtime transport currently remains in compatibility runtime. |
| Notification Service | ⏳ | Target service only. |
| Operator Service | ⏳ | Existing admin app consumes compatibility API; independent service not extracted. |

## Applications

| Application | Status | Notes |
|---|---|---|
| Rider Flutter app | 🔁 | Existing app foundation is tied primarily to V1 compatibility flow. Do not migrate until Marketplace V2 E2E and concurrency tests are green. |
| Driver Flutter app | 🔁 | Existing app foundation is tied primarily to V1 compatibility flow. |
| Operator/Admin Next.js app | 🔁 | Existing operations console uses compatibility API. |
| Landing page | ✅ | Public OpenRide positioning/branding exists. |

## Infrastructure

| Capability | Status | Notes |
|---|---|---|
| PostgreSQL/PostGIS compatibility stack | ✅ | Existing runtime persistence. |
| Redis compatibility stack | ✅ | Existing geo/hot-state implementation. |
| Marketplace PostgreSQL | ✅ | Dedicated database and append-only migrations in the microservices topology. |
| NATS JetStream | ✅ | Marketplace stream bootstrap exists for `openride.marketplace.>` subjects. |
| Marketplace Docker image | ✅ | Independently buildable service image path exists. A fresh networked build/test run is still required after recent dependency and Marketplace changes. |
| Service-level readiness | ✅ | Marketplace readiness performs PostgreSQL ping and verifies NATS connection state. |
| OpenTelemetry | ⏳ | Architectural requirement, not yet wired across services. |
| Kubernetes / service mesh | ⏳ | Deliberately not required for the current pre-1.0 phase. |

## Known P0/P1 gaps

1. Execute the real PostgreSQL integration suites for generic idempotency replay/conflict and 100-way competing quote acceptance once CI runner capacity is restored.
2. Put Marketplace behind an authenticated Edge/BFF that strips client-supplied internal actor headers and reconstructs trusted actor context.
3. Add outbox retry backoff, dead-letter/operator visibility and consumer-side inbox dedupe.
4. Run SDK↔Marketplace E2E tests and enforce one documented money range/serialization rule across all public clients.
5. Add production observability and operational evidence for ranking/idempotency/concurrency paths before upgrading their status beyond Foundation.

## What OpenRide does **not** claim today

OpenRide does not currently claim:

- production readiness for carrying real passengers;
- completed multi-service ride booking from request through settlement;
- fully extracted microservices for every bounded context;
- production-proven generic idempotency under real PostgreSQL concurrency;
- production-proven V2 ranking behavior under service-level load;
- cryptographically signed or hashed Agreements;
- proven correctness under high-contention PostgreSQL acceptance races;
- production-grade fraud, insurance, KYC, incident response or local regulatory compliance;
- complete saga/consumer coverage across NATS;
- multi-region/high-availability deployment;
- federation between independent OpenRide operators.

Those are explicit work items, not hidden gaps.

## Current migration path

```text
Legacy V1 runtime
      │
      │ compatibility edge / future authenticated BFF
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
