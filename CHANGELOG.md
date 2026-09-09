# Changelog

All notable changes to OpenRide will be documented here.

The project is pre-1.0. Until the first stable release, entries may include architecture migrations and compatibility notes that would normally be reserved for major versions.

## Unreleased

### Added

- Marketplace V2 service boundary and microservices topology.
- Portable Go marketplace core and first-party service modules.
- Versioned language-neutral contracts.
- JavaScript/TypeScript SDK foundation.
- Transactional outbox/inbox persistence foundation.
- OpenRide project status, versioning, migration, release and legacy-compatibility documentation.
- Maintainer CODEOWNERS, Dependabot, issue templates and pull-request template.
- CI checks for standalone Go module resolution, Docker image build and microservices health/readiness.

### Changed

- Public project identity is OpenRide.
- Legacy `services/api` is explicitly treated as a compatibility runtime during staged extraction.
- Admin/operator package and public configuration naming are migrating from FlashX to OpenRide while preserving compatibility fallbacks where required.
- Contribution guidance now matches the bounded-context microservices architecture.

### Fixed

- AGPL license file now contains the complete standard license text with project attribution separated into `NOTICE`.
- Marketplace database integrity constraints are extended through additive migrations rather than editing shipped migration history.
- Public documentation now distinguishes implemented, foundation, compatibility and planned capabilities to avoid overstating project readiness.

### Compatibility

Some internal identifiers still contain `flashx` because changing them in place could break deployed databases, Redis keys, mobile application identifiers, sessions or V1 API clients. See `docs/LEGACY_COMPATIBILITY.md` for the migration policy.
