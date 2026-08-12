# Project Overview — FlashX

> **Cập nhật nghiệp vụ: 12/08/2026**  
> Source of truth chi tiết: [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md)

## 1. Tầm nhìn

FlashX là nền tảng **tài xế lái hộ và hỗ trợ phương tiện theo yêu cầu**, kết nối khách hàng có phương tiện riêng với tài xế/đối tác đã được xác minh.

FlashX **không còn được định nghĩa là Grab/Uber clone**. Điểm khác biệt cốt lõi: trong dịch vụ lái hộ, tài xế FlashX tới vị trí khách và **lái chính phương tiện của khách**.

MVP tập trung pilot tại một khu vực/thành phố trước khi mở rộng.

## 2. Phạm vi MVP đã chốt

FlashX MVP chỉ tập trung 3 dịch vụ:

1. **Lái hộ ô tô** (`designated_driver_car`).
2. **Lái hộ xe máy** (`designated_driver_bike`).
3. **Đăng kiểm hộ** (`vehicle_inspection_assist`).

Các vertical như taxi fleet, food delivery, giao hàng, thuê xe tự lái không thuộc MVP mặc định.

## 3. Thành phần sản phẩm

- **Customer/Rider App:** khách hàng đặt dịch vụ, quản lý xe của mình, theo dõi job/tài xế, thanh toán và lịch sử.
- **Driver App:** tài xế/đối tác KYC, online/offline, nhận offer, tới khách, nhận/bàn giao xe và thực hiện dịch vụ.
- **Admin / Operations Portal:** điều hành job, KYC, pricing, support, audit và theo dõi realtime.
- **Go Backend:** API, auth, business rules, dispatch, pricing, payment orchestration, object-storage signing và realtime.
- **PostgreSQL + PostGIS:** dữ liệu durable, job history, user/vehicle/KYC metadata và spatial data.
- **Redis:** online state, latest location, GEO index, distributed lock, offers/cache.
- **Third-party adapters:** Maps/Routes/Places, SMS/OTP, Push, Payment Gateway khi cần, S3-compatible Object Storage.

## 4. Mục tiêu MVP

### Customer

1. Đăng nhập bằng số điện thoại/OTP.
2. Lưu/chọn phương tiện của mình.
3. Chọn một trong 3 dịch vụ MVP.
4. Chọn điểm nhận xe/điểm đến hoặc thông tin đăng kiểm phù hợp.
5. Chọn ngay bây giờ hoặc hẹn giờ theo scope triển khai.
6. Nhận ETA + giá dự kiến.
7. Tạo yêu cầu và được match với tài xế/người thực hiện phù hợp.
8. Theo dõi trạng thái/GPS realtime khi phù hợp.
9. Xác nhận hoàn thành/bàn giao, thanh toán, đánh giá và xem lịch sử.

### Driver

1. Onboarding/KYC và được Admin duyệt.
2. Upload tài liệu trực tiếp lên object storage qua presigned URL.
3. Bật Online và gửi location theo policy.
4. Nhận offer đúng service capability, có countdown.
5. Accept/Reject atomically.
6. Dẫn đường tới khách.
7. Thực hiện workflow nhận xe → làm việc → bàn giao.
8. Xem lịch sử và thu nhập cơ bản.

### Operations

1. Theo dõi dashboard và live jobs.
2. Duyệt Driver/KYC.
3. Theo dõi/hỗ trợ job bất thường.
4. Cấu hình pricing/service rules.
5. Audit action nhạy cảm.

## 5. Domain quan trọng cần chuyển đổi

### Customer Vehicle

Business mới cần entity/aggregate **xe của khách hàng**, tối thiểu:
- owner user;
- type car/motorbike;
- plate number;
- brand/model/color;
- transmission;
- seats nếu áp dụng;
- notes/photo.

Vehicle của tài xế trong Grab-like model không còn là trung tâm của booking.

### Job/Service

Foundation `Trip` hiện tại có thể tái sử dụng nhưng semantics phải chuyển sang service/job theo 3 loại nghiệp vụ mới.

Đăng kiểm hộ có workflow dài hơn trip thông thường và cần state/substate đủ biểu diễn nhận xe → đăng kiểm → trả/bàn giao.

### Pricing

Pricing phải hỗ trợ service fee chứ không chỉ “giá xe chở khách/km”:
- base/minimum fee;
- distance;
- duration;
- waiting;
- scheduled fee;
- night/holiday surcharge;
- cancellation;
- fixed/package fee cho đăng kiểm hộ.

## 6. Stack chuẩn

| Layer | Technology |
|---|---|
| Customer/Rider | Flutter |
| Driver | Flutter |
| Admin | Next.js + TypeScript |
| Backend | Go |
| Database | PostgreSQL + PostGIS |
| Hot state / Geo | Redis |
| Realtime | WebSocket |
| Maps | Google Maps Platform hoặc provider tương đương |
| Push | FCM / APNs |
| Storage | S3-compatible object storage |
| Runtime | Docker / Docker Compose |

## 7. Nguyên tắc kiến trúc

- Modular monolith trước, service extraction sau khi có tải thật.
- PostgreSQL là source of truth cho durable business state.
- Redis giữ realtime/hot/geo/lock state.
- Dispatch phải có TTL, idempotency và distributed lock chống double assignment.
- Business transition phải do backend validate.
- Third-party provider đi qua adapter/interface.
- API/event versioned khi có production clients.
- Không lưu secret trong source/docs.

### Object storage — constraint bắt buộc

File media/KYC **không đi qua backend**.

```text
Client → presigned PUT → S3-compatible storage
Backend → sign URL + authorize + metadata only
```

Backend không nhận multipart/file binary cho flow này.

## 8. UX/UI direction

FlashX hướng tới chất lượng UX của các mobility app lớn tại Việt Nam, tham khảo cách tổ chức information hierarchy của Green SM nhưng giữ brand riêng.

Visual identity:
- Deep Navy;
- Electric Yellow;
- White;
- lightning X;
- rounded cards;
- rõ ràng, premium, tin cậy.

Customer home phải ưu tiên trực tiếp 3 dịch vụ MVP, không làm loãng bằng feature ngoài scope.

## 9. Không phải mục tiêu Phase 1

- Taxi fleet kiểu Grab/Uber.
- Food/grocery delivery.
- Giao hàng đại trà.
- Ví điện tử nội bộ phức tạp.
- AI dispatch/ML surge.
- Kubernetes nếu chưa có nhu cầu thực.
- Embedded turn-by-turn navigation tự xây.
- Mở rộng toàn quốc trước pilot.

## 10. KPI kỹ thuật & vận hành cần theo dõi

- Job creation success rate.
- Time-to-match.
- No-driver/no-match rate.
- Driver location freshness.
- WebSocket reconnect rate.
- Offer acceptance/expiry rate.
- Job completion/cancellation rate.
- ETA accuracy.
- API p95/error rate.
- KYC review turnaround.
- Job chờ quá SLA.
- Redis/Postgres saturation.

## 11. Trạng thái hiện tại

Foundation kỹ thuật đã có đáng kể từ phiên bản ride-hailing ban đầu và **nên tái sử dụng**:
- auth/session;
- dispatch;
- Redis GEO;
- realtime/WebSocket;
- GPS;
- payment cash ledger;
- rating/history;
- KYC/object storage;
- Admin foundation;
- Docker/CI.

Giai đoạn hiện tại là **business-domain realignment**: chuyển model/UX/API/data từ semantics Grab-like sang Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ trước khi đóng Release Candidate.

## 12. Quy mô Phase 1

MVP được thiết kế để pilot và kiểm chứng thị trường trong một khu vực/thành phố. Capacity production phải được benchmark/load test trên hạ tầng thật; không cam kết traffic cố định khi chưa có production profile.
