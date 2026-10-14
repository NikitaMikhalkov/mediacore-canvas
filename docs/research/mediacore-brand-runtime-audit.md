# Mediacore 品牌改动运行审计

## 结论与边界

未发现已复现、可归因于品牌替换的产品运行故障。检查覆盖品牌提交的全部 44 个文件差异，并运行受影响的外观、身份、认证、插件和审批专项检查。这个结论不等同于全产品无 BUG，也不替代实际生成验收。

本次 `json: unknown field "prompt"` 具有独立的来源证据：上游固定版本已经包含产生此错误的请求拼装与严格解码组合，Mediacore 品牌提交没有修改这些文件。审批修复由并行任务处理，本报告不将其计为品牌回归。

审计使用 `diagnosing-bugs`：先运行现有可执行的行为与合同检查，再沿品牌差异核对边界。本轮没有复现独立的品牌故障，因此没有伪造红色测试，也没有进入针对某个假定品牌根因的假设、修复阶段。部分前端检查为源码合同断言；后端检查包含隔离数据库、资源、身份和插件行为。没有访问生产服务、浏览器、用户数据，未执行付费模型调用、完整测试套件或发布。

## 差异范围

- 固定上游：`26dfb8d91bddf78ada93f7cba079368f1634db6d`（v1.6.7）。
- 品牌提交：`56678b8`、`92aaeee`；完整比较范围 `d6acbdb..92aaeee`，44 文件、110 行新增、97 行删除。
- 来源对照固定到修复前提交 `27c9a93e49fa45a5220bf649a87498ad455ac53a`，不随后续 HEAD 移动。运行检查使用该提交的工作目录及当时并行任务已加入的审批请求字段白名单和对应测试。本报告不修改这些文件。
- 已阅读仓库 `AGENTS.md`、`GLOSSARY.md`，以及 ADR-0005、ADR-0007、ADR-0010。调查遵循保留必要内部标识的既有决定。

| 检查面 | 品牌改动与证据 | 判定 |
| --- | --- | --- |
| 外观默认值、恢复默认 | 前后端默认名称、slug、SEO、版权及 Agent 默认名同步；外观配置仍使用 `appearance` 键和原 schema。外观保存、恢复、旧设置回填、资源保留、Live2D 专项通过 | 未复现配置损坏或资源丢失 |
| Logo 和资源路径 | 修改 `logo.svg` 图形；`BrandLogo` 的默认分支改为统一使用 `/logo.svg` mask。全部调用点均提供尺寸类，资源仍在原路径；自定义 Logo 的失败占位分支保留 | 未发现路径破坏；本轮未做浏览器像素验收 |
| 模块和构建依赖 | 没有重命名 Go module、TS import、包名、锁文件或资源目录；`backend/go.mod`、`go.sum`、`web/package.json`、`bun.lock` 与上游逐字节相同 | 排除品牌替换引起依赖/import 漂移 |
| 插件和存储 scope | `yingce.plugin/v1`、`yingce.plugin/v2`、插件 ID、`yingce:plugin-storage:`、`yingce:active-user-scope`、localforage 的 `yingce/app_state` 保留；相关实现与上游逐字节相同。插件注册、权限、编辑器宿主、平台开关检查通过 | 排除直接重命名导致插件失配和已有缓存不可读 |
| 认证、Cookie、域名 | 保留 `open_ai_canvas_session`、认证 API、`CANVAS_CORS_ORIGINS`、HTTPS Cookie 和 nginx 实现；相关文件与上游逐字节相同。服务协议与密码重置检查通过 | 未发现品牌改动导致登录/找回密码合同失配；生产域名配置不在本轮运行验证范围 |
| 对象存储和文件名 | 品牌 slug 只影响展示、诊断包命名与管理员主动采用的新路径前缀；存储设置页仍保留已保存前缀，未自动重写对象路径 | 未发现既有素材路径被批量迁移 |
| Agent 提示、身份、工具 | Node runtime 仅改 provider 的显示 `name`，`providerID = "yingce"` 未变；媒体策略只改标题，policy ID/version、工具及审批规则保留。创作与剪辑提示仅替换身份名称。Agent 身份测试确认改名不改变工具、权限或冻结历史 | 未发现品牌替换改变工具路由或授权 |
| 安装、镜像和更新 | fork 的源码 URL、默认分支、GHCR owner、更新仓库改为自有来源；包后缀、服务名、环境变量、安装目录及 socket 保留。发布 workflow 的 owner 生成方式与镜像名与脚本一致；更新器相关专项通过 | 当前源码部署链无已复现品牌故障；自有 Release/GHCR 在线发布仍未在本轮验证 |
| 其他显示与来源 | 首页、预演台、积分卡片、字幕样例、技能作者默认文案、OpenAPI 标题、PowerShell 输出、README、安全报告入口和上游贡献者说明逐项检查 | 未修改业务键或命令执行逻辑 |

## 审批错误与品牌的来源区分

对下面文件分别运行 `git show 26dfb8d91bddf78ada93f7cba079368f1634db6d:<文件>` 和 `git show 27c9a93e49fa45a5220bf649a87498ad455ac53a:<文件>`，按完整字节比较，结果全部 `UNCHANGED`。等价差异范围为 `git diff 26dfb8d91bddf78ada93f7cba079368f1634db6d..27c9a93e49fa45a5220bf649a87498ad455ac53a -- <下列六个文件>`，无输出。该结论限定于修复前提交，不适用于包含后续审批修复的新提交：

```text
web/src/services/api/agent.ts
web/src/lib/canvas/agent-media-approval.ts
web/src/components/canvas/canvas-agent-video-approval-settings.tsx
web/src/components/canvas/canvas-cloud-agent-panel.tsx
backend/internal/handler/agent.go
backend/internal/app/cloud_agent_media_approval.go
```

上游 `agentVideoApproval()` 的预览数据本来就带 `prompt`、参考素材和源节点；视频选项编辑器会展开该预览对象；原 `decideAgentApproval()` 不筛字段就提交 `mediaSettings`。后端 `CloudAgentMediaSettings` 不接受 `prompt`，handler 使用 `DisallowUnknownFields()`。这是前端预览对象与写入 DTO 混用的已有合同缺口，与品牌文案、Agent 名称和 4aion 型号无直接因果关系。

工作目录中的并行修复为 API 边界列出可编辑字段；新增 image/video 两项实际请求序列化检查均通过。该事实只能证明本机修复后的请求合同，不代表已部署或付费生成通过。

## 已执行检查

以下前端命令直接使用本机已缓存的 Bun 1.3.9，在 `web/` 执行；没有重新安装依赖。表中 `bun` 指该已核实版本的可执行文件。

```bash
bun test test/appearance-bootstrap.test.ts test/site-appearance-and-skins.test.ts test/canvas-appearance.test.ts test/appearance-refresh.test.ts test/official-application-plugins.test.ts test/plugin-permission-check.test.ts test/agent-media-approval.test.ts test/agent-approval-request.test.ts test/creative-agent-contract.test.ts test/welcome-story.test.ts test/workspace-sidebar-checkin.test.ts
# 68 pass, 0 fail; 11 files; 342 assertions

bun test test/ai-command-schema.test.ts test/auth-registration-agreement.test.ts test/agent-api-reliability.test.ts test/plugin-state-switch.test.ts test/plugin-upload-merge.test.ts test/asset-storage-revision.test.ts
# 55 pass, 0 fail; 6 files; 195 assertions
```

后端在 `backend/` 执行：

```bash
go test ./internal/auth ./internal/hostupdate ./internal/prompts ./internal/app -run 'Test(Appearance|Live2D|RegisterRequiresAcceptedTerms|LinuxDORegistrationAgreement|PasswordReset|CloudAgentIdentity|CloudAgent.*Identity|Plugin)' -count=1
# auth PASS; app PASS
# hostupdate/prompts 此筛选没有匹配项，不能计为这两个包的测试通过证据

go test ./internal/hostupdate ./internal/prompts ./internal/app -run 'Test(LoadAgentPolicies|ParsePolicyDocument|ComposeProjectName|WriteComposeEnvOverride|DeployComposeAgent|.*Appearance|.*ApplicationPlugin|.*EditorShell|CloudAgentIdentity)' -count=1
# hostupdate PASS; prompts PASS; app PASS
```

`git diff --check d6acbdb..92aaeee` 通过。最初通过 `npx` 启动 Bun 等待网络且未产生测试输出，已取消并改为缓存中的相同版本；不将这次未运行结果计为通过。

## 剩余边界与处理建议

继续保留内部兼容标识，无需为了去品牌化重命名协议、数据库、Cookie、包名和服务名。`Mediacore` 与 `MediaCore` 的大小写属于展示口径，不构成当前已发现运行故障。

尚未验证自有 GitHub Release、GHCR 镜像在线更新和 Windows 安装脚本实际执行。这些是发布路径的边界，不是当前画布审批故障的证据；当前 README 已说明先使用源码构建、自有镜像发布后再启用镜像部署。不要直接把上游 tag 视为自有品牌的可用发布版本。

本轮没有品牌源代码修复或新增测试工程，没有遗留调试日志、脚本、数据库或后台进程。唯一新增仓库文件为本报告。审批协议故障的最终部署与 Chrome 原场景回验由主任务记录。
