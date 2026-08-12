# FlashX — Tóm tắt thread dự án đến 12/08/2026

## 1. Mục đích tài liệu

Tài liệu này tóm tắt các quyết định, thay đổi phạm vi, trạng thái kỹ thuật và mong muốn khách hàng được chốt trong thread làm việc hiện tại.

Đây là checkpoint để đội BA/Product/Design/Engineering có thể mở thread mới và tiếp tục mà không phải đọc lại toàn bộ lịch sử chat.

> Yêu cầu nghiệp vụ mới nhất nằm tại [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md) và được ưu tiên hơn các giả định ride-hailing cũ.

## 2. Timeline quyết định chính

### 2.1 Khởi đầu — ý tưởng app kiểu Grab/Uber

Dự án bắt đầu từ câu hỏi xây một app tương tự Grab/Uber và ước tính chi phí.

Từ đó đã hình thành scope kỹ thuật ban đầu:
- Rider App;
- Driver App;
- Admin Portal;
- Go backend;
- PostgreSQL/PostGIS;
- Redis/Redis GEO;
- WebSocket realtime;
- Maps/routing;
- auth/OTP;
- dispatch;
- pricing;
- payment;
- CI/Docker và tài liệu bàn giao.

### 2.2 Chọn stack và kiến trúc

Định hướng đã chốt:
- backend: **Go**;
- mobile: **Flutter**;
- admin: **Next.js + TypeScript**;
- durable data: **PostgreSQL + PostGIS**;
- realtime/hot state/geo/locks: **Redis**;
- realtime client update: **WebSocket**;
- object storage: **S3-compatible**;
- modular monolith trước, chỉ tách service khi tải thực tế yêu cầu.

Lý do chính: tốc độ phát triển MVP, concurrency tốt, dễ mở rộng và không over-engineer sớm.

### 2.3 Project được tách khỏi `codex-mcp`

Ban đầu code được tạo tạm trong repo `codex-mcp`.

Sau đó project được move ra:

```text
~/Documents/flashx
```

Project có Git repo riêng và không còn phụ thuộc cấu trúc repo `codex-mcp`.

### 2.4 Đổi brand thành FlashX

Tên dự án được đổi từ Grap/Grap-like sang **FlashX**.

Đã đổi phần lớn:
- folder/project name;
- Go module/imports;
- Rider/Driver package naming;
- Android package naming;
- Admin branding;
- env naming;
- Docker Compose naming;
- docs/config;
- UI brand direction.

Visual identity chốt:
- Deep Navy;
- Electric Yellow;
- White;
- lightning-bolt X;
- premium / nhanh / tin cậy / hiện đại.

### 2.5 Docker cleanup và chuẩn hóa

Trước đó quá trình test thủ công tạo nhiều container `grap-*`.

Đã xử lý:
- restart Docker Desktop sau lỗi containerd metadata;
- xóa container/volume/network Grap cũ;
- reclaim nhiều GB Docker resource không dùng;
- giữ volume của project khác;
- chuẩn hóa Compose thành **một group `flashx`**.

Stack dài hạn chuẩn:

```text
flashx
├── api
├── admin
├── postgres
└── redis
```

Test container dùng:

```text
docker compose run --rm ...
```

để không tích rác container.

### 2.6 Core backend đã được production-hardening đáng kể

Trong thread đã triển khai/kiểm tra các foundation:
- OTP/Auth;
- access/refresh tokens;
- Rider/Driver/Admin roles;
- driver approval;
- trip state foundation;
- Redis GEO;
- dispatch offers;
- distributed trip lock;
- Redis-backed offer store;
- restart/recovery của pending offer;
- WebSocket realtime;
- GPS/location;
- pricing foundation;
- payment ledger cash-first;
- rating;
- history/driver earning summary;
- Admin operational APIs;
- Google Routes adapter boundary;
- SMS webhook provider boundary;
- Docker/CI.

Backend đã từng chạy full:

```text
go test ./...
```

và pass sau các thay đổi object storage/KYC gần nhất.

## 3. Object storage / KYC — quyết định kiến trúc quan trọng

Từ project `MediaUpload` trong Documents, phần S3-compatible storage được tham khảo nhưng **không bê cơ chế server-upload**.

Yêu cầu trực tiếp của chủ dự án:

> Backend không được upload/proxy media. Tất cả file phải đi trực tiếp từ client tới storage; backend chỉ cấp chữ ký.

Kiến trúc chuẩn:

```text
Client
  ↓ request upload signature
FlashX API
  ↓ presigned PUT
Client
  ↓ upload binary trực tiếp
S3-compatible Object Storage
  ↓ complete callback/command
FlashX API
  ↓ lưu object key + metadata
PostgreSQL
```

Khi xem file:

```text
Client/Admin → API kiểm tra quyền → presigned GET → Object Storage
```

Đã thêm AWS SDK Go V2 vào riêng `services/api` để ký SigV4/presigned requests.

Nguyên tắc:
- secret storage chỉ ở backend env;
- mobile không biết access secret;
- backend không nhận multipart/file binary;
- KYC file access phải signed + authorization.

## 4. Maps / định vị

Một điểm đã làm rõ với chủ dự án:

- **GPS vị trí hiện tại của user không cần Google Maps API key.**
- Thiết bị iOS/Android cung cấp location qua Location Services.
- API key cần cho map rendering, places search, routing/ETA/geocoding.

Trong development có thể dùng Google Maps Demo Key/quota dev phù hợp.

Production sẽ cần key tách/restrict theo:
- Rider Android;
- Driver Android;
- Rider iOS;
- Driver iOS;
- backend Routes/Places.

Không commit key/secret vào Git.

## 5. UI/UX — quá trình thay đổi

### 5.1 Ban đầu

Code UI ban đầu còn mang cảm giác app dev/Grab clone và theme xanh cũ.

### 5.2 FlashX redesign

Đã chuyển direction sang:
- Navy + Electric Yellow;
- Rider map-first;
- Driver dark cockpit;
- Admin SaaS/Operations;
- CTA lớn;
- status/timeline rõ;
- loading/error/empty states;
- hạn chế false affordance.

Rider/Driver đã có các vòng analyze/test xanh trong thread sau redesign; Admin cũng đã có production build pass ở các checkpoint.

### 5.3 Tham khảo Green SM

Khách/chủ dự án yêu cầu mock có chất lượng/structure tương tự app Green SM.

Nguyên tắc áp dụng:
- học cách tổ chức home/service cards, hierarchy, khoảng trắng, rounded cards;
- **không sao chép logo/brand/asset độc quyền**;
- giữ identity FlashX riêng.

### 5.4 Bộ mock đã tạo

Đã tạo 3 nhóm mock client-facing:

1. **Customer/Rider App**
   - Home;
   - chọn Lái hộ ô tô;
   - xác nhận booking;
   - matching/tracking;
   - summary/payment;
   - history/support.

2. **FlashX Driver**
   - dashboard/cockpit;
   - offer + countdown;
   - navigation;
   - service execution workflow;
   - earnings/history;
   - KYC/profile.

3. **Admin/Operations**
   - overview KPI;
   - live jobs;
   - KYC review;
   - pricing/service configuration;
   - support/order detail.

Mock là **design target**, không đồng nghĩa mọi control trong mock đã được implement.

## 6. Business clarification quan trọng nhất của thread

Khách hàng gửi ví dụ GOCheap/DriverX/BUTL và nói rõ:

> Mô hình mong muốn là **dịch vụ tài xế lái hộ**. Khách đặt qua app; tài xế tới và **lái chính xe của khách**.

Đây là thay đổi phạm vi lớn so với Grab/Uber clone ban đầu.

### MVP mới được chốt

Chỉ tập trung:

1. **Lái hộ ô tô**
2. **Lái hộ xe máy**
3. **Đăng kiểm hộ**

Từ thời điểm này, FlashX phải được BA/UX/code theo business mới.

## 7. Đối thủ/thị trường đã tham khảo

Các tên được nhắc tới:
- GOCheap / DriverX;
- BUTL;
- THUELAI;
- một số dịch vụ lái hộ nhỏ khác tại Việt Nam.

Kết luận nghiên cứu trong thread:
- thị trường **không trắng**, đã có đối thủ và nhu cầu thật;
- chưa thấy category bị một super-app chiếm tuyệt đối giống ride-hailing truyền thống;
- cạnh tranh vẫn có cửa nhưng moat không nằm ở code alone.

Các yếu tố cạnh tranh quan trọng:
- mật độ/độ phủ tài xế;
- ETA;
- KYC/trust;
- chất lượng tài xế;
- bảo hiểm/trách nhiệm;
- quy trình bàn giao tài sản;
- CSKH/incident handling;
- partnerships với nhà hàng/bar/khách sạn/sân golf/showroom/bảo hiểm;
- pricing minh bạch.

Ý tưởng go-to-market đã thảo luận: pilot tập trung một thành phố/khu nightlife thay vì phủ toàn quốc ngay.

## 8. Gap giữa code hiện tại và business mới

### Tái sử dụng tốt

Khoảng lớn infrastructure/core có thể giữ:
- auth/session;
- Rider/Driver/Admin structure;
- Postgres/PostGIS;
- Redis GEO;
- realtime location;
- WebSocket;
- dispatch/offer/countdown;
- atomic assignment;
- history/rating;
- KYC/object storage;
- Admin operational foundation;
- Docker/CI.

### Cần refactor business domain

#### 8.1 Service type

Không dùng `bike/car` theo nghĩa Grab.

Source of truth mới đề xuất:

```text
designated_driver_car
designated_driver_bike
vehicle_inspection_assist
```

#### 8.2 Customer Vehicle

Cần thêm aggregate/entity xe thuộc khách:
- owner;
- type;
- plate;
- brand/model/color;
- transmission;
- seats;
- notes/photo.

#### 8.3 Pricing

Cần chuyển từ ride fare sang service pricing:
- base/minimum fee;
- distance;
- duration;
- waiting;
- scheduled fee;
- night/holiday surcharge;
- cancellation;
- fixed/package pricing cho đăng kiểm hộ.

#### 8.4 Scheduling

`Hẹn giờ` từng nằm ngoài scope cũ nhưng với business lái hộ/đăng kiểm hộ trở thành yêu cầu quan trọng và cần đưa lên P0/P1 gần nhất.

#### 8.5 Handover

Business mới cần khái niệm:
- nhận xe;
- bàn giao xe;
- tình trạng xe;
- bằng chứng trước/sau;
- PIN/OTP handover ở phase phù hợp.

#### 8.6 Đăng kiểm hộ

Không nên coi là trip đơn thuần. Cần job workflow riêng hoặc subtype/substate đủ rõ để biểu diễn:
- nhận xe;
- đi trung tâm;
- đang thực hiện;
- hoàn tất;
- trả xe;
- bàn giao.

## 9. Pháp lý

Trước đó đã có tài liệu/checklist pháp lý Go-Live Việt Nam cho nền tảng gọi xe/transport platform.

Sau khi business đổi sang **designated driver + vehicle inspection assistance**, phần pháp lý phải được **review lại theo đúng mô hình thực tế**, đặc biệt:
- ai là bên cung cấp dịch vụ;
- ai chịu trách nhiệm khi tài xế lái xe khách;
- hợp đồng khách–tài xế/platform;
- bảo hiểm;
- đăng kiểm hộ/ủy quyền/giấy tờ;
- xử lý dữ liệu/KYC;
- TMĐT/platform obligations.

Không được mặc định rằng toàn bộ kết luận pháp lý của Grab-like model áp dụng nguyên xi cho business mới.

## 10. Trạng thái tech hiện tại theo thread

### Có foundation tốt

- Go API.
- Flutter Rider.
- Flutter Driver.
- Next.js Admin.
- PostgreSQL/PostGIS.
- Redis.
- WebSocket.
- Dispatch.
- GPS.
- Auth/session.
- Payment cash ledger.
- Rating/history.
- KYC direct S3 upload.
- Docker Compose.
- CI.
- Tài liệu kỹ thuật tương đối đầy đủ.

### Chưa coi là production handover cuối

Cần tiếp tục:
- refactor domain theo business mới;
- customer vehicle;
- 3 service flow end-to-end;
- scheduled booking;
- pricing mới;
- đăng kiểm workflow;
- handover/trust flow;
- push production;
- SMS production;
- Maps/Places/Routes credential thật;
- release signing/build iOS/Android;
- full E2E/UAT;
- legal revalidation;
- final handover docs.

## 11. Thứ tự triển khai đề xuất từ checkpoint này

### P0.1 — khóa business contract

1. Chốt `service_type` mới.
2. Chốt `CustomerVehicle`.
3. Chốt state/workflow cho 3 dịch vụ.
4. Chốt pricing inputs.
5. Chốt immediate vs scheduled booking.

### P0.2 — backend/data migration

1. Migrations CustomerVehicle/job fields.
2. Compatibility/migrate service types cũ.
3. Pricing v2.
4. Scheduled booking.
5. Inspection-assist workflow.
6. API contract update.

### P0.3 — Customer UX

1. Home chỉ tập trung 3 service.
2. “Xe của tôi”.
3. Form Lái hộ ô tô.
4. Form Lái hộ xe máy.
5. Form Đăng kiểm hộ.
6. Confirm/matching/tracking/handover/summary.

### P0.4 — Driver UX

1. Capability/service eligibility.
2. Offer đúng domain.
3. Countdown.
4. Vehicle/customer info.
5. Receive vehicle.
6. Execute/complete/handover.
7. Inspection job states.

### P0.5 — Admin/Operations

1. Live jobs theo 3 service.
2. KYC review.
3. Customer vehicles/job detail.
4. Pricing configuration.
5. Manual support/dispatch tools cần thiết.
6. Audit.

### P0.6 — release hardening

1. Full backend tests.
2. Flutter analyze/test/device test.
3. Admin production build.
4. Postgres/Redis integration.
5. S3 E2E.
6. Maps E2E.
7. Full 3-service UAT.
8. Android/iOS release build.
9. Runbook/handover/legal sign-off.

## 12. Working agreement từ thread

- Không mở feature lan man trước khi MVP 3 dịch vụ chạy end-to-end.
- UX/UI phải đẹp đủ mức sản phẩm thương mại, không chỉ “dev UI”.
- Backend foundation đang tốt thì tái sử dụng, không rewrite vì đổi business.
- Mọi binary upload phải direct-to-storage.
- Third-party secret không đưa vào chat/docs/Git.
- Test container phải ephemeral `--rm`.
- Mọi thay đổi business lớn phải cập nhật docs trước/đồng thời với code.

## 13. Câu mô tả ngắn để mở thread mới

Có thể dùng đoạn sau khi cần tiếp tục trong thread khác:

> Tiếp tục dự án `@Code FlashX` tại `~/Documents/flashx`. Source of truth nghiệp vụ nằm ở `docs/CUSTOMER_REQUIREMENTS.md` và checkpoint thread ở `docs/THREAD_SUMMARY_2026-08-12.md`. FlashX hiện là nền tảng tài xế lái hộ/hỗ trợ phương tiện, MVP chỉ gồm Lái hộ ô tô, Lái hộ xe máy và Đăng kiểm hộ. Không tiếp tục thiết kế như Grab clone. Ưu tiên refactor domain → CustomerVehicle → pricing/scheduling → 3 flow end-to-end → UX polish → UAT/release.

## 14. Quyết định mới nhất sau checkpoint

### 14.1 Chọn Full Marketplace

Founder đã chốt **Full Marketplace**, không dùng Lean/Lite làm phương án mặc định.

Mặt khách hàng vẫn phải rất đơn giản, ít nút và tập trung 3 dịch vụ, nhưng hệ thống phía sau giữ đầy đủ khả năng tự động vận hành.

### 14.2 Bản đồ và theo dõi hành trình là bắt buộc

“App đơn giản” không có nghĩa cắt bản đồ hoặc vị trí thời gian thực.

Customer phải thấy được:
- tài xế đang ở đâu sau khi ghép thành công;
- tài xế đang di chuyển tới điểm nhận;
- thời gian dự kiến tài xế tới;
- hành trình từ điểm nhận tới điểm đến;
- vị trí tài xế/xe cập nhật trong quá trình thực hiện chuyến;
- trạng thái vẫn khôi phục đúng sau khi mất mạng/mở lại ứng dụng.

Admin/Operations cần xem được job và vị trí phù hợp để hỗ trợ vận hành.

### 14.3 Mốc tài chính đang dùng thống nhất

- Phần mềm Full Marketplace: **600 triệu VNĐ** mức dự kiến.
- Setup doanh nghiệp + Go-Live: **500 triệu VNĐ**.
- Zero → Go-Live: **1,10 tỷ VNĐ**.
- Burn rate: **~400 triệu VNĐ/tháng**.
- 6 tháng + 10% dự phòng: **~3,85 tỷ VNĐ**, làm tròn quản trị **~4 tỷ**.

**600 triệu chưa phải giá fix cứng theo hợp đồng.** Khi khóa phạm vi cuối cùng phải tận dụng tối đa những phần đã có sẵn, bóc lại từng hạng mục rồi mới chốt chi phí chính thức.

### 14.4 Cách diễn đạt khi trao đổi với Founder/Cổ đông

Ưu tiên tiếng Việt dễ hiểu, hạn chế từ kỹ thuật và tiếng Anh không cần thiết.

Cách mô tả ngắn:

> 600 triệu là mức dự kiến cho toàn bộ hệ thống gồm ứng dụng cho khách đặt dịch vụ, ứng dụng cho tài xế nhận chuyến, hệ thống quản lý dành cho công ty và toàn bộ phần phía sau để ứng dụng hoạt động ổn định như tìm tài xế gần khách, theo dõi vị trí tài xế, hiển thị hành trình chuyến đi, tính giá, quản lý hồ sơ tài xế và lưu trữ dữ liệu. Mức này chưa phải giá chốt cứng; khi chốt đầy đủ yêu cầu sẽ tận dụng tối đa những phần đã có sẵn và bóc lại chi phí chính thức.
