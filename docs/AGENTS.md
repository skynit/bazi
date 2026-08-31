# Documentation

本目录保存持久化维护者文档；根级 `AGENTS.md` 继续适用。

## Document Ownership

- `testing.md`：唯一测试政策和证据层级。
- `adr/`：架构、流程、测试和长期取舍的决策记录。
- 研究计划/路线图：问题范围、来源和验收，不复制测试命令或 ADR 理由。

## Authoring Rules

- 每个事实只有一个 owner；其他文档使用相对链接。
- 当前状态文档使用现在时，不追加聊天日志或推理过程。
- 链接使用相对路径；命令必须来自真实脚本或配置。
- 普通文档建议少于 1500 字；根 `AGENTS.md` 少于 2500 字符，子树 `AGENTS.md` 少于 1400 字符。
- ADR 路径、状态和必填章节遵循 [ADR policy](adr/README.md)，不要绕过 validator。

## Local Verification

运行 `bash scripts/check-governance.sh`，确认链接、预算、Git ignore、ADR 状态和命令清单一致。
