# Third-party Services & Cost Ownership

## 1. Nguyên tắc thương mại

Chi phí phát triển phần mềm và chi phí vận hành bên thứ ba là hai nhóm khác nhau.

Ví dụ: hạng mục “tích hợp bản đồ” là công dev để tích hợp SDK/API, hiển thị map, route, geocoding, ETA. Phí Google Maps/Mapbox/HERE thu theo usage là operating cost của doanh nghiệp và không nằm trong phí dev trừ khi hợp đồng ghi rõ.

## 2. Maps

Provider có thể là:
- Google Maps Platform;
- Mapbox;
- HERE;
- provider khác được chấp thuận.

Các API có thể phát sinh phí:
- Maps SDK;
- Places/Autocomplete;
- Geocoding;
- Routes/Directions;
- Distance Matrix;
- Navigation SDK.

Ownership khuyến nghị:
- tài khoản/billing: Khách hàng/doanh nghiệp;
- kỹ thuật tích hợp/config: đội phát triển;
- usage invoice: Khách hàng/doanh nghiệp.

## 3. SMS/OTP

Phí phụ thuộc provider và số OTP gửi.

Đội phát triển:
- tích hợp;
- rate limit;
- retry/failure handling.

Doanh nghiệp:
- tài khoản provider;
- số dư/billing;
- sender/brandname nếu cần.

## 4. Push notification

FCM/APNs thường là hạ tầng push chính, nhưng các dịch vụ bổ sung/analytics có thể có phí riêng theo chính sách tại thời điểm sử dụng.

## 5. Payment Gateway

Có thể phát sinh:
- transaction fee;
- settlement fee;
- refund fee;
- chargeback/dispute fee;
- các khoản theo hợp đồng provider.

Đội phát triển chỉ tích hợp technical flow; hợp đồng merchant và phí giao dịch thuộc doanh nghiệp.

## 6. Cloud

Các nhóm chi phí:
- compute/container;
- managed PostgreSQL;
- managed Redis;
- object storage;
- bandwidth/egress;
- load balancer;
- monitoring/logging;
- backup;
- CDN/WAF.

Cloud cost tăng theo traffic và retention; không thể xem là giá cố định của app development.

## 7. App stores

Doanh nghiệp nên sở hữu:
- Apple Developer account;
- Google Play Console account.

Phí account/store và thuế/liên quan store không mặc định nằm trong hợp đồng development.

## 8. Object storage

KYC/ảnh cần private storage. Cost phụ thuộc:
- GB stored;
- requests;
- egress;
- retention/versioning.

## 9. Monitoring

Sentry, Datadog, Grafana Cloud hoặc provider khác có thể miễn phí ở mức nhỏ và tính phí khi scale. Phải budget riêng khi chọn tool.

## 10. Ownership matrix

| Hạng mục | Dev tích hợp | Tài khoản Production | Phí usage |
|---|---:|---|---|
| Maps | Có | Khách hàng | Khách hàng |
| SMS/OTP | Có | Khách hàng | Khách hàng |
| Payment | Có | Khách hàng | Khách hàng |
| Cloud | Có cấu hình | Khách hàng | Khách hàng |
| Object Storage | Có | Khách hàng/cloud project | Khách hàng |
| Apple/Google Store | Hỗ trợ submit | Khách hàng | Khách hàng |
| Monitoring | Có cấu hình | Theo thỏa thuận | Khách hàng/ops budget |

## 11. Cost controls kỹ thuật

- quota/budget alert;
- Maps autocomplete debounce;
- cache hợp lệ;
- adaptive GPS frequency;
- log retention;
- autoscaling upper bound;
- image compression;
- signed direct upload;
- query/index optimization.

## 12. Báo giá MVP hiện tại

Tài liệu thương mại đã thống nhất sơ bộ tổng development budget khoảng `650.000.000 VNĐ` cho scope MVP đầy đủ. Đây là planning figure, không thay thế hợp đồng/quotation ký chính thức.

Chi phí third-party không nằm trong con số trên trừ khi phụ lục/hợp đồng ghi rõ.

## 13. Nguyên tắc khi báo khách

Luôn ghi rõ:

> Phí phát triển/tích hợp dịch vụ bên thứ ba đã nằm trong scope tương ứng; phí API, cloud, transaction, SMS, store và license do nhà cung cấp bên thứ ba thu theo usage/chính sách riêng không bao gồm trong tổng chi phí phát triển phần mềm.
