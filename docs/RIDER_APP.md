# Rider App — Legacy Compatibility Notes

> **Deprecated product semantics:** phần lớn tài liệu bên dưới mô tả hướng FlashX cũ và không còn là source of truth cho OpenRide Marketplace V2. Không implement UX mới từ các flow Lái hộ/Đăng kiểm hộ này. Thiết kế OpenRide hiện hành nằm ở [`UX_UI_SYSTEM.md`](./UX_UI_SYSTEM.md), [`OPENRIDE_UX_BA_ROADMAP.md`](./OPENRIDE_UX_BA_ROADMAP.md), và trạng thái thật ở [`PROJECT_STATUS.md`](./PROJECT_STATUS.md). File này chỉ còn hữu ích để tham khảo foundation Flutter/realtime/error handling trong giai đoạn migration.

## 1. Stack

Flutter, tách feature theo domain. Shared package chỉ chứa core/networking/design system/common models cần thiết.

## 2. Feature modules

- auth
- profile
- maps
- booking
- trip
- history
- rating
- promotions
- notifications

## 3. Main flow

```text
Launch -> Auth -> Home Map -> Pickup/Destination -> Estimate
-> Confirm Booking -> Searching -> Driver Assigned
-> Tracking -> In Trip -> Completed -> Rating
```

## 4. State management

Chọn một approach thống nhất (Riverpod/BLoC hoặc tương đương) khi implement. Không trộn nhiều pattern trong cùng app.

State quan trọng:
- auth session;
- current booking draft;
- active trip snapshot;
- realtime connection state;
- map/route state.

## 5. Networking

- REST cho snapshot/command.
- WebSocket cho realtime.
- HTTP client có timeout, retry có chọn lọc và request id.
- Không retry tự động command không idempotent.

## 6. Realtime recovery

Khi app resume/reconnect:
1. kiểm tra session;
2. gọi active trip snapshot;
3. reconnect WebSocket;
4. subscribe trip channel;
5. reconcile UI theo server state.

## 7. Map UX

- debounce place search;
- tránh route request dư thừa;
- marker pickup/destination rõ;
- map camera không giật khi location update;
- location permission có empty/error state tốt.

## 8. Error UX

Các nhóm lỗi phải có UX riêng:
- GPS disabled;
- permission denied;
- no internet;
- map provider unavailable;
- no driver found;
- offer/match cancelled;
- payment failed;
- session expired.

## 9. Security

- Token lưu bằng secure storage.
- Không log token/PII.
- Không nhúng server secret.
- Map mobile key phải restriction.

## 10. Analytics events

Tối thiểu:
- app_open
- estimate_requested/succeeded/failed
- booking_created
- driver_matched
- rider_cancelled
- trip_completed
- rating_submitted

Analytics không được chứa raw secret/KYC data.

## 11. Testing

- unit: pricing display/view models, validators;
- widget: booking screens/state;
- integration: happy path booking;
- device test: GPS/permission/background-resume/network switching.

## 12. Definition of Done

Rider feature chỉ Done khi có loading/error/empty states, analytics cần thiết, test tương ứng và hoạt động với API contract hiện tại.
