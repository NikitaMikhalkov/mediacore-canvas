# Issue tracker: Linear

团队：Nikita96（NIK），ID：a952195b-e521-48de-8888-6a0db8f6cc47。
项目：Mediacore（P-NIK-4）。
项目地址：https://linear.app/nikita96/project/mediacore-1d6254ea120c
代码仓库：https://github.com/NikitaMikhalkov/mediacore-canvas

## 工作流

- 拆票前读取唯一执行入口 `docs/plans/mediacore-execution-plan.md`。原生能力以验收证据完成；只为实际配置、设计、必要局部调整建任务，缺失能力列 needs-info，不自动转成自研任务。
- NIK-269 保存产品需求总览，NIK-268 保存研究完成记录；均不作为整站开发指令。子任务记录具体范围，避免复制第二份完整执行计划。

- 通过已连接的 Linear 工具读取和维护该项目的任务；不在上游仓库或旧项目创建任务。
- 创建前搜索同一目标的已有任务，优先更新已有记录。
- 每个实现任务记录范围、验收条件、依赖、验证证据及代码或文档链接。
- 仓库文档保存研究、术语和决策；Linear 保存状态与依赖，避免维护第二份本地任务台账。
- 技能要求 publish/fetch issue 时，使用 Linear；GitHub 用于代码、分支和 PR。
- 外部 PR 不自动作为任务需求来源。
- 原型在独立 throwaway 分支保存，Linear 任务记录问题、预览方式及用户结论；正式实现单独验证。
- 源码完成、验证完成、部署完成和用户验收分别记录，不互相替代。
