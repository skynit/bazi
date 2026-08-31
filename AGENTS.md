# bazi

bazi 是 Go + Vue 3 的八字、运势与紫微斗数 Web 应用。修改代码前先定位所属子树的 `AGENTS.md`；测试策略以 [docs/testing.md](docs/testing.md) 为唯一来源，架构或流程决策以 [ADR policy](docs/adr/README.md) 为唯一来源。

## Repository Map

| Path | Ownership |
|---|---|
| `src/` | Go API、领域计算、持久化与命令入口 |
| `vue/` | Vue 3 前端、组件、路由和 API 客户端 |
| `scripts/` | 本地检查、研究、生成、抓取与 E2E 工具 |
| `docs/` | 测试政策、ADR 和维护者文档 |
| `library/` | 只读命理参考资料 |
| `.github/workflows/` | CI 检查定义 |

## Standing Orders

- 未收到明确修改指令时保持只读；实施变更时无需为旧实现保留兼容层。
- 保留工作区已有修改和未跟踪文件；不得擅自 `stash`、`reset`、`checkout`、提交或改写历史。
- 重构必须保持输入、输出、副作用、错误、顺序和文案合同；行为变化与清理不得混在同一改动中。
- 每条持久规则只有一个 owner；其他文件使用相对链接，不复制命令、版本或理由。

## Execution Boundaries

在工作区内读取、编辑、运行聚焦测试/构建/lint 和只读 Git 命令属于普通操作，无需额外批准。

发布、部署、访问凭据、修改系统或服务状态、对真实服务写数据、删除数据、调用付费/外部 API、以及工作区外写入必须先获得针对目标和影响范围的明确批准。缺少批准时只阻止该高风险动作，不阻止本地开发。

不得打印 token、密码、Cookie、认证头或 `.env` 内容。对外部输入和生成脚本设置明确目标、输出目录与失败边界。

## Verification

- 只运行受影响包或文件的最小检查，不运行全量测试；精确命令和证据边界见 [docs/testing.md](docs/testing.md)。
- 绿色源代码检查不是浏览器、容器、部署、真实数据库或真实设备证明；报告必须区分证据层级。
- CI 只重放检查，不生成或刷新期望产物。

## Change Records

- 每个非机械性的架构、流程、公共合同或长期取舍变更，必须在同一变更中新增或更新一份 ADR；机械格式化和无决策的拼写修正可豁免。
- ADR 状态由目录表达，格式和生命周期见 [docs/adr/README.md](docs/adr/README.md)。
- 提交信息使用常见前缀（`feat`、`fix`、`refactor`、`test`、`docs`、`ci`）；正文记录实际行为、失败模式和验证，不保存推理过程。

## Product Invariants

- 后端在 `DB_HOST` 为空时使用 SQLite，否则使用 MySQL；数据库相关结论必须说明验证的后端。
- 用户可见的传统规则、评分和证据标签不得表述为现实结果、概率或决策建议。
- API、持久化模型和前端 DTO 的字段/空值/顺序属于跨层合同；修改时必须同时检查生产者、消费者和相关测试。
