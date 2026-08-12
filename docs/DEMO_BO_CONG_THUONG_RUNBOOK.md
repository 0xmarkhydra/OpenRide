# FlashX — Runbook demo Bộ Công Thương

**Mục tiêu:** trình diễn một hệ thống end-to-end đang chạy thật, không phải mock UI: Customer App → Dispatch → Driver App → Realtime/State → Payment/History → Admin Operations.

## 1. Phạm vi demo đã khóa

FlashX hiện trình diễn đúng 3 dịch vụ sử dụng **phương tiện của khách hàng**:

1. **Lái hộ ô tô** — `designated_driver_car`
2. **Lái hộ xe máy** — `designated_driver_bike`
3. **Đăng kiểm hộ** — `vehicle_inspection_assist`

Thông điệp sản phẩm dùng trong demo:

> **Tài xế của bạn, khi bạn cần.**

Không giới thiệu FlashX MVP hiện tại như taxi/Grab/Uber. Kiến trúc vẫn chừa đường cho ride-hailing và xe ghép trong các giai đoạn sau.

## 2. Khởi động backend + database + admin

Tại root repo:

```bash
cd /Users/levanmong/Documents/flashx
docker compose up -d --build postgres redis migrate api admin
```

Kiểm tra:

```bash
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/ready
```

Admin Operations:

```text
http://127.0.0.1:3000
```

Admin đăng nhập bằng số điện thoại cấu hình tại `ADMIN_PHONE` / `NEXT_PUBLIC_ADMIN_PHONE`. Ở môi trường development, OTP debug được hiển thị để demo; production không được dựa vào debug OTP.

## 3. Gate tự động trước buổi demo

Chạy smoke test end-to-end cả 3 dịch vụ:

```bash
python3 scripts/demo_smoke.py
```

Kết quả bắt buộc cuối cùng:

```text
ALL 3 FLASHX DEMO SERVICES PASSED
```

Smoke test tự tạo dữ liệu riêng và giữ các job đã hoàn thành để chúng xuất hiện trên Admin Operations.

Các gate kỹ thuật khác:

```bash
cd services/api && go test ./...
cd ../../apps/rider && flutter test
cd ../driver && flutter test
cd ../admin && npm run build
```

## 4. Chạy Customer App

### Android Emulator

```bash
cd apps/rider
flutter run
```

Android Emulator mặc định gọi API tại `http://10.0.2.2:8080`.

### iOS Simulator

```bash
cd apps/rider
flutter run
```

iOS Simulator mặc định gọi API tại `http://127.0.0.1:8080`.

### Điện thoại thật cùng Wi-Fi với máy demo

Dùng IP LAN của máy chạy backend:

```bash
flutter run --dart-define=API_BASE_URL=http://<IP-LAN-MAC>:8080
```

Nếu có Google Maps key/app configuration hợp lệ thì có thể bật bản đồ thật:

```bash
flutter run \
  --dart-define=API_BASE_URL=http://<IP-LAN-MAC>:8080 \
  --dart-define=MAPS_ENABLED=true
```

Nếu chưa cấu hình Maps production, app sử dụng map fallback để luồng nghiệp vụ vẫn trình diễn được.

## 5. Chạy Driver App

Tương tự Customer App:

```bash
cd apps/driver
flutter run
```

Hoặc với thiết bị thật:

```bash
flutter run --dart-define=API_BASE_URL=http://<IP-LAN-MAC>:8080
```

Driver demo phải ở trạng thái **đã duyệt** và **Online**. Tài xế mới đăng ký bằng OTP mặc định ở trạng thái chờ Admin duyệt — đây là behavior thật cần giữ khi demo quy trình onboarding.

## 6. Kịch bản demo khuyến nghị

### A. Lái hộ ô tô

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
6. **Đã nhận xe**.
7. **Bắt đầu dịch vụ**.
8. **Bàn giao xe**.
9. **Hoàn thành**.

Customer App phải cập nhật trạng thái tương ứng; job xuất hiện trong lịch sử. Admin Operations nhìn thấy job và trạng thái.

### B. Lái hộ xe máy

Lặp lại quy trình trên với xe máy của khách. Flow mặc định hiện tại là tài xế lái xe máy của khách tới điểm đến theo BA MVP.

### C. Đăng kiểm hộ

Customer App:

1. Chọn ô tô đã lưu.
2. Chọn **Đăng kiểm hộ**.
3. Chọn điểm nhận/trả xe.
4. Xem phí dịch vụ và đặt.

Driver App:

1. Nhận việc.
2. Đến nhận xe.
3. Xác nhận nhận xe.
4. Đi tới trung tâm đăng kiểm.
5. Xác nhận tới nơi.
6. Bắt đầu đăng kiểm.
7. Chọn kết quả: đạt / không đạt / hoãn / không thực hiện được.
8. Trả xe.
9. Bàn giao.
10. Hoàn thành dịch vụ.

Lưu ý khi thuyết trình: **kết quả đăng kiểm** và **trạng thái FlashX đã hoàn thành dịch vụ** là hai khái niệm riêng. Xe có thể không đạt đăng kiểm nhưng FlashX vẫn hoàn thành công việc sau khi trả xe cho khách.

## 7. Những điểm nên chủ động trình bày

- Khách sử dụng **xe của chính mình** trong 3 dịch vụ MVP.
- Giá được hiển thị trước khi xác nhận.
- Driver không thể nhận hai job đang thực hiện cùng lúc trong MVP.
- Sau khi tài xế xác nhận đã nhận xe, khách không thể hủy theo flow hủy thông thường; phải đi qua hỗ trợ/sự cố/bàn giao an toàn.
- Driver onboarding có trạng thái chờ duyệt.
- KYC file đi trực tiếp từ Driver App tới object storage bằng presigned URL; binary không đi xuyên FlashX API.
- Job state và audit/history được lưu backend; UI không phải nguồn sự thật duy nhất.
- Admin Operations có dữ liệu backend thật và tự refresh trong demo.

## 8. Phần cố ý chưa coi là production-final

Buổi demo này chứng minh sản phẩm và hệ thống vận hành end-to-end. Các mục sau vẫn cần hardening trước commercial Go-Live quy mô lớn:

- cấu hình Maps/Routes production và API key do chủ dự án sở hữu;
- SMS OTP production;
- payment gateway online ngoài cash-first MVP;
- push notification production;
- realtime hub đa instance qua Redis Pub/Sub khi scale ngang;
- monitoring/alerting/SLA production;
- rà soát pháp lý, bảo hiểm, hợp đồng tài xế và quy trình xử lý sự cố thực địa;
- một số chi tiết UX/BA có thể tiếp tục điều chỉnh sau buổi demo.

Không trình bày các mục trên như đã production-complete nếu môi trường demo chưa cấu hình nhà cung cấp bên thứ ba tương ứng.
