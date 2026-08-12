# Driver App

> **Current product direction:** Driver App phục vụ tài xế/người thực hiện cho **Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ**. Không coi xe của tài xế là phương tiện chở khách như Grab. KYC tập trung vào năng lực tài xế; workflow nhận/bàn giao **xe của khách** phải được bổ sung theo [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md).

## 1. Stack

Flutter. Driver App có yêu cầu background location và realtime cao hơn Rider nên phải test trên thiết bị thật sớm.

## 2. Feature modules

- auth
- onboarding/KYC
- vehicle
- availability
- dispatch offers
- active trip
- navigation handoff
- earnings
- history
- notifications

## 3. Main flow

```text
Launch -> Auth -> KYC status -> Approved
-> Online -> Location streaming
-> Offer -> Accept/Reject
-> Navigate pickup -> Arrived -> Start
-> Navigate destination -> Complete
-> Earnings/History
```

## 4. Availability

Backend là authority. UI chỉ hiển thị trạng thái server xác nhận.

Driver không được Online nếu:
- chưa approved;
- account suspended;
- không có active vehicle hợp lệ;
- app không có location permission cần thiết theo policy.

## 5. Background location

Phải xử lý riêng iOS/Android:
- foreground/background permission;
- battery optimization;
- app lifecycle;
- OS kill/restart;
- GPS off;
- network loss;
- location accuracy.

Tần suất update adaptive theo trạng thái như mô tả trong `REALTIME_LOCATION.md`.

## 6. Offer handling

Offer hiển thị:
- pickup summary;
- khoảng cách/ETA tới pickup nếu policy cho phép;
- service type;
- thông tin giá cần thiết;
- countdown.

Accept phải gọi backend atomic endpoint; UI không coi accept thành công trước khi server xác nhận.

## 7. Trip commands

Các action:
- arrived
- start
- complete
- cancel theo policy

Mỗi action phải disable double tap và hỗ trợ idempotency/server conflict handling.

## 8. Navigation

MVP ưu tiên deep-link/open external navigation app. Embedded navigation là phase riêng.

## 9. Offline/network recovery

Nếu mất mạng:
- giữ UI state gần nhất với cảnh báo rõ;
- queue location ngắn hạn nếu hợp lý, không replay location quá cũ như realtime;
- trip commands cần retry có kiểm soát/idempotency;
- sau reconnect fetch server snapshot.

## 10. Security/KYC

- Upload KYC bắt buộc direct-to-object-storage bằng presigned URL; backend chỉ cấp chữ ký và lưu metadata, không nhận file binary/multipart.
- Không cache ảnh KYC lâu hơn cần thiết.
- Không log PII/document identifiers nhạy cảm.

## 11. Telemetry

Theo dõi:
- online session duration;
- location update success/failure;
- stale GPS rate;
- offers received/accepted/rejected/expired;
- accept latency;
- trip command failures;
- reconnects.

## 12. Testing bắt buộc

- physical Android/iPhone;
- screen off/background;
- network Wi-Fi -> 4G/5G -> offline -> reconnect;
- GPS disabled/enabled;
- permission denied then granted;
- app killed/resumed;
- offer race/expiry;
- long trip/battery test.
