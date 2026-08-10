# Security & Privacy

## 1. Mục tiêu

Bảo vệ account, trip, payment, location và KYC data. Security phải nằm ở backend/runtime, không phụ thuộc frontend.

## 2. Authentication

- OTP/login endpoint rate limit theo phone/IP/device signals phù hợp.
- Access token short-lived; refresh/session revoke được.
- Admin authentication tách policy và có thể yêu cầu MFA trước production.
- Không ghi OTP/token vào logs.

## 3. Authorization

- Rider chỉ truy cập resource của mình.
- Driver chỉ truy cập offer/trip được assign cho mình.
- Admin API dùng RBAC.
- Mọi object lookup phải chống IDOR; không chỉ dựa vào UUID khó đoán.

## 4. Transport

- HTTPS/TLS bắt buộc production.
- WebSocket dùng WSS.
- Không gửi secret/PII nhạy cảm qua query string nếu tránh được.

## 5. Secrets

- Không commit `.env`, API key, database credential.
- Dev/Staging/Prod dùng secret riêng.
- Rotate secret khi lộ hoặc nhân sự/quyền truy cập thay đổi.
- Production dùng secret manager/cloud secret nếu có.

## 6. Password/OTP

Nếu sau này có password, hash bằng thuật toán hiện đại như Argon2id/bcrypt với config phù hợp. OTP phải có expiry, attempt limit và chống replay.

## 7. KYC documents

- Lưu ở private object storage.
- Truy cập qua signed URL ngắn hạn.
- Chỉ role được phép xem.
- Không public bucket.
- Có retention/delete policy.
- Audit document access nếu yêu cầu compliance.

## 8. Location privacy

Driver location là dữ liệu nhạy cảm.
- Rider chỉ nhận location driver trong trip liên quan.
- Admin chỉ truy cập theo quyền/operational need.
- Không public endpoint trả vị trí driver.
- Không lưu history chi tiết vô hạn nếu không có business/legal reason.

## 9. Payment

- Không tự lưu raw card data nếu payment gateway có hosted/tokenized flow.
- Verify webhook signature.
- Webhook idempotent.
- Amount/currency được verify server-side, không tin client.

## 10. Input validation

Validate:
- phone;
- UUID;
- enum/state;
- lat/lng;
- timestamps;
- file mime/size;
- promo code;
- monetary values.

Không interpolate user input vào SQL; dùng parameterized query/query builder.

## 11. Rate limiting

Áp dụng theo endpoint risk:
- OTP request/verify;
- login/refresh;
- trip creation;
- location ingestion theo driver/session;
- admin search/export;
- payment/webhook abuse protection.

## 12. Idempotency & replay

Command tài chính/assignment phải có idempotency. Signed callback/webhook phải kiểm tra timestamp/signature/replay theo provider.

## 13. Audit logging

Audit log không chứa secret nhưng cần:
- actor;
- action;
- target;
- timestamp;
- request_id;
- reason/metadata an toàn.

## 14. Logging hygiene

Không log:
- access/refresh token;
- OTP;
- API secret;
- full payment credentials;
- raw KYC document;
- location history nhiều hơn cần thiết.

Phone/email có thể mask trong logs tùy use case.

## 15. Mobile security

- Secure storage cho token.
- Key restriction cho Maps mobile key.
- Không embed backend secret.
- Detect rooted/jailbroken device chỉ là signal, không phải security boundary chính.

## 16. Admin security

- MFA trước production khuyến nghị mạnh.
- RBAC least privilege.
- Sensitive action audit.
- Session timeout.
- Có thể IP allowlist/VPN nếu mô hình vận hành cho phép.

## 17. Dependency/security updates

- Pin/lock dependency.
- Automated dependency alerts.
- Patch critical CVE theo SLA nội bộ.
- Container/base image scan trước production nếu pipeline hỗ trợ.

## 18. Privacy/legal

Chính sách thu thập, lưu, chia sẻ và xóa dữ liệu phải được doanh nghiệp/pháp lý chốt trước launch. Tech implementation phải hỗ trợ policy đó; docs này không thay thế tư vấn pháp lý.

## 19. Incident basics

Nếu nghi credential/data leak:
1. revoke/rotate credential;
2. isolate affected component;
3. preserve logs/evidence;
4. assess scope;
5. patch root cause;
6. follow notification/legal process của doanh nghiệp.
