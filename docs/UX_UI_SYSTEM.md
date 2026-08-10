# UX/UI System — FlashX MVP

## 1. Product UX principles

1. Một màn hình chỉ có một primary action rõ ràng.
2. Trạng thái chuyến phải dễ hiểu hơn thuật ngữ kỹ thuật.
3. Rider và Driver luôn thấy hệ thống đang làm gì tiếp theo.
4. Không optimistic-update các action có race/financial impact trước khi backend xác nhận.
5. Mạng yếu là trạng thái bình thường, không phải edge case hiếm.
6. Bản đồ hỗ trợ task, không được làm UI quá tải.
7. Các action nguy hiểm (cancel, suspend, pricing change) cần confirmation phù hợp.

## 2. Visual direction

Phong cách: hiện đại, tin cậy, ít nhiễu, ưu tiên khả năng đọc ngoài trời.

Token semantic thay vì hard-code brand color trong feature:
- `primary`: CTA/chuyến đang hoạt động.
- `success`: completed/approved/online.
- `warning`: reconnect/pending/arriving attention.
- `danger`: cancel/reject/suspended/error.
- `surface`: card/sheet.
- `background`: page/map overlay background.
- `textPrimary`, `textSecondary`, `border`.

Brand color cuối cùng có thể thay bằng token mà không đổi feature UI.

## 3. Typoflashxhy

Ưu tiên system-friendly sans-serif, hierarchy:
- Display: màn hình trạng thái quan trọng.
- Title: tiêu đề screen/sheet.
- Body: nội dung chính.
- Label: button/field.
- Caption: metadata/ETA/helper/error.

Không dùng text quá nhỏ ở Driver App vì phải thao tác trong điều kiện di chuyển/ngoài trời.

## 4. Spacing & shape

- Base spacing: 4px; thường dùng 8/12/16/24/32.
- Touch target tối thiểu ~44–48px.
- Card radius nhất quán.
- Bottom sheet là primitive chính cho Rider map flow.
- Driver offer dùng layout high-attention, CTA lớn, countdown rõ.

## 5. Shared components

- PrimaryButton / SecondaryButton / DangerButton.
- AppTextField / PhoneField / OTPField.
- AppTopBar.
- StatusChip.
- EmptyState / ErrorState / OfflineBanner.
- LoadingSkeleton / BlockingProgress.
- ConfirmationSheet.
- MoneyText.
- TripStatusCard.
- DriverCard.
- LocationRow.
- MapBottomSheet.

## 6. Rider information architecture

Bottom navigation MVP:
- Home.
- Trips.
- Account.

Home:
```text
Map
  + Current-location control
  + Pickup pin
  + Bottom Sheet
      Where to?
      Recent locations (nếu có)
```

Booking:
```text
Pickup/Destination
-> Route + service options
-> Fare estimate
-> Confirm ride
-> Searching
-> Driver assigned
-> Tracking
-> Completed
-> Rating
```

### Rider state copy
Technical `SEARCHING` -> "Đang tìm tài xế gần bạn"
`ACCEPTED/ARRIVING` -> "Tài xế đang đến đón bạn"
`ARRIVED` -> "Tài xế đã đến điểm đón"
`IN_PROGRESS` -> "Bạn đang trên chuyến đi"
`COMPLETED` -> "Chuyến đi đã hoàn thành"

## 7. Driver information architecture

Bottom navigation MVP:
- Home.
- Earnings.
- Trips.
- Account.

Home offline:
- KYC/status warning nếu có.
- Online toggle primary.

Home online:
- Map/background context.
- Online duration/state.
- Current demand hint chỉ khi có dữ liệu thật.

Offer:
- Pickup summary.
- Distance/ETA to pickup.
- Service type.
- Fare/earning info theo business policy.
- Countdown.
- Accept primary, reject secondary.

Active trip:
- Rider/pickup info.
- Navigation CTA.
- One valid next action based on state: Arrived -> Start -> Complete.

## 8. Admin UX

Desktop-first, density cao hơn mobile.

Navigation:
- Dashboard
- Trips
- Drivers
- Riders
- Pricing
- Promotions
- Audit
- Settings

List pages:
- Search/filter ở đầu.
- Table có server pagination.
- Status dùng semantic chip.
- Detail mở page/drawer tùy complexity.

Action nhạy cảm luôn hiển thị reason/confirmation và quyền hiện tại.

## 9. Error/offline states

Rider:
- GPS off: giải thích + mở Settings.
- Location denied: cho nhập địa điểm thủ công khi có thể.
- No internet: giữ snapshot + reconnect indicator.
- No driver: thông báo rõ + retry/edit pickup.
- Session expired: re-auth nhưng không mất context nếu có thể.

Driver:
- GPS permission thiếu: không cho Online.
- GPS stale: warning và backend loại khỏi candidate mới.
- Network lost: offline banner, preserve active trip snapshot.
- Offer expired: đóng offer + thông báo ngắn, không để accept tiếp.

## 10. Accessibility

- Contrast đủ cho ngoài trời.
- Không dùng màu là tín hiệu duy nhất.
- Screen-reader label cho icon-only controls.
- Dynamic text không phá layout chính.
- Haptic/audio chỉ là bổ sung, không thay thông tin visual.

## 11. UX acceptance

Một flow chưa Done nếu chỉ có happy path. Mỗi screen phải định nghĩa tối thiểu loading, success, empty/error và reconnect/offline behavior nếu có network.
