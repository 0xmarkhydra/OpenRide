# API Contract

> **Compatibility note 12/08/2026:** API đang chạy vẫn dùng nhiều endpoint/field `rider`, `trip`, `car/bike` từ MVP ride-hailing cũ. Đây là **implementation compatibility contract**, không phải business semantics cuối cùng. API mới phải migrate theo `CUSTOMER_REQUIREMENTS.md` / `PRD_MVP.md`: 3 service mới, `CustomerVehicle`, scheduled booking, handover và workflow Đăng kiểm hộ. Không mở rộng semantics `car/bike` cũ cho feature mới.

## 1. Nguyên tắc

- Base path: `/v1`.
- JSON UTF-8.
- HTTPS bắt buộc trên production.
- Auth: Bearer token cho Rider/Driver/Admin; cơ chế token cụ thể sẽ được ADR hóa khi implement.
- Tất cả timestamp dùng ISO-8601 UTC ở API.
- Money dùng integer, không dùng floating point.
- Endpoint command quan trọng hỗ trợ idempotency khi cần.

## 2. Response envelope

Success có thể trả resource trực tiếp hoặc envelope thống nhất. Khuyến nghị:

```json
{
  "data": {},
  "meta": {}
}
```

Error:

```json
{
  "error": {
    "code": "TRIP_INVALID_STATE",
    "message": "Trip cannot be started from current state",
    "request_id": "req_...",
    "details": {}
  }
}
```

Client logic dựa trên `code`, không parse `message`.

## 3. HTTP status convention

- `200` success query/update.
- `201` created.
- `204` success no body.
- `400` malformed/validation.
- `401` unauthenticated.
- `403` authenticated nhưng không có quyền.
- `404` resource not found.
- `409` state conflict/idempotency/assignment conflict.
- `422` business rule violation nếu cần phân biệt validation.
- `429` rate limited.
- `500` unexpected server error.
- `503` dependency unavailable/degraded.

## 4. Health

### GET `/health`
Không auth. Chỉ báo process sống.

### GET `/ready`
Dành cho readiness; kiểm tra dependency quan trọng khi triển khai production.

## 5. Auth MVP

### POST `/v1/auth/otp/request`
Input:
```json
{"phone":"+84...","purpose":"login"}
```

### POST `/v1/auth/otp/verify`
Input:
```json
{"phone":"+84...","code":"123456","role":"rider"}
```
Output: access/refresh token hoặc session equivalent.

### POST `/v1/auth/refresh`
Refresh session/token.

## 6. Rider

### GET `/v1/me`
Current rider profile.

### PATCH `/v1/me`
Update profile fields được phép.

### POST `/v1/trips/estimate`
Input:
```json
{
  "pickup":{"lat":21.0,"lng":105.8},
  "destination":{"lat":21.1,"lng":105.9},
  "service_type":"bike",
  "promotion_code":null
}
```
Output gồm distance, duration, route summary và fare breakdown/version.

### POST `/v1/trips`
Tạo trip từ pickup/destination/service type. Nên nhận estimate/pricing token hoặc recompute server-side để tránh client sửa giá.

Header đề xuất:
`Idempotency-Key: <uuid>`

### GET `/v1/trips/{id}`
Snapshot state của trip.

### POST `/v1/trips/{id}/cancel`
Body gồm reason code.

### GET `/v1/trips`
Trip history với cursor pagination.

### POST `/v1/trips/{id}/rating`
Rating sau completed trip.

## 7. Driver

### GET `/v1/driver/me`
Driver profile/KYC/availability.

### PATCH `/v1/driver/me`
Update profile.

### POST `/v1/driver/documents/upload-url`
Xin presigned PUT request cho một file KYC. Backend chỉ ký request; file bytes không đi qua FlashX API.

### POST `/v1/driver/documents/complete`
Sau khi client PUT trực tiếp lên object storage, ghi metadata/object key vào PostgreSQL.

### GET `/v1/driver/documents`
Danh sách metadata KYC của tài xế hiện tại.

### GET `/v1/driver/documents/{id}/view-url`
Cấp presigned GET URL ngắn hạn sau khi kiểm tra ownership.

### POST `/v1/driver/availability`
```json
{"status":"online"}
```
Server kiểm tra approved + active vehicle trước khi cho online.

### POST `/v1/driver/offers/{offer_id}/accept`
Atomic accept. Có thể trả `409 OFFER_EXPIRED` hoặc `409 DRIVER_ALREADY_BUSY`.

### POST `/v1/driver/offers/{offer_id}/reject`
Reject với reason optional.

### POST `/v1/driver/trips/{id}/arrived`
Transition ARRIVING/ACCEPTED -> ARRIVED theo state model.

### POST `/v1/driver/trips/{id}/start`
ARRIVED -> IN_PROGRESS.

### POST `/v1/driver/trips/{id}/complete`
IN_PROGRESS -> COMPLETED; backend chốt final fare.

## 8. Location ingestion

Ưu tiên WebSocket cho live session; HTTP endpoint có thể làm fallback:

### POST `/v1/driver/location`
```json
{
  "lat":21.0285,
  "lng":105.8542,
  "accuracy_m":8.5,
  "heading_deg":180,
  "speed_mps":8.2,
  "captured_at":"2026-08-10T10:00:00Z"
}
```

Server phải reject point quá cũ/vô lý theo policy.

## 9. Admin

Prefix: `/v1/admin`

Các endpoint nhóm:
- `/riders`
- `/drivers`
- `/drivers/{id}/approval`
- `/drivers/{id}/documents`
- `/drivers/{id}/documents/{documentID}/review`
- `/trips`
- `/trips/{id}`
- `/pricing-rules`
- `/promotions`
- `/audit-logs`
- `/dashboard`

Admin API bắt buộc RBAC và audit action nhạy cảm.

## 10. Pagination

Ưu tiên cursor pagination:
```json
{
  "data": [],
  "meta": {"next_cursor":"..."}
}
```

Không dùng offset cho stream/table lớn nếu có thể tránh.

## 11. Idempotency

Áp dụng tối thiểu cho:
- create trip;
- accept offer;
- payment create/capture/refund;
- provider webhook processing.

Idempotency record phải bind user/action/body hash trong TTL hoặc durable storage phù hợp.

## 12. API versioning

Breaking change tạo version mới hoặc cơ chế compatibility. Không đổi field semantics âm thầm khi Rider/Driver app production chưa bắt buộc update.

## 13. OpenAPI

Khi endpoint được implement, API contract machine-readable phải được xuất thành OpenAPI và xem là source gần code. File này mô tả business-level contract; OpenAPI mô tả schema chi tiết.
