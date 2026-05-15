# Learnings - Vue Fortune Components

## Design System
- Tailwind v4 with CSS @theme: bazi-red (#C41E3A), bazi-blue (#2B3A42), bazi-paper (#F5F0E8), bazi-ink (#1A1A1A)
- Page background: #FAF8F3 (warm off-white)
- Card border: #E8E3D5 (subtle warm gray)
- Element colors: 金=#FFD700, 木=#228B22, 水=#4169E1, 火=#DC143C, 土=#DAA520
- Typography: system Chinese fonts (PingFang SC, Hiragino Sans GB, Microsoft YaHei)
- Spacing: 0.25rem/0.5rem/0.75rem/1rem increments (Tailwind-ish scale)
- Border radius: 0.5rem-0.75rem cards, 2rem pill buttons

## API Integration
- Axios client at @/api/client with base /api, auto Bearer token, 401→login redirect
- Daily fortune: POST /api/fortune {chart_id, query_date:"YYYY-MM-DD"} → FortuneResponse
- Weekly fortune: POST /api/fortune/weekly {chart_id, start_date} → WeeklyFortuneResponse  
- Monthly fortune: POST /api/fortune/monthly {chart_id, year, month} → MonthlyFortuneResponse
- Chart list: GET /api/charts?page=X&page_size=Y → {charts, total, page, page_size}
- element_trend field is a JSON string (not parsed) — must JSON.parse on frontend

## Backend Response Shape Quirks
- FortuneResponse has many almanac fields but handler only fills: solar_date, day_gan_zhi, element_images
- Weekly/Monthly handlers fill: solar_date, day_gan_zhi, yi_ji, element_images
- Monthly response uses field name "weekly_score" (copy-paste in Go DTO)
- element_trend is JSON-stringified ElementTrendPoint[] (date, score, metal, wood, water, fire, earth)
- No lucky_color, lucky_numbers, wealth_dir in current API response — UI has placeholder props

## Component Patterns
- Vue 3 `<script setup>` with TypeScript
- Props defined with `defineProps<T>()` + `withDefaults()`
- Route query params via `useRoute().query`
- Named exports from .vue files work (export interface for type sharing)
- vue-echarts v7: import VChart, register modules via echarts/core `use()`
