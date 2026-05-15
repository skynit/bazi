# 八字算命 API 文档

**Base URL**: `http://localhost:8080`

---

## 认证说明

除 `/health`、注册、登录外，所有接口需 `Authorization: Bearer <token>` 请求头。

---

## 1. 健康检查

```
GET /health
```

**响应** `200`
```json
{"status":"ok"}
```

---

## 2. 用户认证

### 注册
```
POST /api/auth/register
Content-Type: application/json
```
```json
{"username":"test","email":"t@t.com","password":"test123"}
```
**响应** `201`
```json
{"token":"eyJ...","user":{"id":1,"username":"test","email":"t@t.com"}}
```

### 登录
```
POST /api/auth/login
Content-Type: application/json
```
```json
{"username":"test","password":"test123"}
```
**响应** `200`
```json
{"token":"eyJ..."}
```

### 当前用户
```
GET /api/auth/me
Authorization: Bearer <token>
```
**响应** `200`
```json
{"user":{"id":1,"username":"test","email":"t@t.com"}}
```

---

## 3. 八字排盘

### 创建命盘
```
POST /api/chart
Authorization: Bearer <token>
Content-Type: application/json
```
```json
{
  "birth_year": 1990,
  "birth_month": 6,
  "birth_day": 15,
  "birth_hour": 12,
  "birth_min": 0,
  "calendar_type": "SOLAR",
  "gender": "MALE",
  "name": "张三"
}
```
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| birth_year | int | ✅ | 出生年 |
| birth_month | int | ✅ | 出生月 |
| birth_day | int | ✅ | 出生日 |
| birth_hour | int | ✅ | 出生时（0-23） |
| birth_min | int | | 出生分（默认0） |
| calendar_type | string | ✅ | SOLAR/LUNAR/BAZI |
| gender | string | ✅ | MALE/FEMALE |
| name | string | | 姓名 |

**响应** `200`
```json
{
  "id": 1,
  "year_pillar":  {"gan":"庚","zhi":"午"},
  "month_pillar": {"gan":"壬","zhi":"午"},
  "day_pillar":   {"gan":"庚","zhi":"寅"},
  "hour_pillar":  {"gan":"壬","zhi":"午"},
  "five_elements": {"金":15,"木":4,"水":8,"火":12,"土":5},
  "na_yin": {
    "year": "路旁土",
    "month": "杨柳木",
    "day": "松柏木",
    "hour": "杨柳木"
  }
}
```

### 命盘列表
```
GET /api/charts?page=1&page_size=10
Authorization: Bearer <token>
```
**响应** `200`
```json
{"charts":[...], "total":1, "page":1, "page_size":10}
```

### 命盘详情
```
GET /api/charts/:id
Authorization: Bearer <token>
```
**响应** `200` — 同创建命盘返回结构

---

## 4. 运势查询

### 今日运势
```
POST /api/fortune
Authorization: Bearer <token>
Content-Type: application/json
```
```json
{"chart_id": 1}
```
| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| chart_id | int | ✅ | 命盘ID |
| query_date | string | | 查询日期（默认今天，如"2026-05-12"） |

**响应** `200`
```json
{
  "score": 65,
  "lucky_color": "红色系",
  "lucky_number": 7,
  "wealth_direction": "正南",
  "clash_zodiac": "鼠",
  "auspicious_hours": ["巳时","午时"],
  "yi": [{"activity":"嫁娶","reason":"日干相合"}],
  "ji": [{"activity":"动土","reason":"日支相冲"}],
  "element_images": [
    {"element":"金","image_url":"/images/elements/metal.svg","description":"五行金"}
  ]
}
```

### 周运势
```
POST /api/fortune/weekly
Authorization: Bearer <token>
```
```json
{"chart_id": 1, "start_date": "2026-05-12"}
```

### 月运势
```
POST /api/fortune/monthly
Authorization: Bearer <token>
```
```json
{"chart_id": 1, "year": 2026, "month": 5}
```

---

## 5. AI 分析（占位）

```
POST /api/fortune/ai
Authorization: Bearer <token>
```
```json
{}
```
**响应** `200`
```json
{"status":"coming_soon","message":"AI分析功能即将上线"}
```

---

## 6. 紫微斗数

### 排盘
```
POST /api/ziwei/chart
Authorization: Bearer <token>
```
```json
{"birth_year":1990,"birth_month":6,"birth_day":15,"birth_hour":12,"birth_min":0,"calendar_type":"SOLAR","gender":"MALE","name":""}
```
**响应** `200`
```json
{
  "palaces": [{"name":"命宫","main_stars":["紫微"],"aux_stars":["左辅"],"brightness":{"紫微":"庙"},"four_hua":["化禄"]}, ...],
  "body_palace": "夫妻宫",
  "life_master": "禄存",
  "body_master": "天相",
  "five_bureau": "金四局",
  "patterns": ["紫府同宫格"]
}
```

### 大限/流年/流月/流日
```
POST /api/ziwei/period
Authorization: Bearer <token>
```
```json
{"chart_id": 1, "period_type": "dayun"}
```
| period_type | 说明 |
|-------------|------|
| dayun | 大限（十年） |
| liunian | 流年 |
| liuyue | 流月 |
| liuri | 流日 |

### 流年叠盘
```
POST /api/ziwei/overlay
Authorization: Bearer <token>
```
```json
{"chart_id": 1, "year": 2026}
```

---

## 7. 查询历史

```
GET /api/fortune/history?chart_id=1
Authorization: Bearer <token>
```

---

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或token过期 |
| 500 | 服务器内部错误 |

**错误响应**
```json
{"error":"错误描述"}
```
