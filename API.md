# 八字算命 API 对接规范

当前后端实际监听端口由 `SERVER_PORT` 控制，默认 `8088`。前端 Vite 开发服务器默认 `5174`，通过 `vue/vite.config.ts` 将 `/api` 代理到 `http://localhost:8088`。

| 环境 | Base URL | 说明 |
|------|----------|------|
| 后端直连 | `http://localhost:8088` | 调试 Go API |
| 前端调用 | `/api` | `vue/src/api/client.ts` 统一 axios baseURL |
| Docker/Nginx | `/api` | 由 nginx 反代到 backend |

除 `GET /health`、`POST /api/auth/register`、`POST /api/auth/login` 外，所有 `/api/*` 接口都需要：

```http
Authorization: Bearer <jwt>
Content-Type: application/json
```

错误响应统一使用兼容式结构：`error` 保留为旧前端可直接展示的字符串，`code/message` 供新代码按稳定错误码处理。

```json
{"error":"错误描述","code":"INVALID_REQUEST","message":"错误描述"}
```

常用错误码：`INVALID_REQUEST`、`UNAUTHORIZED`、`FORBIDDEN`、`NOT_FOUND`、`CONFLICT`、`SERVICE_ERROR`、`SERVICE_DISABLED`。

## 端点矩阵

| Method | Path | Auth | 前端封装 | Handler |
|--------|------|------|----------|---------|
| GET | `/health` | 否 | 无 | `src/cmd/main.go` |
| POST | `/api/auth/register` | 否 | `vue/src/api/auth.ts` | `handler.AuthHandler.Register` |
| POST | `/api/auth/login` | 否 | `vue/src/api/auth.ts` | `handler.AuthHandler.Login` |
| GET | `/api/auth/me` | 是 | `vue/src/api/auth.ts` | `handler.AuthHandler.Me` |
| POST | `/api/chart` | 是 | `vue/src/api/chart.ts` | `handler.ChartHandler.Chart` |
| GET | `/api/charts` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.ListCharts` |
| GET | `/api/charts/:id` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.GetChart` |
| POST | `/api/fortune` | 是 | `vue/src/api/fortune.ts` | `handler.FortuneHandler.CalculateDaily` |
| POST | `/api/fortune/weekly` | 是 | `vue/src/api/fortune.ts` | `handler.WeeklyFortuneHandler.Weekly` |
| POST | `/api/fortune/monthly` | 是 | `vue/src/api/fortune.ts` | `handler.MonthlyFortuneHandler.HandleMonthly` |
| POST | `/api/fortune/ai` | 是 | 暂无页面调用 | `handler.AIStubHandler.AnalyzeFortune` |
| GET | `/api/fortune/history` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.FortuneHistoryList` |
| POST | `/api/ziwei/chart` | 是 | `vue/src/api/ziwei.ts` | `handler.ZiWeiChartHandler.Calculate` |
| POST | `/api/ziwei/period` | 是 | `vue/src/api/ziwei.ts` | `handler.ZiWeiPeriodHandler.Period` |
| POST | `/api/ziwei/overlay` | 是 | `vue/src/api/ziwei.ts` | `handler.ZiWeiPeriodHandler.Overlay` |
| POST | `/api/interpretation/bazi` | 是 | `vue/src/api/interpretation.ts` | `handler.InterpretationHandler.Bazi` |
| POST | `/api/feedback` | 是 | `vue/src/api/feedback.ts` | `handler.FeedbackHandler.Create` |
| GET | `/api/feedback/summary` | 是 | `vue/src/api/feedback.ts` | `handler.FeedbackHandler.Summary` |

## 认证

### 注册

```http
POST /api/auth/register
```

```json
{"username":"test","email":"t@example.com","password":"test123"}
```

响应：

```json
{"token":"eyJ...","user":{"id":1,"username":"test","email":"t@example.com"}}
```

### 登录

```http
POST /api/auth/login
```

```json
{"username":"test","password":"test123"}
```

响应：

```json
{"token":"eyJ..."}
```

### 当前用户

```http
GET /api/auth/me
```

响应：

```json
{"user":{"id":1,"username":"test","email":"t@example.com"}}
```

## 八字命盘

### 创建命盘

```http
POST /api/chart
```

```json
{
  "birth_year": 2003,
  "birth_month": 4,
  "birth_day": 15,
  "birth_hour": 14,
  "birth_min": 0,
  "calendar_type": "SOLAR",
  "gender": "MALE",
  "name": ""
}
```

约定：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `birth_year` | int | 是 | 公历或农历出生年 |
| `birth_month` | int | 是 | 出生月 |
| `birth_day` | int | 是 | 出生日 |
| `birth_hour` | int | 是 | 0-23 小时 |
| `birth_min` | int | 否 | 默认 0 |
| `calendar_type` | string | 是 | `SOLAR` / `LUNAR` / `BAZI` |
| `gender` | string | 是 | `MALE` / `FEMALE` |
| `name` | string | 否 | 空字符串允许 |

响应包含新建 `id`、四柱、五行、身强、十神、纳音、大运等。创建响应中的大运字段为 `da_yun`；历史命盘详情的持久化字段为 `da_yun_start`，前端需要兼容两者。

### 命盘列表

```http
GET /api/charts?page=1&page_size=10
```

响应为 `ChartSummaryResponse` 列表，不返回数据库用户字段或 GORM 元数据：

```json
{
  "charts": [
    {
      "id": 87,
      "name": "",
      "gender": "MALE",
      "birth_year": 2003,
      "birth_month": 4,
      "birth_day": 15,
      "birth_hour": 14,
      "birth_min": 0,
      "calendar_type": "SOLAR",
      "created_at": "2026-06-16T02:09:23+08:00",
      "updated_at": "2026-06-16T02:09:23+08:00"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 10
}
```

分页约定：`page < 1` 会按 `1` 处理；`page_size < 1` 或 `page_size > 100` 会回退为 `10`。

### 命盘详情

```http
GET /api/charts/:id
```

响应为 `ChartDetailResponse` DTO，不直接返回 `BirthChart` 存储模型；包含基础出生信息、四柱与分析 JSON、`da_yun_start`、兼容别名 `da_yun`、`ziwei_result`、`ziwei_computed`。不会返回 `user_id`、`DeletedAt`、关联 `User` 等持久化内部字段。

## 运势

### 今日运势

```http
POST /api/fortune
```

```json
{"chart_id":87,"query_date":"2026-06-16"}
```

`query_date` 必填，格式为 `YYYY-MM-DD`。响应包含 `score`、`lucky_color`、`wealth_direction`、`guide`、`yi/ji`、日课推算字段、格局字段等。

### 周运势

```http
POST /api/fortune/weekly
```

```json
{"chart_id":87,"start_date":"2026-06-16"}
```

响应：

```json
{"daily_fortunes":[],"weekly_score":70,"element_trend":"[]","summary":{}}
```

### 月运势

```http
POST /api/fortune/monthly
```

```json
{"chart_id":87,"year":2026,"month":6}
```

### 运势历史

```http
GET /api/fortune/history?chart_id=87&page=1&page_size=10
```

## 紫微斗数

### 紫微排盘

推荐使用命盘 ID，让后端读取八字命盘并缓存紫微结果：

```http
POST /api/ziwei/chart
```

```json
{"chart_id":87}
```

兼容临时排盘：

```json
{
  "birth_year": 2003,
  "birth_month": 4,
  "birth_day": 15,
  "birth_hour": 14,
  "birth_min": 0,
  "gender": "MALE",
  "algorithm": "default"
}
```

响应核心结构：

```json
{
  "palaces": [
    {
      "name": "命宫",
      "branch": "酉",
      "heavenly_stem": "辛",
      "is_body_palace": false,
      "stars": [{"name":"廉贞","type":"major","scope":"origin","brightness":"陷"}],
      "four_hua": ["破军化禄"],
      "adjective_stars": [],
      "sanfang_sizheng": {"opposite":"迁移","trine1":"财帛","trine2":"事业"}
    }
  ],
  "life_master": "文曲",
  "body_master": "天相",
  "five_bureau": "木三局",
  "body_palace": "亥",
  "patterns": ["廉贞破军同宫"]
}
```

### 紫微周期与解读

统一入口：

```http
POST /api/ziwei/period
```

通用请求字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `chart_id` | number | 必填，当前用户命盘 ID |
| `period_type` | string | 必填，见下表 |
| `year` | number | 流年/解释类可选，默认当前年 |
| `month` | number | 流月/流日解释可选，默认当前月 |
| `day` | number | 流日解释可选，默认当天 |
| `palace_idx` | number | `palace_reading` 必填，0-11 |
| `chart_id2` | number | `heming` 必填 |

| `period_type` | 响应字段 | 说明 |
|---------------|----------|------|
| `dayun` | `periods` | 大限列表 |
| `liunian` | `periods[0]` | 流年星曜分布 |
| `liuyue` | `periods[0]` | 流月星曜分布 |
| `liuri` | `periods[0]` | 流日星曜分布 |
| `sihua_feixing` | `periods` | 四化飞星 |
| `sihua_chain` | `chain` | 四化链式分析 |
| `self_mutagen` | `self_mutagens` | 自化检测 |
| `palace_reading` | `reading` | 单宫结构化详解 |
| `heming` | `heming` | 合盘 |
| `liunian_interpretation` | `periods[0]` | 流年文字解释 |
| `liuyue_interpretation` | `periods[0]` | 流月文字解释 |
| `liuri_interpretation` | `periods[0]` | 流日文字解释 |
| `period_summary` | `summary` | 流年/月/日汇总 |
| `liu_nian_stars` | `palaces` | 流年星曜列表 |

`palace_reading` 的新增结构化字段：

```json
{
  "reading": {
    "main_star_analysis": "...",
    "aux_star_influence": "...",
    "sihua_influence": "...",
    "sanfang_analysis": "...",
    "pattern_notes": "...",
    "summary": "...",
    "key_points": ["..."],
    "evidence": [{"type":"main_star","label":"主星","value":"廉贞(陷)","impact":"..."}],
    "sanfang_context": {"opposite":"迁移","trine1":"财帛","trine2":"事业","notes":[]},
    "pattern_details": [{"name":"廉贞破军同宫","palace":"命宫","stars":["廉贞","破军"],"basis":"廉贞、破军同在命宫","confidence":0.92}],
    "advice": ["..."],
    "risk_flags": ["..."],
    "confidence": 0.88
  }
}
```

### 流年叠盘

```http
POST /api/ziwei/overlay
```

```json
{"chart_id":87,"year":2026}
```

## 经典依据解读

```http
POST /api/interpretation/bazi
```

```json
{"chart_id":87,"focus":"overview"}
```

`focus` 可用值由前端约定为 `overview`、`pattern`、`tiaohou`、`ten_gods`。响应包含：

```json
{
  "status": "ok",
  "reason": "",
  "chart_id": 87,
  "focus": "overview",
  "summary": "...",
  "sections": [{"title":"...","content":"...","citations":[1]}],
  "citations": [{"id":1,"book":"...","quote":"...","score":0.91}]
}
```

## 反馈

### 创建反馈

```http
POST /api/feedback
```

```json
{
  "chart_id": 87,
  "target_type": "section",
  "target_id": "ming-palace",
  "rating": "accurate",
  "tags": ["紫微","命宫"],
  "comment": "这段比较准",
  "event_year": 2026,
  "event_category": "career",
  "consent_research": false,
  "consent_training": false
}
```

校验：

| 字段 | 规则 |
|------|------|
| `chart_id` | 必填，且必须属于当前用户 |
| `rating` | `accurate` / `inaccurate` / `too_generic` / `confusing` / `helpful` |
| `comment` | 最多 1000 个 Unicode 字符 |
| `tags` | 去重、去空、单个最多 32 字符、最多 10 个 |
| `consent_training` | 默认 false，需用户显式授权 |

响应：

```json
{"id":1,"status":"ok"}
```

### 反馈聚合

```http
GET /api/feedback/summary?chart_id=87
```

响应：

```json
{"chart_id":87,"total":3,"items":[{"rating":"accurate","count":2}]}
```

## 前后端对接约定

1. 前端页面和组件不直接 import `api/client.ts`；新增接口先放到 `vue/src/api/*.ts`。
2. `api/client.ts` 是唯一 Axios 实例，负责 `baseURL=/api`、JWT 注入、401 跳转。
3. 后端 handler 使用统一错误响应 helper，错误体必须包含 `error/code/message`。
4. 对外响应使用 DTO，不直接返回 GORM 模型；列表用 summary DTO，详情用 detail DTO。
5. 后端新增业务接口时，在 `src/cmd/main.go` 注册，并在对应 `src/internal/handler/*.go` 提供 handler。
6. JSON 字段统一使用 snake_case，前端 API 类型保持后端字段名；页面内部可按需映射为 camelCase 展示模型。
7. `chart_id` 相关接口必须由后端校验当前用户所有权，前端不能依赖本地状态做权限判断。
8. 新增接口需要同步更新本文件的端点矩阵和 `vue/src/api` 封装。

## 项目结构速查

更完整的结构说明见 `docs/frontend-backend-structure.md`。注意当前 `.gitignore` 忽略 `*.md`，该扩展文档若要提交需使用 `git add -f`。

后端目录：

| 目录 | 职责 |
|------|------|
| `src/cmd/` | 程序入口、依赖组装、路由注册 |
| `src/internal/handler/` | HTTP 参数解析、权限校验、JSON 响应 |
| `src/internal/model/` | DTO、GORM entity、常量 |
| `src/internal/service/` | 八字、运势、紫微、经典依据解读等业务规则 |
| `src/internal/store/` | 数据库读写封装 |
| `src/internal/middleware/` | JWT、ETag 等跨接口中间件 |
| `src/migrations/` | 数据库迁移 |
| `src/tools/` | 本地补算、导入、报告工具 |

后端调用链：

```text
cmd/main.go -> handler -> service/store -> model
```

前端目录：

| 目录 | 职责 |
|------|------|
| `vue/src/api/` | 所有后端接口封装和 DTO 类型 |
| `vue/src/stores/` | Pinia 全局状态 |
| `vue/src/router/` | 路由和登录守卫 |
| `vue/src/views/` | 页面容器，组织状态和业务流程 |
| `vue/src/components/` | 可复用 UI/业务展示组件 |
| `vue/src/components/fortune/` | 运势专用展示组件 |

前端 API 层现状：

| 文件 | 对应后端 |
|------|----------|
| `vue/src/api/auth.ts` | `/auth/*` |
| `vue/src/api/chart.ts` | `/chart`、`/charts`、`/fortune/history` |
| `vue/src/api/fortune.ts` | `/fortune`、`/fortune/weekly`、`/fortune/monthly` |
| `vue/src/api/ziwei.ts` | `/ziwei/chart`、`/ziwei/period`、`/ziwei/overlay` |
| `vue/src/api/interpretation.ts` | `/interpretation/bazi` |
| `vue/src/api/feedback.ts` | `/feedback`、`/feedback/summary` |
