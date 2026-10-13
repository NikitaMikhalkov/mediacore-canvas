# Mediacore 首期公开注册与商业运营：代码证据审查

基线：`ddcat-ai/open-ai-canvas@26dfb8d91bddf78ada93f7cba079368f1634db6d`。范围：只读源代码和官方资料；未启动产品、未运行测试、未调用付费模型，以下“已有实现”不等于部署验收通过。首期已接受：公开注册、积分及兑换码运营、覆盖现有创作入口、优先复用上游。

## 结论

上游已有多用户、注册验证、额度账本、兑换码、私有资源、管理员设置、任务恢复及 Linux 部署基础；无需另建认证、账务或任务系统。首期尚不能据源码宣称可公开运营：必须闭合管理员初始化、邮箱验证、三镜像来源、授权例外、跨账号验收、真实 4aion 成本结算与恢复演练。

## 最小配置与具体风险

1. **先受控创建管理员，再公开入口。** 首个账号绕过后续用户的注册开关/邮箱验证分支，并被赋予 admin；`registrationMu` 是进程内锁，初始化阶段不宜多实例并发开放。证据：[auth.go:149–226](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/auth.go#L149)。这不是要求重写认证，而是部署顺序要求。
2. **公开注册不仅是环境变量。** 后续注册要求注册开关及验证码策略/邮件验证；数据库已有 registration 设置时优先于环境变量。应配置并实际验收邮件投递、验证码、找回密码、域名允许策略和 Mediacore 服务条款。证据：[auth.go:158–196](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/auth.go#L158)、[registration.go:122–153](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/auth/registration.go#L122)。条款必须勾选的服务端校验已存在（auth.go:131）。注册入口已有按客户端 IP 的频率限制：[handler/auth.go:36–46](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/handler/auth.go#L36)，但未验证实际反代链路 IP 和公开滥用承受能力。
3. **首期复用积分与兑换码。** 积分发放带 referenceKey、账本及事务；兑换码使用现有 repository 实现，不需要接入自动支付才能售卖额度。证据：[finance.go:26](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/repository/finance.go#L26)、[finance_redeem.go:136](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/repository/finance_redeem.go#L136)。必须运行一次发码→兑换→重复兑换拒绝→成功扣费→失败退款→不明费用待核对链路，不能把“存在兑换码页面”当财务验收。
4. **多用户基础是用户归属，不等于组织协作租户。** 项目、资源查询明确组合 id 与 user_id：[repository_projects.go:38](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/repository/repository_projects.go#L38)、[repository_storage_resources.go:192](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/repository/repository_storage_resources.go#L192)。首期建议按现有个人账号/项目隔离运营，不额外自研组织与团队权限。抽样证明归属查询存在，不能替代全接口权限审计。必须用 A/B 两账号检查项目、画布、任务、素材、下载、Agent 记忆以及退出切换缓存；共享链接是显式例外，需要单独验收。
5. **平台密钥复用服务端加密。** AES-GCM 与 `.settings-key` 实现在 [settings.go:369–453](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/settings.go#L369)；丢失该文件会影响旧设置/恢复数据解密。数据库与数据目录需要成套保护/备份。普通用户自定义渠道与平台渠道是不同密钥边界；4aion 平台商业账户应由管理员渠道配置，不将平台密钥发到用户端。
6. **管理员和插件信任需保留现成边界。** 插件上传路由包含 RequireAdmin，见 [handler/plugin.go:107](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/handler/plugin.go#L107)；上传清单不能自行选择权限作用域，[plugin_management.go:36](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/plugin_management.go#L36)。Web 插件声明 sandbox/worker 仅是 Web 层约束，[protocol/package.go:157–195](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/protocol/package.go#L157)；支付 RPC 插件通过宿主子进程运行，[payment/rpc.go:111–124](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/payment/rpc.go#L111)，不能据“插件”二字认为有 OS 沙箱。首期无需开放陌生支付插件；只安装已审查来源，管理员权限应视作高信任权限。

## 许可证、品牌及可选组件

- 根许可证为 MIT，含允许修改及销售和保留版权/许可文本的条件：[LICENSE:1](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/LICENSE#L1)。NOTICE 明确三方包及图标有各自权利：[NOTICE:1](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/NOTICE#L1)。因此可把产品品牌改为 Mediacore，但不应机械替换 LICENSE/NOTICE 内作者和法律归属；UI 品牌与源代码来源声明分开处理。
- 绘图已有默认 Excalidraw 选择：[drawing_engine.go:38](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/app/drawing_engine.go#L38)。依赖仍含 tldraw 5.2.5：[web/package.json:80](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/web/package.json#L80)。[tldraw 官方许可](https://tldraw.dev/community/license)要求商业生产取得商业许可（评估试用不能作为长期运营策略）。最小方案先复用 Excalidraw；若“全部功能”要求包括 tldraw 引擎，需确认商业授权并实际验证有效域名/密钥。关闭 UI 不等于已完成打包分发条款审查。
- Cubism Core 已随静态文件发布，非项目 MIT：[web/public/live2d/NOTICE.txt:1](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/web/public/live2d/NOTICE.txt#L1)、[外观文档:22](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/canvas-agent-appearance.mdx#L22)。默认动态球可复用；启用 Live2D 需核对 SDK 使用场景及具体模型授权，不能用根 MIT 代替。当前 overview 与专题文档有旧描述差异，应以实际文件和构建为准。
- **InfiniteTalk 不宜泛化结论。** 本固定源码全仓普通文本搜索未找到 InfiniteTalk 字样，未证明它是内置必备服务；二进制/外部服务不能由该搜索排除。[官方 GitHub](https://github.com/MeiGen-AI/InfiniteTalk)当前写模型为 Apache 2.0，且有 [LICENSE.txt](https://github.com/MeiGen-AI/InfiniteTalk/blob/main/LICENSE.txt)；[官方演示页](https://meigen-ai.github.io/InfiniteTalk/)又称部分素材和生成内容限学术使用。该页限制可能针对演示素材，不能推断所有模型输出禁商用，也不能忽略。若 4aion/插件实际提供此能力，应锁定具体模型/权重/供应商服务条款并澄清适用范围；禁止把演示素材直接拿来商用。它应登记为条件项，而非阻断全部画布的无证据结论。

## Linux、镜像与恢复

- 标准生产编排是 PostgreSQL 17、Redis 7.4、迁移 job、backend、web、yingce-agent，默认只暴露 web 3000：[docker-compose.deploy.yml:1](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docker-compose.deploy.yml#L1)。配置明确 HTTPS 反代、CORS、公网 URL、私有上游精确白名单和数据卷；不要为了 4aion 随意打开全私网访问。
- **源码构建 override 不完整覆盖第三镜像。** [docker-compose.build.yml:1](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docker-compose.build.yml#L1)只定义 migrate/backend/web 的 build，yingce-agent 仍来自 deploy 的 ddcat 默认镜像：[deploy:40](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docker-compose.deploy.yml#L40)。Mediacore 要构建、记录并固定 web/backend/agent 三个镜像同一来源与 digest；内部服务名可暂保留，避免无意义改协议标识。上游 updater 的下载源、镜像源和 socket 需要一起审查，否则后续自动更新可能覆盖品牌和适配。
- 已有迁移前备份实现：生成 PostgreSQL dump 和整个 backend /data tar，并校验 ZIP/checksum：[hostupdate/manager_ops.go:406–469](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/hostupdate/manager_ops.go#L406)。回滚确实调用数据库恢复：[manager.go:367–382](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/backend/internal/hostupdate/manager.go#L367)。不能说上游完全没有备份，也不能把这些代码等同现服务器已具备定时异地备份；该 ZIP 不代表外部 OSS 对象及全部宿主配置已备份。
- 任务在 DB，Redis 协调。交付失败有加密恢复记录、暂存复用、重下载和有限重试；源地址过期且无暂存时不承诺恢复：[task-runtime-stability.mdx:18–35](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/task-runtime-stability.mdx#L18)。跨实例需要共享暂存和稳定 settings-key；首期单 backend 配合既有 worker 并发最省改动。恢复测试必须覆盖停机中已有请求“不重复生成/扣费”，而非只看重启成功。
- 不明费用不能自动视为失败退款，上游已明确待核对边界：[task-runtime-stability.mdx:10](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/task-runtime-stability.mdx#L10)、[同文件:66](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/docs/content/docs/backend/task-runtime-stability.mdx#L66)。4aion 验收尤其需要保留 request/task ID 与账本一致性。

## 测试现状与剩余验收

[质量工作流](https://github.com/ddcat-ai/open-ai-canvas/blob/26dfb8d91bddf78ada93f7cba079368f1634db6d/.github/workflows/quality.yml#L1)含 Go tests、PostgreSQL/Redis 依赖检查、支付插件双架构 smoke、前端 typecheck/lint/tests 和 Previs Chrome E2E。其 PostgreSQL 测试版本 16 与部署 17 不同；工作流存在不代表本 SHA 在本环境通过。此次未运行任何测试、构建、部署或实际服务调用。

首期上线完成证明应为：

1. 在目标 Linux 架构构建三镜像、迁移空库、受控管理员初始化、HTTPS 登录/公开注册/找回密码。
2. A/B 用户归属隔离与管理员接口拒绝；公开分享例外和登录切换缓存验证。
3. 4aion 每项已有创作入口的成功、异步轮询、SSE 重连、取消/超时、失败退款及不明费用待核对。
4. 积分发放、兑换码一次性消费/过期/停用和流水导出可核对；外部收款流程与额度交付可追溯。
5. PostgreSQL+data/key+外部素材备份恢复到隔离实例，核对账户、余额、画布、资源及未完任务；迁移失败与回滚演练。
6. 真实浏览器跑通用户要求的完整功能清单，标注需外部服务/商业许可的选项；未知不计入“通过”。

这些是待执行验收，不是本次已完成成果。最小原则允许采用成熟配置、上游策略和既有恢复机制；不允许用“第一阶段”推迟公开商业运营所需的真实验证。
