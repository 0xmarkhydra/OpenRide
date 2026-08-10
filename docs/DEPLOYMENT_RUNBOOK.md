# Deployment Runbook

## 1. Mục tiêu

Quy trình release phải reproducible, rollback được và không làm mất dữ liệu. Tài liệu này mô tả baseline; khi chọn cloud cụ thể cần bổ sung command/provider detail.

## 2. Preconditions

Trước production deploy:
- CI xanh;
- migration reviewed;
- secrets/config production sẵn sàng;
- backup gần nhất thành công;
- third-party credentials/quota hợp lệ;
- release notes có thay đổi/risk;
- mobile API compatibility đã kiểm tra.

## 3. Backend release flow

1. Build immutable image từ commit/tag.
2. Run unit/integration/security checks.
3. Deploy Staging.
4. Run smoke + UAT relevant.
5. Verify migration dry-run/compatibility.
6. Approve Production.
7. Apply backward-compatible migration nếu có.
8. Rollout API/worker gradually nếu platform hỗ trợ.
9. Run production smoke.
10. Theo dõi metrics/logs tối thiểu trong release window.

## 4. Smoke checks

- `/health` OK.
- `/ready` OK.
- DB/Redis connection healthy.
- login/OTP test account flow.
- trip estimate.
- create/cancel test trip ở safe environment/account.
- WebSocket connect.
- Admin login/read.
- provider health cơ bản.

## 5. Database migration

Ưu tiên expand-and-contract:
1. Add new columns/tables/index compatible.
2. Deploy code có thể chạy cả old/new shape.
3. Backfill nếu cần.
4. Switch reads/writes.
5. Enforce constraint/remove old field ở release sau.

Tránh destructive schema change cùng release với app code nếu chưa có rollback path.

## 6. Rollback

Rollback app bằng previous immutable image/tag.

Không rollback database migration một cách mù quáng. Nếu migration đã ghi dữ liệu theo schema mới, phải dùng forward-fix hoặc migration rollback đã được thiết kế/test.

## 7. Feature flags

Feature có risk cao nên có server-side flag/config khi có thể:
- new dispatch strategy;
- new pricing rule;
- new payment provider;
- experimental realtime feature.

Không dùng feature flag để giữ dead code vô thời hạn.

## 8. Mobile release

- build signed Rider/Driver riêng;
- staging/internal testing trước;
- version code/build number tăng;
- release notes;
- verify production API base URL/maps keys/push config;
- phased rollout nếu store hỗ trợ.

Backend phải tương thích ít nhất với app version còn active theo policy.

## 9. Admin release

Admin có thể deploy nhanh hơn mobile nhưng breaking API vẫn phải phối hợp backend.

## 10. Incident during deploy

Nếu error/latency tăng vượt threshold:
1. stop rollout;
2. xác định app vs dependency vs migration;
3. rollback app nếu an toàn;
4. disable feature flag nếu liên quan;
5. communicate incident status;
6. preserve logs/metrics;
7. postmortem sau khi ổn định.

## 11. Release notes template

```text
Version:
Commit/tag:
Date:
Owner:
Changes:
Migrations:
Feature flags:
Third-party changes:
Known risks:
Rollback target:
Verification results:
```

## 12. Post-deploy checks

- 5xx/error rate;
- p95 latency;
- DB/Redis saturation;
- dispatch success/time-to-match;
- location freshness;
- payment/provider errors;
- mobile crash spike.

## 13. Production access

Production write/admin access theo least privilege. Không dùng shared root credential cho cả team. Mọi thao tác manual nhạy cảm cần audit khi hạ tầng cho phép.
