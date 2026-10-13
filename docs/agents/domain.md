# Domain docs

本工程使用 single-context：根 `GLOSSARY.md` 保存领域术语；`docs/adr/` 保存架构决策，沿用现有编号与格式。

探索代码前读取词汇表及相关 ADR。文件缺失时继续工作，由 domain-modeling 在术语明确或决策成立时按需创建，不预建空文档。

问题、规格、实现与沟通采用词汇表中的同一术语。词汇表只定义领域概念，不存放实现细节或待办。决策与既有 ADR 冲突时明确指出，再按确认后的决定更新或替代，不静默覆盖。

上游用户和开发者文档继续放在 `docs/content/docs/`；本工程调查证据放在 `docs/research/`，阶段范围和访谈决策放在 `docs/plans/`，通过入口文档互相链接，任务状态以 Linear 为准。
