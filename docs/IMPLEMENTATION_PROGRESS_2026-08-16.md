# FlashX — Báo cáo tiến độ triển khai 16/08/2026

## Mục tiêu checkpoint

Đưa 3 dịch vụ MVP từ trạng thái demo nền tảng sang flow đủ chắc để UAT/demo Bộ Công Thương, đồng thời không khóa kiến trúc mở rộng marketplace về sau.

## Đã hoàn thành trên branch hiện tại

- 3 service semantics: lái hộ ô tô, lái hộ xe máy, đăng kiểm hộ.
- Inspection checklist end-to-end: template versioned, snapshot theo job, Rider khai báo, Driver đối chiếu, Admin theo dõi, gate trước nhận xe.
- Scheduled booking worker độc lập.
- Route/polyline/ETA Rider + Driver; Google Routes khi cấu hình, fallback khi provider ngoài không khả dụng.
- Tìm địa chỉ bằng chữ qua backend Google Places Text Search; shortcut và pin map vẫn là fallback.
- Rider incident/support dùng cùng incident state với Driver/Admin.
- Pickup grace period + server-enforced customer no-show + Driver countdown + Admin config.
- Rider rating UI + backend.
- Driver KYC onboarding: direct presigned upload, không proxy binary qua API.
- Driver qualification: capability do Admin cấp, GPLX class/expiry, quyền lái số sàn; dispatch lọc trước offer/accept/manual assign.
- Cash payment ledger/reconciliation hiển thị nhất quán Rider/Driver/Admin; không suy diễn payout/commission tài xế.
- Push foundation: device registration, token không trả ngược API, durable outbox, worker, event hooks. Provider disabled/development ghi `skipped`, không giả `sent`.
- Rider/Driver theme chuyển sang palette canonical trắng + xanh FlashX.

## Còn phụ thuộc quyết định/credential ngoài code

1. **Push thật FCM/APNs:** cần Firebase/APNs credential + mobile messaging SDK/config native. Backend foundation đã sẵn sàng.
2. **Waiting/night/holiday/cancellation fee:** cần Founder/Product chốt công thức và policy tiền trước khi code tự tính.
3. **Driver payout/commission/refund/invoice:** giữ tách khỏi customer cash ledger cho đến khi unit economics/policy được chốt.
4. **Pháp lý/bảo hiểm/ủy quyền đăng kiểm:** cần review chuyên môn trước Go-Live.
5. **Store/release polish:** cần UAT trên thiết bị thật và final visual/accessibility pass.

## Verification gần nhất

- `git diff --check` — PASS.
- `services/api: go test ./...` — PASS, bao gồm dispatch/manual transmission, payments, notifications, places và HTTP contracts.
- `apps/rider: flutter test` — PASS.
- `apps/driver: flutter test` — PASS.
- `apps/admin: npm run build` — PASS.

## Nguyên tắc release

- Chưa merge/deploy production từ branch này cho tới khi diff review + APK demo + smoke 3 dịch vụ hoàn tất.
- Railway chỉ build `dev`; không chạm production trong checkpoint này.
- Migrations mới phải chạy qua pre-deploy migration script hiện có.
- Push thật phải fail-closed: không có credential/provider thật thì không được báo `sent`.
- Search địa chỉ là tiện ích; pin map luôn giữ làm fallback để booking không phụ thuộc tuyệt đối vào Google Places.
