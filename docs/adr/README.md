# ADR Policy

ADR（Architecture Decision Record）记录长期有效的架构、流程、测试、安全、公共合同或简化取舍。它解释为什么当前方案成立以及放弃了什么，不保存聊天记录、执行流水或逐日进度。

## Directory Lifecycle

```text
docs/adr/
  proposed/<class>/YYYY-MM-DD-topic.md
  implemented/<class>/YYYY-MM-DD-topic.md
  rejected/<class>/YYYY-MM-DD-topic.md
  archived/<class>/YYYY-MM-DD-topic.md
```

允许的 `<class>`：`architecture`、`process`、`testing`、`security`、`simplification`。

- `proposed`：方案尚未实施；使用 `Problem`、`Proposal`、`Alternatives considered`、`Acceptance criteria`、`Risks`。
- `implemented`：已成为当前事实；使用 `Problem`、`Decision`、`Alternatives considered`、`Consequences`。
- `rejected`：明确不采用；使用 `Problem`、`Decision`、`Alternatives considered`、`Consequences`。
- `archived`：曾经有效但不再指导当前实现；从原状态移动后冻结正文，不再编辑。

## When Required

非机械性的架构、流程、公共 API/数据合同、测试策略、安全边界、数据库或跨包长期取舍必须在同一变更中新增或更新一份 ADR。格式化、拼写修正、生成物刷新和没有新决策的机械重命名可豁免。

## Authoring Rules

1. 从 [TEMPLATE.md](TEMPLATE.md) 开始，文件名使用小写英文短语和日期。
2. 第一行是 `# ADR: ...`，第二行必须是与目录一致的 `Status: ...`。
3. 只记录真实比较过的替代方案，不虚构候选。
4. Proposal 实施后移动到 `implemented/`，改为现在时的 `Decision` 和 `Consequences`；不要追加历史日志。
5. 实现变化使 ADR 过时时，在同一变更中更新或归档它。
6. 运行 `bash scripts/check-governance.sh` 校验路径、状态、章节和链接。

提交记录说明发生了什么和怎样验证；ADR 说明为什么这个决定长期成立以及代价是什么。
