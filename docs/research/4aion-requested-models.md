# 4aion 指定九型号的原生协议映射

结论：新增六型号都已有可复用的原生协议候选，无证据要求新增适配器；逐型号生成仍待验收。Gemini 原生目录只读鉴权已通过，预览别名的实际后端仍不明；`minimax-m3` 是文本型号，原示例放在视频类有误。本轮只考虑增加六型号，既有 GPT 文本、GPT 图片、Seedance Mini 三项保持不动。

## 证据范围

- 已读取用户示例的脱敏正文：九个示范全部使用 `chat.completions`，连图片/视频也提交同一条量子纠缠文本。这只能提供精确型号和 SDK Base URL，不能证明媒体协议、能力或成功。
- 本次本机读取 [4aion 文档](https://4aion.com/docs)失败；主代理通过只读公开 GET 取得 2026-10-14 首页、[公开目录](https://4aion.com/api/pricing)快照供本研究核对。当前目录 87 项，`pricing_version=a42d372ccf0b5dd13ecf71203521f9d2`；九个精确型号全部存在。未登录价格为零不能解释为免费。目录字段是 `supported_endpoint_types`，全局 `supported_endpoint` 只是协议路径映射。
- 主代理另提供同日目标 Key 的脱敏 `/v1/models` 结果：54 个精确 ID，九项全部命中。**账号目录可见仍不等于生成成功。** 本研究代理没有读取 Key、账号正文或原始 RTF、访问服务器或调用生成；只读取脱敏 ID 和公开页面快照。
- 今日官方入口索引 `/docs/` 仍加载文档块 `6799.806c78eb7c.js`；已读取主代理重新抓取的 35,581 字节正文，与[前次调研](4aion-integration.md)的公开文档快照逐字节相同。另读取 Google、xAI、MiniMax、火山引擎官方资料及仓库插件；原厂规格不能替代 4aion 对参数、价格和实际后端的承诺。
- Gemini 只读鉴权旁证：主代理以 `x-goog-api-key` 请求 `/v1beta/models` 返回 200、54 模型；同一头访问 `/v1/models` 返回 401。原生 Gemini 路径可复用，不能把不同协议的模型列表路径混用。此检查不调用 `generateContent`，不能证明图片生成或所有字段支持。

## 九型号配置表

表中 `B` 表示协议从服务端 Key 自动产生 `Authorization: Bearer <KEY>`；JSON 自动设置 `Content-Type: application/json`。`M` 表示额外渠道头 `X-Aion-Model: <该行精确型号>`；视频创建和轮询都须携带。Base URL 均指 Mediacore 渠道配置，所有实际调用只走 4aion。

| 精确 Provider Model Key | 能力 / 原生协议 | Base URL | 头部 | 参数与未解决项 |
|---|---|---|---|---|
| `gpt-image-2-plus` | 图片 / `openai-image` | `https://4aion.com/v1` | B；M 保持既有配置 | 既有项不改。目录明确 1K/2K；`model,prompt,n,size,quality`、`data[].url/b64_json`。编辑、蒙版、透明背景、格式不能由默认能力推断。 |
| `grok-imagine-image` | 图片 / `grok-image` | `https://4aion.com/v1` | B；M 是否必需待核 | `aspect_ratio,resolution,response_format,n`；不用 OpenAI 的像素 `size` 代替比例。网关是否接受 `/images/edits` JSON 单参考图未证实。 |
| `gemini-3.1-flash-image-preview` | 图片 / `gemini-image` | `https://4aion.com` | 自动 `x-goog-api-key`；原生目录鉴权已通过 | `contents[].parts`、`generationConfig.responseModalities=[TEXT,IMAGE]`、`imageConfig.aspectRatio/imageSize`；返回候选的 `inlineData`。预览别名实际映射和图片生成未证实。 |
| `doubao-seedance-2-0-260128` | 视频 / `volcengine-ark-video` | `https://4aion.com` | B + M | 非空文本 `content[]`、`duration,ratio,resolution,generate_audio,watermark`；建议先限定 4–15 秒、480p/720p/1080p，4K 待 4aion 确认。 |
| `doubao-seedance-2-0-fast-260128` | 视频 / `volcengine-ark-video` | `https://4aion.com` | B + M | 同一协议，独立能力：4–15 秒、480p/720p；不能沿用默认 1080p。参考/首尾帧/音画组合仍待网关验收。 |
| `doubao-seedance-2-0-mini-260615` | 视频 / `volcengine-ark-video` | `https://4aion.com` | B + M | 保持既有渠道；原厂范围同为 4–15 秒、480p/720p。本文未重新生成或重新证明既有结果。 |
| `minimax-m3` | **文本** / `chat-completion` | `https://4aion.com/v1` | B；M 是否必需待核 | `model,messages,stream,max_tokens`；4aion 当前标记 `Chat,Reasoning,Long Context,Cache`，与原厂语言模型分类一致。图像/视频理解输入不等于视频生成；实际工具、视觉、上下文仍待核实。 |
| `claude-fable-5` | 文本 / 首选 `chat-completion` | `https://4aion.com/v1` | B；M 是否必需待核 | 复用 OpenAI 兼容文本合同。当前公开目录另列 Anthropic；若改 `claude-api`，使用 `/v1/messages`、`x-api-key`、`anthropic-version: 2023-06-01`、顶层 `system` 与必填 `max_tokens`。不可从别名推断原厂规格。 |
| `gpt-6.1-sol` | 文本 / `chat-completion` | `https://4aion.com/v1` | B；M 可保持既有配置 | 保持既有文本路线；`messages,stream,max_tokens`，工具、JSON Schema、视觉、缓存和上下文不可由协议默认推断。 |

4aion 的 [图片](https://4aion.com/docs#image)、[视频](https://4aion.com/docs#video)、[鉴权](https://4aion.com/docs#auth)、[端点](https://4aion.com/docs#endpoints)：图片为 `/v1/images/generations`；Seedance 为 `POST /api/v3/contents/generations/tasks` 和 `GET .../{id}`，成功读取 `content.video_url`，失败读取 `error.code/message`。本次[公开目录](https://4aion.com/api/pricing)声明 Claude 为 `anthropic,openai`，Gemini 为 `gemini,openai`，其余七项为 `openai`；该标签不证明媒体须走 Chat，也不证明具体媒体参数。

原厂旁证：xAI 的[精确型号页](https://docs.x.ai/developers/models/grok-imagine-image)将其定义为图片输出，[Images API](https://docs.x.ai/developers/rest-api-reference/inference/images)区分 generation 与 edit；[MiniMax 模型调用文档](https://platform.minimax.io/docs/guides/text-generation)明确 M3 是语言模型并支持 OpenAI 兼容文本路线；[火山模型目录](https://docs.volcengine.com/docs/ark/model-list?lang=zh)及[视频技术说明](https://docs.volcengine.com/docs/LakeAIService/VideoGenerationEnhancedEditionBasicEdition?lang=zh)列出三个精确 Seedance 型号及各自分辨率/4–15 秒范围。这些地址不作为 Mediacore 的供应商地址。

Google [官方更新记录](https://ai.google.dev/gemini-api/docs/changelog#may-28-2026)宣布 `gemini-3.1-flash-image-preview` 于 2026-06-25 停用。4aion 当前目录仍列该名字，可能是保留别名；实际映射未知，**不能自行换成 GA 型号，也不能把原厂停用等同 4aion 必然不可用**。原生图片参数参见 [Google 图片文档](https://ai.google.dev/gemini-api/docs/image-generation)；当前最新示例已转向新型号/Interactions，不能原样替换仓库现有 generateContent 合同。

## 仓库复用边界与必要差异

1. **模型和参数用配置。** [ChannelModelRequest](../../backend/internal/app/channel_models.go)已有精确 `providerModelKey`、`protocol`、`capabilityConfig` 和价格档；[默认能力](../../backend/internal/app/model_capability_defaults.go)不是 4aion 白名单。Ark 默认时长从 1 秒开始，且默认包含 1080p；新增标准/Fast 应配置 4–15 秒，Fast 去掉 1080p，既有 Mini 保持不动。Grok 使用 1K/2K、比例枚举、最多单参考图，关闭 mask、透明背景和不支持的格式；Gemini 先选现有 1K/2K/4K 子集，关闭蒙版/透明背景，参考数量以网关承诺收紧。
2. **三个 Seedance 分别渠道。** [provider.go](../../backend/internal/app/provider.go)从渠道加载 `HeadersJSON`，模型本身没有独立头配置。精确 M 头不同，沿用 Mini、另建标准/Fast 渠道即可；不要把三个型号塞进同一个固定型号头渠道。没有固定 M 头要求的文本型号可共享原生文本渠道。
3. **媒体原生插件已经存在。** [OpenAI 图片](../../plugin-packages/openai-images/manifest.json)、[Grok 图片](../../plugin-packages/xai-grok-images/manifest.json)、[Gemini 图片](../../plugin-packages/google-gemini-image/manifest.json)、[Seedance](../../plugin-packages/volcengine-ark-seedance/manifest.json)分别处理请求和结果。当前 OpenAI/Grok 插件有参考图时已切换 `/v1/images/edits`（前者 JSON `images[].image_url`，后者 JSON `image.url`）；旧的内置说明仍称没有 edit，研究以源码为准。4aion 没有充分逐模型编辑编码合同，不能仅见 edits 路径就宣称兼容。必要时只修改现有插件/能力配置并补对应 fixture，不新增 4aion 总适配器。
4. **Gemini 鉴权可原样复用。** [原生鉴权](../../backend/internal/app/provider_protocol_auth.go)发 `x-goog-api-key`，与本次 `/v1beta/models` 只读结果相符；图片请求走 `/v1beta/models/{model}:generateContent`。通用文档的 Bearer 说明不足以推翻已观测的原生鉴权，暂无改鉴权代码的理由。[出站头校验](../../backend/internal/outbound/outbound.go)禁止自定义 `Authorization` 或 `x-goog-api-key`，也不需要绕过。不要用 Chat 文本解析器接图片。
5. **任务状态机继续复用。** Seedance 插件已创建/轮询/解析 `content.video_url`；[原生轮询](../../backend/internal/app/provider_video_polling.go)间隔 30 秒，满足前次 4aion 建议不频于约 15 秒。`seed/camera_fixed` 已有扩展参数；素材上传 Action、取消、幂等、编辑/延长、视频回调等不在本次简单换型号的已证明范围。

## 开放前仍需的证据

九个精确型号和账号目录、Gemini 原生目录鉴权已核对；还需确认 Gemini 别名实际映射、Grok/各图像编辑编码、逐模型参数、实际价格和限额，再在用户给定次数与预算内分别验收成功、失败、费用和资源保存。文本需按用途核实 SSE usage、工具循环和视觉；视频需校验同一任务轮询/下载，避免重复提交。公开目录、原厂能力、静态配置和既有三模型验收均不能替代新增六模型的真实结果。

本研究没有真实生成、付费调用、上传、素材注册、创建/取消任务或修改代码；没有把任何新增型号标记为已接通。
