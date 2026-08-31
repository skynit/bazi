# Domain Services

本目录拥有八字、运势、紫微、解释、精度和本地检索等领域实现。

## Local Rules

- 计算规则保持确定性；时间、历法、顺序、nil/空集合和中文文案均可能是公开合同。
- 传统解释与确定性投影分层表达；`not_adjudicated`、`is_outcome_conclusion=false` 等证据边界不得在下游被弱化。
- 规则表和知识数据保留来源、版本和验证状态；不要把研究材料直接变成现实结论。
- 抽取 helper 只收敛真实同构职责；保留流年/流月/流日、不同历法和不同结果类型的领域差异。
- 修改规则或计算顺序时同步修改能因该回归失败的合同测试；不得只改期望值让测试变绿。

## Local Verification

在 `src/` 运行具体 owner 及其同名 `test` 包，例如 `go test ./internal/service/ziwei ./internal/service/ziwei/test`。bazi、fortune 同理；跨服务公共合同变化时只增加直接受影响包。

## Change Record

规则来源、算法边界、证据语义或公共结果结构的长期决策使用 `docs/adr/` 中的 `architecture`、`testing` 或 `simplification` 类 ADR。
