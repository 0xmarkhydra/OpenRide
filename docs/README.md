# OpenRide Documentation

**English** · [简体中文](i18n/zh-CN/README.md) · [हिन्दी](i18n/hi/README.md) · [Español](i18n/es/README.md)

This directory is the documentation source of truth for OpenRide.

> The previous index still described the legacy FlashX product and its three-service MVP. That index is no longer authoritative for OpenRide. Historical documents must be explicitly classified before they are used to implement new OpenRide behavior.

## Four-language policy

All maintained public OpenRide documentation is published in:

1. English (`en`) — canonical technical source.
2. Simplified Chinese (`zh-CN`).
3. Hindi (`hi`).
4. Spanish (`es`).

English source documents remain at `docs/*.md`. Translations mirror the same filename under:

```text
docs/i18n/zh-CN/
docs/i18n/hi/
docs/i18n/es/
```

A translation must preserve API paths, identifiers, event names, JSON fields, code blocks, currencies and invariant semantics. If a translation temporarily lags a code change, it must be marked stale; it must not silently present old behavior as current.

## OpenRide source-of-truth order

1. `PROJECT_STATUS.md` — implemented vs foundational vs planned.
2. `PRODUCT_VISION.md` — product direction and marketplace philosophy.
3. `OPENRIDE_MANIFESTO.md` — non-negotiable ecosystem principles.
4. `MICROSERVICES_ARCHITECTURE.md` — target service ownership/boundaries.
5. `PACKAGE_ARCHITECTURE.md` — portable Core/modules/contracts/SDK boundaries.
6. `DOMAIN_MODEL.md` — domain objects and state transitions.
7. `DATA_MODEL.md` — persistence model.
8. `API_CONTRACT_V2.md` — V2 public contract target.
9. `OPENRIDE_MIGRATION_PLAN_V2.md` — staged migration from compatibility runtime.
10. `ADR_OPENRIDE_V2.md` — V2 architecture decisions.

## Marketplace invariant

Each driver owns their own tariff. `per_km` is configured **per driver + service tariff**, not once for every driver in a country.

```text
Driver A → own per_km / minimum / pickup fee / quote mode
Driver B → own per_km / minimum / pickup fee / quote mode
Driver C → own per_km / minimum / pickup fee / quote mode
```

Country/instance configuration may define currency, regulation, safety constraints and operator limits. It does not replace driver-owned commercial pricing.

## Translation coverage rule

Every maintained public `docs/*.md` document must have a matching translation using the same filename in all three translation directories. New product/architecture/API changes should update affected translations in the same PR. Historical FlashX-specific documents must be marked historical/deprecated or migrated before being treated as OpenRide requirements.

## Contribution rule

Do not claim a diagram, translated document or roadmap item is implemented unless `PROJECT_STATUS.md` and executable code/tests support that claim. Documentation must distinguish current implementation from target architecture.
