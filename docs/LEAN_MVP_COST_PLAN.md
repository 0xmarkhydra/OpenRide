# FlashX — Lean MVP Cost Plan

> Cập nhật: 12/08/2026  
> **Trạng thái:** phương án tham khảo, **không còn là phương án được chọn**. Founder sau đó chốt **Full Marketplace**; xem [`FULL_MARKETPLACE_COST_PLAN.md`](./FULL_MARKETPLACE_COST_PLAN.md).  
> Bối cảnh ban đầu của file: Founder muốn UI đơn giản như mẫu; kết luận hiện tại là **UI vẫn đơn giản nhưng backend/marketplace phải đầy đủ**.

## 1. Kết luận điều hành

Nếu FlashX chỉ cần một ứng dụng đơn giản, dễ dùng, tập trung đặt dịch vụ và vận hành pilot 1 thành phố, thì không nên giữ planning figure 650 triệu cho phần mềm.

Có 3 mức triển khai:

| Phương án | Mô tả | Budget phần mềm |
|---|---|---:|
| Lite | Customer App + Admin, điều phối tài xế thủ công/semi-manual | 120–160 triệu |
| Lean MVP — khuyến nghị | Customer App + Driver App + Admin + backend booking/KYC/GPS cơ bản | 220–280 triệu |
| Full marketplace | Auto dispatch/realtime mạnh, pricing phức tạp, production hardening cao | 450–650 triệu |

**Khuyến nghị Founder:** chọn Lean MVP khoảng **250 triệu VNĐ**.

Mục tiêu của Lean MVP là chứng minh được:
- có khách đặt dịch vụ;
- có tài xế nhận và hoàn thành dịch vụ;
- Operations theo dõi/điều phối được;
- thị trường có willingness-to-pay;
- chưa burn tiền vào các hệ thống phức tạp trước khi có traction.

---

## 2. Lean MVP — scope khuyến nghị

### Customer App
- OTP/login.
- Home đơn giản như mẫu.
- 3 service cards:
  - Lái hộ ô tô;
  - Lái hộ xe máy;
  - Đăng kiểm hộ.
- Chọn địa chỉ nhận/trả.
- Hẹn giờ.
- Nhập/chọn xe của khách.
- Giá dự kiến hoặc bảng giá package.
- Tạo yêu cầu.
- Xem trạng thái.
- Xem tài xế sau khi được assign.
- Lịch sử.
- Tài khoản.

### Driver App
- Login.
- KYC cơ bản.
- Online/Offline.
- Nhận offer/job.
- Accept/Reject.
- Xem thông tin khách/xe.
- Đã tới → Nhận xe → Bắt đầu → Hoàn thành/Bàn giao.
- GPS cơ bản khi đang thực hiện job.
- Lịch sử/thu nhập đơn giản.

### Admin
- Login.
- Dashboard cơ bản.
- Danh sách khách/tài xế/job.
- Duyệt KYC.
- Gán tài xế thủ công khi cần.
- Theo dõi trạng thái.
- Cấu hình bảng giá cơ bản.
- Hỗ trợ/hủy job.

### Backend
- Auth/OTP.
- PostgreSQL.
- API booking/job.
- CustomerVehicle.
- Driver profile/KYC.
- Object storage presigned upload.
- Basic location.
- Basic notifications.
- Pricing đơn giản.
- Admin RBAC.

### Chưa cần ở Lean MVP
- Auto dispatch nhiều vòng như Grab.
- Surge pricing.
- Ví nội bộ.
- Payment gateway phức tạp.
- Loyalty/referral.
- AI/fraud engine.
- Call masking.
- Embedded navigation tự xây.
- Multi-city architecture.
- Microservices.
- ML demand forecasting.

---

## 3. Bóc giá Lean MVP — planning 250 triệu

| Hạng mục | Budget (triệu VNĐ) |
|---|---:|
| BA + UX/UI + prototype | 20 |
| Backend/API/DB/Auth | 55 |
| Customer App Flutter | 45 |
| Driver App Flutter | 35 |
| Admin Portal | 25 |
| Maps/SMS/S3/notification integration | 15 |
| QA/UAT/bug fixing | 20 |
| DevOps/release/docs | 15 |
| Contingency nhỏ | 20 |
| **Tổng planning** | **250** |

Khoảng thương mại hợp lý: **220–280 triệu VNĐ** tùy mức polish, payment, Maps và workflow đăng kiểm thực tế.

---

## 4. Nếu muốn rẻ nhất có thể — Lite 120–160 triệu

Có thể bỏ Driver App giai đoạn đầu.

Flow:

Customer App → tạo yêu cầu → Admin/Operations thấy job → gọi/Zalo tài xế → gán tài xế → cập nhật trạng thái → khách theo dõi trạng thái cơ bản.

Ưu điểm:
- ra thị trường nhanh;
- ít code;
- ít bug;
- dễ thay đổi quy trình sau khi học từ thị trường.

Nhược điểm:
- Operations tốn người;
- không scale tốt;
- trải nghiệm tài xế chưa đẹp;
- phải nâng cấp Driver App khi volume tăng.

Phương án này phù hợp nếu mục tiêu chỉ là pilot 50–100 job/ngày trở xuống và Founder muốn kiểm chứng thị trường trước.

---

## 5. Chi phí thành lập + launch theo mô hình Lean

Không cần bắt đầu với bộ máy 11 người và burn 400 triệu/tháng.

### Setup một lần

| Khoản | Budget (triệu VNĐ) |
|---|---:|
| Phần mềm Lean MVP | 250 |
| Thành lập doanh nghiệp + kế toán/chữ ký số/hóa đơn | 15 |
| Legal/contracts/privacy/compliance baseline | 35 |
| Domain/brand/store accounts | 5 |
| Cloud/production/devices/setup | 20 |
| Tuyển/KYC/onboarding nhóm tài xế đầu | 40 |
| Content/creative/launch marketing | 50 |
| Promotion/incentive launch | 50 |
| Quỹ sự cố/bồi hoàn/risk reserve ban đầu | 40 |
| Workspace/coworking/setup | 20 |
| **Tổng setup ban đầu** | **525 triệu** |

---

## 6. Bộ máy Lean 5–6 người

| Vai trò | Số lượng | Budget/tháng (triệu VNĐ) |
|---|---:|---:|
| Founder/CEO allowance | 1 | 10 |
| Operations Lead | 1 | 22 |
| Driver Ops + Customer Support | 2 | 28 |
| Growth/Partnership | 1 | 16 |
| Finance/Admin part-time/outsource | 1 | 6 |
| Tech maintenance/support outsource | — | 15 |
| **People/tech core** | | **97** |

Founder có thể tự kiêm Product/Sales để giữ burn thấp.

---

## 7. Burn rate Lean hàng tháng

| Khoản | Budget/tháng (triệu VNĐ) |
|---|---:|
| People + technical maintenance | 97 |
| Coworking/điện/internet | 8 |
| Cloud/Maps/SMS/storage | 5 |
| Marketing always-on | 15 |
| Driver incentives/acquisition | 12 |
| Kế toán/pháp lý/SaaS | 5 |
| CS/điện thoại/di chuyển | 5 |
| Incident/misc reserve | 8 |
| **Burn rate Lean** | **155 triệu/tháng** |

Planning range: **130–180 triệu/tháng**.

---

## 8. Vốn cần có để khởi động

### 3 tháng runway
- Setup: 525 triệu
- Burn 3 tháng: 465 triệu
- Tổng: 990 triệu
- +10% buffer: **~1,09 tỷ**

### 6 tháng runway — khuyến nghị
- Setup: 525 triệu
- Burn 6 tháng: 930 triệu
- Tổng: 1,455 tỷ
- +10% buffer: **~1,60 tỷ**

### 12 tháng runway
- Setup: 525 triệu
- Burn 12 tháng: 1,86 tỷ
- Tổng: 2,385 tỷ
- +10% buffer: **~2,62 tỷ**

> **Mục tiêu vốn khuyến nghị cho Founder nếu đi Lean: khoảng 1,5–1,7 tỷ để launch và có 6 tháng thử thị trường.**

Nếu phần mềm hiện tại đã được coi là sunk cost/đã sở hữu, phần cash mới cần bỏ ra sẽ thấp hơn tương ứng.

---

## 9. Khi nào mới nâng từ Lean lên Full?

Chỉ đầu tư auto-dispatch/realtime phức tạp hơn khi có một trong các tín hiệu:
- >100–200 completed jobs/ngày;
- Operations bắt đầu không điều phối thủ công nổi;
- ETA/no-match trở thành vấn đề lớn;
- có >300–500 tài xế active;
- có kế hoạch mở thành phố thứ hai;
- unit economics đã chứng minh được.

Khi đó mới thêm:
- automatic geo matching;
- dispatch retry/radius expansion;
- dynamic pricing;
- push/offline reliability nâng cao;
- payout/settlement;
- richer monitoring/anti-fraud.

---

## 10. Khuyến nghị cuối

Với yêu cầu Founder “chỉ cần app đơn giản như mẫu”, phương án hợp lý nhất là:

**Lean MVP ~250 triệu phần mềm + ~275 triệu setup launch + burn ~155 triệu/tháng.**

Như vậy:
- Zero → Launch: khoảng **525 triệu**.
- Launch + 6 tháng runway: khoảng **1,5–1,6 tỷ**.

Không nên tiếp tục plan 4 tỷ ngay từ đầu nếu Founder chưa cần một operation lớn hoặc Grab-like automation.

Tư duy đúng là:

> **Build nhỏ → chạy thật → đo đơn → học vận hành → mới scale công nghệ và đội ngũ.**
