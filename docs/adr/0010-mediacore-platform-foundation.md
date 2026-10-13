---
status: accepted
---

# Mediacore 采用独立平台基线

本项目以既有开源基线的独立 fork 为开发主线，独立部署与初始化数据，不安排旧账号、余额、工程或任务迁移。旧实现仅作为玻璃风格和成熟能力参考，避免继承另一套系统依赖。

用户消费 Mediacore 积分，管理员统一管理用户、平台渠道、密钥和模型。中转、模型与账务沿用原生机制；实施边界由 [ADR-0014](0014-minimal-reuse-no-custom-systems.md) 规定，执行顺序见[执行方案](../plans/mediacore-execution-plan.md)。

当前阶段整理调研、需求与任务；正式产品代码、真实调用和部署在后续执行阶段开展。
