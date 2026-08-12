# FlashX — Project Documentation

Thư mục này là nguồn tài liệu chuẩn (source of truth) cho dự án FlashX.

## Đọc theo thứ tự này trước khi triển khai

1. [CUSTOMER_REQUIREMENTS.md](./CUSTOMER_REQUIREMENTS.md) — **source of truth nghiệp vụ mới nhất**: FlashX là nền tảng lái hộ/hỗ trợ phương tiện; MVP gồm Lái hộ ô tô, Lái hộ xe máy, Đăng kiểm hộ.
2. [THREAD_SUMMARY_2026-08-12.md](./THREAD_SUMMARY_2026-08-12.md) — toàn bộ context/pivot/quyết định của thread hiện tại để tiếp tục ở cuộc hội thoại khác.
3. [PRD_MVP.md](./PRD_MVP.md) — yêu cầu sản phẩm MVP chi tiết.
4. [PRODUCT_BACKLOG.md](./PRODUCT_BACKLOG.md) — backlog P0/P1/P2 đã rebaseline theo business lái hộ/đăng kiểm hộ.
5. [UX_UI_SYSTEM.md](./UX_UI_SYSTEM.md) — design system và flow UX mới cho Customer/Driver/Admin.
6. [FULL_MARKETPLACE_COST_PLAN.md](./FULL_MARKETPLACE_COST_PLAN.md) — phương án tài chính/triển khai Founder đã chốt: UI đơn giản, engine Full Marketplace.
7. [IMPLEMENTATION_GAP_REVIEW_2026-08-12.md](./IMPLEMENTATION_GAP_REVIEW_2026-08-12.md) — đối chiếu code hiện tại với yêu cầu Full Marketplace và danh sách blocker P0 cần xử lý.

Nếu tài liệu legacy còn mô tả `car/bike` theo nghĩa taxi/ride-hailing, **7 tài liệu trên được ưu tiên** và tài liệu legacy phải được migrate trước khi dùng để code feature mới.

## Mục lục

- [PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md) — mục tiêu, phạm vi và định nghĩa sản phẩm.
- [PRD_MVP.md](./PRD_MVP.md) — yêu cầu sản phẩm MVP chi tiết.
- [ARCHITECTURE.md](./ARCHITECTURE.md) — kiến trúc hệ thống và nguyên tắc scale.
- [DOMAIN_MODEL.md](./DOMAIN_MODEL.md) — domain, aggregate và state machine.
- [DATA_MODEL.md](./DATA_MODEL.md) — schema dữ liệu, PostGIS và Redis key model.
- [API_CONTRACT.md](./API_CONTRACT.md) — quy ước REST API, auth, lỗi và endpoint MVP.
- [REALTIME_LOCATION.md](./REALTIME_LOCATION.md) — WebSocket, GPS ingestion và fan-out.
- [DISPATCH_ENGINE.md](./DISPATCH_ENGINE.md) — matching, scoring, locking và retry.
- [MAPS_AND_GEO.md](./MAPS_AND_GEO.md) — Maps provider, route, geocoding và chi phí bên thứ ba.
- [RIDER_APP.md](./RIDER_APP.md) — kiến trúc và luồng app khách hàng.
- [DRIVER_APP.md](./DRIVER_APP.md) — kiến trúc, background location và luồng tài xế.
- [ADMIN_PORTAL.md](./ADMIN_PORTAL.md) — nghiệp vụ Admin/Operations.
- [SECURITY_PRIVACY.md](./SECURITY_PRIVACY.md) — auth, KYC, secrets, privacy và audit.
- [INFRA_DEVOPS.md](./INFRA_DEVOPS.md) — môi trường, Docker, CI/CD và production topology.
- [OBSERVABILITY.md](./OBSERVABILITY.md) — logs, metrics, traces, alerts và SLO.
- [TESTING_QA.md](./TESTING_QA.md) — chiến lược test và tiêu chí UAT.
- [DEPLOYMENT_RUNBOOK.md](./DEPLOYMENT_RUNBOOK.md) — quy trình release, rollback và incident cơ bản.
- [ROADMAP_SCALING.md](./ROADMAP_SCALING.md) — roadmap MVP → scale, tiêu chí tách service/Rust.
- [THIRD_PARTY_COSTS.md](./THIRD_PARTY_COSTS.md) — dịch vụ bên thứ ba và ownership chi phí.
- [INITIAL_COMPANY_FINANCIAL_PLAN.md](./INITIAL_COMPANY_FINANCIAL_PLAN.md) — bức tranh tài chính ban đầu: chi phí MVP, thành lập doanh nghiệp, setup Go-Live, burn rate và runway 3/6/9/12 tháng.
- [PROJECT_PLAN.md](./PROJECT_PLAN.md) — milestone, timeline, risk và Definition of Done.
- [ADR.md](./ADR.md) — Architecture Decision Records.
- [LEGAL_COMPLIANCE_VIETNAM.md](./LEGAL_COMPLIANCE_VIETNAM.md) — định hướng pháp lý, giấy phép và checklist Go-Live tại Việt Nam; cần review lại theo business lái hộ/đăng kiểm hộ trước production.
- [OBJECT_STORAGE.md](./OBJECT_STORAGE.md) — S3-compatible direct upload, presigned URL và KYC media flow.

## Quy tắc cập nhật

- Mọi thay đổi lớn về nghiệp vụ hoặc kiến trúc phải cập nhật docs tương ứng trong cùng PR/commit.
- `ARCHITECTURE.md`, `DATA_MODEL.md`, `API_CONTRACT.md` và `ADR.md` phải khớp với implementation hiện tại.
- Không ghi API key, password, token hoặc secret thật vào docs.
- Khi một quyết định kiến trúc thay đổi, thêm ADR mới thay vì xóa lịch sử lý do.

## Trạng thái hiện tại

Dự án đang ở giai đoạn **business-domain realignment + MVP hardening**. Foundation kỹ thuật đã có đáng kể: Go modular monolith; PostgreSQL + PostGIS; Redis realtime/geo/locks; WebSocket; Flutter Customer/Driver; Next.js Admin; Docker/CI; S3-compatible presigned direct upload. Từ 12/08/2026, mọi feature mới phải bám phạm vi **Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ** trong `CUSTOMER_REQUIREMENTS.md`, không tiếp tục mở rộng theo giả định Grab/Uber clone cũ.
