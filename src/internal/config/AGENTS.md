# Configuration

本包从环境变量构造运行配置，并决定 SQLite/MySQL 模式和服务端口。

## Local Rules

- 默认值和环境变量名属于部署合同；修改时同步检查 Compose、Dockerfile 和入口调用方。
- 不读取、记录或提交真实凭据；测试使用进程内临时环境并负责清理。
- 缺失配置应只在实际需要该能力的边界失败，不阻塞无关本地开发。

## Local Verification

在 `src/` 运行 `go test ./internal/config`。
