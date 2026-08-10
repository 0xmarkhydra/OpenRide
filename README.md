# FlashX

Ride-hailing platform MVP inspired by the core operating model of Grab/Uber.

## Stack

- Rider app: Flutter
- Driver app: Flutter
- Admin: Next.js + TypeScript
- Backend: Go
- Primary database: PostgreSQL + PostGIS
- Realtime/cache/geo: Redis
- Realtime transport: WebSocket
- Maps: third-party provider such as Google Maps Platform
- Push: FCM/APNs
- Storage: S3-compatible object storage

## Repository layout

```text
flashx/
├── apps/
│   ├── rider/
│   ├── driver/
│   └── admin/
├── services/
│   └── api/
├── infrastructure/
│   └── migrations/
├── docs/
├── docker-compose.yml
└── Makefile
```

## Architecture rule

Start as a modular monolith in Go. Keep business domains isolated so Dispatch, Location, Payment and Notification can later be extracted into services without rewriting the whole product.

## Docker layout

Local Docker is intentionally kept as one Compose project named `flashx`.

```text
flashx
├── api
├── admin
├── postgres
└── redis
```

Backend test containers use the Compose `test` profile and always run with `--rm`, so they disappear immediately after the test finishes instead of accumulating in Docker Desktop.

## Local development

1. Copy `.env.example` to `.env`.
2. Start the full stack with `make stack-up`.
3. API health: `GET http://localhost:8080/health`.
4. Admin: `http://localhost:3000`.

Useful commands:

```bash
make stack-ps                 # show only FlashX Compose services
make stack-logs               # follow FlashX logs
make stack-down               # stop/remove FlashX containers, preserve DB volumes
make infra-up                 # only PostGIS + Redis
make api-run                  # run API directly on the host
make api-test                 # run Go tests directly on the host
make docker-test              # ephemeral backend test container
make docker-integration-test  # ephemeral integration test container
make stack-reset              # remove FlashX containers AND local DB/Redis volumes
```

Do not create manually named `flashx-test-*` containers for routine testing. Use the Make targets above so test containers are automatically removed.
