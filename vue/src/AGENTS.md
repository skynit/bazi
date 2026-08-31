# Frontend Source

`src/` 负责应用启动、路由、状态、API、页面和组件；根级及 `vue/AGENTS.md` 规则继续适用。

## Local Boundaries

- `api/` 拥有 wire DTO 和请求封装。
- `stores/` 拥有跨页面会话状态。
- `views/` 获取数据并组装页面流程。
- `components/` 接收明确 props、发出 events 并渲染交互。
- `router/` 拥有路径、懒加载和访问控制。

## Local Rules

- 将可测试的领域映射放在明确模块中，不把请求、状态和复杂展示都塞进一个组件。
- 不在多个页面复制认证、错误解析或 API URL；复用对应 owner。
- 修改跨层类型时同时检查 API 类型、页面映射、组件 props 和测试夹具。

## Local Verification

运行变更文件的定向 lint 和最邻近 spec；没有 spec 的行为变化应新增一个能因回归失败的测试。
