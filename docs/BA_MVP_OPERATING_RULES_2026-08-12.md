# FlashX — BA vận hành 3 dịch vụ MVP

> **Ngày khóa BA:** 12/08/2026  
> **Vai trò:** Baseline nghiệp vụ để Product/Design/Engineering/Operations cùng triển khai.  
> **Ưu tiên:** 3 dịch vụ hiện tại phải chạy ổn ngoài đời trước; kiến trúc vẫn phải chừa đường cho gọi xe, xe ghép và các dịch vụ mobility sau này.

## 1. Mục tiêu tài liệu

Tài liệu này chốt các quyết định nghiệp vụ mặc định để đội phát triển không phải dừng lại hỏi Founder ở từng tình huống nhỏ.

Nguyên tắc làm việc:
- BA chủ động chọn phương án hợp lý nhất cho MVP.
- Chỉ đưa lên Founder khi quyết định ảnh hưởng lớn tới pháp lý, bảo hiểm, mức phí/commission, vốn hoặc mô hình vận hành.
- Không vì chuẩn bị cho tương lai mà làm 3 dịch vụ hiện tại khó dùng hoặc khó vận hành.
- Không rewrite foundation kỹ thuật đang chạy nếu có thể migrate từng bước.

## 2. Ba dịch vụ MVP

1. `designated_driver_car` — Lái hộ ô tô.
2. `designated_driver_bike` — Lái hộ xe máy.
3. `vehicle_inspection_assist` — Đăng kiểm hộ.

Trong MVP, phương tiện thuộc khách hàng. Gọi xe kiểu Grab/Uber, xe của tài xế/fleet và xe ghép chưa bật ở production MVP.

## 3. Các quyết định BA mặc định đã chốt

### 3.1 Tài xế tự chủ động việc di chuyển tới khách và rời điểm cuối

MVP không tổ chức đội xe riêng để chở tài xế.

- Tài xế tự chọn cách phù hợp để đến điểm nhận: xe cá nhân, xe điện gấp, phương tiện công cộng, xe ôm công nghệ hoặc cách hợp pháp khác.
- Sau khi bàn giao xe cho khách, tài xế tự chủ động rời điểm cuối.
- Chi phí di chuyển hợp lý phải được phản ánh trong giá/thu nhập của job theo pricing policy.
- Tài xế không được tự xin thêm tiền khách ngoài số tiền hiển thị/được khách chấp thuận trên FlashX.
- Offer cần hiển thị đủ pickup, destination/return point và thu nhập dự kiến để tài xế tự quyết định có nhận job hay không.

### 3.2 Lái hộ xe máy mặc định là đưa cả khách và xe về

Flow mặc định của MVP:

```text
Tài xế tới khách
→ nhận xe máy của khách
→ khách đi cùng trên chính xe của mình
→ tài xế lái tới điểm đến
→ bàn giao xe và kết thúc dịch vụ
```

Case “chỉ đưa xe, khách không đi cùng” chưa phải flow mặc định; data model không được chặn để có thể thêm sau.

Trước khi bắt đầu, tài xế có quyền từ chối nếu điều kiện thực tế gây mất an toàn: khách mất kiểm soát, xe không an toàn để vận hành, số người chở không hợp lệ hoặc tình huống khác trái quy định.

### 3.3 Người đặt không bắt buộc phải là người trực tiếp giao/nhận xe

MVP hỗ trợ:
- chính khách đặt dịch vụ; hoặc
- người được khách chỉ định.

Thông tin người được chỉ định tối thiểu:
- tên;
- số điện thoại;
- vai trò giao xe/nhận xe;
- mã xác nhận khi policy yêu cầu.

Không được bàn giao xe cho người không khớp thông tin mà không có xác nhận của khách/Operations.

### 3.4 Giao nhận xe có bằng chứng ngay trong MVP

Evidence không để hoàn toàn sang P1 nữa vì đây là tài sản có giá trị.

MVP tối thiểu tại lúc nhận và trả xe:
- ảnh tổng quan phương tiện;
- ảnh khu vực đồng hồ/odometer hoặc dashboard khi phù hợp;
- ghi chú tình trạng bất thường nếu có;
- xác nhận người giao/nhận;
- thời gian + vị trí của mốc bàn giao;
- audit actor thực hiện.

Ảnh/video upload trực tiếp object storage; backend chỉ cấp signed URL và lưu metadata, tương tự nguyên tắc KYC.

Không bắt khách/tài xế nhập quá nhiều trường gây chậm thao tác. Evidence chi tiết theo từng góc xe có thể cấu hình thêm sau pilot.

### 3.5 Sau `VEHICLE_RECEIVED` không còn hủy job thông thường

Trước khi tài xế nhận xe, customer cancellation có thể áp dụng theo policy/phí từng giai đoạn.

Sau khi tài xế đã nhận xe:
- không hiển thị “Hủy chuyến” như một booking bình thường;
- nếu khách đổi ý, tạo yêu cầu đưa xe trở lại/điểm bàn giao mới hoặc liên hệ Operations;
- job chuyển qua luồng return/support/incident phù hợp;
- các chi phí thực tế phát sinh được xử lý theo policy.

Lý do: từ thời điểm nhận xe, FlashX đang quản lý trách nhiệm với tài sản của khách.

### 3.6 Phí phát sinh phải minh bạch và có phê duyệt

Các khoản có thể phát sinh:
- cầu đường;
- bãi xe;
- xăng/sạc;
- phí tại trung tâm đăng kiểm;
- chi phí hợp lý khác.

Quy tắc:
- khoản đã biết trước phải nằm trong quote hoặc mô tả giá;
- khoản ngoài quote phải tạo yêu cầu phát sinh;
- khách được thấy số tiền + lý do và chọn đồng ý/từ chối;
- tài xế không tự sửa giá cuối hoặc tự thu thêm ngoài hệ thống;
- trường hợp khẩn cấp để bảo vệ người/tài sản có thể do Operations xử lý trước rồi ghi audit, sau đó đối soát theo chính sách.

### 3.7 Đăng kiểm “không đạt” không đồng nghĩa FlashX “không hoàn thành dịch vụ”

Tách hai khái niệm:

- `service_status`: FlashX đã hoàn thành công việc nhận xe → thực hiện → trả xe hay chưa.
- `inspection_result`: kết quả kiểm định của phương tiện.

Kết quả đăng kiểm tối thiểu:
- `passed`;
- `failed`;
- `deferred`;
- `unavailable`.

Ví dụ: FlashX đã đưa xe tới trung tâm, xe bị đánh giá không đạt, sau đó FlashX đưa xe về và bàn giao đầy đủ → `service_status = completed`, `inspection_result = failed`.

MVP không tự động biến một job đăng kiểm thất bại thành job sửa xe. Khách có thể tạo yêu cầu tiếp theo khi business mở dịch vụ đó.

### 3.8 FlashX chủ động đề xuất nơi đăng kiểm phù hợp

MVP:
- khách có thể ghi trung tâm mong muốn;
- FlashX/Operations có thể đề xuất hoặc chọn trung tâm phù hợp theo khu vực, khả năng tiếp nhận và vận hành thực tế;
- không cam kết trung tâm cụ thể nếu chưa có xác nhận;
- hệ thống không phụ thuộc cứng vào một app/API đặt lịch của bên thứ ba.

Thông tin trung tâm phải là dữ liệu có thể thay đổi, không hard-code trong mobile.

### 3.9 Sự cố là một luồng riêng, không phải “Cancel”

Các sự cố tối thiểu cần hỗ trợ:
- tai nạn;
- xe hỏng;
- thủng lốp;
- hết xăng/pin;
- mất chìa khóa;
- mất/hỏng giấy tờ;
- tranh chấp tình trạng xe;
- khách/tài xế mất liên lạc;
- khách gây mất an toàn;
- sự cố tại trung tâm đăng kiểm.

Khi báo sự cố:
1. lưu loại sự cố;
2. lưu vị trí/thời gian;
3. cho phép đính kèm ảnh/evidence;
4. thông báo Operations;
5. job không tiếp tục tự động theo flow bình thường nếu sự cố đang mở;
6. Operations quyết định tiếp tục, trả xe, thay người hoặc đóng sự cố;
7. mọi action nhạy cảm phải audit.

### 3.10 Không tự động đổi tài xế sau khi đã nhận xe

- Trước `VEHICLE_RECEIVED`: có thể reassign theo policy.
- Sau `VEHICLE_RECEIVED`: không được đổi `assigned_driver` trực tiếp.
- Nếu bắt buộc thay người, phải tạo một lần bàn giao có xác nhận, evidence và audit trước khi người mới chịu trách nhiệm.

### 3.11 Tiền khách trả khác tiền tài xế kiếm được

MVP phải tách tối thiểu:
- tổng tiền khách phải trả;
- khoản tài xế được hưởng;
- phần FlashX;
- khoản phát sinh/hoàn tiền/điều chỉnh nếu có.

Không được dùng `final_fare` làm thẳng “thu nhập tài xế”.

Với cash-first MVP:
- khách có thể trả số tiền hiển thị trên app theo policy;
- hệ thống vẫn phải ghi nhận đầy đủ gross charge và driver earning;
- đối soát commission/settlement có thể vận hành bán thủ công ở pilot nhưng dữ liệu phải tách đúng từ đầu.

### 3.12 Pilot theo vùng nhỏ, không mở toàn tỉnh ngay

Khuyến nghị BA hiện tại:
- pilot Thanh Hóa;
- ưu tiên vùng Hạc Thành trước;
- mở sang Sầm Sơn khi supply/ETA/support đủ tốt;
- chỉ mở thêm vùng khi đạt ngưỡng vận hành nội bộ.

Ứng dụng chỉ nên cho đặt ở vùng thật sự có khả năng phục vụ. Không quảng cáo phủ toàn tỉnh trong khi supply còn mỏng.

## 4. Quy tắc hủy và chờ — baseline MVP

Không khóa số tiền cứng trong code. Mọi mức phí do pricing config quản lý.

Các giai đoạn nghiệp vụ:

1. `SEARCHING`: khách có thể hủy; thường không/phí thấp tùy policy.
2. `ACCEPTED/ARRIVING`: có thể tính cancellation fee vì tài xế đã di chuyển.
3. `ARRIVED`: có grace period chờ khách, sau đó có waiting fee/cancellation fee theo config.
4. `VEHICLE_RECEIVED` trở đi: không còn normal cancellation; chuyển return/support/incident.

Default BA cho grace period pilot: **10 phút**, nhưng phải là cấu hình backend/Admin, không hard-code mobile.

## 5. Luồng chuẩn — Lái hộ ô tô

```text
Customer chọn dịch vụ
→ chọn ô tô của mình
→ pickup
→ destination
→ ngay/hẹn giờ
→ quote
→ xác nhận
→ SEARCHING
→ offer cho tài xế phù hợp
→ ACCEPTED
→ ARRIVING
→ ARRIVED
→ kiểm tra/evidence xe
→ xác nhận nhận xe
→ VEHICLE_RECEIVED
→ bắt đầu dịch vụ
→ IN_PROGRESS
→ tới điểm bàn giao
→ evidence trả xe
→ HANDOVER
→ thanh toán/đối soát
→ COMPLETED
→ đánh giá
```

### Không được bỏ qua
- ownership xe;
- capability số sàn/tự động;
- GPLX còn hiệu lực;
- GPS tài xế còn fresh;
- evidence nhận/trả;
- tracking/reconnect;
- support/incident.

## 6. Luồng chuẩn — Lái hộ xe máy

```text
Customer chọn dịch vụ
→ chọn xe máy
→ pickup
→ destination
→ ngay/hẹn giờ
→ quote
→ xác nhận
→ SEARCHING
→ ACCEPTED
→ ARRIVING
→ ARRIVED
→ kiểm tra/evidence xe
→ xác nhận nhận xe
→ VEHICLE_RECEIVED
→ khách đi cùng trên xe theo flow mặc định
→ IN_PROGRESS
→ tới điểm đến
→ HANDOVER
→ thanh toán
→ COMPLETED
→ đánh giá
```

### Case phải xử lý
- xe số/ga/côn tay/xe điện theo capability;
- điều kiện khách không an toàn để chở;
- số người vượt quy định;
- thiếu mũ bảo hiểm/điều kiện an toàn;
- xe hỏng/hết pin/xăng;
- trời mưa/điều kiện thời tiết bất thường theo Operations policy;
- người nhận khác người đặt.

## 7. Luồng chuẩn — Đăng kiểm hộ

```text
Customer chọn ô tô
→ địa chỉ nhận
→ thời gian mong muốn
→ checklist giấy tờ
→ quote/package
→ xác nhận
→ SCHEDULED hoặc SEARCHING
→ ACCEPTED
→ ARRIVING_FOR_PICKUP
→ ARRIVED_FOR_PICKUP
→ evidence + nhận xe/giấy tờ
→ VEHICLE_RECEIVED
→ EN_ROUTE_TO_INSPECTION
→ ARRIVED_AT_INSPECTION_CENTER
→ INSPECTION_IN_PROGRESS
→ INSPECTION_COMPLETED
→ lưu inspection_result
→ RETURNING_VEHICLE
→ ARRIVED_FOR_RETURN
→ evidence + trả xe/giấy tờ
→ HANDOVER
→ COMPLETED
```

### Checklist đăng kiểm phải là dữ liệu cấu hình
Không hard-code danh sách giấy tờ trong app vì quy định có thể thay đổi.

Mỗi item cần tối thiểu:
- loại/nhãn giấy tờ;
- bắt buộc hay không;
- đã nhận;
- đã trả;
- ghi chú.

## 8. State và trách nhiệm tài sản

Ranh giới quan trọng nhất:

```text
Trước VEHICLE_RECEIVED
= FlashX đang thực hiện marketplace/điều phối

Từ VEHICLE_RECEIVED đến HANDOVER
= FlashX/tài xế đang có trách nhiệm trực tiếp với tài sản khách

Sau HANDOVER
= trách nhiệm bàn giao vật lý đã kết thúc, nhưng payment/rating/support có thể còn mở
```

Mọi query xác định “driver đang bận” phải coi toàn bộ giai đoạn custody là active/occupied.

Không được có trường hợp tài xế đang giữ xe khách nhưng hệ thống lại cấp thêm job mới.

## 9. Driver eligibility

Tài xế chỉ được nhận offer khi đồng thời:
- account/KYC approved;
- không suspended;
- GPLX còn hiệu lực;
- online;
- GPS đủ mới;
- không bận theo policy;
- có capability đúng service;
- phù hợp loại xe/transmission;
- nằm trong vùng hoạt động;
- không có incident/open restriction chặn nhận job.

Đăng kiểm hộ yêu cầu tối thiểu khả năng lái ô tô phù hợp + capability hỗ trợ đăng kiểm, không chỉ một flag `inspection` đơn lẻ.

## 10. Realtime và mất mạng

Mất mạng/app background/app restart được coi là tình huống bình thường phải thiết kế.

Flow phục hồi:
1. phát hiện WebSocket mất kết nối;
2. reconnect theo backoff;
3. refresh session nếu cần;
4. lấy snapshot active job từ backend;
5. đồng bộ version/state;
6. tiếp tục subscribe location/status.

Backend snapshot là nguồn sự thật. Không giả định client đã nhận đủ mọi realtime event.

## 11. Admin/Operations bắt buộc cho pilot

Operations phải xem được:
- job đang searching quá lâu;
- scheduled job sắp tới giờ nhưng chưa có tài xế;
- GPS stale;
- tài xế tới lâu nhưng chưa nhận xe;
- đã nhận xe nhưng chưa start;
- job in-progress quá lâu;
- đăng kiểm tại trung tâm quá lâu;
- xe đang trả nhưng mất GPS;
- incident đang mở;
- bàn giao chưa hoàn tất;
- payment chưa xử lý.

Operations được can thiệp có kiểm soát, mọi action nhạy cảm phải có reason + actor + timestamp + audit.

## 12. Các case bắt buộc phải test trước UAT sign-off

### Case chung
- no driver found;
- offer reject/timeout;
- hai tài xế accept đồng thời;
- driver accept rồi hủy;
- customer hủy theo từng giai đoạn;
- customer không xuất hiện;
- duplicate create request/retry;
- app kill/reopen;
- mất mạng giữa job;
- GPS stale;
- scheduled job tới giờ chưa match;
- payment retry;
- Admin can thiệp;
- incident mở/đóng.

### Lái hộ
- xe không đúng owner;
- sai loại xe;
- số sàn nhưng driver không có capability;
- xe không đủ điều kiện vận hành;
- khách đổi destination trước/sau nhận xe;
- xe hỏng/tai nạn;
- người nhận xe khác người đặt;
- bàn giao không thành công;
- tranh chấp tình trạng xe.

### Đăng kiểm
- thiếu giấy tờ;
- trung tâm không nhận/đóng cửa;
- phải đổi trung tâm;
- inspection passed;
- inspection failed;
- inspection deferred/unavailable;
- phát sinh khoản tiền cần khách duyệt;
- khách từ chối khoản phát sinh;
- mất/hỏng giấy tờ;
- không liên lạc được người nhận lúc trả xe.

## 13. Definition of Done nghiệp vụ cho 3 dịch vụ

Một service chỉ được coi là sẵn sàng pilot khi:
- happy path chạy E2E trên thiết bị thật;
- các case bắt buộc ở mục 12 có kết quả rõ;
- không có đường state bỏ qua custody/handover;
- reconnect phục hồi đúng active job;
- giá backend-authoritative;
- tài xế chỉ nhận job phù hợp capability;
- evidence nhận/trả lưu được;
- incident có thể báo và Operations nhìn thấy;
- payment/history/rating lưu đúng;
- Admin cứu được job kẹt mà không phá audit;
- không có Blocker/Critical bug.

## 14. Ràng buộc để mở rộng về sau nhưng không làm phức tạp MVP

MVP chưa bật gọi xe/xe ghép, nhưng code mới không được khóa cứng các giả định sau:
- service luôn bằng loại phương tiện;
- tài xế chỉ có một service type;
- mọi job luôn chỉ có pickup + destination;
- customer luôn là người trực tiếp sử dụng/nhận tài sản;
- `final_fare` luôn bằng driver earning;
- mọi service luôn dùng xe của khách;
- future shared ride luôn là 1 booking = 1 physical run.

Kiến trúc dài hạn hướng tới tách:
- nhu cầu khách (`Booking`);
- quá trình thực hiện thực tế (`Fulfillment`);
- stops/legs;
- provider/vehicle ownership;
- capacity/reservation cho shared ride.

**Nhưng các abstraction tương lai chỉ được triển khai sâu sau khi 3 service MVP ổn.** Giai đoạn hiện tại chỉ cần tránh những quyết định khiến sau này phải rewrite toàn bộ.

## 15. Những việc vẫn cần Founder/Legal/Finance phê duyệt trước production

BA tự quyết phần luồng sản phẩm, nhưng các mục sau không được coi là quyết định cuối nếu chưa có người có thẩm quyền duyệt:
- điều khoản trách nhiệm/bồi thường khi tai nạn hoặc hư hỏng tài sản;
- mô hình bảo hiểm/quỹ đảm bảo;
- commission chính thức và payout/settlement;
- bảng giá cụ thể từng service;
- mức cancellation/waiting/night/holiday fee cụ thể;
- yêu cầu giấy tờ pháp lý chính thức của tài xế;
- hợp đồng tài xế/đối tác;
- quy trình ủy quyền/giấy tờ đăng kiểm theo tư vấn pháp lý hiện hành;
- phạm vi vùng pilot cuối cùng và ngày Go-Live.

---

**Quy tắc triển khai:** Nếu code/UI/API khác với các rule trong tài liệu này, đội phát triển phải coi đó là gap cần sửa hoặc phải cập nhật lại BA bằng một quyết định mới có lý do rõ ràng.