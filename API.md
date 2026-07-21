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
| POST | `/api/chart/preview` | 是 | `vue/src/api/chart.ts` | `handler.ChartHandler.Preview` |
| POST | `/api/chart` | 是 | `vue/src/api/chart.ts` | `handler.ChartHandler.Chart` |
| GET | `/api/charts` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.ListCharts` |
| GET | `/api/charts/:id` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.GetChart` |
| POST | `/api/fortune` | 是 | `vue/src/api/fortune.ts` | `handler.FortuneHandler.CalculateDaily` |
| POST | `/api/fortune/weekly` | 是 | `vue/src/api/fortune.ts` | `handler.WeeklyFortuneHandler.Weekly` |
| POST | `/api/fortune/monthly` | 是 | `vue/src/api/fortune.ts` | `handler.MonthlyFortuneHandler.HandleMonthly` |
| POST | `/api/fortune/ai` | 是 | 暂无页面调用 | `handler.AIStubHandler.AnalyzeFortune` |
| GET | `/api/fortune/history` | 是 | `vue/src/api/chart.ts` | `handler.HistoryHandler.FortuneHistoryList` |
| GET | `/api/buyi/today` | 是 | `vue/src/api/buyi.ts` | `handler.BuyiHandler.Today` |
| POST | `/api/buyi/today` | 是 | `vue/src/api/buyi.ts` | `handler.BuyiHandler.DrawToday` |
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
{"username":"test","email":"t@example.com","password":"test12345"}
```

用户名须为 3–32 个 Unicode 字母、数字或 `_-.`，密码至少 8 个字符，邮箱须为有效地址。

响应：

```json
{"token":"eyJ...","user":{"id":1,"username":"test","email":"t@example.com"}}
```

### 登录

```http
POST /api/auth/login
```

```json
{"username":"test","password":"test12345"}
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

### 预览并校验命盘

```http
POST /api/chart/preview
```

预览接口负责标准化出生输入并完成八字排盘，但**不会保存命盘**。前端应先展示 `birth_validation` 与八字四柱，让用户确认后再使用相同请求调用 `POST /api/chart`。校验内容只包含历法与八字信息，不包含紫微基础盘。

```json
{
  "birth_year": 2024,
  "birth_month": 1,
  "birth_day": 1,
  "birth_hour": 8,
  "birth_min": 30,
  "birth_sec": 15,
  "calendar_type": "LUNAR",
  "lunar_leap_month": false,
  "gender": "MALE",
  "zi_hour_policy": "late_zi_next_day",
  "name": "",
  "birth_place": "上海",
  "timezone": "Asia/Shanghai",
  "birth_utc_offset_seconds": 28800,
  "longitude": 121.4737,
  "use_true_solar_time": true,
  "time_uncertain": true,
  "uncertainty_seconds": 900
}
```

请求字段：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `birth_year` | int | 是 | 输入历法下的出生年 |
| `birth_month` | int | 是 | 输入历法下的出生月 |
| `birth_day` | int | 是 | 输入历法下的出生日 |
| `birth_hour` | int | 是 | 0–23 小时 |
| `birth_min` | int | 否 | 0–59 分钟，默认 0 |
| `birth_sec` | int | 否 | 0–59 秒，默认 0；会参与节令交接与起运边界计算 |
| `calendar_type` | string | 否 | `SOLAR` / `LUNAR`，空值按 `SOLAR` |
| `lunar_leap_month` | bool | 否 | 农历输入是否为闰月；公历输入应为 false |
| `gender` | string | 是 | `MALE` / `FEMALE` |
| `zi_hour_policy` | string | 否 | `late_zi_next_day`（默认，23:00 后日柱算次日）或 `late_zi_same_day`（23:00–23:59 日柱仍算当天） |
| `name` | string | 否 | 空值自动生成 `YYYY-MM-DD 命盘` |
| `birth_place` | string | 否 | 展示与审计用出生地 |
| `timezone` | string | 否 | IANA 时区，默认 `Asia/Shanghai` |
| `birth_utc_offset_seconds` | int | 条件必填 | 本地钟表时间在夏令时回拨时重复出现，必须从错误信息列出的 UTC 偏移秒中选择一个；普通时刻可省略 |
| `longitude` | number | 条件必填 | 经度，启用真太阳时时必填，范围 -180–180 |
| `use_true_solar_time` | bool | 否 | 是否按经度与均时差修正计算时间 |
| `time_uncertain` | bool | 否 | 是否启用时间不确定区间；必须与 `uncertainty_seconds > 0` 保持一致 |
| `uncertainty_seconds` | int | 条件必填 | 中心时刻前后对称误差，1–86400 秒；`time_uncertain=true` 时必填 |
| `candidate_id` | string | 创建时条件必填 | 误差区间跨越四柱边界时，提交预览响应中的候选 ID；预览请求不需要 |

预览响应的 `id` 为 `0`，并返回完整排盘结果及以下校验信息：

```json
{
  "id": 0,
  "name": "2024-01-01 命盘",
  "calendar_type": "LUNAR",
  "engine_version": "bazi-engine-2026-07-15.30",
  "zi_hour_policy": "late_zi_next_day",
  "rule_version": "...",
  "birth_validation": {
    "normalization_version": "birth-normalization-2026-07-15.2",
    "input_calendar": "LUNAR",
    "original_date_time": "2024-01-01 08:30:15",
    "converted_solar_date_time": "2024-02-10 08:30:15",
    "calculation_date_time": "2024-02-10 08:21:58",
    "lunar_date": "...",
    "current_solar_term": "...",
    "current_solar_term_started_at": "...",
    "birth_place": "上海",
    "timezone": "Asia/Shanghai",
    "utc_date_time": "2024-02-10T00:30:15Z",
    "local_time_ambiguous": false,
    "possible_utc_offset_seconds": [28800],
    "longitude": 121.4737,
    "true_solar_time_applied": true,
    "true_solar_adjustment_minutes": -8,
    "timezone_offset_seconds": 28800,
    "mean_solar_adjustment_seconds": 354,
    "equation_of_time_seconds": -851,
    "true_solar_adjustment_seconds": -497,
    "true_solar_algorithm": "usno-approx-solar-coordinates-j2000",
    "true_solar_source": "USNO Astronomical Applications Department: Approximate Solar Coordinates",
    "true_solar_within_validated_range": true,
    "true_solar_uncertainty_seconds": 6,
    "time_uncertain": true,
    "zi_hour_policy": "late_zi_next_day",
    "uncertainty_seconds": 900,
    "notices": []
  },
  "uncertainty": {
    "seconds": 900,
    "algorithm_uncertainty_seconds": 6,
    "effective_seconds": 906,
    "input_range_start": "2024-02-10 08:15:15 +08:00",
    "input_range_end": "2024-02-10 08:45:15 +08:00",
    "evaluation_range_start": "2024-02-10 08:15:09 +08:00",
    "evaluation_range_end": "2024-02-10 08:45:21 +08:00",
    "calculation_range_start": "2024-02-10 08:06:52",
    "calculation_range_end": "2024-02-10 08:37:04",
    "crossed_boundaries": []
  },
  "candidate_charts": [
    {
      "candidate_id": "32-hex-character-id",
      "calculation_range_start": "2024-02-10 08:06:52",
      "calculation_range_end": "2024-02-10 08:37:04",
      "representative_time": "2024-02-10 08:21:58",
      "year_pillar": {"gan":"甲","zhi":"辰"},
      "month_pillar": {"gan":"丙","zhi":"寅"},
      "day_pillar": {"gan":"...","zhi":"..."},
      "hour_pillar": {"gan":"...","zhi":"..."},
      "da_yun_start_at_min": "...",
      "da_yun_start_at_max": "..."
    }
  ],
  "stable_fields": ["year_pillar", "month_pillar", "day_pillar", "hour_pillar"],
  "unstable_fields": [],
  "requires_candidate_selection": false,
  "year_pillar": {"gan":"甲","zhi":"辰"},
  "month_pillar": {"gan":"丙","zhi":"寅"},
  "day_pillar": {"gan":"...","zhi":"..."},
  "hour_pillar": {"gan":"...","zhi":"..."}
}
```

真太阳时总修正为 `mean_solar_adjustment_seconds + equation_of_time_seconds`。前者使用出生瞬间的真实时区偏移（包括夏令时），后者使用 USNO J2000 近似太阳坐标公式；秒级结果直接进入四柱、节令边界与起运计算。USNO 对该近似公式公开说明的适用窗口为 1800–2200 年。窗口内使用 `6` 秒候选扩展工程量：约 1 角分太阳坐标精度折算 4 秒，另计 UTC/UT1 差异 1 秒和秒级取整 1 秒；这不是严格天文误差上界，经度输入误差另计。窗口外返回 `0` 表示误差未知，并在 `notices` 中警告。

不确定区间的端点会分别执行完整历法、时区和真太阳时标准化，不复用中心时刻的时区偏移或均时差。后端按“四柱完全相同的最大连续秒段”合并候选；`da_yun_start_at_min/max` 表示该候选段内的起运范围，不会因起运时刻连续变化而逐秒生成候选。若启用真太阳时且处于算法适用窗口，名义算法误差会加入 `effective_seconds` 并参与边界评估。

IANA 时区中的夏令时跳跃时刻会被拒绝；夏令时回拨造成的重复本地时刻不会再静默选择其中一次。此时接口返回可用偏移秒，例如纽约 `2024-11-03 01:30:00` 需要显式提交 `-14400` 或 `-18000`，所选偏移会持久化并参与候选 ID 校验。

`zi_hour_policy` 是四柱计算合同的一部分。`late_zi_next_day` 使用 `tyme4go DefaultEightCharProvider`，23:00 起日柱算次日；`late_zi_same_day` 使用 `LunarSect2EightCharProvider`，23:00–23:59 日柱仍算当天，但子时时柱保持该 Provider 的原始算法。实现按单次计算选择 Provider，不修改进程全局变量，因此并发请求不会互相改变口径。候选边界分别标记为 `zi_hour_day_boundary`（子初换日）或 `civil_day`（午夜换日）。

### 确认并创建命盘

```http
POST /api/chart
```

请求体以预览接口为基础。若 `requires_candidate_selection=true`，必须增加预览返回的 `candidate_id`；缺失时返回 `409 CONFLICT`，不存在或与当前出生输入不匹配时返回 `400 INVALID_REQUEST`，且两种情况都不会持久化。后端会重新计算候选并校验 ID，再保存采用的标准化出生输入、误差秒数、候选 ID、完整八字结果快照、`engine_version` 与规则版本，不信任客户端提交的候选内容。即使用户时间标为精确，真太阳时名义误差在极端边界上产生多个候选时也执行同一门禁。

响应除 `id`、`birth_validation`、`engine_version`、`rule_version`、`school`、`rule_meta` 外，还包含四柱、五行、身强、十神、纳音、大运等完整八字字段。创建响应的大运字段为 `da_yun`。

`pattern_analysis` 只返回本地检测器的可审计候选证据，不返回扁平的唯一格局、描述、喜忌五行或现实解释：

```json
{
  "pattern_analysis": {
    "rule_id": "bazi.pattern-candidate-set-v34",
    "schema_version": "pattern-candidates-2026-07-17.27",
    "detector_profile": "classical_structural_detectors_v45",
    "detector_count": 10,
    "detector_manifest_sha256": "6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8",
    "detector_profiles": [
      {"rule_id": "pattern.aux.jinshen", "algorithm_sha256": "326f3feb4586cd0bf618737a597807f1da18b0ce70dda666c6589203993597e6", "behavior_sha256": "be71d3d52d03a913a3e2f8010c92480c2da401397c68d82fcdd4492881b2fe09", "profile_sha256": "c09c584d0ae7e0b55e6167cae417f6f85c9f1a514978ff0640d9dde9c128f63e"},
      {"rule_id": "pattern.aux.kuigang", "algorithm_sha256": "552b8d2dd8cef0d7f5343638c4c07d870791a34283d50d6e7247e514334b4db6", "behavior_sha256": "40f31c0dd264242bb03b8e13a1befab913faeddb75e40a4a11154e5c72f8bfe9", "profile_sha256": "5473b6110b3cb1bfbc2570f000c6e84da7d1ad5a6738ef1f1ca462d4c40d8e02"},
      {"rule_id": "pattern.aux.ride", "algorithm_sha256": "a5d3c9dc5ca9dcc125898be9a7305c8f1b88da9dc31f32f39e250d9c092276b2", "behavior_sha256": "7be77b6771f51945042d3fe7ddb0a305c8840147f89c97b84754d2dc67c65a1c", "profile_sha256": "d6624103cfe639446158c5dfa6909f4979dc1b995a1da1b245acf1621431ab02"},
      {"rule_id": "pattern.aux.sanqi", "algorithm_sha256": "69f4d63d045612b192a7276e3951e04104eb6b4eb8431aea62f6b7a712eca25c", "behavior_sha256": "7a9e9fb9f8a63f9bfd908c4b27187c8329e132cb63d410d371adc9b51de2348a", "profile_sha256": "b54db26c03be56842109528860f258280937d5dafbaf5b5c5dd61cfbd6956028"},
      {"rule_id": "pattern.lu.jianlu", "algorithm_sha256": "a8ba8785978eecf10745ffbb81d86542f3a948ded86d569ca8ffc8ad565c603e", "behavior_sha256": "d51707e2f5627ade0d053f10c9885eb7406e0b55defb6ea3b7ed6ad25c91a97b", "profile_sha256": "29314b503c9d6b5dcd248121d9abb189ae5304e2984648299be0580bb04b5322"},
      {"rule_id": "pattern.lu.riren", "algorithm_sha256": "0f124e093fbc6009e6fadb279ae9020205493dd5e4b2e2d07a4fc8bfa5a8799d", "behavior_sha256": "d3258cb7609e23df271a827467f048799339f31623a47949de1c55e1930cbc1b", "profile_sha256": "bb3a422fa07b0661abac8c7925aaa039262f4a1a6e8cde0b1c6d5d20aa847983"},
      {"rule_id": "pattern.lu.yueren", "algorithm_sha256": "3955ef0ad0405984cbc41bb2a58571f9409b1a75a51ac00be88f1278ffa6fd28", "behavior_sha256": "cb4d3b6db008d758dbefa4730bf876c8ec6edf82a116733ed3bcd8131b0a6df3", "profile_sha256": "80c34b3028f8a9cb8b8dab004ee89fc9355ed2b8feb5b451c1113fe41d58f97e"},
      {"rule_id": "pattern.lu.zhuanlu", "algorithm_sha256": "4ac683a19f228d45b2acd2bdbe22b0c0ecd22d161f824295f6d1ca712a302067", "behavior_sha256": "3e5fa70c2fb6805ecb5e4c0d57318242840866b2ee3b3bb7057fd6521deb4581", "profile_sha256": "94f7305566788f872cbcceadd7aa62e15ab41d42421869026c509798ae256b2b"},
      {"rule_id": "pattern.special.liangqi", "algorithm_sha256": "a4feed92bef94cf35624c525895a9872e1014f59aa273dd1c69cace155fc9c23", "behavior_sha256": "454a7144a3ec3f6fd6a5cf2a7e2f2af5ec668fcf02c513f5ecd833a09763b55e", "profile_sha256": "20ed815f5411b2add63e4d98931f8f775b7af1b4f98a2b41f91a6b9feae9c610"},
      {"rule_id": "pattern.special.zhuanwang", "algorithm_sha256": "ef1fe91027ba0e0101789f43e2bef18f3f17fd273a4913ca88d980deee66e53f", "behavior_sha256": "6193bc4d839aead50015b4ce6f155151e528437eb9fd269b7d49cb3657c235e7", "profile_sha256": "daedf9a821d62349eadcaf07757cc678f86ff67f3adb76e966d3d8d3685cd57c"}
    ],
    "detector_profile_change_contract": {
      "scheme": "layered_detector_digest_delta_v1",
      "alignment_key": "rule_id",
      "classes": [
        "detector_added",
        "detector_removed",
        "algorithm_digest_changed",
        "behavior_evidence_digest_changed",
        "semantic_profile_digest_changed",
        "layered_digests_unchanged"
      ],
      "behavior_evidence_scope": "simple_full_truth_table_complex_partial_contract",
      "inference_boundary": "digest_evidence_only"
    },
    "detector_profile_migration": {
      "ledger_id": "bazi.pattern-detector-profile-migrations",
      "schema": "pattern_detector_profile_migration_ledger_v2",
      "sha256": "a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825",
      "migration_count": 4,
      "latest_migration_id": "bazi.pattern-candidate-set-v33_to_v34",
      "latest_from_snapshot_id": "bazi.pattern-candidate-set-v33",
      "latest_to_snapshot_id": "bazi.pattern-candidate-set-v34",
      "change_scheme": "layered_detector_digest_delta_v1",
      "claim_boundary": "digest_evidence_only",
      "chain_scheme": "pattern_detector_profile_migration_chain_v1",
      "chain_head_sha256": "07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd"
    },
    "detector_profile_release_anchor": {
      "schema": "pattern_detector_profile_release_anchor_v1",
      "anchor_id": "bazi.pattern-detector-profile-release-anchor-v34",
      "artifact_path": "release/pattern-detector-profile-anchor.json",
      "sha256": "ebd6323f28715695aa3c4ee9038e74d261c9fa34b422037266c4b097e3086a2e",
      "verification_profile": "repository_ci_cross_check_v1",
      "trust_boundary": "unsigned_repository_ci_artifact",
      "claim_boundary": "digest_evidence_only"
    },
    "inputs": {
      "pillars": ["丙子", "丙寅", "甲寅", "戊辰"],
      "month_branch": "寅"
    },
    "candidates": [
      {
        "rule_id": "pattern.lu.jianlu",
        "pattern_name": "建禄格",
        "category": "结构格局",
        "source": "《三命通会》PDF第230-232页月令建禄十干表及取用条件"
      },
      {
        "rule_id": "pattern.lu.zhuanlu",
        "pattern_name": "专禄格",
        "category": "结构格局",
        "source": "《三命通会》PDF第190页甲寅乙卯庚申辛酉专禄结构"
      }
    ],
    "status": "observed",
    "validation_status": "not_validated",
    "interpretation_status": "not_adjudicated",
    "limitations": [
      "detector conditions are local classical-text Profiles without expert Gold adjudication",
      "candidate list order is deterministic serialization only and does not rank or adjudicate patterns",
      "candidates do not determine favorable elements or real-world outcomes"
    ]
  }
}
```

当前10个检测器由单一 `patternDetectorRegistry` 声明并统一执行。每次分析创建一份独立注册表快照，规则 ID、来源、category、调用函数和由其长度派生的 `detector_count` 都绑定该次局部快照；修改一份清单不会污染同进程的后续分析。每条注册规则从实际运行的封闭表、柱位范围、算法参数及规范化 Go AST 调用闭包计算独立 `profile_sha256`；构建期契约会重算 `go_ast_detector_closure_v1` 指纹，检测根函数或其同包 helper 的逻辑变化必须改变摘要，运行时不读取源码。专旺与两气 Profile 还包含基线成立盘及单原子输入变异的 `behavior_witnesses`，用于证明结构缺失、克神、未知干支、柱数、第三气和非四四均分等拒绝条件可重复触发。专旺以 `metamorphic_policies` 约束地支排列、重复/缺失必需支及全部结构外地支；两气则约束十种无序五行对、八个位点四选四、第三气和5:3非均分的对称关系。其余8个简单检测器以 `canonical_truth_table_v1 behavior_manifest` 绑定有限输入域、案例数、命中数和完整行为清单摘要。`detector_profiles` 按规则 ID 规范排序公开10条逐规则摘要，合法与非法结果均从同一次执行快照生成。`detector_manifest_sha256` 对规则身份、来源、category、算法摘要、允许输出名称与 Profile 摘要计算总 SHA-256。专旺结构表每次返回独立嵌套值，十干禄位由纯 `canonicalLuProfile/luBranchForStem` 定义供格局、神煞与身强共同消费，不保留可跨调用修改的 map 或启动时副本。命中项按规则 ID 稳定排列，数组顺序只用于确定性序列化，不表示主格、兼格、优先级或成立裁决。历史快照按权威输入重算并逐字段校验候选、逐规则摘要及总摘要合同。已失败关闭的八格、复合格局、从格及化气快捷检测器不会继续发布候选；辅助特征通过 `category` 与结构格局区分。现存候选尚未取得专家 Gold，因此不得决定喜忌、性格、事业、财富、婚姻、行运或其他现实结果。

`detector_profiles` 每项只公开 `rule_id/algorithm_sha256/behavior_sha256/profile_sha256`。比较两个结果时：`algorithm_sha256` 不同表示检测根函数或同包调用闭包改变；算法相同而 `behavior_sha256` 不同表示简单规则的完整有限域输出或复杂规则的见证/形变行为合同改变；前两者相同而 `profile_sha256` 不同表示来源、分类、输出名称、共享上下文或其他语义 Profile 改变。专旺与两气的行为摘要并非全四柱穷举，不能与八个简单检测器的完整真值表等量解释；三层摘要都不证明传统口径正确或现实预测有效。

`detector_profile_change_contract` 固定采用 `layered_detector_digest_delta_v1`。比较两份快照时先验证每项规则 ID 非空且唯一、三条摘要均为64位小写十六进制，再按 `rule_id` 对齐：只在一侧出现分别标记 `detector_added` 或 `detector_removed`；两侧都有时逐层枚举 `algorithm_digest_changed`、`behavior_evidence_digest_changed`、`semantic_profile_digest_changed`，三层都相同时才标记 `layered_digests_unchanged`。输入无效必须拒绝分类，不得猜测。`layered_digests_unchanged` 只表示当前三条工程证据摘要相同；尤其复杂规则的行为证据只是见证与形变合同，不能据此宣称完整四柱行为、传统裁决或现实预测等价。

`detector_profile_migration` 引用仓库内嵌的 `rules/pattern_detector_profile_migrations.json` 账本。`pattern_detector_profile_migration_ledger_v2` 将重复的逐规则摘要收敛为 `profile_sets`，再由版本快照引用；迁移项连接前后快照和预期变化集。加载时会用 `layered_detector_digest_delta_v1` 重算每条迁移，并按 `pattern_detector_profile_migration_chain_v1` 重算内容摘要链：首项 `previous_migration_sha256` 固定为64个零，其余项必须引用前项 `migration_sha256`；单项摘要绑定前后完整快照、解析后的逐规则摘要和预期分类，不只散列引用ID。链尾仍须与当前运行时完全一致。响应发布规范化账本摘要和链头，历史完整快照不在每次响应重复传输。该链可以区分保留旧摘要的正常追加与迫使后继摘要变化的旧记录重写，但仍只证明工程版本证据一致，不证明规则传统正确或现实预测有效。

`detector_profile_release_anchor` 引用根目录 `release/pattern-detector-profile-anchor.json`。该文件不嵌入生产二进制，固定当前五项版本、总manifest、逐规则摘要集摘要、账本ID/schema/摘要、迁移数和链头；GitHub Actions只运行 `internal/service/bazi` 的锚合同，将文件与当前运行时证据交叉验证后作为独立CI artifact上传。响应发布锚文件规范摘要、路径和验证Profile。`trust_boundary=unsigned_repository_ci_artifact` 表示它仍与源码处于同一仓库且没有签名tag、透明日志或外部时间戳，不能证明整个仓库未被共同重写，也不能作为传统准确性或现实预测验证。

检测器内部只保留候选名称和类型，不保存隐藏的描述、子类型或喜忌五行；所有可审计依据均通过候选 `source`、规则元数据和版本化合同公开。内部字段不能作为绕过当前 API 边界恢复扁平格局解释的来源。

格局候选以非空且唯一的 `rule_id` 作为唯一身份，不再重复返回与其相同的 `candidate_id`。这里的规则命中不得与出生时间不确定性预览中的哈希 `candidate_id` 混淆；后者仍用于选择跨四柱边界的出生候选。

`validation_status` 与 `interpretation_status` 只在 `pattern_analysis` 集合级返回，表示当前整组检测器尚未通过专家 Gold 验证、现实解释未裁决。单个候选不再重复这两项固定状态，避免集合与候选行在历史快照中形成矛盾。

候选进入 `candidates` 数组本身即表示对应检测器条件命中，不再额外返回固定值 `basis: local_detector_conditions_matched`。每项 `rule_id` 与 `source` 继续提供稳定身份和可审计来源；其他领域具有实际多种取值的 `basis` 字段不受影响。

每个候选只用 `category` 区分 `结构格局` 与 `辅助特征`，不再同时返回未经专家裁决的 `pattern_type`。旧双分类会让魁罡、日德一面属于辅助特征、一面又标为特殊格局；删除后数组不再包含相互冲突的分类结论。

`month_command_evidence` 只在月支某一藏干本字透于年干、月干或时干时记录。例如申中庚金必须见庚透干；只见同属金但阴阳不同的辛，不再冒充“庚透”并生成偏印格候选。同五行异干仍保留在十神与五行事实层。

规则 ID 前缀与 category 同步：辅助特征使用 `pattern.aux.*`，结构候选使用 `pattern.lu.*` 或 `pattern.special.*`。金神辅助候选的正式 ID 为 `pattern.aux.jinshen`，旧 `pattern.special.jinshen` 不再发布。

`body_strength` 当前只表示 `local_fuyi_weighted_score_v3` 的本地加权评分证据。旧 `verdict`、`like`、`dislike`、`summary` 字段已删除；五档标签改为 `score_band_candidate`，明确由固定阈值和后验修正产生，且 `is_strength_conclusion` 恒为 `false`：

```json
{
  "body_strength": {
    "rule_id": "bazi.body-strength-score-candidate-v3",
    "schema_version": "body-strength-evidence-2026-07-18.3",
    "rule_version": "bazi-rules-2026-07-17.27",
    "school": "子平八字-扶抑调候-v2",
    "scoring_profile": "local_fuyi_weighted_score_v3",
    "yue_ling_rule_id": "bazi.body-strength.yue-ling-seasonal-state.v1",
    "yue_ling_profile": "local_ziping_yue_ling_5x12_v1",
    "yue_ling_table_sha256": "sha256:76d21a2761256db42976144f236056b080a32e367bc50cbb9b927eb04d1c26a6",
    "inputs": {
      "pillars": ["戊辰", "丙巳", "甲子", "壬申"],
      "day_stem": "甲",
      "day_element": "木",
      "month_branch": "巳"
    },
    "score_band_candidate": "偏旺",
    "band_selection_basis": "ordered_fixed_local_thresholds_then_posterior_adjustments",
    "band_rules": [
      {"candidate":"身旺","operator":"gt","threshold":0.7},
      {"candidate":"偏旺","operator":"gt","threshold":0.5},
      {"candidate":"中和","operator":"gt","threshold":0.4},
      {"candidate":"偏弱","operator":"gt","threshold":0.3},
      {"candidate":"身弱","operator":"otherwise"}
    ],
    "total_score": 0.62,
    "ling_score": 1,
    "di_score": 2.4,
    "shi_score": 0.8,
    "sheng_score": 1,
    "lu_bonus": 0,
    "components": [
      {
        "rule_id": "bazi.body-strength.yue-ling-seasonal-state.v1",
        "key": "ling",
        "name": "得令",
        "raw_score": 1,
        "normalized_score": 0.3333,
        "weight": 0.4,
        "weighted_score": 0.1333,
        "basis": "local_weighted_component_profile",
        "status": "observed",
        "validation_status": "not_validated",
        "description": "月令分项的本地评分说明"
      }
    ],
    "evidence": [
      {
        "rule_id": "bazi.body-strength.yue-ling-seasonal-state.v1.lookup",
        "component": "ling",
        "polarity": "support",
        "source": "月令",
        "item": "巳",
        "score": 1,
        "basis": "local_component_scoring_rule",
        "status": "observed",
        "interpretation_status": "not_adjudicated",
        "reason": "木日主遇巳月，按本地季令表记录原始分1.00。"
      }
    ],
    "adjustments": [],
    "status": "observed",
    "validation_status": "not_validated",
    "interpretation_status": "not_adjudicated",
    "is_strength_conclusion": false,
    "limitations": [
      "component weights and normalizers are local profile parameters without Gold calibration",
      "score-band thresholds and posterior adjustments are not learned from Train Gold",
      "the score-band candidate does not determine favorable elements or real-world outcomes"
    ]
  }
}
```

实际响应的 `components/evidence/adjustments` 会完整返回规则 ID、原始分、归一化分、权重、加权分、取值依据与验证状态，上例数组只展示一条代表记录。历史快照必须能由四柱重算出相同的输入、连续分数、组件、逐条证据、修正和分段候选，否则会使用当前规则重算。该结构提升可复现性，不证明权重、阈值或分段标签具有传统正确性或预测效度。

从 `bazi-rules-2026-07-15.23` 起，`rule_meta.body_strength.yue_ling` 公开日主五行顺序、十二月支顺序、完整 5x12 分值表、旺相休囚死分值语义及规范 SHA-256。哈希输入是 `day_element_order`、`month_branch_order`、`scores` 三字段按该顺序组成的 UTF-8、无空白紧凑 JSON，数值使用最短十进制表示；当前四土月采用 `simplified_earth_prosperous_in_chou_chen_wei_xu` Profile，验证状态为 `engineering_complete_expert_gold_pending`。这些字段证明响应使用了哪张表，不表示该表已经通过专家 Gold 或预测效度验证。

从 `bazi-rules-2026-07-15.26` 起，`score_band_candidate` 必须严格等于最终 `total_score` 按公开 `band_rules` 计算的结果。`得令不旺` 和 `失令不衰` 只修改连续分，不能再无条件手工升降一级；两者分别使用 `bazi.body-strength.adjustment.de-ling-bu-wang.v1` 与 `bazi.body-strength.adjustment.shi-ling-bu-shuai.v1`。触发修正并不保证跨过分段阈值。

从 `bazi-rules-2026-07-15.27` 起，`rule_meta.body_strength.root` 公开藏干本中余气权重、十二长生阶段权重、通根基数、透干乘数和范围；当前透干范围明确为 `all_four_heaven_stems_including_day_master`。`rule_meta.body_strength.bonus` 公开十干禄表、五阳干刃表、`no_yang_ren_bonus` 阴干口径、四项加分和表 SHA-256。运行路径直接消费这些字段，通根与禄刃证据分别绑定 `bazi.body-strength.root-evidence.v1` 和 `bazi.body-strength.lu-yang-ren-bonus.v1`，但这些数值仍未通过专家 Gold 校准。

`rule_meta.body_strength.influence` 公开年月时可见天干范围、比肩/劫财、官杀、食伤、财的本地系数，以及四支藏干 `all_four_branches_restrict_only` 范围和乘数。印星明确归 `sheng`，同气根明确归 `di`；天干和四支藏干共用 `bazi.body-strength.influence-evidence.v2` 评分规则。此前只统计日支藏干的财官食伤，会遗漏年、月、时支的克泄耗，现已统一覆盖四支。

从 `bazi-rules-2026-07-15.29` 起，得势使用 `centered_logistic_v1`，有正负方向的原始分 `0` 映射为中性 `0.5`；只可能非负的得生使用 `zero_origin_logistic_v1`，原始分 `0` 必须映射为 `0`，不再无条件贡献一半组件权重。`rule_meta.body_strength.adjustment_force` 同时公开后验天干力量、藏干乘数、比例公式、支持范围和中性目标。

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
      "name": "2003-04-15 命盘",
      "gender": "MALE",
      "zi_hour_policy": "late_zi_next_day",
      "birth_year": 2003,
      "birth_month": 4,
      "birth_day": 15,
      "birth_hour": 14,
      "birth_min": 30,
      "birth_sec": 0,
      "calendar_type": "SOLAR",
      "lunar_leap_month": false,
      "birth_place": "上海",
      "timezone": "Asia/Shanghai",
      "birth_utc_offset_seconds": 28800,
      "longitude": 121.4737,
      "use_true_solar_time": true,
      "time_uncertain": false,
      "uncertainty_seconds": 0,
      "selected_candidate_id": "...",
      "engine_version": "bazi-engine-2026-07-15.30",
      "stored_rule_version": "...",
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

响应为 `ChartDetailResponse` DTO，不直接返回 `BirthChart` 存储模型。除列表字段外，详情包含 `birth_validation`、八字完整展示字段、当前输出使用的 `rule_version`、持久化版本 `stored_rule_version`、`da_yun_start` 及其别名 `da_yun`、`ziwei_result`、`ziwei_computed`。

历史详情优先读取创建时保存的完整八字快照，确保同一命盘可追溯；若快照缺失，再按保存的标准化出生输入重算，最后才回退原始出生字段。详情不会返回 `user_id`、`DeletedAt`、关联 `User` 等持久化内部字段。

`tong_guan` 已退役。旧实现用本地五行计分中的“双方至少 5 分且合计至少占一半”构造通关候选，该门槛没有古籍封闭条件、专家 Gold 或外部实现验证，不能据此进入用神综合。

`tiaohou` 当前返回 `selection_basis: "first_table_entry_candidate"`、`depth_affects_selection: false` 和 `depth_evidence`。日期输入以出生时刻所在的前后节令交接时刻计算 `elapsed_seconds`、`interval_seconds`、`position` 及前/中/后段；四柱输入无法还原具体时刻，返回 `status: "unavailable"`。由于当前调候资料尚无经双人 Gold 复核的条件适用标签，区间位置只作为寒暖燥湿证据，后续救应或替代条目也不会自动成为本命盘用神。`rules` 内保留 `xi_shen`、未裁决的 `ji_shen`、`source_text` 与证据状态。

`gan_zhi_analysis.zhi_relations` 从 `bazi-rules-2026-07-15.5` 起是地支关系的权威图。每条关系提供稳定 `id/rule_id`、全部 `pillars/branches`、`priority`、结构 `status`、`structure_status`、`transformation_status`、目标五行及冲突证据。两支只会标为 `半合`、`拱合` 或 `半会`；三支齐备才是 `三合局`、`三会局` 或 `三刑`。六合、完整三合和三会只证明结构齐备，成化保持 `unadjudicated`；与冲、刑、害、破共存时改为 `disputed` 并保留双方，不按优先级静默删除任何关系。从 `bazi-rules-2026-07-15.19` 起，重复且字段更少的 `clash_harmony` 兼容视图已删除，天干与地支完整关系图必须能从四柱逐字段重算，篡改或缺失关系证据的历史快照会被拒绝并重新计算。

`gan_zhi_analysis.gan_relations` 从 `bazi-rules-2026-07-15.6` 起完整覆盖十干比和、相生、相克与五合，不再使用只覆盖单一阴阳干的生克表。五合仅返回 `complete_structure`，目标五行的成化状态保持 `unadjudicated`；`transformation_evidence` 记录月支/月令五行、目标五行是否另透天干及地支根气，但不会自动据此宣称成化。一个天干同时参与多组五合时，相关关系标为 `disputed`，通过 `conflicts_with` 与 `dispute_reasons` 暴露争合/妒合；`proximity` 区分贴合与遥合。

`missing_elements` 从 `bazi-rules-2026-07-15.7` 起只表示原始五行计分事实，不是喜用神或补救结论。数组固定按木、火、土、金、水排序：得分为 0 进入 `missing_elements`，得分大于 0 且小于 5 进入 `weak_elements`；`scores` 保留五行原始得分，`missing_count` 仅为客观计数。响应固定返回 `status: "observed"`、`rule_id: "wuxing.raw-score-presence-v1"`、`is_yongshen_conclusion: false` 与 `remedy_status: "not_adjudicated"`。旧 `severity`、`remedy_elements`、`remedy_advice` 字段已删除，不再把缺失数量解释为严重程度，也不生成颜色、饰品、行业或方位建议。

`wuxing_flow` 已退役。旧实现仅按聚合五行分数是否达到经验阈值 10 拼接“木生火”等路径，无法证明柱位之间存在实际生化流通，也没有专家 Gold 校准。当前响应保留 `five_elements`、`element_detail`、`gan_zhi_analysis` 与 `missing_elements`，不再从总分推断流通路径、通关、完整循环、平衡或吉凶。

重复拼接五行观察的 `flow_pattern_desc`、`wuxing_flow`、`tong_guan` 与按单个大运天干五行生成结果判断的 `dayun_flow` 已从命盘、详情 DTO 和前端删除。五行观察只保留原始 `five_elements`、`element_detail` 与 `missing_elements`；大运结构以 `da_yun.pillars` 和日课 `fortune_layers` 为准。系统不再自动输出大运“增强”“减弱”“泄秀”，也不再由此推断竞争、求财、压力或其他现实影响。

历史快照中的 `five_elements`、`element_detail` 和 `missing_elements` 必须能由四柱完整重算。校验会重建 EightChar，重新计算天干与藏干分数及明细，再逐字段重算缺失/低分观察。

从 `bazi-rules-2026-07-15.21` 起，`ten_gods`、`hidden_stems`、`pillar_details`、`ming_gong`、`ten_god_proportion`、`ten_god_analysis` 以及全部命盘神煞字段形成统一的四柱派生事实合同。校验从四柱重建 EightChar，并使用标准化出生输入中的性别重算逐柱与全局神煞；数据库原始性别字段不参与裁决。只在名称数组与详情数组之间保持内部一致、但无法匹配四柱的篡改结果同样会被拒绝。

从 `bazi-rules-2026-07-15.22` 起，日期盘历史快照的 `da_yun_info` 与完整 `tiaohou` 必须和标准化出生时刻的实时重算结果逐字段一致，不再仅检查大运非空或调候区间值落在合法范围。`rule_meta` 也必须匹配权威清单中的全部表项、规则数量、来源提交、文件 SHA-256 和身强参数；只修改版本/门派标签可以保留历史标识，修改任何实际规则证据都会触发重算。

从 `bazi-rules-2026-07-15.9` 起，八字响应及其 `ten_god_analysis` 不再返回 `health_note`。五行原始分数和十神比例不能证明器官状态或医学风险，因此旧五行脏腑疾病映射及按比例生成的高血压、心血管、失眠、免疫力、手术等提示均已删除。干支关系文本也不再把冲克直接解释为健康波动。旧 `jin_bu_huan` 字段按日干输出夭折、损寿等断语，现已连同数据表删除。神煞原始字符串只保留名称与命中目标，不再拼接传统断语；`致死`、`天杀`、`死符`、`自缢煞`、`产厄` 等直接死亡、自伤或妊娠风险标签不再进入产品结果。命理输出只描述采用规则下的传统结构与解释，不作为疾病、寿命或人身风险预测。

从 `bazi-rules-2026-07-15.10` 起，命宫按月支与时支的寅起月序相加后，以 14/26 反推宫支，再用年干五虎遁取宫干；旧实现错误地把子起时支索引直接加到月序，导致大量命宫错位。命宫响应只保留干支、纳音和传统神煞名称；无 Gold 标签支持的 `shen_sha_desc`、`zhi_detail` 及其寿命、婚姻、财富、职业结果断语已删除，前端也不再按命宫神煞自动标吉凶。`rule_meta.tables` 新增 `calendar_core`，公开固定 `lunar-javascript` 与 `cnlunar` 提交、核心文件 SHA-256、MIT 许可证和 `silver_external/cross_checked_not_gold` 状态。可重建 fixture 含 21 个双源四柱共识案例、2024 年 12 个节令交接前一秒/交接秒共 24 案，以及 1 个明确隔离的上游争议案；45 个可裁判案例 exact 比较四柱、命宫、纳音、大运顺逆和前八柱。`cnlunar` 只有节气日期粒度，不参与秒级边界裁决；`lunar-javascript` 与生产依赖 tyme4go 同作者，边界证据具有相关性。因此这些测试只证明基础结构在声明口径下可复现，不进入传统解释准确率或 Gold 分母。

从 `bazi-rules-2026-07-15.11` 起，所有季节信息只从排盘后的 `month_pillar.zhi` 派生，不再使用公历出生月。新字段 `month_season` 返回月支、寅起传统月序、春夏秋冬归属、规则 ID、取值依据和 `observed` 状态。日期输入和四柱输入使用同一映射，节令交接秒会随月柱同步切换。未经 Gold 裁决且包含旺衰、贫夭、婚姻、伤残等结果断语的 `season_text`、`season_text_month`、`wuxing_season_note` 已从后端、详情 DTO 和前端删除；缺少或篡改 `month_season` 的旧快照不会作为完整快照使用。

```json
{
  "month_season": {
    "rule_id": "month-season.branch-order-v1",
    "month_branch": "寅",
    "traditional_month": 1,
    "season": "春",
    "basis": "month_pillar_branch",
    "status": "observed"
  }
}
```

从 `bazi-rules-2026-07-15.12` 起，`na_yin` 只返回柱干支到纳音名称和五行的固定映射，并明确携带 `rule_id`、`basis` 与 `observed` 状态。外部 Silver fixture 已对 45 个可裁判案例的四柱纳音名称做 exact 比较，但这不构成性格或人生结果的 Gold 证据。旧纳音 `image_desc`、`personality`、`energy_stage`、`modern_ext`、`judgments` 字段已删除；未经裁决的日柱复合文本 `ri_zhu_desc`、`ri_zhu_poem`、`ri_zhu_source`、`ri_zhu_comment`、`ri_zhu_hour_detail` 和六十甲子吉凶对象 `jia_zi_detail` 同步退出 API 与前端。缺少完整纳音事实证据的旧快照不会作为完整快照使用。

```json
{
  "na_yin": {
    "year": {
      "rule_id": "nayin.sixty-cycle-v1",
      "gan_zhi": "甲辰",
      "name": "覆灯火",
      "element": "火",
      "basis": "pillar_gan_zhi",
      "status": "observed"
    }
  }
}
```

从 `bazi-rules-2026-07-15.13` 起，神煞详情只记录传统名称、稳定规则 ID、查表依据、`observed` 命中状态与 `not_adjudicated` 解释状态。旧 `category`、`polarity`、`priority`、`source`、`description` 字段以及宣称“日柱最重要”的 `shen_sha_summary` 已删除；`shen_sha_by_pillar` 按年、月、日、时自然顺序返回，不再附加人生阶段角色或吉凶排序。前端只展示名称、命中目标与查法，不再自动着色或归类为吉凶。日课 `activated_shen_sha` 采用同一证据结构，旧“吉神/凶煞”类型、事件断语及每项固定加减 2 分逻辑已删除，神煞命中不再进入运势分数。旧 `ganzhi.go` 事件解释器也已删除；当前干支关系输出仅来自结构化关系图，不包含牢狱、第三者、婚姻质量或晚景结果推断。

```json
{
  "day_shen_sha_details": [
    {
      "name": "天乙贵人",
      "rule_id": "shensha.天乙贵人",
      "basis": "日干或年干查贵人地支",
      "status": "observed",
      "interpretation_status": "not_adjudicated"
    }
  ]
}
```

从 `bazi-rules-2026-07-15.14` 起，`ten_god_analysis` 只表示十神等权出现次数的稳定排序。计算口径固定为三处非日主透干加四支全部藏干，每次出现等权计数；百分比由计数重新计算，并列最高项通过 `dominant_gods` 全部保留，不再按固定数组顺序静默选出唯一主导项。响应明确标记 `validation_status: "not_validated"`、`interpretation_status: "not_adjudicated"` 和未考虑藏干深浅、月令强度的限制。旧 `personality`、`interpersonal`、`career_fortune`、`emotion_relation`、`taboo`、`summary` 以及逐十神 `meaning/advice` 字段已删除；命盘页和古籍解释接口不再从占比自动生成性格、事业、财富、婚姻、投资或行动建议。缺少完整新证据的历史快照会重新计算。

```json
{
  "ten_god_analysis": {
    "rule_id": "bazi.ten-god-occurrence-ranking-v1",
    "calculation_method": "three_visible_stems_and_all_hidden_stems_counted_equally",
    "total_occurrences": 12,
    "dominant_gods": ["伤官", "七杀"],
    "dominant_percent": 25,
    "ranked_gods": [
      {
        "rank": 1,
        "god": "伤官",
        "count": 3,
        "percent": 25,
        "basis": "three_visible_stems_and_all_hidden_stems_counted_equally",
        "status": "observed",
        "interpretation_status": "not_adjudicated"
      }
    ],
    "status": "observed",
    "validation_status": "not_validated",
    "interpretation_status": "not_adjudicated",
    "limitations": [
      "visible stems and hidden stems are counted equally",
      "hidden-stem depth and seasonal strength are not weighted",
      "occurrence share is not influence strength or outcome probability"
    ]
  }
}
```

`tiaohou` 只返回日干、月支查表条目和节令区间事实。`table_primary_candidate` 仅表示资料表第一条的顺序，固定标记 `validation_status: "not_validated"` 与 `interpretation_status: "not_adjudicated"`；出生时刻在前后“节”之间的位置由 `depth_evidence` 独立记录，但 `depth_affects_selection: false`，不会据此改选条目。每条表规则携带稳定规则 ID、`xi_shen`、`ji_shen`、原始 `source_text`、取值依据和未裁决状态。后续条件性救应或替代条目只保留为原始证据，不会无条件进入用神候选；旧多来源 `useful_god_evidence` 综合已退役。古籍解释和前端不再从调候表自动生成现实吉凶或行动建议。

```json
{
  "tiaohou": {
    "rule_id": "bazi.tiaohou.table-candidates-v3",
    "stem": "甲",
    "month": "卯",
    "rules": [
      {
        "rule_id": "bazi.tiaohou.table.甲.卯.1",
        "xi_shen": "庚",
        "ji_shen": "戊",
        "source_text": "资料表原始规则文本",
        "basis": "day_stem_month_branch_table",
        "status": "observed",
        "validation_status": "not_validated",
        "interpretation_status": "not_adjudicated"
      }
    ],
    "table_primary_candidate": "庚",
    "selection_basis": "first_table_entry_candidate",
    "depth_affects_selection": false,
    "depth_evidence": {
      "rule_id": "bazi.tiaohou.solar-term-depth-v1",
      "status": "observed",
      "basis": "solar_term_jie_interval",
      "phase": "前段",
      "start_term": "惊蛰",
      "end_term": "清明",
      "position": 0,
      "note": "区间位置仅作寒暖燥湿证据；当前规则表没有经 Gold 复核的分段适用标签，不据此改选主用。",
      "interpretation_status": "not_adjudicated"
    },
    "status": "observed",
    "validation_status": "not_validated",
    "interpretation_status": "not_adjudicated",
    "limitations": [
      "table order is not an independently adjudicated unique selection",
      "solar-term depth does not change candidate order",
      "table candidates do not imply favorable real-world outcomes"
    ]
  }
}
```

```json
{
  "missing_elements": {
    "status": "observed",
    "rule_id": "wuxing.raw-score-presence-v1",
    "missing_elements": ["木"],
    "weak_elements": ["火"],
    "scores": {"木": 0, "火": 3, "土": 8, "金": 9, "水": 10},
    "missing_count": 1,
    "is_yongshen_conclusion": false,
    "remedy_status": "not_adjudicated",
    "note": "原始五行计分中的缺失或偏低不等于喜用神，也不自动代表需要补入相应五行。"
  }
}
```

## 运势

### 今日运势

```http
POST /api/fortune
```

```json
{"chart_id":87,"query_date":"2026-06-16"}
```

`query_date` 必填，格式为 `YYYY-MM-DD`。今日、周、月接口中的每日数值统一使用同一结构关系流水线：中性起分 → 天干关系启发式权重 → 地支关系启发式权重。该数值只是干支结构的本地映射，未经 Gold 数据验证，不是现实事件概率，也不用于推断事业、财富、感情、健康、贵人、学业、投资、出行或是非结果。

关键评分与解释字段：

```json
{
  "engine_version": "fortune-engine-2026-07-15.8",
  "bazi_engine_version": "bazi-engine-2026-07-15.30",
  "bazi_resolution_source": "bazi_snapshot",
  "rule_version": "bazi-rules-2026-07-15.30",
  "score": 60,
  "evidence_completeness": 100,
  "supporting_evidence": [
    {
      "code": "relation.stem.shengWo",
      "stage": "relation",
      "category": "天干生克",
      "label": "流日生扶日主",
      "impact": 18,
      "description": "流日天干五行与日主天干五行构成生我关系。",
      "source": "本地启发式权重；结构取法参考《三命通会》十神生克规则",
      "evidence_basis": "empirical",
      "validation_status": "not_validated",
      "interpretation_status": "not_adjudicated",
      "is_outcome_conclusion": false
    }
  ],
  "counter_evidence": [],
  "score_breakdown": {
    "pipeline_version": "fortune-score-pipeline-2026-07-15.3",
    "score_kind": "structural_relation_index",
    "evidence_basis": "empirical",
    "validation_status": "not_validated",
    "interpretation_status": "not_adjudicated",
    "is_outcome_probability": false,
    "base_score": 50,
    "relation_score": 60,
    "final_score": 60,
    "evidence_completeness": 100,
    "supporting_evidence": [],
    "counter_evidence": []
  },
  "season_element": {
    "rule_id": "fortune.season-element.month-branch-v1",
    "reference_stem": "甲",
    "reference_element": "木",
    "query_month_branch": "未",
    "season": "夏",
    "basis": "reference_day_stem_element_and_query_month_branch",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  },
  "ten_god": {
    "rule_id": "rikuyo.ten-god-day-stem-v1",
    "reference_stem": "甲",
    "query_stem": "丙",
    "name": "食神",
    "basis": "reference_day_stem_and_query_day_stem",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  },
  "seasonal_state": {
    "rule_id": "rikuyo.seasonal-state-v1",
    "query_stem": "丙",
    "query_element": "火",
    "query_month_branch": "未",
    "season": "夏",
    "state": "旺",
    "basis": "query_day_stem_element_and_query_month_branch",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  },
  "fortune_layers": {
    "rule_version": "bazi-rules-2026-07-15.30",
    "school": "子平法",
    "dayun": {
      "rule_id": "fortune.layer.dayun-v2",
      "key": "dayun",
      "name": "大运",
      "pillar": "乙丑",
      "gan": "乙",
      "zhi": "丑",
      "start_at": "2030-12-12T07:19:00",
      "end_at_exclusive": "2040-12-12T07:19:00",
      "ten_god": {
        "rule_id": "rikuyo.ten-god-day-stem-v1",
        "reference_stem": "甲",
        "query_stem": "乙",
        "name": "劫财",
        "basis": "reference_day_stem_and_query_day_stem",
        "status": "observed",
        "interpretation_status": "not_adjudicated"
      },
      "relations": [],
      "shen_sha_details": [],
      "basis": "exact_start_time_and_query_time",
      "status": "observed",
      "interpretation_status": "not_adjudicated"
    },
    "liunian": {},
    "liuyue": {},
    "xiaoyun": {}
  }
}
```

从 `fortune-engine-2026-07-15.4` 起，未经裁决的九维事件模块退出生产路径。API 不再返回 `analysis`，评分拆解不再返回 `detail_score`，事业、财运、感情、健康、贵人、学业、投资、出行和是非的固定权重、星级、趋势、断语与建议均不参与总分。前端主数值改称“结构指数”，不再根据数值显示“大吉、良好、欠佳、低迷”等现实结果评价。旧版 `analysis.lucky_guide` fallback 同步删除，避免历史响应重新激活旧断语。

从 `fortune-engine-2026-07-15.5` 起，幸运色、幸运数字、财位、吉时、自动宜忌、开运指南和加持资源全部退出日、周、月响应，前端“运势加持”入口同步删除。周月汇总只保留可复现的描述统计：`structural_relation_index`、最高/最低指数及日期、均值、标准差、五行分布和十神频次最高项。旧 `weekly_score`、`monthly_score`、`best_*`、`worst_*`、`peak_days`、`low_days`、`good_streak`、`bad_streak` 与 `key_advice` 已删除；不再按任意阈值生成吉凶分段、连续段或行动建议。

从 `fortune-engine-2026-07-15.6` 起，`season_element_advice` 与 `flow_impact` 结果文本退出日、周、月响应。季节字段改为 `season_element`，只记录日主天干五行、查询日期月支及月支对应季节；十神字段改为结构化 `ten_god`，只记录日主天干与查询日干的十神映射。两者均携带规则 ID、取值依据、状态与 `interpretation_status: "not_adjudicated"`。旧 `today_ten_god`、`ten_god_favorable` 和 `ten_god_desc` 已删除，前端不再把十神着色为喜忌，也不再展示贵人、财富、事业、压力或行动建议等未经裁决解释。

从 `fortune-engine-2026-07-15.7` 起，日课嵌套对象只返回可复核的结构事实。`hidden_stems`、`stem_relations`、`branch_relations` 和 `fortune_layers` 不再包含 `favorable`、`is_favorable`、`score`、`description` 或自生成 `evidence` 等结果字段；每条记录统一携带规则 ID、输入、取值依据、状态和 `interpretation_status: "not_adjudicated"`。旧的独立 `dayun_influence`、`liunian_influence`、`advance_retreat`、`yongshen_impact`、`pattern_name`、`pattern_type`、`pattern_favorable` 与 `pattern_unfavorable` 已删除。月令旺相休囚死查表改由中性 `seasonal_state` 返回；大运、流年、流月和小运统一收敛到 `fortune_layers`，其中大运优先使用精确起运时刻与查询时刻选段，缺失起运时间时才明确标记 `integer_age_fallback`。

从 `fortune-engine-2026-07-15.8` 起，旧数据的 `integer_age_fallback` 与精确起运路径使用相同的覆盖边界语义。查询年龄小于 `start_age` 时返回 `active: false`、`index: -1`、`status: "before_start"`；超过 `start_age + 10 * 大运柱数` 时返回 `active: false`、末端索引和 `status: "after_covered_periods"`。旧实现会把起运前查询强行夹到第一步，并在已有大运柱全部结束后永久夹到最后一步，现已修正。完整日期盘仍以 `start_at` 和每十年的 `[start_at, end_at_exclusive)` 精确半开区间为权威；整数年龄只用于缺少起运时间的明确降级数据。

`fortune_layers` 是查询时刻周期结构的权威合同；命盘中的 `da_yun.pillars` 是大运序列与年龄边界的权威合同。二者只记录可复核结构，不从单个大运天干或五行关系生成现实结果。已删除的 `dayun_flow` 不再作为大运柱 fallback，也不会由历史快照重新激活。

从 `fortune-engine-2026-07-15.3` 起，传统日课标签只返回可审计的查表事实：

```json
{
  "twelve_stage": {
    "rule_id": "rikuyo.twelve-stage-v1",
    "reference_stem": "甲",
    "query_branch": "午",
    "name": "死",
    "basis": "reference_day_stem_and_query_day_branch",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  },
  "jian_chu": {
    "rule_id": "rikuyo.jianchu-month-branch-v1",
    "month_branch": "未",
    "query_branch": "午",
    "name": "闭",
    "basis": "query_month_branch_and_query_day_branch",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  },
  "huang_dao": {
    "rule_id": "rikuyo.twelve-officers-month-branch-v1",
    "month_branch": "未",
    "query_branch": "午",
    "name": "朱雀",
    "basis": "query_month_branch_qinglong_start_and_query_day_branch",
    "status": "observed",
    "interpretation_status": "not_adjudicated"
  }
}
```

建除和值神均以查询日期的月柱地支为基准，不再误用出生月支。上述名称不自动映射为吉凶，也不参与分数、健康状态、行动建议或时段建议。旧 `stage_favorable`、`stage_desc`、`stage_flexible`、`overall_verdict`、`favor_score` 字段已删除；日课内的彭祖禁忌与日课总断语也已删除。顶层 `peng_zu` 仍是独立历书原始字段，仅供查阅，不参与日课解释和评分。

日运、周运和月运统一从命盘恢复八字：先校验保存的标准化出生信息，再验证完整快照的版本、结构以及四柱一致性；可信快照用于保持创建时规则结果可复现，不可信或缺失的快照按标准化出生信息重算，最后才完整标准化原始出生字段。农历、时区和真太阳时输入不会被当作普通公历字段直接重排。`bazi_resolution_source` 的值为 `bazi_snapshot`、`normalized_birth` 或 `normalized_raw_birth`，`bazi_engine_version` 标识实际生成本次八字输入的引擎版本。

`evidence_completeness` 表示规则所需资料的覆盖程度（0–100），不是事件发生概率。接口不再使用 `confidence` 表达运势可信度。前端按普通、进阶、专业三层展示：普通层展示结构概览，进阶层展示正反证据，专业层展示指数拆解、证据代码/来源以及引擎和规则版本。

### 周运势

```http
POST /api/fortune/weekly
```

```json
{"chart_id":87,"start_date":"2026-06-16"}
```

响应：

```json
{
  "daily_fortunes": [],
  "structural_relation_index": 70,
  "element_trend": "[]",
  "summary": {
    "highest_index_day": "2026-06-18",
    "lowest_index_day": "2026-06-21",
    "highest_index": 82,
    "lowest_index": 44,
    "element_distribution": {"木": 0.2, "火": 0.2, "土": 0.2, "金": 0.2, "水": 0.2},
    "dominant_element": "木",
    "dominant_ten_god": "正印",
    "average_index": 70,
    "index_standard_deviation": 10.4
  }
}
```

### 月运势

```http
POST /api/fortune/monthly
```

```json
{"chart_id":87,"year":2026,"month":6}
```

月运响应与周运使用同一中性汇总契约，顶层均值字段同为 `structural_relation_index`。

### 运势历史

```http
GET /api/fortune/history?chart_id=87&page=1&page_size=10
```

## 卜易

卜易为登录用户维度的每日一卦，不绑定命盘。每日边界按服务端 `Asia/Shanghai` 日期计算。

### 今日卜易状态

```http
GET /api/buyi/today
```

未卜响应：

```json
{"date":"2026-07-05","has_record":false,"already_drawn":false,"record":null}
```

### 今日起卦

```http
POST /api/buyi/today
```

无请求体。若今日已卜，会返回同一条记录，并设置 `already_drawn=true`。

```json
{
  "date": "2026-07-05",
  "has_record": true,
  "already_drawn": false,
  "record": {
    "id": 1,
    "hexagram_number": 11,
    "hexagram_name": "地天泰",
    "score": 86,
    "level": "大吉",
    "summary": "今日得地天泰，上下交通，泰平盛世。",
    "human_way": "上下交通，泰平盛世。阴阳交泰，在家里是丈夫体贴、妻子贤惠，在公司是老板和员工一条心。",
    "image_reading": "三人同行、喜报、日月同辉（大吉之象）。",
    "advice": "今日气势较顺，可主动把握关键机会；仍以守正、谦和为底，不因顺境而躁进。",
    "source": "倪海厦天纪六十四卦详解",
    "created_at": "2026-07-05T09:00:00+08:00"
  }
}
```

## 紫微斗数

### 紫微排盘

推荐使用命盘 ID，让后端读取八字命盘并缓存紫微结果：

```http
POST /api/ziwei/chart
```

```json
{"chart_id":87,"profile":"ziwei-local-composite-v2"}
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
  "profile": "ziwei-local-composite-v2"
}
```

响应核心结构：

```json
{
  "profile_id": "ziwei-local-composite-v2",
  "engine_version": "ziwei-local-go-2026-07-16.45",
  "rule_version": "ziwei-rules-2026-07-16.45",
  "rule_school": "紫微斗数-本地综合规则-v2",
  "rule_sources": [{
    "rule_id": "ziwei.sihua.ten-stem.iztro-v1",
    "repository": "https://github.com/SylarLong/iztro",
    "commit": "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13",
    "path": "src/data/heavenlyStems.ts",
    "sha256": "f50a96b4fda42f834b2ff7aea36533a38bd5bcee94b74d5ecc1deb22ee07279c",
    "license": "MIT",
    "source_tier": "silver_external",
    "validation_status": "cross_checked_not_gold"
  }, {
    "rule_id": "ziwei.star-brightness.iztro-v1",
    "repository": "https://github.com/SylarLong/iztro",
    "commit": "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13",
    "path": "src/data/stars.ts",
    "sha256": "87d2fdbb4501db7b6aa054237cba4de38210e7b3dbd97e849ce1936c0f735b95",
    "license": "MIT",
    "source_tier": "silver_external",
    "validation_status": "cross_checked_not_gold"
  }, {
    "rule_id": "ziwei.period-chronology.iztro-normal-v1",
    "repository": "https://github.com/SylarLong/iztro",
    "commit": "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13",
    "path": "src/astro/FunctionalAstrolabe.ts",
    "sha256": "1748e1c8210a19aac1da7479e50237b30b160ac694d3ecbef5bb8ba4a0822f27",
    "license": "MIT",
    "source_tier": "silver_external",
    "validation_status": "cross_checked_not_gold"
  }, {
    "rule_id": "ziwei.transit-stars.iztro-v1",
    "repository": "https://github.com/SylarLong/iztro",
    "commit": "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13",
    "path": "src/star/horoscopeStar.ts",
    "sha256": "0cce679f149711bee610f62cdeabbbc12697a242847f297275cc846669bf29f8",
    "license": "MIT",
    "source_tier": "silver_external",
    "validation_status": "cross_checked_not_gold"
  }],
  "runtime_rule_tables_schema": "ziwei-runtime-rule-tables-v1",
  "runtime_rule_tables_hash": "d07c0eb29ec41f7ccf3b905ca86fab5aaa2b9e29955295346c7e12d4778deebc",
  "plugin_manifest": [],
  "plugin_manifest_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "calculation_input": {
    "calendar_type": "SOLAR",
    "year": 2003,
    "month": 4,
    "day": 15,
    "hour": 14,
    "minute": 0,
    "gender": "男",
    "basis": "normalized_solar_minute"
  },
  "input_fingerprint": "<规范出生输入的 SHA-256>",
  "content_hash": "<完整公开命盘内容的 SHA-256>",
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

`profile` 为空时使用 `ziwei-local-composite-v2`，未知 Profile 返回 `400`。Profile 固定声明插件的 `id@version` 和执行顺序；依赖必须出现在使用者之前，冲突、缺失或版本不符会在计算前失败。响应和缓存同时校验 `plugin_manifest` 与顺序敏感的 SHA-256，旧插件状态或不同顺序不能命中缓存。已废弃但曾被静默忽略的 `algorithm` 参数现在返回 `400`，调用方必须选择真实注册的 Profile。

准确率口径：上述版本和插件字段证明一次计算可以复现，不等于证明排盘正确。当前 `src/internal/service/testdata/ziwei_full_chart_gold.json` 尚无冻结、双人独立复核的真实全盘 Gold 案例，因此精度报告会阻止发布紫微准确率。旧 `ziwei_cases.json` 为 quarantined Bronze，只用于计算 smoke test 和十二宫、十四主星、四化等结构不变量，不参与任何产品准确率分母。未来 Gold 每盘必须固定同一 Profile/引擎/规则版本，提供来源页、许可证、来源哈希和两名复核者，并对命身宫、命身主、五行局、十二宫、主辅星、四化、四组十二神及完整大限执行 127 项 exact 检查。

从 `ziwei-rules-2026-07-15.2` 起，紫微核心排盘结构保持不变，但解释层不再返回流月/流日 `health` 字段，也不再依据地支五行映射肝胆、脾胃、心脑血管、肺肾等器官提示。周期建议删除自动体检、固定饮水量和健康管理内容；疾厄宫只作为传统宫位名称展示星曜、四化及三方四正结构，不据此推导个体身体状态。原“寿星入庙”检测器实际只验证天梁亮度为庙或旺，现按真实条件更名为“天梁庙旺”，并删除长寿、寿命和职业推断。上述边界不证明其余紫微解释正确，真实解释精度仍须由独立 Gold 数据验证。

从 `ziwei-rules-2026-07-15.3` 起，紫微解释曾统一标记 `evidence_basis: "empirical"`、`validation_status: "not_adjudicated"` 和 `is_outcome_conclusion: false`。单宫解释将原 `confidence` 改为 0-100 的 `evidence_completeness`，只表示输入结构覆盖，不是预测准确率；格局明细将原置信度改为 `structure_status: "matched"`。周期解释删除 `emotional_state`，并停止依据规则分给出投资、职业、婚期、具体时段或人生结果建议。合盘删除 `marriage_timing` 和预测式 `confidence`，只保留可复现的结构分、分数区间和未裁决状态。流年叠盘的 `good/watch` 改为中性的 `resource/constraint`，`advice` 改为 `review_note`。周期汇总的 `recommendations/risks` 改为 `review_notes/limitations`。其中周期分析与流年叠盘统一使用 `empirical` 的历史口径已由 `.22/.8` 的混合证据合同替代；单宫等其他解释合同不因此自动改变。

从 `ziwei-rules-2026-07-15.4` 起，十干四化只保留 `SiHuaTable` 一个权威表源，排盘、四化飞星、飞星链和流年叠盘不得各自维护副本。旧飞星链曾把庚干写成“太阳化禄、武曲化权、天同化科、太阴化忌”，与排盘主表不一致；现统一为 Profile 采用的 iztro 口径“太阳化禄、武曲化权、太阴化科、天同化忌”。`rule_sources` 固定返回上游仓库、提交、文件路径和许可证；该来源属于 `silver_external` 交叉验证，不是独立 Gold，因此不会增加准确率分母。缓存命盘缺少来源元数据或来源提交不一致时必须重新计算。

从 `ziwei-rules-2026-07-15.5` 起，固定 iztro 提交生成 20 个全盘 Silver 差分案例，覆盖十天干、男女、多个时辰、闰月和立春边界，并 exact 比较命身宫、命身主、五行局、十二宫宫名宫干、主辅星位置/亮度/四化、四组十二神和大限。差分发现并修正壬干禄存应在亥（索引 11），连带校正壬年擎羊、陀罗、博士十二神及流年流禄/流羊/流陀。亮度只保留 iztro `stars.ts` 明确提供表值的文昌、文曲、擎羊、陀罗、火星、铃星；左辅、右弼、天魁、天钺、禄存、天马、地空、地劫不再生成无来源亮度。两个上游文件均以 `sha256` 固定，来源变更或旧缓存元数据不完整时重算。20 案仍标记为 `silver_external/cross_checked_not_gold`，不能用于发布准确率；真实全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-15.6` 起，紫微缓存不再只凭 Profile、版本、来源和插件元数据命中。命盘新增 `input_fingerprint`，按规范化后的公历年月日时分与男/女性别生成 SHA-256；`content_hash` 对除自身外的完整公开命盘 JSON 生成 SHA-256，覆盖十二宫、星曜、亮度、四化、四组十二神、三方四正、格局和运限星表。命盘与周期两个缓存入口都必须同时通过 Profile、插件清单、来源、输入指纹和内容哈希校验；出生字段变化、旧缓存缺少合同或任何公开命盘内容被修改都会重新计算。该哈希用于一致性和意外篡改检测，不是密码学签名，也不证明排盘规则正确。

从 `ziwei-rules-2026-07-15.7` 起，保存命盘、临时排盘与周期三个入口不再绕过统一出生标准化。保存命盘优先使用创建时持久化的 `NormalizedBirth`，因此农历转换、时区、真太阳时修正和不确定时间候选选择会贯通紫微计算；旧记录缺失标准化对象时，才按完整原始输入和候选 ID 重建，无法唯一恢复时明确失败。临时排盘也会处理完整 `ChartRequest`，跨候选边界却未提交 `candidate_id` 时返回冲突而不静默强选。响应新增 `calculation_input`，公开紫微实际消费的规范公历年月日时分、性别和 `normalized_solar_minute` 口径，缓存输入指纹也绑定该对象。当前紫微核心只消费分钟粒度，标准化秒数不会参与排盘；涉及分钟边界的研究必须把这一限制纳入 Profile 与 Gold 标注。

从 `ziwei-rules-2026-07-15.8` 起，流年、流月、流日派生盘不再继承本命盘已经失效的 `content_hash`。派生响应改为返回 `derivation_type`、规范公历查询对象 `derivation_input`、`derivation_fingerprint`、`base_content_hash` 和 `derived_content_hash`；父哈希绑定实际本命盘公开内容，派生哈希覆盖除自身外的父哈希、查询参数和完整派生盘公开 JSON。该中间合同可发现未同步哈希的父盘、查询或流曜变化，但同步重算无密钥派生哈希仍可绕过；现行完整校验以 `.25` 的父盘恢复与派生重建合同为准。派生盘没有本命 `content_hash`，不能被误当作可缓存本命盘；非法年、月或不存在的公历日期返回 `400`，不再按系统当前时间静默替换。哈希只用于一致性审计，不是签名，也不增加规则正确性或预测准确率证据。

从 `ziwei-rules-2026-07-15.9` 起，紫微运限统一采用固定 iztro 提交的 `horoscopeDivide: normal` / `fixLeap: true` 历法口径。`liunian.year` 表示农历年份标签；流月和流日必须由完整目标公历年月日转换为农历周期，普通流月以农历初一分界，闰月十五日后按 iztro 口径进入下一月干支。旧实现以公历月中旬的节气月柱计算流月四化，解释层又按出生日干和月/日序号生成另一套干支，现均已删除。派生盘、结构分析和文字解释只读取同一 `derivation_input.period_gan_zhi`，查询参数与派生合同不一致时拒绝分析；`period_summary` 也先按目标公历日期解析农历年标签，春节前不会再把当前公历年流年与上一农历年的流月、流日混合。固定上游生成器新增 8 个运限 Silver 案例，覆盖官方样例、农历新年前后、闰月十五/十六和普通日期；这些案例是外部实现一致性证据，不是专家 Gold。

从 `ziwei-rules-2026-07-15.10` 起，运限星曜不再用不完整的四化数组冒充流耀分布。固定 iztro `horoscopeStar.ts` 的魁钺昌曲禄羊陀马鸾喜规则进入 Profile：`liu_nian_stars` 每盘固定安置 10 颗流耀并增加 1 颗年解，`liu_yue_stars` 与 `liu_ri_stars` 各安置 10 颗对应月/日流耀。周期四化分别进入 `liu_nian_four_hua`、`liu_yue_four_hua`、`liu_ri_four_hua`，不再混入流耀数组；分析层的 `focus_palaces[].four_hua` 只表示本层四化，不再错误复制本命宫位四化。三层十二宫名称分别进入 `liu_nian_palaces`、`liu_yue_palaces`、`liu_ri_palaces`；流月以目标农历年流年命宫为起点，叠加出生农历月、出生时支、目标农历月及闰月十五日修正，流日再叠加目标农历日序。`focus_palaces[].period_palace` 明确返回同一地支上的本层宫名，避免把本命宫名当作流年/月/日宫名。运限 Silver 扩展到 18 例，其中连续 10 个流日覆盖甲至癸全部天干，并对三层逐地支宫位名称、流耀位置和四化集合执行 exact 差分。该修复提高固定 Profile 下的结构完整性，不证明宫位、流耀或评分具有预测效度。

从 `ziwei-rules-2026-07-15.11` 起，导出的杂曜与四组十二神 helper 不再根据出生年干支、月份和五行局重新排盘，而是按地支索引直接投影核心命盘中已经计算并公开的 `adjective_stars`、`changsheng_12`、`boshi_12`、`jiangqian_12` 与 `suiqian_12`。旧 helper 使用另一套重复表和混合坐标口径，新算盘会与核心十二宫发生确定性错宫；命盘经 JSON 缓存后，未序列化的运行时出生字段还会造成进一步漂移。重复的 `ziwei_stars.go` 旧表已删除，新算盘与缓存回放必须 exact 返回相同的权威宫位字段。该修复只消除同一 Profile 内的双算法与序列化漂移，不增加真实全盘 Gold，也不证明传统排法或现实预测有效。

从 `ziwei-rules-2026-07-15.12` 起，宫位解读中的三方四正四化会照直接投影对宫和两个三合宫已经公开的 `four_hua`。旧实现只有在对宫/三合宫化曜名称同时匹配本宫主星时才输出“照/拱”说明；十四主星在全盘唯一安置，该条件基本不可能成立，导致真实存在的四化会照长期被静默丢弃。现对宫每项生成“化曜照本宫”，两个三合宫每项生成“化曜拱本宫”，并对新算盘与 JSON 回放逐宫 exact 验证。该变化修复结构解读缺失，不把传统四化会照解释升级为现实结果证据，也不增加 Gold。

从 `ziwei-rules-2026-07-15.13` 起，`sihua_chain` 与 `self_mutagen` 改用宫干飞化语义。每宫以公开 `heavenly_stem` 查询十干四化表，发出宫与化曜在公开 `stars` 中的实际落宫相同才标记自化；旧实现只读取本命年干四化，又把化曜落宫同时写成源宫与目标宫，导致所有命中项都被误报为自化。`.13` 中间版本曾把直接跨宫边包装成 `chain_depth=1`，再求和为 `total_chain_depth`，并输出 `key_mutagens`；这些字段已由 `.23` 的直接飞行边合同替代。新算盘与 JSON 回放继续执行 exact 比较；该变化修复结构语义和缓存复现性，不验证自化解释的经验效度，也不增加 Gold。

从 `ziwei-rules-2026-07-15.14` 起，`sihua_feixing` 直接投影命盘十二宫已经公开的本命 `four_hua`，不再用未序列化的 `YearStem` 重新选表，也不再从 `json:"-"` 的 `MainStars` / `AuxStars` 重建星曜落宫。旧实现对新算盘可以返回四条本命四化，但仅将持久化命盘 JSON 反序列化后，年干会是零值且私有星曜数组为空，四组飞星无法独立复现；当前周期接口虽会通过 `AttachBirthData` 补回这些运行时字段，正确性仍额外依赖调用者不得遗漏该补水步骤。现逐宫解析严格的“星名 + 化禄/化权/化科/化忌”后缀，目标宫直接取该 `four_hua` 所在宫，并复用统一的结构说明生成器；非法标签被忽略而不会猜测。真实完整命盘的新算结果与仅 JSON 回放结果执行 exact 比较，每种四化固定验证一条。该修复使结果成为持久化命盘自身可复现的确定性投影，不增加专家 Gold，也不证明四化解释具有预测效度。

从 `ziwei-rules-2026-07-15.15` 起，大限排布和 `ziwei-period-2026-07-15.4` 大限结构分析只从持久化命盘的公开合同恢复计算参数。顺逆行所需生年干阴阳与性别由规范化 `calculation_input` 重建，起限年龄由 `five_bureau` 解析，起宫由 `earthly_branch_of_soul_palace` 定位，每限星曜直接投影对应宫位的公开 `stars`；不再读取 JSON 中不会保存的 `birthData`、`YearStem`、`JuValue`、`SoulBranch`、`MainStars` 或 `AuxStars`。旧实现对新算盘可正常返回大限，但纯 JSON 回放会把局数退成 0、顺逆依据退成零值并清空全部星曜；真实失败案例的第一限曾从“5 岁命宫”漂移为“0 岁父母宫”。现阳男、阳女、阴男、阴女四种方向组合均要求新算盘与不调用 `AttachBirthData` 的 JSON 回放在完整 `Dayun` 和大限分析上 exact 一致；内容哈希、输入指纹、公开出生输入、五行局、命宫地支或十二宫地支映射不完整时拒绝计算，不再用零值或数组位置猜测。当前 Handler 仍可为其他运限解释恢复运行时出生数据，但大限本身不再依赖该补水步骤。该修复提高缓存复现性，不增加专家 Gold，也不证明大限解释或现实预测有效。

从 `ziwei-rules-2026-07-15.16` 起，流年、流月、流日派生图和 `ziwei-period-2026-07-15.5` 周期解释统一从持久化本命盘公开合同恢复出生上下文。恢复路径必须同时通过 `content_hash`、`input_fingerprint` 与规范化 `calculation_input` 校验，再由公开出生年月日时分和性别重建生年支、月柱支、日干支及时支；流月、流日宫位不再读取未序列化的 `birthData`，周期十神、命局刑冲合、经验评分、流年叠盘说明和三层摘要也不再依赖调用方先执行 `AttachBirthData`。旧实现中，纯 JSON 本命盘仍可生成流年结构，但流月与流日会直接返回空，流年叠盘会静默退回 `score=60` 和空十神，三类周期分析还可能因空出生对象触发 panic。现三层派生图公开 JSON、三层结构分析、流年叠盘解释和周期摘要均要求新算盘与无补水 JSON 回放 exact 一致；本命合同无效、派生哈希无效或 `base_content_hash` 不指向当前本命盘时明确拒绝，不再制造默认解释。`NewPeriodInterpreterFromChart` 提供同一公开恢复入口，原 `NewPeriodInterpreter` 仅保留给已经持有可信 `BirthData` 的内部调用。该修复提高缓存可复现性与失败可见性，不增加专家 Gold，也不证明周期解释或现实预测有效。

从 `ziwei-rules-2026-07-15.17` 起，本地格局检测与合盘结构分析统一以十二宫公开 `stars`、星曜 `brightness` 和规范化 `calculation_input` 为权威输入。旧格局检查器仍有多处直接读取 `json:"-"` 的 `MainStars`、`AuxStars` 与 `Brightness`；真实命盘仅经 JSON 回放后会丢失“机月同梁格”并误报“空宫”。旧合盘评分同样从这些私有字段计分，并直接读取不序列化的 `YearStem`，使缓存命盘的宫位分数退化、双方年干同时变成零值并错误落入同干结构。现公开 `stars` 即使为空数组也具有权威性，运行时私有字段只允许作为没有公开数组的测试夹具兼容层；新算盘、无补水 JSON 回放以及私有字段恶意污染三种路径必须得到 exact 相同的格局列表和合盘结果。合盘还必须先通过双方 `content_hash`、`input_fingerprint` 与规范化出生输入校验，任一合同非法时返回空结果。该修复提高结构复现性和篡改可见性；合盘分数仍标记 `empirical/not_adjudicated`，不会进入 Gold，也不证明关系预测有效。

从 `ziwei-rules-2026-07-15.18` 起，合盘删除误导性的 `stem_structure` 字符串。旧实现把同干称为“帝旺格”、索引相邻称为“长生格”，又把实际构成甲己、乙庚、丙辛、丁壬、戊癸五合的索引差 5 错称为“墓绝格”，混淆了十二长生与天干关系。新 `stem_relation` 只返回双方生年天干、各自五行、`five_combine / same_element / generates / controls` 关系类型、确定方向、规则 ID、结构状态和成化状态。五合优先标记并给出目标五行，但 `transformation_status` 固定为 `unadjudicated`，不得解释为已经合化；所有关系均声明 `is_outcome_conclusion: false`，不参与关系质量、婚姻结果或事件时点推断。十干 10×10 有序组合均执行确定性门禁，新算与纯 JSON 回放必须 exact 一致。

从 `ziwei-rules-2026-07-15.19` 起，合盘删除 `rule_score`、`shuang_gong_lian_can`、`wu_bu_xing_yun` 及其所有宫位分、差值档和分数档。旧“五步行运”只是把命宫、夫妻、福德、事业、财帛五个单盘宫位经验分再次平均，与“双宫联参”逐项重复；“夫妻宫对冲”“福德宫共鸣”“官禄宫合作”等名称没有实现任何跨盘对冲、共鸣或合作规则，因此最终分数不是兼容性测量。新 `palace_comparisons` 只并列五个同名宫位的完整公开字段投影，并返回按 A 盘顺序去重的共同星曜、共同四化标签和共同杂曜集合；共同项不表示契合，差异项也不表示冲突。双方还必须属于完全相同的当前计算 Profile，并通过十二宫名和地支唯一、宫干合法、恰好一个身宫标记的结构门禁，否则返回空结果。

从 `ziwei-rules-2026-07-15.20` / `ziwei-period-2026-07-15.6` 起，流年、流月、流日、时辰、周期分析和流年叠盘不再公开 `score`、`tone`、`level`、`overall_tone`、`key_tips`、`effect` 或 `qi_zi_effect`，也不再把十神、刑冲合、星曜亮度、四化或流曜按固定基线和加减权重换算成 0–100 分、高中低区间或现实事件强弱。大限阶段和触发宫位不再按启发式权重排序或截断，改为按规则生成顺序与命盘十二宫顺序完整返回。周期地支与本命年、月、日、时四柱逐柱返回 `relation_evidence`；该中间版本曾使用泛化 `ziwei.period.branch-relation-v1`，现已由下一段 `.21/.7` 的逐关系稳定 ID 合同替代。同一地支对同时成立的冲、合、刑关系全部保留。该版本同时修正十二地支五行映射为子亥水、寅卯木、巳午火、申酉金、丑辰未戌土，并将自刑集合修正为辰、午、酉、亥。上述变化只提高结构正确性、透明度和回放复现性，不增加专家 Gold；真实紫微全盘 Gold 仍为 0，也不证明周期预测有效。

从 `ziwei-rules-2026-07-15.21` / `ziwei-period-2026-07-15.7` 起，周期地支关系与项目内八字权威关系图统一：补全子未、丑午、寅巳、卯辰、申亥、酉戌六害，以及子酉、丑辰、寅亥、卯午、巳申、未戌六破。逐柱两支关系不再把寅巳申或丑戌未中的任意两支称为完整“三刑”，统一输出 `relation: "相刑"`，并以 `subtype` 区分 `无礼之刑`、`无恩之刑`、`恃势之刑` 与 `自刑`；只有三支齐备的独立结构合同才允许称“三刑”。每条证据改用 `ziwei.period.branch.{fuyin|clash|liuhe|harm|punish|break}.<地支对>-v1` 形式的具体稳定 `rule_id`，并固定公开 `transformation_status` 与 `is_outcome_conclusion: false`。六合只确认两支结构完整，返回目标五行，同时将成化状态固定为 `unadjudicated`；冲、刑、合、害、破同时成立时全部保留，不静默取舍。该修复是内部规则一致性门禁，不是独立专家 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-15.22` / `ziwei-period-2026-07-15.8` 起，周期分析与流年叠盘不再以笼统的 `evidence_basis: "empirical"` 混合描述两类性质不同的输出。干支、四化、流曜落点、宫位投影和地支关系属于确定性规则投影，统一声明 `placement_basis: "deterministic_rule_projection"`；资源、约束、移动等文本仍只是未裁决的传统解释标签，统一声明 `interpretation_basis: "traditional_rule_labels"`、`interpretation_status: "not_adjudicated"`。外层 `evidence_basis` 固定为 `mixed_deterministic_projection_and_unadjudicated_traditional_labels`，`validation_status` 保持 `not_adjudicated`，且 `is_outcome_conclusion: false`。大限阶段、周期重点、聚焦宫位、证据，以及叠盘的方法步骤、四化触发、流曜触发和聚焦宫位均逐项携带同一语义字段，避免调用方把确定性落点误当作已经验证的现实解释。该拆分不裁决传统标签，不验证现实结果，也不增加专家 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.23` 起，本命年干四化落宫、宫干四化飞行和自化检测统一收敛为纯确定性投影合同。旧 `sihua_chain` 只生成源宫到目标宫的一次直连，并未递归计算链，却公开 `chain_depth` 与求和后的 `total_chain_depth`；旧 `star_affinity` 只是把目标宫全部非主星计数，既未区分辅弼与煞杂曜，也没有“亲和度”的裁决依据；三个入口还输出混合传统主题的 `effect`。现删除 `effect`、`chain_depth`、`total_chain_depth`、`star_affinity`、`key_mutagens` 与 `mutagen_type`。本命四化使用 `transformed_star/hua_type/target_palace`；宫干飞化逐边使用 `source_palace/source_palace_stem/transformed_star/hua_type/target_palace/flight_scope/is_self_mutagen`；自化项明确返回同宫结构。外层和逐项均固定声明十干四化 `rule_id`、`source_tier: "silver_external"`、`placement_basis: "deterministic_rule_projection"`、`validation_status: "cross_checked_not_gold"` 与 `is_outcome_conclusion: false`。历史 `period_type: "sihua_chain"` 名称暂保留为路由枚举，但响应 `analysis_kind` 明确为 `direct_palace_stem_four_hua_flights`，不再声称计算了链、亲和度或现实效果。该修复消除伪结构指标，不增加 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.24` 起，当前无插件 Profile 的缓存命中不再只验证版本元数据和 `content_hash`。内容哈希只能证明公开 JSON 与哈希自洽；旧合同会接受被修改后重新计算哈希的错误命盘，例如交换十二宫数组顺序、修改宫干或命身字段、移动核心星曜、替换四化、十二神、三方四正或格局列表。现 `chartMatchesProfile` 先从已认证的 `calculation_input/input_fingerprint` 重建农历出生上下文，再 exact 重算命身宫、宫位顺序与地支宫干、命身主、五行局、完整核心主辅星输出及亮度、四化、杂曜、长生/博士/岁前/将前十二神、三方四正和格局；任何字段不一致即拒绝缓存，重新哈希也不能绕过。插件 Profile 必须先提供对应的结构校验器；当前门禁对非空插件清单采取拒绝策略，避免未经校验的插件变更获得缓存信任。该校验提高结构完整性和事故可见性，但无密钥内容哈希仍不是抗攻击签名，也不增加专家 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.25` 起，流年、流月、流日派生合同也从“哈希自洽”升级为完整重建。`ValidDerivedChartContract` 会从派生 JSON 清除三层流曜、流四化、流宫名和派生元数据，以 `base_content_hash` 恢复内嵌本命盘；恢复结果必须先通过 `.24` 的当前 Profile 结构门禁，再用规范 `derivation_input` 重新执行对应流年、流月或流日算法，最终派生指纹、父哈希和完整 `derived_content_hash` 必须 exact 一致。同步重算派生哈希后篡改流曜、四化、流宫名、内嵌本命星曜，或把合法的次日查询元数据与前一日结果拼接，都会被拒绝。三个派生计算入口也先要求本命盘通过当前 Profile，而不再只依赖输入指纹和内容哈希。该修复提高父子结构完整性和缓存污染可见性；它仍只证明固定 Profile 的确定性重建，不裁决运限规则或现实结果，真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.26` 起，当前 Profile 的本命结构门禁统一进入全部生产服务消费者。`ZiWeiService` 的格局检测、四化飞星、大限、宫干四化直接飞行、单宫解读、查询视图、自化和合盘，在读取命盘结构前都必须通过 `chartMatchesProfile`；结构错误但同步重算哈希的命盘统一返回 `nil`，不能只因 Handler 通常先校验缓存就绕过服务层防线。流年、流月、流日已在 `.25` 接入同一门禁。无 Profile 元数据的合成命盘仍可调用包内纯计算函数进行规则单测，但不能通过生产服务方法。`sihua_chain` Handler 的旧递归链说明同步改为“宫干四化直接飞行”，与 `.23` 的实际数据合同一致。该变化提高入口一致性与失败可见性，不增加专家 Gold，也不验证格局、飞化或解读的现实效度；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.27` / `ziwei-period-2026-07-16.9` 起，周期解释 Handler 与分析构造器完成同一合同收敛。旧 `liunian_interpretation`、`liuyue_interpretation`、`liuri_interpretation` 分支仍访问已经从结果类型删除的 `overall_tone`、`key_tips`、`score`、`effect`、`qi_zi_effect` 与时辰分数字段，并通过运行时 `GetBirthData()` 创建解释器；现 Handler 使用 `NewPeriodInterpreterFromChart`，直接序列化当前 `LiunianResult/LiuyueResult/LiuriResult`，公开干支、十神、逐柱关系证据、时辰结构、结构摘要及非结果声明，不再手工维护易漂移的字段白名单。解释器绑定创建它的本命 `content_hash`，每次解释前执行 `.25` 派生完整重建，不能分析另一张本命盘的有效派生结果。`BuildDayunAnalysis` 也必须通过当前 Profile，并将传入阶段与本命盘重算的大限 exact 比较；伪造起限年龄、宫位或星曜的阶段会被拒绝。该修复恢复 Handler 与服务类型的一致性并封堵分析层旁路，不增加专家 Gold，也不验证周期现实结果；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.28` / `ziwei-period-2026-07-16.10` 起，流日十二时辰不再返回含义不明的 `hour: 1,3,...,23`。该数字实际接近区间结束边界，却没有说明子时跨越午夜，也容易被界面误解为开始小时。现每个时辰逐项返回天干、地支、干支、`interval_start_hour`、右开区间 `interval_end_hour_exclusive`、可读区间、`crosses_midnight`、日干依据、边界口径与五鼠遁规则 ID。子时固定表示 `23:00-00:59`、起点 `23`、结束边界 `1`、`crosses_midnight: true`；其余十一时辰为不跨日的连续两小时槽。时干按 `(日干五组索引×2 + 时支索引) mod 10` 的五鼠遁表 exact 计算，覆盖十日干×十二时支。当前紫微流日查询只确定目标日期的流日干，因此边界口径固定声明 `traditional_two_hour_branch_slots_no_civil_date_assignment`：展示传统时辰槽，但不裁决子初属于哪个民用日期，也不生成行动时段建议。该修复消除时间标签歧义，不增加 Gold 或现实预测证据；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.29` / `ziwei-period-2026-07-16.11` 起，三层 `PeriodSummary` 删除残留的 `advice` 合同。该字段实际只包含“仅核对结构”“以现实证据为准”等安全边界文本，却可能被客户端按算命行动建议展示。现类型更名为 `PeriodReviewNotes`，JSON 固定使用 `review_notes`，并按流年、流月、流日分组；结构摘要、关系证据和 `evidence_basis` 保持不变。序列化安全门禁要求紫微周期结果不出现 `"advice"`，同时必须出现 `review_notes`。该变化只修正语义与下游误用风险，不增加 Gold 或预测建议，真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.30` / `ziwei-period-2026-07-16.12` 起，紫微出生与运限干支统一使用严格解析器。旧 `stemFromRune/branchFromRune` 在未知字符时静默返回 `0`，会把异常上游值伪造成甲或子；出生四柱字符串长度异常还可能在取 rune 下标时 panic。现干支必须 trim 后恰好两个汉字，第一字命中十天干、第二字命中十二地支，且干支索引阴阳奇偶一致，甲丑、乙子等不可能出现在六十甲子中的配对明确拒绝。出生年、月、日、时柱全部通过同一入口，解析后的时支还必须与输入民用小时映射一致；流年/月/日 `period_gan_zhi` 同样使用严格解析。测试覆盖完整六十甲子、全部 24 个出生小时以及长度、未知字符和奇偶错配反例。该修复消除静默甲子兜底和潜在越界，不增加 Gold 或传统解释证据；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.31` / `ziwei-period-2026-07-16.13` 起，紫微流年、流月、流日与十二时辰的十神统一调用八字权威 `ClassifyTenGod`，不再维护第二套可能漂移的五行生克实现。周期入口会先严格校验所见天干与本命日干索引，任一越界即返回失败，不再把非法索引经五行辅助函数静默当作木而生成伪造十神。测试逐项比较十干×十日主共 100 种组合，并覆盖非法索引、三层周期和十二时辰输出。该门禁只证明项目内分类一致性和失败可见性，不验证周期解释或现实预测准确性；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.32` 起，后端删除 `SimplifiedZiWei`、`SimplifiedDayun` 与 `SimplifiedLiunian` 占位 API。旧实现忽略绝大多数出生输入，在十二宫重复填入相同主辅星、固定格局和伪大限，却返回正式 `ZiWeiChart/Dayun` 类型，存在被包内调用者误当作真实排盘的风险。唯一测试消费者改用明确的空合成夹具；生产排盘只保留 `ZiWeiService.CalculateChart/CalculateChartWithProfile`，继续执行规范出生输入、当前 Profile、结构重建和内容哈希合同。该删除不改变合法命盘结果，也不增加 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.33` 起，星曜亮度描述不再把未知星名、未知等级或缺失等级静默回退为“平”。`GetStarBrightness` 只接受亮度知识表中的精确星曜与等级组合，并显式返回是否命中；单宫主星分析、亮度摘要和证据也保留“亮度未提供”状态，不生成 `(平)` 或“亮度等级为平”的伪证据。当前 Profile 的合法命盘仍由结构重建门禁 exact 校验主星亮度，因此正常排盘结果不变。该修复提高异常与不完整夹具的可见性，不裁决亮度解释或现实结果；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.34` 起，缓存运行时补水 `AttachBirthData` 必须先验证服务当前 Profile 的完整命盘结构，并要求调用方出生输入与持久化 `calculation_input` exact 一致。旧实现可把另一组出生输入挂到命盘上，命宫或身宫地支未知时还分别回退到出生年支或命宫支，使错误缓存获得看似可用的私有字段。现 Profile、结构、输入、命宫地支、身宫地支任一不合法即报错，且在全部检查通过前不修改命盘；合法 JSON 回放仍可恢复私有字段，但所有公开计算继续只依赖认证后的持久化合同。该修复提高错误缓存可见性，不改变合法排盘结果，不新增 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.35` / `ziwei-period-2026-07-16.14` 起，周期解释器只能通过 `NewPeriodInterpreterFromChart` 从通过当前声明 Profile 完整校验的本命盘创建。旧导出 `NewPeriodInterpreter(*BirthData)` 可直接注入任意日干与四支，且空 `baseContentHash` 被解释为跳过父盘绑定，使调用者能用伪出生上下文解释任意有效派生盘。现删除原始构造器，解释器的父盘内容哈希必须非空并与派生盘 `base_content_hash` exact 一致；缺失或错误哈希的三层解释全部失败。该修复保证周期十神与刑冲合上下文来自派生盘的认证父盘，不裁决周期现实结果；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.36` 起，`ZiWeiChart` 不再导出 `GetBirthData()`。旧方法直接返回内部 `*BirthData` 指针，外部调用者可原地修改新算盘的生年干支、日干和四支，形成与公开 `calculation_input/content_hash` 不一致的双重状态。该方法已经没有生产消费者；周期解释、运限和知识消费者均从认证公开命盘重建上下文。现运行时出生对象只保留为包内兼容状态，不能通过公开方法读取或修改。该删除不改变 JSON、合法排盘或周期结果，不新增 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.37` 起，`ZiWeiChart` 彻底删除不序列化的 `birthData`、`YearStem`、`YearBranch`、`SoulBranch`、`BodyBranch`、`SoulStem` 与 `JuValue`。这些字段只在新算盘组装或遗留补水中写入，生产运限、周期、格局、合盘与解读已经全部从认证公开合同重建，保留它们只会让新算盘和 JSON 回放拥有不同内存状态。历史 `AttachBirthData` 方法现仅验证当前 Profile、完整结构和 exact 出生输入，不再修改命盘；五行局与年干测试也改从公开 `five_bureau/calculation_input` 恢复。合法 JSON 与结果不变。该清理建立单一事实来源，不新增 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.38` 起，`PalaceInfo` 删除不序列化的 `MainStars`、`AuxStars` 与 `Brightness` 影子字段。旧组装器和插件会同时写公开 `stars` 与三组私有投影，部分格局、叠盘和解读 helper 在 `stars=nil` 时还回退到私有字段，使新算盘、合成夹具与 JSON 回放可能选择不同事实来源。现主辅星名称、类型和亮度只从 `stars` 读取；插件只追加 `StarOutput`，无来源亮度只清空对应 `StarOutput.brightness`。所有合成测试也直接构造公开星曜数组。该清理不改变合法 JSON，消除星曜双写和回退漂移，不新增 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.39` 起，紫微单宫解读删除 `evidence_completeness`。旧值以 64 为任意基线，按是否有主星、四化、三方备注、格局和证据条数固定加减，再强制夹在 45–95；它没有需求清单、覆盖分母或 Gold 校准，不能表示证据完整率。现单宫结果直接公开逐条 `evidence`、三方上下文、格局明细、限制、依据和验证状态，由调用方查看实际缺失项，不再压缩成伪百分比。fortune 接口中基于另一套输入检查的同名字段不在本次范围。该删除不改变排盘，不新增 Gold；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.40` / `ziwei-period-2026-07-16.15` 起，紫微单宫和周期证据删除含因果暗示的 `impact` 字段，统一改为结构性 `basis`。身宫、十二神、三方四正、煞曜与四化证据只说明公开落位、索引关系和传统标签，不再输出“行动权重上升”“事件触发”“控制风险暴露”或四化现实主题。单宫旧 `evidence_basis: empirical` 同时退出，改为 `mixed_deterministic_projection_and_unadjudicated_traditional_labels`，并分别公开 `placement_basis`、`interpretation_basis`、`interpretation_status` 与 `is_outcome_conclusion: false`。该修复收紧证据语义，不改变星曜排布，也不证明传统标签具有现实预测效度；真实紫微全盘 Gold 仍为 0。

从 `ziwei-rules-2026-07-16.41` 起，单宫解读只能通过 `ZiWeiService.GetPalaceReading` 从通过当前服务 Profile、内容哈希和完整本命盘结构重算校验的命盘生成。旧包级导出 `GetPalaceReading` 只检查指针和宫位索引，能够绕过认证入口，对不完整、篡改或派生盘产生看似有效的解读；现该生成器改为包内私有函数，处理器和测试统一走服务入口，派生盘与越界宫位明确拒绝。该修复保证解读证据绑定认证本命盘，不新增 Gold 或现实结果裁决。

从 `ziwei-rules-2026-07-16.42` / `ziwei-query-2026-07-16.1` 起，四化飞星链、自化、合命结构和查询视图删除可绕过服务认证的包级导出生成器，只保留 `ZiWeiService` 方法作为公开入口。查询视图接受当前服务 Profile 下通过完整结构校验的本命盘，或能由认证本命盘与规范派生输入逐字段重建的流年、流月、流日盘；仅重算 `derived_content_hash` 的篡改派生盘仍会拒绝。HTTP 本命与周期响应统一由服务方法生成 `query_view`，不再直接读取任意 `ZiWeiChart`。底层合成结构测试继续调用包内私有生成器，但生产调用无法绕过认证。该修复不改变星曜、四化或宫位算法，也不新增专家 Gold。

从 `ziwei-rules-2026-07-16.43` 起，三方四正索引、宫名映射、命盘映射和四化会照生成器全部改为包内私有函数。命盘相关生成器遇到空盘或越界宫位直接返回失败，不再回退到固定宫名或对非法索引取模；合法排盘仍按命盘公开地支定位对宫与两个三合宫。无生产消费者的 `ComputeAdjectiveStars`、`ComputeTwelveShen` 复制投影同时删除，杂曜与四套十二神只以认证命盘 `palaces` 中的权威字段发布，JSON 回放测试直接比较这些字段。该修复减少未认证重复投影和错误回退，不改变合法命盘结果，也不新增专家 Gold。

从 `ziwei-rules-2026-07-16.44` 起，删除包级导出 `CalculateZiWeiChart(*BirthData)`。旧入口可接受调用方自行拼装且互相矛盾的公历、农历、干支和闰月字段，直接返回缺少 Profile、插件清单、规范输入指纹和内容哈希的正式 `ZiWeiChart`；现核心函数改为包内私有，只能由 `ZiWeiService.CalculateChart/CalculateChartWithProfile` 在完成公历出生输入规范化后调用，并在返回前盖上完整可复现合同。非法日期、小时、分钟和性别统一失败，不存在原始排盘兜底。该修复不改变合法出生输入的排盘算法，也不新增专家 Gold。

从 `ziwei-rules-2026-07-16.45` 起，Profile 与命盘新增 `runtime_rule_tables_schema` 和 `runtime_rule_tables_hash`，以规范 JSON 对当前进程实际消费的天干地支、五行局、主星序列、亮度、四化、辅星、杂曜、命身主、十二神、流曜和结构解读表计算 SHA-256。外部来源 SHA 证明参考文件版本，该运行时摘要另行证明内存中的实现表与固定 Profile 一致；两者不能互相替代。计算前、缓存命中和派生盘重建都会重新计算摘要，任一表被进程内代码修改时，新排盘与旧缓存同时失败关闭。当前 schema 为 `ziwei-runtime-rule-tables-v1`，摘要为 `d07c0eb29ec41f7ccf3b905ca86fab5aaa2b9e29955295346c7e12d4778deebc`。该门禁提高可复现性，不表示规则已通过专家 Gold 或现实预测验证。

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
| `year` | number | `liunian` 中表示农历年份标签；流月/流日中是目标公历年 |
| `month` | number | 流月/流日的目标公历月 |
| `day` | number | 流月/流日的目标公历日，用于唯一确定所在农历月及流日 |
| `palace_idx` | number | `palace_reading` 必填，0-11 |
| `chart_id2` | number | `heming` 必填 |

| `period_type` | 响应字段 | 说明 |
|---------------|----------|------|
| `dayun` | `periods` | 大限列表 |
| `liunian` | `periods[0]` | 流年星曜分布 |
| `liuyue` | `periods[0]` | 流月星曜分布 |
| `liuri` | `periods[0]` | 流日星曜分布 |
| `sihua_feixing` | `periods` | 四化飞星 |
| `sihua_chain` | `chain` | 宫干四化直接飞行边（历史枚举名） |
| `self_mutagen` | `self_mutagens` | 自化检测 |
| `palace_reading` | `reading` | 单宫结构化详解 |
| `heming` | `heming` | 合盘 |
| `liunian_interpretation` | `periods[0]` | 流年文字解释 |
| `liuyue_interpretation` | `periods[0]` | 流月文字解释 |
| `liuri_interpretation` | `periods[0]` | 流日文字解释 |
| `period_summary` | `summary` | 流年/月/日汇总 |
| `liu_nian_stars` | `palaces` | 流年星曜列表 |

四化投影项只描述规则落点，不返回主题解释或评分。`sihua_feixing` 的落点项使用 `transformed_star/hua_type/target_palace`；`sihua_chain` 的直接宫干飞行边形状如下：

```json
{
  "analysis_kind": "direct_palace_stem_four_hua_flights",
  "rule_id": "ziwei.sihua.ten-stem.iztro-v1",
  "source_tier": "silver_external",
  "placement_basis": "deterministic_rule_projection",
  "validation_status": "cross_checked_not_gold",
  "is_outcome_conclusion": false,
  "hua_lu": [{
    "source_palace": "命宫",
    "source_palace_stem": "甲",
    "transformed_star": "廉贞",
    "hua_type": "化禄",
    "target_palace": "命宫",
    "flight_scope": "same_palace",
    "is_self_mutagen": true,
    "rule_id": "ziwei.sihua.ten-stem.iztro-v1",
    "source_tier": "silver_external",
    "placement_basis": "deterministic_rule_projection",
    "validation_status": "cross_checked_not_gold",
    "is_outcome_conclusion": false
  }]
}
```

`dayun`、`liunian`、`liuyue`、`liuri` 的 `analysis`，以及 `/api/ziwei/overlay` 的 `overlay_analysis`，都使用以下混合证据边界；其阶段、重点、证据、方法、触发项和聚焦宫位也逐项返回后四个语义字段：

```json
{
  "evidence_basis": "mixed_deterministic_projection_and_unadjudicated_traditional_labels",
  "placement_basis": "deterministic_rule_projection",
  "interpretation_basis": "traditional_rule_labels",
  "interpretation_status": "not_adjudicated",
  "validation_status": "not_adjudicated",
  "is_outcome_conclusion": false
}
```

`placement_basis` 只覆盖干支、星曜、四化、宫位与地支关系等规则落点；传统解释标签没有随落点一起获得现实有效性。周期分析和叠盘外层不得回退为 `evidence_basis: "empirical"`。

`liuri_interpretation.periods[0].hourly_analysis` 的子时项示例：

```json
{
  "stem": "甲",
  "branch": "子",
  "stem_branch": "甲子",
  "interval_start_hour": 23,
  "interval_end_hour_exclusive": 1,
  "interval_label": "23:00-00:59",
  "crosses_midnight": true,
  "day_stem_basis": "period_derivation_day_stem",
  "boundary_policy": "traditional_two_hour_branch_slots_no_civil_date_assignment",
  "rule_id": "ziwei.period.hour-stem.five-rat-v1",
  "shi_shen": "...",
  "relation_to_ming": "...",
  "relation_evidence": [],
  "structural_summary": "...",
  "evidence_basis": "deterministic_rule_projection",
  "validation_status": "not_adjudicated",
  "is_outcome_conclusion": false
}
```

结束小时采用右开边界；子时的 `1` 表示到 `01:00` 前，不表示开始小时或推荐时刻。该合同不把跨午夜的两段自动归属于某个民用日期。

`period_summary.summary` 的安全提示字段：

```json
{
  "review_notes": {
    "liunian": ["仅核对流年干支、十神和刑冲合结构", "重要决定以现实证据和专业意见为准"],
    "liuyue": ["仅核对流月干支、十神和刑冲合结构", "周期结构不得作为财务或职业决策依据"],
    "liuri": ["仅核对流日与时辰规则结构", "时辰结构不代表吉凶或行动建议"]
  },
  "evidence_basis": "deterministic_rule_projection",
  "validation_status": "not_adjudicated",
  "is_outcome_conclusion": false
}
```

紫微周期汇总不返回 `advice`；上述文本是使用限制，不是命盘生成的行动方案。

`heming.stem_relation` 为结构事实合同，不是匹配度或婚姻结论。响应不再包含旧 `stem_structure`：

```json
{
  "stem_a": "甲",
  "stem_b": "己",
  "element_a": "木",
  "element_b": "土",
  "relation_type": "five_combine",
  "relation_label": "天干五合",
  "direction": "mutual",
  "five_combine_target": "土",
  "structure_status": "complete_structure",
  "transformation_status": "unadjudicated",
  "rule_id": "heming.year-stem.five-combine.jia-ji",
  "evidence_basis": "deterministic_traditional_rule",
  "validation_status": "not_adjudicated",
  "is_outcome_conclusion": false,
  "notes": "仅确认五合配对，不裁决合化成功或现实结果。"
}
```

`heming.palace_comparisons` 固定按命宫、夫妻、福德、事业、财帛排列。每项的 `chart_a` 与 `chart_b` 是对应宫位公开字段的精确投影，`shared_*` 只是集合交集：

```json
{
  "palace": "命宫",
  "chart_a": {
    "branch": "午",
    "heavenly_stem": "戊",
    "is_body_palace": false,
    "stars": [],
    "four_hua": [],
    "adjective_stars": [],
    "changsheng_12": "帝旺",
    "boshi_12": "博士",
    "jiang_qian_12": "岁驿",
    "sui_qian_12": "岁建"
  },
  "chart_b": {
    "branch": "申",
    "heavenly_stem": "庚",
    "is_body_palace": true,
    "stars": [],
    "four_hua": [],
    "adjective_stars": [],
    "changsheng_12": "临官",
    "boshi_12": "力士",
    "jiang_qian_12": "攀鞍",
    "sui_qian_12": "晦气"
  },
  "shared_stars": [],
  "shared_four_hua": [],
  "shared_adjective_stars": [],
  "comparison_basis": "same_named_palace_exact_public_projection",
  "interpretation_status": "not_adjudicated",
  "is_compatibility_result": false
}
```

合盘响应不提供总分、宫位分、差值档或“对冲/共鸣/合作”结论。

流月、流日及对应解释/汇总请求必须同时提供 `year/month/day`，或者三个字段全部省略以使用服务器当天日期；部分日期返回 `400`。这是必要约束，因为一个公历月通常横跨两个农历月，只有 `year/month` 无法唯一确定流月。

`liunian`、`liuyue`、`liuri` 的 `periods[0]` 额外返回派生完整性合同。本命输入仍由 `calculation_input` / `input_fingerprint` 标识；周期查询由以下独立字段标识：

```json
{
  "derivation_type": "liuri",
  "derivation_input": {
    "calendar_type": "SOLAR",
    "year": 2026,
    "month": 7,
    "day": 15,
    "basis": "target_solar_date_resolved_to_lunar_day",
    "boundary_policy": "iztro_normal_lunar_boundaries_fix_leap_day_15",
    "resolved_lunar_date": {
      "year": 2026,
      "month": 6,
      "day": 2,
      "is_leap_month": false
    },
    "period_gan_zhi": "乙未"
  },
  "derivation_fingerprint": "<规范周期查询的 SHA-256>",
  "base_content_hash": "<本命盘公开内容的 SHA-256>",
  "derived_content_hash": "<完整派生盘公开内容的 SHA-256>"
}
```

派生盘不返回 `content_hash`；该字段只属于可缓存本命盘。流年输入使用 `calendar_type: "LUNAR_YEAR"`、`month/day: 0` 和 `basis: "target_lunar_year_label"`。流月与流日都记录完整公历日期、解析后的农历日期和实际干支，`basis` 分别为 `target_solar_date_resolved_to_lunar_month`、`target_solar_date_resolved_to_lunar_day`。

运限结构字段按层分离：

```json
{
  "liu_nian_stars": [["流魁"], ["流钺"], ["流昌"], ["流曲"], ["流禄"], ["流羊"], ["流陀"], ["流马"], ["流鸾"], ["流喜"], ["年解"], []],
  "liu_nian_four_hua": [["廉贞化禄"], ["破军化权"], ["武曲化科"], ["太阳化忌"], [], [], [], [], [], [], [], []],
  "liu_nian_palaces": ["命宫", "父母", "福德", "田宅", "事业", "交友", "迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟"],
  "liu_yue_stars": [["月魁"], ["月钺"], ["月昌"], ["月曲"], ["月禄"], ["月羊"], ["月陀"], ["月马"], ["月鸾"], ["月喜"], [], []],
  "liu_yue_four_hua": [["廉贞化禄"], ["破军化权"], ["武曲化科"], ["太阳化忌"], [], [], [], [], [], [], [], []],
  "liu_yue_palaces": ["命宫", "父母", "福德", "田宅", "事业", "交友", "迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟"],
  "liu_ri_stars": [["日魁"], ["日钺"], ["日昌"], ["日曲"], ["日禄"], ["日羊"], ["日陀"], ["日马"], ["日鸾"], ["日喜"], [], []],
  "liu_ri_four_hua": [["廉贞化禄"], ["破军化权"], ["武曲化科"], ["太阳化忌"], [], [], [], [], [], [], [], []],
  "liu_ri_palaces": ["命宫", "父母", "福德", "田宅", "事业", "交友", "迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟"]
}
```

数组顺序与响应 `palaces` 一一对应，不按固定子/丑/寅索引解释；跨实现差分必须先用每个宫位的 `branch` 转成地支映射。宫名数组每层恰好包含一次十二宫名。示例只说明字段形状，实际宫名、星曜和四化由出生信息、目标农历日期及本命星曜位置决定。

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
    "evidence": [{"type":"main_star","label":"主星","value":"廉贞(陷)","basis":"廉贞为本宫主星，亮度等级为陷；具体个体结果未裁决"}],
    "sanfang_context": {"opposite":"迁移","trine1":"财帛","trine2":"事业","notes":[]},
    "pattern_details": [{"name":"廉贞破军同宫","palace":"命宫","stars":["廉贞","破军"],"basis":"廉贞、破军同在命宫","structure_status":"matched","validation_status":"not_adjudicated"}],
    "review_notes": ["..."],
    "limitations": ["..."],
    "evidence_basis": "mixed_deterministic_projection_and_unadjudicated_traditional_labels",
    "placement_basis": "deterministic_rule_projection",
    "interpretation_basis": "traditional_rule_labels",
    "interpretation_status": "not_adjudicated",
    "validation_status": "not_adjudicated",
    "is_outcome_conclusion": false
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
  "status": "fallback",
  "reason": "citation_metadata_incomplete",
  "chart_id": 87,
  "focus": "overview",
  "summary": "...",
  "sections": [{"title":"...","content":"...","citations":[]}],
  "citations": [{
    "id": 1,
    "book": "渊海子平",
    "author": "unrecorded",
    "edition": "same_corpus_web_export_2026_unverified",
    "volume": "",
    "chapter": "001",
    "page": "",
    "locator": "chapter:001",
    "path": "bazi/渊海子平/001.md",
    "artifact_path": "library/渊海子平.pdf",
    "artifact_sha256": "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5",
    "document_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    "quote": "...",
    "quote_sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
    "source_tier": "classical_text_local",
    "verification_status": "artifact_hash_verified_same_corpus_export_not_independent",
    "artifact_kind": "chromium_web_export",
    "provenance_status": "same_corpus_web_export_detected",
    "independence_status": "not_independent_from_markdown",
    "coverage_status": "same_corpus_export_chapter_structure_observed",
    "catalog_schema": "bazi_rag_source_catalog_v1",
    "catalog_version": "2026-07-17.3",
    "catalog_sha256": "5a012c68833eaa1163a175f579833cca2337de2d7b22eeb2ec9ba396038059d5",
    "claim_eligible": false,
    "score": 0.91
  }]
}
```

引用只有在作者、版次、页码或稳定 locator、artifact/Markdown/短引文 SHA-256、来源层级、目录 schema/版本/文件 SHA-256、书目 provenance、来源独立性、全文覆盖和人工目录资格全部完整时，才会设置 `claim_eligible=true` 并进入段落 `citations`。`chromium_web_export`、`not_independent_from_markdown` 或非 `complete_primary_text_verified` 的不完整 artifact 均不得支撑声明。元数据不完整返回 `citation_metadata_incomplete`；引用完整但原文不支持当前焦点时返回 `citation_not_supporting_claim`。两种情况都保留审计引用，但不会声称“结合经典条文”生成结论。

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
  "consent_research": false,
  "consent_training": false,
  "engine_version": "bazi-engine-2026-07-15.30",
  "rule_version": "..."
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
| 事件字段 | Pilot 伦理、隐私与评分方案批准前，公开请求不接受或持久化事件年份、类别或结果 |
| `engine_version` | 可选，最多 64 个 Unicode 字符；空值使用命盘保存版本，旧命盘回退当前版本 |
| `rule_version` | 可选，最多 64 个 Unicode 字符；空值使用命盘保存版本，旧命盘回退当前版本 |

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
