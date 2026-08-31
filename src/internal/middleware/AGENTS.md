# Middleware

本目录拥有认证和请求级横切行为。

## Local Rules

- JWT 算法、过期、认证头解析、context key 和 401 响应均是安全合同。
- 不记录 token、认证头或签名密钥；测试使用合成密钥。
- 中间件失败必须终止请求链，成功路径只写入最小身份信息。

## Local Verification

在 `src/` 运行 `go test ./internal/middleware`。
