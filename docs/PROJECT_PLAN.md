# Project Plan

## 1. Delivery model

MVP dự kiến triển khai theo milestone, ưu tiên tạo vertical slice chạy end-to-end sớm thay vì hoàn thiện từng app riêng rẽ rồi mới ghép cuối.

## 2. Timeline tham khảo

### Tuần 1–2 — Discovery & Foundation
- chốt flow và business rules;
- API/domain baseline;
- CI/CD cơ bản;
- DB/Redis/local env;
- auth skeleton;
- design/wireframe baseline.

### Tuần 3–5 — Rider/Driver onboarding + Maps
- Rider auth/profile;
- Driver auth/KYC/vehicle;
- map/location permission;
- place search/route/estimate;
- Admin driver review skeleton.

### Tuần 5–8 — Trip core + Dispatch
- Trip state machine;
- pricing;
- Redis GEO;
- offers;
- atomic assignment;
- cancellation;
- basic realtime.

### Tuần 8–10 — Active trip realtime
- background location;
- Rider tracking;
- arrived/start/complete;
- notification;
- trip history.

### Tuần 10–12 — Admin/Payment/Operations
- Admin trip/rider/driver;
- pricing rules;
- payment adapter nếu trong scope;
- promotions cơ bản;
- audit.

### Tuần 12–15 — QA/UAT/Hardening
- integration/E2E;
- concurrency tests;
- physical device tests;
- load test;
- security checks;
- observability;
- bug fixing.

### Tuần 16 — Launch
- production infra;
- migration;
- store submission/release;
- handover/runbook.

Timeline thay đổi nếu scope/third-party/legal/store review thay đổi.

## 3. Team tham khảo

- 1 PM/BA (có thể part-time theo giai đoạn).
- 1–2 Backend Go.
- 1–2 Flutter.
- 1 Frontend/Admin (có thể shared).
- 1 QA.
- UI/UX theo workload.
- DevOps shared/part-time giai đoạn đầu.

## 4. Milestone deliverables

### M1 Foundation
- repo structure;
- docs;
- local environment;
- auth/database skeleton.

### M2 Booking vertical slice
- Rider estimate/create trip;
- Driver appears online;
- test dispatch path.

### M3 End-to-end trip
- match;
- tracking;
- arrived/start/complete;
- history.

### M4 Operations
- Admin;
- KYC;
- pricing;
- notification/payment integration.

### M5 Release candidate
- QA/UAT;
- load/security;
- production readiness.

## 5. Definition of Done — Feature

Một feature chỉ Done khi:
- acceptance criteria pass;
- backend validation/security đầy đủ;
- unit/integration test phù hợp;
- docs/API updated nếu contract thay đổi;
- logs/metrics cần thiết;
- loading/error/retry state;
- QA sign-off;
- không chứa secret/debug artifact.

## 6. Definition of Done — MVP

- Rider/Driver/Admin core flow hoạt động.
- Dispatch không double-assign trong race tests.
- Background location đã test thiết bị thật.
- Trip transitions được audit/history.
- Production config/secrets riêng.
- Backup/monitoring/alerts baseline.
- Không Blocker/Critical bug core.
- Release/runbook/handover docs hoàn chỉnh.

## 7. Risks

### Maps/API cost
Mitigation: account của khách, quota/budget alert, debounce/cache phù hợp.

### Background GPS reliability
Mitigation: prototype/test iOS/Android sớm, adaptive interval, reconnect/snapshot.

### Dispatch race
Mitigation: atomic lock + DB conditional update + concurrency tests.

### Scope creep
Mitigation: PRD baseline + Change Request.

### Third-party delays
Mitigation: sandbox/mock provider trước, account production chuẩn bị sớm.

### Store review
Mitigation: submit sớm, compliance/privacy content chuẩn bị trước.

### Production scale uncertainty
Mitigation: load test + pilot traffic, không over-promise capacity.

## 8. Change management

Bất kỳ thay đổi ảnh hưởng:
- user flow;
- business rule;
- API contract;
- data model;
- third-party;
- timeline/cost

phải được ghi thành ticket/CR và cập nhật docs liên quan.

## 9. Handover

Bàn giao tối thiểu:
- source code;
- repository access;
- architecture/docs;
- env/config inventory không chứa secret trong docs;
- migration;
- deployment runbook;
- third-party account ownership list;
- known issues/backlog;
- production access transfer theo thỏa thuận.
