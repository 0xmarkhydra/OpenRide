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

## Local development

1. Copy `.env.example` to `.env`.
2. Start PostGIS and Redis with `docker compose up -d`.
3. Run the API with `make api-run`.
4. Health check: `GET http://localhost:8080/health`.
