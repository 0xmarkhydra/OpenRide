# FlashX — Implementation Gap Review 12/08/2026

> Mục tiêu: đối chiếu code hiện tại với yêu cầu mới nhất trong `CUSTOMER_REQUIREMENTS.md`: **Full Marketplace**, 3 dịch vụ Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ, phương tiện là **xe của khách**, có vị trí tài xế tới nhận xe và theo dõi hành trình theo thời gian thực.

## 1. Kết luận nhanh

Code hiện tại có **foundation tốt để tái sử dụng**, đặc biệt auth, PostgreSQL/PostGIS, Redis GEO, dispatch offer, atomic accept, GPS ingestion, WebSocket, payment ledger, rating, KYC/S3 và khung Customer/Driver/Admin.

Tuy nhiên phần business/UI đang chạy vẫn chủ yếu là **ride-hailing MVP cũ**. Chưa thể coi là Full Marketplace theo yêu cầu mới nếu chưa xử lý các blocker dưới đây.

### Blocker P0

1. Chưa có `CustomerVehicle` / “Xe của tôi” trong DB/API/app.
2. `service_type` trong code vẫn dùng `car` / `bike`; 3 service mới mới chỉ có trong docs.
3. Pricing hiện vẫn là `base + km` cho `car/bike`, chưa có pricing v2 và package Đăng kiểm hộ.
4. Chưa có scheduled booking / `scheduled_at`.
5. Chưa có workflow Đăng kiểm hộ.
6. Chưa có state nhận xe / bàn giao xe cho designated-driver.
7. Driver chưa có capability theo dịch vụ, bằng số sàn/tự động, hạng GPLX/expiry ở domain chính.
8. Customer map hiện chỉ vẽ marker; chưa vẽ route/polyline tài xế → khách và pickup → destination, chưa tính ETA tài xế tới khách.
9. Realtime client mobile hiện chưa có reconnect/backoff/resubscribe đủ để đảm bảo mở lại app/mất mạng vẫn bám đúng active job.
10. Admin “bản đồ vận hành” hiện là UI mô phỏng, chưa phải live map thật.

---

## 2. Những phần có thể giữ lại

### Backend

- OTP/auth + access/refresh session.
- Role Rider/Driver/Admin hiện có thể migrate naming dần, không cần rewrite auth.
- Trip persistence/state foundation có thể dùng làm base cho `Job/ServiceRequest` compatibility layer.
- Redis GEO nearby candidate discovery.
- Offer TTL/countdown.
- Distributed/atomic accept và chống double assignment.
- Driver GPS ingestion.
- WebSocket publish `driver.location` tới customer đang có foundation.
- Cash-first payment ledger.
- Rating/history.
- KYC/object storage direct upload.
- Google Routes provider boundary.

### Customer App

- Auth/session.
- Google Maps widget.
- GPS current location.
- Estimate/create/cancel/history/rating foundation.
- Có nhận event `driver.location` và hiển thị marker tài xế.

### Driver App

- Auth/session.
- Online/Offline.
- GPS stream và gửi location.
- Offer/Accept/Reject/countdown.
- Basic trip lifecycle.
- KYC direct-upload client foundation.

### Admin

- Auth.
- Dashboard foundation.
- Driver approval.
- Job/trip list foundation.

---

## 3. Gap chi tiết theo domain

### 3.1 Service type — phải migrate

Code backend và test vẫn đang dùng:

```text
bike
car
```

Target:

```text
designated_driver_car
designated_driver_bike
vehicle_inspection_assist
```

Cần compatibility migration để app cũ không vỡ trong quá trình chuyển đổi.

### 3.2 Customer Vehicle — chưa implement

Migration hiện tại có bảng `vehicles` gắn với `driver_id`, phản ánh mô hình tài xế có xe kiểu ride-hailing.

Target cần bảng riêng cho **phương tiện của khách**:
- owner/customer id;
- loại xe;
- biển số;
- hãng/model/màu;
- số sàn/tự động;
- số chỗ;
- ghi chú/ảnh.

Trip/job phải tham chiếu `customer_vehicle_id`.

### 3.3 Pricing v2 — chưa implement

`pricing.Service` hiện hard-code:
- bike: base + per km;
- car: base + per km.

Cần pricing theo 3 dịch vụ:
- minimum/base;
- distance;
- duration;
- waiting;
- scheduled fee;
- night/holiday;
- cancellation;
- fixed/package cho Đăng kiểm hộ;
- version/effective time;
- cấu hình qua Admin.

### 3.4 Scheduling — chưa implement

Không tìm thấy contract/domain cho `scheduled_at` trong backend hiện tại.

Cần:
- immediate/scheduled;
- timezone/lead time;
- dispatch trước giờ hẹn;
- reminder;
- edit/cancel policy;
- Admin filter.

### 3.5 Job workflow — còn taxi semantics

State hiện tại:

```text
searching -> accepted -> arriving -> arrived -> in_progress -> completed
```

Lái hộ cần ít nhất thêm semantics:

```text
accepted -> arriving -> arrived -> vehicle_received -> in_progress -> handover -> completed
```

Đăng kiểm hộ cần workflow/substatus riêng từ nhận xe/giấy tờ → trung tâm → đăng kiểm → trả xe → bàn giao.

### 3.6 Driver capability — chưa đủ

Driver model hiện có một `service_type` duy nhất.

Cần chuyển sang capability/profile:
- nhận được service nào;
- bằng lái/hạng bằng + expiry;
- số sàn/tự động;
- kinh nghiệm;
- inspection support nếu có;
- Admin duyệt capability.

Dispatch phải filter theo capability chứ không chỉ `driver.service_type == trip.service_type`.

---

## 4. Realtime map & hành trình — foundation có, UX chưa đạt yêu cầu

### Đã có

- Driver App lấy GPS và POST location.
- Backend publish `driver.location` cho rider của active trip.
- Customer controller nhận event.
- Customer map có marker pickup, destination và driver.

### Còn thiếu P0

- Route/polyline **driver → điểm nhận** sau khi match.
- ETA tài xế tới khách cập nhật theo vị trí mới.
- Camera fit/animate theo driver + pickup/destination hợp lý.
- Route/polyline **pickup → destination** khi bắt đầu chuyến.
- Remaining ETA/distance trong chuyến.
- Re-route khi tài xế lệch tuyến nếu policy cần.
- Reconnect/backoff và resync snapshot sau mất mạng/background.
- Lưu/sampling trip path tối thiểu cho support/audit nếu business yêu cầu.

Google Routes provider hiện chỉ request `distanceMeters` + `duration`; chưa lấy encoded polyline để mobile vẽ route.

---

## 5. Customer App — hiện vẫn phải redesign theo yêu cầu mới

Home hiện tại vẫn hỏi `Bạn muốn đi đâu?` và chỉ có:
- `FlashX Bike`;
- `FlashX Car`.

Đây là semantics taxi cũ.

P0 cần đổi thành service-first:
- Lái hộ ô tô;
- Lái hộ xe máy;
- Đăng kiểm hộ;
- “Xe của tôi”;
- Ngay/Hẹn giờ.

Copy active trip hiện còn các câu như:
- `Điểm đón của bạn`;
- `Bạn đang trên chuyến đi`;
- `kiểm tra biển số trước khi lên xe`.

Phải chuyển sang đúng business: tài xế tới **nhận xe của khách** và lái xe của khách.

---

## 6. Driver App — cần đổi từ chở khách sang lái xe khách hàng

Offer hiện chưa hiển thị:
- service type mới;
- xe khách hàng;
- biển số/model/transmission;
- schedule;
- checklist đăng kiểm.

Execution copy còn:
- `Đi tới điểm đón của khách`;
- `Chỉ bắt đầu sau khi khách đã lên xe`.

Phải đổi thành:
- tới vị trí khách;
- xác nhận nhận xe;
- kiểm tra xe/thông tin bàn giao;
- bắt đầu lái xe khách;
- bàn giao xe;
- flow riêng cho đăng kiểm.

Phần `Phương tiện` trong tài khoản Driver hiện cũng phản ánh taxi semantics và cần đổi thành `Năng lực & loại dịch vụ` hoặc tương đương.

---

## 7. Admin — hiện mới là dashboard foundation

Admin hiện:
- có KPI;
- danh sách tài xế chờ duyệt;
- danh sách trip;
- “bản đồ vận hành” mới chỉ là mock surface, chưa có map/location thật.

P0 cần:
- Live Jobs thật;
- filter 3 service;
- customer vehicle;
- driver capability;
- scheduled booking;
- inspection timeline;
- pricing v2;
- signed KYC review đầy đủ;
- support/reassign/manual intervention;
- live map và job detail/timeline.

---

## 8. Hạ tầng dữ liệu cần migration mới

Không sửa migration đã chạy. Tạo migration tiếp theo, dự kiến gồm:
- `customer_vehicles`;
- thêm `customer_vehicle_id` vào job/trip compatibility model;
- `scheduled_at` và booking mode;
- service type migration/compatibility;
- job/service metadata cho inspection;
- handover/status event metadata;
- driver capabilities;
- pricing v2 fields/rules;
- indexes phù hợp.

Không nên rename/drop bảng cũ ngay trong một bước; migrate incremental để giữ khả năng rollback.

---

## 9. Thứ tự triển khai đề xuất

1. Khóa service constants + compatibility.
2. Migration + API `CustomerVehicle`.
3. Driver capability model.
4. Job contract + scheduling.
5. Pricing v2.
6. Lái hộ ô tô end-to-end.
7. Realtime route/ETA/reconnect hoàn chỉnh.
8. Lái hộ xe máy end-to-end.
9. Đăng kiểm hộ workflow end-to-end.
10. Admin Operations theo domain mới.
11. Full regression + UAT thiết bị thật + production adapters.

---

## 10. Trạng thái review

- Review này đối chiếu source code hiện tại với docs business mới.
- Chưa coi các build/test checkpoint cũ là trạng thái hiện tại; phải chạy lại gate sau khi bắt đầu refactor.
- Chưa thực hiện domain refactor lớn trong review này để tránh trộn thay đổi tài liệu với migration/code feature.
