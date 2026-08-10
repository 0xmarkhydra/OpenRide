# PRD — FlashX MVP

## 1. Product objective

Xây phiên bản thương mại đầu tiên cho mô hình gọi xe theo yêu cầu, ưu tiên luồng đặt chuyến end-to-end và khả năng vận hành thực tế.

## 2. Personas

### Rider
- Muốn đặt xe nhanh, biết giá trước, theo dõi tài xế và xem lịch sử chuyến.

### Driver
- Muốn đăng ký, được duyệt, bật online, nhận chuyến, điều hướng và theo dõi thu nhập.

### Operations/Admin
- Muốn duyệt tài xế, theo dõi chuyến, xử lý bất thường và cấu hình giá.

## 3. Rider requirements

### Account
- Đăng ký/đăng nhập bằng số điện thoại.
- OTP qua provider bên thứ ba.
- Hồ sơ cơ bản.

### Booking
- Lấy vị trí hiện tại.
- Tìm/chọn pickup và destination.
- Lấy route, distance và ETA từ map provider.
- Hiển thị service type và estimated fare.
- Tạo booking.
- Hủy chuyến theo policy.

### Tracking
- Nhận thông tin driver sau khi match.
- Xem vị trí driver realtime.
- Xem trạng thái chuyến.

### After trip
- Hiển thị final fare.
- Ghi nhận phương thức/trạng thái thanh toán.
- Rating và feedback.
- Trip history.

## 4. Driver requirements

### Onboarding/KYC
- Tạo tài khoản.
- Nhập thông tin cá nhân.
- Thêm phương tiện.
- Upload tài liệu KYC.
- Admin approve/reject/request-more-info.

### Availability
- Online/Offline.
- Chỉ tài xế approved mới được Online.
- Khi Online, app gửi location theo policy.

### Trip execution
- Nhận trip offer.
- Accept/Reject trước timeout.
- Navigate tới pickup.
- Arrived → Start trip → Complete trip.
- Xem trip history và earning summary cơ bản.

## 5. Admin requirements

- Dashboard tổng quan.
- Quản lý Riders.
- Quản lý Drivers/KYC/Vehicles.
- Quản lý Trips.
- Cấu hình service types và pricing rules.
- Promotion cơ bản.
- Khóa/mở tài khoản.
- Audit các thao tác quan trọng.

## 6. Trip state machine

```text
SEARCHING
  -> ACCEPTED
  -> ARRIVING
  -> ARRIVED
  -> IN_PROGRESS
  -> COMPLETED

SEARCHING/ACCEPTED/ARRIVING/ARRIVED -> CANCELLED (theo policy)
```

Mọi transition phải được backend validate; mobile client không được tự ý set state tùy ý.

## 7. Pricing MVP

Fare estimate có thể gồm:
- base fare;
- distance component;
- duration component nếu áp dụng;
- fixed surcharge;
- promotion/discount.

Pricing config phải nằm ở backend/Admin, không hard-code trong mobile.

## 8. Dispatch MVP

- Tìm candidate theo Redis GEO.
- Lọc theo service type, approved/online/available.
- Score theo khoảng cách + idle time + quality signals cơ bản.
- Lock assignment để tránh một driver nhận hai trip.
- Timeout/retry sang candidate tiếp theo.

## 9. Non-functional requirements

- HTTPS/TLS production.
- Auth token và RBAC Admin.
- Idempotency cho các action quan trọng.
- Structured logs.
- Error tracking.
- Backup DB production.
- Config/secrets tách khỏi source code.
- Graceful handling cho mạng yếu/reconnect.

## 10. Acceptance criteria MVP

MVP đạt yêu cầu khi:
1. Rider tạo chuyến thành công.
2. Driver phù hợp nhận được offer.
3. Một và chỉ một driver được assign.
4. Rider thấy location/status driver realtime.
5. Driver hoàn thành trip.
6. Trip history lưu đủ các mốc chính.
7. Admin có thể tra cứu và quản lý trip/driver/rider.
8. Không còn blocker/critical bug trên luồng core.

## 11. Out of scope mặc định

- Wallet/ledger phức tạp.
- Dynamic surge pricing theo ML.
- Multi-stop.
- Scheduled booking.
- Food delivery.
- AI fraud detection.
- In-app turn-by-turn navigation tự xây.
- Call masking/VoIP riêng.
