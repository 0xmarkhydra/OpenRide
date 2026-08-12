# PRD — FlashX MVP

> **Business baseline: 12/08/2026**  
> Chi tiết yêu cầu khách hàng: [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md)

## 1. Product objective

Xây phiên bản thương mại đầu tiên của FlashX cho mô hình **tài xế lái hộ / hỗ trợ phương tiện theo yêu cầu**.

FlashX kết nối khách hàng có phương tiện riêng với tài xế/người thực hiện đã được xác minh. MVP phải chứng minh được 3 flow end-to-end và khả năng vận hành thực tế, thay vì tiếp tục mở rộng như một Grab/Uber clone.

## 2. MVP services

Chỉ gồm:

```text
designated_driver_car       = Lái hộ ô tô
designated_driver_bike      = Lái hộ xe máy
vehicle_inspection_assist   = Đăng kiểm hộ
```

Mọi UI/API/pricing mới phải dùng semantics này.

## 3. Personas

### Customer
- Có ô tô/xe máy của mình.
- Muốn đặt người lái hộ hoặc đưa xe đi đăng kiểm hộ.
- Quan tâm đặc biệt tới trust, ETA, giá, vị trí người thực hiện và trạng thái xe/công việc.

### Driver / Service Partner
- Muốn đăng ký, KYC, được duyệt, bật Online và nhận job phù hợp với năng lực.
- Cần flow thực hiện rõ ràng và thu nhập minh bạch.

### Operations/Admin
- Duyệt tài xế/KYC.
- Theo dõi live jobs.
- Cấu hình giá/service.
- Hỗ trợ sự cố/đơn chờ lâu.
- Audit hành động nhạy cảm.

## 4. Customer requirements

### 4.1 Account
- Đăng ký/đăng nhập bằng số điện thoại.
- OTP qua provider abstraction.
- Access/refresh session.
- Hồ sơ cơ bản.

### 4.2 Customer Vehicle
Customer phải lưu/chọn được phương tiện của mình.

MVP fields tối thiểu:
- type: car/motorbike;
- license plate;
- brand/model;
- color;
- transmission nếu áp dụng;
- seats nếu áp dụng;
- note;
- ảnh đại diện optional.

### 4.3 Home / service selection
Trang chủ chỉ ưu tiên ba service card:
- Lái hộ ô tô;
- Lái hộ xe máy;
- Đăng kiểm hộ.

Search/map/context là hỗ trợ, không được làm loãng lựa chọn dịch vụ.

### 4.4 Booking — Lái hộ ô tô
- chọn xe của khách;
- pickup/địa chỉ nhận xe;
- destination;
- current location/manual place selection;
- route distance + ETA;
- ngay bây giờ / hẹn giờ;
- ghi chú;
- estimated fare;
- confirm request;
- cancel theo policy.

### 4.5 Booking — Lái hộ xe máy
Tương tự ô tô nhưng field phương tiện đơn giản hơn; service/pricing/capability tách riêng.

### 4.6 Booking — Đăng kiểm hộ
Form riêng, tối thiểu:
- xe;
- địa chỉ nhận xe;
- địa chỉ trả xe nếu khác;
- thời gian mong muốn;
- trung tâm đăng kiểm nếu được chọn;
- checklist/ghi chú giấy tờ bàn giao;
- giá dự kiến/package.

### 4.7 Matching/tracking
- nhận driver/partner sau match;
- xem ETA tới điểm nhận;
- xem location realtime khi policy cho phép;
- timeline trạng thái bằng tiếng Việt;
- gọi/nhắn/hỗ trợ theo capability MVP.

### 4.8 Completion
- final fare/payment state;
- xác nhận bàn giao/hoàn thành;
- rating/feedback;
- job/service history.

## 5. Driver requirements

### 5.1 Onboarding/KYC
- tài khoản;
- profile;
- CCCD/identity;
- selfie/portrait;
- GPLX + hạng bằng + expiry;
- kinh nghiệm;
- lý lịch tư pháp theo policy;
- tài khoản ngân hàng;
- capability: car/bike/inspection support;
- transmission capability nếu cần;
- Admin approve/reject/request-more-info.

### 5.2 KYC media
Upload phải là direct-to-object-storage:

```text
Driver App -> request signed URL -> S3 PUT trực tiếp
Driver App -> complete metadata -> API
```

Backend không nhận file binary/multipart cho KYC media.

### 5.3 Availability/location
- Online/Offline.
- Approved mới được Online.
- Location permission/GPS đáp ứng policy mới được nhận offer.
- Khi Online, location gửi theo adaptive policy.

### 5.4 Offer
Offer hiển thị:
- service type;
- pickup;
- destination/inspection location nếu áp dụng;
- ETA/distance tới customer;
- vehicle summary cần thiết;
- fare/earning theo policy;
- countdown;
- Accept/Reject.

Accept phải atomic/idempotent; UI không tự coi là thành công trước server response.

### 5.5 Lái hộ workflow
Target UX/business flow:

```text
OFFERED
-> ACCEPTED
-> ARRIVING
-> ARRIVED
-> VEHICLE_RECEIVED
-> IN_PROGRESS
-> HANDOVER
-> COMPLETED
```

Implementation có thể giữ compatibility state machine ngắn hạn nhưng API/UI phải có kế hoạch migrate rõ.

### 5.6 Đăng kiểm hộ workflow
Target business flow:

```text
OFFERED
-> ACCEPTED
-> ARRIVING_FOR_PICKUP
-> VEHICLE_RECEIVED
-> EN_ROUTE_TO_INSPECTION
-> INSPECTION_IN_PROGRESS
-> INSPECTION_COMPLETED
-> RETURNING_VEHICLE
-> HANDOVER
-> COMPLETED
```

MVP có thể map một số bước vào status + substatus/event history, nhưng Operations phải nhìn được tiến trình thực tế.

### 5.7 Earnings/history
- completed job history;
- gross earning summary;
- payment/settlement state khi phase hỗ trợ.

## 6. Admin requirements

### Dashboard
- jobs hôm nay;
- split theo 3 service;
- searching/active/completed/cancelled;
- online drivers;
- time-to-match/no-match;
- revenue;
- alerts.

### Live Jobs
- search/filter;
- service/status/driver/customer;
- live map sampled khi phù hợp;
- job detail/timeline;
- support notes/action.

### Driver/KYC
- profile;
- capability;
- documents qua signed GET;
- approve/reject/request more info;
- suspend/unsuspend;
- audit.

### Customer/Vehicles
- customer search/detail;
- customer vehicle detail;
- job history;
- support context.

### Pricing
- service-specific pricing rules;
- version/effective time;
- waiting/night/holiday/scheduled/cancellation components;
- inspection package/fixed fee;
- audit pricing changes.

## 7. Job state & invariants

Backend là authority cho state transition.

Các invariant nền tảng vẫn giữ:
- một active job chỉ có một assigned driver tại một thời điểm;
- driver không nhận hai active job nếu policy chưa cho phép;
- final fare do backend chốt;
- transition sai thứ tự bị reject;
- action quan trọng ghi history/audit;
- accept race được bảo vệ bằng distributed/atomic locking.

State machine implementation hiện tại có thể được migrate incremental để không phá foundation đang chạy.

## 8. Pricing MVP v2

Pricing input có thể gồm:
- service type;
- pickup/destination route;
- distance;
- duration;
- booking time/schedule;
- waiting duration;
- zone/time rules;
- vehicle attributes cần thiết;
- promotion nếu phase có.

Components:
- base/minimum service fee;
- distance fare;
- duration fare;
- waiting fee;
- schedule fee;
- night/holiday surcharge;
- cancellation fee;
- inspection fixed/package fee;
- discount.

Pricing config nằm ở backend/Admin, có version/effective time; mobile không hard-code.

## 9. Dispatch MVP

- Redis GEO candidate discovery.
- Filter approved/online/available/fresh-location.
- Filter theo service capability.
- Score theo distance + idle time + quality signals cơ bản.
- Offer TTL/countdown.
- Distributed lock/atomic assignment.
- Retry/search-radius expansion.
- No-driver-found outcome rõ ràng.
- Operations/manual intervention có thể P1 nếu pilot cần.

## 10. Trust & Safety MVP

P0:
- KYC tài xế;
- GPLX/expiry;
- GPS realtime;
- rating/history;
- Admin audit;
- support path;
- nhận/bàn giao xe ở mức cơ bản.

P1:
- ảnh xe trước/sau;
- odometer/fuel;
- damage notes;
- OTP/PIN handover;
- incident claim flow;
- SOS/share job;
- bảo hiểm integration/process.

## 11. Payment MVP

- Cash-first được chấp nhận cho MVP.
- Payment state tách khỏi job state.
- Mọi completed service có ledger/payment record phù hợp.
- Online gateway chỉ đưa vào P0 nếu khách cung cấp merchant account và chốt provider đúng hạn.

## 12. UX/UI requirements

### Brand
- Deep Navy + Electric Yellow + White.
- Lightning X.
- Premium, nhanh, tin cậy.

### Reference quality
- Tham khảo mức độ polish/hierarchy của các mobility app lớn như Green SM.
- Không sao chép brand asset/identity của đối thủ.

### Customer
- service-first;
- 3 service CTA rõ ngay home;
- map hỗ trợ task;
- one primary action/screen;
- status copy tiếng Việt.

### Driver
- dark cockpit;
- CTA lớn;
- countdown offer;
- next action rõ;
- readability ngoài trời.

### Admin
- operations-first;
- dashboard/table/filter/drawer/map;
- semantic status chips;
- responsive desktop/tablet hợp lý.

## 13. Non-functional requirements

- HTTPS/TLS production.
- Secure token storage.
- RBAC Admin.
- Idempotency.
- Structured logs/request ID.
- Error tracking/observability.
- Backup/restore production.
- Config/secrets ngoài source.
- Graceful network/reconnect.
- Rate limit cho auth/OTP/critical public endpoints.
- KYC/PII privacy/access policy.

## 14. Acceptance criteria MVP

MVP đạt yêu cầu khi:

1. Customer tạo và hoàn thành Lái hộ ô tô end-to-end.
2. Customer tạo và hoàn thành Lái hộ xe máy end-to-end.
3. Customer tạo và hoàn thành Đăng kiểm hộ với workflow phù hợp.
4. Customer quản lý/chọn phương tiện của mình.
5. Driver approved nhận đúng loại offer.
6. Một và chỉ một driver được assign.
7. Realtime location/status hoạt động và reconnect đúng.
8. KYC direct upload + Admin review signed access hoạt động.
9. Pricing của 3 service do backend/Admin điều khiển.
10. Payment/history/rating lưu đúng.
11. Admin quản lý được customer/driver/job/KYC/pricing/support context.
12. Unit/integration/race/mobile/admin checks xanh.
13. UAT trên thiết bị thật qua đủ 3 flow.
14. Không còn Blocker/Critical bug core.
15. Production config không chứa dev/mock/secret sai môi trường.

## 15. Out of scope mặc định

- Taxi fleet kiểu Grab/Uber.
- Food delivery.
- Parcel delivery đại trà.
- Multi-stop.
- Wallet phức tạp.
- Referral/loyalty nâng cao.
- ML surge/AI fraud.
- Embedded navigation tự xây.
- Call masking/VoIP riêng.
- Launch toàn quốc ngay Phase 1.

## 16. Open decisions cần khách hàng chốt

- Pilot ở thành phố/khu vực nào.
- Bảng giá 3 service.
- Hẹn giờ là P0 bắt buộc hay rollout sau immediate booking.
- Waiting/cancellation/night/holiday policy.
- Quy trình nhận/bàn giao xe tối thiểu.
- Checklist giấy tờ cho đăng kiểm hộ.
- Bảo hiểm/trách nhiệm sự cố.
- SLA support.
- Payment online có trong MVP không.
- SMS OTP provider production.
