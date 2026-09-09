# OpenRide Package Architecture

OpenRide is being shaped as a **mobility marketplace framework**, not a single deployment that every contributor must fork and rewrite.

The repository therefore separates four concepts:

```text
┌───────────────────────────────────────────────────────────────┐
│                         Applications                          │
│ Rider • Driver • Operator • Community-specific frontends     │
├───────────────────────────────────────────────────────────────┤
│                       Operator Runtime                        │
│ API • auth • persistence • realtime • jobs • observability   │
├───────────────────────────────────────────────────────────────┤
│                         OpenRide Core                         │
│ marketplace domain • invariants • engine • extension ports   │
├───────────────────────────────────────────────────────────────┤
│                 Modules / Adapters / Contracts                │
│ service modules • Redis/PostGIS • routing • payment • schemas │
└───────────────────────────────────────────────────────────────┘
```

## 1. What belongs in Core

Core should contain business rules that remain true regardless of infrastructure or operator:

- request, tariff, quote, agreement and ride concepts;
- money and geography value objects;
- state-machine invariants;
- service-module contracts;
- candidate -> quote -> rank orchestration;
- versioned domain-event contracts;
- rules that protect accepted terms from silent mutation.

Core must **not** import PostgreSQL, Redis, HTTP routers, WebSocket libraries, map SDKs, payment SDKs or cloud-provider SDKs.

The first package is:

```text
packages/core-go
```

Its canonical Go module is:

```text
github.com/0xmarkhydra/OpenRide/packages/core-go
```

## 2. What belongs in a Service Module

A service module describes a mobility vertical without changing the marketplace kernel.

Examples:

```text
passenger.car
passenger.bike
carpool.intercity
parcel.instant
designated-driver.car
vehicle.inspection-assist
```

Each module exposes a manifest:

```json
{
  "id": "passenger.car",
  "version": "1.0.0",
  "display_name": "Passenger Car",
  "category": "passenger",
  "capabilities": ["passenger", "scheduled"]
}
```

The module may validate service-specific request attributes, lifecycle rules or commercial constraints. It must not weaken Core invariants such as accepted-price snapshotting.

A new vertical should normally be added by registering a module, **not by adding another boolean or nullable column to every shared table**.

## 3. What belongs in an Adapter

Adapters connect Core ports to infrastructure.

Typical adapters:

```text
CandidateSource -> Redis GEO / PostGIS
QuoteProvider   -> tariff store + manual quote channel
Ranker          -> default fairness ranker / operator ranker
EventPublisher  -> Redis Streams / Kafka / NATS
Router          -> OSRM / Mapbox / Google Maps / custom
PaymentGateway  -> Stripe / MoMo / VNPay / community cash flow
ObjectStorage   -> S3 / MinIO / local
```

The existing `services/api/internal/*` code is treated as the current operator runtime and adapter source. It will migrate toward these ports gradually instead of being deleted in a big-bang rewrite.

## 4. What belongs in Contracts

`packages/contracts` is language-neutral. It exists so Dart, TypeScript, Rust or another runtime can share stable wire/domain contracts without importing Go code.

Contracts are versioned explicitly:

```text
service-manifest.v1
marketplace-event.v1
```

Breaking wire changes create a new version. Old versions may remain supported during a migration window.

## 5. Dependency direction

Allowed:

```text
apps -> runtime -> core
runtime -> adapters -> core ports
modules -> core
SDKs -> contracts
core -> standard library only (plus intentionally tiny pure dependencies if ever approved)
```

Avoid:

```text
core -> runtime
core -> database
core -> web framework
core -> one operator's pricing policy
module A -> module B internals
```

This rule keeps OpenRide usable by communities that do not share our deployment choices.

## 6. Model evolution

Shared marketplace objects intentionally provide generic `Attributes`, `Constraints` and metadata snapshots at extension boundaries. This is not permission to put arbitrary data everywhere.

The rule is:

1. If a concept is universal and has invariants, promote it into a typed Core field/value object.
2. If a concept belongs to one vertical, keep it in the service module contract.
3. If a concept belongs to one infrastructure provider, keep it in the adapter.
4. If a concept crosses processes/languages, define or version it in `packages/contracts`.

## 7. Versioning policy

Before `1.0.0`, package APIs may evolve quickly, but every breaking change must include:

- why the old model was insufficient;
- migration notes;
- any contract/event version changes;
- whether stored data needs a migration;
- whether adapters need an update.

After `1.0.0`, public Core packages follow semantic versioning.

Domain events always carry an explicit name version, for example:

```text
marketplace.offers_ready.v1
marketplace.agreement_created.v1
ride.status_changed.v1
```

## 8. Packaging roadmap

```text
Phase A  ✅ core-go foundation + contracts
Phase B     move marketplace runtime behavior behind Core ports
Phase C     reference passenger-car service module
Phase D     default PostGIS/Redis/routing/payment adapter packages
Phase E     generated TypeScript + Dart contracts/SDKs
Phase F     operator starter distribution + Docker/Helm templates
Phase G     stable Core v1
```

## 9. Definition of a good extension

A community should be able to say:

> "We want to run OpenRide with our own ranking algorithm, local payment method and a carpool service."

and the answer should be:

> "Implement these ports/modules and register them. Do not fork the marketplace kernel."

That is the architectural standard for OpenRide.
