# Observability

## 1. Mục tiêu

Khi có lỗi production, team phải trả lời được: lỗi ở đâu, ảnh hưởng ai, bắt đầu khi nào, mức độ bao nhiêu và có rollback/fix gì.

## 2. Ba trụ cột

- Logs: sự kiện có cấu trúc.
- Metrics: số đo theo thời gian.
- Traces: đường đi request/event qua component.

## 3. Structured logs

Backend Go log JSON với field tối thiểu:
- timestamp
- level
- service
- environment
- request_id
- trace_id nếu có
- user/driver/trip identifiers đã cân nhắc privacy
- event/action
- error_code
- latency_ms

Không log secret, OTP, token, raw KYC.

## 4. Request ID

Mỗi REST request có `request_id`. Nếu client gửi hợp lệ có thể propagate, nếu không server tạo. Webhook/event/background job cũng cần correlation id.

## 5. Metrics API

- request_count by route/status
- request_duration p50/p95/p99
- error_rate
- active requests
- DB query latency
- DB connection pool usage
- Redis operation latency/error

## 6. Metrics ride-hailing

### Dispatch
- trips_searching
- candidates_found
- no_candidate_rate
- offer_accept/reject/expire
- time_to_match
- assignment_conflicts

### Location
- location_ingest_rate
- location_invalid_rate
- location_stale_rate
- location_freshness
- websocket_connections
- reconnect_rate

### Trips
- created/completed/cancelled
- transition_failure
- active trips
- completion rate

### Third-party
- maps latency/error/cost proxy counts
- SMS success/failure
- payment webhook/error
- push send failure

## 7. Tracing

OpenTelemetry là hướng chuẩn. Trace ít nhất các flow:
- trip estimate -> maps -> pricing;
- trip create -> dispatch -> Redis -> DB;
- driver accept -> lock -> DB transition -> notify;
- payment -> provider;
- webhook -> payment update.

## 8. Error tracking

Mobile/Admin dùng Sentry hoặc công cụ tương đương để bắt crash/unhandled error. Backend có exception/panic/error aggregation và alert.

## 9. Dashboards

Tối thiểu có:
1. API health.
2. Dispatch health.
3. Realtime/location health.
4. Database/Redis.
5. Third-party dependencies.
6. Business trip funnel.

## 10. Alerts

Alert dựa trên symptom quan trọng, tránh alert noise.

Ví dụ:
- API 5xx vượt threshold 5 phút.
- p95 latency tăng mạnh.
- no-match rate tăng bất thường.
- Redis unavailable.
- DB connection saturation.
- payment webhook failures.
- location freshness degraded.

## 11. SLO ban đầu

SLO production phải chốt sau load test và pilot. Có thể bắt đầu bằng mục tiêu nội bộ cho availability/latency rồi điều chỉnh theo thực tế, không ghi cam kết SLA thương mại khi chưa đo.

## 12. Retention

- Application logs: retention theo chi phí/điều tra.
- Audit logs: retention dài hơn theo policy.
- Metrics: đủ để so trend.
- Traces: sample để kiểm soát chi phí.

## 13. Operational correlation

Mỗi trip detail Admin nên hỗ trợ liên kết/copy:
- trip_id
- rider_id
- driver_id
- request_id gần nhất
- payment reference

để support và dev truy vết nhanh.

## 14. Privacy

Observability data vẫn là dữ liệu thật. Mask phone/email và tránh raw coordinate history nếu không cần cho debugging.
