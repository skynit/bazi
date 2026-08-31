# Commands

`cmd/` 包含服务入口和离线工具；入口负责配置、依赖组装、路由/任务注册及进程生命周期。

## Local Rules

- 领域算法放在 `internal/service`，持久化放在 `internal/store`；命令只编排。
- 命令参数、退出码、输出文件和 stdout/stderr 是 CLI 合同。
- 默认不启动服务、修改数据库、抓取网络或写外部目录；这些效果需要明确目标和授权。

## Local Verification

运行受影响命令包的测试，例如 `go test ./cmd/precision-report`；没有测试的入口至少执行编译或只读参数检查，并明确证据限制。
