# Infrastructure & DevOps

## 1. Environments

Tối thiểu:
- Local
- Staging
- Production

Dev team có thể dùng preview environment nếu CI/CD hỗ trợ nhưng không bắt buộc MVP.

## 2. Local stack hiện tại

`docker-compose.yml` cung cấp:
- PostGIS `postgis/postgis:16-3.4`
- Redis 7 Alpine

Database local:
- db: `flashx`
- user: `flashx`
- password local-only: `flashx`

Production không dùng credential local mặc định.

## 3. Production topology đề xuất

```text
Internet
  -> CDN/WAF (nếu dùng)
  -> Load Balancer
  -> Go API instances >= 2 khi cần HA
       |-> Managed PostgreSQL/PostGIS
       |-> Managed Redis
       |-> Object Storage
       |-> Queue/worker
       |-> Third-party providers
```

Admin có thể deploy riêng static/server app.

## 4. Containerization

- Backend đóng Docker image multi-stage.
- Run bằng non-root user khi có thể.
- Image immutable/tag theo commit SHA/release.
- Không bake secret vào image.

## 5. Database

Production ưu tiên managed PostgreSQL có:
- automated backup;
- point-in-time recovery nếu ngân sách cho phép;
- monitoring connection/storage/CPU;
- SSL connection;
- maintenance window.

PostGIS extension phải được provider hỗ trợ.

## 6. Redis

Production ưu tiên managed Redis có persistence/HA phù hợp. Vì Redis chứa dispatch hot state, outage có thể ảnh hưởng booking mới.

Không dùng Redis như durable source of truth cho payment/trip history.

## 7. CI/CD

Pipeline đề xuất:

```text
PR
 -> format/lint
 -> unit tests
 -> static/security checks
 -> build
 -> integration tests
 -> merge
 -> staging deploy
 -> smoke/UAT
 -> production approval
 -> deploy
 -> smoke checks
```

## 8. Branch/release

Khuyến nghị trunk-based hoặc short-lived feature branches. Release tag theo SemVer hoặc date/version convention thống nhất.

## 9. Configuration

Config qua environment/secret manager:
- database URL;
- Redis URL;
- auth secrets;
- maps provider secrets;
- SMS provider;
- payment provider;
- object storage;
- observability endpoints.

Mỗi env có credential riêng.

## 10. Migrations

Migrations chạy có kiểm soát trước/đồng thời với app deploy theo compatibility rule.

Không tự động chạy destructive migration không review trên startup app production.

## 11. Scaling

Scale API horizontally khi CPU/connections/latency yêu cầu. Scale database bằng tuning/index/read replicas trước khi sharding.

Location/Dispatch chỉ tách service khi measured load hoặc ownership cần.

## 12. WebSocket scale

Nhiều API instances cần shared messaging layer cho event cross-instance. Redis Pub/Sub/Streams hoặc NATS là lựa chọn phù hợp trước Kafka.

## 13. CDN/Object Storage

KYC private; asset public nếu cần có bucket/prefix riêng. Upload file lớn nên dùng signed upload URL thay vì proxy toàn bộ qua API.

## 14. Backup/restore

- Automated DB backup.
- Xác định RPO/RTO trước launch.
- Test restore định kỳ, không chỉ tin rằng backup “có chạy”.
- Object storage versioning/lifecycle tùy dữ liệu.

## 15. Infrastructure as Code

Khi production ổn định, dùng Terraform/OpenTofu hoặc provider IaC để version hạ tầng. MVP bootstrap có thể deploy thủ công nhưng production config phải được ghi chép/reproducible.

## 16. Cost control

- Budget alerts.
- Autoscaling có upper bound.
- Log retention hợp lý.
- Maps/SMS/payment usage monitoring.
- Không provision Kubernetes/multi-region trước nhu cầu.

## 17. Production readiness checklist

- domain/TLS;
- secrets rotated khỏi local defaults;
- backup enabled;
- monitoring/alerts;
- migration verified;
- rate limits;
- admin RBAC/MFA policy;
- third-party production accounts;
- store builds signed;
- incident contacts/runbook.
