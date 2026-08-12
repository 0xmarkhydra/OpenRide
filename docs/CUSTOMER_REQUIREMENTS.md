# FlashX — Mong muốn khách hàng & phạm vi sản phẩm hiện tại

> **Source of truth nghiệp vụ — cập nhật 12/08/2026**
>
> Tài liệu này ghi nhận yêu cầu mới nhất của khách hàng và **thay thế giả định cũ rằng FlashX là một Grab/Uber clone thuần túy**. Nếu tài liệu cũ còn dùng từ `ride-hailing`, `bike/car` theo nghĩa xe của tài xế chở khách, phải diễn giải lại theo tài liệu này và cập nhật dần trong cùng đợt refactor domain.

## 1. Kết luận nghiệp vụ đã chốt

FlashX là nền tảng **tài xế lái hộ / hỗ trợ phương tiện theo yêu cầu**.

Khách hàng sử dụng **phương tiện của chính mình**; FlashX kết nối khách với tài xế/đối tác phù hợp để thực hiện dịch vụ.

### MVP chỉ tập trung 3 dịch vụ

1. **Lái hộ ô tô** — tài xế tới vị trí khách, nhận bàn giao và lái ô tô của khách tới điểm đến.
2. **Lái hộ xe máy** — tài xế tới vị trí khách, nhận bàn giao và lái xe máy của khách tới điểm đến.
3. **Đăng kiểm hộ** — nhận xe/giấy tờ theo quy trình, đưa xe đi thực hiện đăng kiểm và bàn giao lại cho khách.

### Không còn là định hướng MVP

- Grab/Uber clone theo mô hình tài xế dùng xe của mình để chở khách.
- Food delivery.
- Giao hàng đại trà.
- Thuê xe tự lái.
- Mở rộng nhiều vertical không liên quan trước khi 3 dịch vụ lõi chạy ổn định.

## 2. Định vị sản phẩm

Định vị đề xuất:

> **FlashX — Tài xế của bạn, khi bạn cần.**

Giá trị cốt lõi:
- **An toàn:** KYC tài xế chặt, lịch sử rõ ràng, GPS realtime, quy trình bàn giao xe.
- **Nhanh chóng:** tìm tài xế gần/phù hợp và hiển thị ETA rõ ràng.
- **Minh bạch:** giá, trạng thái công việc, lịch sử, thanh toán và hỗ trợ đều theo dõi được.
- **Tin cậy:** khách đang giao tài sản có giá trị lớn cho tài xế, nên trust phải là ưu tiên số 1.

## 3. Nhóm người dùng

### 3.1 Khách hàng

Khách có ô tô/xe máy và cần:
- người lái hộ ngay bây giờ;
- đặt trước tài xế;
- người thay mặt đưa xe đi đăng kiểm;
- theo dõi người thực hiện và trạng thái xe/công việc.

### 3.2 Tài xế / đối tác dịch vụ

Tài xế cần:
- đăng ký và KYC;
- khai báo năng lực/loại dịch vụ nhận được;
- online/offline;
- nhận/từ chối yêu cầu;
- dẫn đường tới khách;
- thực hiện quy trình nhận xe → làm việc → bàn giao;
- xem lịch sử và thu nhập.

### 3.3 Admin / Operations

Operations cần:
- theo dõi job realtime;
- duyệt và quản lý tài xế/KYC;
- cấu hình dịch vụ và giá;
- hỗ trợ khách/tài xế;
- xử lý job chờ lâu, sự cố, khiếu nại;
- audit hành động nhạy cảm.

## 4. Customer App — yêu cầu MVP

### 4.1 Auth & hồ sơ

- Đăng nhập/đăng ký bằng số điện thoại + OTP.
- Access/refresh session.
- Hồ sơ khách hàng cơ bản.

### 4.2 Trang chủ

UI theo hướng mobility hiện đại, dễ dùng như các app lớn tại Việt Nam nhưng mang nhận diện FlashX riêng.

Trang chủ cần ưu tiên 3 CTA rõ ràng:
- **Lái hộ ô tô**
- **Lái hộ xe máy**
- **Đăng kiểm hộ**

Không đưa quá nhiều dịch vụ phụ làm loãng MVP.

### 4.3 “Xe của tôi” — domain bắt buộc cần bổ sung

Khách phải có thể lưu/chọn phương tiện khi tạo yêu cầu.

Thông tin tối thiểu:
- `id`
- `owner_user_id`
- `type`: `car | motorbike`
- `license_plate`
- `brand`
- `model`
- `year` nếu có
- `color`
- `transmission`: `automatic | manual | n/a`
- `seats` nếu là ô tô
- `notes`
- `photo_object_key`/ảnh đại diện nếu có

### 4.4 Flow — Lái hộ ô tô

```text
Chọn Lái hộ ô tô
→ Chọn xe của tôi
→ Chọn điểm nhận xe/đón khách
→ Chọn điểm đến
→ Ngay bây giờ / Hẹn giờ
→ Xem ETA + giá dự kiến
→ Xác nhận yêu cầu
→ Tìm tài xế
→ Tài xế tới
→ Xác nhận bàn giao xe
→ Bắt đầu hành trình
→ Theo dõi realtime
→ Hoàn thành
→ Bàn giao xe
→ Thanh toán
→ Đánh giá
```

Thông tin nên hiển thị trước khi xác nhận:
- phương tiện;
- điểm nhận xe;
- điểm đến;
- quãng đường/ETA;
- giá dự kiến;
- thời gian đặt;
- ghi chú cho tài xế.

### 4.5 Flow — Lái hộ xe máy

Tương tự ô tô nhưng form phương tiện đơn giản hơn.

```text
Chọn Lái hộ xe máy
→ Chọn xe máy
→ Điểm nhận
→ Điểm đến
→ Ngay / Hẹn giờ
→ Báo giá
→ Match tài xế
→ Nhận xe
→ Thực hiện
→ Bàn giao
→ Thanh toán / đánh giá
```

### 4.6 Flow — Đăng kiểm hộ

Đây là workflow công việc, không nên ép y hệt trip chở khách.

Form MVP cần:
- xe cần đăng kiểm;
- địa chỉ nhận xe;
- địa chỉ trả xe nếu khác;
- thời gian mong muốn/hẹn giờ;
- trung tâm đăng kiểm (nếu khách chọn hoặc hệ thống điều phối);
- ghi chú;
- checklist giấy tờ cần bàn giao.

State UX đề xuất:

```text
Đã tạo yêu cầu
→ Đã ghép người thực hiện
→ Đang đến nhận xe
→ Đã nhận xe/giấy tờ
→ Đang di chuyển tới trung tâm
→ Đang thực hiện đăng kiểm
→ Đã hoàn tất đăng kiểm
→ Đang trả xe
→ Đã bàn giao
→ Hoàn thành
```

MVP có thể rút gọn state kỹ thuật nhưng UI phải thể hiện đúng bản chất công việc.

### 4.7 Theo dõi realtime và hành trình — BẮT BUỘC

UI trang chủ có thể đơn giản, nhưng sau khi khách đặt dịch vụ FlashX **bắt buộc có màn hình bản đồ realtime**.

Giai đoạn tài xế đang đến nhận khách/nhận xe:
- hiển thị vị trí realtime của tài xế trên bản đồ;
- hiển thị pickup của khách;
- vẽ tuyến đường tài xế → điểm nhận;
- hiển thị ETA còn bao lâu tài xế tới;
- hiển thị thông tin tài xế, trạng thái và nút liên hệ phù hợp.

Sau khi nhận xe và bắt đầu dịch vụ lái hộ:
- tiếp tục hiển thị vị trí realtime của tài xế/xe;
- vẽ hành trình từ điểm nhận → điểm đến;
- hiển thị tiến trình chuyến, quãng đường/ETA còn lại khi có dữ liệu;
- cập nhật trạng thái qua WebSocket/realtime;
- reconnect được khi app mất mạng/ngắt nền tạm thời;
- không để khách phải refresh thủ công để thấy tài xế di chuyển.

Đối với Đăng kiểm hộ:
- hiển thị vị trí người thực hiện khi policy cho phép;
- timeline nhận xe → đến trung tâm → đăng kiểm → trả xe;
- có thể hiển thị map hành trình nhận/trả xe thay vì ép toàn bộ workflow thành một trip duy nhất.

Ngoài tracking realtime:
- Timeline trạng thái dễ hiểu.
- Gọi/nhắn tin/hỗ trợ theo khả năng MVP.
- Lịch sử dịch vụ.
- Rating/feedback.

> **Quy tắc sản phẩm:** “App đơn giản” chỉ nói về số lượng chức năng/độ gọn của UI; **không được cắt GPS realtime, ETA, bản đồ tài xế đến đón và bản đồ hành trình chuyến đi** khỏi Full Marketplace.

## 5. Driver App — yêu cầu MVP

### 5.1 Onboarding/KYC

KYC ưu tiên **năng lực con người**, không coi xe của tài xế là tài sản phục vụ chuyến như Grab.

Tối thiểu nên có:
- CCCD/giấy tờ định danh;
- ảnh chân dung/selfie;
- GPLX + hạng bằng;
- ngày hết hạn GPLX;
- số năm kinh nghiệm;
- lý lịch tư pháp theo chính sách vận hành;
- tài khoản ngân hàng;
- năng lực xe số sàn/số tự động;
- loại dịch vụ có thể nhận;
- loại/dòng xe có kinh nghiệm nếu cần.

### 5.2 Upload tài liệu

**Nguyên tắc kiến trúc đã chốt:**

```text
Driver App → S3/Object Storage trực tiếp
Backend → chỉ cấp presigned URL/chữ ký + lưu metadata
```

Backend **không nhận/proxy file binary**.

Flow:
1. Client xin `upload-url`.
2. Backend kiểm tra auth/quyền/type/metadata và ký presigned PUT.
3. Client PUT file trực tiếp lên object storage.
4. Client gọi `complete`.
5. Backend lưu object key + metadata.
6. Khi xem tài liệu, backend cấp presigned GET theo quyền.

### 5.3 Driver Home

- ONLINE/OFFLINE cực rõ.
- GPS/status realtime.
- Số job/thu nhập tóm tắt.
- Capability/service preferences.
- Cảnh báo KYC nếu hồ sơ chưa đủ.

### 5.4 Offer

Offer phải hiển thị:
- loại dịch vụ: Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ;
- pickup;
- destination/trung tâm đăng kiểm nếu áp dụng;
- khoảng cách/ETA tới khách;
- giá/thu nhập theo policy;
- thông tin phương tiện cần thiết;
- **countdown**;
- Accept/Reject CTA lớn.

### 5.5 Workflow thực hiện

Lái hộ:

```text
Accepted
→ Arriving
→ Arrived
→ Nhận xe / xác nhận tình trạng
→ Start
→ In progress
→ Complete
→ Bàn giao
```

Đăng kiểm hộ:
- cần workflow riêng hoặc state/substate đủ biểu diễn nhận xe → đăng kiểm → trả xe.

### 5.6 Earnings & history

- job history;
- thu nhập theo ngày/tuần/tháng tối thiểu;
- trạng thái thanh toán/đối soát theo phase.

## 6. Admin / Operations — yêu cầu MVP

### Dashboard
- tổng job hôm nay;
- job theo 3 loại dịch vụ;
- tài xế online;
- doanh thu;
- completion/cancellation/no-match;
- cảnh báo job chờ lâu/KYC sắp hết hạn.

### Live Jobs
- filter theo 3 service type;
- trạng thái;
- khách;
- tài xế;
- pickup/destination;
- live map khi phù hợp;
- detail drawer/timeline.

### Driver/KYC
- danh sách pending/approved/rejected;
- xem document bằng presigned GET;
- approve/reject/request-more-info;
- audit decision.

### Pricing
- cấu hình giá theo service;
- version/effective time;
- phụ phí giờ đêm/giờ cao điểm/ngày lễ nếu áp dụng;
- waiting/cancellation rules;
- không hard-code giá trong mobile.

### Support
- tra cứu job;
- xem timeline;
- liên hệ khách/tài xế;
- support note;
- incident/audit trail.

## 7. Service types — nguồn chuẩn mới

Tên business đề xuất:

```text
designated_driver_car
designated_driver_bike
vehicle_inspection_assist
```

Các giá trị `car`/`bike` cũ mang nghĩa ride-hailing phải được migrate/compat theo kế hoạch, không tiếp tục mở rộng semantics cũ.

## 8. Pricing — thay đổi bắt buộc

Pricing cũ dạng “base + giá/km theo xe chở khách” không đủ cho business mới.

Pricing engine cần hỗ trợ các component:
- minimum/base service fee;
- distance component;
- duration component;
- waiting fee;
- scheduled booking fee nếu có;
- night/holiday surcharge;
- intercity component nếu mở rộng;
- fixed fee/package cho đăng kiểm hộ;
- cancellation fee theo policy;
- promotion/discount sau MVP nếu cần.

Tất cả cấu hình từ backend/Admin và có version để audit.

## 9. Trust & Safety — khác biệt sản phẩm quan trọng

Khách giao phương tiện có giá trị lớn, nên FlashX phải coi trust là feature cốt lõi.

Nên triển khai theo mức ưu tiên:

### P0/MVP
- KYC tài xế;
- GPLX và expiry;
- GPS realtime;
- lịch sử job;
- rating;
- support;
- audit Admin;
- xác nhận nhận/bàn giao xe ở mức cơ bản.

### P1
- ảnh hiện trạng xe trước/sau;
- odometer/fuel;
- ghi nhận vết xước;
- OTP/PIN bàn giao;
- incident workflow;
- share trip/job;
- SOS;
- bảo hiểm/quy trình claim theo đối tác kinh doanh.

## 10. UX/UI direction đã chốt

### Phong cách
- chất lượng/độ rõ tương đương các mobility app lớn tại Việt Nam;
- tham khảo cách tổ chức thông tin của Green SM nhưng **không sao chép nhận diện**;
- FlashX có identity riêng: **Deep Navy + Electric Yellow + White**, lightning X;
- rounded cards, hierarchy rõ, ít nhiễu, thao tác một tay tốt.

### Customer App
- service-first + map/context;
- 3 dịch vụ MVP nổi bật ngay trang chủ;
- một primary CTA/màn hình;
- trạng thái công việc viết bằng tiếng Việt, không phơi technical state.

### Driver App
- cockpit navy;
- CTA lớn;
- countdown offer;
- trạng thái tiếp theo luôn rõ;
- ưu tiên đọc ngoài trời/khi di chuyển.

### Admin
- desktop SaaS/operations;
- dark navy sidebar + yellow accent;
- table/filter/drawer/map/KPI;
- density cao nhưng dễ scan.

Mock đã được xây theo 3 nhóm:
1. Customer/Rider UX board.
2. FlashX Driver UX board.
3. FlashX Admin/Operations board.

Các mock là **visual target**, không phải bằng chứng mọi feature đã implement.

## 11. Technical baseline đã có và nên tái sử dụng

- Flutter Rider/Customer App.
- Flutter Driver App.
- Next.js Admin.
- Go modular backend.
- PostgreSQL + PostGIS.
- Redis + Redis GEO.
- WebSocket realtime.
- OTP/auth + refresh token.
- Dispatch offers + TTL + distributed lock.
- GPS/location pipeline.
- Trip/job lifecycle foundation.
- Payment ledger cash-first.
- Rating/history foundation.
- Docker Compose project `flashx`.
- CI baseline.
- S3-compatible presigned object storage flow.

Không đập lại các foundation này nếu business refactor có thể tái sử dụng.

## 12. Third-party / ENV cần cho dev và production

### Dev hiện tại
- GPS của thiết bị **không cần Google API key**.
- Map/routes/place search có thể dùng Google Maps Demo Key hoặc provider dev phù hợp.
- S3/object storage cần endpoint/bucket/access credentials ở local env, không gửi secret vào chat/docs.

### Production sau này
- restricted Maps keys theo Android/iOS/backend;
- SMS OTP provider thật;
- push FCM/APNs;
- object storage production;
- JWT/secret production;
- payment gateway nếu khách đưa vào scope;
- monitoring/error tracking.

## 13. Out of scope MVP mặc định

Trừ khi khách Change Request:
- food delivery;
- parcel delivery;
- taxi fleet;
- ví điện tử nội bộ;
- referral/loyalty phức tạp;
- dynamic ML surge;
- call masking riêng;
- embedded navigation tự xây;
- mở rộng toàn quốc trước pilot.

## 14. Definition of Done — MVP theo business mới

MVP chỉ được gọi là bàn giao khi tối thiểu:

1. Customer tạo được yêu cầu **Lái hộ ô tô** end-to-end.
2. Customer tạo được yêu cầu **Lái hộ xe máy** end-to-end.
3. Customer tạo được yêu cầu **Đăng kiểm hộ** end-to-end với workflow phù hợp.
4. Customer chọn/lưu được phương tiện của mình.
5. Driver approved nhận đúng loại offer và accept atomically.
6. Realtime location/status hoạt động và reconnect được.
7. Driver KYC upload trực tiếp object storage; Admin review được bằng signed access.
8. Pricing của 3 service nằm ở backend/Admin.
9. Cash-first payment/history/rating hoạt động ở các flow áp dụng.
10. Admin xem được job/driver/KYC/timeline và hỗ trợ vận hành.
11. Rider/Driver/Admin test/build release gate xanh.
12. Có full E2E/UAT trên thiết bị thật.
13. Không còn dev secret/mock provider trong production config.
14. Tài liệu deploy/runbook/handover được cập nhật theo domain mới.

## 15. Những điểm cần khách hàng quyết định trước production

- Thành phố/khu vực pilot đầu tiên.
- Giá từng dịch vụ và cancellation/waiting policy.
- Hẹn giờ bắt buộc ở MVP hay rollout ngay sau MVP.
- Quy trình nhận/bàn giao xe và bằng chứng hiện trạng tối thiểu.
- Quy trình đăng kiểm hộ chi tiết và giấy tờ khách phải bàn giao.
- KYC/lý lịch tư pháp/đào tạo tài xế theo chính sách vận hành.
- Mô hình bảo hiểm/trách nhiệm khi có sự cố.
- SMS OTP provider.
- Payment online có nằm trong MVP không hay cash-first.
- SLA hỗ trợ khách hàng.

---

**Quy tắc:** Khi yêu cầu khách hàng thay đổi, cập nhật file này trước; sau đó mới cập nhật PRD/domain/API/data model/backlog để tránh các tài liệu mâu thuẫn nhau.
