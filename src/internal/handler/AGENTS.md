# HTTP Handlers

Handler 解析请求、调用 service/store 并映射稳定的 HTTP/JSON 合同；领域计算留在 service。

## Local Rules

- 通过依赖注入取得 service/store；不要在 handler 内创建数据库连接或复制算法。
- 认证身份从 middleware 写入的 Gin context 获取；不要信任客户端提交的用户 ID。
- 请求字段、状态码、错误 JSON、空值和响应顺序是公共 API 合同，修改前检查前端消费者。
- 使用 `ShouldBindJSON` 等结构化解析，不手写 JSON 或字符串拼接。

## Local Verification

在 `src/` 运行 `go test ./internal/handler ./internal/handler/test`；如果修改 service 或 model 合同，再运行对应直接包。
