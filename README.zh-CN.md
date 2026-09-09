<div align="center">

# OpenRide

### 开源出行市场基础设施

**司机自己定价。乘客自己选择。算法负责连接。社区可以自行部署。**

[English](README.md) · **简体中文** · [हिन्दी](README.hi.md) · [Español](README.es.md)

</div>

---

## 如果网约车是一种基础设施，而不是一个控制市场的中间平台？

OpenRide **不是另一个 Grab/Uber 克隆**。它是一个开源基础设施，让司机社区、合作社、本地运营商和创业公司无需从零重建整套技术，就可以运营自己的出行市场。

```text
乘客发布出行需求
        ↓
司机按照自己的商业条款报价
        ↓
OpenRide 透明排序
        ↓
乘客选择报价，或在乘客约束范围内使用 Quick Match
        ↓
Agreement 固化双方接受的条款
        ↓
行程按照 Agreement 执行
```

平台可以推荐，但不能偷偷修改双方已经同意的价格或条件。

> **当前实现状态：** Core、第一方模块、contracts、JS/TS SDK 和可独立部署的 Marketplace Service 进程已经存在。Marketplace 的持久化 schema 已建立基础，但完整的 request → quote → agreement 持久化流程、Outbox relay、Location Service 和 Ride Service 尚未全部完成。准确状态以 `docs/PROJECT_STATUS.md` 为准。

## 核心市场模型

**每个司机拥有自己的 tariff。** 价格不是“每个国家所有司机统一设置一个每公里价格”。国家或实例配置可以规定货币、法规约束和安全边界，但每个司机独立设置自己的商业条件，例如 `per_km`。

同一个 10 km 需求可以得到：

```text
司机 A → 5,000 VND/km → 50,000 VND → 10 分钟到达
司机 B → 6,000 VND/km → 60,000 VND →  3 分钟到达
司机 C → 5,500 VND/km → 55,000 VND →  6 分钟到达
```

OpenRide 不会简单地执行 `sort(price ASC)`。乘客应能看到价格、接驾时间、质量、可靠性和个人偏好的权衡。司机可以选择手动、自动或混合报价，并控制自己的价格边界。

### 不可妥协的原则

1. **司机控制自己的商业条款，包括自己的每公里价格。**
2. **乘客保留最终选择权。**
3. **最低价不应自动等于最佳选择。**
4. **拒绝不合适的订单不应自动被视为不良行为。**
5. **已接受的条款必须固化，不能被静默修改。**
6. **定价和排序必须可解释。**
7. **OpenRide 必须保持可自托管，并独立于单一运营商。**

## Marketplace 核心对象

```text
MobilityRequest
DriverTariff
Quote
Marketplace ranking
Agreement
Outbox / Inbox
```

核心不变量：

```text
Request → Quote → Agreement → Ride
金额使用最小货币单位的整数表示
每个司机拥有自己的 tariff
接受后的商业条款形成不可变快照
ranking 可以评分和重排，但不能改写 quote
```

## 微服务方向

OpenRide 按 bounded context 和数据所有权拆分服务，而不是为了看起来“微服务化”而创建空目录。

```text
一个 service → 一个 bounded context
每个 service 拥有自己的数据和 migrations
禁止跨 service 查询数据库
同步内部调用 → 需要立即响应时使用 gRPC
异步集成 → NATS JetStream
状态变更 + 事件 → transactional outbox
consumer side effects → inbox / idempotency
公共流量 → Edge Gateway / BFF
```

Marketplace 是第一个正在抽离的 V2 domain service。当前已经有进程、schema、health/readiness、service catalog 和 request validation；完整的 durable Marketplace API 仍在实现中。

## 可靠性

```text
DB transaction
  ├── 修改本服务拥有的 domain state
  └── 写入 outbox event
COMMIT
      ↓
outbox relay
      ↓
NATS JetStream
      ↓
consumer inbox 去重
      ↓
consumer-owned state change
```

## 快速运行

```bash
make packages-test
make core-example
make marketplace-test
make marketplace-run
```

开发微服务环境：

```bash
make micro-config
make micro-up
make micro-logs
make micro-down
```

## 文档

OpenRide README 和技术文档采用四语言策略：**English、简体中文、हिन्दी、Español**。文档入口与翻译状态见 `docs/README.md`。当代码更新后翻译暂时落后时，英文技术文档是 canonical source。

关键文档：`PROJECT_STATUS`、`PRODUCT_VISION`、`OPENRIDE_MANIFESTO`、`MICROSERVICES_ARCHITECTURE`、`PACKAGE_ARCHITECTURE`、`API_CONTRACT_V2`。

## 项目状态

OpenRide 当前为 **pre-1.0**，尚未宣称已经适合承载真实乘客的生产环境。真实运营还需要验证当地交通法规、保险、KYC、支付、事故响应、反欺诈、隐私和安全要求。

## License

OpenRide 使用 **GNU AGPL-3.0-or-later** 许可证。详见 `LICENSE`。

---

<div align="center">

### 为出行社区建设基础设施，而不是让社区依赖另一个封闭平台。

⭐ Star · 🧩 构建模块 · 🛠️ 自行运营 · 🤝 共同改进市场

</div>
