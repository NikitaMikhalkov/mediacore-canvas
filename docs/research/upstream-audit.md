# Mediacore：影策上游与最小复用调研

调研时间：2026-10-13。固定基线：`ddcat-ai/open-ai-canvas@26dfb8d91bddf78ada93f7cba079368f1634db6d`，提交时间 2026-10-13 05:38:08 UTC，提交内容为数据库升级兼容旧工作台 v51 账本；根 VERSION 为 `v1.6.7`。这不是旧工程 `ede6cc1…` 的继续实施或验收。下列链接均固定到本次 commit；事实来自源码与一手文档，方案是建议，未做部署或付费调用。

## 结论

官方目标明确为 [ddcat-ai/open-ai-canvas](https://github.com/ddcat-ai/open-ai-canvas)。它已经是 React/Go 的多用户工作台，包含认证、用户隔离、管理员、系统渠道、任务、资源、积分、兑换码、在线支付与对账。Mediacore 不需要另建认证、计费、队列、模型网关或素材系统。最小路线是固定上游基线，先完成原生功能的部署和验收，复用后台配置接入渠道与品牌；仅针对证实的缺口做小补丁。

“现有代码具有该能力”不能视为“公开商业运营已通过验收”。上游自身仍提示未经安全配置不应直接公网多用户；大量功能位于待测试清单。完整功能还存在部署依赖缺口，尤其时间线导出与转写。来源：[README](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/README.md)、[待测试清单](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/progress/pending-test.mdx)。

## 1. 工程布局与约定

| 单元 | 原生方案 | Mediacore 最小动作 |
|---|---|---|
| Web | Vite、React 19、TypeScript、React Router、Ant Design、Tailwind、Zustand、TanStack Query；Bun 锁文件 | 保留真实页面与状态模型，修改原有入口 |
| Backend | Go 1.25、Gin、GORM，SQLite 或 PostgreSQL、Redis | 生产直接采用 PostgreSQL + Redis，不新建 Node/BFF |
| Agent | 本地内嵌 Node/Pi；生产独立 yingce-agent 容器 | 保留现有运行与令牌合同，不擅自去掉独立服务 |
| Docs | Next.js、Fumadocs、MDX | 调研/ADR 与现有专题并存，避免重写 README |
| Plugins | 声明式模型协议、支付原生 RPC 等，104 个 manifest（本次文件计数） | 优先现有协议；仅精确差异新增插件或修正映射 |

根存在 AGENTS.md，不存在 CLAUDE.md、GLOSSARY.md、GLOSSARY-MAP.md、根 package.json 或 pnpm-workspace.yaml。适合单上下文领域词汇表与根 docs/adr；不能因前后端分目录就推定要采用复杂多上下文配置。原生调用方向是 HTTP → handler → service 别名 → app/领域包 → repository/model → database/resource。来源：[AGENTS](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/AGENTS.md)、[插件目录](https://github.com/ddcat-ai/open-ai-canvas/tree/26dfb8d91bddf78ada93f7cba079368f1634db6d/plugin-packages)。

## 2. 功能范围与一期验收表

| 原生能力 | 应验证的完整闭环 |
|---|---|
| 无限画布 | 创建/保存/刷新恢复、节点与连线、布局、撤销重做、导入导出、版本、只读分享 |
| 影视生产 | 项目、剧本、角色、场景、风格板、分镜、参考素材与 3D 导演台；真实产出写回 |
| 文本/图片/视频/音频 | 原生支持的每种操作与参数，经 4aion 选定模型真实成功、失败、取消和恢复 |
| 任务中心 | 排队、进度、日志、SSE 重连、取消/重试、上游任务 ID 恢复、幂等保存 |
| 素材 | 上传、下载、分类、引用保护、私有访问、短时签名 URL、过期与跨用户拒绝 |
| 跨设备同步 | 同一账号重开与跨设备、账号切换、断网/恢复、并发写冲突；本地缓存不冒充服务端保存 |
| Agent/技能 | 服务启动、流式回放、读画布、权限/审批、实际支持的写操作；README“只读阶段”与现有权限代码存在文档/能力口径差异，必须以真实用例确认 |
| 时间线 | 修剪/拆分/字幕、语音转写、ffmpeg 渲染、资源入库、下载成片 |
| 运营 | 注册/登录/找回、用户封禁、功能开关、积分/兑换码、充值、对账、公告、审计 |

来源：[功能清单](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/overview/features.mdx)、[画布同步合同](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/canvas-operations-sync.mdx)、[Agent 权限测试](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/cloud_agent_permissions_test.go)。

“完全功能”建议定义为该固定基线原生实现且经确认纳入的功能清单，逐项记载通过/失败/依赖供应商/未测试；不把供应商不存在的模型能力伪装成适配任务，也不将上游未来愿景自动列为一期自研。

## 3. 多用户、公开注册和商业运营

代码确认支持普通用户/管理员、本地密码、bcrypt、服务端 Cookie 会话、邮箱验证码、短信/邮箱策略、Linux.do OAuth、密码重置和会话撤销。第一个账号直接成为管理员；公开注册默认关闭，但可通过后台/环境开启。默认邮箱注册策略要求邮件服务，必须先配置 SMTP 或已支持的短信方式。注册要求同意条款；OAuth 的 accepted_terms 是一次性登录状态字段，并非长期条款签署存证。来源：[auth.go](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/auth.go)、[registration.go](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/registration.go)、[verification_policy.go](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/verification_policy.go)、[数据库文档](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/backend-database.mdx)。

已有登录/注册/验证码限流，资源访问检查 UserID 归属，账号切换要求清理 scoped 缓存。其主要模型是用户私有数据，并未在本次模型目录检索中发现组织/团队成员实体；“多用户独立账户”与“团队共享协作/SaaS 组织租户”应分别决定，后者不可凭多用户标签推定原生完成。来源：[认证 handler](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/handler/auth.go)、[资源鉴权](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/resource_access.go)、[model 目录](https://github.com/ddcat-ai/open-ai-canvas/tree/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/model)。

原生财务包含账户、流水、冻结/结算、成本快照、规格售价、兑换码批次、支付商品、订单、回调收件箱、不可变支付配置版本与对账。支付入口可复用微信 Native、支付宝电脑网站、虎皮椒等官方包，选择实际商户渠道后验收即可。真实商户账号和密钥、资费、退款规则和收费授权仍是业务决策。来源：[兑换码事务](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/repository/finance_redeem.go)、[支付协议](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/plugins/payment-plugin-runtime.mdx)、[模型价格与成本](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/backend-database.mdx)。

公网验收最低覆盖：先在非公网完成首个管理员注册；HTTPS 与明确 CORS；两个普通账户交叉读写/删除/分享/下载拒绝；验证码/密码重置/限流；余额不足、并发扣费、重复回调、失败返还、超时待核对；系统渠道密钥不进入浏览器/日志；数据备份包含数据库、媒体与 .settings-key 并完成恢复。上游明确要求前四类部署安全边界，后面的闭环为商业上线建议验收，不是已通过的安全审计。[SECURITY](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/SECURITY.md)。

## 4. Linux 部署与已发现依赖缺口

官方生产 Compose 已包含 PostgreSQL 17、Redis 7.4、独立 migrate、yingce-agent、backend、web；只有 web 映射主机端口 3000。后端与 web 镜像要求显式参数，文案要求 digest；Agent 当前仍有 tag 回退兼容旧更新器。建议 Mediacore 全部固定自有构建 digest，并保留原 Compose 服务结构及迁移路径。数据库升级先迁移后启动，不能仅回滚镜像假定恢复数据库。[生产 Compose](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docker-compose.deploy.yml)。

**代码级依赖差距：** backend/Dockerfile 运行镜像仅安装 ca-certificates、tzdata、wget，没有 ffmpeg/ffprobe；task_render.go 查找 ffmpeg，缺失即失败。转写实现要求 CANVAS_WHISPER_BASE_URL；生产 Compose 没有 whisper 服务，也没有传递该变量。成片导出优先补齐镜像内 FFmpeg，复用渲染逻辑；转写则受本轮全部 AI 仅走 4aion 的决定约束，不能补装本地 Whisper 当作合规完成，应验证 4aion ASR 合同后局部适配，或者按已确认例外暂时关闭。来源：[Dockerfile](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/Dockerfile)、[渲染实现](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/task_render.go)、[转写实现](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/transcription_whisper.go)。

资源支持本地、阿里云 OSS、腾讯云 COS。模型参考图/Agent 看图需要上游实际可访问的签名链接；本地存储时设置 CANVAS_PUBLIC_BASE_URL。跨境上游可达性、签名时效、大文件上传、视频 URL 保存与 CDN 403 必须从目标 Linux 实测。worker/channel 默认并发各 3，不应据此推算商用容量。压力测试和容量目标尚未完成。[环境样例](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/.env.example)、[资源存储合同](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/backend-database.mdx)。

支付 RPC 插件当前控制工作目录、环境、输出上限和超时，但尚无通用 OS 沙箱、签名校验或通用网络白名单。公开运营先限制为审查过的可信包，不给普通用户上传并执行任意插件的能力。[支付运行时安全边界](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/plugins/payment-plugin-runtime.mdx)。

## 5. 4aion 全面接入：应复用的位置

本次在 backend、web/src、plugin-packages 检索 `4aion|4ai` 无结果，因此没有证据声称已有专用集成。已有系统渠道、模型能力、上游模型 ID、价格档、规格 SKU 与协议插件可承载中转商；OpenAI chat/responses/images/audio/videos、Gemini、Anthropic、NewAPI 等插件可作为逐合同映射的候选。品牌是中转商不代表所有 API 都是同一种 OpenAI 协议。[插件清单](https://github.com/ddcat-ai/open-ai-canvas/tree/26dfb8d91bddf78ada93f7cba079368f1634db6d/plugin-packages)、[渠道数据模型](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/backend-database.mdx)。

建议每项记录：4aion 精确模型 ID → 能力/操作 → URL/鉴权 → 请求字段与参考媒体格式 → 同步/异步与轮询 → 状态/错误 → 返回媒体 → usage/结算 → 现有插件 → 配置或最小差异补丁 → 实测证据。重点不是只测模型列表，而是文生图/图编辑、多图、首尾帧/参考视频、TTS/ASR、流式文本/工具调用等真实行为。4aion 的完整供应商合同与账号可用模型清单应由另一专项报告补齐。

异步视频已有持久化 Provider Task ID、30 秒轮询、短暂 task_not_exist 容错、瞬时错误继续查询原 ID 的语义，优先复用，避免在不确定任务状态时二次 create 导致重复收费。[NewAPI 视频合同](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/newapi-channel-2.mdx)。

## 6. Mediacore 品牌替换与兼容边界

优先使用后台站点外观：brandName、brandSlug、深浅 Logo、登录标题/视频、SEO、页脚；Agent 名称/欢迎文案也已有独立配置。前端 use-appearance-store 将公开配置写入标题、OG 与 Twitter meta。仍有硬编码默认品牌、恢复默认提示、YINGCE STUDIO 回退、Agent 默认名等；应做展示面清单与默认值小补丁，确保冷启动、配置失败/重置时仍是 Mediacore。代码检索在 web/index.html 与 web/src 的 ts/tsx/json 中约 88 行相关匹配，仅作为检索规模，不是需替换行数。[后端 appearance](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/appearance.go)、[前端外观](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/web/src/stores/use-appearance-store.ts)、[后台外观页](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/web/src/pages/admin/settings/appearance-settings-page.tsx)。

建议一期保留 `yingce.plugin/v1`、`yingce.payment/v1`、`.yingce-plugin`、Go module/import、Cookie/localStorage key、环境变量与内部服务名，除非明确要求且能证明迁移安全。这些是数据与插件合同，不等同用户可见品牌；全局 replace 将破坏现有包、会话、缓存和升级路径。README/营销链接、默认图标、演示域名、品牌素材与自有发布更新源需单列验证，确保运行环境不会自动更新回官方影策镜像。

许可证在本次基线明确为 MIT，允许修改/分发/销售但必须保留原版权与许可通知。NOTICE 还指出第三方各自许可，以及 web/public/icons 中产品名称/Logo 的权属。可以将产品展示改为 Mediacore，但不能把 LICENSE/NOTICE 的作者署名机械替换成自有版权，也不能据仓库 MIT 推断所有外部素材、模型输出或第三方标志可重新授权。来源：[LICENSE](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/LICENSE)、[NOTICE](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/NOTICE)。

## 7. 现有测试与完成证据

Quality workflow 已配置 Go 测试、PostgreSQL 集成、Redis 可执行程序、依赖测试不得静默 skip、支付插件 Linux amd64/arm64 执行、前端 frozen Bun 安装/typecheck/lint/test、3D Chrome E2E。值得直接复用；CI 存在不代表该 commit 测试已通过，本调研未查询对应运行结果、未运行全套测试。[质量工作流](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/.github/workflows/quality.yml)。

一期交付建议附：固定源码 commit/镜像 digest/数据库版本；上述功能矩阵；4aion 模型协议矩阵；双账号隔离结果；收费与对账证据；备份恢复演练；部署与回滚记录；所有未测试/供应商不支持项。健康检查不能替代登录、SSE、媒体生成与下载验证。用户确认才是验收，不以“代码有测试”冒充。

## 8. 建议拆分到新 Linear 项目的工作

1. 固定上游、独立工程与 Matt skills 配置，记录许可/上游 remote 与升级规则。
2. 原生能力和待测试清单冻结，建立一期可验收矩阵。
3. Linux 基础部署与迁移、媒体/转写依赖补齐、备份恢复。
4. 公开注册、SMTP/短信、条款配置、双用户隔离与运营安全验收。
5. 4aion 精确协议/模型/操作矩阵，配置优先、按差距最小适配。
6. 原生积分/兑换码/支付/对账的商业闭环和失败路径验证。
7. Mediacore 可见品牌、默认值、素材、SEO 与自有更新源替换；保留许可和兼容合同。
8. 完整功能回归与一期验收；二期功能单独排期，不将未经批准自研塞入一期。

## 9. 按最终用户约束补充的完整功能矩阵

已确认的本轮业务边界：平台统一保管 4aion 密钥，以平台积分收费；所有 AI 调用只允许 4aion；对外品牌 Mediacore，保留法定署名及内部兼容协议。以下验收不是宣称已完成。路由取自 [router.tsx](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/web/src/router.tsx)，内部机制取自前述功能/数据库文档。

| 分类 | 真实入口 | 依赖 | 复用机制 | 阶段一验收 |
|---|---|---|---|---|
| 项目管理 | /projects、/projects/:projectId、chapters/:chapterId | 认证、DB、项目资源 | 原生项目/章节/制作流程 | 建项目→章节→保存→重开；另一用户不可访问 |
| 画布管理 | /canvas、/canvas/:id | DB、浏览器 localForage、资源 | operations journal、版本、节点/边与撤销 | 全节点与连线、导入导出、离线恢复、冲突、持久化 |
| 自由创作 | /、/create | 4aion 文本/图/音/视频模型、计费 | 创作会话、统一任务与素材同步 | 每类创作→历史→结果引用→收费闭环 |
| 生图/视频工作台 | /toolbox/image、/toolbox/video；后台 workbench-tools | 功能开关、管理员工具配置、4aion | 工具卡片/预设/参数、统一任务 | 开启入口、预设、参考图/视频、历史与资源；模型能力不虚报 |
| Agent | 画布助手与创作 Agent 入口 | yingce-agent/Pi、4aion 文本工具调用、SSE | 原生对话持久化、上下文、工具权限和审批 | 对话/读画布/允许的写操作/拒绝越权、重连、工具结果持久化；监测出站只去获准 4aion 端点 |
| 剧本与分镜 | /projects/:projectId/:view、workflow/:unitId/:stage | 4aion 文本、多媒体、项目资产 | 剧本、角色/场景、结构化分镜、批量任务 | 原文→结构化剧本→角色/场景→分镜→生成→回写→制作状态 |
| 3D 导演/预演 | 项目/画布中的导演台 | 浏览器图形能力、3D 素材；AI 部分 4aion | 原生预演、控制图、镜头与场景 | 交互/相机/控制图导出、返回画布、引用后生成；Chrome E2E 加真实浏览器 |
| 素材库 | /assets、/admin/resources | 本地或 OSS/COS、签名公开 URL | 用户资源、分类、引用保护、删除队列 | 上传/分类/移动/下载/跨用户拒绝/引用阻止删除/物理删除故障 |
| 任务中心 | /tasks 与画布任务面板 | worker、Redis、4aion | 任务状态、SSE、provider_request_id、重试 | 排队、取消、失败、恢复、断线、重复同步与单次收费 |
| 时间线编辑 | /projects/:projectId/editor（:view） | ffmpeg/ffprobe；字幕 ASR 当前为 whisper.cpp | 原生命令/历史、编辑器插件、转写任务、渲染计划 | 编排/修剪/撤销→转写字幕→输出成片→入库→下载；ASR 路线未解决前不得勾全通过 |
| 分享 | /share/canvas/:token | 分享 token、资源投影 | 原生只读分享 | 访问、撤销、媒体、禁止编辑、禁止越权访问未分享数据 |
| 灵感/提示词 | /inspirations、/admin/prompt-templates | 随包目录与用户本地状态 | 原生灵感、模板、收藏/导入 | 浏览/搜索/收藏/插入；明确自建灵感本地持久化边界 |
| 技能 | /skills、/admin/skill-curation | GitHub 技能来源、Agent | 原生安装、分类、版本与策展 | 安装可信技能、搜索、使用、卸载；技能不能绕开统一供应商/计费 |
| 插件 | /plugins、/admin/plugins | 声明式包/可信 RPC、权限 | 原生 manifest/版本/runtime/权限 | 安装受支持包、开关、版本、拒绝未授权执行；模型插件出站仍受 4aion 约束 |
| 用户与账户 | /login、/register、/forgot-password、/settings、/admin/users | SMTP 或短信插件、Cookie、DB | 原生认证/找回/头像/封禁/策略 | 公开注册、验证、找回、密码更新撤会话、限流、账号切换隔离 |
| 积分与支付 | /wallet 弹窗、/admin/credit-operations、redemption-codes、payments | 商户渠道、原生支付插件、DB | 积分账本、兑换码、幂等订单/回调/对账 | 充值/兑换→额度→调用→结算/返还，重复并发请求与对账异常 |
| 模型与渠道 | /admin/channels、/admin/models | 4aion 合同/系统密钥 | 系统渠道、模型/规格/售价/成本配置 | 用户只能用平台目录；系统密钥不泄露；非 4aion 自定义渠道/直连路径无法绕过 |
| 运营与品牌 | /admin/announcements、banner-announcements、settings/appearance、settings/features | 配置、资源 | 原生公告、功能开放、站点与助手外观 | Mediacore 冷启动/正常/降级/重置、Logo/SEO/分享/登录、公告与开关 |
| 运维与安全 | /admin/logs、login-logs、settings/runtime-policy、storage、system-update、system-performance | Linux、PG、Redis、日志、备份、自有镜像 | 原生审计/限流/存储/更新机制 | 健康+真实闭环、告警、重启任务恢复、备份还原、自有更新源、受控升级 |

**sole-4aion 的明确冲突：** 原生字幕 ASR 是本地 whisper.cpp HTTP 服务，不经过 4aion。若“所有 AI 只走 4aion”包含 ASR，就不能直接部署 whisper 服务当作满足要求。最小补丁应在既有 transcription 客户端边界添加或选用 4aion 实际支持的 ASR 协议，保留时间线任务、资源、字幕与错误流程；前提是确认供应商端点、文件限制、时间戳/分段合同及计费。如果供应商没有满足字幕时间戳需求的 ASR，则“全部功能”和“仅 4aion”之间是产品约束冲突，必须明确例外或接受阶段缺口，不能暗中改成本地 Whisper。

用户 Q12 已确认处理原则：若 4aion 确实缺少某项原生 AI 能力，允许暂时关闭该能力并首期上线，保留代码并公开例外清单，不静默切换其他供应商。因此 ASR 差距阻塞的是“字幕转写这一项完整能力通过”，不必阻塞首期整体发布。应记录供应商缺失证据、关闭入口与后端行为、用户可见提示和后续恢复条件。

同样，用户自带渠道、插件、外部工作流（例如 AutoDL/ComfyUI）、第三方账号生成入口等原生功能可能绕开统一 4aion。建议区分保留上游代码与对外开放能力：复用功能开关/后台渠道权限关闭未授权调用入口，并验证后端真正阻止；不能仅隐藏前端。ffmpeg 的本地视频渲染属于媒体处理而非模型推理，可继续原生复用。

## 调研限制与本地证据

完整只读 clone 位于 `/tmp/mediacore-upstream-audit`，HEAD 与上述 commit 一致，git status 干净。已核对源码/配置/文档，未执行应用、自动测试、支付或模型调用，也未连接 Linux 服务器。未完成专项渗透/负载审计、完整依赖许可证清点或全部 104 个插件逐文件检查；4aion 供应商合同由专项调研承担。没有 fork、部署、发表消息或更改既有项目。
