# FlashX — Full Marketplace Cost Plan

> **Founder decision:** 12/08/2026 — chọn **Full Marketplace**.  
> UI có thể vẫn đơn giản như mẫu Founder gửi, nhưng phía sau phải là marketplace tự động hóa đầy đủ.

## 1. Định nghĩa Full Marketplace

FlashX Full Marketplace gồm 3 ứng dụng/hệ thống:

1. **Customer App** — Lái hộ ô tô / Lái hộ xe máy / Đăng kiểm hộ.
2. **Driver App** — KYC, online/offline, GPS realtime, offer, accept/reject, workflow thực hiện, earnings.
3. **Admin/Operations** — live jobs, KYC, pricing, driver/customer management, support, manual override, metrics.

Backend phải hỗ trợ:
- auth/OTP/session;
- CustomerVehicle;
- pricing engine;
- scheduled booking;
- Redis GEO candidate discovery;
- automatic matching/dispatch;
- offer TTL/countdown;
- atomic assignment/double-assignment protection;
- GPS realtime + WebSocket;
- push/SMS fallback;
- job state machine cho lái hộ và đăng kiểm hộ;
- payment/ledger;
- rating/history;
- KYC direct-to-S3;
- audit/security/observability;
- production CI/CD/backup/recovery.

UI không cần phức tạp. **Sự phức tạp nằm ở engine vận hành phía sau.**

Tuy nhiên Full Marketplace **bắt buộc giữ trải nghiệm bản đồ realtime**:
- trước khi đón: Customer thấy tài xế đang ở đâu, tuyến tài xế → điểm nhận và ETA;
- trong chuyến lái hộ: Customer thấy vị trí tài xế/xe, route điểm nhận → điểm đến và tiến trình hành trình;
- trạng thái cập nhật realtime qua WebSocket/GPS, không phải refresh thủ công;
- Admin Operations có live-location/live-job view phù hợp;
- Đăng kiểm hộ có tracking nhận xe/trả xe và timeline nghiệp vụ phù hợp.

“UI đơn giản” **không có nghĩa là cắt GPS/map/tracking**; chỉ có nghĩa là giảm số chức năng và làm luồng dễ dùng.

---

## 2. Planning budget phần mềm

Khoảng thương mại: **450–650 triệu VNĐ**.

Planning target để làm bài bản: **600 triệu VNĐ**.

| Hạng mục | Budget (triệu VNĐ) |
|---|---:|
| BA/Product/Domain + UX/UI | 55 |
| Backend/API/Postgres/Redis/Realtime | 125 |
| Dispatch/Matching/Pricing/Scheduling | 75 |
| Customer App Flutter | 80 |
| Driver App Flutter | 90 |
| Admin/Operations | 55 |
| Maps/SMS/Push/S3/Payment integrations | 35 |
| QA/Security/E2E/UAT | 40 |
| DevOps/Monitoring/Release | 25 |
| Documentation/Handover/Training | 10 |
| Contingency kỹ thuật | 10 |
| **Tổng planning** | **600** |

Nếu cần báo giá có buffer cho Change Request/production hardening mạnh, có thể giữ trần **650 triệu**.

### Lưu ý code hiện tại

FlashX hiện đã có nhiều foundation: Go backend, Flutter Customer/Driver, Next.js Admin, PostgreSQL/PostGIS, Redis GEO, WebSocket, auth, dispatch nền, payment ledger, KYC/S3 và Docker/CI.

Do đó **600–650 triệu là replacement/full-project commercial value**, không nhất thiết là lượng tiền mặt còn phải chi từ hôm nay. Cần audit code hiện tại để tính incremental completion cost riêng.

---

## 3. Setup doanh nghiệp + Go-Live ngoài phần mềm

| Nhóm | Budget (triệu VNĐ) |
|---|---:|
| Thành lập DN + kế toán/chữ ký số/hóa đơn | 15 |
| Legal/compliance/contracts/privacy/TMĐT | 60 |
| Brand/domain/IP/store accounts | 15 |
| Production infra/device/security setup | 35 |
| Tuyển/KYC/onboarding tài xế ban đầu | 70 |
| Marketing/creative/launch | 80 |
| Promotion + driver incentive | 100 |
| Risk/incident/customer compensation reserve | 80 |
| Workspace/setup | 45 |
| **Tổng ngoài phần mềm** | **500** |

**Zero → Go-Live planning:** khoảng **1,10 tỷ VNĐ** nếu tính 600 triệu software + 500 triệu setup.

---

## 4. Team vận hành Full Marketplace pilot 1 thành phố

| Vai trò | Số lượng | Budget/tháng (triệu VNĐ) |
|---|---:|---:|
| Founder/CEO allowance | 1 | 10 |
| COO / Operations Lead | 1 | 30 |
| Product/Tech Lead | 1 | 30 |
| Driver Operations | 2 | 30 |
| Customer Support | 2 | 24 |
| Partnership/Sales | 1 | 20 |
| Marketing/Growth | 1 | 18 |
| Finance/Admin | 1 | 13 |
| QA/Technical Support | 1 | 15 |
| **Lương planning** | **11** | **190** |

Reserve payroll/benefits/recruitment: **25 triệu/tháng**.

People cost planning: **~215 triệu/tháng**.

---

## 5. Burn rate hàng tháng

| Khoản | Budget/tháng (triệu VNĐ) |
|---|---:|
| People + payroll reserve | 215 |
| Office/Internet | 22 |
| Cloud/Maps/SMS/Storage/Monitoring | 15 |
| Technical maintenance/on-call | 35 |
| Accounting/legal retainer | 8 |
| Marketing always-on | 50 |
| Driver acquisition/incentive | 30 |
| CS/SaaS/phone | 7 |
| Incident/compensation reserve | 10 |
| Partnership/travel/misc | 8 |
| **Burn rate** | **400 triệu/tháng** |

Target quản trị: **350–400 triệu/tháng** trong pilot có tổ chức đầy đủ.

---

## 6. Vốn cần chuẩn bị

Giả định:
- Software: 600 triệu.
- Setup ngoài software: 500 triệu.
- Zero → Go-Live: 1,10 tỷ.
- Burn: 400 triệu/tháng.
- Contingency: 10%.

### 3 tháng runway

1,10 tỷ + 1,20 tỷ = 2,30 tỷ  
+ 10% = **~2,53 tỷ VNĐ**.

### 6 tháng runway — khuyến nghị

1,10 tỷ + 2,40 tỷ = 3,50 tỷ  
+ 10% = **~3,85 tỷ VNĐ**.

=> Làm tròn **4 tỷ VNĐ**.

### 12 tháng runway

1,10 tỷ + 4,80 tỷ = 5,90 tỷ  
+ 10% = **~6,49 tỷ VNĐ**.

=> Planning **6,5 tỷ VNĐ**.

---

## 7. Kết luận Founder

### Quyết định hiện tại

**FlashX đi Full Marketplace.**

Không làm Lite/Lean làm default nữa.

Tuy nhiên UI Customer vẫn giữ nguyên triết lý Founder yêu cầu:
- đơn giản;
- ít nút;
- 3 dịch vụ chính;
- thao tác đặt yêu cầu nhanh;
- không biến giao diện thành một Grab clone nhiều vertical.

### Cách giải thích ngắn gọn

> **Mặt trước đơn giản — hệ thống phía sau đầy đủ.**

Khách chỉ thấy vài nút, nhưng hệ thống tự xử lý matching, GPS, offer, KYC, giá, trạng thái, notification, admin và dữ liệu vận hành.

### Mốc tài chính quản trị

- Software Full Marketplace: **~600 triệu** planning, range **450–650 triệu**.
- Zero → Go-Live: **~1,1 tỷ**.
- Burn: **~400 triệu/tháng**.
- 6 tháng runway: **~4 tỷ**.
- 12 tháng runway: **~6,5 tỷ**.

Nếu audit code hiện tại cho thấy phần lớn Full Marketplace foundation đã hoàn thành, cần lập thêm **Incremental Completion Budget** để xác định số tiền thực tế còn phải bỏ ra từ hôm nay.
