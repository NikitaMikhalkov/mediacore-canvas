# Mediacore 品牌替换与玻璃 UI 调研

基线：上游 `ddcat-ai/open-ai-canvas`，提交 `ede6cc1e707addceff03c2a0ad4bfe1767afa23e`。结论来自固定提交源码、仓库设计文档、既有玻璃样式和许可证原文。本文是静态研究，不是开发任务，也没有执行产品改动、浏览器验收、性能测试或部署。当前唯一执行入口是[执行计划](../plans/mediacore-execution-plan.md)，产品需求以[平台规格](../plans/mediacore-platform-spec.md)为准。

已确认的产品边界：Mediacore 基于上游工程独立部署；旧选型未投入生产，不涉及数据迁移；面向服务器多用户；产品品牌大小写为 **Mediacore**；整体采用玻璃风格，组件、排版和位置保持原样。正式需求见[平台规格](../plans/mediacore-platform-spec.md)。

## 结论

1. 上游已有可配置品牌与三层主题系统。仅在后台设置名称不足以完成品牌切换：前后端内置默认、独立 Agent 名称、恢复默认动作、静态欢迎素材和文档站均是独立落点。
2. 全站存在 Ant Design 与自研 Celadon/React Aria 两套正在迁移中的控件。玻璃风格应先落入共享语义令牌，再映射两个控件体系；本次不顺带继续上游组件替换工程。
3. 现有皮肤可以表达 6/8 位 HEX 表面、明暗颜色和控件参数，但不直接表达多层背景渐变与 backdrop blur 材质。需要有界的玻璃材质令牌与场景适配，不能只保存一份颜色配置，也不宜追加无范围全局样式。
4. 用户可见品牌、技术协议标识、许可证署名、真实第三方品牌是四类不同对象。全面换品牌不等于全文替换字符串；尤其不应把模型、支付渠道、Eagle、Blender、tldraw 等真实名称冒充为 Mediacore。

## 品牌配置与遗漏面

### 已有可用入口

后端 `system_settings.appearance` 存储外观；实际代码 schemaVersion 为 **10**，设计文档仍描述 8，应以代码为准。公开接口返回当前 activeSkin，不公开未启用主题草稿。后台“站点及外观”有品牌识别、登录页、欢迎页、SEO、页尾备案、主题等分区。前端 `useAppearanceStore` 与 `BrandLogo` 消费同一投影。

| 对象 | 当前机制与证据 | Mediacore 改造落点 |
| --- | --- | --- |
| 页面品牌、品牌标识 | `brandName`、`brandSlug`；默认“影策”/`open-ai-canvas` | 产品名为 `Mediacore`；目标 slug 为 `mediacore`，前后端默认值需一致 |
| 明暗 Logo、favicon | 两套 Logo URL，`BrandLogo`；metadata 动态更新 favicon | 替换 `/logo.svg` 内置资产与上传配置默认；图片缺失保持中性占位，不能退回旧 Logo |
| 浏览器与分享 metadata | title、description、keywords、OG、Twitter、theme-color、canonical | 标题及说明同步 Mediacore；不是只改导航栏 |
| 登录/注册/找回 | 共用 `AuthScene`；标题、视频、封面、版权可配置 | 保留当前结构，覆盖各步骤/错误/成功态；审查默认外链视频和封面 |
| Agent 身份 | **独立** `canvas.agentName` 与 `{agentName}` 文案模板 | 前端和后端默认仍为上游品牌；需核对配置值及服务端身份提示词消费链 |
| 邮件 | `auth/email.go` 的注册文案和发件人继承站点品牌 | 注册/重置邮件主题、正文、默认发件人，管理员自填独立名称策略需确定 |
| 更新公告 | 有可配置版本和内容；关闭后回到内置更新日志 | Mediacore 自己的发布说明；保留产品入口，不能继续展示上游为本产品版本承诺 |
| 后台恢复默认 | 硬编码“恢复影策默认品牌标识？”、“已恢复影策默认…” | 操作提示与实际重置目标分别硬编码，属于两个品牌配置落点 |
| 批量导出、诊断包 | 品牌文档要求展示名投影；导出实现另有协议标记 | 可见文件名/README 改为产品名；协议标记按下一节分类处理 |

源码：[前端默认与 metadata](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/stores/use-appearance-store.ts)、[后端 appearance](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/appearance.go)、[BrandLogo](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/components/brand/brand-logo.tsx)、[Agent 前端默认](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/canvas/agent-appearance.ts)、[Agent 后端默认](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/appearance_canvas.go)、[邮件](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/auth/email.go)、[后台恢复](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/pages/admin/settings/appearance-settings-page.tsx)。

### 源码揭示的额外风险

- `brandStudioLabel()` 对内置品牌特判输出 `YINGCE STUDIO`，自定义 slug 输出大写。不能简单把默认名称常量改为 Mediacore 后保留此特判，否则新默认仍返回旧品牌。
- 外观文档称启动时等待外观请求再导入应用；**实际 `main.tsx` 并行启动 `bootstrapAppearance()` 与 `import('./application')`**，`application.tsx` 随后直接 render。源码存在初始内置身份/皮肤先渲染的可能；4 秒外观超时还会采用内置默认。尚未浏览器复现，是否出现闪现仍未知；默认外观资源会影响该首屏状态。
- `/welcome` 是独立 `welcome-application` 入口。欢迎页仍有固定上游 GitHub URL、代码标签 `aria-label="影策创作之旅"`，以及可能包含上游 UI 的 `web/public/welcome/workbench-*.webp` 截图。
- `web/public/welcome/credits.html` 是独立静态页，现有文案包含“返回影策”“不是影策生成的案例”；Blender 作者与 CC BY 署名属于许可信息。该页自身有暗色 CSS。
- 主静态 Web 清单中未发现独立 PWA `manifest.webmanifest`/favicon.ico；当前 favicon 动态来自 Logo。`web/public/inspirations/manifest.json` 是素材目录，不应误认为 PWA。
- `backend/internal/hostupdate/manager.go` 默认仓库为 `ddcat-ai/open-ai-canvas`，安装与状态目录也使用上游标识；后台“代码仓库”有该默认值。这是发布配置差异，当前状态以执行计划为准。

源码：[启动](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/main.tsx)、[应用渲染](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/application.tsx)、[外观 bootstrap](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/services/appearance-bootstrap.ts)、[欢迎页](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/pages/welcome/index.tsx)、[素材署名页](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/public/welcome/credits.html)、[宿主升级器](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/hostupdate/manager.go)。

### 品牌标识分类

| 类别 | 源码范围与约束 |
| --- | --- |
| 自有产品展示 | 上游品牌标识分布于无障碍标签、空态、异常页、邮件、导出说明和帮助文档；产品品牌为 Mediacore |
| 自有产品资产/运营链接 | Logo、截图、视频封面、社群/支持/购买链接、更新公告、欢迎页 GitHub/文档入口逐项配置；未知 Mediacore 地址不能编造 |
| 技术兼容标识 | Go module `yingce/backend`、环境变量、服务名、DB/存储 key、内部事件协议均属技术标识；重命名影响需单独评估 |
| 导出和插件协议 | `CanvasExportFile` 当前 `app: "yingce", version: 4`；`.yingce-plugin` 等格式应确认是否保留互操作。即使无旧生产数据，也仍有上游导入/插件互操作成本 |
| 开源来源 | LICENSE/NOTICE 作者、历史 ADR 引用、第三方许可证保留原始归属；署名不随产品品牌变化 |
| 真实供应商/支付/模型 | OpenAI、Claude、Gemini、支付宝、微信等保留正确名称；渠道呈现为 4aion 是路由产品策略，不把底层模型名重命名为 Mediacore |
| 外部内容 UI | 外部站点 iframe 不可控，宿主外壳和加载/错误态可以统一；不能声称所有跨域页面都能注入玻璃样式 |

导出源码：[canvas-export.ts](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/canvas/canvas-export.ts)。技术标识与品牌区别也见上游[站点外观合同](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/docs/design/site-appearance-and-skins.mdx)。

## 许可与素材边界

根代码是 MIT，允许修改和商用，但复制或分发实质代码时须保留版权及许可声明。`LICENSE` 的 ddcat / Open AI Canvas contributors 不属于应抹除的产品品牌。`NOTICE` 说明第三方包和 `web/public/icons` 的名称/Logo 各归其权利人。此处为许可原文摘要，不是对所有依赖、素材的完整法律审计。[MIT 原文](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/LICENSE)、[NOTICE](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/NOTICE)。

可选 tldraw SDK 的生产使用适用其独立许可；主画布不是 tldraw，当前新绘图默认 Excalidraw，tldraw 路径需要 license key。选择 Excalidraw 不妨碍主画布玻璃化，不应为了统一品牌擅自去掉受许可约束的标识。[tldraw 官方许可](https://tldraw.dev/community/license)、[绘图引擎选择代码](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/canvas/canvas-drawing-engine.ts)。

欢迎页 Blender 开放电影素材采用页面注明的 CC BY 4.0，作者、来源、许可和改动说明属于法定署名范围；当前外链登录视频/封面不能仅因仓库 MIT 就推断取得其素材使用权。Live2D Core/模型同样受各自许可约束。

## UI 系统与覆盖地图

`AGENTS.md` 指向的 `docs/ui-design-system.md` 在基线不存在；真实活规范是 `docs/plans/ui-design-system.mdx`，ADR-0008 记录自研 Celadon 迁移。其上游 Composio 数值基准需被本项目已确认的 Mediacore 玻璃视觉决策取代，但继续保留 token 分层、组件契约、焦点、语义状态和减少动画规则。不能把源码中的上游作者本地设计包路径当成本项目必需依赖。

现有全局入口为 `styles/globals.css`，含三层 CSS token 与大量页面/控件规则；同时存在 `globals-1..7.css`、`admin-ui*.css`、各页面和节点私有 CSS，实际 import 与级联须逐项追踪，不能看到文件名就假定全部生效。`skin-themes.ts` 将皮肤映射 CSS 与 AntD overrides；`AppProviders` 应用主题并通过 `ConfigProvider` 配置 AntD。自研 Select/Tooltip 已使用 React Aria，浮层出入在 `styles/shared/overlays.css`。

| 覆盖组 | 真实入口 / 状态 | 玻璃验收重点 |
| --- | --- | --- |
| 认证 | `/login`、`/register`、`/forgot-password` + 共用 AuthScene | 双栏与表单不移动；验证码、密码、报错、提交中、成功态统一 |
| 公开页 | `/welcome` 独立入口、`/share/canvas/:token`、not-found、route-error | 独立 CSS、分享访客/失效/无权限，品牌及素材署名 |
| 工作台 | `/`、`/create`、`/canvas`、`/projects`、`/inspirations` | 现有侧栏、顶栏、分类、卡片/列表、搜索、空态、加载态 |
| 画布 | `/canvas/:id` | 背景网格、工具条、Dock、小地图、节点、组、连线、选择/拖拽/焦点；不改变坐标/命中区域 |
| 节点与创作工具 | image/text/video/audio/drawing、批量表格、剧本、角色、frame、预演，以及插件节点 | 内部输入、生成/停止、状态徽章、参数弹层；媒体本体不覆玻璃滤镜 |
| Agent | launcher、panel、composer、history、settings、审批、附件、Skills | 组件位置与交互属于产品规格约束；本报告只标出覆盖面 |
| 画布浮层 | 右键菜单、模型/风格选择、资源/角色库、分享、历史抽屉、裁剪/擦除/拆图/光照/镜头/字幕/时间线等 | body portal/嵌套浮层、滚动、键盘、遮罩对比；透明叠加不穿透干扰读字 |
| 项目影视流程 | `/projects/:projectId`、view、chapter、workflow、编辑器 | 表格、脚本、角色/场景、分镜、时间线密度和轨道位置不变 |
| 媒体工具 | `/toolbox/image`、`/toolbox/video`、`/assets`、`/tasks` | 批量选择、上传进度、生成任务状态、媒体预览 |
| 用户 | `/settings`、用户弹出框、钱包/充值/兑换 | 原渠道及金额布局保持；二维码需要足够不透明底，支付状态易读 |
| 扩展 | `/skills`、`/plugins`、`/plugins/eagle`、自带编辑器插件 | 自有外壳统一；真实插件品牌/版权保留；外部 iframe 内部单列边界 |
| 管理端 | `/admin` 下用户、登录日志、渠道、模型、插件、支付、提示模板、公告、资源、积分、兑换码、日志与各设置 | 数据表、筛选器、危险操作、禁用、校验色，保持原密度及固定列 |

路由证据：[router.tsx](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/router.tsx)。UI 基础证据：[活规范](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/docs/plans/ui-design-system.mdx)、[三层 token](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/styles/globals.css)、[皮肤映射](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/skin-themes.ts)、[AntD adapter](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/app-theme.ts)、[Providers](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/components/layout/app-providers.tsx)。

## Mediacore 玻璃设计基准

参考实际旧版 `canvasTrackpad1012/apps/web/src/assets/main.scss`，不是最早 Horizon 示例的营销排版。仅吸收材质，不复制 Element Plus/Vue 实现。

| 材质 | 浅色参考 | 深色参考 |
| --- | --- | --- |
| 基础玻璃 | 白色 64% | RGB 30/38/56，72% |
| 高密度玻璃 | 白色 92% | RGB 24/31/46，94% |
| 次级表面 | 白色 40% | RGB 44/53/74，42% |
| 描边 | 白色 84% | RGB 204/218/255，15% |
| 分隔 | RGB 63/78/108，12% | RGB 204/218/255，13% |
| 阴影 | 0 8px 32px，RGB 43/48/76，7% | 0 12px 40px，黑色 24% |
| 页背景 | #f9fbff + 顶部浅蓝/浅紫径向光 | #101622 + 蓝紫 12% |
| 画布背景 | #fbfcff + 很淡蓝紫 7%/6% | #101622 + 蓝紫各 6% |
| blur | 旧版主令牌 22px | 同令牌；不代表新画布应所有节点强制 22px |

建议材质分层：页面背景光晕；导航/独立浮动条轻玻璃；节点与浮层高密度玻璃；表格、设置、聊天正文维持充分不透明度。选中、错误、禁用继续独立语义，不能全部变透明白。组件边距、字号/行高、控件高度、宽度、gap、flex/grid、fixed/absolute 坐标先冻结；字库替换也可能改变换行，因此首期保留字体度量。圆角、边界、颜色、阴影可改变观感而不移动布局。

**性能影响是待验证假设**：大量节点叠加实时 blur 可能增加渲染成本；本研究未测该成本或 FPS。若执行计划包含视觉实现，应在实际多节点拖拽/缩放下比较。图片/视频内容应保持清晰；嵌套控件避免重复 blur，浮层动效优先 opacity/transform 并尊重 reduced-motion。源码已有视口筛选、LOD、挂载预算。

## 玻璃材质研究观察

待验证问题：在现有布局内，哪档 Mediacore 玻璃材质兼顾品牌一致性与密集画布可读性？布局不属于此次材质比较变量。

- A：忠实当前 Mediacore，蓝紫环境光 + 高密度白玻璃节点/菜单；建议默认。
- B：更清透，导航与空白卡片降低不透明度，节点正文和后台表格保持高密度。
- C：更克制，减弱背景光和 blur，依靠边界/阴影区分节点；用于可读性与性能对照。

如果后续安排视觉比较，A/B/C 应共用同一组件树、尺寸和状态，只改变材质参数；代表场景可覆盖 `/create`、`/canvas/:id`、登录页、设置弹窗和管理员表格。该建议仅描述研究方法，不是独立执行清单。

布局比较应以相同 fixture、字体完成加载后的稳定 DOMRect（x/y/width/height）、滚动容器、层叠与焦点次序为参考。研究未做浏览器比较。

## 完成证明

源码覆盖范围、兼容标识和许可边界如上所列；执行顺序见执行方案，实际实现与验收状态只维护在 Linear 对应任务中。

视觉检查同视口对比关键组件 bounding boxes，确保布局/位置/输入命中区未移动；浅深主题、portal 浮层、焦点/键盘、滚动、错误/禁用/加载、空素材图标与媒体清晰度均需验证。性能以同一机器和同一工程实际测量，不用后端 Go 或 UI 技术栈名称替代数据。研究阶段没有运行这些验证。
