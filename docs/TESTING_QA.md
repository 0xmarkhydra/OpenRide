# Testing & QA Strategy

## 1. Mục tiêu

Ưu tiên chất lượng cho các flow có rủi ro cao: auth, dispatch race, trip state, location realtime, payment và permission/background mobile.

## 2. Test pyramid

### Unit tests
Backend:
- trip state transition;
- pricing calculation;
- dispatch scoring/filter;
- validation;
- payment state machine.

Mobile/Admin:
- state/view model;
- validators;
- permission state;
- formatting/money/time.

### Integration tests
- PostgreSQL/PostGIS repository;
- Redis GEO/lock;
- REST handler -> domain -> DB;
- WebSocket auth/subscription;
- provider adapter với mock/stub.

### End-to-end
- Rider creates trip;
- Driver receives and accepts;
- Rider sees assigned driver;
- Driver arrives/starts/completes;
- Rider sees completion/history;
- Admin sees trip timeline.

## 3. Critical concurrency tests

- Hai driver accept cùng offer/trip gần đồng thời: chỉ một người thắng.
- Một driver nhận hai trip đồng thời: chỉ một assignment hợp lệ.
- Rider cancel và Driver accept race.
- Complete command retry: không double finalization/payment.
- Payment webhook duplicate: xử lý idempotent.

## 4. Location tests

- valid GPS update;
- stale timestamp;
- invalid coordinates;
- poor accuracy;
- offline driver update;
- location freshness expiry;
- GEO search excludes stale/busy driver;
- reconnect + snapshot sync.

## 5. Mobile device matrix

Trước production cần test trên nhiều thiết bị thật, ưu tiên:
- Android phiên bản thấp nhất hỗ trợ;
- Android phổ biến hiện tại;
- iPhone iOS thấp nhất hỗ trợ;
- iOS hiện tại.

Driver app phải test đặc biệt background/location/battery.

## 6. Network tests

- Wi-Fi -> cellular;
- cellular -> offline;
- high latency;
- packet loss;
- reconnect sau 30s/5m;
- request timeout;
- duplicate tap/retry.

## 7. Provider failure tests

- Maps timeout/5xx/quota error;
- SMS provider fail;
- push fail;
- payment timeout/webhook delay;
- object storage upload failure.

System phải fail rõ và không corrupt Trip state.

## 8. Load tests

Các scenario:
- concurrent API traffic;
- nhiều driver location updates/sec;
- WebSocket concurrent connections;
- dispatch burst;
- Redis GEO query;
- DB write/read under active trips.

Không công bố capacity nếu chưa chạy load test trên topology gần production.

## 9. Security tests

- IDOR Rider/Driver/Admin;
- expired/invalid token;
- admin role bypass;
- OTP brute-force/rate limit;
- webhook signature/replay;
- file upload type/size;
- SQL injection/input validation.

## 10. Regression suite

Mỗi bug production/critical phải có regression test nếu có thể tự động hóa.

## 11. UAT checklist

### Rider
- login;
- select route;
- estimate;
- booking;
- cancel;
- tracking;
- completion/history/rating.

### Driver
- KYC status;
- online/offline;
- offer accept/reject/expiry;
- arrived/start/complete;
- background GPS;
- earnings/history.

### Admin
- KYC approve/reject;
- rider/driver search;
- trip detail/timeline;
- pricing update;
- RBAC.

## 12. Severity

- Blocker: core flow không thể dùng/data corruption/security critical.
- Critical: chức năng core lỗi nghiêm trọng, không có workaround hợp lý.
- Major: feature quan trọng lỗi nhưng có workaround.
- Minor: cosmetic/edge case nhỏ.

Go-live không có Blocker/Critical known issues trên scope core, trừ khi có written risk acceptance.

## 13. Definition of Done cho ticket

- code review;
- unit/integration test phù hợp;
- docs/API updated nếu contract đổi;
- loading/error/empty state;
- observability cần thiết;
- security/privacy checked;
- QA pass acceptance criteria.
