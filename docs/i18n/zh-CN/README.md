# OpenRide 文档

[English](../../README.md) · **简体中文** · [हिन्दी](../hi/README.md) · [Español](../es/README.md)

OpenRide 的公开文档采用四语言策略：English、简体中文、हिन्दी、Español。英文技术文档是 canonical source，中文文档必须保持相同的 API 路径、字段名、事件名、代码和业务不变量。

## 最重要的 Marketplace 不变量

每个司机拥有自己的 tariff。`per_km` 是按 **driver + service tariff** 独立配置的，不是一个国家的所有司机共用一个价格。国家/instance 可以定义货币、法规、安全约束和运营边界，但不能替代司机自己的商业定价。

## 文档结构

英文源文件位于 `docs/*.md`，中文镜像使用相同文件名放在 `docs/i18n/zh-CN/`。维护中的公开文档最终必须完整覆盖四种语言。若翻译暂时落后，必须明确标记 stale，不能把旧行为描述成当前行为。

优先同步：`PROJECT_STATUS.md`、`PRODUCT_VISION.md`、`OPENRIDE_MANIFESTO.md`、`MICROSERVICES_ARCHITECTURE.md`、`PACKAGE_ARCHITECTURE.md`、`DOMAIN_MODEL.md`、`DATA_MODEL.md`、`API_CONTRACT_V2.md`、`OPENRIDE_MIGRATION_PLAN_V2.md`、`ADR_OPENRIDE_V2.md`。
