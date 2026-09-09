# OpenRide Versioning Policy

OpenRide is a multi-module, pre-1.0 repository. Versioning must make it possible to consume packages and services without relying on a developer's local `go.work` file.

## Semantic versioning

Public packages and service contracts follow Semantic Versioning. Before `1.0.0`, breaking changes are allowed but must be explicit and documented.

## Go multi-module tags

Go modules inside this repository use path-prefixed tags.

Examples:

```text
packages/core-go/v0.1.0
packages/modules-go/v0.1.0
services/marketplace/v0.1.0
```

Do **not** publish one root `v0.1.0` tag and assume Go will resolve every nested module from it.

Until the first module tags are published, inter-module dependencies may use valid commit pseudo-versions. Placeholder dependencies such as `v0.0.0` are not acceptable on `main` because they only work accidentally through the repository workspace.

## JavaScript packages

JavaScript packages use normal package semantic versions:

```text
@openride/contracts
@openride/sdk
```

A package version bump must correspond to the wire/API compatibility impact of the change.

## Contract versioning

Wire contracts are versioned independently from package releases.

Examples:

```text
service-manifest.v1
marketplace-event.v1
request.passenger-car.v1
```

A breaking change creates a new contract major version. Existing meanings must not be silently changed under the same contract identifier.

## Event versioning

Event subjects/names include the major contract version:

```text
openride.marketplace.request.opened.v1
openride.marketplace.quote.created.v1
openride.marketplace.agreement.created.v1
```

Consumers must not infer compatibility from repository commit hashes.

## Release checklist

Before publishing a package/service version:

1. tests and vet/type checks pass;
2. the module/package is consumable without the root workspace;
3. public API changes are documented;
4. contract/event changes are versioned explicitly;
5. migrations are append-only and documented;
6. release notes describe upgrade and rollback considerations;
7. no credentials, local paths or private deployment assumptions are embedded.
