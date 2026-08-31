# Vue Components

本目录拥有可复用业务组件和 UI primitives。

## Local Rules

- Props 和 emits 是公共组件 API；保持类型明确，避免无说明的 `any`。
- 数据获取和路由跳转默认由 view/父组件负责；组件只承担可复用展示或局部交互。
- 抽取 composable/helper 仅在它隔离真实职责或被多个消费者共同使用时进行。
- ECharts 等库按需注册；挂载、observer、timer 和 listener 必须在卸载时释放。
- 用户可见传统解释保持证据边界，不能把规则标签渲染成现实结论。

## Local Verification

在 `vue/` 运行 `npm run lint -- src/components/<file>` 和 `npm test -- src/components/<file>.spec.ts`。布局变化另需桌面与 390px 视口检查。
