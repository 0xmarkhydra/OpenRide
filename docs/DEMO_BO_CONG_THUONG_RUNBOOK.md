# FlashX — Runbook demo Bộ Công Thương

**Mục tiêu:** trình diễn một hệ thống end-to-end đang chạy thật, không phải mock UI: Customer App → Dispatch → Driver App → Custody Evidence → Realtime/State → Payment/History → Admin Operations.

> **Quy ước vận hành hiện tại:** local chạy trực tiếp bằng `.env`, **không dùng Docker local**. Môi trường `dev` trên Railway dùng shared PostgreSQL + shared Redis đã có của project.

## 1. Phạm vi demo đã khóa

FlashX hiện trình diễn đúng 3 dịch vụ sử dụng **phương tiện của khách hàng**:

1. **Lái hộ ô tô** — `designated_driver_car`
2. **Lái hộ xe máy** — `designated_driver_bike`
3. **Đăng kiểm hộ** — `vehicle_inspection_assist`

Thông điệp sản phẩm dùng trong demo:

> **Tài xế của bạn, khi bạn cần.**

Không giới thiệu FlashX MVP hiện tại như taxi/Grab/Uber. Kiến trúc vẫn chừa đường cho ride-hailing và xe ghép trong các giai đoạn sau.

## 2. Môi trường demo ưu tiên: Railway `dev`

Các URL public hiện dùng cho demo web:

```text
Backend: https://flashx-be-dev.up.railway.app
Admin:   https://flashx-admin-dev.up.railway.app
Landing: https://flashx-landing-dev.up.railway.app
```

Backend Railway `dev` dùng:

- shared PostgreSQL, database logic `flashx_dev`;
- shared Redis, logical DB `1`;
- migration chạy trước khi backend start;
- source branch `dev`.

Admin Operations đăng nhập bằng số điện thoại cấu hình tại `ADMIN_PHONE`. Super Admin demo hiện tại:

```text
0999999999
```

Ở môi trường development, OTP debug được backend trả về để demo; production không được dựa vào debug OTP.

## 3. Chạy backend local khi cần test kỹ thuật

Không chạy Docker local.

Tại repo:

```bash
cd /Users/levanmong/Documents/flashx/services/api
```

Nạp các biến môi trường local theo file `.env` của dự án, sau đó chạy trực tiếp Go API:

```bash
go run ./cmd/api
```

Local có thể dùng `PERSISTENCE=memory` cho test cô lập, hoặc kết nối hạ tầng dev theo cấu hình được cấp. Không tạo Postgres/Redis mới chỉ để chạy FlashX.

Kiểm tra:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/ready
```

## 4. Gate tự động trước buổi demo

Các gate bắt buộc:

```bash
cd services/api && go test ./...
cd ../../apps/rider && flutter test
cd ../driver && flutter test
cd ../admin && npm run build
```

Build APK debug để bắt cả lỗi native/plugin:

```bash
cd apps/rider && flutter build apk --debug
cd ../driver && flutter build apk --debug
```

Khi có một API dev/local đang chạy, có thể chạy smoke test 3 dịch vụ:

```bash
python3 scripts/demo_smoke.py
```

Kết quả mong đợi cuối cùng:

```text
ALL 3 FLASHX DEMO SERVICES PASSED
```

Smoke cũ kiểm chứng lifecycle 3 dịch vụ. Custody Evidence có test riêng ở Go HTTP/domain suite, bao gồm cả chế độ strict khi object storage được bật.

## 5. Chạy Customer App

### Android Emulator

```bash
cd apps/rider
flutter run
```

Android Emulator mặc định gọi API local tại `http://10.0.2.2:8080`.

### iOS Simulator

```bash
cd apps/rider
flutter run
```

iOS Simulator mặc định gọi API local tại `http://127.0.0.1:8080`.

### Dùng Railway dev

```bash
flutter run \
  --dart-define=API_BASE_URL=https://flashx-be-dev.up.railway.app
```

Nếu Maps app configuration hợp lệ thì bật bản đồ thật:

```bash
flutter run \
  --dart-define=API_BASE_URL=https://flashx-be-dev.up.railway.app \
  --dart-define=MAPS_ENABLED=true
```

Nếu chưa cấu hình Maps production, app dùng map fallback để luồng nghiệp vụ vẫn trình diễn được.

## 6. Chạy Driver App

Ví dụ với Railway dev:

```bash
cd apps/driver
flutter run \
  --dart-define=API_BASE_URL=https://flashx-be-dev.up.railway.app
```

Driver demo phải ở trạng thái **đã duyệt** và **Online**. Tài xế mới đăng ký bằng OTP mặc định ở trạng thái chờ Admin duyệt — đây là behavior thật cần giữ khi demo onboarding.

### Custody Evidence và object storage

Khi object storage được cấu hình:

1. tài xế ghi tình trạng xe;
2. chụp/chọn tối thiểu 2 ảnh;
3. tài xế xác nhận;
4. Customer App xem đúng bộ ảnh/tình trạng và xác nhận;
5. backend mới cho phép chuyển qua mốc **Đã nhận xe** hoặc **Bàn giao xe**.

Nếu backend development trả `OBJECT_STORAGE_UNAVAILABLE`, Driver App **không giả rằng ảnh đã được lưu**. App hiển thị cảnh báo “Kho ảnh demo chưa cấu hình” và chỉ khi đó mới xuất hiện action có hậu tố **`· Demo`**. Production không có bypass này vì production bootstrap yêu cầu object storage.

Có thể chủ động bật đường demo bằng build flag trong môi trường thử nghiệm:

```bash
flutter run \
  --dart-define=API_BASE_URL=https://flashx-be-dev.up.railway.app \
  --dart-define=DEMO_MODE=true
```

Không dùng `DEMO_MODE=true` để mô tả production flow.

## 7. Kịch bản demo — Lái hộ ô tô

Customer App:

1. Đăng nhập OTP.
2. Mở **Xe của tôi** và thêm/chọn ô tô.
3. Chọn **Lái hộ ô tô**.
4. Chọn điểm nhận xe và điểm đến.
5. Xem giá trước khi đặt.
6. Đặt dịch vụ.

Driver App:

1. Online.
2. Nhận offer có đúng thông tin xe khách.
3. Nhấn **Nhận việc**.
4. **Đang đến nhận xe**.
5. **Đã tới điểm nhận**.
6. Mở **Bằng chứng nhận xe**:
   - mô tả tình trạng;
   - odo/xăng/pin nếu cần;
   - chụp/chọn ít nhất 2 ảnh;
   - tài xế xác nhận.

Customer App:

7. Mở card **Bằng chứng nhận xe**.
8. Xem tình trạng + ảnh + thông số.
9. Nhấn xác nhận.

Driver App:

10. Khi evidence đã `ready`, nhấn **Xác nhận đã nhận xe**.
11. **Bắt đầu dịch vụ**.
12. Khi tới điểm bàn giao, tạo **Bằng chứng trả xe** theo cùng quy tắc.

Customer App:

13. Xem và xác nhận bộ bằng chứng trả xe.

Driver App:

14. **Bàn giao xe**.
15. **Hoàn thành**.

Customer App phải cập nhật trạng thái tương ứng; job xuất hiện trong lịch sử. Admin Operations nhìn thấy job, Flowline và hai bộ Custody Evidence.

## 8. Kịch bản demo — Lái hộ xe máy

Lặp lại quy trình trên với xe máy của khách. Flow MVP là tài xế lái **xe máy của khách**, đưa khách và xe tới điểm đến.

Custody Evidence vẫn dùng hai đầu nhận/trả; các trường odo/xăng/pin đều tùy chọn để phù hợp nhiều loại xe.

## 9. Kịch bản demo — Đăng kiểm hộ

Customer App:

1. Chọn ô tô đã lưu.
2. Chọn **Đăng kiểm hộ**.
3. Chọn điểm nhận/trả xe.
4. Xem phí dịch vụ và đặt.

Driver App:

1. Nhận việc.
2. Đến nhận xe.
3. Ghi **Bằng chứng nhận xe** + khách xác nhận.
4. Xác nhận nhận xe.
5. Đi tới trung tâm đăng kiểm.
6. Xác nhận tới nơi.
7. Bắt đầu đăng kiểm.
8. Chọn kết quả: đạt / không đạt / hoãn / không thực hiện được.
9. Trả xe.
10. Tại điểm trả, ghi **Bằng chứng trả xe** + khách xác nhận.
11. Bàn giao.
12. Hoàn thành dịch vụ.

Lưu ý khi thuyết trình: **kết quả đăng kiểm** và **trạng thái FlashX đã hoàn thành dịch vụ** là hai khái niệm riêng. Xe có thể không đạt đăng kiểm nhưng FlashX vẫn hoàn thành công việc sau khi trả xe cho khách.

## 10. Những điểm nên chủ động trình bày

- Khách sử dụng **xe của chính mình** trong 3 dịch vụ MVP.
- Giá được hiển thị trước khi xác nhận.
- Driver không thể nhận hai job đang thực hiện cùng lúc trong MVP.
- Sau khi tài xế đã nhận xe, khách không thể hủy theo flow hủy thông thường; phải đi qua hỗ trợ/sự cố/bàn giao an toàn.
- Driver onboarding có trạng thái chờ duyệt.
- KYC và Custody Evidence file đi trực tiếp từ app tới object storage bằng presigned URL; binary không đi xuyên FlashX API.
- Một bộ Custody Evidence chỉ `ready` khi có mô tả, ít nhất 2 ảnh và xác nhận cả tài xế lẫn khách.
- Nếu tài xế sửa tình trạng hoặc thêm ảnh sau khi đã xác nhận, backend reset xác nhận hai bên để không thay bằng chứng âm thầm.
- Job state và audit/history được lưu backend; UI không phải nguồn sự thật duy nhất.
- Admin Operations có bản đồ vận hành, job detail, evidence, KYC, incident, pricing, RBAC và cấu hình vận hành từ dữ liệu backend thật.

## 11. Phần cố ý chưa coi là production-final

Buổi demo chứng minh sản phẩm và hệ thống vận hành end-to-end. Các mục sau vẫn cần hardening trước commercial Go-Live quy mô lớn:

- Maps/Routes production và API key do chủ dự án sở hữu;
- SMS OTP production;
- payment gateway online ngoài cash-first MVP;
- push notification production;
- realtime hub đa instance qua Redis Pub/Sub khi scale ngang;
- monitoring/alerting/SLA production;
- hoàn thiện checklist giấy tờ đăng kiểm thành cấu hình nghiệp vụ đầy đủ;
- quy trình người nhận bàn giao ủy quyền/PIN/OTP nếu triển khai thực địa;
- rà soát pháp lý, bảo hiểm, hợp đồng tài xế và xử lý sự cố thực địa;
- một số chi tiết UX/BA tiếp tục được điều chỉnh sau demo.

Không trình bày các mục trên như đã production-complete nếu môi trường demo chưa cấu hình nhà cung cấp bên thứ ba tương ứng.
