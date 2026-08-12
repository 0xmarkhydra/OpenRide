# Admin Portal

> **Current product direction:** Operations quản lý job theo 3 dịch vụ **Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ**. Dashboard/live jobs/KYC/pricing/support phải ưu tiên ba service này; xem [`CUSTOMER_REQUIREMENTS.md`](./CUSTOMER_REQUIREMENTS.md) để tránh tiếp tục triển khai theo ride-hailing semantics cũ.

## 1. Mục tiêu

Admin Portal phục vụ Operations, Support và quản trị hệ thống; không phải dashboard chỉ để xem số liệu.

## 2. Stack

- Next.js + TypeScript.
- Server-side data fetching hoặc client query tùy màn hình.
- Tất cả action nhạy cảm gọi Admin API và được RBAC/audit ở backend.

## 3. Roles MVP

Đề xuất:
- `super_admin`: toàn quyền.
- `operations`: quản lý trip/driver/rider.
- `driver_review`: duyệt KYC.
- `support`: tra cứu và hỗ trợ, quyền chỉnh sửa hạn chế.
- `finance`: payment/refund/report nếu phase có.

Không chỉ ẩn button ở frontend; backend bắt buộc enforce quyền.

## 4. Dashboard

Hiển thị tối thiểu:
- trips hôm nay;
- searching/active/completed/cancelled;
- online drivers;
- time-to-match;
- zero-driver/no-match rate;
- riders/drivers mới;
- error/operational alerts quan trọng.

## 5. Driver management

- search/filter driver;
- xem profile;
- xem vehicle;
- xem KYC documents qua signed access;
- approve/reject/request more info;
- suspend/unsuspend;
- xem trip history;
- audit lịch sử thay đổi.

## 6. Rider management

- search bằng phone/id;
- status;
- trip history;
- suspend/unsuspend theo quyền;
- support notes nếu implement.

## 7. Trip management

- filter theo time/status/city/driver/rider;
- trip detail;
- pickup/destination;
- fare breakdown;
- timeline `trip_status_history`;
- driver/rider liên quan;
- payment state;
- operational notes/audit.

Manual state override không nên có ở MVP trừ khi business bắt buộc; nếu có phải cực kỳ hạn chế và audit đầy đủ.

## 8. Pricing

Admin có thể quản lý pricing rule có version/effective time.

Không update giá production mà mất lịch sử. Mỗi thay đổi phải lưu:
- old config/new config hoặc version;
- actor;
- effective_at;
- reason optional.

## 9. Promotions

- code;
- type/value;
- max discount;
- valid window;
- usage limit;
- status.

Validation logic ở backend.

## 10. Live operations map

Phase MVP+ nên có màn hình:
- driver online sampled;
- active trips;
- searching trips;
- operational hotspot.

Không broadcast toàn bộ raw GPS nếu không cần; aggregate/sampling để bảo vệ privacy và hiệu năng.

## 11. Audit

Action bắt buộc audit:
- KYC decision;
- suspend/unsuspend;
- pricing/promotion change;
- refund/financial override;
- role change;
- manual trip intervention.

## 12. Security UX

- session timeout phù hợp;
- re-auth hoặc step-up auth cho action rất nhạy cảm nếu cần;
- không hiển thị KYC data vượt quyền;
- mask PII ở list view khi phù hợp.

## 13. QA

Mỗi role phải có permission matrix test. Test cả việc user không có quyền gọi API trực tiếp, không chỉ test UI.
