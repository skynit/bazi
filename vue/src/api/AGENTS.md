# Frontend API Layer

本目录拥有 Axios client、认证拦截器、wire DTO 和领域请求函数。

## Local Rules

- 所有请求复用 `client.ts`；调用方只使用领域函数，不拼接 URL 或认证头。
- API 字段保持后端 snake_case；仅展示所需的 camelCase 转换放在明确的 view/lib 映射层。
- 超时、401、取消和错误消息是共享行为；修改前检查所有消费者。
- 类型必须反映真实可空/可选字段，不用 `any` 隐藏后端合同变化。
- 不记录 token、请求认证头或包含个人信息的响应体。

## Local Verification

在 `vue/` 运行目标文件 lint；client/interceptor 变化运行 `npm test -- src/api/client.spec.ts`，领域 DTO 变化补充直接消费者测试。
