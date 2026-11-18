# Mediacore · AI 文档索引

面向 AI 的短索引。详细文档维护规则见 [AGENTS.md](../AGENTS.md) 第 9 节「文档与交付」。

## Mediacore 当前工作入口

- [执行方案](plans/mediacore-execution-plan.md)：唯一执行入口；现有能力、最小工作、待决差异及完成证明。
- [产品规格](plans/mediacore-platform-spec.md)：已确认需求，保留 65 条用户故事；不是新增开发清单，对应 [NIK-269](https://linear.app/nikita96/issue/NIK-269)。
- [ADR-0014](adr/0014-minimal-reuse-no-custom-systems.md)：优先复用、最小二开、不自研。
- [领域术语](../GLOSSARY.md)：本项目唯一术语表。
- [4aion 调研](research/mediacore-aion-integration.md)、[服务器调研](research/mediacore-server-readiness.md)、[品牌与玻璃 UI 调研](research/mediacore-brand-glass-audit.md)：固定版本的静态事实、证据和差异，不是运行验收。
- [Linear 项目](https://linear.app/nikita96/project/mediacore-1d6254ea120c)：唯一任务状态与依赖来源。

## 上游设计参考

以下文档解释已有实现或上游提案，保留供定位源码时参考；其中的实施计划、自研组件迁移和未完成建议不属于 Mediacore 当前任务。是否开展工作以当前执行方案和实际子任务为准。

- [资源访问与 OSS/CDN 分发架构重构](design/resource-delivery-architecture.md)：资源身份、场景授权与分发策略分离，统一访问合同、全入口迁移、旧实现退场及流量/安全验收门槛（核心代码已实施，待真实 OSS/CDN 链路验收）。

- [短信渠道插件与登录注册策略](design/sms-channels-and-auth-policy.md)：阿里云/腾讯云官方 SDK 候选基线、多渠道与模板路由、发送记录、验证码安全、短信和邮箱登录注册组合及分阶段验收（核心代码已实施，待真实 OSS/CDN 链路验收）。

- [画布批量生成一致性治理](design/canvas-consistency-repair.mdx)：节点丢失、批次状态与引用解析的根因、已实施边界重构、回归证据和仍待验收的性能/同步场景。

- [画布 Agent 外观与 Live2D](content/docs/backend/canvas-agent-appearance.mdx)：独立助手名称、文案模板、模型包边界、Core 部署和验收要求。

- [云端 Agent 架构优化方案](design/cloud-agent-architecture-optimization.md)：跨端生成合同、引用绑定、审批依赖、报价预算、运行存储与工作上下文的现状审查、分阶段方案和验收门槛（提案，未实施）。

- [品牌首页设计参考](design/story-creation-homepage.md)：六幕电影卷轴、公开入口、故事素材、工作台预览与响应式降级。

- [插件平台与市场演进调研](design/plugin-platform-and-marketplace-research.md)：插件机制代码审计、对外回应、外部 SDK 与隔离运行时、独立插件验收、受控目录到公开市场的分阶段方案（调研建议，未实施）。

- [站点及外观与皮肤主题设计合同](design/site-appearance-and-skins.mdx)：品牌一致性、登录页与邮件、SEO/备案、三层皮肤令牌、无闪屏启动顺序和验收边界。

- [工作区外壳设计沉淀](design/workspace-shell-design.mdx)：侧栏（260px 可折叠导航 + 分组折叠）、主区卡片、顶部栏（账户/公告/主题）的设计决策与样式约束。

- [画布节点可读性设计沉淀](design/canvas-node-visual-contrast.mdx)：节点外壳、空态和图片创作面板在浅色/深色画布上的表面、边界、阴影与控件状态约束。

- [画布浮动控件设计沉淀](design/canvas-floating-controls.mdx)：顶部操作区、底部 Dock、小地图和右下角工作模式切换的浮动面板、定位与响应式约束。

- [用户诊断包设计](design/user-diagnostic-bundle.mdx)：面向普通用户的一键日志导出、前后端链路关联、脱敏、权限与排障方案。

- [AI 审美批改画布插件方案](design/ai-art-critique-solution.md)：云端视觉分析、并行 Reviewer、问题定位、AI 修改提示词与前端 SVG 标注的职责边界和交互设计。

- [LLM、Image、Video 主流请求协议全景与现有平台兼容性调查](design/model-request-protocol-landscape.md)：主流原生协议、聚合网关、图片/视频异步任务、参考素材 role、当前插件映射缺口与 MiniMax H3 专项审计。

- [编辑器预设插件化实施规格](plans/editor-preset-plugin-implementation.md)：参考 open-vetta 万物皆可插件，把编辑器做成预设插件的分阶段实施计划（SDK v2、命令状态机、8 个 editor 预设插件（含 AI 助手）、后端转写/导出任务、权限执行校验、AI 对话式剪辑），含产品视图、接口草案与文件规划。
- [编辑器实施 Runbook](plans/editor-implementation-runbook.md)：分步执行计划——M0~M6 里程碑 + 原子步明细（每步改动文件/验证/完成标准）、依赖关系、验证命令速查、高风险步与回退。解决「一次性实施效果差」：每步可验证、可回退、看得见进度。

## 上游决策记录（源码背景）

- [ADR-0001：编辑器整合边界](adr/0001-editor-integration-boundary.md)：借鉴 Concat 架构形态但以 TS 时间线状态机为真相源，不移植 Rust 引擎；媒体重活放 Go 后端。
- [ADR-0002：编辑命令协议](adr/0002-edit-command-protocol.md)：时间线唯一修改入口为可序列化编辑命令，手势 echo、有界快照撤销、防抖保存。
- [ADR-0003：预览与导出分层](adr/0003-preview-export-layering.md)：交互近似预览 + 单一滤镜图计划驱动导出，禁止手写第二套滤镜串。
- [ADR-0004：自动字幕转写服务化](adr/0004-autocaption-transcription-service.md)：转写为 Go 后端异步任务，结果写回字幕轨道。
- [ADR-0005：编辑器预设插件架构](adr/0005-editor-preset-plugin-architecture.md)：参考 open-vetta「万物皆可插件」，编辑器全部 UI 与能力为预设插件贡献，插件 SDK 演进为 v2（UI 插槽 + 预设分发 + 权限执行校验）。
- [ADR-0006：语音滤镜与项目模板](adr/0006-voice-filters-and-templates.md)：补充评估 Concat 剩余两个缺口——语音滤镜走 Web Audio 预览近似 + ffmpeg 音频滤镜图导出，模板走项目快照服务化；均为预设插件且服从单一滤镜图计划。
- [ADR-0007：AI 编辑交互](adr/0007-ai-editing-interaction.md)：对话式剪辑作为预设插件——AI 输出受约束命令 JSON（schema 校验 fail-closed），与手势命令同队列同撤销栈；≤3 条直接执行、批量改动 diff 预览待确认。
 - [ADR-0008：自研 UI 组件层与 AntD 替换](adr/0008-ui-component-layer-antd-replacement.md)：`components/ui` 重建为自研组件层（Celadon UI），吸收 dbx 契约形态（族目录+barrel）与 BoardUI 设计深度（语义状态矩阵/复合排版/RAC 原语/动效纪律/agentic 产品块），分阶段替换 AntD（依赖面 180 文件/41 导入）最终脱离；试点画布工具条/节点徽章，唯一实施计划见 [plans/ui-kit-rollout-plan.mdx](plans/ui-kit-rollout-plan.mdx)，活规范见 [plans/ui-design-system.mdx](plans/ui-design-system.mdx)。
- [ADR-0009：媒体产出模型与选用模型分开记录](adr/0009-produced-model-separate-from-selection.md)：生成成功时冻结产出模型，选用模型继续随下拉变化；角标按当前目录解析展示名。

## Mediacore 决策记录

- [ADR-0014：最小二开边界](adr/0014-minimal-reuse-no-custom-systems.md)：原生验证、配置、必要局部调整；能力缺口待决。
- [ADR-0010：独立平台基线](adr/0010-mediacore-platform-foundation.md)：独立数据、优先复用和平台统一供给。
- [ADR-0011：成本倍率与待核对账单](adr/0011-cost-multiplier-and-uncertain-billing.md)：价格快照、未知不重发和冻结积分核对。

- [ADR-0012：费用授权上限与平台差额](adr/0012-authorized-cost-ceiling.md)：用户扣费受事前授权上限约束，超额由平台承担。

- [ADR-0013：最终无法交付的媒体任务退款](adr/0013-undeliverable-media-refund.md)：已收供应商费用而最终无法交付时，由平台承担成本并释放或退还用户积分。

## 按约定维护的文档（`docs/content/docs/`）

以下页面记录原有功能、代码地图和待验证背景；Mediacore 的实际任务及状态以 Linear 为准：

- [AI 审美批改画布插件](content/docs/plugins/ai-art-critique.mdx)
- [功能](content/docs/overview/features.mdx)
- [本地开发](content/docs/backend/local-development.mdx)
- [HTTP API 合同](content/docs/backend/http-api.mdx)
- [后端数据库](content/docs/backend/backend-database.mdx)
- [代码地图](content/docs/backend/code-map.mdx)
- [待测试](content/docs/progress/pending-test.mdx)
