# OpenRide Contracts

Language-neutral, versioned schemas for service modules, domain events and future SDK generation.

These contracts intentionally live outside `core-go` so Flutter/Dart, TypeScript, Rust or other runtimes can integrate without depending on Go implementation details.

## Current contracts

```text
schemas/service-manifest.v1.schema.json
schemas/marketplace-event.v1.schema.json
```

## Rules

- contract filenames include a major version;
- breaking wire changes create a new version instead of mutating an old meaning;
- unknown additive fields should be tolerated where the schema allows them;
- service-specific payloads belong under explicit extension fields, not hidden overloaded primitives;
- event names themselves are versioned, e.g. `marketplace.offers_ready.v1`.

Future generated SDKs should treat this directory as their source of truth.
