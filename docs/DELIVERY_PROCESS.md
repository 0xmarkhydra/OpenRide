# Quy trình phát triển và bàn giao FlashX MVP

## 1. Nguyên tắc điều hành

Dự án được vận hành như một product team thống nhất thay vì các nhóm tách rời. Mỗi quyết định phải ưu tiên khả năng bàn giao một MVP chạy thật, dễ kiểm thử, dễ vận hành và có đường mở rộng.

Thứ tự ưu tiên khi có xung đột:
1. Correctness của luồng đặt/chạy chuyến.
2. An toàn dữ liệu và tiền.
3. Khả năng vận hành/debug.
4. UX rõ ràng trong trạng thái lỗi/mạng yếu.
5. Tốc độ phát triển.
6. Tối ưu hiệu năng nâng cao.

## 2. Mô hình team

Một backlog chung cho Product/BA, UX/UI, Backend, Rider, Driver, Admin, QA và DevOps. Mọi feature có một owner logic nhưng Definition of Done luôn end-to-end.

Vai trò logic:
- Product/BA: scope, business rules, acceptance criteria.
- UX/UI: flow, state, design system, accessibility.
- Backend: domain, API, persistence, realtime, dispatch.
- Mobile: Rider/Driver production behavior.
- Web: Admin operations.
- QA: test strategy, regression, release gate.
- DevOps: environments, deployment, observability, rollback.

## 3. Delivery model

Phát triển theo vertical slice:

1. Foundation: project rules, health, config, database, logging.
2. Rider creates trip.
3. Driver online + location.
4. Dispatch + offer + atomic accept.
5. Rider tracks assigned driver.
6. Arrived -> Start -> Complete.
7. Admin observes and manages data.
8. KYC, notification, pricing, payment integration.
9. Hardening, UAT, release.

Không triển khai một module lớn hoàn toàn riêng biệt nếu chưa tạo giá trị end-to-end.

## 4. Sprint flow

Mỗi sprint:
1. BA chốt rule + acceptance criteria.
2. UX định nghĩa happy/error/loading/empty/offline states.
3. API/domain contract được cập nhật.
4. Backend + client triển khai cùng contract.
5. Unit/integration/E2E test.
6. Telemetry/logging.
7. Demo trên staging.
8. Docs cập nhật cùng code.

## 5. Definition of Ready

Một story chỉ Ready khi có:
- user/business outcome;
- actor;
- precondition;
- business rules;
- happy path;
- error/edge cases;
- acceptance criteria;
- API/data impact;
- security/privacy impact nếu có.

## 6. Definition of Done

Một story chỉ Done khi:
- code reviewable và domain rule không nằm rải rác trong UI;
- happy path hoạt động end-to-end;
- error/loading/empty/offline state phù hợp;
- auth/permission được enforce ở backend;
- logs/metrics cần thiết tồn tại;
- test tự động cho logic quan trọng;
- migration backward-safe nếu thay DB;
- docs liên quan cập nhật;
- không có blocker/critical bug.

## 7. Quy tắc scope

Nếu một yêu cầu mới có nguy cơ làm trễ MVP, ưu tiên:
1. Giữ core ride flow.
2. Chuyển yêu cầu sang MVP+.
3. Chỉ mở scope nếu nó là điều kiện pháp lý/vận hành bắt buộc.

## 8. Release gate

Không release production nếu còn một trong các vấn đề:
- có thể assign một driver cho hai active trip;
- trip state có thể đi sai thứ tự;
- rider/driver thấy dữ liệu của user khác;
- payment webhook không idempotent;
- KYC data lộ vượt quyền;
- background GPS không có hành vi xác định;
- không có rollback cho backend/migration nguy hiểm;
- blocker/critical bug chưa xử lý.

## 9. Handover

Bàn giao cuối gồm source code, migrations, environment template, CI/CD config, API docs, architecture docs, runbook, test report/UAT, release artifacts và danh sách tài khoản third-party do khách hàng sở hữu.
