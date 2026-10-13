# Mediacore 安全策略

## 支持范围

安全修复优先应用于当前 Mediacore 开发分支和本工程最新发布版本。历史版本不保证单独回补。

## 报告漏洞

请不要通过公开 Issue 披露尚未修复的漏洞。Mediacore 仓库目前尚未启用 GitHub 私密漏洞报告；部署时需公布本工程管理员的私密联系方式。在渠道公布前，不要把漏洞细节或密钥贴入公开任务。

本工程的报告入口应归属 [NikitaMikhalkov/mediacore-canvas](https://github.com/NikitaMikhalkov/mediacore-canvas)。上游影策的[安全策略](https://github.com/ddcat-ai/open-ai-canvas/blob/main/SECURITY.md)仅用于上游项目问题，不作为 Mediacore 的运营联系渠道。

报告中请包含受影响版本、复现步骤、影响范围和可行的缓解建议。请勿访问、下载或修改不属于你的数据，也不要对公开实例进行持续压测。

维护者会尽快确认收到报告，在验证影响后同步修复进度。漏洞修复公开前，请为项目和用户保留合理的修复窗口。

## 部署者责任

- 生产环境必须使用 HTTPS，并将 `CANVAS_CORS_ORIGINS` 设置为明确的前端 Origin。
- 首个管理员应在服务暴露公网前完成注册，公开注册默认保持关闭。
- 数据目录、PostgreSQL/SQLite、备份和 `.settings-key` 必须限制为服务账号可读。
- 系统模型渠道和 OSS 使用最小权限密钥，并设置供应商侧预算与告警。
- 不要将 `.env`、数据库、日志、备份或真实密钥提交到 Git。
