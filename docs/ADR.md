# Architecture Decision Records

ADR lưu các quyết định kiến trúc quan trọng và lý do tại thời điểm ra quyết định. Khi đổi quyết định, thêm ADR mới hoặc cập nhật trạng thái; không xóa lịch sử.

---

## ADR-001 — Backend chính dùng Go

**Status:** Accepted

**Context:** Ride-hailing cần concurrency cao cho API, realtime, GPS và dispatch; đồng thời MVP cần tốc độ phát triển và mở rộng team tốt.

**Decision:** Backend chính dùng Go.

**Consequences:**
- concurrency/network workload phù hợp;
- binary deploy đơn giản;
- team dễ maintain hơn Rust cho toàn bộ backend;
- Rust chỉ dùng cho measured hot-path về sau.

---

## ADR-002 — Modular Monolith trước Microservices

**Status:** Accepted

**Decision:** Phase 1 deploy một Go application với module boundaries rõ.

**Reason:** Giảm network/distributed-system complexity, transaction dễ hơn, ra MVP nhanh hơn.

**Extraction triggers:** scale độc lập, ownership độc lập, measured bottleneck hoặc deployment cadence khác biệt.

---

## ADR-003 — PostgreSQL + PostGIS là durable source of truth

**Status:** Accepted

**Decision:** Transactional business data nằm trong PostgreSQL; spatial durable data dùng PostGIS.

**Reason:** Trip/payment/account cần consistency; PostGIS hỗ trợ geo data/zone/query khi cần.

---

## ADR-004 — Redis cho realtime hot state và GEO

**Status:** Accepted

**Decision:** Redis giữ online drivers, latest location, GEO index, dispatch locks và cache ngắn hạn.

**Constraint:** Redis không là durable source of truth cho trip/payment history.

---

## ADR-005 — WebSocket cho realtime, REST cho snapshot/commands

**Status:** Accepted

**Decision:** WebSocket dùng cho location/event realtime. REST dùng cho snapshot và command quan trọng như create/cancel/accept/start/complete.

**Reason:** Command semantics/idempotency/error handling rõ hơn; realtime vẫn low-latency.

---

## ADR-006 — Maps là third-party provider

**Status:** Accepted

**Decision:** MVP dùng Google Maps Platform hoặc provider tương đương cho map/search/geocode/route/ETA.

**Reason:** Tự xây map/routing không nằm trong core business MVP và chi phí/rủi ro rất lớn.

**Commercial consequence:** API usage fee là third-party operating cost, tách khỏi development fee.

---

## ADR-007 — Flutter cho Rider/Driver

**Status:** Accepted

**Decision:** Hai mobile app dùng Flutter để giảm duplication và tăng tốc delivery đa nền tảng.

**Constraint:** Driver background location vẫn cần platform-specific configuration/testing.

---

## ADR-008 — Next.js + TypeScript cho Admin

**Status:** Accepted

**Decision:** Admin web dùng Next.js + TypeScript, không dùng Flutter Web.

**Reason:** Web operations dashboard phù hợp ecosystem web, table/form/admin tooling và iteration nhanh.

---

## ADR-009 — Không Kafka/Kubernetes ở MVP

**Status:** Accepted

**Decision:** Không mặc định Kafka/Kubernetes ở Phase 1.

**Reason:** Chưa có load/team topology chứng minh cần; chi phí vận hành lớn hơn lợi ích.

**Future:** NATS/Kafka/K8s chỉ được thêm khi metrics/use case cho thấy rõ.

---

## ADR-010 — Payment state tách khỏi Trip state

**Status:** Accepted

**Decision:** Trip completed không đồng nghĩa payment captured. Payment là state machine riêng.

**Reason:** Provider timeout/failure/refund có lifecycle khác Trip và cần idempotency/audit riêng.

---

## ADR-011 — Không persist mọi GPS point vào transactional DB

**Status:** Accepted

**Decision:** Latest location ở Redis; trip path chỉ sample/persist theo nhu cầu audit/support/analytics.

**Reason:** Giảm write amplification/storage và tránh Postgres trở thành bottleneck do telemetry.

---

## ADR-012 — Go first, Rust only after profiling

**Status:** Accepted

**Decision:** Không rewrite component sang Rust vì benchmark lý thuyết. Chỉ dùng Rust khi profiling production/load test chứng minh CPU/memory/latency bottleneck và Go optimization không đủ.

**Candidate components:** custom routing, dense matching, geospatial stream processing.

---

## Template cho ADR mới

```text
## ADR-XXX — Title

Status: Proposed | Accepted | Deprecated | Superseded
Date:

Context:

Decision:

Alternatives considered:

Consequences:

Migration/Rollback notes:
```
