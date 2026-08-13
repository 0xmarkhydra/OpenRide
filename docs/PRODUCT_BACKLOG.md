# Product Backlog — FlashX MVP

> **Rebaseline 12/08/2026 — business mới đã chốt.**  
> Source of truth: [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md) và [`PRD_MVP.md`](./PRD_MVP.md).
>
> FlashX **không còn là Grab/Uber clone**. MVP chỉ gồm **Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ**. Khách sử dụng phương tiện của chính mình.

## 1. Quy tắc ưu tiên

- Không mở thêm vertical mới trước khi 3 service MVP chạy end-to-end.
- Foundation auth/realtime/dispatch/Postgres/Redis đang có phải được tái sử dụng, không rewrite chỉ vì pivot business.
- Business semantics mới ưu tiên hơn compatibility semantics `bike/car` cũ.
- UX phải phản ánh việc **tài xế nhận và lái xe của khách**, không được tiếp tục dùng copy/flow kiểu taxi.
- Binary KYC/media upload trực tiếp từ client tới object storage; backend chỉ cấp presigned URL và lưu metadata.

---

# P0 — Bắt buộc để UAT/bàn giao MVP

## P0.1 — Business/domain contract

- [ ] Chuẩn hóa `service_type`:
  - `designated_driver_car`
  - `designated_driver_bike`
  - `vehicle_inspection_assist`
- [ ] Quy định compatibility/migration cho `car/bike` cũ.
- [ ] Định nghĩa Job/Service Request thay vì giả định mọi nghiệp vụ là taxi trip.
- [ ] Chốt state machine Lái hộ.
- [ ] Chốt state/substate workflow Đăng kiểm hộ.
- [ ] Chốt rule immediate booking vs scheduled booking.

## P0.2 — Customer Vehicle / “Xe của tôi”

- [ ] Migration `customer_vehicles`.
- [ ] CRUD xe của khách.
- [ ] Ownership/authorization check.
- [ ] Trường tối thiểu:
  - type;
  - license plate;
  - brand/model;
  - color;
  - transmission;
  - seats nếu áp dụng;
  - notes;
  - ảnh đại diện optional.
- [ ] Customer chọn đúng xe khi tạo service request.
- [ ] Driver/Admin xem đúng subset thông tin cần thiết.

## P0.3 — Customer App

### Home
- [ ] Home chỉ nổi bật 3 service MVP.
- [ ] Copy chính: “Bạn cần tài xế làm gì?” hoặc tương đương.
- [ ] Không hiển thị food/delivery/taxi/thuê xe tự lái trong MVP.
- [ ] GPS/current location state.
- [ ] Empty/error/offline/location-denied states.

### Lái hộ ô tô
- [ ] Chọn “Xe của tôi”.
- [ ] Điểm nhận xe/điểm đón.
- [ ] Điểm đến.
- [ ] Ngay bây giờ / Hẹn giờ.
- [ ] Ghi chú tài xế.
- [ ] Route distance + ETA.
- [ ] Giá dự kiến.
- [ ] Confirm service request.
- [ ] Matching/searching.
- [ ] Driver assigned + ETA realtime.
- [ ] Map vẽ tuyến tài xế → điểm nhận và cập nhật vị trí tài xế realtime.
- [ ] Xác nhận nhận/bàn giao xe tối thiểu.
- [ ] In-progress tracking: route pickup → destination + vị trí + quãng đường/ETA còn lại.
- [ ] Completion/payment/rating/history.

### Lái hộ xe máy
- [ ] Flow tương đương ô tô nhưng form phương tiện/pricing/capability riêng.
- [ ] Handover phù hợp xe máy.

### Đăng kiểm hộ
- [ ] Chọn xe.
- [ ] Địa chỉ nhận xe.
- [ ] Địa chỉ trả xe nếu khác.
- [ ] Hẹn thời gian.
- [ ] Trung tâm đăng kiểm nếu khách/hệ thống chọn.
- [ ] Checklist giấy tờ cần bàn giao.
- [ ] Giá/package dự kiến.
- [ ] Timeline nhận xe → đăng kiểm → trả xe.
- [ ] Completion + evidence/status phù hợp.

## P0.4 — Driver App

### KYC/capability
- [ ] CCCD/identity.
- [ ] Selfie/portrait.
- [ ] GPLX + hạng bằng + expiry.
- [ ] Số năm kinh nghiệm.
- [ ] Lý lịch tư pháp theo policy.
- [ ] Tài khoản ngân hàng.
- [ ] Capability theo 3 service.
- [ ] Capability xe số sàn/số tự động nếu cần.
- [ ] Chỉ approved driver mới Online/nhận offer.

### Direct object storage
- [x] Backend signer/presigned architecture.
- [x] AWS SDK Go V2 signer dependency.
- [x] Backend regression sau S3/KYC integration.
- [ ] Driver UI chọn/chụp tài liệu.
- [ ] Client PUT file trực tiếp S3/VNDATA.
- [ ] Complete metadata API E2E với bucket thật.
- [ ] Admin signed GET review E2E.

### Availability/offer
- [x] Online/offline foundation.
- [x] Realtime GPS/location foundation.
- [x] Redis GEO candidate discovery foundation.
- [x] Offer TTL/countdown foundation.
- [x] Distributed/atomic accept foundation.
- [ ] Filter theo service capability mới.
- [ ] Offer hiển thị vehicle summary của khách.
- [ ] Offer hiển thị đúng flow Đăng kiểm hộ.

### Execution
- [ ] Lái hộ: Accepted → Arriving → Arrived → Vehicle Received → In Progress → Handover → Completed.
- [ ] Đăng kiểm: nhận xe → tới trung tâm → đang đăng kiểm → hoàn tất → trả xe → bàn giao.
- [ ] Một primary CTA hợp lệ theo state.
- [ ] Network/reconnect không làm mất active job.
- [ ] History/earning phản ánh đúng service type mới.

## P0.5 — Pricing v2

- [ ] Không dùng duy nhất mô hình Grab-style `base + km theo loại xe chở khách`.
- [ ] Versioned pricing rules.
- [ ] Base/minimum service fee.
- [ ] Distance component.
- [ ] Duration component.
- [ ] Waiting fee.
- [ ] Scheduled-booking fee nếu policy áp dụng.
- [ ] Night/holiday surcharge.
- [ ] Cancellation rule/fee.
- [ ] Fixed/package pricing cho Đăng kiểm hộ.
- [ ] Backend là authority cho estimate/final price.
- [ ] Admin cấu hình được; mobile không hard-code.

## P0.6 — Scheduling

- [ ] `scheduled_at`/timezone contract.
- [ ] Validation lead time.
- [ ] Scheduled job lifecycle.
- [ ] Dispatch timing trước giờ hẹn.
- [ ] Reminder/push hook.
- [ ] Customer edit/cancel policy trước giờ hẹn.
- [ ] Admin filter scheduled jobs.

## P0.7 — Handover, Evidence & Incident

### MVP bắt buộc
- [ ] Xác nhận tài xế đã tới.
- [ ] Xác nhận nhận xe.
- [ ] Vehicle/customer/job summary rõ.
- [ ] Ảnh/evidence tối thiểu khi nhận xe và trả xe.
- [ ] Dashboard/odometer evidence khi phù hợp.
- [ ] Ghi chú tình trạng bất thường.
- [ ] Lưu thời gian/vị trí/người xác nhận tại mốc bàn giao.
- [ ] Xác nhận bàn giao hoàn tất.
- [ ] Sau `VEHICLE_RECEIVED` không còn normal cancellation.
- [ ] Không reassign trực tiếp sau `VEHICLE_RECEIVED`; nếu bắt buộc phải qua controlled handover.
- [ ] Incident tối thiểu: loại sự cố + ảnh + vị trí + note + trạng thái + Operations owner.
- [ ] Incident đang mở có thể chặn flow tự động tiếp tục theo policy.
- [ ] Audit event cho các mốc/action quan trọng.
- [ ] Support path rõ ràng.
- [ ] Evidence media upload direct-to-object-storage, backend chỉ lưu metadata/signed access.

### Sau MVP nhưng data model không được chặn
- [ ] Bộ ảnh/evidence chi tiết theo loại xe.
- [ ] Structured fuel/battery/odometer.
- [ ] Damage annotation.
- [ ] OTP/PIN/signature bàn giao nâng cao.
- [ ] Claim workflow hoàn chỉnh.
- [ ] SOS/share job nâng cao.

## P0.8 — Admin / Operations

- [x] Dashboard foundation.
- [x] Driver approval foundation.
- [x] Trip/job list foundation.
- [ ] Dashboard split theo đúng 3 service.
- [ ] Live Jobs + filter service/status/driver/customer.
- [ ] Customer Vehicle detail.
- [ ] Driver capability detail.
- [ ] KYC signed document review.
- [ ] Pricing v2 configuration.
- [ ] Scheduled jobs view.
- [ ] Inspection workflow timeline.
- [ ] Support/job detail drawer.
- [ ] Manual intervention/reassign policy tối thiểu nếu pilot cần.
- [ ] Audit các action nhạy cảm.

## P0.9 — Maps / Places / Routing

- [x] GPS device location foundation.
- [x] Map provider boundary/foundation.
- [x] Routes provider boundary.
- [ ] Google Maps Demo Key hoặc dev provider được cấu hình local.
- [ ] Places/autocomplete destination search.
- [ ] Reverse geocoding nếu UX cần.
- [ ] Route/ETA thật cho 3 flow.
- [ ] Restricted production keys tách mobile/backend trước go-live.

## P0.10 — Auth/Notifications/Production adapters

- [x] OTP/auth/refresh foundation.
- [x] SMS webhook provider boundary.
- [ ] SMS provider thật trước production.
- [ ] Push provider FCM/APNs.
- [ ] Driver offer push/fallback.
- [ ] Assigned/arrived/completed/scheduled reminders.
- [ ] Production JWT/secrets/config hardening.
- [ ] Rate limiting OTP/auth/critical endpoints.

## P0.11 — Payment/history/rating

- [x] Cash-first payment ledger foundation.
- [x] Rating foundation.
- [x] History foundation.
- [ ] Map ledger/history sang service/job semantics mới.
- [ ] Đăng kiểm hộ có payment/final summary phù hợp.
- [ ] Online payment chỉ P0 nếu khách chốt provider + merchant account.

## P0.12 — QA / Release Gate

- [x] Backend Go test baseline đã xanh sau object-storage integration.
- [x] Rider analyze/test đã có checkpoint xanh.
- [x] Driver analyze/test đã có checkpoint xanh.
- [x] Admin production build đã có checkpoint xanh.
- [ ] Unit tests domain mới.
- [ ] API integration tests CustomerVehicle/scheduling/pricing v2.
- [ ] Race tests assignment/cancel/scheduled dispatch.
- [ ] S3 direct-upload E2E bucket thật.
- [ ] Maps/Places/Routes E2E.
- [ ] UAT Lái hộ ô tô trên thiết bị thật.
- [ ] UAT Lái hộ xe máy trên thiết bị thật.
- [ ] UAT Đăng kiểm hộ trên thiết bị thật.
- [ ] Android AAB signed Rider/Driver.
- [ ] iOS archive/TestFlight path.
- [ ] Staging deployment.
- [ ] Production runbook + rollback + backup/restore.
- [ ] Legal/business-model revalidation trước go-live.

---

# P1 — Nên có ngay sau MVP/pilot

- [ ] Evidence/condition capture nâng cao theo loại xe.
- [ ] PIN/OTP/signature bàn giao nâng cao.
- [ ] Claim/incident resolution hoàn chỉnh.
- [ ] Share job/SOS nâng cao.
- [ ] Promo code basic.
- [ ] Driver earning breakdown/settlement tự động hơn.
- [ ] Better manual dispatch/reassign console.
- [ ] Partner/referral QR cho nhà hàng/bar/khách sạn.
- [ ] Insurance/claim process integration.
- [ ] Customer support templates/CRM-lite.

# P2 — Scale sau khi chứng minh pilot

- [ ] Lái đi tỉnh package nâng cao.
- [ ] Thuê tài xế theo giờ/ngày.
- [ ] Enterprise/B2B accounts.
- [ ] Valet/event service.
- [ ] Loyalty/referral nâng cao.
- [ ] Wallet nếu business thật sự cần.
- [ ] Advanced fraud/risk scoring.
- [ ] Dynamic supply/demand pricing.
- [ ] Multi-city operational zoning.

---

# Delivery order hiện tại

1. Khóa business/domain contract + 3 service semantics.
2. Centralize active-state/driver-occupied rules + state machine.
3. CustomerVehicle.
4. Driver capability/KYC eligibility.
5. Scheduling.
6. Quote + Pricing v2.
7. Handover + Evidence + Incident foundation.
8. Lái hộ ô tô end-to-end.
9. Realtime route/ETA/reconnect.
10. Lái hộ xe máy end-to-end.
11. Inspection details/checklist + Đăng kiểm hộ end-to-end.
12. Marketplace financial separation: customer charge / driver earning / FlashX fee.
13. Admin Operations theo domain mới.
14. Maps/S3/SMS/Push provider production.
15. Full regression/UAT/release.

> Chưa triển khai sâu gọi xe/xe ghép trước khi 3 dịch vụ MVP qua UAT. Code/domain mới chỉ cần tránh các giả định khiến tương lai phải rewrite toàn bộ.

## Definition of Done chung

Một backlog item chỉ Done khi:
- backend contract + authorization đúng;
- UI có loading/success/error/empty/offline state nếu liên quan;
- logs/audit hợp lý;
- test tương ứng;
- docs liên quan được cập nhật;
- không tái đưa semantics Grab/Uber cũ vào business mới.
