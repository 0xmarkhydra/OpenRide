# OpenRide Packages

OpenRide packages are designed to be consumed independently from the full operator runtime.

```text
packages/
├── core-go/     # portable marketplace kernel and extension ports
└── contracts/   # versioned language-neutral schemas (npm-ready)
```

## Dependency rule

Packages may depend inward on stable domain contracts, but they must not depend on the full `services/api` runtime.

```text
applications -> operator runtime -> packages/core-go
SDKs/modules ---------------------> packages/contracts
```

This lets communities adopt only the pieces they need:

- import Core and build a custom operator runtime;
- consume only contracts to generate another-language SDK;
- use the full OpenRide runtime and replace selected ports;
- contribute a service module without forking the kernel.

See [`../docs/PACKAGE_ARCHITECTURE.md`](../docs/PACKAGE_ARCHITECTURE.md).
