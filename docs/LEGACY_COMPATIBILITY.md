# Legacy compatibility during the OpenRide V2 migration

OpenRide is being migrated from an earlier FlashX codebase. Public product language and new domain work use the OpenRide name, but some internal identifiers intentionally remain unchanged until their replacements can be shipped without breaking existing deployments.

Examples may include:

- the Go module path `flashx/services/api`;
- Redis key prefixes using `flashx`;
- existing database names or environment variables;
- Flutter class/theme names;
- Android/iOS bundle/package identifiers;
- the Admin compatibility proxy path `/api/flashx`;
- local-storage keys used by existing sessions.

These are **implementation compatibility identifiers**, not the target product model or branding.

## Why they are not renamed in one commit

Renaming all of them together can invalidate imports, mobile signing/bundle identity, stored sessions, Redis keys, deployed environment variables, database connections and existing automation. A cosmetic big-bang rename would create risk without improving marketplace correctness.

## Migration rule

1. New public product language uses **OpenRide**.
2. New V2 domain code uses marketplace-oriented names (`MobilityRequest`, `DriverTariff`, `Quote`, `Agreement`, `Ride`).
3. Legacy identifiers may remain behind compatibility boundaries while old runtime paths are still active.
4. A legacy identifier is removed only with a migration path, tests and deployment notes.
5. No new business-domain feature should introduce a fresh `FlashX` product dependency.

## What is not legacy

The old FlashX service verticals — designated driver and vehicle inspection assistance — may remain useful OpenRide service types in the future. What is being retired is the assumption that those verticals define the entire platform.

The OpenRide business kernel is the marketplace itself:

```text
MobilityRequest -> Quote -> Agreement -> Ride
```

This document exists so contributors can distinguish intentional compatibility debt from the target architecture.
