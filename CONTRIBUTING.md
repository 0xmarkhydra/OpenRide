# Contributing to OpenRide

Thank you for helping build OpenRide.

OpenRide is intended to become community-owned mobility infrastructure. Contributions should preserve technical rigor, explicit ownership boundaries and the marketplace principles that distinguish OpenRide from a conventional platform-controlled ride-hailing clone.

## Read this first

Before a non-trivial contribution, read:

1. [`README.md`](README.md)
2. [`docs/PROJECT_STATUS.md`](docs/PROJECT_STATUS.md) — implemented vs foundational vs planned
3. [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md)
4. [`docs/MICROSERVICES_ARCHITECTURE.md`](docs/MICROSERVICES_ARCHITECTURE.md)
5. [`docs/PACKAGE_ARCHITECTURE.md`](docs/PACKAGE_ARCHITECTURE.md)
6. [`docs/VERSIONING.md`](docs/VERSIONING.md)
7. [`docs/DOMAIN_MODEL.md`](docs/DOMAIN_MODEL.md)
8. [`docs/OPENRIDE_MIGRATION_PLAN_V2.md`](docs/OPENRIDE_MIGRATION_PLAN_V2.md)

Do not infer implementation status from an architecture diagram. `PROJECT_STATUS.md` is the authoritative implementation-status document.

## Non-negotiable marketplace constraints

A contribution should not casually introduce behavior that:

- silently overrides driver-owned commercial terms;
- ranks only by lowest fare;
- automatically punishes drivers merely for skipping unsuitable work;
- mutates accepted commercial terms after an Agreement is created;
- creates a hidden mandatory dependency on one operator/provider;
- collapses Request, Quote, Agreement and Ride back into one giant mutable object;
- allows a ranking extension to fabricate, hide or mutate canonical Quotes.

If a proposal intentionally revisits one of these principles, open an RFC and document the trade-off explicitly.

## Architecture standard

OpenRide uses **microservices at deployable bounded-context boundaries** and clean/hexagonal architecture inside each service.

A service must have a business ownership reason to exist. Do not create empty services merely to match the target architecture diagram.

Hard rules:

```text
one service owns one bounded context
service owns its data + migrations
no cross-service SQL reads
no cross-service database foreign keys
sync cross-service calls only when the caller needs an immediate answer
async integration through versioned events / NATS JetStream
state + event publication through transactional Outbox
consumer side effects through Inbox/idempotency
multi-service business workflows through Saga/process managers
```

Do not share persistence repositories across services. Shared packages may contain stable value objects, contracts and carefully bounded pure domain invariants, not another service's state access.

The existing `services/api` runtime is a **compatibility boundary**, not the target home for new Marketplace V2 business logic.

## Choosing where code belongs

Use this decision rule:

1. **Universal pure invariant/value object?** → `packages/core-go`.
2. **Mobility vertical behavior?** → service module under `packages/modules-go` when it can stay infrastructure-neutral.
3. **Wire/event schema?** → `packages/contracts`, versioned explicitly.
4. **Marketplace-owned state/use case?** → `services/marketplace`.
5. **Legacy V1 fix needed during migration?** → `services/api`, clearly marked compatibility where relevant.
6. **Future bounded context not extracted yet?** → prefer an RFC/design issue before creating a new deployable service folder.

## Database ownership and migrations

Each extracted service owns its migrations.

Rules:

- never rewrite a migration already merged to `main`;
- add the next numbered migration;
- prefer backward-compatible expansion before contraction;
- add/backfill/validate restrictive constraints deliberately;
- use database constraints for critical invariants where practical;
- never make another service query this database directly;
- accepted commercial truth must remain durable;
- Redis/NATS are not substitutes for the owning service's durable state.

For Marketplace migrations read [`services/marketplace/migrations/README.md`](services/marketplace/migrations/README.md).

## API and event contracts

Breaking wire changes require a new contract major version.

Examples:

```text
service-manifest.v1
marketplace-event.v1
openride.marketplace.agreement.created.v1
```

Do not silently change the meaning of an existing versioned event/schema.

When a PR changes a public command/query/event shape, update the corresponding documentation/schema in the same PR.

## Reliability requirements

Critical commands such as request creation, quote submission/withdrawal, quote acceptance, payment and settlement must define retry/idempotency behavior.

For cross-service state changes:

```text
DB transaction
  ├── owned state change
  └── outbox row
COMMIT
      ↓
relay → NATS JetStream
      ↓
inbox dedupe → consumer-owned state
```

Do not introduce distributed database transactions.

Document degraded behavior when a dependency such as Location, NATS, routing or payment provider is unavailable.

## Pull requests

Use the repository PR template. A strong PR explains:

- the problem, not only the implementation;
- owning bounded context/package;
- data ownership changes;
- API/event contract changes;
- marketplace principle impact;
- reliability/idempotency behavior;
- migration/compatibility/rollback path;
- tests and validation performed.

Keep PRs reviewable. Large migrations should be decomposed into coherent vertical slices rather than arbitrary file-count slices.

## Tests and validation

### Portable packages

```bash
make packages-test
```

### Marketplace Service

```bash
make marketplace-test
```

To prove the service is not accidentally rescued by the monorepo workspace:

```bash
cd services/marketplace
GOWORK=off go mod download
GOWORK=off go test ./...
GOWORK=off go vet ./...
```

### Compatibility API

```bash
make api-test
```

### Microservices topology

```bash
make micro-config
make micro-up
# check health/readiness and logs
make micro-logs
make micro-down
```

### Flutter apps

Run in the relevant app:

```bash
flutter pub get
dart analyze
flutter test
```

### Operator/Admin

```bash
cd apps/admin
npm ci
npm run build
```

CI also builds service images and boots the first microservices development slice. If CI infrastructure itself is unavailable, state that explicitly; do not report a check as green when it did not execute.

## Multi-module versioning

Nested Go modules require path-prefixed module tags or valid pseudo-versions. Never commit a fake `v0.0.0` dependency that succeeds only because `go.work` happens to replace resolution locally.

See [`docs/VERSIONING.md`](docs/VERSIONING.md).

## Updating project status

If your PR changes a capability from planned → foundation or foundation → implemented, update [`docs/PROJECT_STATUS.md`](docs/PROJECT_STATUS.md).

A capability should not be marked ✅ merely because:

- an interface exists;
- a table exists;
- an architecture document describes it;
- an empty service process can start.

At minimum it should have an owning implementation, tests, runnable path and documented boundary/failure behavior appropriate to that capability.

## Compatibility debt

Some `flashx` identifiers remain intentionally to preserve old deployments, data volumes, mobile identifiers and V1 clients. Do not copy those identifiers into new V2 code.

Read [`docs/LEGACY_COMPATIBILITY.md`](docs/LEGACY_COMPATIBILITY.md) before renaming or adding to compatibility paths.

## Security and privacy

Do not commit:

- credentials or production tokens;
- private KYC documents;
- rider/driver phone numbers or exact trip data from production;
- private GPS traces;
- payment secrets.

Security vulnerabilities should follow [`SECURITY.md`](SECURITY.md), not a public issue with exploitable details.

## Community standard

Challenge architecture and product decisions with evidence. Be rigorous about trade-offs, but keep criticism focused on code, behavior and decisions rather than contributors.
