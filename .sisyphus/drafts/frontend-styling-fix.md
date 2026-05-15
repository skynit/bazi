# Draft: Frontend Styling Fix

## Issues Identified

### FortuneView (运势) — 界面风格不符合规范

**P1: 页面背景与设计系统不匹配**
- `.fortune-page { background: #FAF8F3 }` — 暖白色, 但 App 是深色主题 (`#0A0815`)
- DailyFortune 子组件也使用白色/浅色系卡片
- 整个运势页与主应用的深色主题完全脱节

**P2: text-bazi-* 类不存在**
- 模板中大量使用 `text-bazi-blue`, `text-bazi-red`, `text-bazi-ink`, `text-bazi-blue/60`, `text-bazi-ink/70`, `text-bazi-ink/80` 等
- 这些不是 Tailwind 标准类，也未在 v4 配置中定义
- 所有使用这些类的文本颜色完全不生效，回退为继承色或透明

**P3: .glass-card 在浅色背景上不可见**
- 分析区块使用 `class="glass-card"` → `background: rgba(255,255,255,0.025)` 
- 在 `#FAF8F3` 背景上几乎不可见

**P4: 字体-背景对比度不足（运势解析）**
- `.lucky-label { color: #999 }` → 白底灰字，WCAG AA 失败 (2.84:1)
- `.element-desc { color: #888 }` → 对比度 3.54:1，勉强
- `.yiji-empty { color: #bbb }` → 几乎不可见 (1.87:1)
- `.element-desc.placeholder { color: #ccc }` → 不可见
- `.modal-close { color: #999 }` → 低对比度
- `.lunar-date { color: #888 }`, `.week-day { color: #666 }` → 低对比度
- Data: 白底(#fff) + 灰字(#999) = 2.84:1 (需要 4.5:1)

### ZiWeiView (紫微斗数) — 白屏

**P1: text-bazi-* 类缺失（同上）**
- 模板中同样使用 `text-bazi-blue` 等不存在的类
- 标题、正文颜色全部不生效

**P2: 可能的运行时白屏原因**
- 方案 A: API 调用成功但 chartData 为空 → v-else 渲染但内容为空 → 看到白底页面
- 方案 B: chartData.palaces 为空 → ZiWeiChart 渲染但无内容
- 方案 C: JavaScript 异常导致组件未挂载 → 看到深色 App 外壳
- 用户说"白屏" → 更可能是 A 或 B

**P3: 缺少空数据/部分数据的优雅降级**
- 当 API 返回 chart 但 ziwei 数据为空时，只显示出生信息栏，其余空白

## Design System Reference
- 深色主题: `--bg: #0A0815`, `--text: #F0EDE4`
- 金色: `--gold: #D4A84B`
- 红色: `--crimson: #C41E3A`
- 毛玻璃: `.glass-card`, `.glass-panel`
- 按钮: `.btn-gold`, `.btn-ghost`

## Decisions Needed
- [ ] Fortune/运势页应使用深色主题还是保持浅色？
  - 推荐: 深色主题，与 App 一致
- [ ] text-bazi-* 类应定义为 Tailwind 主题扩展还是 CSS 变量？
  - 推荐: Tailwind v4 `@theme` 扩展
