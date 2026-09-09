# OpenRide First-Party Service Modules

`modules-go` contains mobility verticals that plug into `core-go` through the public extension contract.

Current modules:

- `passenger.car` — point-to-point passenger rides by car;
- `carpool.intercity` — seat-aware shared intercity mobility with its own ride lifecycle.

The separation is intentional:

```text
core-go            -> marketplace invariants and extension interfaces
modules-go         -> service-specific validation and lifecycle
services/api       -> operator runtime / adapters / HTTP / persistence
packages/contracts -> language-neutral schemas
packages/sdk       -> client SDK
```

A new service should normally be added as a module instead of adding service-specific fields or `if service == ...` branches to Core.

## Example manifest

```go
func (MyService) Manifest() extension.Manifest {
    return extension.Manifest{
        ID:          "mycompany.special-ride",
        Version:     "1.0.0",
        DisplayName: "Special Ride",
        Category:    "passenger",
        Contracts: map[string]string{
            "request_schema": "https://example.com/schemas/special-ride.v1.json",
        },
    }
}
```

## Validation

From the repository root:

```bash
make modules-test
```

The first-party catalog is available at `packages/modules-go/catalog` and is bridged into the current API runtime through `services/api/internal/openridecore`.
