# Roadmap & Scaling Strategy

## 1. Nguyên tắc

Scale theo dữ liệu thực tế, không theo cảm giác. Mỗi bước thêm độ phức tạp phải giải quyết một bottleneck hoặc business requirement đã quan sát được.

## 2. Phase 0 — Bootstrap

Mục tiêu:
- repo/CI skeleton;
- Go API health/config;
- PostgreSQL/PostGIS;
- Redis;
- Flutter Rider/Driver;
- Next.js Admin;
- docs nền.

Exit criteria:
- local environment chạy được;
- migration chạy được;
- CI baseline.

## 3. Phase 1 — Core MVP

### Backend
- Auth/OTP.
- Rider/Driver profiles.
- Driver KYC/vehicle.
- Pricing estimate.
- Trip state machine.
- Redis GEO dispatch.
- WebSocket realtime/location.
- Push notifications.
- Admin core.

### Mobile
- Rider booking/tracking/history.
- Driver onboarding/online/offer/trip.

### Exit criteria
- end-to-end trip chạy ổn;
- UAT pass;
- critical race tests pass;
- pilot production readiness.

## 4. Phase 2 — Pilot Operations

Mục tiêu:
- chạy thật ở khu vực giới hạn;
- thu telemetry;
- tune GPS interval;
- tune dispatch radius/TTL/scoring;
- hoàn thiện support/admin;
- payment production nếu cần.

Các số phải đo:
- trips/day;
- peak concurrent drivers;
- location updates/sec;
- websocket connections;
- time-to-match;
- no-match rate;
- cancellation rate;
- DB/Redis load.

## 5. Phase 3 — Scale một thành phố

Chỉ sau khi có số liệu pilot.

Có thể cần:
- API horizontal scaling;
- shared realtime pub/sub;
- DB index/tuning;
- Redis HA;
- worker separation;
- queue rõ ràng;
- service-area polygons;
- operational live map;
- better fraud signals.

## 6. Khi nào tách Location Service

Tách khi một hoặc nhiều điều đúng:
- location traffic chiếm phần lớn API throughput;
- cần scale CPU/network độc lập;
- deploy location thay đổi thường xuyên;
- process memory/GC/network profile khác rõ core API;
- cần specialized ingestion pipeline.

## 7. Khi nào tách Dispatch Service

Tách khi:
- dispatch latency bị ảnh hưởng bởi workload khác;
- cần scale/tuning độc lập;
- thuật toán trở nên phức tạp;
- ownership team riêng;
- event flow giữa location/dispatch đã rõ.

## 8. Khi nào dùng NATS/Kafka

### Redis Pub/Sub/Streams đủ khi
- topology đơn giản;
- event volume vừa;
- không cần retention/replay phức tạp.

### NATS cân nhắc khi
- nhiều service cần pub/sub request/reply nhẹ, latency thấp.

### Kafka cân nhắc khi
- event throughput rất lớn;
- cần durable log/replay/consumer ecosystem;
- analytics pipeline lớn.

Không dùng Kafka chỉ vì hệ thống “giống Grab”.

## 9. Khi nào dùng Kubernetes

Chỉ cân nhắc khi:
- nhiều service/container;
- cần autoscaling/rolling deploy/service discovery phức tạp;
- team có năng lực vận hành K8s;
- managed simpler platform trở thành giới hạn.

## 10. Khi nào dùng Rust

Backend chính vẫn Go. Rust chỉ cho hot-path được profiling:
- custom route computation;
- dense matching optimizer;
- high-volume geospatial stream processor;
- CPU/memory critical component.

Điều kiện trước khi rewrite:
1. có benchmark/profiling;
2. Go optimization hợp lý vẫn chưa đủ;
3. latency/cost thực sự ảnh hưởng business;
4. team có khả năng maintain Rust.

## 11. Database scaling order

Trước sharding:
1. fix slow query;
2. indexes;
3. connection pool;
4. cache;
5. partition large append tables;
6. read replica cho read-heavy;
7. archive/retention;
8. chỉ sau đó mới đánh giá sharding.

## 12. Multi-city

Khi mở nhiều thành phố:
- city/service region là first-class dimension;
- Redis GEO keys partition theo city/service type;
- pricing config version theo city;
- operational dashboard filter city;
- rollout feature/config theo city.

## 13. Multi-region

Không cần MVP. Khi có nhu cầu thật phải giải quyết:
- data residency;
- latency;
- active-active vs active-passive;
- cross-region trip ownership;
- failover;
- payment/provider regional dependency.

## 14. Capacity planning

Mỗi bước scale phải có load model:

```text
peak drivers online
x location updates/minute
= location events/minute

peak new trips/minute
x average candidates evaluated
= dispatch candidate operations/minute
```

Dùng load test gần production để xác nhận, không suy ra capacity từ benchmark máy cá nhân.
