# Go Backend

本文件补充根级规则，只适用于 `src/`。Go 版本和依赖以 `go.mod` 为准；容器工具链必须与其保持一致。

## What Lives Here

- `cmd/`：服务和离线工具入口。
- `internal/handler/`：HTTP 参数解析与响应装配。
- `internal/service/`：八字、运势、紫微和证据计算。
- `internal/store/`、`internal/model/`：持久化与跨层数据合同。
- `internal/config/`、`internal/middleware/`：配置和认证边界。
- `migrations/`：数据库迁移与种子数据。

## Local Rules

- 依赖方向保持为 handler -> service/store -> model；handler 不承载领域算法。
- `internal` 包不从 `cmd` 导入；命令入口只做组装、配置和生命周期管理。
- 新依赖必须有当前需求，禁止为单一调用引入框架或通用注册层。
- 数据库行为必须考虑 SQLite/MySQL 差异，并明确本次实际验证的数据库。

## Local Verification

在 `src/` 运行受影响包测试，例如 `go test ./internal/service/ziwei`；禁止用 `go test ./...` 代替范围选择。命令选择和证据含义见 `../docs/testing.md`。
