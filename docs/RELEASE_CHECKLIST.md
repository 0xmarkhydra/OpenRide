# OpenRide maintainer release checklist

This checklist is intentionally short and operational. It exists to prevent a polished README from getting ahead of what the repository can actually build and run.

## Pull request gate

- [ ] `go test ./...` and `go vet ./...` pass for changed Go modules.
- [ ] Marketplace resolves with `GOWORK=off`.
- [ ] Contract validation passes.
- [ ] SDK tests and `npm pack --dry-run` pass.
- [ ] Admin builds with `npm ci && npm run build`.
- [ ] Rider and Driver apps analyze/test when touched.
- [ ] Docker Compose files validate.
- [ ] Marketplace Docker image builds and the microservices slice reaches `/healthz` and `/readyz`.
- [ ] New DB changes are append-only migrations.
- [ ] Public docs distinguish implemented, foundation, compatibility and planned capabilities.
- [ ] No credentials, production tokens, KYC data or private user data are included.

## Before tagging

- [ ] Update `CHANGELOG.md`.
- [ ] Confirm all package/module versions follow `docs/VERSIONING.md`.
- [ ] Confirm published package manifests contain only intended files.
- [ ] Confirm `LICENSE`, `NOTICE`, `SECURITY.md` and contribution documents are current.
- [ ] Confirm the GitHub repository description/topics describe OpenRide, not legacy FlashX branding.
- [ ] Confirm the release commit is green in CI.

## Production-readiness rule

Pre-1.0 releases must not be described as production-ready for real passenger transport unless safety, privacy, local regulation, insurance, payments, KYC, incident response, fraud controls, monitoring and operational recovery have been validated for the target operator and jurisdiction.
