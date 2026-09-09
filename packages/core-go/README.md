# OpenRide Core for Go

`core-go` is the portable business kernel of OpenRide.

It is intentionally independent from HTTP, PostgreSQL, Redis, WebSocket, FCM/APNs, map providers and payment providers. Those belong in adapters or operator runtimes.

## Packages

```text
core-go/
├── money/        # minor-unit money value object
├── geo/          # provider-neutral coordinates
├── marketplace/  # Request, Tariff, Quote, Agreement, Ride + invariants
├── extension/    # service-module manifests and registry
└── engine/       # candidate -> quote -> rank orchestration ports
```

## Design rule

Core owns invariants. Extensions own vertical-specific behavior. Adapters own infrastructure.

```text
                         ┌─────────────────────────────┐
                         │       Operator Runtime       │
                         │ HTTP • DB • Redis • Realtime │
                         └──────────────┬──────────────┘
                                        │ adapters
                         ┌──────────────▼──────────────┐
                         │       OpenRide Core          │
                         │ request • quote • agreement  │
                         │ engine • state machines      │
                         └──────────────┬──────────────┘
                                        │ extension contract
                         ┌──────────────▼──────────────┐
                         │      Service Modules         │
                         │ passenger • carpool • parcel │
                         │ designated-driver • ...      │
                         └─────────────────────────────┘
```

## Run the example

From the repository root:

```bash
make core-example
```

Or directly:

```bash
cd packages/core-go
go run ./examples/minimal
```

The example registers a `passenger.car` service, discovers three drivers, creates driver-authorized offers and ranks them with an explainable custom algorithm.

## Import

Until the first tagged Core release, use the module directly from this repository:

```go
import (
    "github.com/0xmarkhydra/OpenRide/packages/core-go/engine"
    "github.com/0xmarkhydra/OpenRide/packages/core-go/extension"
    "github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)
```

## Stability

`core-go` is pre-1.0. Public APIs may evolve, but changes should follow these rules:

- domain names should be transport-provider neutral;
- infrastructure dependencies must not leak into Core;
- accepted commercial terms must be snapshot-safe;
- service-specific fields should be modeled through extension contracts instead of hardcoding every vertical into the kernel;
- new breaking APIs require a migration note;
- event names are versioned (`*.v1`, `*.v2`, ...).

See [`../../docs/PACKAGE_ARCHITECTURE.md`](../../docs/PACKAGE_ARCHITECTURE.md).
