# OpenRide Packages

OpenRide packages are designed to be consumed independently from the full operator runtime.

```text
packages/
├── core-go/       # portable marketplace kernel and extension ports
├── modules-go/    # first-party mobility service modules
├── contracts/     # versioned language-neutral schemas
└── sdk/           # zero-dependency JS/TS Marketplace V2 client
```

## Dependency rule

Packages depend inward. They do not import the full `services/api` runtime.

```text
apps
  ↓
operator runtime / adapters
  ↓
core-go ← modules-go
   ↑          ↑
contracts ← sdk / other-language tooling
```

This lets communities adopt only the pieces they need:

- import Core and build a custom operator runtime;
- use first-party modules or register their own mobility vertical;
- consume contracts to generate Dart/Rust/Python/TypeScript integrations;
- use `@openride/sdk` without adopting OpenRide's web/mobile UI;
- use the full OpenRide runtime and replace selected ports;
- contribute a service module without forking the kernel.

## Local validation

```bash
make packages-test
```

That validates Core, service modules, contracts and SDK independently from the full applications.

See [`../docs/PACKAGE_ARCHITECTURE.md`](../docs/PACKAGE_ARCHITECTURE.md).
