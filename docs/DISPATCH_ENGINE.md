# Dispatch Engine

## 1. Mục tiêu

Dispatch chọn và gán tài xế phù hợp cho Trip mới với ba ưu tiên:
1. match nhanh;
2. tránh double assignment;
3. có thể giải thích/tuning được.

MVP không dùng ML. Thuật toán deterministic trước để dễ debug và vận hành.

## 2. Input

- trip_id
- pickup lat/lng
- service_type
- city/zone
- rider context cần thiết
- candidate search policy

## 3. Candidate discovery

Redis GEO index theo city/service type:

```text
geo:drivers:{city}:{service_type}
```

Flow:

```text
Trip SEARCHING
  -> GEOSEARCH radius R1
  -> filter stale/unavailable
  -> score candidates
  -> offer
  -> nếu không thành công: expand R2/R3 hoặc retry policy
```

Không query toàn bộ drivers trong PostgreSQL cho mỗi booking.

## 4. Candidate filters

Driver bị loại nếu:
- không approved;
- offline;
- busy;
- latest location quá stale;
- sai service type;
- sai region;
- suspended;
- đang có active dispatch lock/trip;
- vehicle không hợp lệ.

## 5. Scoring MVP

Score có thể chuẩn hóa về higher-is-better:

```text
score =
  w_distance * distance_score
+ w_idle     * idle_time_score
+ w_accept   * acceptance_score
+ w_quality  * quality_score
- w_cancel   * cancellation_penalty
```

MVP có thể bắt đầu rất đơn giản bằng distance + idle time, sau đó tuning khi có dữ liệu.

Không đưa những tín hiệu chưa có dữ liệu đáng tin vào scoring chỉ để làm hệ thống “thông minh”.

## 6. Offer strategies

### Sequential
Offer D1, timeout, rồi D2.
- ít race;
- match chậm hơn khi driver không phản hồi.

### Small batch
Offer top N nhỏ, driver đầu tiên accept thắng.
- match nhanh;
- phải xử lý race/lock tốt.

MVP nên cấu hình được strategy và batch size. Bắt đầu sequential hoặc batch rất nhỏ.

## 7. Atomic accept

Accept phải là thao tác atomic.

Pseudo flow:

```text
BEGIN ACCEPT
  verify offer exists + not expired
  acquire trip assignment lock
  verify trip == SEARCHING
  acquire/check driver availability lock
  verify driver == AVAILABLE
  assign driver to trip
  mark driver BUSY
  invalidate other offers
COMMIT
```

Có thể dùng Redis distributed lock kết hợp DB transaction/conditional update. DB vẫn là source of truth cho assignment durable.

## 8. Double assignment protection

Các guard tối thiểu:
- unique/conditional invariant cho active trip per driver ở application/database level;
- trip version hoặc `WHERE status='searching' AND driver_id IS NULL` khi update;
- Redis short lock để giảm race;
- idempotent accept endpoint.

## 9. Offer TTL

Offer có expiry rõ ràng, ví dụ 8–15 giây để bắt đầu thử nghiệm. Giá trị thực tế cần tuning bằng telemetry.

Driver accept sau expiry nhận business error `OFFER_EXPIRED`.

## 10. Search expansion

Ví dụ policy ban đầu:
- Wave 1: 1.5 km
- Wave 2: 3 km
- Wave 3: 5 km

Đây chỉ là default thử nghiệm; city density và service type quyết định radius phù hợp.

## 11. Cancellation interaction

Nếu Rider hủy khi SEARCHING:
- transition trip CANCELLED;
- cancel all offers;
- release locks;
- notify candidate/assigned driver nếu cần.

Nếu đã ACCEPTED trở đi, cancellation rule có thể phát sinh fee/business policy riêng.

## 12. Driver goes offline/stale

Nếu driver mất heartbeat/location freshness:
- không đưa vào candidate mới;
- active offer có thể expire;
- nếu đang active trip, không tự cancel ngay; chuyển sang degraded/ops monitoring theo policy.

## 13. Metrics

Bắt buộc đo:
- dispatch requests/sec;
- candidates found per wave;
- zero-candidate rate;
- offer acceptance rate;
- median/p95 time-to-match;
- offer timeout rate;
- assignment conflict rate;
- search radius at success;
- cancellation before match;
- stale driver exclusion count.

## 14. Future improvements

Sau khi có data thật:
- ETA-to-pickup thay vì straight-line distance;
- supply/demand balancing;
- driver fairness;
- batching/multi-objective optimization;
- zone heatmap/repositioning;
- surge pricing signal;
- ML acceptance probability;
- Rust matching engine nếu Go service được profiling là CPU bottleneck thực sự.

## 15. Không làm ở MVP

- Không global optimizer toàn thành phố.
- Không reinforcement learning.
- Không tự xây routing engine.
- Không rank theo hàng chục feature chưa có dữ liệu.

MVP phải dễ quan sát và giải thích trước khi tối ưu sâu.
