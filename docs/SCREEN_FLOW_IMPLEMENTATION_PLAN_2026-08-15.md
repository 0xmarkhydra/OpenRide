# FlashX — Screen Flow & Implementation Plan

> **Ngày audit:** 15/08/2026  
> **Mục đích:** nguồn sự thật để Founder/Product/BA/Design/Engineering mở lại và biết ngay **mỗi màn phải có gì, phần nào đã làm, phần nào mới làm một phần, logic nào còn thiếu và thứ tự phải hoàn thiện**.  
> **Baseline production/dev đã merge:** `dev @ 30c7fd0` — `merge: vehicle custody evidence end-to-end`.  
> **Nhánh đang phát triển:** `feat/inspection-checklist`.  
> **Cập nhật triển khai:** 16/08/2026 — branch hiện tại đã mở rộng vượt checklist: route/ETA, place search, scheduler, no-show, rating, KYC, payment visibility, driver qualification và push foundation đều đã được nối source và test; chưa merge `dev` tại thời điểm cập nhật tài liệu này.

---

## 0. Cách đọc trạng thái

| Ký hiệu | Ý nghĩa |
|---|---|
| ✅ **Đã xong** | Có code nối end-to-end ở mức MVP, đã merge `dev` hoặc đã kiểm chứng ở baseline hiện tại. |
| 🟡 **Một phần** | Foundation/core đã có nhưng còn thiếu UX, automation, production edge case hoặc một mắt xích quan trọng. |
| 🚧 **Đang làm** | Đang có code trên branch hiện tại nhưng chưa hoàn tất/commit/merge. |
| 🔴 **Chưa làm** | Chưa có implementation đáng tin cậy hoặc mới chỉ có trong tài liệu. |
| ⚪ **Sau MVP** | Không phải blocker của demo 3 dịch vụ hiện tại. |

> **Nguyên tắc:** tài liệu này đánh giá theo **code thật**, không đánh dấu “xong” chỉ vì endpoint/model đã tồn tại.

---

# 1. Kết luận điều hành

FlashX hiện **không còn là Grab clone**. Nền tảng marketplace 3 dịch vụ đã hình thành khá rõ:

1. `designated_driver_car` — Lái hộ ô tô.
2. `designated_driver_bike` — Lái hộ xe máy.
3. `vehicle_inspection_assist` — Đăng kiểm hộ.

Phần mạnh nhất hiện tại:

- OTP/auth/session.
- CustomerVehicle.
- 3 service semantics.
- Immediate + scheduled booking foundation.
- Pricing versioned.
- Redis GEO/driver matching/offer/locking.
- GPS + WebSocket + reconnect/backoff.
- Job lifecycle cho lái hộ và đăng kiểm.
- Custody Evidence nhận/trả xe hai bên.
- Incident foundation.
- Admin Operations khá đầy đủ.
- Admin live map.
- RBAC/audit/settings.
- Landing mới.

Ở checkpoint **16/08/2026**, phần lớn blocker sản phẩm đã được xử lý trên branch hiện tại và đã qua test kỹ thuật:

1. ✅ Checklist giấy tờ Đăng kiểm hộ end-to-end, có template version + snapshot theo job + gate nhận xe.
2. ✅ Mobile map có route/polyline + ETA, refresh có throttle và fallback khi provider ngoài lỗi.
3. ✅ Chọn điểm đến có 3 đường: tìm địa chỉ bằng chữ, shortcut demo và ghim pin trên bản đồ; cùng dùng một tọa độ chuẩn cho estimate/booking.
4. ✅ Scheduled booking có worker kích hoạt độc lập, không phụ thuộc heartbeat tài xế.
5. 🟡 Push notification: device registration + durable outbox + event hook + worker đã xong; **gửi push thật** còn phụ thuộc FCM/APNs credential và mobile SDK.
6. ✅ Pickup grace/no-show đã có policy server-side + countdown Driver + cấu hình Admin. **Waiting fee/cancellation fee bằng tiền chưa tự tính** vì chưa có chính sách tài chính được Founder chốt.
7. 🟡 Pricing core/versioned đã ổn; waiting/night/holiday/cancellation surcharge chưa bật vì công thức chưa được chốt.
8. ✅ Rider rating UI + backend rating đã nối.
9. ✅ Payment cash-first có ledger/reconciliation + trạng thái Rider/Driver/Admin. Payout/commission/refund/invoice vẫn tách riêng, không tự suy diễn.
10. ✅ Driver capability có Admin control, GPLX class/expiry và quyền lái số sàn; dispatch lọc lại cả lúc offer/accept/manual assign.
11. ✅ Driver KYC direct-upload onboarding + trạng thái review đã nối, Admin review giữ nguyên presigned-object-storage flow.
12. 🟡 Mobile đã chuyển token màu canonical xanh-trắng và nhiều màn chính đã cập nhật; vẫn cần một vòng visual polish cuối trên thiết bị thật trước release store.
13. 🔴 Pháp lý/bảo hiểm/ủy quyền đăng kiểm vẫn cần review cuối trước Go-Live; đây không phải thứ nên tự suy diễn bằng code.

---

# 2. Kiến trúc flow tổng thể cần giữ

```text
CUSTOMER
  ↓
Chọn dịch vụ
  ↓
Chọn xe của khách
  ↓
Pickup / Destination / Schedule
  ↓
Estimate / Pricing Version
  ↓
Create Job
  ↓
Dispatch / Offer / Matching
  ↓
DRIVER ACCEPT
  ↓
Driver tới điểm nhận
  ↓
[Inspection: checklist giấy tờ]
  ↓
Custody Evidence nhận xe
  ↓
Tài xế + Khách xác nhận
  ↓
VEHICLE_RECEIVED
  ↓
Service-specific workflow
  ↓
Custody Evidence trả xe
  ↓
Tài xế + Khách xác nhận
  ↓
HANDOVER
  ↓
COMPLETED
  ↓
Payment / Rating / History
```

Phần dùng chung phải được giữ reusable:

```text
Auth
Customer Vehicle
Booking
Pricing
Scheduling
Dispatch
Driver Capability
Realtime/GPS
Custody
Incident
Payment
Rating
Audit

        ↓
 Service Workflow
        ↓
Car / Bike / Inspection / future services
```

Không nhét Đăng kiểm hộ thành “taxi trip”.

---

# 3. CUSTOMER APP — Màn hình và trạng thái implementation

## C01 — App bootstrap / Splash / Session Recovery

**Mục tiêu:**

- Hiện branding FlashX khi khởi động.
- Restore access/refresh session.
- Nếu đang có job → quay lại đúng active job.
- Nếu chưa login → OTP.
- Load operational config/service availability.

**Hiện tại:** 🟡 **Một phần**

Đã có:

- `AuthController.initialize()`.
- Loading state khi app khởi động.
- Restore auth/session.
- `loadHistory()` chọn job chưa `completed/cancelled` làm `activeTrip`.
- Realtime reconnect/backoff.

Còn thiếu:

- Splash/launch experience đúng design system.
- Hiển thị network/recovery status rõ ràng khi restore thất bại.
- Operational settings/service enabled state chưa được Customer App fetch để disable service trước khi bấm.
- App lifecycle resume/background recovery cần test kỹ trên iOS/Android thật.

---

## C02 — Login OTP

**Mục tiêu:** số điện thoại → OTP → session.

**Trạng thái:** ✅ **Đã xong ở mức MVP**

Đã có:

- Login OTP.
- Access/refresh session foundation.
- Development OTP support.
- Logout.

Cần hardening trước production:

- Rate-limit UX / resend countdown.
- Anti-abuse/OTP provider production.
- Terms/privacy acceptance version.

---

## C03 — Home / 3 dịch vụ

**Mục tiêu:** chỉ nổi bật 3 dịch vụ hiện tại.

**Trạng thái:** ✅ **Đã có chức năng**, 🟡 **UI còn cần migrate**

Đã có 3 service card:

- Lái hộ ô tô.
- Lái hộ xe máy.
- Đăng kiểm hộ.

Đã có active-job pill nếu đang chạy.

Gap:

- Mobile Rider còn theme navy/vàng cũ; chưa đồng bộ hoàn toàn design system iOS-inspired trắng/xanh.
- Chưa có service availability từ Admin settings ở Home.
- Chưa có scheduled jobs/upcoming section rõ ràng ngoài active state.

---

## C04 — “Xe của tôi” / Vehicle list

**Trạng thái:** 🟡 **Một phần**

Đã có:

- `CustomerVehicle` backend.
- List vehicles.
- Chọn vehicle theo service.
- Car/motorbike compatibility.
- Account tab hiển thị xe.

Còn thiếu:

- Edit vehicle.
- Delete/archive vehicle.
- Set default vehicle.
- Vehicle photo.
- Year/seats/notes chưa được khai thác đầy đủ trong UI.
- Ownership/plate duplicate verification policy.

---

## C05 — Thêm xe

**Trạng thái:** ✅ **MVP cơ bản**

Đã nhập:

- type.
- license plate.
- brand.
- model.
- color.
- transmission với ô tô.

Gap:

- Validate format biển số tốt hơn.
- Year/seats/notes/photo.
- Duplicate vehicle UX.
- Manual/automatic phải được dùng thực sự vào matching driver capability — hiện chưa đủ domain.

---

## C06 — Chọn điểm nhận

**Trạng thái:** 🟡 **Demo được, production chưa đủ**

Đã có:

- GPS hiện tại qua `Geolocator`.
- Thanh Hóa fallback khi demo/permission fail.
- Map pickup marker.

Còn thiếu P0:

- Search địa chỉ/Google Places.
- Kéo pin/map picker.
- Nhập địa chỉ khác GPS hiện tại.
- Pickup note: cổng nào/tầng nào/quán nào.
- Saved address: Nhà/Công ty.
- Geocoding display address.

---

## C07 — Chọn điểm đến / nơi đăng kiểm

**Trạng thái:** 🟡 **Demo only**

Hiện destination đang hard-code:

- Quảng trường Lam Sơn.
- Vincom Plaza Thanh Hóa.
- Sầm Sơn.
- Trung tâm đăng kiểm Thanh Hóa.

Cần làm:

- Places/search/pin thật.
- Recent/saved destinations.
- Với Đăng kiểm: danh sách trung tâm phù hợp hoặc Operations quyết định trung tâm.
- Không hard-code một trung tâm trong production.

---

## C08 — Đặt ngay / Hẹn giờ

**Trạng thái:** 🟡 **Một phần**

Đã có:

- UI immediate/scheduled.
- DatePicker/TimePicker.
- `scheduled_at` persistence.
- `scheduled` job status.
- `ActivateDueScheduled()` backend.

**Production gap quan trọng:**

- Job due hiện được activate khi có driver heartbeat/availability path gọi `ActivateDueScheduled()`.
- Chưa có scheduler/worker độc lập chạy theo clock.
- Nếu không có driver heartbeat đúng thời điểm, scheduled job có thể không được kích hoạt đúng SLA.

Cần:

- Dedicated scheduler/worker hoặc periodic durable job.
- Pre-dispatch trước giờ hẹn X phút.
- Reminder cho khách/tài xế.
- Late driver / no candidate escalation.
- Scheduled cancellation window.

---

## C09 — Estimate / Báo giá

**Trạng thái:** ✅ **Core pricing/versioning đã xong**, 🟡 **Business pricing chưa đầy đủ**

Đã có:

- Route abstraction.
- Distance/duration estimate.
- Base fare.
- Per-km.
- Service fee.
- Minimum fare.
- Pricing version.
- Admin edit/version persistence.
- Inspection service package base (`299k` default demo) + distance.

Còn thiếu:

- Waiting fee.
- Night surcharge.
- Holiday/peak surcharge.
- Scheduled booking fee.
- Cancellation fee.
- Extra inspection expenses approval.
- Price ceiling/rounding policy.
- Fare finalization logic nếu thực tế khác estimate.

---

## C10 — Confirm booking

**Trạng thái:** ✅ **MVP**

Đã có:

- Estimate preview.
- Vehicle/service/destination.
- Distance/duration.
- Cash method display.
- Confirm booking.
- Idempotency key.

Gap:

- Pickup address text/note.
- Cancellation policy disclosure.
- Detailed price components.
- Legal confirmation/terms for handing over vehicle.

---

## C11 — Searching / Matching

**Trạng thái:** ✅ **Backend mạnh**, 🟡 **Customer UX còn cơ bản**

Đã có:

- Redis GEO / spatial candidate lookup.
- Radius/freshness config.
- Capability filter.
- Offer timeout.
- Atomic accept / chống double assignment.
- Offer retry foundation.
- Customer state `searching`.

Còn thiếu:

- Searching animation/estimated search time rõ hơn.
- Expand radius strategy UX.
- “Không tìm được tài xế” fallback/Operations escalation.
- Scheduled job supply reservation strategy.

---

## C12 — Driver matched / Tài xế đang tới

**Trạng thái:** 🟡 **Một phần**

Đã có:

- Driver assigned ID.
- Driver GPS realtime.
- Map driver marker.
- Status timeline.

Còn thiếu P0/P1:

- Tên/avatar/rating driver đầy đủ trong Rider UI.
- ETA tài xế → pickup.
- Route/polyline tài xế → pickup.
- Camera fit/recenter.
- Call/chat.
- Phone masking/privacy.
- Driver late/no-show UX.

---

## C13 — Customer Live Map / Theo dõi dịch vụ

**Trạng thái:** 🟡 **Một phần**

Đã có:

- GoogleMap.
- Pickup marker.
- Destination marker.
- Driver marker.
- WebSocket driver location.
- Reconnect/backoff.

Chưa có:

- Polyline route.
- ETA realtime.
- Remaining distance/time.
- Driver→pickup route riêng và service route riêng.
- Smooth marker interpolation.
- Auto camera tracking.
- Explicit stale-location indicator.

Đây là **P0 UX** vì Founder đã yêu cầu map/ETA/realtime rõ ràng.

---

## C14 — Custody Evidence: nhận xe

**Trạng thái:** ✅ **Đã xong end-to-end**

Đã có:

- Driver nhập condition note.
- Odometer.
- Fuel/battery optional.
- ≥2 ảnh.
- Presigned upload trực tiếp object storage.
- Driver confirm.
- Customer review evidence.
- Customer confirm.
- Sửa metadata/thêm ảnh → reset confirmations.
- Backend production gate `VEHICLE_RECEIVED` khi storage/enforcement bật.
- Demo fallback có nhãn rõ nếu storage unavailable.

Còn cần production policy:

- Retention period.
- Data privacy/access retention.
- Export evidence khi tranh chấp.
- Quy định ảnh bắt buộc theo góc/loại xe.

---

## C15 — Active service / In-progress

**Trạng thái:** ✅ **State workflow**, 🟡 **tracking UX partial**

Lái hộ:

- `VEHICLE_RECEIVED → IN_PROGRESS → HANDOVER → COMPLETED`.

Đăng kiểm có state riêng.

Gap:

- Route/ETA.
- Customer support/contact action.
- Customer incident/report issue action chưa có trong Rider UI.
- Background push when state changes.

---

## C16 — Custody Evidence: trả xe

**Trạng thái:** ✅ **Đã xong end-to-end**

Giống pickup evidence và production handover gate.

Gap pháp lý/vận hành:

- Người nhận thay mặt/ủy quyền.
- Khách không có mặt khi trả xe.
- Timeout chờ customer confirm.
- Dispute state nếu customer không đồng ý condition.

---

## C17 — Completion / Thanh toán

**Trạng thái:** 🟡 **Một phần**

Đã có:

- Trip complete.
- Fare stored.
- Payment domain/model cash-first foundation.
- UI demo hiển thị Tiền mặt.

Chưa đủ production:

- Customer xác nhận cash paid / driver cash received.
- QR/card/e-wallet.
- Payment retry/failure.
- Refund.
- Invoice/receipt.
- Cancellation charge.
- Extra cost approval.
- Driver payout/commission settlement.

---

## C18 — Rating

**Trạng thái:** 🟡 **Backend xong, UI chưa nối**

Đã có:

- Rating backend/service.
- Validation completed trip + assigned driver.
- `RiderTripController.rateTrip()`.

Nhưng `rateTrip()` hiện **không được gọi từ Rider UI**.

Cần:

- Rating bottom sheet sau completion.
- 1–5 stars.
- Comment.
- Prevent duplicate UX.
- Driver average rating display.

---

## C19 — Hoạt động / History

**Trạng thái:** 🟡 **List xong, detail thiếu**

Đã có:

- History list.
- Service/status/fare.
- Inspection result text.

Cần:

- History detail screen.
- Full timeline.
- Customer/driver/vehicle data.
- Custody evidence.
- Inspection checklist/result.
- Payment/receipt.
- Rating.
- Incident history.

---

## C20 — Account / Support

**Trạng thái:** 🟡 **Basic**

Đã có:

- Phone/profile shell.
- Vehicle list.
- Cash demo info.
- Logout.

Còn thiếu:

- Edit profile.
- Notifications settings.
- Saved addresses.
- Legal/privacy/terms screens.
- Support center.
- Emergency contact.
- Delete account/data request.

---

# 4. DRIVER APP — Màn hình và trạng thái

## D01 — Driver bootstrap/Login OTP

**Trạng thái:** ✅ **MVP**

- Auth/session.
- OTP.
- Restore profile.
- Home if authenticated.

Gap: branded splash + recovery/network UX.

---

## D02 — Driver onboarding/KYC upload

**Trạng thái:** 🟡 **Backend/client có, màn onboarding chưa nối hoàn chỉnh**

Đã có backend:

- presigned upload URL.
- direct upload.
- complete metadata.
- list docs.
- signed view URL.
- Admin approve/reject.

Có `DriverDocumentClient`, nhưng audit hiện tại không thấy flow UI Driver dùng client này thành màn onboarding upload hoàn chỉnh.

Cần:

- KYC checklist screen.
- CCCD front/back.
- GPLX.
- Portrait.
- Expiry date/class metadata.
- Re-upload rejected document.
- Review status/reason.

---

## D03 — Home / Online-Offline

**Trạng thái:** ✅ **Core**

Đã có:

- Pending/approved gate.
- Online/offline switch.
- Busy protection.
- GPS permission.
- Location stream.
- Demo location fallback.

Gap:

- Shift/break mode.
- Service zone preference.
- Background location production validation on iOS/Android.
- Battery/network health warning.

---

## D04 — Job Offer

**Trạng thái:** ✅ **MVP**

Hiển thị:

- Service.
- Customer vehicle.
- Transmission display.
- Distance to pickup.
- Estimated service value.
- Offer expiry countdown.
- Accept/reject.

Backend:

- lock/atomic accept.
- capability + online + fresh location filter.

Gap:

- Destination/return logistics presentation đầy đủ.
- ETA tới pickup.
- Driver payout thực thay vì customer fare/service value.
- Scheduled offer semantics.

---

## D05 — Navigation tới khách

**Trạng thái:** 🟡 **State + markers, navigation thiếu**

Đã có:

- Arriving/arrived actions.
- Driver GPS marker.
- Pickup/destination markers.

Thiếu:

- Route/polyline.
- ETA.
- Turn-by-turn/deep link Google Maps.
- Arrival radius validation/geofence.
- Wrong-location correction.

---

## D06 — Custody Evidence nhận xe

**Trạng thái:** ✅ **End-to-end**

Đã có:

- Condition note.
- odo/fuel/battery.
- Camera/gallery.
- image picker.
- direct S3 upload.
- ≥2 photos.
- Driver confirm.
- Wait customer confirm.
- `ready` gate.
- HEIC/HEIF handling/fallback.

Gap: production photo policy, dispute workflow.

---

## D07 — Vehicle Received / Start service

**Trạng thái:** ✅ **State machine**

- Backend `allowed_actions` drives UI.
- Custody production gate.
- Incident can block flow.

Gap:

- Customer no-show/wait policy.
- Driver cannot safely abort/return vehicle under exceptional circumstances with explicit workflow.

---

## D08 — Lái hộ in-progress

**Trạng thái:** ✅ **Core workflow**, 🟡 **map/navigation partial**

- Start.
- In-progress.
- Handover.
- Complete.
- Incident.

Thiếu route/navigation/ETA và special exceptions.

---

## D09 — Đăng kiểm progress

**Trạng thái:** ✅ **State workflow**

Đã có actions:

- tới pickup.
- nhận xe.
- tới nơi đăng kiểm.
- bắt đầu đăng kiểm.
- complete inspection.
- result: passed/failed/deferred/unavailable.
- returning vehicle.
- arrived return.
- handover.
- complete.

Còn thiếu:

- Checklist giấy tờ end-to-end — 🚧.
- Trung tâm đăng kiểm thực tế.
- Queue/appointment.
- Trung tâm đóng cửa/quá tải.
- Phí phát sinh & customer approval.
- Receipt/photo result.
- Retry/remediation flow nếu failed/deferred.

---

## D10 — Inspection Checklist

**Trạng thái:** 🚧 **Đang làm trên `feat/inspection-checklist`**

Đã code nhưng chưa commit/merge:

- migration `010_inspection_checklist.sql`.
- versioned template domain.
- snapshot per job.
- customer status.
- driver verification status.
- required/optional rules.
- customer change resets driver verification.
- readiness logic.
- memory/Postgres store.
- service tests đang được hoàn thiện.

Chưa có:

- API routes.
- Runtime bootstrap/dependency wiring.
- Customer App UI.
- Driver App UI.
- Admin template manager UI.
- Admin job checklist viewer.
- State gate production.
- HTTP/end-to-end tests.

Known current issue:

- Test mới gọi nhầm `MarkArrivingForPickup()` / `MarkArrivedForPickup()`; trip service thật dùng `MarkArriving()` / `MarkArrived()`.

---

## D11 — Custody trả xe / handover

**Trạng thái:** ✅ **End-to-end**

Gap:

- Authorized receiver.
- Customer absent.
- Refusal/dispute.
- Waiting charge.

---

## D12 — Báo sự cố

**Trạng thái:** ✅ **MVP basic**, 🟡 **Operations workflow còn đơn giản**

Driver có:

- vehicle issue.
- accident.
- document issue.
- customer unreachable.
- other.
- note.

Admin có incident open/close.

Thiếu:

- Severity.
- Ownership/assignee.
- SLA.
- Escalation.
- Evidence attachment.
- Emergency action.
- Insurance claim workflow.

---

## D13 — Giá trị dịch vụ / Payout

**Trạng thái:** 🟡 **Demo only**

UI cố ý ghi rõ **không phải thu nhập ròng**.

Chưa có:

- Driver payout calculation.
- Commission ledger.
- Wallet balance.
- Cash reconciliation.
- Settlement cycle.
- Withdraw/bank account.
- Tax/withholding.

---

## D14 — Driver history

**Trạng thái:** 🟡 **List only**

Cần detail screen + payout + evidence/incident/history.

---

## D15 — Driver account/capabilities

**Trạng thái:** 🟡 **Một phần**

Đã có:

- service capabilities array.
- approval status.
- full name.

**Domain còn thiếu quan trọng:**

- GPLX class.
- GPLX expiry.
- manual/automatic competency.
- vehicle-type restrictions.
- inspection authorization/training.
- capability expiry/revalidation.

Hiện capability matching chủ yếu theo **service type**, chưa đủ để khẳng định driver phù hợp từng chiếc xe thực tế.

---

# 5. ADMIN WEB — Màn hình và trạng thái

## A01 — Admin OTP Login

**Trạng thái:** ✅

- Phone OTP.
- Dev OTP only in development.
- session/refresh/logout.
- RBAC role preserved.

---

## A02 — Operations Dashboard

**Trạng thái:** ✅ **MVP mạnh**

- active jobs.
- drivers online.
- incidents.
- pending approvals.
- priority queue.
- auto refresh.

P1:

- richer KPI/time-series.
- service SLA.
- supply-demand metrics.

---

## A03 — Live Operations Map

**Trạng thái:** ✅ **Live marker map**, 🟡 **route/ETA thiếu**

- MapLibre.
- OpenFreeMap.
- online/busy driver markers.
- active job pickup markers.
- incident markers.
- freshness filter.
- marker click opens Driver/Job.

Thiếu:

- Route lines.
- ETA.
- heatmap.
- zone/supply visualization.

---

## A04 — Jobs List

**Trạng thái:** ✅

- search.
- service filter.
- status filter.
- incident filter.
- sort.
- pagination 20/50/100.
- desktop/mobile layouts.

---

## A05 — Job Detail

**Trạng thái:** ✅ **Core**, 🟡 **inspection checklist pending**

Đã có:

- customer.
- driver.
- vehicle.
- price.
- pickup/destination.
- timeline.
- incident.
- inspection result.
- custody pickup/return evidence.

Còn thiếu:

- inspection checklist viewer 🚧.
- payment/settlement detail.
- event timeline/audit per job đầy đủ hơn.

---

## A06 — Manual Dispatch / Reassign

**Trạng thái:** ✅

- eligible candidates.
- distance.
- assign.
- reassign.
- reason required.
- block after `VEHICLE_RECEIVED`.
- invalidate old offer.
- audit/realtime.

P1:

- candidate ETA/route instead of straight distance only.
- reason codes.

---

## A07 — Drivers

**Trạng thái:** ✅ **Operations core**

- list.
- online/offline/busy.
- capabilities.
- approval.
- last location.
- approve/reject/suspend.
- reason + audit.

Gap: license fields/capability detail model.

---

## A08 — KYC Review

**Trạng thái:** ✅

- signed URL.
- view document.
- approve/reject individual docs.
- notes.
- direct upload architecture preserved.

Gap: expiry alerts/revalidation workflow.

---

## A09 — Customers

**Trạng thái:** ✅ **Read operations view**

- list.
- phone/status.
- vehicle count.
- job count/history.

P1:

- account suspend/unsuspend.
- privacy/data request.
- support notes.

---

## A10 — Customer Vehicles

**Trạng thái:** ✅ **Read operations view**

- owner.
- plate.
- type.
- brand/model.
- transmission.
- history.

P1: admin corrections/verification flags.

---

## A11 — Pricing

**Trạng thái:** ✅ **Versioned editable core**

- 3 services.
- base/per-km/service/minimum.
- version persistence.
- Super Admin edit.
- Operations read-only.
- audit.
- old job fare unaffected.

Missing business components:

- waiting.
- scheduled.
- night.
- holiday/peak.
- cancel fee.
- inspection extra costs.

---

## A12 — Incidents

**Trạng thái:** ✅ **Basic operations**

- incident list.
- job context.
- notes.
- close incident.
- resolution note.
- audit.

Need P1 incident management:

- severity.
- assignee.
- SLA.
- escalation.
- attachments.
- reopen.
- insurance/legal flag.

---

## A13 — Inspection Checklist Template

**Trạng thái:** 🚧 **Backend domain in progress; Admin UI missing**

Target:

- active template version.
- item key/label.
- required/optional.
- sort order.
- activate/publish new version.
- reason for change.
- never mutate old job snapshots.

---

## A14 — Audit Log

**Trạng thái:** ✅

- actor.
- action.
- resource.
- timestamp.
- key admin operations audited.

P1: richer before/after diff & filters/export.

---

## A15 — Admin Accounts / RBAC

**Trạng thái:** ✅

Roles:

- `super_admin`.
- `operations`.

- OTP-only account.
- create/update/enable/disable.
- reason.
- cannot remove last active Super Admin.
- backend permission checks, not only hidden UI.

---

## A16 — Operational Settings

**Trạng thái:** ✅

Super Admin:

- enable/disable each service.
- dispatch radius.
- max location age.
- reason/version/audit.

Operations: read-only.

Intentional restriction:

- secrets/DB/Redis/API keys are not editable from UI.

---

## A17 — Finance / Payout / Refund Console

**Trạng thái:** 🔴 **Chưa làm**

Cần trước scale production:

- payments.
- cash reconciliation.
- driver payout.
- FlashX commission.
- refunds.
- disputes.
- settlement batches.
- invoice/reference.

---

## A18 — Notification Center

**Trạng thái:** 🔴 **Chưa làm**

- push history.
- failed notification.
- scheduled reminder.
- incident alert.
- broadcast/operations alert.

---

# 6. LANDING WEB

## L01 — Landing page

**Trạng thái:** ✅ **Demo/brand ready**

Đã có:

- iOS-inspired green/white direction.
- 3 services.
- product truth.
- process sections.
- custody/trust messaging.
- inspection section.
- Thanh Hóa pilot.
- responsive static deployment.

Còn cần trước public launch:

- real App Store/Google Play links.
- legal links.
- privacy/terms.
- support/contact.
- company/legal entity info.
- analytics/consent if needed.
- SEO/OpenGraph assets.

---

# 7. CROSS-CUTTING LOGIC — Status Matrix

| Domain | Status | Nhận xét |
|---|---|---|
| OTP/Auth/session | ✅ | Core dùng được. |
| 3 service semantics | ✅ | Không còn Grab `car/bike` làm domain chính. |
| CustomerVehicle | ✅ | Backend + create/list UI; CRUD UX chưa đủ. |
| Immediate booking | ✅ | Có estimate/create/matching. |
| Scheduled booking | 🟡 | Persist + due activation có; thiếu scheduler độc lập/SLA. |
| Pricing version | ✅ | Base/km/service/minimum editable. |
| Pricing surcharge v2 | 🔴 | Chờ/đêm/lễ/hủy/hẹn giờ chưa có. |
| Dispatch GEO | ✅ | Capability + online + fresh location. |
| Offer timeout/atomic accept | ✅ | Foundation tốt. |
| Manual dispatch/reassign | ✅ | Safe gate trước nhận xe. |
| Driver service capabilities | ✅ | Service-level. |
| Driver license/transmission capability | 🔴 | GPLX class/expiry/manual-auto chưa thành domain đầy đủ. |
| Customer GPS | ✅ | Có GPS + fallback. |
| Driver GPS | ✅ | Có streaming. |
| WebSocket realtime | ✅ | Có. |
| Mobile reconnect/backoff | ✅ | Có exponential capped 15s. |
| App active-job restore | ✅ | Từ history/non-terminal job. |
| Push notification | 🔴 | Không tìm thấy FCM/APNs integration. |
| Mobile map markers | ✅ | Có. |
| Route/polyline/ETA | 🔴 | Mobile chưa có. |
| Admin map | ✅ | Live marker MapLibre. |
| Custody Evidence | ✅ | Strong end-to-end. |
| Inspection lifecycle | ✅ | States/result/return flow có. |
| Inspection document checklist | 🚧 | Domain đang làm, chưa end-to-end. |
| Incident | 🟡 | Report + Admin close; thiếu escalation/SLA. |
| Cash payment foundation | 🟡 | Model/core có, operational settlement thiếu. |
| Driver payout/commission | 🔴 | Chưa có. |
| Rating backend | ✅ | Có. |
| Rating Rider UI | 🔴 | Controller có method nhưng UI chưa gọi. |
| History list | ✅ | Rider + Driver có list. |
| History detail | 🔴 | Chưa có mobile detail. |
| Admin RBAC | ✅ | Backend enforced. |
| Audit | ✅ | Có. |
| Object storage direct upload | ✅ | KYC/Custody architecture đúng. |
| Driver KYC upload UI | 🟡 | Backend/client có, full onboarding screen chưa nối. |
| Customer call/chat | 🔴 | Chưa có. |
| Cancellation policy/fees | 🔴 | State cancel cơ bản có, business fee/no-show/waiting thiếu. |
| Legal/insurance/authorization | 🟡 | Docs cảnh báo cần review; chưa đủ để Go-Live production. |

---

# 8. BUSINESS LOGIC CÒN THIẾU — Các case bắt buộc phải giải quyết

## 8.1 Location / Map

1. Khách muốn pickup khác vị trí hiện tại.
2. GPS sai/chưa bật.
3. Địa chỉ không rõ cổng/tòa/tầng.
4. Tài xế đến sai phía đường.
5. ETA tăng do giao thông.
6. Location tài xế stale.
7. Driver đi lệch route bất thường.

**Cần:** Places + pin + address note + route + ETA + stale warning.

---

## 8.2 Matching / Supply

1. Không có tài xế trong bán kính.
2. Nhiều tài xế reject.
3. Offer hết hạn liên tục.
4. Scheduled job tới giờ vẫn chưa có tài xế.
5. Driver accept rồi mất mạng.
6. Driver accept rồi cancel/no-show.
7. Admin cần manual reassign.

Manual reassign đã có; các SLA/escalation còn thiếu.

---

## 8.3 Waiting / No-show / Cancellation

Phải chốt và code:

- Free waiting minutes khi driver arrived.
- Waiting fee/minute.
- Customer no-show.
- Driver no-show.
- Rider cancel trước match.
- Rider cancel sau accept.
- Rider cancel sau arrived.
- Scheduled cancellation window.
- Cancellation fee destination.
- Driver compensation.
- FlashX commission/refund.

Hiện cancel state có nhưng **chưa phải cancellation product hoàn chỉnh**.

---

## 8.4 Custody / Dispute

Đã có evidence tốt, nhưng cần quyết định:

- Khách không xác nhận trong X phút.
- Khách từ chối evidence.
- Hai bên tranh luận vết xước.
- Người nhận xe khác chủ tài khoản.
- Customer unreachable khi trả xe.
- Vehicle cannot be safely returned.
- Evidence retention/export.

---

## 8.5 Lái hộ ô tô

Còn thiếu logic:

- GPLX class hợp lệ.
- GPLX expiry.
- Manual/automatic competency.
- Xe đặc biệt/xe điện/xe lớn.
- Fuel/charging issue giữa chuyến.
- Parking/toll/extra expense approval.
- Accident/insurance process.

---

## 8.6 Lái hộ xe máy

BA hiện chốt default:

> Tài xế lái chính xe máy của khách và khách đi cùng.

Cần làm rõ/code:

- Helmet responsibility.
- Có cho thêm passenger/luggage không.
- Vehicle load/safety suitability.
- Trẻ em/pregnancy/high-risk passenger policy.
- Rain/extreme weather cancellation/safety.
- Driver license suitability.
- Custody evidence tối giản phù hợp xe máy.

---

## 8.7 Đăng kiểm hộ

Đây là nhóm gap lớn nhất hiện tại.

Phải hoàn tất:

1. Checklist giấy tờ versioned.
2. Customer declaration.
3. Driver verification.
4. Required document gate.
5. Admin template management.
6. Admin per-job checklist viewer.
7. Trung tâm đăng kiểm selection/assignment.
8. Authorization/ủy quyền rule.
9. Trung tâm từ chối nhận xe.
10. Trung tâm quá tải/đóng cửa.
11. Phát sinh chi phí phải xin customer approval.
12. Receipt/evidence/result attachment.
13. Failed inspection remediation.
14. Retry/reschedule.
15. Document return/lost document incident.
16. Kết quả `failed` không được hiểu thành FlashX job failed.

---

## 8.8 Payment / Money

MVP demo cash có thể chạy, nhưng production cần:

- payment state displayed to customer/driver/admin.
- cash collected confirmation.
- commission calculation.
- driver payout ledger.
- refund.
- cancellation charge.
- extra expense.
- settlement.
- invoice/receipt.
- reconciliation report.

---

## 8.9 Notifications

Không nên phụ thuộc user mở app.

Cần:

- Driver new offer push.
- Customer driver matched push.
- Driver arriving/arrived.
- Scheduled reminder.
- Custody waiting customer confirmation.
- Inspection status/result.
- Return/handover.
- Incident/Operations alert.

---

# 9. Thứ tự hoàn thiện được khuyến nghị

## P0-A — Đóng đủ 3 dịch vụ demo thật

1. **Hoàn tất Inspection Checklist end-to-end** 🚧
   - fix service tests.
   - bootstrap store/service.
   - HTTP API.
   - Customer declaration UI.
   - Driver verification UI.
   - Admin template + per-job view.
   - required gate.
   - tests.
   - merge `dev`.

2. **Address search/pin**
   - pickup.
   - destination.
   - inspection center input.

3. **Route + ETA mobile**
   - driver → pickup.
   - service journey.
   - remaining distance/time.

4. **Dedicated scheduled-job worker**
   - không phụ thuộc driver heartbeat.

5. **Rating UI**
   - backend đã sẵn.

6. **Customer incident/support action**.

7. **Driver KYC onboarding/upload UI**.

## P0-B — Trước pilot người dùng thật

8. Cancellation/waiting/no-show.
9. Push notification/background behavior.
10. Driver GPLX/transmission capability model.
11. Payment cash reconciliation + commission/payout minimum viable ledger.
12. Inspection unexpected expense/receipt/retry flow.
13. Legal/insurance/ủy quyền review.
14. Production object storage + retention policy.

## P1 — Sau khi pilot ổn

15. Call/chat + phone masking.
16. Vehicle edit/delete/photos/default.
17. Full mobile history detail.
18. Incident severity/SLA/escalation.
19. Admin finance/payout/refund.
20. Analytics/report/export.
21. Driver shift/zone/preferences.
22. Better map interpolation/heatmap.

## P2 — Expansion platform

- Ride-hailing bằng xe tài xế.
- Xe ghép/carpool.
- Driver by hour.
- Bảo dưỡng hộ.
- Giao/nhận xe.
- Cứu hộ.
- Fleet/partner companies.

Chỉ mở P2 sau khi `Job + ServiceWorkflow` 3 MVP ổn định.

---

# 10. Definition of Done theo từng dịch vụ

## Lái hộ ô tô được coi là “pilot-ready” khi

- [ ] Customer chọn pickup/destination thật, không hard-code.
- [ ] Customer chọn đúng xe.
- [ ] Estimate đúng pricing version.
- [ ] Matching đúng capability + GPLX/transmission.
- [ ] Driver nhận offer.
- [ ] Route/ETA driver→customer hiển thị.
- [ ] Driver arrived.
- [ ] Custody pickup evidence đủ 2 bên.
- [ ] Vehicle received.
- [ ] Journey route/tracking hiển thị.
- [ ] Incident path test.
- [ ] Return custody đủ 2 bên.
- [ ] Payment/cash reconcile.
- [ ] Rating UI.
- [ ] App kill/reopen vẫn phục hồi job.
- [ ] Scheduled flow test nếu đặt trước.

## Lái hộ xe máy được coi là “pilot-ready” khi

Ngoài các điểm chung:

- [ ] Chốt helmet/passenger/luggage policy.
- [ ] Driver license suitability.
- [ ] Custody flow phù hợp xe máy.
- [ ] Weather/safety exception.
- [ ] Customer-rides-along flow được test thực tế.

## Đăng kiểm hộ được coi là “pilot-ready” khi

- [ ] Inspection checklist template active.
- [ ] Customer khai báo giấy tờ.
- [ ] Driver verify giấy tờ.
- [ ] Required docs gate.
- [ ] Custody pickup.
- [ ] Inspection center selection.
- [ ] En-route + arrived + in-progress states.
- [ ] Result passed/failed/deferred/unavailable.
- [ ] Receipt/result evidence.
- [ ] Extra cost approval policy.
- [ ] Return journey.
- [ ] Custody return.
- [ ] Document return accountability.
- [ ] Failed inspection still completes service correctly.
- [ ] Retry/reschedule scenario.

---

# 11. Demo Bộ Công Thương — Minimum safe demo flow

## Customer

```text
Login
→ Home 3 services
→ Add/select vehicle
→ Choose service
→ Pickup
→ Destination / inspection
→ Immediate / schedule
→ Estimate
→ Confirm
→ Searching
→ Driver assigned
→ Realtime status
→ Custody confirm
→ Track service
→ Return custody confirm
→ Complete
→ History
```

## Driver

```text
Login
→ Approved profile
→ Online
→ Receive offer
→ Accept
→ Arriving
→ Arrived
→ [Inspection: verify checklist]
→ Capture custody evidence
→ Wait customer confirm
→ Vehicle received
→ Execute service
→ [Inspection states/result]
→ Return
→ Capture return evidence
→ Wait customer confirm
→ Handover
→ Complete
```

## Admin

```text
Login
→ Dashboard
→ Live map
→ Watch job
→ View customer/vehicle/driver
→ Manual dispatch if needed
→ KYC view
→ Incident handling
→ View custody evidence
→ [Inspection: view checklist]
→ Audit
```

---

# 12. UI/UX status

## Admin Web

✅ Đã chuyển sang FlashX iOS-inspired white + green design system.

## Landing

✅ Đã chuyển sang FlashX iOS-inspired white + green design system.

## Rider/Customer Mobile

🟡 Logic đã phát triển đáng kể nhưng **visual vẫn còn nhiều navy/yellow legacy**.

Cần migrate:

- semantic green tokens.
- white surfaces.
- Flowline.
- buttons/cards/bottom sheets.
- states/empty/error/loading.
- map overlays.

## Driver Mobile

🟡 Tương tự Rider: logic mới đã mạnh nhưng visual còn navy/yellow legacy.

**Không nên để demo cuối cùng có Admin/Landing một brand, mobile một brand khác.**

---

# 13. Current branch checkpoint — PHẢI ĐỌC KHI TIẾP TỤC

**Current branch:** `feat/inspection-checklist`

**Last committed base:** `30c7fd0` — Custody Evidence merged.

**Untracked work — KHÔNG XÓA:**

```text
infrastructure/migrations/010_inspection_checklist.sql
services/api/internal/inspectionchecklist/
```

Inspection checklist implementation đang có:

- model.
- memory store.
- PostgreSQL store.
- service.
- versioned template.
- snapshot per inspection job.
- customer declaration states.
- driver verification states.
- readiness.
- tests.

**Known immediate fix:**

Trong test đang gọi method cũ/không tồn tại:

```text
MarkArrivingForPickup()
MarkArrivedForPickup()
```

Trip service hiện dùng:

```text
MarkArriving()
MarkArrived()
```

Sau khi fix test, tiếp tục theo thứ tự:

```text
Go tests
→ runtime bootstrap
→ HTTP handlers/routes
→ Customer checklist UI
→ Driver verify UI
→ Admin template UI
→ Admin job checklist view
→ production gate
→ HTTP tests
→ Flutter tests/build
→ Admin build
→ commit checkpoint
→ merge dev
```

---

# 14. Tài liệu nào là nguồn ưu tiên sau audit này

Khi có mâu thuẫn về **trạng thái implementation**, ưu tiên:

1. `SCREEN_FLOW_IMPLEMENTATION_PLAN_2026-08-15.md` — trạng thái màn/logic mới nhất.
2. `CUSTOMER_REQUIREMENTS.md` — phạm vi sản phẩm.
3. `BA_MVP_OPERATING_RULES_2026-08-12.md` — quy tắc nghiệp vụ.
4. `design-system/flashx/MASTER.md` — UI/UX source of truth.
5. `.claude/skills/flashx-uiux/SKILL.md` — instruction cho coding agent.
6. `PRD_MVP.md`.
7. `PRODUCT_BACKLOG.md`.
8. `IMPLEMENTATION_GAP_REVIEW_2026-08-12.md` — **chỉ dùng lịch sử**, vì nhiều P0 trong file này đã được xử lý sau 12/08.

---

# 15. Quy tắc cập nhật tài liệu này

Khi merge một nhóm chức năng lớn:

1. Đổi trạng thái màn tương ứng.
2. Ghi commit/branch checkpoint mới ở mục 13.
3. Xóa gap chỉ khi **UI + API + business rule + test** đều đủ.
4. Không đánh dấu ✅ nếu mới có model/endpoint nhưng người dùng chưa thao tác được.
5. Không đánh dấu production-ready nếu chỉ chạy bằng demo fallback.
6. Mọi business exception mới phải thêm vào mục 8.

---

# 16. Kết luận cuối audit 15/08/2026

FlashX hiện đã vượt xa checkpoint 12/08: foundation marketplace và hai luồng lái hộ đã rõ, Admin Operations mạnh, Custody Evidence là một điểm tốt và có thể trở thành lợi thế niềm tin của sản phẩm.

**Phần cần tập trung ngay không phải thêm tính năng mới.** Việc đúng nhất là:

```text
Đóng Inspection Checklist
→ Address/Search/Pin
→ Route/ETA
→ Scheduler
→ Cancellation/Waiting
→ Notification
→ Rating UI
→ KYC onboarding Driver
→ Payment/Payout minimum viable
→ Legal/UAT
```

Sau đó mới tuyên bố **3 dịch vụ MVP pilot-ready** và mới bắt đầu mở nhánh cho ride-hailing/xe ghép.
