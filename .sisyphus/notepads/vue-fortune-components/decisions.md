# Decisions - Vue Fortune Components

## Component Split
- DailyFortune is a separate display component (not page-level) — reusable if needed from other contexts
- FortuneChart is a separate ECharts wrapper — shared by WeeklyFortuneView and MonthlyFortuneView

## API Data Handling
- WeeklyFortuneView and MonthlyFortuneView parse element_trend JSON string at computed() level
- DailyFortune accepts all possible props with defaults — gracefully handles partial API data
- Future API enhancements (lucky_color, score in daily response) won't break UI

## YiJi Parsing
- yi_ji field is a string like "宜: 出行, 嫁娶 忌: 动土, 安葬"
- Parsed with regex in DailyFortune: /宜[:：]?\s*(.+?)(?:忌|$)/ and /忌[:：]?\s*(.+)/
- Split by Chinese/comma separators

## Chart Configuration
- Single ECharts instance with 6 series (score + 5 elements)
- Dual Y-axis: left for score (0-100), right for element percentages (0-100)
- Score line: red (#C41E3A), thicker (2.5px), with dashed baseline at y=60
- Element lines: colored, thinner (1.5px), smooth, no symbols
- Date labels truncated to MM/DD format for readability

## Routing
- /fortune?chart_id=X → daily fortune
- /fortune/weekly?chart_id=X → weekly fortune (new route)
- /fortune/monthly?chart_id=X → monthly fortune (new route)
- /history → chart list with pagination
- Click chart card → /chart/:id (existing route)
- "查看运势历史" button → /fortune?chart_id=X (reuses daily fortune flow)

## LSP / Build
- No vue-language-server or typescript-language-server installed in environment
- No node_modules (dependencies not installed)
- Manual review used for verification (imports, types, template bindings)
- tsconfig has noUnusedLocals — verified all imports are consumed
