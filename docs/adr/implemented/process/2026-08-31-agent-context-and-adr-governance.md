# ADR: Track scoped agent instructions and decision records
Status: implemented

## Problem

仓库原有 20 份 `AGENTS.md` 被通配 `*.md` 忽略，无法随 checkout 和 CI 传播；根与子树重复测试命令并包含过期的 Go、依赖和脚本描述。仓库也没有持久化 ADR 生命周期，长期决策只能散落在计划、聊天或实现中。

## Decision

所有 `AGENTS.md` 纳入 Git：根文件只拥有全局 standing orders、权限边界、测试/ADR 链接和跨层不变量，子树文件只补充局部 owner、规则和最小验证。测试命令和证据层级由 `docs/testing.md` 唯一拥有。

ADR 使用 `docs/adr/<status>/<class>/YYYY-MM-DD-topic.md`，状态由路径和第二行 `Status:` 同时表达。非机械性的架构、流程、测试、安全、公共合同或长期简化取舍必须在同一变更中新增或更新 ADR；治理脚本检查 AGENTS 预算/链接及 ADR 路径、状态和必填章节。

## Alternatives considered

**继续把 AGENTS 保留为本机生成文件。** 这避免增加仓库文件，但其他 checkout、贡献者和 CI 无法获得同一约束，过期事实也没有 review 或 gate。

**只跟踪根 AGENTS。** 文件更少，但 Go、Vue、脚本、迁移和文档的局部风险会重新挤入根文件，或退化为未跟踪的隐性规则。

**只使用提交信息记录决策。** 提交适合记录行为和验证，不适合表达可按状态、类别检索且随实现更新的长期取舍。

## Consequences

新 checkout 能获得一致且分层的 agent 指令，重复规则和版本漂移可以被机械检查。维护者需要在局部边界变化时更新对应 AGENTS，并为非机械性长期决策维护 ADR；仓库因此增加少量文档和 gate 成本。现有全量 CI、分支保护和真实环境验证仍是独立能力，不能由本决策推断为已存在。
