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
| Driver tariff persistence | ✅ | V2 create/list path persists driver-owned tariffs. Public input uses flat minor-unit fields. |
| Mobility request persistence | ✅ | V2 request creation/read/cancel path is persisted. Cancel is transactionally coupled with pending-quote invalidation. |
| Quote persistence | ✅ | Driver quote submission and withdrawal are persisted. Database guards reject pending quotes for requests that are no longer open. |
| Quote acceptance / Agreement | 🟡 | Atomic PostgreSQL transaction creates one Agreement, changes request/quote state, stores acceptance idempotency and appends an outbox event. Acceptance locks Request before Quote and retries PostgreSQL serialization/deadlock failures up to three attempts. A 100-way PostgreSQL integration test path exists, but hosted runner infrastructure has not executed it yet. |
| Agreement commercial snapshot | 🟡 | Accepted quote metadata is detached into a typed snapshot and the Agreement table is append-only at the database layer. Canonical content hashing/signing is not implemented. |
| Ranking engine | 🟡 | Core ranking is deterministic/explainable, isolates rankers from canonical quote data, and the V2 rider-offers path ranks persisted selectable quotes through the hardened Core entrypoint. Executable service CI proof is still blocked by hosted runner availability. |
| Transactional outbox relay | 🟡 | Relay claims work with short DB leases, publishes outside the claim transaction, retries with bounded exponential backoff, dead-letters after 12 failed attempts and keeps deterministic `Nats-Msg-Id`. Consumer inbox processing is still not wired because there is not yet an extracted Marketplace event consumer. |
| Public money contract | 🟡 | V2 defines every `*_minor` JSON field as an exact non-negative integer in `0..2^53-1`. HTTP validates nested minor fields before float decoding, PostgreSQL enforces the same ceiling for persisted tariff/quote/Agreement money, and the JS SDK rejects unsafe inputs before network I/O. Executable service/SDK CI proof is still blocked by hosted runners. |
| JavaScript / TypeScript SDK | 🟡 | Fetch-based SDK matches the current request/quote/offer/Agreement V2 paths. A full Compose E2E path now drives the real SDK through authenticated Edge HTTP into Marketplace/PostgreSQL/NATS, but hosted runners have not executed it yet. |
| Generic command idempotency | 🟡 | Create request, create tariff, submit quote, cancel request and withdraw quote persist payload hash, resource identity, response status and response snapshot in the same PostgreSQL transaction as the mutation. The full E2E suite exercises replay/conflict for these commands against real PostgreSQL, but that suite has not executed while hosted runners are unavailable. Acceptance retains its specialized durable Agreement replay path. |

## Services

| Service | Status | Data ownership |
|---|---|---|
| Compatibility API / edge | 🟡 | Owns legacy V1 runtime data and is the public V2 boundary. It authenticates bearer tokens, enforces rider/driver roles, strips client `X-OpenRide-*` headers, reconstructs actor identity and authenticates itself to Marketplace with a separate shared gateway secret. |
| Marketplace Service | 🟡 | Owns Marketplace V2 state in its dedicated database. Runtime requires `MARKETPLACE_GATEWAY_TOKEN`; direct `/v2/*` calls without the trusted Edge secret are rejected even if an actor header is spoofed. Network-private deployment is still recommended. |
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
| Rider Flutter app | 🔁 | Existing app foundation is tied primarily to V1 compatibility flow. Do not migrate until Marketplace V2 E2E and concurrency tests actually execute successfully. |
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
| Marketplace Docker image | ✅ | Independently buildable service image path exists. The dev microservices port is bound to loopback only and V2 additionally requires the gateway trust secret. |
| Full SDK→Edge→Marketplace E2E topology | 🟡 | `docker-compose.e2e.yml` boots real compatibility persistence/auth, Edge, Marketplace PostgreSQL migrations, NATS and Marketplace; the Node E2E script covers auth, actor anti-spoofing, ranking, money and all current idempotent commands. Awaiting an executing CI runner. |
| Service-level readiness | ✅ | Marketplace readiness performs PostgreSQL ping and verifies NATS connection state. |
| OpenTelemetry | ⏳ | Architectural requirement, not yet wired across services. |
| Kubernetes / service mesh | ⏳ | Deliberately not required for the current pre-1.0 phase. |

## Known P0/P1 gaps

1. Restore/replace CI runner capacity and actually execute the 100-way PostgreSQL acceptance test plus the full SDK→Edge→Marketplace E2E suite. Until then, concurrency/idempotency/ranking remain implemented but not execution-proven in this repository environment.
2. Add structured metrics/tracing and an operator surface for dead-lettered outbox events, idempotency conflicts/retries and PostgreSQL concurrency retries.
3. Wire consumer inbox dedupe when the first extracted NATS consumer is introduced; the schema exists but there is no consumer lifecycle to protect yet.
4. Decide whether Agreement snapshots need canonical hashing/signing before real-money or regulated deployments.
5. Migrate Rider/Driver Flutter flows from V1 only after the V2 E2E/concurrency evidence is green; do not multiply service boundaries before that proof exists.

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
Public clients
      │
      ▼
Compatibility Edge / Auth boundary 🟡
      │ bearer auth + trusted gateway secret
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
