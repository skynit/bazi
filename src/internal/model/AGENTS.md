# Models And DTOs

本目录拥有 GORM 实体、共享请求/响应 DTO 和持久化 JSON 结构。

## Local Rules

- JSON/GORM tag、字段空值和类型属于持久化或 wire 合同；变更时检查 handler、service、store 和迁移。
- 不把仅供展示的派生字段写入数据库实体，除非有明确查询或重放需求。
- 新数据库字段必须有迁移策略；公开 DTO 变化必须有合同测试或调用方验证。

## Local Verification

在 `src/` 运行 `go test ./internal/model`，并补充直接生产者/消费者包。
