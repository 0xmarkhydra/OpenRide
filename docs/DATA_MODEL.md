# Data Model

> **Compatibility note 12/08/2026:** schema hiện tại phản ánh foundation ride-hailing cũ (`vehicles.driver_id`, `trips`, `car/bike`). Không được coi bảng `vehicles` hiện tại là “Xe của tôi”. Target Full Marketplace cần migration mới cho **phương tiện của khách**, 3 service mới, lịch hẹn, handover và workflow Đăng kiểm hộ; xem `IMPLEMENTATION_GAP_REVIEW_2026-08-12.md`.

## 1. Mục tiêu

PostgreSQL + PostGIS là source of truth cho dữ liệu nghiệp vụ. Redis giữ hot state và spatial index realtime. Thiết kế phải tránh ghi GPS tần suất cao trực tiếp vào bảng transactional chính.

## 2. Schema hiện có

Migration `001_init.sql` hiện tạo:
- `users`
- `drivers`
- `vehicles`
- `trips`
- `trip_status_history`
- PostGIS extension

## 3. users

```text
id UUID PK
phone VARCHAR UNIQUE
full_name VARCHAR
status VARCHAR
created_at TIMESTAMPTZ
updated_at TIMESTAMPTZ
```

Khuyến nghị bổ sung khi triển khai auth thật:
- phone_verified_at
- avatar_object_key
- last_login_at
- version hoặc optimistic locking field nếu cần.

## 4. drivers

```text
id UUID PK
phone VARCHAR UNIQUE
full_name VARCHAR
approval_status VARCHAR
availability_status VARCHAR
created_at
updated_at
```

Khuyến nghị bổ sung:
- current_vehicle_id hoặc mapping active vehicle;
- approved_at;
- suspended_at;
- rating aggregates;
- service region/city.

## 5. vehicles

```text
id UUID PK
driver_id UUID FK
service_type VARCHAR
plate_number VARCHAR
brand VARCHAR
model VARCHAR
created_at
```

Cần unique/index phù hợp cho plate number và driver lookup khi implement đầy đủ.

## 6. trips

Hiện dùng PostGIS `GEOGRAPHY(POINT,4326)` cho pickup/destination.

```text
id UUID PK
rider_id UUID FK
driver_id UUID nullable FK
service_type
status
pickup GEOGRAPHY POINT
destination GEOGRAPHY POINT
estimated_distance_m
estimated_duration_s
estimated_fare_minor
final_fare_minor
currency
created_at
accepted_at
started_at
completed_at
cancelled_at
```

### Money
Tất cả tiền lưu bằng integer minor unit, không dùng float.

Ví dụ VND có thể lưu trực tiếp số đồng trong trường `*_minor` theo convention thống nhất của dự án.

### Index hiện có
- GIST pickup
- GIST destination
- `(status, created_at DESC)`

## 7. trip_status_history

Mọi transition quan trọng ghi append-only:
- trip_id
- status
- actor_type
- created_at

Nên bổ sung trong migration sau:
- actor_id nullable
- reason_code
- metadata JSONB
- sequence/version.

## 8. Driver documents / object storage metadata

### driver_documents
- id
- driver_id
- document_type
- object_key — private S3-compatible object key, không phải public URL
- filename
- content_type
- size_bytes
- review_status
- review_note
- created_at/updated_at
- reviewed_at

File bytes không nằm trong PostgreSQL và không đi qua backend. Client upload trực tiếp tới object storage bằng presigned PUT; bảng này chỉ là business metadata/KYC review state.

## 9. Tables dự kiến tiếp theo

### pricing_rules
- id
- service_type
- city/zone
- base_fare
- per_km
- per_minute
- minimum_fare
- effective_from/effective_to
- version

### payments
- id
- trip_id
- provider
- external_reference
- amount
- currency
- status
- idempotency_key
- created_at/updated_at

### promotions
- id
- code
- type
- value
- max_discount
- starts_at/ends_at
- usage limits
- status

### ratings
- id
- trip_id
- rider_id
- driver_id
- score
- comment
- created_at

### admin_users / admin_roles
RBAC cho cổng vận hành.

### audit_logs
Append-only log cho thao tác nhạy cảm.

## 9. Redis key model đề xuất

### Online driver geo index
```text
geo:drivers:{city}:{service_type}
```
Member: `driver_id`

### Latest location
```text
driver:{driver_id}:location
```
Payload/hash:
- lat
- lng
- accuracy
- heading
- speed
- captured_at
- received_at

TTL/freshness policy phải được áp dụng để driver stale không xuất hiện trong dispatch.

### Availability
```text
driver:{driver_id}:state
```

### Dispatch lock
```text
lock:driver:{driver_id}
lock:trip:{trip_id}:assignment
```
Dùng atomic Redis operation/script hoặc database guard để chống race.

### Offer
```text
dispatch:trip:{trip_id}:offer:{driver_id}
```
Có TTL.

## 10. Location persistence

Không lưu từng GPS ping vào `trips`.

Các lựa chọn khi cần trip path:
- sample point theo thời gian/khoảng cách;
- append vào time-series table partitioned;
- encode polyline sau trip;
- chuyển sang analytics store khi scale.

MVP chỉ nên lưu lượng cần thiết cho audit/support.

## 11. Migration rules

- Migration là append-only; không sửa migration đã chạy production.
- Mọi schema change phải backward-compatible trong rolling deployment nếu có nhiều instance.
- Add nullable/default trước, backfill sau, enforce constraint cuối nếu table lớn.
- Index creation trên production lớn phải cân nhắc lock/concurrently.

## 12. Backup/retention

Production phải có automated backup và test restore định kỳ. Retention của KYC, location history và audit log phải được chốt theo yêu cầu pháp lý/chính sách doanh nghiệp trước launch.
