# Pinia Stores

本目录拥有跨页面会话状态；当前认证 store 管理 token、用户和登录生命周期。

## Local Rules

- Store 只保存跨页面状态，不缓存可由当前 props/API 响应直接推导的数据。
- 登录、注册、登出和 401 清理必须保持 token 与用户状态一致。
- 持久化 key 和恢复语义属于客户端合同；变更时考虑刷新、过期和多标签页。
- 不记录或展示 token，不把额外敏感字段写入浏览器存储。

## Local Verification

运行 store 文件定向 lint；会话生命周期变化必须增加 store/client 联动测试。
