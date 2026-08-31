# Vue Views

View 对应路由页面，负责参数读取、数据加载、加载/错误状态和子组件组装。

## Local Rules

- 通过 `src/api` 和 stores 访问外部状态；不要在 view 内复制 client/interceptor 逻辑。
- 异步流程必须处理 loading、空数据、错误和过期响应；不得吞掉影响主流程的异常。
- 页面只保存该路由需要的状态；可复用计算移到 lib/composable 或组件 owner。
- 路由参数先验证再转为数字或日期；用户文案不得暴露内部枚举、源码路径或原始异常。

## Local Verification

运行目标 view 及直接子组件的 spec/定向 lint。交互或布局变化按 `docs/testing.md` 记录浏览器和视口证据。
