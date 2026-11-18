# Mediacore 服务端与多用户部署调研

调研基线：上游 `ddcat-ai/open-ai-canvas` 提交 `ede6cc1e707addceff03c2a0ad4bfe1767afa23e`。本文是静态研究，不是开发任务，也不是线上安全审计、容量测试或生产验收证明。当前唯一执行入口是[执行计划](../plans/mediacore-execution-plan.md)，产品需求以[平台规格](../plans/mediacore-platform-spec.md)为准。部分工作区文件尚在检出，读取使用 `git show HEAD:path`；一次广域 `git grep` 因 GitHub 对象下载失败未完成，不以它的缺失结果作不存在某功能的证据。

## 已确认的产品边界

- 旧选型未投入生产，不存在生产数据迁移任务。
- Mediacore 独立部署，面向多用户。平台统一持有 4aion 密钥，普通用户消费 Mediacore 积分；管理员管理用户、密钥、中转和模型。
- 优先复用上游已有中转、模型目录、任务和计费；兼容差异作为配置验证或待决事项记录。
- 玻璃风格覆盖产品，但组件、布局、位置保持原样；本文不通过改后端流程来擅自改变 UI 操作路径。

## 结论与建议架构

建议直接以原生 Go 后端作为 Mediacore 用户、权限、模型目录、生成任务和积分账本的唯一权威；保留 PostgreSQL 持久化、Redis 协调、Node Pi Agent runtime、React 客户端。4aion 应接入原生系统渠道及协议层，浏览器只得到公开模型目录与站内接口，不能得到平台密钥。现有源码已有这一分层，全面重写认证或再引入独立钱包会增加对账、故障恢复和运维成本。[仓库边界](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/AGENTS.md)、[原生建任务入口](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_creation.go)、[生产编排](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/docker-compose.deploy.yml)。

这是有服务器多用户基础的产品，而不是仅将个人画布放到服务器。它已经具备认证、管理员、用户资源归属、任务恢复、积分冻结与支付订单；但这些能力不能替代下面列出的故障场景验收。尤其是普通画布生成请求的客户端重试幂等、平台模式下关闭用户自定义渠道、二开自己的发布源，需要在收费上线之前落实。

## 认证、邮箱与会话

| 范围 | 当前源码事实 | Mediacore 落地要求 |
| --- | --- | --- |
| 账号密码 | bcrypt 哈希；密码至少 8 个 Unicode 字符；用户名规则 3–32 位字母、数字、下划线、连字符 | 密码长度上限、错误提示和邮箱作为主身份的产品表达按平台规格核对 |
| 初始管理员 | 用户数为零时，注册者直接取得 admin；随后注册受总开关及验证策略约束 | 首次启动在受控入口完成管理员初始化，完成前不向公网开放注册入口。多副本首次初始化也须验证，进程内 `registrationMu` 不等同数据库全局锁 |
| 注册验证 | 邮箱、短信或双验证策略；默认策略为邮箱注册；新验证 ticket 10 分钟，含用途、用户/联系信息、策略哈希、验证码哈希及尝试记录 | 首期优先邮箱，配置 Mediacore 发件身份、SMTP 及送达监测。不能只把按钮开放、未配置邮件服务 |
| 找回密码 | 邮箱 6 位码、10 分钟；未知邮箱返回成功避免直接暴露是否存在；验证码消费与改密在事务内，清除该用户全部会话 | 可复用。验收重复用码、并发用码、旧会话失效、禁用用户和邮件失败路径 |
| 会话 | 数据库存 session ID 与 token hash；Cookie 为 HttpOnly、SameSite=Lax；检测 TLS 或转发 https 后置 Secure；默认固定 **30 天** | 源码当前默认 30 天；与平台规格确定的期限对齐时，需核对服务端 session 及 Cookie MaxAge |
| 禁用与管理员 | CurrentUser 每次检查用户 active；管理端支持账号角色、禁用、改密和审计，停用清登录态 | 普通用户伪造 admin API、已禁用账户旧标签页和进行中任务须纳入验收 |

证据：[auth.go](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/auth/auth.go)、[验证 ticket](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/auth/verification.go)、[验证策略](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/auth/verification_policy.go)、[找回密码](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/auth/password_reset.go)、[验证码与会话事务](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/repository_users.go)、[HTTP Cookie 与限流](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/handler/auth.go)、[用户管理](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/admin_users.go)。

认证入口已有按 IP、账号等作用域限流，Redis 协调器不可用时相关请求失败关闭。CORS 中间件会拒绝不允许的 Origin；会读取 X-Forwarded-Host，Cookie Secure 判定读取 X-Forwarded-Proto，因此反向代理必须覆盖客户端传入的转发头，并仅从可信代理到达 backend。不能仅凭 SameSite/Lax 和 CORS 就宣称所有 CSRF/代理欺骗场景通过。[限流和代理路径限制](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/handler/security.go)、[CORS 实现](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/cmd/server/main.go)。

## 多用户隔离、缓存与文件

前端有用户 scope 存储键；切换身份时暂停生成消费者、等待本地持久化、复位远程同步，清空资源访问/Blob 缓存和 React Query，重置画布/素材/插件内存并按新 scope 恢复。异步初始化还带 session generation 检查，降低旧请求晚到污染新用户状态的风险。该实现值得保留，不能在全站换肤时改成全局共享缓存。[用户作用域](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/user-scope.ts)、[切换会话实现](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/user-session.ts)。

后端示例读取均带 userID：TaskForUser、任务列表/日志、ResourceForUser、CanvasProjectForUser。资源下载先做归属校验，再生成具体变体和有效期的签名地址；签名地址作为 bearer capability，可在有效期内被持有者访问，因此它不是“每次访问都重新要求用户 Cookie”。给 4aion 拉取输入素材时需正确设置公网 HTTPS Base URL、签名有效期和对应资源权限。[任务仓储](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/repository_tasks.go)、[资源访问](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/resource_access.go)。

画布分享有随机 token、哈希索引、加密保存原 token、到期和轮换；公开资源读取仅允许分享项目所引用且归属分享者的资源。创建时 ExpiresDays=0 表示无到期日；应由运营决定是否保留永久公开分享，而不是认为所有分享链接自动到期。[分享实现](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/canvas/canvas_share.go)。

资源支持本地及 OSS/对象存储，记录实际 provider/bucket/object key。二开后不能只改当前全局存储配置就认为老资源已迁移。大型视频输出必须回存自有对象存储，不能仅保存上游可能过期的 URL；原生已有媒体恢复检查点，优先沿用。[资源存储](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/resource_store.go)、[任务模型中的媒体恢复字段](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_task.go)。

**删除语义存在需记录的文档差异**：AGENTS 写有“有引用则保留、物理删除失败不得删除记录”的一般约定，但此提交 `resource_delete.go` 已明确采用删除 Outbox；用户确认的 `purge=true` 允许让项目/画布/历史引用失效，其他素材复用的对象保留，物理对象异步删除。不能将一般约定当作现有代码行为。需要实测普通删除、回收站、彻底删除、项目历史恢复和删除 worker 重试，避免上线后出现“列表已删、旧引用仍在”误解。[实际删除路径](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/resource_delete.go)。

隔离验收至少覆盖同浏览器 A→B→A 切换、并行标签页、A 的资源/工程/任务 ID 在 B 下访问、已撤销分享、过期签名、下载 Range、删除后的缓存与离线同步重放。上述抽样源码存在归属控制，不等于所有路由已完成越权测试。

## 平台密钥、中转与模型管理

原生系统渠道已把创建/更新/删除置于 RequireAdmin 后；平台 APIKey/SecretKey 经 AES-GCM 加密，根密钥保存在数据目录 `.settings-key`。客户端系统模型配置用 `apiKey: "system"` 和站内 `/api/<channelId>`，真正请求由后端取系统渠道密钥。站内公开模型目录可直接支撑“用户只选模型、不填 Key”的运营模式。[系统渠道](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/admin_channels.go)、[秘密加密与任务秘密保护](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/settings.go)、[公开目录客户端映射](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/user-session.ts)。

原生仍支持用户自定义渠道。仅隐藏前端 Key 输入不足：需禁用 FeatureCustomChannels 并验证直调 `/api/ai/custom`、自定义 task input 和工作流 provider 等入口均不能绕过平台供给政策。通用 CreateTask 会重新解析系统渠道、模型启用状态、能力及价格，不信任客户端提交的协议或定价；该合同应保留。[任务 admission](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_creation.go)。

4aion 是否在所有文本、Agent 工具调用、图片编辑、异步视频、音频、取消和 usage 上与现有协议兼容，需要逐模型合同核对，不能把“OpenAI 兼容 Base URL”当成全面接入完成。具体协议调研见独立渠道报告；这里要求调用最终落入已有任务/计费路径，不能为某工具新增不扣费的浏览器直连捷径。

密钥加密是静态保护，不抵御已取得运行时权限的攻击者；必须同时备份 `.settings-key` 和数据库，权限与备份访问隔离。仅备份 PG 而丢失这个文件，系统渠道、邮件/支付等加密配置可能无法恢复。明文密钥不应出现在普通用户模型目录、任务 InputJSON 对外响应、日志、诊断包或错误信息中；这些是必须实测的验收点。

## 任务持久性、重试、取消与积分

| 边界 | 当前已有机制 | 不能据此作出的推断 |
| --- | --- | --- |
| 建任务与冻结积分 | PG 事务内校验模型、活动任务数、冻结 available→reserved、写账单/流水与 task | 不等于所有客户端重试只会创建一次任务 |
| Worker 接管 | 数据库任务租约；PG `FOR UPDATE SKIP LOCKED`；到期接管；写入附带 owner/expiry fencing | 不等于上游已发出的 HTTP 请求能恰好执行一次 |
| 并发控制 | Redis 全局 worker slots 与渠道协调；任务 lease 定期续租；外部视频轮询可延后调度 | 不是 Redis 作为唯一持久任务队列；DB 才保存任务事实 |
| 结算 | 整数 microcredits；价格快照；固定请求、时长、token 路径；冻结、结算、退款流水；部分差额结算 | 上游 usage 缺失、收费口径不同不能靠估算随意补账 |
| uncertain | 不明确的账单保留冻结，进入人工核对；重试资格明确排除 uncertain；后台巡检输出陈旧订单 | 不等于超时自动退款，也不等于自动重新生成 |
| 取消 | 本地 task cancelled 与 provider cancel 状态区分；尝试远端取消并查询确认 | 停止 UI/断开 SSE 不保证上游停止，也不保证不收费 |
| 回存失败 | 保存加密 media recovery 检查点；恢复可重做资源保存而非重新生成 | 不能以页面没有显示素材判定供应商没完成 |

证据：[建任务与积分事务](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/finance.go)、[账务状态及算术](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/finance_billing.go)、[任务租约](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/repository_tasks.go)、[Worker](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_worker.go)、[计费重试规则](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_billing.go)、[核对巡检](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/billing_review.go)、[取消对账](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_task_cancellation.go)、[失败视频恢复](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/provider_task_recovery.go)。

**静态对比发现普通画布生成提交的幂等差异。** `POST /tasks` 读取请求后直接调用 CreateTask，每次创建新的 task ID；`Task.RequestID` 只有普通索引。taskBillingOrder 的幂等键使用 `task:<taskID>:<newID>`，账单表的 `(user_id,idempotency_key)` 唯一约束无法识别两次独立 HTTP 提交是否来自同一用户动作。充值、CreationSubmission 已有专门幂等语义，但不能据此推断该入口也具备请求级重放保护。[HTTP 入口](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/handler/routes.go)、[Task 字段](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_task.go)、[taskBillingOrder](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/finance.go)、[账单唯一约束](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/model/models_finance.go)。这是源码路径结论，尚未用重复 POST 验证运行现象；完成标准以执行计划为准。

存储配额 admission 有进程内 storageMu、pendingStorage 和用户占用检查；活动任务上限在事务里是 count 后创建。由于这些机制不等同跨副本统一临界区，第一阶段建议单 backend 实例运行、使用其现有 worker 并发。多实例扩容前专项验证配额与活动任务数并发，不把已有 Redis worker slots 推广为所有业务写入已分布式安全。[配额实现](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/storage_quota.go)、[生成前容量检查](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/task_storage_admission.go)。

## 充值与管理运营

原生充值订单支持用户作用域 idempotencyKey、商品金额/积分快照、支付渠道配置、回调验签、通知 inbox、订单查询、关单、对账恢复和原子入积分。已入账重复回调不重复加积分；金额与币种要与订单一致。支付渠道来自校验后的支付插件包，缺可用插件时失败关闭，不能凭界面出现支付宝/微信选项就认定可收款。[支付服务](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/payment.go)、[订单与入账事务](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/repository/payment.go)、[支付渠道注册](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/app/payment_providers.go)。

账本和支付插件机制已存在；首期按平台规格采用管理员发放和兑换码，真实支付暂未开放。任意金额输入的服务端限额、最小货币单位、积分换算及订单快照仍需在相应支付验收中核查。

管理员功能已具备角色校验与操作审计，可复用用户、渠道、模型、资源和账务管理。发布前应明确管理员能否修改余额、谁处理 uncertain、保存何种供应商凭证；“管理员能管理”不代表允许无记录改余额或删账单。插件进程与 Node runtime 也应纳入受信任部署边界，不能将任意第三方插件视为隔离租户沙箱。

## 部署拓扑、发布与恢复

生产 Compose 是 **5 个长期服务**：web/Nginx、Go backend、独立 Node yingce-agent、PostgreSQL 17、Redis 7.4，加一次性 migrate job。只有 web 默认发布 3000；backend 8080 和 agent 8081 留内网。backend 关闭自动迁移，等待 migrate 成功后启动；PG/Redis/data/web chunks 都有持久卷。PG、Redis 镜像标签还不是内容 digest；backend/web 要求自己的发布 digest，Agent 有上游镜像标签回退逻辑。[完整编排](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/docker-compose.deploy.yml)。

backend runtime 镜像仍包含 Node 22 与 Pi 依赖，生产使用独立 Agent 服务；不得在远端 Agent 失败时回退本地重执行。runtime 使用非 root 用户。当前 backend Dockerfile 未安装 FFmpeg；如果开放依赖原生 FFmpeg 的时间线渲染、转码等能力，应检查可执行文件配置并在自有镜像中补齐、做专项验收，不能只看到功能入口就认为部署已有对应二进制。[backend 镜像](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/Dockerfile)、[Agent 镜像](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/yingce-agent/Dockerfile)、[禁止远程失败回退的仓库约定](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/AGENTS.md)。

host updater 默认仓库 `ddcat-ai/open-ai-canvas`，Compose 默认 Agent 镜像也仍指向其 GHCR。这是发布配置与上游的差异；实际更新源和发布流程状态以执行计划为准。内部服务名、Go module 等属于技术标识，不等同于用户可见品牌。[updater 默认源](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/cmd/host-updater/main.go)。

投产配置与可观测证明：

1. 独立域名/HTTPS、明确 CORS Origin、可信反代转发头、PUBLIC_BASE_URL；PG/Redis/backend/Agent 不对公网开放。
2. 受控初始化管理员、邮箱服务、默认关闭未准备好的注册和支付；所有平台 Key 仅后端配置。
3. 独立 PG/Redis/数据卷/对象存储，保存 `.settings-key`；备份数据库、加密根密钥、上传对象、支付配置版本、镜像 digest 与部署配置。
4. 备份必须有离线恢复演练，证明用户、余额、冻结账单、任务和素材均能恢复。程序回滚不自动撤销数据迁移或已经发出的供应商订单。
5. 监控任务排队时间、执行时间、失败率、uncertain 冻结、回存失败、PG/Redis 可用性、磁盘及对象存储增长、SMTP 失败、支付通知积压；健康检查只证明入口可用。

这些是建议的部署交付控制，本次没有触碰任何真实环境或凭证，没有执行部署。

## 性能、成本与维护

默认 runtime policy 有任务数、资源大小/容量、请求速率、worker/渠道并发、Agent 会话与 token/超时等界限。源码默认 worker/channel 并发为 8，但生产 Compose 环境默认覆盖为 3；Agent 默认最大 30 会话，Compose 配有 512 MB Node old-space 参数。这些是配置值，不是已验证的承载能力，不能据此承诺 30 用户同时无卡顿。[runtime policy](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/platform/runtime_policy.go)、[Compose](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/docker-compose.deploy.yml)。

首期维护面包括 React/TypeScript UI、Go/GORM 业务与 SQL 迁移、Node/Pi Agent、PG/Redis、媒体存储/带宽、邮件及支付插件。Go 的部署便利不代表消除了这些成本。性能测量要分开浏览器帧耗时、工程载入与同步、排队、4aion 上游耗时、回存耗时；全站玻璃 UI 对浏览器绘制的影响与 Go API 的吞吐是两件事。服务端已有有界读缓存（条数、字节、并发回源限制），应保留并验证私有数据缓存键包含身份。[有界缓存](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/backend/internal/platform/read_cache.go)。

建议容量验证按活跃用户和并发任务定义，而不是注册账号数：例如空闲 100 个会话、10 个同时编辑、3 个视频生成、多个 Agent SSE，与大量视频节点的单用户画布是不同压力模型。具体机器规格、并发目标和可接受排队时间需要业务决定；本次未测成本或容量，不提供虚构的服务器报价/承载人数。

## 许可证与二开边界

当前根 LICENSE 为 MIT，须保留原版权与许可文本；NOTICE 记录第三方及继承材料。可将面向用户的产品品牌改为 Mediacore，但法律归属/第三方许可不能全部替换成 Mediacore。模型商图标、第三方插件、字体和素材仍各有权利边界。[LICENSE](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/LICENSE)、[NOTICE](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/NOTICE)。

新绘图默认 Excalidraw；源码对 tldraw 单独要求配置 license key 才视为可用，旧无 engine 标记节点按 tldraw 处理。新工程不迁移旧数据，可先使用 Excalidraw，避免把全部依赖都理解为根 MIT。tldraw、Live2D 和内置第三方技能材料有独立依赖/素材许可证边界，与产品品牌切换无关。[绘图引擎合同](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/web/src/lib/canvas/canvas-drawing-engine.ts)。

## 静态证据与验证边界

上游有 Go 测试、PG/Redis 依赖型测试防跳过检查、前端类型/测试、Chrome E2E、支付包多架构构建与 smoke 流程。本次只阅读其存在及关键路径，没有安装依赖、运行测试、构建镜像或触发付费模型调用；CI 定义存在不表示本 fork 当前全部通过。[quality workflow](https://github.com/ddcat-ai/open-ai-canvas/blob/ede6cc1e707addceff03c2a0ad4bfe1767afa23e/.github/workflows/quality.yml)。

本次只读取源码、配置、已有测试定义和部署清单；没有安装依赖、运行测试、构建镜像、访问真实环境或触发付费调用。会话期限、注册/充值策略、分享默认值、用户容量、生产资源及上游发布地址中尚未形成实证的差异，应按平台规格和执行计划验证。报告中的源码事实不等于线上行为已经通过验收。
