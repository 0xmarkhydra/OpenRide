# Product Backlog — FlashX MVP

## P0 — bắt buộc để bàn giao

### Foundation
- [ ] Environment local/staging/production.
- [ ] Config/secrets management.
- [ ] PostgreSQL/PostGIS migrations.
- [ ] Redis connection và health/readiness.
- [ ] Structured logging + request ID.
- [ ] Error model thống nhất.

### Auth & identity
- [ ] OTP abstraction + development provider.
- [ ] Rider session.
- [ ] Driver session.
- [ ] Admin auth + RBAC.
- [ ] Refresh/revoke session.

### Rider
- [ ] Home map + permission states.
- [ ] Pickup/destination selection.
- [ ] Estimate route/fare.
- [ ] Create trip idempotently.
- [ ] Searching state.
- [ ] Driver assigned state.
- [ ] Realtime tracking.
- [ ] Cancel trip.
- [ ] Completed summary.
- [ ] Rating.
- [ ] Trip history.

### Driver
- [ ] Profile + KYC status.
- [ ] Vehicle.
- [ ] Admin approval gate.
- [ ] Online/offline.
- [ ] Foreground/background location policy.
- [ ] Trip offer with expiry.
- [ ] Atomic accept/reject.
- [ ] Arrived/start/complete.
- [ ] Navigation handoff.
- [ ] Trip history/earning summary.

### Dispatch
- [ ] Redis GEO candidate discovery.
- [ ] Driver freshness/availability filter.
- [ ] Deterministic scoring.
- [ ] Offer TTL/retry.
- [ ] Atomic assignment.
- [ ] Double-assignment protection.
- [ ] No-driver-found outcome.

### Pricing
- [ ] Service type.
- [ ] Base fare + distance fare.
- [ ] Versioned pricing rules.
- [ ] Fare estimate breakdown.
- [ ] Final fare server-side.

### Admin
- [ ] Dashboard.
- [ ] Rider list/detail.
- [ ] Driver list/detail/KYC decision.
- [ ] Vehicle view.
- [ ] Trip list/detail/timeline.
- [ ] Pricing configuration.
- [ ] Promotion basic.
- [ ] Suspend/unsuspend.
- [ ] Audit log.

### Notifications
- [ ] Push abstraction.
- [ ] Driver offer notification/fallback.
- [ ] Driver assigned/arrived/completed notifications.

### Maps
- [ ] Provider abstraction boundary.
- [ ] Place search.
- [ ] Route distance/duration.
- [ ] Rider map rendering.
- [ ] Driver external navigation handoff.

### QA/Release
- [ ] Unit tests core state/pricing/dispatch.
- [ ] API integration tests.
- [ ] Race test assignment.
- [ ] Mobile booking happy path.
- [ ] Driver background/network recovery test.
- [ ] Admin permission matrix.
- [ ] Staging deployment.
- [ ] Production runbook + rollback.

## P1 — nên có nếu không đe dọa release
- [ ] Promo code apply.
- [ ] Support notes.
- [ ] Live operations map sampled.
- [ ] Better driver earning breakdown.
- [ ] Basic cancellation reason analytics.
- [ ] Payment gateway online nếu khách có merchant account đúng hạn.

## P2 — MVP+
- [ ] Scheduled booking.
- [ ] Multi-stop.
- [ ] Wallet/ledger.
- [ ] Referral.
- [ ] Loyalty.
- [ ] Surge pricing.
- [ ] Embedded turn-by-turn navigation.
- [ ] Call masking.
- [ ] Advanced fraud detection.
- [ ] Food/delivery verticals.

## Delivery order

Không ưu tiên theo danh sách module. Thứ tự triển khai thực tế:
1. Foundation + Trip domain.
2. Create/get/cancel trip API.
3. Driver availability/location.
4. Dispatch/offer/accept.
5. Trip execution state machine.
6. Rider/Driver UI nối API.
7. Admin operations.
8. KYC/pricing/notifications/maps providers.
9. Hardening/UAT/deploy.
