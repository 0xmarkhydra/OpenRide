# Project Overview

## 1. Tầm nhìn

FlashX là nền tảng ride-hailing theo yêu cầu, cho phép khách hàng đặt chuyến và kết nối với tài xế gần nhất/phù hợp nhất theo thời gian thực. MVP tập trung vào một khu vực vận hành ban đầu, với kiến trúc đủ sạch để mở rộng sang nhiều thành phố và nhiều loại dịch vụ.

## 2. Thành phần sản phẩm

- Rider App: ứng dụng khách hàng trên iOS/Android.
- Driver App: ứng dụng tài xế trên iOS/Android.
- Admin Portal: cổng điều hành trên web.
- Go Backend: API, business logic, trip lifecycle, dispatch, pricing, payment orchestration và realtime.
- PostgreSQL + PostGIS: dữ liệu bền vững và spatial query.
- Redis: tài xế online, vị trí mới nhất, geo index, lock và cache.
- Third-party: Maps, SMS/OTP, Push, Payment Gateway, Object Storage.

## 3. Mục tiêu MVP

1. Khách chọn điểm đón/đến và nhận giá dự kiến.
2. Hệ thống tìm tài xế phù hợp trong vùng.
3. Tài xế nhận/từ chối chuyến.
4. Khách theo dõi tài xế realtime.
5. Tài xế bắt đầu/kết thúc chuyến.
6. Hệ thống lưu lịch sử, giá cuối, thanh toán và đánh giá.
7. Admin quản lý người dùng, tài xế, chuyến và giá.

## 4. Không phải mục tiêu Phase 1

- Không tự xây bản đồ hoặc routing engine.
- Không triển khai microservices hàng loạt.
- Không Kubernetes nếu chưa có nhu cầu scale thực tế.
- Không AI dispatch ở MVP.
- Không ví điện tử nội bộ phức tạp.
- Không tối ưu cho hàng triệu chuyến/ngày từ ngày đầu.

## 5. Stack chuẩn

| Layer | Technology |
|---|---|
| Rider | Flutter |
| Driver | Flutter |
| Admin | Next.js + TypeScript |
| Backend | Go |
| Database | PostgreSQL 16 + PostGIS |
| Hot state / Geo | Redis |
| Realtime | WebSocket |
| Maps | Google Maps Platform hoặc provider tương đương |
| Push | FCM / APNs |
| Storage | S3-compatible object storage |
| Runtime | Docker |

## 6. Nguyên tắc kiến trúc

- Modular monolith trước, service extraction sau khi có dữ liệu tải thật.
- Dữ liệu durable nằm trong PostgreSQL; state realtime ngắn hạn nằm trong Redis.
- Trip lifecycle phải được quản lý bằng state machine rõ ràng.
- Dispatch phải có lock/idempotency để chống double assignment.
- Tất cả third-party provider đi qua adapter/interface, không trộn trực tiếp vào domain logic.
- API và event phải versioned khi có client production.

## 7. Các KPI kỹ thuật cần theo dõi

- Trip creation success rate.
- Time-to-driver-match.
- Driver location freshness.
- WebSocket connection success/reconnect rate.
- Dispatch acceptance rate.
- Trip completion rate.
- API p95 latency.
- Error rate theo endpoint/service.
- Redis/Postgres saturation.

## 8. Quy mô kỳ vọng Phase 1

MVP được thiết kế để chạy ổn cho giai đoạn kiểm chứng thị trường. Capacity cụ thể phải được benchmark bằng load test trước launch; không cam kết một con số traffic cố định nếu chưa có cấu hình cloud production và profile tải thực tế.
