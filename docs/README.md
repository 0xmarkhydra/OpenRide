# FlashX — Project Documentation

Thư mục này là nguồn tài liệu chuẩn (source of truth) cho dự án FlashX.

## Mục lục

1. [PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md) — mục tiêu, phạm vi và định nghĩa sản phẩm.
2. [PRD_MVP.md](./PRD_MVP.md) — yêu cầu sản phẩm MVP chi tiết.
3. [ARCHITECTURE.md](./ARCHITECTURE.md) — kiến trúc hệ thống và nguyên tắc scale.
4. [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) — domain, aggregate và state machine.
5. [DATA_MODEL.md](./DATA_MODEL.md) — schema dữ liệu, PostGIS và Redis key model.
6. [API_CONTRACT.md](./API_CONTRACT.md) — quy ước REST API, auth, lỗi và endpoint MVP.
7. [REALTIME_LOCATION.md](./REALTIME_LOCATION.md) — WebSocket, GPS ingestion và fan-out.
8. [DISPATCH_ENGINE.md](./DISPATCH_ENGINE.md) — matching, scoring, locking và retry.
9. [MAPS_AND_GEO.md](./MAPS_AND_GEO.md) — Maps provider, route, geocoding và chi phí bên thứ ba.
10. [RIDER_APP.md](./RIDER_APP.md) — kiến trúc và luồng app khách hàng.
11. [DRIVER_APP.md](./DRIVER_APP.md) — kiến trúc, background location và luồng tài xế.
12. [ADMIN_PORTAL.md](./ADMIN_PORTAL.md) — nghiệp vụ Admin/Operations.
13. [SECURITY_PRIVACY.md](./SECURITY_PRIVACY.md) — auth, KYC, secrets, privacy và audit.
14. [INFRA_DEVOPS.md](./INFRA_DEVOPS.md) — môi trường, Docker, CI/CD và production topology.
15. [OBSERVABILITY.md](./OBSERVABILITY.md) — logs, metrics, traces, alerts và SLO.
16. [TESTING_QA.md](./TESTING_QA.md) — chiến lược test và tiêu chí UAT.
17. [DEPLOYMENT_RUNBOOK.md](./DEPLOYMENT_RUNBOOK.md) — quy trình release, rollback và incident cơ bản.
18. [ROADMAP_SCALING.md](./ROADMAP_SCALING.md) — roadmap MVP → scale, tiêu chí tách service/Rust.
19. [THIRD_PARTY_COSTS.md](./THIRD_PARTY_COSTS.md) — dịch vụ bên thứ ba và ownership chi phí.
20. [PROJECT_PLAN.md](./PROJECT_PLAN.md) — milestone, timeline, risk và Definition of Done.
21. [ADR.md](./ADR.md) — Architecture Decision Records.
22. [LEGAL_COMPLIANCE_VIETNAM.md](./LEGAL_COMPLIANCE_VIETNAM.md) — định hướng pháp lý, giấy phép và checklist Go-Live tại Việt Nam.

## Quy tắc cập nhật

- Mọi thay đổi lớn về nghiệp vụ hoặc kiến trúc phải cập nhật docs tương ứng trong cùng PR/commit.
- `ARCHITECTURE.md`, `DATA_MODEL.md`, `API_CONTRACT.md` và `ADR.md` phải khớp với implementation hiện tại.
- Không ghi API key, password, token hoặc secret thật vào docs.
- Khi một quyết định kiến trúc thay đổi, thêm ADR mới thay vì xóa lịch sử lý do.

## Trạng thái hiện tại

Dự án đang ở giai đoạn bootstrap MVP. Backend dùng Go modular monolith; PostgreSQL + PostGIS lưu dữ liệu bền vững; Redis giữ hot realtime/geo state; Rider/Driver dùng Flutter; Admin dùng Next.js.
