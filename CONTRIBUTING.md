# Contributing to OpenRide

Thank you for helping build OpenRide.

OpenRide is intended to be community-owned mobility infrastructure, so contributions should preserve both technical quality and the project's marketplace principles.

## Before contributing

Please read:

1. [`README.md`](README.md)
2. [`docs/OPENRIDE_MANIFESTO.md`](docs/OPENRIDE_MANIFESTO.md)
3. [`docs/PRODUCT_VISION.md`](docs/PRODUCT_VISION.md)
4. [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)
5. [`docs/DOMAIN_MODEL.md`](docs/DOMAIN_MODEL.md)
6. [`docs/OPENRIDE_MIGRATION_PLAN_V2.md`](docs/OPENRIDE_MIGRATION_PLAN_V2.md)

## Non-negotiable product constraints

A contribution should not casually introduce behavior that:
- silently overrides driver-owned pricing;
- ranks only by lowest fare;
- automatically punishes drivers merely for skipping unsuitable work;
- mutates accepted commercial terms after agreement;
- creates a hidden dependency on one mandatory operator/provider;
- collapses Request, Quote, Agreement and Ride back into one giant business object.

If a change intentionally revisits one of these principles, open an architecture/product discussion and document the decision explicitly.

## Development approach

OpenRide currently favors a Go modular monolith with clear domain boundaries.

Please avoid adding infrastructure complexity without demonstrated need, including:
- new microservices;
- Kafka/NATS;
- Kubernetes;
- service mesh;
- global event sourcing;
- ML ranking.

Such changes are welcome when backed by a concrete production requirement and migration plan.

## Working on the marketplace migration

The current repository still contains compatibility code from earlier ride-hailing/FlashX versions.

Follow the staged migration plan rather than deleting old flows immediately.

Preferred order:

```text
MobilityRequest
-> DriverTariff
-> Quote
-> Marketplace Ranking
-> Agreement
-> Ride
```

Each phase should:
- add tests;
- preserve currently supported runtime behavior until replacement is ready;
- use additive DB migrations;
- keep API changes explicit/documented;
- keep critical operations idempotent.

## Database changes

Rules:
- never rewrite a migration that has already shipped;
- create a new numbered migration;
- prefer backward-compatible additions first;
- backfill before adding restrictive constraints when needed;
- keep accepted price/agreement state durable in PostgreSQL;
- keep Redis for hot/ephemeral state, not as the only business source of truth.

## Pull requests

A good PR should explain:
- problem being solved;
- product/domain impact;
- architecture impact;
- migrations/API changes;
- tests added/changed;
- compatibility/rollout concerns.

Keep PRs focused when possible.

## Tests

For backend changes run:

```bash
cd services/api
go test ./...
go vet ./...
```

For Flutter apps run the relevant:

```bash
flutter pub get
dart analyze
flutter test
```

For Admin:

```bash
cd apps/admin
npm ci
npm run build
```

Persistent integration tests use PostgreSQL/PostGIS + Redis through the existing repository tooling/CI.

## Documentation

Business-changing code requires documentation updates in the same PR.

At minimum consider whether the change affects:
- PRODUCT_VISION;
- CUSTOMER_REQUIREMENTS;
- DOMAIN_MODEL;
- ARCHITECTURE;
- DATA_MODEL;
- Marketplace Matching Engine;
- API contract;
- migration plan.

## Security

Do not commit credentials, private KYC data, production tokens or user secrets.

Security-sensitive findings should not be published with exploitable details before maintainers can coordinate a fix.

## Community standard

Debate architecture and product decisions strongly, but treat contributors respectfully. Focus criticism on code, behavior, trade-offs and evidence.
