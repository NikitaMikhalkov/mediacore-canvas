# 4aion 接入调研（2026-10-13，补证修订版）

## 结论与证据边界

官网为 [4aion.com](https://4aion.com)。目前存在可匿名读取的[官方开发文档](https://4aion.com/docs)，覆盖文本、图片、Seedance 视频、素材管理、存储上传、音频端点与错误处理。文档为 SPA；网页抽取工具打不开正文，但普通 HTTPS GET 能取得其[当前文档脚本](https://4aion.com/static/js/async/6799.806c78eb7c.js)。本报告已用当前网络重新读取该脚本，不仅引用旧报告。脚本文件名可能变化，长期引用 `/docs#章节`。

补证线索来自旧工程 `docs/research/mediacore-aion-integration.md`。**旧工程的用户确认、平台积分/倍率选择、平台 key 选择和已通过结论均不继承。** 初查时官网 `docs_link` 只回指首页，因此曾未发现文档；当前报告已更正这一初查缺口。

仅公开 GET 与只读文档调查；未读取密钥、Cookie、真实用户数据，未登录、上传、注册素材、创建任务或付费生成，未调用带 key 的 `/v1/models`。文档声明不等于实测可用。

## 身份、文档与模型目录

[官网](https://4aion.com) HTML 标题 AION；[公开状态接口](https://4aion.com/api/status) 返回 `system_name=Aion` 和 `server_address=https://4aion.com`。不要与 Aion Labs、4All API、4AICode 混同。官网脚本含 koureikishi/Aion GitHub 链接，但本次公开仓库访问 404，不作为服务端可审计证据。

[公开定价接口](https://4aion.com/api/pricing) 本次有 87 个模型条目、`pricing_version=a42d372ccf0b5dd13ecf71203521f9d2`。未登录 `current_group=""`、`usable_group={}`、`group_ratio={}`、每项 `enable_groups=[]`、价格与倍率字段为零。不能解释为免费或当前账号可调用。官方说明 `GET /v1/models` 的 `data[].id` 按 key 权限过滤；显示名与精确模型 ID 必须分开。[模型文档](https://4aion.com/docs#models)

公开目录 `supported_endpoint` 声明 OpenAI Chat、Anthropic Messages、Gemini generateContent。图片/视频条目也可能仅标 `openai`；标签不能代替具体模型协议。名称相似、大小写不同、后缀不同不能自行合并。[目录](https://4aion.com/api/pricing)

## 当前官方协议

| 能力 | 文档声明的端点/行为 | 仍需核实 |
|---|---|---|
| 模型 | GET `/v1/models`，按 key 权限过滤 | 目标 key 的实际模型集和分组 |
| OpenAI 文本 | POST `/v1/chat/completions`、`/v1/responses`，Bearer；`stream:true` 使用 SSE | 精确模型 tool calls、JSON schema、usage、缓存、上下文限制 |
| Claude | POST `/v1/messages`，`x-api-key` + `anthropic-version`，示例版本 2023-06-01 | 目标模型参数与工具多轮兼容 |
| Gemini | 公开目录声明 POST `/v1beta/models/{model}:generateContent` | 原生鉴权、streamGenerateContent、文件/图片/工具合同，不能由标签推断 |
| 图片 | POST `/v1/images/generations`，model/prompt/size/n；返回 `data[0].url` 或 `b64_json` | 各模型尺寸、数量、格式、上限、价格与实际结果 |
| 图片编辑 | 端点表有 POST `/v1/images/edits` | multipart/JSON、多参考图、mask、透明背景合同缺少逐模型说明 |
| OpenAI 视频 | POST `/v1/videos`、GET `/v1/videos/{video_id}`、GET `/content` | 创建体编码、每模型输入/输出/状态、取消 |
| Seedance | `/api/v3/contents/generations/tasks` 创建/查询，见下节 | 具体型号参数组合、限额、取消与幂等保证 |
| 音频 | POST `/v1/audio/transcriptions`、`translations`、`speech` | 公共目录没有独立音频型号证明；voice/格式/时间戳/费用等未详述 |
| Embeddings/Rerank | POST `/v1/embeddings`、`/v1/rerank` | 是否为影策现有产品链路所需，勿仅因端点存在新增产品 |

来源：[端点总览](https://4aion.com/docs#endpoints)、[鉴权](https://4aion.com/docs#auth)、[文本](https://4aion.com/docs#text)、[图片](https://4aion.com/docs#image)、[公开目录](https://4aion.com/api/pricing)。

OpenAI SDK Base URL 应为 `https://4aion.com/v1`。除 Anthropic 例外，文档统一要求 Bearer；JSON UTF-8 与 Content-Type。模型特定协议仍优先逐项验证，不能由通用鉴权段推断所有 SDK 行为。

## Seedance 视频任务

官方文档明确 Seedance 2.0/2.5 使用 `/api/v3/contents/generations/tasks`。提交与查询示例带 Bearer 和精确 `X-Aion-Model`。创建响应 `id` 是应保存的公开任务 ID。`content` 数组包含非空文本以及 `image_url`/`video_url`/`audio_url` 的嵌套 URL；model 和非空文本必填。常用参数 duration、ratio、resolution、generate_audio、watermark、seed；callback_url、return_last_frame、service_tier、execution_expires_after、draft、tools、frames、camera_fixed 等仅当前模型支持时生效。[视频文档](https://4aion.com/docs#video)

状态为 queued、running、succeeded、failed。成功读取 `content.video_url`，样例亦有 `last_frame_url`；失败读取 `error.code/message`。建议轮询间隔不短于约 15 秒并设总超时；未知状态记录后按总超时退出。中英文轮询措辞略有歧义，但英文明确 “no more frequently than about every 15 seconds”，不应解读为至少每 15 秒必须请求一次。回调支持因模型而异，仍需轮询兜底；调用方须独立验真并幂等处理。[视频文档](https://4aion.com/docs#video)

文档没有清晰承诺取消端点、Idempotency-Key 保存窗口、创建超时后的查询检索方法、重复提交退款、任务保留期及产物 URL TTL。不能以 Ark 官方或影策现有实现替代 4aion 保证。

## 素材 Action 与存储

Seedance 素材与视频共用 AION Bearer key，无需额外 AK/SK。每次 Action 必须带稳定 ProjectName 与精确 X-Aion-Model；文档声明客户、项目、模型隔离，Seedance 2.0/2.5 与不同项目不能混用素材。新平台仍需自身证明用户资源归属，不能直接把共享供应商账户隔离等同平台租户隔离。[素材文档](https://4aion.com/docs#videoAssets)

流程为 `CreateAssetGroup → CreateAsset → GetAsset` 等到 Active，保存 AION `Result.Id`，再用 `asset://<ID>` 创建视频，不可构造/改写 ID。Action 路径 `/api/v3/?Action=<ACTION>&Version=2024-01-01`。支持 Create/Get/List/Update/DeleteAsset，素材组 Create/List/Update/DeleteAssetGroup，没有 GetAssetGroup。分页 PageNumber/PageSize（1–100）；非空组删除 ResourceInUse；Action JSON 最大 8 MiB。素材操作本身不计视频生成费用；此说法不证明所有存储/流量均免费。[素材文档](https://4aion.com/docs#videoAssets)

CreateAsset.URL 只接受供应商服务端可读 HTTP/HTTPS URL，不接受本地路径、multipart 或二进制。可复用自有对象存储限时只读签名 GET URL，不要求 Cookie/额外 header，至少保持到 Active；无需向供应商交桶凭据。无自有存储时使用：

1. POST `/v1/storage/upload-credentials`：filename、content_type、size_bytes、purpose=video_reference、related_model、asset_type。
2. 按返回 upload_headers PUT 文件到 upload_url，**不可发送 AION Authorization 给 upload_url**。
3. POST `/v1/storage/objects/{object_id}/complete`（示例 etag、size_bytes）。
4. GET `/v1/storage/objects/{object_id}/download-url`，将 download_url 传 CreateAsset.URL。

单文件上限取返回的 `max_size_bytes`；其他限制由模型/媒体类型决定。Content-Type 应与 Image/Video/Audio 匹配；不存在可泛化到全部模型的固定最大文件大小。[素材文档](https://4aion.com/docs#videoAssets)

Action 使用 `ResponseMetadata`（RequestId/Action/Version/Error）和 Result 信封；视频任务使用 error 对象。Action 错误包含 InvalidAction、AuthFailure、InvalidRequest、ProjectAccessDenied、ResourceAccessDenied、InvalidAsset、ResourceNotFound、ResourceInUse、InternalError、MappingError。

## 错误、重试、费用与密钥

官方错误文档区分 400 参数错误、401 key 缺失/无效/过期/撤销、403 账户/分组/scope、404 端点/模型/任务、429 速率/并发、500 网关、502/503 模型暂不可用。先看 HTTP 状态，再兼容 `error.code/type/message` 或顶层 `code/message`；部分网关错误为 `{success:false,message}`，不要依赖可变 message 分支。[错误文档](https://4aion.com/docs#errors)

文档建议仅重试 429、502、503 与网络超时；429 优先 Retry-After，否则有限指数退避与随机抖动。**创建任务中断仍需先核对原任务，不是无条件 POST 重试授权。** 数值 RPM/TPM/并发上限、队列限额、退款与实际计费 API 未公开充分说明。

官方要求 key 只在服务端环境变量保存、不同应用独立 key，不进入浏览器、移动包、日志或 Git，并设置连接/读取超时、记录请求 ID。[鉴权文档](https://4aion.com/docs#auth)

当前公开文档不能证明请求实扣查询、分组倍率、成本/积分换算或失败退款规则。公开目录价、目标 key 分组价、最终实际扣费必须分开。代理商专用 API 在前端中有独立权限章节，不代表普通 Mediacore 账户拥有该权限，本次未调用。

## 最小二开建议与验收（工程判断）

1. 将全面接入定义为影策现有功能逐项映射，建立“功能→现有 provider/插件→模型→请求合同→差异→验收”矩阵；不要直接等同 87 型号全启用。
2. 优先复用原系统渠道、协议插件、模型目录、任务与账务、素材保存和权限控制。普通路径/header/字段映射先查配置能力；多阶段素材 Action 如确有缺口，只补现有资源准备链路最小适配。
3. 本轮已确认平台管理员统一管理 4aion key 和模型，普通用户仅消费 Mediacore 积分。必须在后端隔离用户资源与费用；保存 provider 绑定时包含项目/精确模型，避免跨型号素材复用。成本倍率和商业价格尚未决定，不继承旧规格。
4. 先用离线模拟契约验证，再在明确模型/次数/预算授权后付费 smoke。成功、失败、超时未知、刷新/重启恢复、下载失败重试、跨用户读取拒绝都需证据。
5. 复用原轮询策略能满足约 15 秒下限则不改。不得为更快进度展示无限加速轮询；创建未知结果不得自动第二次扣费请求。
6. 用户已明确所有 AI 调用只能走 4aion，不允许其他 provider 兜底；也已允许确认不支持的 AI 能力暂时关闭后首期上线，保留代码并公开例外清单，不把关闭项记为通过。首期按每项保留能力至少一个通过验收的 4aion 模型开放，不要求全部 87 个型号通过。

### 转写与本地 Whisper 的具体缺口

4aion 官方端点表确实声明 POST `/v1/audio/transcriptions` 和 `/v1/audio/translations`；不是完全没有音频声明。但本次当前文档没有转写请求示例、multipart 字段、模型 ID、返回结构、时间戳粒度、字幕格式、语言/翻译选项、文件上限或费用。当前公开 87 型号目录没有明确 ASR/TTS 型号，不能证明目标 key 具备能力。[官方端点表](https://4aion.com/docs#endpoints)

当前上游功能文档明确编辑器转写依赖本地 whisper.cpp（见同目录上游报告）。它不能仅改 Base URL 就视为兼容 OpenAI `/audio/transcriptions`。单供应商约束下本地 Whisper 不能作为隐式兜底。最小路线是先确认 4aion 具体 ASR 模型及合同，再复用影策转写入口、字幕/时间轴消费结构，仅在既有调用边界适配请求与响应；必须验证逐词/逐段时间戳满足原功能。若供应商无法满足，按本轮用户已确认的例外规则暂时关闭该能力；这不阻塞其余功能的首期上线，但阻塞该项完整能力验收，报告和产品入口都应如实说明。

## 剩余关键未知

目标 key 模型与分组、真实单价/倍率/账单与退款、逐模型图片编辑/音频/Gemini 参数、视频取消与幂等、素材生命周期、任务与 URL TTL、数量/大小/时长限制、速率与并发、数据保留/训练/删除/SLA。公开状态 privacy_policy_enabled/user_agreement_enabled=false 仅为显示配置，不能推断实际服务条款。需要官方文档补充或去密钥响应样本，不索取明文 key 写进研究。

## 当前公开目录（原样模型 ID，仅元数据）

| model_name | endpoint 标签 |
|---|---|
| `gemini-2.5-flash-image-token` | gemini, openai |
| `gemini-2.5-flash-image` | gemini, openai |
| `gemini-3.7-flash` | openai, gemini |
| `claude-opus-5` | openai, anthropic |
| `grok-imagine-image` | openai |
| `gemini-3.1-flash-lite-image-token` | gemini, openai |
| `gpt-6-luna` | openai |
| `doubao-seedance-2.0-fast` | openai |
| `video-mini` | openai |
| `claude-haiku-4-5-20251001` | anthropic, openai |
| `gpt-image-2.5` | openai |
| `kimi-k2.7-code` | openai |
| `gpt-5.6-terra` | openai |
| `glm-5.1` | openai |
| `claude-opus-4-7` | anthropic, openai |
| `gpt-image-2-plus` | openai |
| `doubao-seedance-2-0-260128` | openai |
| `minimax-m2.7` | openai |
| `grok-build-0.1` | openai |
| `grok-4.5` | openai |
| `claude-opus-4-7-thinking` | anthropic, openai |
| `claude-opus-4-5-20251101` | anthropic, openai |
| `claude-sonnet-5` | anthropic, openai |
| `gemini-3.1-flash-lite-image` | gemini, openai |
| `glm-5` | openai |
| `gemini-2.5-pro` | openai, gemini |
| `gemini-3.8-flash` | gemini, openai |
| `claude-fable-5` | anthropic, openai |
| `gemini-3.6-flash` | openai, gemini |
| `deepseek-v4.1-flash` | openai |
| `kimi-k3` | openai |
| `gemini-3-pro-image-token` | gemini, openai |
| `claude-opus-4-6-thinking` | anthropic, openai |
| `gemini-3.1-flash-lite` | openai |
| `gpt-5.6-sol` | openai |
| `glm-5.3` | openai |
| `gpt-6-sol` | openai |
| `doubao-seedance-2.0` | openai |
| `doubao-seedance-2.0-mini` | openai |
| `gemini-3-flash-preview` | openai, gemini |
| `deepseek-v4-pro` | openai |
| `qwen3.6-plus` | openai |
| `claude-opus-4-6` | anthropic, openai |
| `gemini-3-pro-image-preview` | gemini, openai |
| `gemini-3.5-flash-lite` | openai |
| `kimi-k2.6` | openai |
| `minimax-m2.5` | openai |
| `gemini-3.1-flash-image-token` | gemini, openai |
| `glm-5.3-flash` | openai |
| `Doubao-Seedance-2.0-fast` | openai |
| `seedance-2.0-international` | openai |
| `doubao-seedance-2.5` | openai |
| `video-pro-fast` | openai |
| `Doubao-Seedance-2.5` | openai |
| `kimi-k2.5` | openai |
| `doubao-seedance-2-0-fast-260128` | openai |
| `video-pro` | openai |
| `gpt-6-astra` | openai |
| `gemini-3.1-pro-preview` | openai, gemini |
| `claude-opus-4-8` | anthropic, openai |
| `claude-sonnet-4-6` | anthropic, openai |
| `gpt-image-2-c` | openai |
| `gpt-5.6-luna` | openai |
| `seedance-2.0` | openai |
| `gemini-3.1-flash-lite-preview` | gemini, openai |
| `gemini-3.5-flash` | openai, gemini |
| `Doubao-Seedance-2.0` | openai |
| `minimax-m3` | openai |
| `claude-opus-4-8-thinking` | anthropic, openai |
| `claude-fable-5-1` | anthropic, openai |
| `gpt-image-2.5-sunburst` | openai |
| `gpt-image-2-pro` | openai |
| `gpt-image-2.5-flare` | openai |
| `gpt-5.5` | openai |
| `gpt-4o-mini` | openai |
| `grok-imagine-video` | openai |
| `Doubao-Seedance-2.0-mini` | openai |
| `claude-opus-5-5` | openai, anthropic |
| `grok-imagine-image-quality` | openai |
| `grok-4.3` | openai |
| `gpt-6.1-sol` | openai |
| `doubao-seedance-2-5-260628` | openai |
| `doubao-seedance-2-0-mini-260615` | openai |
| `deepseek-v4-flash` | openai |
| `glm-5.2` | openai |
| `grok-imagine-video-1.5` | openai |
| `gemini-3.1-flash-image-preview` | gemini, openai |

来源：[4aion /api/pricing](https://4aion.com/api/pricing)，2026-10-13 未登录读取。实时列表可能变化，型号名称本身不证明上游真实性、功能等价、授权或价格。
