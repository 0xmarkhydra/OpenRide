## What changed

<!-- Describe the user/business/architecture problem first, then the implementation. -->

## Boundary / ownership

- Owning package/service:
- Data owned or changed:
- API/event contracts changed:
- Other services affected:

## Marketplace principles

- [ ] Driver commercial terms are not silently rewritten.
- [ ] Rider choice / constraints remain explicit.
- [ ] Ranking/pricing behavior is explainable where applicable.
- [ ] No cross-service database reads were introduced.

## Reliability

- [ ] Critical commands are retry-safe/idempotent where applicable.
- [ ] Cross-service state changes use events/sagas rather than distributed DB transactions.
- [ ] Migration is append-only or rollback implications are documented.
- [ ] Failure/degraded behavior is covered.

## Validation

- [ ] Tests added/updated.
- [ ] Static analysis/type checks pass locally where applicable.
- [ ] Docker/service build checked if deployment files changed.
- [ ] Documentation/status matrix updated if implementation status changed.

## Compatibility / rollout

<!-- Explain migration, feature flag, backward compatibility and rollback. Write "none" only when genuinely not applicable. -->
