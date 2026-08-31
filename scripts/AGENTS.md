# Scripts

本目录同时包含 Bash、Go 和 Node 工具，覆盖治理检查、Docker 静态检查、E2E、截图、研究抓取、OCR、索引和资源生成。

## Local Rules

- Bash 脚本使用 `#!/usr/bin/env bash` 和 `set -euo pipefail`；Go/Node 工具遵循各自入口和依赖文件。
- 读取凭据、网络抓取、外部 API、真实服务写入、生成/覆盖仓库产物和数据库变更属于高风险效果，必须获得目标级批准。
- 每个写入脚本声明输出路径、覆盖策略和失败清理；已有目标默认 fail closed，不静默覆盖。
- 不打印 Bearer token、密码、Cookie 或完整响应；日志只保留复现所需的脱敏证据。
- E2E 仅指向明确批准的一次性环境；不得把任意 `BASE_URL` 当作安全目标。

## Local Verification

- Bash：`bash -n scripts/<script>.sh`，再运行不产生外部效果的最小 dry/static check。
- Go/Node：运行对应目录的聚焦测试或 `--help`/只读模式。
- 仓库治理：`bash scripts/check-governance.sh`。
