# Vue Frontend

本目录是 Vue 3 + TypeScript 前端包。依赖、脚本和 Node 行为以 `package.json`、lockfile 和 Vite 配置为准。

## What Lives Here

- `src/api/`：HTTP client 与领域 API。
- `src/stores/`：Pinia 会话状态。
- `src/router/`：路由与认证守卫。
- `src/views/`：页面数据获取和流程编排。
- `src/components/`：可复用展示与交互组件。
- `public/`：原样发布的静态资源；`dist/` 是生成物。

## Local Rules

- 使用 `<script setup lang="ts">` 和现有组合式 API；不要引入第二套状态或样式框架。
- API 调用通过 `src/api` 的领域函数，页面和组件不直接创建 Axios 实例。
- Props、emits、路由参数、snake_case DTO 和用户可见文案属于前端合同。
- 不手改 `dist/`、锁文件派生内容或生成资源；依赖变化才更新 lockfile。
- 保持现有设计系统、键盘操作、移动端布局和 reduced-motion 行为；视觉修改需要对应视口证据。

## Local Verification

在 `vue/` 对改动文件运行 `npm run lint -- <files>`，对组件运行 `npm test -- <target-spec>`。只有构建配置、依赖、公共资源或生产打包变化才运行 `npm run build`；详见 `../docs/testing.md`。
