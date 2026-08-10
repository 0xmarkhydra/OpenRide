# FlashX System Architecture

## 1. Kiến trúc Phase 1

Phase 1 dùng Go modular monolith với module boundary rõ ràng. Mục tiêu là tốc độ phát triển cao nhưng vẫn có đường tách service khi tải thực tế yêu cầu.

```text
Rider Flutter -----------\
                         \
Driver Flutter ------------> Go API + WebSocket
                         /        |
Admin Next.js -----------/        +--> PostgreSQL + PostGIS
                                  +--> Redis GEO / Cache / Locks
                                  +--> Object Storage
                                  +--> Background workers
                                  |
                                  +--> Maps / SMS / Push / Payment
```

## 2. Module boundaries

### auth
Authentication, token/session, OTP orchestration, authorization primitives.

### users
Rider/customer profile, account status.

### drivers
Driver profile, KYC state, vehicle, availability.

### trips
Trip aggregate, lifecycle/state machine, history.

### dispatch
Candidate search, filtering, scoring, offer, timeout, assignment lock.

### location
GPS ingestion, Redis GEO update, freshness, trip location fan-out.

### pricing
Fare estimate, final fare, service type, promotion hooks.

### payments
Payment provider abstraction, payment transaction state, webhook handling.

### notifications
Push/SMS/email orchestration and delivery jobs.

### admin
Admin-only use cases, RBAC và operational actions.

## 3. Data ownership

### PostgreSQL/PostGIS
Dùng cho dữ liệu cần bền vững:
- riders;
- drivers;
- vehicles;
- trips;
- trip status history;
- pricing config;
- payment records;
- promotions;
- audit records.

### Redis
Dùng cho state thay đổi nhanh:
- online driver geo index;
- latest driver location;
- current driver availability;
- dispatch locks;
- trip/socket ephemeral state;
- cache;
- rate limit/idempotency ngắn hạn khi phù hợp.

Redis không phải source of truth cho lịch sử nghiệp vụ lâu dài.

### Object Storage
Dùng cho:
- KYC documents;
- avatars;
- vehicle images;
- support attachments.

## 4. Request flow — create trip

```text
Rider
  -> REST POST /v1/trips/estimate
  -> Maps provider route/distance
  -> Pricing module
  <- Estimate

Rider
  -> REST POST /v1/trips
  -> Trip persisted: SEARCHING
  -> Dispatch starts
  -> Redis GEO candidate lookup
  -> Offer driver via realtime/push
  -> Driver accepts
  -> Atomic assignment lock
  -> Trip: ACCEPTED
  -> Notify Rider
```

## 5. Location flow

```text
Driver GPS
  -> Driver App
  -> WebSocket/HTTP location ingestion
  -> validate timestamp/accuracy
  -> Redis latest location + GEO index
  -> if active trip: fan-out location event to Rider room
  -> periodically/sample durable trip path if required
```

Không ghi mọi GPS point trực tiếp vào bảng chính của PostgreSQL.

## 6. Consistency strategy

- Strong consistency ở assignment/payment transition quan trọng.
- Eventual consistency được chấp nhận cho map marker/location display.
- Trip state transition phải transactional hoặc guarded bằng compare-and-set/version.
- Dispatch phải có distributed lock/idempotent accept.

## 7. Background processing

Worker xử lý các tác vụ không cần block request:
- push notifications;
- SMS;
- receipt;
- analytics event;
- cleanup expired offers;
- retry provider callbacks;
- settlement/earning computation sau này.

MVP có thể dùng Redis-backed queue. Khi throughput/event topology yêu cầu mới cân nhắc NATS/Kafka.

## 8. API style

- REST cho command/query thông thường.
- WebSocket cho realtime trip/location events.
- Webhook cho Payment provider và third-party callback.
- OpenAPI/Swagger cho REST API.

## 9. Failure handling

### Redis unavailable
- Không tạo dispatch mới nếu không đảm bảo assignment safety.
- Existing trip phải degrade an toàn; location có thể tạm stale.

### Maps provider unavailable
- Estimate/route-dependent booking có thể fail rõ ràng hoặc dùng cached/static fallback theo policy.

### WebSocket disconnect
- Mobile reconnect với exponential backoff.
- Sau reconnect phải query snapshot state từ REST trước/hoặc sync event sequence.

### Payment provider failure
- Trip completion và payment state tách biệt.
- Webhook/idempotency đảm bảo không double-charge.

## 10. Scale path

Không tách service vì “trông enterprise”. Chỉ extract khi có ít nhất một lý do rõ:
- module cần scale độc lập;
- deployment cadence độc lập;
- ownership/team boundary;
- resource profile rất khác;
- measured bottleneck.

Thứ tự có khả năng:
1. Location Service.
2. Dispatch Service.
3. Notification Worker Service.
4. Payment Service.
5. Analytics pipeline.

## 11. Go vs Rust

Backend chính dùng Go để tối ưu tốc độ phát triển, concurrency và khả năng tuyển/mở rộng team. Rust chỉ được cân nhắc cho measured hot-path như custom routing, dense geospatial matching hoặc stream processor khi profiling chứng minh cần thiết.

## 12. Nguyên tắc tránh over-engineering

Phase 1 không mặc định dùng:
- Kubernetes;
- Kafka;
- service mesh;
- event sourcing toàn hệ thống;
- CQRS phức tạp;
- multi-region active-active.

Các công nghệ trên chỉ được thêm sau khi có yêu cầu vận hành thực tế.
