# Mediacore 接入 4aion：原生复用与缺口调研

调查时间：2026-10-13。上游基线：`ede6cc1e707addceff03c2a0ad4bfe1767afa23e`。本文是静态研究，不是开发任务，也不代表接入、测试或部署完成。当前唯一执行入口是[执行计划](../plans/mediacore-execution-plan.md)，产品需求以[平台规格](../plans/mediacore-platform-spec.md)为准。

## 已确认边界

- Mediacore 基于上游代码独立部署；旧选型未正式生产，不迁移旧账号、余额、项目或任务。
- 平台持有 4aion 密钥，普通用户使用 Mediacore 积分；管理员管理用户、渠道、密钥和模型。
- 用户选择统一按“4aion 成本 × 倍率”定价，不采用逐 SKU 手工制定售价；具体倍率、成本口径与货币/积分换算尚未确定。
- 中转、模型、任务和账务优先复用现有能力；兼容差异以配置验证或待决事项记录，不从本报告派生新系统。
- 本窗口只完成研究和方案，不开发产品、不运行付费模型、不创建供应商素材、不部署。

## 结论

上游已经具备本次接入所需的大部分平台基础：系统渠道、服务端加密密钥、每模型协议与规格价格、原生积分、异步任务、提交记录、恢复、取消对账和下载持久化。优先验证这些入口的配置覆盖能力。[系统渠道](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/admin_channels.go)、[模型与价格结构](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_channel.go)、[任务计费协调器](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_billing.go)。

**“全面接入”不能用修改一个 Base URL 作为完成标准。** 4aion 的文本、图片、视频、素材 Action、存储上传各有协议；模型目录存在不等于全部输入、工具调用、取消与账务语义均可用。P0 明确缺口主要是 4aion 素材预处理、实际模型能力与价格校准，以及取消和提交结果不明确时的处理。[4aion 文档](https://4aion.com/docs)。

## 证据与限制

本次读取新基线的 Git 对象，避免把尚未完成的 checkout 当成完整源码。使用匿名 HTTP 复核 4aion 当前公开文档；web 浏览工具无法打开该站，但普通浏览器 UA 的 GET 可读页面及官方前端文档源：[当前文档 chunk](https://4aion.com/static/js/async/6799.806c78eb7c.js)。该地址是本次取证快照，长期入口仍为 `/docs`。

既有本地 4aion 文档及模型弹窗调查只作历史线索；未读取密钥、Cookie、真实用户数据，也未复用旧环境的“已通过”结论作为新项目验收。没有调用需要密钥的 `/v1/models`，所以**当前平台 Key 实际可用的全部模型集合、实际收费与限额仍待确认**。公开目录里的通用示例不能替代特定模型的服务端合同。[模型发现](https://4aion.com/docs#models)、[公开模型广场](https://4aion.com/pricing)。

## P0 能力矩阵

“原生具备”指代码中已有机制；“配置候选”指协议形态吻合但未与 4aion 联调；“差异待核”指静态比较发现协议差异、覆盖方式未定；“未验证”不代表可用。

| 能力 | 静态源码事实及原生入口 | 未验证的兼容差异 |
| --- | --- | --- |
| 平台渠道与密钥 | 原生具备：`admin_channels.go` 加密系统密钥，`provider.go/resolveProviderConfig` 从服务端渠道和授权模型解析真实协议、模型 ID、密钥 | 建 4aion 系统渠道；验证普通用户拿不到密钥，不能以请求体覆盖系统上游或调用未授权模型 |
| 模型发现与管理 | 配置候选：`channel_model_catalog.go` 请求 `/models`；模型记录已区分显示名、渠道模型键、真实供应商 ID、协议、能力和价格档 | 使用平台 Key 拉真实目录后人工/规则校准能力；只启用已验收规格。显示名只用模型名，不把营销解释混入模型名；精确供应商 ID 独立保存 |
| 文本 Chat / Responses / Claude | 配置候选：`provider_text.go`、文本 SSE 与声明式协议宿主已有三类路径 | 对每个上线模型验证非流式、SSE、结束信号、usage、图片输入、思考字段和错误；协议插件的路径前缀不同，检查最终 URL 防止 `/v1/v1` |
| Agent 工具调用 | 原生具备结构，4aion 未验证：`provider_agent_protocol.go` 统一会话，`provider_text_agent.go` 展开三类协议；Gemini 有独立声明式插件 | 验证多轮 tool call/result ID、并行工具、SSE 参数拼接、工具错误和中断；文本能回复不代表 Agent 可用。媒体工具必须走同一模型目录、报价和任务创建入口 |
| 普通图片生成 | 配置候选：`provider_image.go` 与 `openai-images` 插件支持 URL/base64 结果 | 逐模型确认请求字段、尺寸、数量和实际输出 MIME；生成后进入现有资源持久化，重开工程仍可见，不依赖临时上游 URL |
| 图片编辑、多参考、蒙版 | 待合同核对：手写路径是 multipart，`openai-images` v2 插件是 JSON `images[].image_url`，且优先分发到已安装声明式适配器 | 4aion 精确模型要求的报文格式、参考顺序、mask 与透明背景处理尚未确认 |
| Gemini 图片/文本 | 原生有 Gemini 图片路径和 generateContent 插件；4aion 当前逐模型鉴权与能力未验证 | 校准 `/v1beta` 路径、header 鉴权、inlineData/fileData、图片输出与 tool 调用；不能用 OpenAI 文本映射冒充完整 Gemini 多模态 |
| OpenAI-compatible 视频 | 配置候选而非直接通过：已有 `/videos` 创建、查询和 content 下载路径；独立 Seedance Videos 插件发送自己的 JSON 数组结构 | 4aion 文档声明 `/v1/videos`，但未给每模型完整编码合同；验证 JSON/multipart、首尾帧和参考结构后再决定现成协议是否足够 |
| Seedance 官方任务 | 原生有 Ark content/tasks 结构：`provider_video.go` 已保存并复用任务 ID、查询 `content.video_url` | 需验证 4aion 的 `X-Aion-Model`、`/api/v3` 路径、字段能力与错误解析能否由现有配置覆盖 |
| Seedance 素材 Action | `provider_ark_private_assets.go` 使用官方 IAM AK/SK/Region，当前只同步图片 | 4aion Bearer Action、素材绑定和生命周期合同与现有实现存在差异；覆盖方式尚未确定 |
| 素材上传与可读 URL | 原生资源体系及签名 URL 已有；公开文档另列 4aion storage 预签名流程 | 自有对象存储签名 GET 与 4aion storage 的兼容路线尚未确定；4aion 流程使用 `upload-credentials → PUT → complete → download-url`，上传头与平台鉴权用途不同 |
| TTS 语音合成 | 配置候选：`provider_audio.go` 有 `/audio/speech` 二进制、格式与签名校验 | 验证上线模型 voice、speed、instructions、格式、时长和费用；原生 `/audio/tasks` 异步音频不是 4aion 公开文档已承诺的通用端点 |
| 转写/翻译 | 编辑器转写使用本地 whisper.cpp `/inference`；4aion 声明 `/v1/audio/transcriptions`、`translations` | 现有转写入口属于覆盖目标；响应格式、时间戳与计费兼容性待验证，不能覆盖则列差异 |
| 任务恢复与去重 | 原生具备：`provider_submission.go` 稳定 attempt key，`provider_task_recovery.go` 原 ID 查询，worker 租约及媒体恢复 | 复用，不新增平行队列；验收创建已受理但响应丢失、进程重启、轮询失败、下载失败。原任务未知时禁止自动创建第二个计费任务 |
| 停止与供应商取消 | 原生有取消对账，**4aion 取消合同未公开确认** | 内置 Ark 会 DELETE task，不能因此认定4aion支持。未确认前明确区分停止等待/已请求取消/确认取消/已完成/费用待核对，不承诺点停止立即退款 |
| 积分、价格和对账 | 原生具备：用户账本、预留、结算、退款、uncertain、规格价格快照与幂等键 | 统一倍率的具体参数、成本目录来源、usage 单位与供应商实扣核对方式仍待配置或确认 |
| 限流、权限和出站安全 | 原生已有渠道并发、worker lease、SSRF、请求头处理、日志脱敏 | 复用控制面；核对全局/每渠道/每用户限额，429 Retry-After、总超时、有限退避；平台模式不能仅隐藏 BYOK 界面而允许后端绕过平台定价与准入 |
| Embeddings / Rerank | 4aion 公开列出端点，本次未找到足以确认现有产品消费链路的证据 | 不据此推导新增产品入口；是否存在既有调用方待核实 |

矩阵代码来源：[文本](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_text.go)、[Agent](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_text_agent.go)、[图片](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_image.go)、[图片插件](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/plugin-packages/openai-images/manifest.json)、[视频](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_video.go)、[音频](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_audio.go)、[转写](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/transcription_whisper.go)。

## 已识别的兼容差异

### 1. 4aion 素材适配

现有 Ark 实现值得复用的是资源就绪/归属检查、任务准备进度、已有绑定恢复和 `asset://` 替换时机；其鉴权、控制面和绑定维度不能原样套用。4aion 要求以稳定项目名与精确模型传入 Action，保留返回 ID，在素材可用后才提交视频。媒体上传和素材注册是不同阶段。[现有 Ark 素材代码](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_ark_private_assets.go)、[供应商素材文档](https://4aion.com/docs#videoAssets)。

静态证据表明，素材注册涉及资源归属、准备状态、供应商项目名/模型和返回 ID；这与单纯上传媒体不是同一步骤。是否能用现有能力及供应商生命周期限制，仍需按执行计划验证。素材永久有效性没有公开保证。

### 2. 声明式协议宿主的覆盖范围

现有宿主已经承接鉴权、请求组装、轮询、响应解析、媒体下载与上游提交 key；普通路径、字段、header、结果映射差异应优先在已有插件机制表达。上传凭证、外部 PUT、complete、素材注册、Active 等多阶段持久化过程不宜硬塞成一个 create 模板。[协议宿主](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_protocol.go)、[请求映射](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_protocol_request.go)、[鉴权](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_protocol_auth.go)、[结果下载](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_protocol_result.go)。

本报告不指定旧路由或适配工程作为实施来源。当前可确认的只是上游协议宿主对鉴权、请求组装、轮询、响应解析、下载及提交 key 的覆盖范围；差异是否可由配置表达，须验证实际生效插件与请求合同。

### 3. 账务与提交不明确

上游账本用整数 microcredits；订单有用户+幂等键约束、价格与规格版本快照、实际用量、上游请求 ID 和人工核对字段。`uncertain` 会保留冻结积分，等待核对；现有协调器阻止未知账单继续重试。这些机制不构成供应商已保证幂等的证据。[账务结构](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_finance.go)、[账务事务](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/finance_billing.go)、[提交记录](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_submission.go)。

4aion 公共文档提示创建中断时先确认原任务，未明确承诺某个 Idempotency-Key 的保存窗口和全端点去重保证。因此不能仅增加一个 header 就宣称不会重复扣费。取消也要由供应商确认后再按账务规则处理。[供应商错误处理](https://4aion.com/docs#errors)、[原生取消对账](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_task_cancellation.go)。

原生视频默认 30 秒查询一次，符合4aion公共文档建议的不高于每15秒一次频率；是否需要更快反馈应由任务事件与UI进度表达解决，不能靠无界轮询。[视频轮询策略](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_video_polling.go)、[任务租约](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_worker.go)、[任务文档](https://4aion.com/docs#video)。

### 4. 统一成本倍率的证据缺口

已经确认采用成本倍率，但 **4aion公开目录价、当前Key所属分组价格、单次实际扣费金额是三个不同数据**。本次公开 `/docs` 没有提供可确认的逐请求实际账单查询端点、统一 `cost` 响应字段、分组折扣计算及最终退款规则；未进行带密钥查询。已有历史模型广场截图只证明当时选定分组的展示价，不能证明本项目平台Key的最终成本。

原生 `ChannelModelPriceTier` 已内嵌 `CostPricing`，`BillingOrder` 内嵌成本快照并保存 `MultiplierBasisPoints`。全局倍率配置及 4aion 成本目录导入/刷新是否已有入口，属于待核差异。[渠道价格结构](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_channel.go)、[订单快照](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_finance.go)。

静态证据允许区分两类成本口径：

逐请求最终成本若可查，可与预估冻结值核对；目录价与 usage 则只能得出规则计算成本，不能等同供应商最终账单。

货币单位、积分兑换率、倍率生效版本、缓存/输入/输出 token、视频规格与音频时长、舍入和最低收费构成计价差异。当前源码与公开文档不足以证明“自动同步 4aion 真实成本”已具备。

## 验证范围与未决差异

本次没有读取平台密钥或真实用户数据，没有调用需密钥的 `/v1/models`、付费模型或供应商素材，也没有安装依赖、启动服务或运行应用验证。可用模型、分组价格、限额、取消合同、幂等保证和素材生命周期均须在执行计划对应阶段核实；未核实的模型或操作不能视为兼容。工作顺序见[执行方案](../plans/mediacore-execution-plan.md)，完成证据与状态记录在 Linear 对应任务中。
