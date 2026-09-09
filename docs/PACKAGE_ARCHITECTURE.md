# OpenRide Package Architecture

OpenRide uses microservices for deployment boundaries and packages for stable reusable contracts/invariants.

These are complementary, not competing ideas.

```text
Applications
   |
   v
Edge Gateway / BFF
   |
   +--> independently deployable domain services
           |
           +--> shared pure packages where appropriate
           +--> service-owned persistence/adapters
           +--> versioned contracts/events
```

## 1. Package rule

A shared package is allowed only when it does **not** create shared service ownership.

Good shared package content:
- money/geo value objects;
- Request/Quote/Agreement/Ride invariants;
- pure ranking/pricing interfaces;
- service-module interfaces;
- versioned JSON schemas;
- client SDK transport helpers.

Forbidden shared package content:
- repositories that access service databases;
- DB models used as cross-service persistence contracts;
- global transaction managers;
- service-owned caches;
- hidden service-to-service calls;
- business workflows owned by a bounded context.

## 2. Current packages

```text
packages/
├── core-go/       # pure marketplace kernel/invariants
├── modules-go/    # first-party mobility vertical modules
├── contracts/     # versioned language-neutral wire contracts
└── sdk/           # zero-dependency JS/TS public API client
```

## 3. Core

Canonical module:

```text
github.com/0xmarkhydra/OpenRide/packages/core-go
```

Core must not import:
- PostgreSQL;
- Redis;
- NATS;
- HTTP routers;
- gRPC servers;
- WebSocket frameworks;
- payment/map/cloud SDKs.

Core owns business invariants that remain valid no matter which service deployment or provider is used.

Examples:
- money is integer minor units;
- quotes belong to requests/drivers;
- accepted terms are immutable snapshots;
- ranking may score/order but cannot rewrite canonical quote terms;
- state transitions are explicit.

## 4. Service modules

A service module models a mobility vertical without redefining the marketplace kernel.

Examples:

```text
passenger.car
carpool.intercity
passenger.motorbike
parcel.instant
designated-driver.car
vehicle-assistance
```

A module may provide:
- manifest metadata;
- request validation;
- service-specific attributes/contracts;
- execution lifecycle policy.

A module does not own database connectivity. The bounded-context service using that module owns persistence.

## 5. Contracts

`packages/contracts` is language-neutral.

Use it for data that crosses process/language boundaries:

```text
service-manifest.v1
marketplace-event.v1
request.passenger-car.v1
request.carpool-intercity.v1
```

Breaking wire semantics create a new major contract version.

Persisted service tables are **not** contracts between services.

## 6. SDK

`@openride/sdk` targets public Edge/BFF APIs.

It must not know internal service topology. A client should not need to know whether `/v2/requests` is served directly by Marketplace or proxied by the Edge.

This lets service extraction happen without forcing every client release to change network topology.

## 7. Service dependency direction

Allowed:

```text
marketplace-service -> core-go
marketplace-service -> modules-go
marketplace-service -> contracts
ride-service        -> core-go/modules-go where invariants apply
edge-gateway        -> SDK/client contracts only where useful
apps                -> public SDK/contracts
```

Avoid:

```text
marketplace-service -> ride-service database package
ride-service -> marketplace repository
service A -> service B internal package
core-go -> any deployable service
shared package -> NATS subjects + DB repository + business workflow all mixed together
```

## 8. Service-local code

Inside each service, prefer:

```text
services/<name>/
├── cmd/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── ports/
│   └── adapters/
├── migrations/
├── Dockerfile
└── go.mod
```

Not every small service needs every folder immediately, but dependency direction should remain clear:

```text
transport/adapters -> application -> domain
                         |
                         v
                       ports
```

## 9. Cross-service integration

Packages are not a substitute for runtime integration.

Use:
- gRPC for immediate internal queries/commands when necessary;
- NATS JetStream for asynchronous integration;
- Outbox/Inbox for reliable delivery/deduplication;
- Saga/process manager for multi-service workflows.

Do not import another service's Go package and call its application service in-process merely because the monorepo makes that easy.

## 10. Monorepo rule

OpenRide may remain a monorepo while still being microservices.

Monorepo advantages:
- atomic contract updates;
- shared CI;
- easier contributor onboarding;
- one place for docs/apps/packages.

But every deployable service must still have:
- independent entrypoint;
- independent Docker image;
- independent database ownership;
- independent migrations;
- explicit network/event contracts.

Monorepo does not mean monolith.

## 11. Versioning

Before 1.0:
- package APIs may evolve;
- breaking changes need migration notes;
- wire/event breaking changes require version bumps;
- service DB changes use append-only migrations.

After Core 1.0, public packages follow semantic versioning.

Events always carry explicit major versions:

```text
openride.marketplace.quote.created.v1
openride.marketplace.agreement.created.v1
openride.ride.ride.completed.v1
```

## 12. Good extension test

A community should be able to say:

> We want our own ranking policy, local payment provider and carpool vertical.

The architecture should answer:

```text
ranking policy -> implement Marketplace Core port/policy
carpool -> service module
local payment -> Payment Service adapter
service integration -> versioned contract/event
```

not:

> Fork the whole backend and edit a giant Trip service.

That is the package architecture standard for OpenRide.
