# UX/UI System — FlashX MVP

> **UX rebaseline 12/08/2026:** FlashX là nền tảng **tài xế lái hộ / hỗ trợ phương tiện**, không còn là Grab/Uber clone. MVP chỉ gồm **Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ**.

## 1. Product UX principles

1. Mỗi màn hình chỉ có một primary action rõ ràng.
2. Khách phải luôn hiểu **ai đang giữ/lái xe của mình và công việc đang ở bước nào**.
3. Trạng thái hiển thị bằng ngôn ngữ nghiệp vụ tiếng Việt, không phơi technical state.
4. Trust là feature cốt lõi: tài xế/KYC/phương tiện/bàn giao phải được thể hiện rõ.
5. Không optimistic-update action có race, tài chính hoặc bàn giao tài sản trước khi backend xác nhận.
6. Mạng yếu/GPS lỗi là trạng thái bình thường cần thiết kế đầy đủ.
7. Bản đồ hỗ trợ task, không chiếm hết trải nghiệm.
8. Một service flow không được ép dùng layout/state của taxi nếu bản chất là workflow khác, đặc biệt Đăng kiểm hộ.
9. Không tạo false affordance: control trông bấm được thì phải có action thật.
10. Tối ưu thao tác một tay và đọc nhanh trên thiết bị thật.

## 2. Visual direction

### FlashX identity

- `navy / primary`: `#0B132B`
- `yellow / accent`: `#FFD600`
- `lime / success accent`: `#C7F36B`
- `mint / support accent`: dùng tiết chế cho trust/safe/secondary highlight
- `canvas`: `#F4F5F7` / `#F5F6F8`
- `surface`: `#FFFFFF`
- `textPrimary`: `#111827`
- `textSecondary`: `#6B7280`
- `border`: `#E5E7EB`
- `success`: `#0E9F6E`
- `warning`: `#F59E0B`
- `danger`: `#E5484D`

Không quay lại màu brand xanh cũ `#0A7A55`.

### Reference quality

Tham khảo mức polish, hierarchy, spacing và cách gom service card của các mobility app lớn tại Việt Nam như Green SM, nhưng:
- không sao chép logo;
- không sao chép asset độc quyền;
- không copy nguyên visual identity;
- giữ FlashX Navy + Electric Yellow + Lightning X.

## 3. Brand personality

FlashX phải truyền được 4 thuộc tính:
- **Nhanh** — ít bước, ETA rõ, CTA rõ.
- **Tin cậy** — KYC, rating, thông tin tài xế, timeline minh bạch.
- **An toàn** — handover/status/support dễ thấy.
- **Chuyên nghiệp** — không dùng UI vui nhộn quá mức khi khách giao tài sản có giá trị cao.

## 4. Typography / spacing / accessibility

- System-friendly sans-serif.
- Base spacing 4px; thường dùng 8/12/16/24/32.
- Touch target 44–48px trở lên.
- Driver CTA cần lớn hơn Customer CTA thông thường.
- Không dùng màu là tín hiệu duy nhất.
- Contrast đủ đọc ngoài trời.
- Icon-only control phải có semantic/screen-reader label.
- Dynamic text không phá CTA/state quan trọng.

## 5. Shared UI primitives

- PrimaryButton / SecondaryButton / DangerButton.
- AppTextField / PhoneField / OTPField.
- AppTopBar.
- ServiceCard.
- VehicleCard.
- DriverCard.
- LocationRow.
- MoneyText.
- StatusChip.
- JobTimeline.
- MapBottomSheet.
- ScheduleSelector.
- HandoverCard.
- DocumentStatusCard.
- EmptyState / ErrorState / OfflineBanner.
- LoadingSkeleton / BlockingProgress.
- ConfirmationSheet.

---

# 6. Customer App information architecture

## 6.1 Bottom navigation MVP

- **Trang chủ**
- **Lịch sử**
- **Thông báo** nếu push/notification center đưa vào MVP UI
- **Tài khoản**

Không dành tab chính cho vertical ngoài MVP.

## 6.2 Home

Home là **service-first**, map/context là hỗ trợ.

```text
Header / greeting / notification
Search/context: “Bạn cần tài xế làm gì?”

3 service CTA lớn:
  1. Lái hộ ô tô
  2. Lái hộ xe máy
  3. Đăng kiểm hộ

Quick context:
  - Xe của tôi
  - Hẹn giờ
  - Hỗ trợ

Map / current-location context nếu phù hợp
Bottom navigation
```

Yêu cầu:
- ba service phải nhìn thấy trong first meaningful viewport;
- không hiển thị Food/Giao hàng/Taxi/Thuê xe tự lái;
- service card phải nói rõ “tài xế lái xe của bạn”.

## 6.3 “Xe của tôi”

### Vehicle list
Mỗi card tối thiểu:
- loại xe;
- biển số;
- hãng/model;
- màu;
- transmission nếu ô tô;
- trạng thái/ảnh optional.

Primary action:
- Chọn xe khi booking, hoặc
- Thêm xe nếu chưa có.

### Add/Edit Vehicle
Không hỏi trường không cần thiết cho service type.

Ô tô:
- biển số;
- hãng/model;
- màu;
- số sàn/tự động;
- số chỗ;
- ghi chú.

Xe máy:
- biển số;
- hãng/model;
- màu;
- ghi chú.

## 6.4 Flow — Lái hộ ô tô

```text
Home
→ Lái hộ ô tô
→ Chọn xe của tôi
→ Điểm nhận xe
→ Điểm đến
→ Ngay bây giờ / Hẹn giờ
→ Estimate
→ Xác nhận
→ Đang tìm tài xế
→ Tài xế nhận job
→ Tài xế đang tới
→ Nhận xe / xác nhận bàn giao
→ Đang thực hiện
→ Bàn giao xe
→ Tổng kết / thanh toán
→ Đánh giá
```

### Service detail screen
Hiển thị:
- vehicle card;
- pickup;
- destination;
- note;
- ETA tài xế tới khách;
- route distance/time;
- estimated price;
- trust strip: tài xế được xác minh / support.

### Booking confirmation
- service name;
- xe;
- địa chỉ;
- thời gian;
- payment method;
- breakdown giá dự kiến;
- cancellation note ngắn;
- CTA `Xác nhận đặt`.

### Matching
Copy gợi ý:
- `Đang tìm tài xế phù hợp...`
- không dùng `Đang gọi xe`.

Sau assignment:
- ảnh/tên tài xế;
- rating;
- số job/kinh nghiệm nếu có;
- GPLX/KYC badge phù hợp;
- ETA;
- gọi/nhắn;
- map route tài xế → khách.

### Handover
MVP:
- `Tài xế đã đến`
- `Xác nhận bàn giao xe`
- vehicle summary
- note hiện trạng cơ bản nếu policy yêu cầu.

P1 có thể thêm ảnh/odometer/fuel/PIN.

## 6.5 Flow — Lái hộ xe máy

Giữ mental model giống ô tô để user dễ học, nhưng:
- vehicle form đơn giản;
- capability tài xế/pricing riêng;
- illustration/icon riêng;
- không hiển thị thông tin ô tô không liên quan.

## 6.6 Flow — Đăng kiểm hộ

Đây là **service job**, không phải trip chở khách.

```text
Home
→ Đăng kiểm hộ
→ Chọn ô tô
→ Địa chỉ nhận xe
→ Thời gian hẹn
→ Trung tâm đăng kiểm (nếu có)
→ Checklist giấy tờ
→ Địa chỉ trả xe
→ Giá/package
→ Xác nhận
→ Đã ghép người thực hiện
→ Đang tới nhận xe
→ Đã nhận xe/giấy tờ
→ Đang tới trung tâm
→ Đang đăng kiểm
→ Hoàn tất đăng kiểm
→ Đang trả xe
→ Đã bàn giao
→ Hoàn thành
```

UI cần có timeline rõ hơn map. Map chỉ xuất hiện khi location thực sự có ích.

## 6.7 Customer state copy

Business copy ưu tiên:

- `SEARCHING` → `Đang tìm tài xế phù hợp`
- `ACCEPTED/ARRIVING` → `Tài xế đang đến nhận xe`
- `ARRIVED` → `Tài xế đã đến`
- `VEHICLE_RECEIVED` → `Tài xế đã nhận xe`
- `IN_PROGRESS` → `Dịch vụ đang được thực hiện`
- `HANDOVER` → `Đang bàn giao xe`
- `COMPLETED` → `Dịch vụ đã hoàn thành`
- `CANCELLED` → `Yêu cầu đã hủy`

Đăng kiểm dùng subcopy cụ thể hơn như `Đang thực hiện đăng kiểm`.

## 6.8 Customer error/offline

- GPS off → giải thích + mở Settings + manual address fallback nếu có.
- Location denied → manual search.
- No internet → snapshot + reconnect indicator.
- No driver → retry / hẹn giờ / chỉnh khu vực theo policy.
- Offer/matching lâu → thông báo Operations đang hỗ trợ nếu có manual dispatch.
- Session expired → refresh/re-auth mà không mất booking draft nếu có thể.
- Scheduled booking invalid → giải thích lead-time rõ.

---

# 7. Driver App information architecture

## 7.1 Bottom navigation

- **Trang chủ**
- **Thu nhập**
- **Lịch sử**
- **Tài khoản**

## 7.2 Driver cockpit

Visual:
- dark/navy dominant;
- yellow action/high-attention;
- Online/Offline rất rõ;
- GPS/KYC state nhìn một phát hiểu ngay.

Home:
```text
Greeting + rating
Online / Offline
Today summary
Earnings summary
Capability/preferences
Current zone / GPS state
```

Không hiển thị demand/earnings giả nếu backend không có dữ liệu thật.

## 7.3 Incoming offer

Offer card phải hiển thị:
- loại service;
- countdown lớn;
- pickup;
- destination/inspection location;
- distance/ETA tới khách;
- vehicle summary;
- estimated earning/fare theo business policy;
- relevant note;
- `Từ chối` secondary;
- `Chấp nhận` primary.

### Service visual
- Lái hộ ô tô → car + vehicle summary nổi bật.
- Lái hộ xe máy → motorbike.
- Đăng kiểm hộ → document/inspection icon + schedule + center/checklist.

## 7.4 Active designated-driver job

```text
Accepted
→ Dẫn đường tới khách
→ Đã tới
→ Nhận xe
→ Bắt đầu
→ Đang thực hiện
→ Bàn giao
→ Hoàn thành
```

Mỗi state chỉ hiển thị **một CTA chính tiếp theo**.

Thông tin cần luôn truy cập được:
- khách;
- gọi khách;
- vehicle;
- pickup/destination;
- support.

## 7.5 Inspection job

UI stepper riêng:
- tới nhận xe;
- nhận xe/giấy tờ;
- tới trung tâm;
- đang đăng kiểm;
- hoàn tất;
- trả xe;
- bàn giao.

Không dùng copy `Đón khách` cho Đăng kiểm hộ.

## 7.6 Driver KYC/Profile

Hiển thị:
- trạng thái hồ sơ;
- document checklist;
- capability được duyệt;
- GPLX expiry;
- banking state;
- action bổ sung hồ sơ.

Document upload client → storage trực tiếp; UI cần show upload progress/retry/complete state.

## 7.7 Driver offline/error states

- KYC chưa approved → không cho Online, CTA tới hồ sơ.
- GPS permission thiếu → không cho Online.
- GPS stale → warning.
- Network lost → offline banner + preserve active job snapshot.
- Offer expired → close offer; disable accept.
- Service capability mismatch → offer không nên xuất hiện; nếu xảy ra show safe error và report.

---

# 8. Admin / Operations UX

## 8.1 Navigation MVP

- Tổng quan
- Công việc (Live)
- Đơn/Yêu cầu
- Tài xế
- Khách hàng
- Cấu hình dịch vụ / Bảng giá
- Hỗ trợ khách hàng
- Báo cáo cơ bản
- Audit/Cài đặt theo quyền

## 8.2 Overview

KPI:
- tổng job hôm nay;
- split 3 service;
- tài xế online;
- job chờ lâu;
- completion/cancellation/no-match;
- doanh thu;
- KYC warning.

Không dùng vanity chart nếu không hỗ trợ quyết định vận hành.

## 8.3 Live Jobs

Table/filter:
- service;
- status;
- scheduled/immediate;
- customer;
- vehicle;
- driver;
- pickup/destination;
- waiting/match time.

Detail drawer:
- job summary;
- customer + vehicle;
- driver;
- timeline;
- support notes;
- map nếu relevant;
- controlled actions theo RBAC.

## 8.4 Driver/KYC

Master-detail UX:
- pending/approved/rejected/suspended tabs;
- document thumbnail/status;
- signed document view;
- capability;
- GPLX expiry;
- Approve / Reject / Request more information;
- reason bắt buộc cho reject/sensitive changes.

## 8.5 Pricing configuration

Tách theo 3 service:
- Lái hộ ô tô;
- Lái hộ xe máy;
- Đăng kiểm hộ.

Pricing screen cần:
- version/effective time;
- base/minimum;
- distance/time;
- waiting;
- schedule;
- night/holiday;
- cancellation;
- inspection package;
- preview/validation trước Save.

## 8.6 Admin visual

- dark navy sidebar;
- white content surfaces;
- yellow active/accent;
- semantic chips cho status;
- dense nhưng dễ scan;
- desktop-first; tablet responsive hợp lý;
- destructive action không dùng yellow primary.

---

# 9. UI states bắt buộc cho mọi feature

Một screen có network/data chưa Done nếu thiếu:
- loading;
- success;
- empty;
- error;
- retry;
- offline/reconnect nếu relevant;
- permission denied nếu relevant;
- disabled/expired state nếu action có TTL.

## 10. Micro-interaction guidelines

- Haptic nhẹ khi toggle Online, accept offer, confirm handover nếu platform hỗ trợ.
- Offer countdown animation không gây distraction.
- Matching/searching có animation nhẹ nhưng phải có text state.
- Timeline transition animate ngắn, không delay action.
- Success animation tiết chế cho completion.
- Skeleton thay spinner dài ở list/dashboard khi phù hợp.

## 11. Mock → implementation rule

Bộ mock Customer/Driver/Admin đã tạo là **visual target**, không phải source of truth nghiệp vụ.

Nếu mock có:
- feature chưa trong MVP;
- số liệu giả;
- bảo hiểm/SLA chưa được khách chốt;
- payment method chưa implement;

thì code **không được tự thêm business promise chỉ vì mock hiển thị**.

Source of truth luôn là:
1. `CUSTOMER_REQUIREMENTS.md`
2. `PRD_MVP.md`
3. API/domain implementation đã được review

## 12. UX Definition of Done

Một flow UX chỉ Done khi:
1. đúng business của một trong 3 service MVP;
2. không dùng semantics taxi/Grab sai ngữ cảnh;
3. customer vehicle được thể hiện đúng;
4. primary CTA rõ;
5. state/timeline rõ;
6. loading/error/offline được xử lý;
7. trust/KYC/handover information đúng mức cần thiết;
8. responsive/device layout không overflow;
9. analyze/test/build gate liên quan xanh;
10. UAT trên thiết bị thật không có blocker về hiểu nhầm thao tác.
