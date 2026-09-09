# Marketplace Service migrations

These migrations are owned exclusively by `marketplace-service`.

## Rules

1. Migrations are **append-only** once merged to `main`.
2. Never edit an applied migration to fix a production invariant; add the next migration.
3. `marketplace-service` is the only service allowed to evolve these tables.
4. Other services reference Marketplace identifiers through APIs/events, not SQL foreign keys into this database.
5. Every destructive or long-running migration must document rollout and rollback implications.
6. Application code must not assume a migration completed until service readiness/deployment orchestration guarantees it.

## Current migrations

- `001_marketplace.sql` — initial Marketplace-owned schema and Outbox/Inbox foundation.
- `002_integrity_guards.sql` — status, coordinate, currency, tariff-bound and Agreement identity integrity constraints.

## Local application

The development topology applies migrations in filename order through the `marketplace-migrate` job:

```bash
docker compose -f docker-compose.microservices.yml up -d --build --wait
```

To inspect the database after startup:

```bash
docker compose -f docker-compose.microservices.yml exec marketplace-db \
  psql -U marketplace -d marketplace
```

The migration files must remain portable PostgreSQL SQL. Provider-specific deployment wrappers belong outside this directory.
