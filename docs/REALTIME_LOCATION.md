# Realtime & Location Design

## 1. Mục tiêu

Realtime layer phục vụ ba việc chính:
- Driver gửi location khi Online/đang trip.
- Rider nhận location/status của driver trong active trip.
- Driver nhận dispatch offer và trip state update.

Realtime không thay thế REST snapshot. Khi reconnect, client phải có cách lấy state chuẩn từ backend.

## 2. Transport

MVP dùng WebSocket. Production có thể đặt sau load balancer có hỗ trợ sticky session hoặc dùng shared pub/sub để nhiều instance fan-out event.

Event envelope đề xuất:

```json
{
  "type":"driver.location.updated",
  "version":1,
  "event_id":"evt_...",
  "occurred_at":"2026-08-10T10:00:00Z",
  "data":{}
}
```

## 3. Authentication

WebSocket connection phải authenticated bằng short-lived token/session. Server bind connection với principal:
- rider:{id}
- driver:{id}
- admin:{id}

Không tin `user_id` do client gửi trong event payload.

## 4. Rooms/channels

Logical channels:
- `user:{rider_id}`
- `driver:{driver_id}`
- `trip:{trip_id}`
- admin operation channels theo quyền nếu cần.

Client chỉ được subscribe channel mà backend authorize.

## 5. Driver location payload

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

Server bổ sung `received_at` để đo network delay và freshness.

## 6. Validation

Reject/ignore point khi:
- lat/lng ngoài range;
- timestamp quá cũ hoặc quá xa tương lai;
- accuracy vượt threshold không chấp nhận cho dispatch;
- tốc độ/teleport bất hợp lý vượt policy;
- driver không authenticated;
- driver offline nhưng vẫn gửi background update không được phép.

GPS spoofing nâng cao không nằm trong MVP, nhưng telemetry phải đủ để điều tra sau này.

## 7. Update frequency

Không hard-code một tần suất cho mọi state.

Đề xuất policy ban đầu:
- Offline: không gửi.
- Online idle: 8–15 giây hoặc adaptive theo di chuyển.
- Có offer/đang tới pickup: 3–5 giây.
- In trip: 2–5 giây tùy pin/network/UX.
- Background/mobile OS constraints: dùng best-effort và platform-specific configuration.

Con số cuối phải được tuning bằng test pin, data và UX trên thiết bị thật.

## 8. Redis updates

Mỗi location hợp lệ cập nhật:
1. latest location hash/value;
2. Redis GEO index của city/service type nếu driver available;
3. freshness timestamp/TTL.

Nếu driver chuyển busy, có thể remove khỏi available GEO index hoặc giữ ở index khác để tracking nhưng không dispatch.

## 9. Rider fan-out

Chỉ fan-out driver location tới Rider nếu:
- Rider là owner của active trip;
- trip đang ở trạng thái cần tracking;
- location không quá stale.

Không broadcast location của mọi driver cho Rider app.

## 10. Event types MVP

### Server -> Rider
- `trip.searching`
- `trip.driver_assigned`
- `driver.location.updated`
- `trip.driver_arrived`
- `trip.started`
- `trip.completed`
- `trip.cancelled`

### Server -> Driver
- `dispatch.offer.created`
- `dispatch.offer.expired`
- `trip.cancelled`
- `trip.updated`

### Driver -> Server
- `driver.location.update`
- heartbeat/connection metadata nếu cần.

Các command thay đổi trip state quan trọng vẫn ưu tiên REST hoặc server-validated command message có idempotency rõ.

## 11. Reconnect

Client reconnect với exponential backoff + jitter.

Sau reconnect:
1. authenticate lại;
2. query active trip snapshot;
3. subscribe room hợp lệ;
4. tiếp tục event stream.

Không giả định client nhận đủ mọi event trong thời gian mất mạng.

## 12. Ordering

Event cho một trip nên mang:
- `event_id`;
- `occurred_at`;
- `trip_version` hoặc sequence khi implement.

Client bỏ qua event cũ hơn snapshot/version hiện tại.

## 13. Horizontal scale

Khi có nhiều Go API instances:
- Redis Pub/Sub/Streams hoặc NATS có thể làm cross-instance event bus;
- connection registry có thể local nhưng routing event phải biết instance hoặc publish shared channel;
- không lưu authoritative trip state trong process memory.

## 14. Persistence

Không persist mọi websocket event. Persist business transition trong DB; location path chỉ sample khi có business/legal reason.

## 15. Metrics

Theo dõi:
- active websocket connections;
- reconnect rate;
- messages/sec;
- location ingest/sec;
- invalid/stale location rate;
- location freshness p50/p95;
- fan-out latency;
- dropped/backpressure events.
