# 八字算命 Web 应用

## TL;DR

> **Quick Summary**: 构建完整的八字+紫微斗数算命 Web 应用。用户一次设置生辰八字，即可每日查看八字运势，以及紫微斗数命盘、大限、流年分析。周/月运势 + ECharts 折线图。预留 AI 分析和五行元素图片接口。Go + Gin + tyme4go/lunar-go + Vue 3 + ECharts + MySQL + Docker Compose。

> **Deliverables**:
> - Go REST API 后端（用户认证、八字排盘、每日/每周/每月运势、**紫微斗数排盘/大限/流年**、AI分析占位）
> - Vue 3 SPA 前端（登录注册、排盘输入、八字命盘、运势折线图、**紫微命盘图**、历史记录）
> - MySQL 数据库（用户、命盘、查询历史、运势查找表）
> - Docker Compose 一键部署（Go + Vue/Nginx + MySQL）

> **Estimated Effort**: Large
> **Parallel Execution**: YES — 6 waves
> **Critical Path**: Scaffolding → DB Schema → Bazi Service → Fortune Engine → API Handlers → Frontend → Integration

---

## Context

### Original Request
用户需要一个八字算命项目。Go 后端（src/），Vue 前端（vue/），引用 6tail/lunar-javascript 相关库。用户输入生辰八字即可计算今日运势。

### Interview Summary
**Key Discussions**:
- 功能范围：完整命盘 + 完整黄历运势（四柱、五行、十神、纳音、藏干、大运、流日、宜忌、幸运色、幸运数字、吉时、冲煞、财神方位）
- 核心引擎：选用 6tail/tyme4go（lunar-go 升级版），Go module 直接引用
- 用户系统：需要注册登录（JWT），MySQL 存储用户+命盘+全历史
- 输入方式：公历/农历/直接八字，全部支持，自动识别
- 需要收集性别（影响大运起运方向）
- 部署：Docker Compose（Go + Vue/Nginx + MySQL）
- 测试策略：TDD（Go testing + Vitest）
- UI：新中式融合风

**Extension (v1.2) — 紫微斗数 (加强)**:
- 紫微斗数命盘排盘（12 宫 + 14 主星 + 辅星 + 四化 + **星曜亮度 + 格局分析**）
- 大限分析 + 流年分析 + **流月分析 + 流日分析**
- **四化飞星详细分析**（化禄/化权/化科/化忌飞入各宫）
- **宫位结构化解读文案**（基于规则模板的每宫详解）
- **流年叠盘交互可视化**（基础命盘叠加流年星曜变化）
- 紫微命盘可视化（**亮度颜色编码 + 格局标注 + 叠盘切换**）

**Research Findings**:
- 6tail/tyme4go（⭐54, MIT）是 lunar-javascript 作者的最新 Go 库
- API：Solar → Lunar → EightChar → WuXing/ShiShen/NaYin/DaYun/LiuNian/LiuRi
- 今日运势 = 本命日柱 vs 今日日柱的生克关系 + 黄历规则库
- **kaecer68/ziwei-zenith**（Go, MIT）— 紫微斗数完整引擎（亮度/格局/神煞/大限/流年）
- **tommitoan/bazica**（Go, ⭐24, MIT, 36 releases）— Go 八字排盘备选方案，1900-2100 范围

**Borrowed from Reference Libraries (research findings)**:
- **china-testing/bazi** (Python, ⭐1,289): 五行分数量化算法、冲刑合会可视化、《三命通会》解读模板
- **jinchenma94/bazi-skill** (Claude Code Skill, ⭐1,279): 9 本经典命理典籍的结构化规则数据
  - `classical-texts.md` — 穷通宝典/三命通会/滴天髓/渊海子平/千里命稿/协纪辨方书/果老星宗/子平真诠/神峰通考 核心论命规则
  - `wuxing-tables.md` — 五行/天干地支/十神/藏干完整参考表
  - `dayun-rules.md` — 大运顺逆排规则、起运年龄计算
  - 计划直接复用其参考数据到解读引擎（Task 40），节省大量模板编写

### Metis Review
**Identified Gaps** (addressed):
- 运势内容定义：确认 → 完整黄历运势（宜忌、幸运色/数字、吉时、冲煞、财神方位）
- 性别字段：确认 → 加入输入表单，传给 tyme4go 计算大运
- 运势文字生成：确认 → 规则模板引擎 + 黄历查找表
- 其他（自动处理）：Pinia 状态管理、Tailwind CSS、Nginx 反向代理、北京时区、日期范围 1900-2100

---

## Work Objectives

### Core Objective
构建八字+紫微斗数双系统算命 Web 应用。用户一次设置生辰，即可查看八字运势（日/周/月+折线图）和紫微命盘（12宫+大限+流年）。

### Concrete Deliverables
- `src/` Go 后端：REST API（auth + chart + daily/weekly/monthly fortune + **ziwei chart + period** + AI stub + history）
- `vue/` Vue 3 前端：SPA（登录、注册、首页、排盘、日/周/月运势 + ECharts、**紫微命盘图**、历史、AI 占位、元素图片占位）
- MySQL：5+1 张表（新增 ziwei_charts 表）
- `docker-compose.yml`：一键部署
- TDD 测试：Go 单元测试 + Vue Vitest 组件测试

### Definition of Done
- [ ] 用户可注册登录
- [ ] 用户可输入出生日期 + 性别
- [ ] 系统正确计算八字四柱（天干地支）+ 五行分布 + 十神 + 纳音 + 大运
- [ ] 系统正确计算紫微斗数命盘（12 宫定位 + 主星辅星安放 + 四化）
- [ ] 系统正确输出紫微大限 + 流年分析
- [ ] 系统正确输出今日/本周/本月八字运势
- [ ] 前端展示八字运势 ECharts 折线图（评分 + 五行占比）
- [ ] 前端展示紫微命盘图（12 宫可视化）
- [ ] AI 分析占位端点 + 五行元素图片占位
- [ ] `docker-compose up` 一键启动全栈
- [ ] 所有 Go 测试通过
- [ ] 所有 Vue 测试通过

### Must Have
- 八字排盘（四柱天干地支 + 五行 + 十神 + 纳音 + 大运）
- 八字运势（日/周/月）+ ECharts 折线图
- **紫微斗数排盘（12 宫 + 主星 + 辅星 + 四化）**
- **紫微斗数大限 + 流年分析**
- **紫微命盘可视化前端组件**
- 用户认证（JWT）+ 查询历史
- AI 分析占位 + 五行元素图片占位
- Docker Compose 一键部署

### Must NOT Have (Guardrails)
- 奇门遁甲等非八字/紫微命理系统
- 付费/订阅系统
- 多语言国际版（仅中文）
- 移动 App
- AI 生成的运势文字（仅规则模板；AI 接口仅为占位）
- 真实的 AI 分析实现（预留）
- 真实的五行元素图片资源（仅占位 URL）
- 紫微流月/流日/流时（仅大限+流年）
- 过度抽象的工具类

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: NO（全新项目，需搭建）
- **Automated tests**: TDD
- **Go Framework**: `go test` (standard testing + testify)
- **Vue Framework**: Vitest + Vue Test Utils

### QA Policy
Every task includes agent-executed QA scenarios. Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **API/Backend**: Bash (curl) — 发送请求，断言状态码 + JSON 响应字段
- **Frontend/UI**: Playwright — 导航、填表、点击、断言 DOM、截图
- **CLI/Build**: Bash — 运行命令，校验退出码 + 输出

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Start Immediately — scaffolding, MAX PARALLEL):
├── Task 1: Go scaffolding + go.mod + deps [quick]
├── Task 2: Vue scaffolding + Vite + Tailwind + Pinia + ECharts [quick]
├── Task 3: DB schema + GORM models + migrations [quick]
├── Task 4: Docker Compose config [quick]
├── Task 5: Shared API types / DTOs (incl. weekly/monthly/AI/element stubs) [quick]
├── Task 6: Go config module [quick]
└── Task 7: Nginx config template [quick]

Wave 2 (After Wave 1 — backend core, MAX PARALLEL):
├── Task 8: User model + JWT middleware (TDD) [quick]
├── Task 9: tyme4go Bazi calculation service (TDD) [deep]
├── Task 10: Input parser service (TDD) [quick]
├── Task 11: Fortune lookup data + seed (宜忌/吉时等) [unspecified-high]
├── Task 12: Fortune computation engine (daily + weekly + monthly) (TDD) [deep]
├── Task 13: Auth API handlers (TDD) [quick]
└── Task 35: ZiWei calculation service (kaecer68/ziwei-zenith) (TDD) [deep]

Wave 3 (After Wave 2 — backend APIs, MAX PARALLEL):
├── Task 14: Bazi chart handler (TDD) [quick]
├── Task 15: Daily fortune handler (TDD) [quick]
├── Task 16: History handler (TDD) [quick]
├── Task 28: Weekly fortune handler (TDD) [quick]
├── Task 29: Monthly fortune handler (TDD) [quick]
├── Task 30: AI analysis stub endpoint [quick]
├── Task 36: ZiWei chart handler (TDD) [quick]
├── Task 37: ZiWei period handler (大限/流年) (TDD) [quick]
├── Task 17: Router + middleware wiring [quick]
└── Task 18: Go Dockerfile + build [quick]

Wave 4 (After Wave 1 — frontend core, MAX PARALLEL):
├── Task 19: Auth pages (Login/Register) + auth store (TDD) [visual-engineering]
├── Task 20: Home page + input form component (TDD) [visual-engineering]
├── Task 21: Bazi chart display component (TDD) [visual-engineering]
├── Task 22: Daily fortune display component (TDD) [visual-engineering]
├── Task 23: History page (TDD) [visual-engineering]
├── Task 31: ECharts fortune line chart component [visual-engineering]
├── Task 32: Weekly fortune page [visual-engineering]
├── Task 33: Monthly fortune page [visual-engineering]
├── Task 34: Element image display + AI entry component [visual-engineering]
├── Task 38: ZiWei chart visualization component [visual-engineering]
├── Task 39: ZiWei page (命盘 + 12宫详解 + 大限/流年) [visual-engineering]
└── Task 24: API client layer (Axios) + request interceptors [quick]

Wave 5 (After Waves 3 + 4 — integration):
├── Task 25: Frontend-backend wiring + E2E flow [unspecified-high]
├── Task 26: Nginx reverse proxy + Vue production build [quick]
└── Task 27: Docker Compose full integration test [unspecified-high]

Wave FINAL (After ALL tasks — 4 parallel reviews):
├── Task F1: Plan compliance audit (oracle)
├── Task F2: Code quality review (unspecified-high)
├── Task F3: Real manual QA (unspecified-high)
└── Task F4: Scope fidelity check (deep)
→ Present results → Get explicit user okay

Critical Path: Task 1 → Task 9 → Task 12 → Task 15 → Task 28 → Task 29 → Task 32 → Task 33 → Task 25 → Task 27 → F1-F4
Parallel Speedup: ~70% faster than sequential
Max Concurrent: 10 (Wave 4)
```

**Critical Path**: Task 1 → Task 9 → Task 12 → Task 15 → Task 25 → Task 27 → F1-F4
**Parallel Speedup**: ~65% faster than sequential
**Max Concurrent**: 7 (Wave 1), 6 (Waves 2 & 4)

### Dependency Matrix

- **1-7**: — — 8-42, Wave 1 foundation
- **8**: 1, 6 — 13
- **9**: 1 — 12, 14
- **10**: 1 — 14
- **11**: 3 — 12
- **12**: 9, 11 — 15, 28, 29
- **35**: 1, 10 — 36, 37, 40
- **40**: 35 — 36
- **13**: 8 — 17
- **14**: 9, 10 — 17, 25
- **15**: 12 — 17, 25
- **16**: 8, 14, 15 — 17, 25
- **28**: 12 — 17, 25
- **29**: 12 — 17, 25
- **30**: 5 — 17
- **36**: 35, 40 — 17, 25
- **37**: 35 — 17, 25
- **17**: 13, 14, 15, 16, 28, 29, 30, 36, 37 — 25
- **18**: 1 — 27
- **19-24**: 2 — 25
- **31-34**: 2, 5 — 25
- **38-42**: 2, 5 — 25
- **25**: 14-17, 19-42 — 27

### Agent Dispatch Summary

| Wave | Count | Profiles |
|------|-------|----------|
| 1 | 7 | quick ×7 |
| 2 | 8 | quick ×3, deep ×4, unspecified-high ×1 |
| 3 | 10 | quick ×9, unspecified-high ×1 |
| 4 | 14 | visual-engineering ×13, quick ×1 |
| 5 | 3 | unspecified-high ×2, quick ×1 |
| FINAL | 4 | oracle ×1, unspecified-high ×2, deep ×1 |

> **Total: 42 implementation + 4 verification = 46 tasks**

---

## TODOs

### Wave 1 — 项目脚手架（7 tasks，全并行）

- [x] 1. Go 后端脚手架搭建

  **What to do**:
  - 创建 `src/go.mod`（module: `bazi`），初始化 Go module
  - 安装依赖：`gin-gonic/gin`, `jinzhu/gorm`, `go-sql-driver/mysql`, `golang-jwt/jwt/v5`, `6tail/tyme4go`, `kaecer68/ziwei-zenith`
  - 创建目录结构：`cmd/`, `internal/handler/`, `internal/service/`, `internal/model/`, `internal/middleware/`, `internal/config/`, `migrations/`
  - 创建 `src/cmd/main.go`：最小 Gin 启动 + health check endpoint
  - 创建 `src/Dockerfile`（多阶段构建：go build → alpine 运行）
  - 跑通 `go mod tidy && go build ./cmd/ && go run ./cmd/`

  **Must NOT do**:
  - 不要写业务逻辑代码
  - 不要创建非 Go 文件

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯文件结构创建 + 依赖安装，机械性任务
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2-7)
  - **Blocks**: Tasks 8-18
  - **Blocked By**: None

  **References**:
  - `https://github.com/6tail/tyme4go` — tyme4go 官方仓库，确认 import path 和 API
  - `https://gin-gonic.com/docs/quickstart/` — Gin 官方快速入门

  **Acceptance Criteria**:
  - [ ] `src/go.mod` 存在，包含 gin, gorm, mysql driver, jwt, tyme4go
  - [ ] `go build ./cmd/` 编译成功
  - [ ] `src/cmd/main.go` 可运行，`curl localhost:8080/health` 返回 200

  **QA Scenarios**:
  ```
  Scenario: Go backend compiles and starts
    Tool: Bash
    Preconditions: Go 1.21+ installed
    Steps:
      1. cd src && go mod tidy
      2. go build -o /tmp/bazi-server ./cmd/
      3. /tmp/bazi-server &
      4. sleep 2
      5. curl -s http://localhost:8080/health
      6. kill %1
    Expected Result: curl 返回 {"status":"ok"} 且 HTTP 200
    Failure Indicators: 编译错误、端口冲突、无响应
    Evidence: .sisyphus/evidence/task-1-build-start.txt (build + curl 输出)
  ```

  **Commit**: YES
  - Message: `chore(backend): init Go project with Gin, GORM, tyme4go deps`
  - Files: `src/go.mod`, `src/go.sum`, `src/cmd/main.go`, `src/Dockerfile`

- [x] 2. Vue 前端脚手架搭建

  **What to do**:
  - 在 `vue/` 目录中用 Vite 创建 Vue 3 + TypeScript 项目
  - 安装依赖：`vue-router`, `pinia`, `axios`, `tailwindcss`, `@tailwindcss/vite`, `element-plus`
  - 配置 Tailwind CSS + 自定义主题（新中式配色：朱砂红 #C41E3A、黛蓝 #2B3A42、宣纸白 #F5F0E8、墨色 #1A1A1A）
  - 创建目录结构：`src/views/`, `src/components/`, `src/api/`, `src/stores/`, `src/router/`
  - 配置 Vue Router（首页 `/`、登录 `/login`、注册 `/register`、命盘 `/chart`、运势 `/fortune`、历史 `/history`）
  - 创建 `vue/Dockerfile`（多阶段：npm build → nginx 静态服务）
  - 跑通 `npm run dev` 确认首页可访问

  **Must NOT do**:
  - 不写业务组件代码
  - 不写 API 调用代码

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
    - Reason: 前端脚手架注重 UI 基础（Tailwind 主题配置、Element Plus 引入）
  - **Skills**: []
  - **Skills Evaluated but Omitted**:
    - `playwright`: 脚手架阶段不需要浏览器测试

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3-7)
  - **Blocks**: Tasks 19-24
  - **Blocked By**: None

  **References**:
  - `https://vite.dev/guide/#scaffolding-your-first-vite-project` — Vite 项目创建
  - `https://tailwindcss.com/docs/installation/using-vite` — Tailwind + Vite 配置
  - `https://element-plus.org/en-US/guide/installation.html` — Element Plus 引入

  **Acceptance Criteria**:
  - [ ] `vue/package.json` 包含 vue, vue-router, pinia, axios, tailwindcss, element-plus
  - [ ] `npm run dev` 启动 Vite 开发服务器，`curl localhost:5173` 返回 HTML
  - [ ] Vue Router 六个路由已注册
  - [ ] Tailwind 自定义颜色生效（可在页面中使用 `bg-bazi-red` 等类）

  **QA Scenarios**:
  ```
  Scenario: Vue dev server starts and serves pages
    Tool: Bash
    Preconditions: Node.js 18+ installed
    Steps:
      1. cd vue && npm install
      2. npm run dev &
      3. sleep 5
      4. curl -s http://localhost:5173 | head -20
      5. kill %1
    Expected Result: curl 返回包含 `<div id="app">` 的 HTML
    Failure Indicators: 端口占用、依赖安装失败、空白页面
    Evidence: .sisyphus/evidence/task-2-dev-server.txt
  ```

  **Commit**: YES
  - Message: `chore(frontend): init Vue 3 + Vite + Tailwind + Pinia project`
  - Files: `vue/` 所有脚手架文件

- [x] 3. 数据库 Schema + GORM 模型 + 迁移脚本

  **What to do**:
  - 创建 5 张表的 GORM 模型：
    - `users`（id, username, email, password_hash, created_at, updated_at）
    - `birth_charts`（id, user_id FK, name, gender, birth_year, birth_month, birth_day, birth_hour, birth_minute, calendar_type, year_pillar, month_pillar, day_pillar, hour_pillar, five_elements JSON, ten_gods JSON, na_yin, da_yun_start JSON, created_at）
    - `fortune_queries`（id, user_id FK, chart_id FK, query_date, day_pillar, daily_fortune JSON, created_at）
    - `auspicious_rules`（id, category, day_stem, day_branch, content JSON, created_at）— 宜忌/吉时/财神/冲煞/幸运色/幸运数字
    - `activity_catalog`（id, category, name, created_at）— 宜忌活动目录
  - 创建 `src/migrations/001_init.sql`（完整建表 SQL）
  - 创建 `src/migrations/002_seed_auspicious.sql`（初始化黄历规则数据种子）
  - GORM AutoMigrate 在 main.go 启动时执行
  - MySQL Docker 容器配置（`docker-compose.yml` 中的 db 服务）

  **Must NOT do**:
  - 不写 API handler 代码
  - 不创建超出以上 5 张表的内容

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 数据库模型定义 + SQL 迁移，结构清晰，无需深度分析
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-2, 4-7)
  - **Blocks**: Tasks 8, 14, 15, 16
  - **Blocked By**: None

  **References**:
  - `https://gorm.io/docs/models.html` — GORM 模型定义
  - `https://gorm.io/docs/migration.html` — AutoMigrate 用法

  **Acceptance Criteria**:
  - [ ] 5 个 GORM 模型文件在 `src/internal/model/` 中
  - [ ] `migrations/001_init.sql` 可直接在 MySQL 中执行建表
  - [ ] GORM AutoMigrate 在应用启动时执行无报错

  **QA Scenarios**:
  ```
  Scenario: MySQL tables created by migration
    Tool: Bash (mysql client)
    Preconditions: MySQL 容器运行中
    Steps:
      1. docker-compose up -d db
      2. sleep 10
      3. docker-compose exec db mysql -u root -proot123 bazi -e "SHOW TABLES;"
    Expected Result: 输出列出 users, birth_charts, fortune_queries, auspicious_rules, activity_catalog
    Failure Indicators: 连接失败、表不存在、SQL 语法错误
    Evidence: .sisyphus/evidence/task-3-tables.txt
  ```

  **Commit**: YES
  - Message: `feat(db): add GORM models and migration for all 5 tables`
  - Files: `src/internal/model/*.go`, `src/migrations/*.sql`

- [x] 4. Docker Compose 编排配置

  **What to do**:
  - 创建根目录 `docker-compose.yml`，定义 3 个服务：
    - `db`：MySQL 8.0，端口 3306，数据卷持久化，环境变量（MYSQL_ROOT_PASSWORD, MYSQL_DATABASE）
    - `backend`：Go 后端，端口 8080，依赖 db，环境变量（DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, JWT_SECRET）
    - `frontend`：Nginx + Vue 静态文件，端口 80，依赖 backend
  - 创建 `.env.example` 放默认环境变量
  - 创建 `mysql/init/` 目录用于初始化脚本

  **Must NOT do**:
  - 不配置 Kubernetes / 云原生部署
  - 不添加监控/日志收集服务

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Docker Compose 配置标准化，纯 YAML 编写
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-3, 5-7)
  - **Blocks**: Task 27
  - **Blocked By**: None

  **References**:
  - `https://docs.docker.com/compose/compose-file/` — Docker Compose 文件格式参考

  **Acceptance Criteria**:
  - [ ] `docker-compose.yml` 定义 3 个服务，含健康检查
  - [ ] `.env.example` 含所有必需环境变量
  - [ ] `docker-compose up -d db` 可单独启动数据库

  **QA Scenarios**:
  ```
  Scenario: Docker Compose starts all services
    Tool: Bash
    Preconditions: Docker + docker-compose 已安装
    Steps:
      1. docker-compose up -d
      2. sleep 15
      3. docker-compose ps
    Expected Result: 3 个服务状态均为 "Up" (healthy)
    Failure Indicators: 容器启动失败、健康检查失败、端口冲突
    Evidence: .sisyphus/evidence/task-4-compose-ps.txt
  ```

  **Commit**: YES
  - Message: `chore(deploy): add docker-compose with Go+Vue+MySQL services`
  - Files: `docker-compose.yml`, `.env.example`

- [x] 5. 共享 API 类型 / DTO 定义

  **What to do**:
  - 创建 `src/internal/model/dto.go`，定义所有 API 请求/响应结构体：
    - `LoginRequest / LoginResponse`（含 JWT token）
    - `RegisterRequest / RegisterResponse`
    - `ChartRequest`（含 birth_year, birth_month, birth_day, birth_hour, birth_minute, calendar_type[SOLAR/LUNAR/BAZI], gender, name）
    - `ChartResponse`（含四柱、五行、十神、纳音、大运起始等信息）
    - `FortuneRequest`（含 chart_id, query_date 或默认今天）
    - `FortuneResponse`（含 day_pillar, auspicious JSON, inauspicious JSON, lucky_color, lucky_number, auspicious_hours, clash_zodiac, wealth_direction, **element_images[]**）
    - **`WeeklyFortuneRequest / WeeklyFortuneResponse`**（含 week_start, daily_fortunes[]每日7条, overall_summary, weekly_score, **element_trend[]五行占比趋势**）
    - **`MonthlyFortuneRequest / MonthlyFortuneResponse`**（含 month, daily_fortunes[]每日~30条, overall_summary, monthly_score, **element_trend[]**）
    - **`AIFortuneStubResponse`**（仅含 status:"coming_soon"）
    - **`ElementImage`**（含 element 五行名, image_url 占位URL, description）
    - `HistoryResponse`（含分页列表）
    - 通用 `ErrorResponse`
  - 创建 `src/internal/model/fortune_output.go`：黄历运势输出的完整数据结构（YiJiItem, LuckyInfo, AuspiciousHour 等）

  **Must NOT do**:
  - 不写 handler / service 代码
  - 不定义数据库无关的类型

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 纯数据结构定义，无业务逻辑
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-4, 6-7)
  - **Blocks**: Tasks 8-17（所有 handler/service 依赖类型定义）
  - **Blocked By**: None

  **References**:
  - tyme4go 的 `EightChar`, `Solar`, `Lunar`, `DaYun` 类型 — 理解字段映射
  - `https://github.com/6tail/tyme4go` — 查看 `EightChar` 结构体含哪些字段

  **Acceptance Criteria**:
  - [ ] `dto.go` 包含所有 8 个请求/响应结构体
  - [ ] `fortune_output.go` 包含黄历运势的完整输出结构
  - [ ] 所有结构体有 json tag

  **QA Scenarios**:
  ```
  Scenario: DTO types compile cleanly
    Tool: Bash
    Preconditions: Go module 已初始化
    Steps:
      1. cd src && go build ./internal/model/
    Expected Result: 编译成功，无错误
    Failure Indicators: 编译错误、循环引用
    Evidence: .sisyphus/evidence/task-5-build.txt
  ```

  **Commit**: YES
  - Message: `feat(model): define all API request/response DTOs`
  - Files: `src/internal/model/dto.go`, `src/internal/model/fortune_output.go`

- [x] 6. Go 配置模块

  **What to do**:
  - 创建 `src/internal/config/config.go`：
    - 从环境变量读取配置（DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, JWT_SECRET, SERVER_PORT）
    - 提供默认值（SERVER_PORT=8080, DB_PORT=3306）
    - 提供 `Load()` 函数返回 `Config` 结构体
    - JWT_SECRET 必需项检查（启动时为空则 panic）
  - 创建 `src/internal/config/config_test.go`（TDD：先写测试）

  **Must NOT do**:
  - 不使用 viper 等重型配置库（简单 env 读取足够）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 简单环境变量读取，单文件模块
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-5, 7)
  - **Blocks**: Tasks 8, 13, 17
  - **Blocked By**: None

  **References**:
  - `os.Getenv()` — Go 标准库环境变量读取

  **Acceptance Criteria**:
  - [ ] `config_test.go`: 3 个测试（默认值/环境变量覆盖/JWT_SECRET 缺失 panic）
  - [ ] `go test ./internal/config/` → PASS

  **QA Scenarios**:
  ```
  Scenario: Config loads defaults and env overrides correctly
    Tool: Bash
    Preconditions: 无相关环境变量
    Steps:
      1. cd src && go test -v ./internal/config/ -run TestLoad
    Expected Result: 3 tests PASS
    Failure Indicators: 测试失败
    Evidence: .sisyphus/evidence/task-6-test-output.txt

  Scenario: JWT_SECRET empty causes panic
    Tool: Bash
    Preconditions: JWT_SECRET 未设置
    Steps:
      1. cd src && JWT_SECRET="" go test -v ./internal/config/ -run TestMissingJWTSecret
    Expected Result: Test confirms panic on missing JWT_SECRET
    Failure Indicators: 测试未捕获 panic
    Evidence: .sisyphus/evidence/task-6-panic-test.txt
  ```

  **Commit**: YES
  - Message: `feat(config): add env-based config loader with TDD`
  - Files: `src/internal/config/config.go`, `src/internal/config/config_test.go`

- [x] 7. Nginx 反向代理配置模板

  **What to do**:
  - 创建 `nginx/default.conf`：
    - `/api/` → 反向代理到 Go 后端 `backend:8080`
    - `/` → 静态文件服务（Vue 构建产物 `/usr/share/nginx/html`）
    - SPA fallback：所有非 `/api/` 路径 fallback 到 `index.html`
    - CORS 头设置
  - 创建 `vue/nginx.conf`（供前端 Dockerfile 使用）

  **Must NOT do**:
  - 不配置 SSL/HTTPS（第一版仅 HTTP）
  - 不配置 CDN / 缓存策略

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 Nginx 配置，模板化
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-6)
  - **Blocks**: Tasks 25, 26, 27
  - **Blocked By**: None

  **References**:
  - `https://nginx.org/en/docs/http/ngx_http_proxy_module.html` — Nginx 反向代理配置
  - `https://router.vuejs.org/guide/essentials/history-mode.html` — Vue Router history mode 需要 Nginx fallback

  **Acceptance Criteria**:
  - [ ] `nginx/default.conf` 含 `/api/` 代理 + SPA fallback
  - [ ] `vue/nginx.conf` 含静态服务 + fallback 配置
  - [ ] Nginx 配置语法无错误

  **QA Scenarios**:
  ```
  Scenario: Nginx config syntax check
    Tool: Bash
    Preconditions: Nginx 可用（或 Docker）
    Steps:
      1. docker run --rm -v $(pwd)/nginx/default.conf:/etc/nginx/conf.d/default.conf:ro nginx:alpine nginx -t
    Expected Result: "syntax is ok" 且 "test is successful"
    Failure Indicators: 语法错误
    Evidence: .sisyphus/evidence/task-7-nginx-test.txt
  ```

  **Commit**: YES
  - Message: `chore(nginx): add reverse proxy config for Go API + Vue SPA`
  - Files: `nginx/default.conf`, `vue/nginx.conf`

---

### Wave 2 — 后端核心模块（6 tasks，最大并行）

- [x] 8. User 模型 + JWT 中间件（TDD）

  **What to do**:
  - 实现 `src/internal/model/user.go`：GORM User 模型（bcrypt 密码哈希）
  - 实现 `src/internal/middleware/auth.go`：JWT 认证中间件
    - `GenerateToken(userID, username)` → JWT string
    - `AuthMiddleware()` → Gin middleware，验证 Bearer token，注入 userID 到 context
    - Token 过期时间：24h
  - TDD：先写 `middleware/auth_test.go`（测试 token 生成/验证/过期/缺失/无效）
  - `go get golang.org/x/crypto/bcrypt golang-jwt/jwt/v5`

  **Must NOT do**:
  - 不写 handler 代码
  - 不用 Redis 做 token 黑名单

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 JWT + bcrypt 模式
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 9-13)
  - **Blocks**: Tasks 13, 16, 17
  - **Blocked By**: Tasks 1, 6

  **Acceptance Criteria**:
  - [ ] TDD：4 tests PASS（token 生成/验证/过期/缺失）
  - [ ] `go test ./internal/middleware/` → ALL PASS

  **QA Scenarios**:
  ```
  Scenario: JWT token lifecycle tests
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/middleware/ -run TestGenerateToken
      2. cd src && go test -v ./internal/middleware/ -run TestValidateToken
      3. cd src && go test -v ./internal/middleware/ -run TestExpiredToken
      4. cd src && go test -v ./internal/middleware/ -run TestMissingToken
    Expected Result: 4/4 PASS
    Evidence: .sisyphus/evidence/task-8-jwt-tests.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): add User model and JWT middleware with TDD`
  - Files: `src/internal/model/user.go`, `src/internal/middleware/auth.go`, `*_test.go`

- [x] 9. tyme4go 八字计算服务（TDD）

  **What to do**:
  - 实现 `src/internal/service/bazi.go`：`BaziService` 结构体
    - `Calculate(solarYear, solarMonth, solarDay, hour, minute, gender int) → *BaziResult`
    - 调用 tyme4go API：`Solar.fromYmdHms()` → `.getLunar()` → `.getEightChar()`
    - 提取四柱：年/月/日/时柱（天干 + 地支）
    - 提取五行、十神、纳音、藏干
    - **五行分数量化**：借鉴 china-testing/bazi 算法，计算金/木/水/火/土各自分数（天干 5 分/字 + 地支藏干加权）
    - **冲刑合会检测**：年/月/日/时柱之间的六冲/三刑/六合/三会/三合/六害关系
    - 提取大运（DaYun）起运信息（gender 参数决定顺逆）
  - TDD：3 个测试用例校验已知八字输出
    - 1990-01-15 08:00 男（根据 tyme4go 实际输出校准断言）
    - 2000-06-01 12:00 女
    - 1985-03-20 16:30 男（含大运）

  **Must NOT do**:
  - 不在 service 中处理 HTTP
  - 不实现运势解读（Task 12）

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 需研究 tyme4go API、八字算法、类型映射
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 8, 10-13)
  - **Blocks**: Tasks 12, 14
  - **Blocked By**: Tasks 1, 5

  **References**:
  - `https://github.com/6tail/tyme4go` — Solar/Lunar/EightChar/DaYun API
  - tyme4go 关键 API：`Solar.fromYmdHms()`, `Lunar.getEightChar()`, `EightChar.getYear()/getMonth()/getDay()/getTime()`

  **Acceptance Criteria**:
  - [ ] TDD：3 个生辰测试 PASS
  - [ ] 八字结果含：四柱（天干+地支）、五行 map、十神 map、纳音、大运起始

  **QA Scenarios**:
  ```
  Scenario: Known birth chart calculated correctly
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/service/ -run TestBaziCalculate
      2. cd src && go test -v ./internal/service/ -run TestBaziDaYunGender
    Expected Result: 四柱输出正确，男女大运不同
    Evidence: .sisyphus/evidence/task-9-bazi-test.txt
  ```

  **Commit**: YES
  - Message: `feat(bazi): add tyme4go Bazi calculation service with TDD`
  - Files: `src/internal/service/bazi.go`, `src/internal/service/bazi_test.go`

- [x] 10. 输入解析服务（TDD）

  **What to do**:
  - 实现 `src/internal/service/parser.go`：`InputParser` 结构体
    - `Parse(input ParseRequest) → ParsedBirth`：自动检测输入类型
    - 公历：YYYYMMDD 或 YYYY-MM-DD 格式 → 直接使用
    - 农历：含农历标识 → 调用 tyme4go `Lunar.fromYmd()` → `.getSolar()` 转换
    - 八字：含天干地支字符 → 识别为直接八字输入
    - 无效输入 → 返回描述性 error
  - TDD：4 个测试（公历/农历/八字/无效）

  **Must NOT do**:
  - 不实现 NLP 自然语言解析
  - 不在 parser 中做八字计算

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 格式匹配解析器
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 14
  - **Blocked By**: Task 5

  **Acceptance Criteria**:
  - [ ] TDD：4 tests PASS
  - [ ] 无效输入返回描述性错误

  **QA Scenarios**:
  ```
  Scenario: Input formats parsed correctly
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/service/ -run TestParse
    Expected Result: 4/4 tests PASS
    Evidence: .sisyphus/evidence/task-10-parser-test.txt
  ```

  **Commit**: YES
  - Message: `feat(parser): add multi-format birth input parser with TDD`
  - Files: `src/internal/service/parser.go`, `src/internal/service/parser_test.go`

- [x] 11. 黄历运势查找数据 + 种子脚本

  **What to do**:
  - 创建 `src/internal/service/auspicious_data.go`：硬编码黄历规则数据
    - 幸运色映射（基于日干五行：甲乙木→绿色、丙丁火→红色...）
    - 幸运数字映射（基于五行生数：木→3/8、火→2/7...）
    - 财神方位（每日固定规则：甲乙→东北、丙丁→正西...）
    - 冲煞（日支对冲生肖）
    - 吉时（每日 12 时辰择吉规则）
  - 创建 `src/internal/model/auspicious.go`：AuspiciousData 结构体
  - 创建 `src/migrations/002_seed_auspicious.sql`：黄历基础规则 INSERT 语句
  - 创建 `src/migrations/003_seed_activities.sql`：宜忌活动目录（嫁娶、出行、开市...）
  - TDD：`service/auspicious_test.go`

  **Must NOT do**:
  - 不依赖外部 API 获取黄历数据（全部内嵌）
  - 不覆盖所有 60 甲子日（至少覆盖 10 个常用日干日支）

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
    - Reason: 需要中国文化知识（黄历规则），数据量大且有规律
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 12
  - **Blocked By**: Task 3

  **Acceptance Criteria**:
  - [ ] 幸运色/幸运数字/财神方位/冲煞规则覆盖全部 10 天干
  - [ ] SQL 种子可正常执行
  - [ ] `go test ./internal/service/ -run TestAuspicious` → PASS

  **QA Scenarios**:
  ```
  Scenario: Lucky color returned for each day stem
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/service/ -run TestLuckyColorByStem
    Expected Result: 甲乙→绿青色系, 丙丁→红紫色系, etc.
    Evidence: .sisyphus/evidence/task-11-lucky-color.txt

  Scenario: Seed SQL executes without errors
    Tool: Bash (mysql client)
    Steps:
      1. docker-compose exec db mysql -u root -proot123 bazi < src/migrations/002_seed_auspicious.sql
      2. docker-compose exec db mysql -u root -proot123 bazi -e "SELECT COUNT(*) FROM auspicious_rules;"
    Expected Result: COUNT(*) > 0
    Evidence: .sisyphus/evidence/task-11-seed-count.txt
  ```

  **Commit**: YES
  - Message: `feat(fortune): add auspicious data rules and seed scripts`
  - Files: `src/internal/service/auspicious_data.go`, `src/migrations/002_*.sql`, `src/migrations/003_*.sql`

- [x] 12. 运势计算引擎（TDD）

  **What to do**:
  - 实现 `src/internal/service/fortune.go`：`FortuneEngine` 结构体
    - `CalculateDaily(userChart *BaziResult, queryDate Solar) → *DailyFortune`
    - **`CalculateWeekly(userChart *BaziResult, weekStart Solar) → *WeeklyFortune`**
      - 循环 7 天：每天调用 CalculateDaily 获取运势
      - 聚合：weekly_score（7天平均）、五行占比趋势（7天五行分布变化）
      - 生成 overall_summary（本周整体宜忌建议）
    - **`CalculateMonthly(userChart *BaziResult, year, month int) → *MonthlyFortune`**
      - 循环当月天数：每天调用 CalculateDaily 获取运势
      - 聚合：monthly_score（月平均）、五行占比趋势
      - 生成 overall_summary
    - 运势评分算法（0-100）：基于本命日干 vs 今日日干的生克强度（生+50分，合+30分，克-20分，冲-40分，基准50分）
    - 五行占比计算：统计日柱+时柱+流日的五行分布
  - TDD：`service/fortune_test.go`（至少 5 个测试用例：daily + weekly + monthly + score + element_trend）

  **Must NOT do**:
  - 不使用 AI 生成运势文字
  - 不调用外部 API

  **Recommended Agent Profile**:
  - **Category**: `deep`
    - Reason: 需要理解五行生克规则、六合三合六冲六害、黄历择吉逻辑
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 15
  - **Blocked By**: Tasks 9, 11

  **References**:
  - tyme4go `Solar`, `EightChar`, 五行生克（木生火、火生土... 金克木、木克土...）
  - 地支关系：六合（子丑合...）、六冲（子午冲...）

  **Acceptance Criteria**:
  - [ ] TDD：至少 2 tests PASS
  - [ ] DailyFortune 含：今日日柱、幸运色、幸运数字、财神方位、冲煞、吉时列表、宜忌活动列表
  - [ ] 生克分析逻辑正确（有测试验证）

  **QA Scenarios**:
  ```
  Scenario: Daily fortune calculated with all fields
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/service/ -run TestDailyFortune
    Expected Result: DailyFortune 所有字段非空，幸运色为有效色值
    Evidence: .sisyphus/evidence/task-12-fortune-test.txt

  Scenario: Sheng-Ke analysis produces correct activities
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/service/ -run TestShengKeActivities
    Expected Result: 日干相生时宜忌与相克时不同
    Evidence: .sisyphus/evidence/task-12-shengke-test.txt
  ```

  **Commit**: YES
  - Message: `feat(fortune): add daily fortune computation engine with TDD`
  - Files: `src/internal/service/fortune.go`, `src/internal/service/fortune_test.go`

- [x] 13. 用户认证 API Handlers（TDD）

  **What to do**:
  - 实现 `src/internal/handler/auth.go`：`AuthHandler` 结构体
    - `POST /api/auth/register`：接收 RegisterRequest → 创建 User → 返回 JWT token
    - `POST /api/auth/login`：接收 LoginRequest → 验证密码 → 返回 JWT token
    - `GET /api/auth/me`：需要 JWT 认证 → 返回当前用户信息
  - TDD：`handler/auth_test.go`（使用 `httptest` 模拟 Gin context）
    - 测试注册成功
    - 测试注册重复用户名
    - 测试登录成功
    - 测试登录密码错误
    - 测试未认证访问 /me

  **Must NOT do**:
  - 不在此任务中整合数据库（使用 mock / sqlite 内存库测试）

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: 标准 CRUD handler
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2
  - **Blocks**: Task 17
  - **Blocked By**: Tasks 8, 5

  **References**:
  - Gin `httptest` 包文档
  - `https://gin-gonic.com/docs/testing/` — Gin 测试指南

  **Acceptance Criteria**:
  - [ ] TDD：5 tests PASS
  - [ ] 注册返回 JWT token，登录验证正确
  - [ ] 密码错误/重复用户名返回 4xx

  **QA Scenarios**:
  ```
  Scenario: Register and login flow
    Tool: Bash (go test)
    Steps:
      1. cd src && go test -v ./internal/handler/ -run TestAuth
    Expected Result: 5/5 PASS
    Failure Indicators: 测试失败
    Evidence: .sisyphus/evidence/task-13-auth-test.txt
  ```

  **Commit**: YES
  - Message: `feat(auth): add register/login/me API handlers with TDD`
  - Files: `src/internal/handler/auth.go`, `src/internal/handler/auth_test.go`

- [x] 35. 紫微斗数计算服务（加强版）（TDD）

  **What to do**:
  - 安装依赖：`go get github.com/kaecer68/ziwei-zenith`
  - 实现 `src/internal/service/ziwei.go`：`ZiWeiService` 结构体
    - `CalculateChart(...) → *ZiWeiChart`：基础排盘
      - 12 宫（每宫含：宫名 + 14 主星 + 辅星 + **星曜亮度[庙旺得利平不陷]** + 四化标注 + 三方四正）
      - **身宫 + 命主/身主 + 五行局**
    - **`DetectPatterns(chart) → []Pattern`**：格局检测（如紫府同宫格、君臣庆会格、机月同梁格等 ≥20 种常见格局）
    - **`AnalyzeFlyingStars(chart) → FlyingStarAnalysis`**：四化飞星分析（化禄/权/科/忌分别飞入何宫、对该宫的影响规则）
    - **`CalculateLiuyue(chart, year, month) → *ZiWeiPeriod`**：流月计算
    - **`CalculateLiuri(chart, year, month, day) → *ZiWeiPeriod`**：流日计算
    - **`GetOverlayData(chart, liunianYear) → *OverlayData`**：流年叠盘数据（基础命盘 + 流年星曜变化）
    - 复用 ziwei-zenith 已有的亮度/格局/神煞模块，映射到我们的 DTO
  - TDD：`service/ziwei_test.go`（≥5 测试：chart/patterns/flyingStars/liuyue/liuri/overlay）

  **Must NOT do**:
  - 不修改 ziwei-zenith 源码
  - 不在 service 中处理 HTTP

  **Recommended Agent Profile**: `deep` | **Skills**: []
  **Parallelization**: Wave 2 | **Blocks**: Tasks 36, 37, 40 | **Blocked By**: Tasks 1, 10

  **References**:
  - `https://github.com/kaecer68/ziwei-zenith` — 核心库 API（pkg/basis 定义层, pkg/engine 计算层）
  - ziwei-zenith 已支持：brightness(亮度), pattern(格局), dayun(大限), liunian(流年), god stars(神煞)
  - 格局列表：紫府同宫, 君臣庆会, 机月同梁, 月朗天门, 日照雷门, 雄宿乾元, 月生沧海, 明珠出海, 石中隐玉, 巨日同宫, 廉贞清白, 刑囚夹印, 马头带箭...≥20

  **Acceptance Criteria**:
  - [ ] TDD：≥5 tests PASS
  - [ ] 每宫主星含 brightness 字段（7 级枚举）
  - [ ] DetectPatterns 返回 ≥2 个匹配格局（若生辰有格局）
  - [ ] 四化飞星含 source（化星）+ target（飞入宫）+ effect（影响规则）

  **QA Scenarios**:
  ```
  Scenario: ZiWei chart with brightness and patterns
    Tool: Bash (go test)
    Steps: cd src && go test -v ./internal/service/ -run TestZiWeiFull
    Expected Result: 12 宫含亮度 + ≥1 格局 + 四化飞星分析
    Evidence: .sisyphus/evidence/task-35-ziwei-full.txt
  ```

  **Commit**: YES | `feat(ziwei): add comprehensive ZiWei service with brightness/patterns/flying-stars`
  Files: `src/internal/service/ziwei.go`, `*_test.go`

- [x] 40. 紫微斗数宫位解读文案引擎（TDD）

  **What to do**:
  - 实现 `src/internal/service/ziwei_interpretation.go`：`ZiWeiInterpreter` 结构体
    - `InterpretPalace(palace *ZiWeiPalace, chart *ZiWeiChart) → *PalaceReading`
    - 基于规则模板 + **bazi-skill 典籍参考数据**生成结构化解读
    - 数据来源：
      - ziwei-zenith 内置规则（格局/亮度/四化）
      - **`classical-texts.md`**：穷通宝典/三命通会/滴天髓/渊海子平/千里命稿/协纪辨方书/果老星宗/子平真诠/神峰通考 — 9 本典籍论命规则
      - **`wuxing-tables.md`**：五行/天干地支/十神/藏干完整参考表
      - **`dayun-rules.md`**：大运顺逆排规则
    - 模板存放 `src/internal/service/ziwei_templates.go`：
      - 14 主星 × 7 亮度 = 98 条模板
      - 14 辅星组合规则
      - 4 化 × 12 宫 = 48 条模板
      - 20+ 格局描述模板
      - **扩展**：引用 bazi-skill 典籍规则补充深度解读
  - TDD：`service/ziwei_interpretation_test.go`

  **References**:
  - `https://github.com/jinchenma94/bazi-skill/tree/main/references` — 9 本典籍结构化论命规则数据
  - `https://github.com/kaecer68/ziwei-zenith` — 紫微斗数计算引擎

  **Recommended Agent Profile**: `deep` | **Skills**: []
  **Parallelization**: Wave 2 | **Blocks**: Task 36 | **Blocked By**: Task 35

  **Commit**: YES | `feat(ziwei): add interpretation engine with classical text rules`
  Files: `src/internal/service/ziwei_interpretation.go`, `ziwei_templates.go`, `*_test.go`

---

### Wave 3 — 后端 API 端点（10 tasks）

- [x] 14. 八字排盘 Handler（TDD）
- [x] 15. 今日运势 Handler（TDD）
- [x] 16. 历史记录 Handler（TDD）
- [x] 28. 周运势 Handler（TDD）
- [x] 37. 紫微斗数流年/流月/流日 + 叠盘 Handler（TDD）

  **What to do**:
  - 实现 `src/internal/handler/ziwei_period.go`：`ZiWeiPeriodHandler`
    - `POST /api/ziwei/period`：period_type=dayun → 10 年大限列表
    - `POST /api/ziwei/period`：period_type=liunian → 当年流年命盘 + 流年四化
    - **`POST /api/ziwei/period`：period_type=liuyue → 当月流月分析**
    - **`POST /api/ziwei/period`：period_type=liuri → 当日流日分析**
    - **`POST /api/ziwei/overlay`：chart_id + year → 基础命盘 + 流年星曜叠盘数据**
      - 返回两套星曜数据：base_stars + liunian_stars，前端可切换/叠加显示
  - TDD：`handler/ziwei_period_test.go`

  **Recommended Agent Profile**: `quick` | **Skills**: []
  **Parallelization**: Wave 3 | **Blocks**: Task 17 | **Blocked By**: Task 35

  **QA Scenarios**:
  ```
  Scenario: All period types return correct data
    Tool: Bash (go test)
    Steps: cd src && go test -v ./internal/handler/ -run TestZiWeiPeriod
    Expected Result: dayun/liunian/liuyue/liuri/overlay 全覆盖
    Evidence: .sisyphus/evidence/task-37-ziwei-period.txt
  ```

  **Commit**: YES | `feat(ziwei): add period/liuyue/liuri/overlay API handlers`
  Files: `src/internal/handler/ziwei_period.go`, `*_test.go`

- [x] 17. 路由注册 + 中间件整合

  **What to do**:
  - 更新 `src/cmd/main.go`：
    - 加载 Config
    - 初始化 GORM（连接 MySQL）
    - AutoMigrate 所有模型
    - 注册路由：`/health`（公开）、`/api/auth/*`（公开）、`/api/chart`（JWT）、`/api/fortune`（JWT）、`/api/fortune/weekly`（JWT）、`/api/fortune/monthly`（JWT）、`/api/fortune/ai`（JWT）、`/api/ziwei/chart`（JWT）、`/api/ziwei/period`（JWT）、`/api/charts`（JWT）
    - CORS 中间件（允许前端跨域）
    - 启动 Gin server
  - 确保 `go build ./cmd/` 编译通过
  - TDD：`cmd/main_test.go` 集成测试（启动 server → 测试 health + auth 端点）

  **Must NOT do**:
  - 不写业务逻辑

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (last in wave, depends on 13-16)
  - **Blocks**: Task 25
  - **Blocked By**: Tasks 13, 14, 15, 16

  **Acceptance Criteria**:
  - [ ] `go build ./cmd/` 编译通过
  - [ ] 所有路由正确注册
  - [ ] JWT 中间件保护 /api/chart, /api/fortune, /api/charts

  **QA Scenarios**:
  ```
  Scenario: API server starts and routes work
    Tool: Bash (curl)
    Steps:
      1. cd src && go run ./cmd/ &
      2. sleep 3
      3. curl -s http://localhost:8080/health
      4. curl -s http://localhost:8080/api/chart (expect 401)
      5. kill %1
    Expected Result: /health → 200, /api/chart → 401
    Evidence: .sisyphus/evidence/task-17-routes.txt
  ```

  **Commit**: YES
  - Message: `feat(router): wire all handlers with JWT middleware and CORS`
  - Files: `src/cmd/main.go`, `src/cmd/main_test.go`

- [x] 18. Go Dockerfile 完善 + 构建验证

  **What to do**:
  - 完善 `src/Dockerfile`：多阶段构建
    - Stage 1：`golang:1.22-alpine`，`go mod download` + `go build -o /app/server ./cmd/`
    - Stage 2：`alpine:3.19`，复制二进制 + 迁移脚本
  - 在 `docker-compose.yml` 中完善 backend 服务配置（环境变量、健康检查、依赖 db）
  - 验证 `docker-compose build backend` 成功

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: []

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3
  - **Blocks**: Task 27
  - **Blocked By**: Task 1

  **Acceptance Criteria**:
  - [ ] `docker-compose build backend` 无错误
  - [ ] 健康检查：`curl localhost:8080/health` → 200

  **QA Scenarios**:
  ```
  Scenario: Backend Docker image builds and starts
    Tool: Bash
    Steps:
      1. docker-compose build backend
      2. docker-compose up -d backend
      3. sleep 5
      4. curl -s http://localhost:8080/health
    Expected Result: {"status":"ok"}
    Evidence: .sisyphus/evidence/task-18-docker-backend.txt
  ```

  **Commit**: YES
  - Message: `chore(docker): finalize Go multi-stage Dockerfile`
  - Files: `src/Dockerfile`, `docker-compose.yml`

---

### Wave 4 — 前端核心页面（6 tasks，最大并行）

- [x] 19. 登录/注册页面 + Auth Store（TDD）

  **What to do**:
  - 创建 `vue/src/stores/auth.ts`：Pinia store（login/register/logout/fetchMe，token 持久化到 localStorage）
  - 创建 `vue/src/views/LoginView.vue`：Element Plus 表单，新中式配色（朱砂红按钮、黛蓝卡片、宣纸白底色），链接到注册
  - 创建 `vue/src/views/RegisterView.vue`：用户名+邮箱+密码+确认密码表单
  - TDD：`stores/auth.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **Acceptance Criteria**:
  - [ ] TDD：store 4 个方法测试 PASS
  - [ ] 登录成功跳转到首页

  **QA Scenarios**:
  ```
  Scenario: Register and login flow via Playwright
    Tool: Playwright
    Steps: goto /register → fill form → submit → goto /login → fill → submit → waitForURL /
    Expected Result: 最终 URL 为 /，页面含"欢迎"
    Evidence: .sisyphus/evidence/task-19-auth-flow.png
  ```

  **Commit**: YES | `feat(auth-ui): add Login/Register pages with Pinia store`
  Files: `vue/src/stores/auth.ts`, `vue/src/views/LoginView.vue`, `vue/src/views/RegisterView.vue`

- [x] 20. 首页 + 输入表单组件（TDD）

  **What to do**:
  - `vue/src/views/HomeView.vue`：欢迎信息 + "排盘"入口
  - `vue/src/components/BirthInputForm.vue`：三模式输入（公历/农历/八字 Tab 切换）+ 时辰下拉 + 性别 radio + 表单验证 + 新中式 UI
  - TDD：`components/BirthInputForm.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Solar birth form submits and navigates
    Tool: Playwright
    Steps: goto / → click start → tab solar → fill 1990/1/15/辰时/男 → submit → waitForURL /chart/
    Evidence: .sisyphus/evidence/task-20-form-submit.png
  ```

  **Commit**: YES | `feat(home): add Home and BirthInputForm with multi-mode input`
  Files: `vue/src/views/HomeView.vue`, `vue/src/components/BirthInputForm.vue`

- [x] 21. 八字命盘展示组件（TDD）

  **What to do**:
  - `vue/src/views/ChartView.vue`：加载命盘数据
  - `vue/src/components/BaziChart.vue`：四柱表格（天干+地支+藏干）+ 五行颜色标注（木绿火红土黄金白水黑）+ 十神 + 纳音 + **冲刑合会关系可视化连线**（六冲红虚线/六合绿实线/三刑橙色）+ "今日运势"按钮
  - TDD：`components/BaziChart.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Bazi chart renders with WuXing colors
    Tool: Playwright
    Steps: goto /chart/1 → waitForSelector [data-testid="bazi-chart"] → screenshot
    Evidence: .sisyphus/evidence/task-21-chart.png
  ```

  **Commit**: YES | `feat(chart-ui): add BaziChart with WuXing color visualization`
  Files: `vue/src/views/ChartView.vue`, `vue/src/components/BaziChart.vue`

- [x] 22. 运势展示组件（TDD）

  **What to do**:
  - `vue/src/views/FortuneView.vue`：运势详情页
  - `vue/src/components/DailyFortune.vue`：今日日柱 + 幸运色卡片（色块预览）+ 幸运数字卡片 + 财神方位 + 冲煞 + 宜忌分栏 + 吉时列表 + 生克分析
  - TDD：`components/DailyFortune.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Fortune page displays all sections
    Tool: Playwright
    Steps: goto /fortune?chart_id=1 → waitForSelector .daily-fortune → screenshot
    Evidence: .sisyphus/evidence/task-22-fortune.png
  ```

  **Commit**: YES | `feat(fortune-ui): add DailyFortune with almanac card display`
  Files: `vue/src/views/FortuneView.vue`, `vue/src/components/DailyFortune.vue`

- [x] 23. 历史记录页面（TDD）

  **What to do**:
  - `vue/src/views/HistoryView.vue`：命盘列表（卡片）+ 分页 + 点击跳转详情 + 运势历史查询 + 空状态
  - TDD：`views/HistoryView.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: History page shows chart list with pagination
    Tool: Playwright
    Steps: goto /history → waitFor chart-list → click item → waitForURL /chart/
    Evidence: .sisyphus/evidence/task-23-history.png
  ```

  **Commit**: YES | `feat(history-ui): add HistoryView with paginated chart list`
  Files: `vue/src/views/HistoryView.vue`

- [x] 31. ECharts 运势折线图组件

  **What to do**:
  - 安装 ECharts：`npm install echarts vue-echarts`
  - 创建 `vue/src/components/FortuneChart.vue`：
    - Props：`dailyData[]`（含 date, score, elements{metal,wood,water,fire,earth}）
    - 多折线图：评分折线（红色，虚线标注 60 分基准线）+ 五行占比折线（金白、木绿、水蓝、火红、土黄）
    - X 轴：日期（日/月）
    - Y 轴左：运势评分 0-100
    - Y 轴右：五行占比 0-100%
    - 响应式：自适应容器宽度
    - Tooltip：悬停显示当日详细数据
    - 新中式配色（黛蓝背景、朱砂红强调色）
  - TDD：`components/FortuneChart.test.ts`（Vitest + mock echarts）

  **Must NOT do**:
  - 不在此组件中加载数据（数据通过 props 传入）

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Tasks 32, 33 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Line chart renders with score and element lines
    Tool: Playwright
    Steps: Mount component with mock 7-day data → screenshot
    Expected Result: 图表含 6 条折线（1 评分 + 5 五行），颜色正确
    Evidence: .sisyphus/evidence/task-31-chart.png
  ```

  **Commit**: YES | `feat(chart): add ECharts multi-line fortune trend chart`
  Files: `vue/src/components/FortuneChart.vue`

- [x] 32. 周运势展示页

  **What to do**:
  - 创建 `vue/src/views/WeeklyFortuneView.vue`：路由 `/fortune/weekly?chart_id=X`
  - 布局：
    - 顶部：本周日期范围 + 整体运势评分（大字）+ 运势摘要文字
    - 中部：**FortuneChart 折线图**（本周 7 天评分+五行趋势）
    - 底部：7 天运势卡片列表（每天一张小卡片：日期、日柱、幸运色色块、简洁宜忌）
  - 从 API 加载数据：`POST /api/fortune/weekly`
  - 新中式 UI 风格
  - TDD：`views/WeeklyFortuneView.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Tasks 2, 31

  **QA Scenarios**:
  ```
  Scenario: Weekly fortune page displays chart + 7 daily cards
    Tool: Playwright
    Steps: goto /fortune/weekly?chart_id=1 → waitFor [data-testid="fortune-chart"] → screenshot
    Expected Result: 折线图可见 + 7 张运势卡片
    Evidence: .sisyphus/evidence/task-32-weekly-page.png
  ```

  **Commit**: YES | `feat(weekly): add WeeklyFortuneView with chart and daily cards`
  Files: `vue/src/views/WeeklyFortuneView.vue`

- [x] 33. 月运势展示页

  **What to do**:
  - 创建 `vue/src/views/MonthlyFortuneView.vue`：路由 `/fortune/monthly?chart_id=X&year=2026&month=5`
  - 布局：同周运势，但数据为当月全部天数（~30 天折线图 + 可滚动卡片列表）
  - 折线图默认展示，卡片列表支持虚拟滚动（如超过 15 条）
  - TDD：`views/MonthlyFortuneView.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Tasks 2, 31

  **QA Scenarios**:
  ```
  Scenario: Monthly fortune page with 30-day chart
    Tool: Playwright
    Steps: goto /fortune/monthly?chart_id=1&year=2026&month=5 → waitFor chart → screenshot
    Expected Result: 折线图 + 当月全部运势卡片
    Evidence: .sisyphus/evidence/task-33-monthly-page.png
  ```

  **Commit**: YES | `feat(monthly): add MonthlyFortuneView with 30-day trend chart`
  Files: `vue/src/views/MonthlyFortuneView.vue`

- [x] 34. 五行元素图片展示 + AI 分析入口组件

  **What to do**:
  - 创建 `vue/src/components/ElementImages.vue`：
    - Props：`elements[]`（含 element 名 + image_url + description）
    - 展示五行元素占位图卡片（金/木/水/火/土 各一张）
    - 每张卡片：元素名 + 占位图片（用纯色 block 代替，标注"图片资源待补充"）+ 描述文字
    - 水平排列，响应式 grid
  - 创建 `vue/src/components/AIAnalysisEntry.vue`：
    - 一个"AI 智能分析"按钮（新中式风格，灰显 disabled 样式）
    - 点击弹出 Modal：显示"AI 分析功能即将上线，敬请期待" + 装饰图标
    - 不调用实际 API（纯前端占位）
  - 在运势详情页（DailyFortune 组件）中集成两者：元素图片放在运势卡片下方，AI 按钮放在页面底部
  - TDD：`components/ElementImages.test.ts` + `components/AIAnalysisEntry.test.ts`

  **Must NOT do**:
  - 不接入真实 AI API
  - 不使用真实图片资源（用 CSS 纯色块 + 占位文字）

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Element images display placeholder cards
    Tool: Playwright
    Steps: Mount ElementImages with mock data → screenshot
    Expected Result: 5 张元素卡片，含"图片资源待补充"标注
    Evidence: .sisyphus/evidence/task-34-elements.png

  Scenario: AI entry shows coming soon modal
    Tool: Playwright
    Steps: Click AI button → waitFor modal → assert text "即将上线"
    Expected Result: Modal 弹出，含"敬请期待"
    Evidence: .sisyphus/evidence/task-34-ai-modal.png
  ```

  **Commit**: YES | `feat(ui): add ElementImages placeholder and AI analysis entry stub`
  Files: `vue/src/components/ElementImages.vue`, `vue/src/components/AIAnalysisEntry.vue`

- [x] 38. 紫微斗数命盘图可视化组件（加强版）

  **What to do**:
  - 创建 `vue/src/components/ZiWeiChart.vue`：紫微命盘图组件
    - 传统方图布局：12 宫格按地支方位排列
    - 每宫显示：宫名 + 主星（大字）+ 辅星（小字）+ **星曜亮度背景色**（庙=红/旺=橙/得=黄/利=绿/平=灰/不=浅蓝/陷=深蓝）
    - 中央显示命宫信息（命主 + 身主 + **五行局**）
    - **格局标注**：若命盘含特殊格局，在方图外以 badge 标签展示
    - **叠盘切换**：toggle 按钮切换"本命盘"/"流年叠盘"模式
      - 流年叠盘：在原宫位叠加显示流年星曜（不同颜色/层级）
    - 四化标注：化禄(红)化权(蓝)化科(绿)化忌(黑)，CSS 动画高亮
    - 响应式 + 新中式配色
  - TDD：`components/ZiWeiChart.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: Wave 4 | **Blocks**: Tasks 39, 42 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: Chart displays brightness colors, pattern badges, and overlay toggle
    Tool: Playwright
    Steps: Mount with full data → screenshot base → toggle overlay → screenshot overlay
    Expected Result: 亮度色阶区分清晰，叠盘切换后星曜变化可见，格局 badge 显示
    Evidence: .sisyphus/evidence/task-38-ziwei-enhanced.png
  ```

  **Commit**: YES | `feat(ziwei-ui): enhance chart with brightness/patterns/overlay`
  Files: `vue/src/components/ZiWeiChart.vue`

- [x] 39. 紫微斗数命盘页面（加强版）

  **What to do**:
  - 创建 `vue/src/views/ZiWeiView.vue`：路由 `/ziwei/:chartId`
  - 布局：生辰信息 + 命盘方图（含亮度/格局/叠盘） + Tab 切换
  - Tab 1 **命盘详解**：点击宫位 → 展开 `ZiWeiInterpretation` 组件（结构化解读文案）
  - Tab 2 **大限分析**：当前大限 + 历史大限时间轴
  - Tab 3 **流年分析**：当前流年 + 叠盘对比视图
  - Tab 4 **流月分析**：当月运势分析
  - Tab 5 **流日分析**：今日运势分析
  - Tab 6 **四化飞星**：化禄/权/科/忌飞入各宫的详细表格
  - 各 Tab 加载状态 + 错误处理
  - TDD + 新中式 UI

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: Wave 4 | **Blocks**: Task 25 | **Blocked By**: Tasks 2, 38

  **QA Scenarios**:
  ```
  Scenario: All 6 tabs load and display correctly
    Tool: Playwright
    Steps: goto /ziwei/1 → click each tab → screenshot each
    Expected Result: 6 个 Tab 全部有内容
    Evidence: .sisyphus/evidence/task-39-ziwei-page.png
  ```

  **Commit**: YES | `feat(ziwei-ui): enhance ZiWei view with 6 tabs and interpretation`
  Files: `vue/src/views/ZiWeiView.vue`

- [x] 41. 紫微斗数宫位解读展示组件

  **What to do**:
  - 创建 `vue/src/components/ZiWeiInterpretation.vue`：
    - 展开式文案卡片：点击某宫 → 展示该宫的结构化解读
    - 分节展示：主星特性 | 辅星影响 | 四化影响 | 三方四正 | 格局标注
    - 每节含标题 + 描述文字 + 相关星曜标签
    - 新中式排版：宣纸白底 + 黛蓝标题 + 朱砂红强调
    - 支持展开/收起动画
  - TDD：`components/ZiWeiInterpretation.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: Wave 4 | **Blocks**: Task 25 | **Blocked By**: Tasks 2, 38

  **QA Scenarios**:
  ```
  Scenario: Interpretation card displays all 5 sections
    Tool: Playwright
    Steps: Mount with mock palace reading → assert 5 sections visible
    Expected Result: 每个 section 有标题和内容
    Evidence: .sisyphus/evidence/task-41-interpretation.png
  ```

  **Commit**: YES | `feat(ziwei-ui): add structured palace interpretation display`
  Files: `vue/src/components/ZiWeiInterpretation.vue`

- [x] 42. 紫微斗数流年叠盘交互组件

  **What to do**:
  - 创建 `vue/src/components/ZiWeiOverlay.vue`：
    - 双模式切换：本命盘 / 流年叠盘
    - 叠盘模式：本命星曜（正常色）+ 流年星曜（半透明彩色叠加层）
    - 流年四化高亮动画（化禄红闪/化权蓝闪/化科绿闪/化忌黑闪）
    - 年份选择器：可切换不同年份的流年叠盘
    - 图例：本命星曜符号 + 流年星曜符号区分
  - TDD：`components/ZiWeiOverlay.test.ts`

  **Recommended Agent Profile**: `visual-engineering` | **Skills**: []
  **Parallelization**: Wave 4 | **Blocks**: Task 25 | **Blocked By**: Tasks 2, 38

  **QA Scenarios**:
  ```
  Scenario: Overlay toggle switches between base and liunian stars
    Tool: Playwright
    Steps: Mount with base+overlay data → click toggle → assert liunian stars visible
    Expected Result: 叠盘切换流畅，流年星曜颜色区分清晰
    Evidence: .sisyphus/evidence/task-42-overlay.png
  ```

  **Commit**: YES | `feat(ziwei-ui): add interactive Liunian overlay chart component`
  Files: `vue/src/components/ZiWeiOverlay.vue`

- [x] 24. API 客户端层 + 请求拦截器

  **What to do**:
  - `vue/src/api/client.ts`：Axios 实例（baseURL /api，JWT interceptor，401 → 跳转登录）
  - `vue/src/api/auth.ts`、`vue/src/api/chart.ts`、`vue/src/api/fortune.ts`、`vue/src/api/fortune_weekly.ts`、`vue/src/api/fortune_monthly.ts`、`vue/src/api/ai.ts`
  - TDD：`api/client.test.ts`

  **Recommended Agent Profile**: `quick` | **Skills**: []
  **Parallelization**: YES（Wave 4）| **Blocks**: Task 25 | **Blocked By**: Task 2

  **QA Scenarios**:
  ```
  Scenario: API client attaches JWT and handles 401
    Tool: Bash (Vitest)
    Steps: cd vue && npx vitest run src/api/client.test.ts
    Evidence: .sisyphus/evidence/task-24-api-test.txt
  ```

  **Commit**: YES | `feat(api): add Axios client with JWT interceptor`
  Files: `vue/src/api/*.ts`

---

### Wave 5 — 集成与部署验证（3 tasks）

- [x] 25. 前后端联调 + E2E 流程

  **What to do**: 确保完整用户流程（注册→登录→排盘→运势→历史）+ 修复 CORS/路由/数据格式 + E2E Playwright 测试 + loading/error/empty 状态

  **Recommended Agent Profile**: `unspecified-high` | **Skills**: [] | **Parallelization**: NO | **Blocks**: 27 | **Blocked By**: 14-24

  **QA Scenarios**:
  ```
  Scenario: Full end-to-end user journey
    Tool: Playwright
    Steps: 注册 → 登录 → 排盘 → 查看命盘 → 查看运势 → 查看历史
    Expected Result: 全流程无报错
    Evidence: .sisyphus/evidence/task-25-e2e-full.png
  ```

  **Commit**: YES | `feat(integration): wire frontend to backend, E2E test pass`
  Files: any modified files

- [x] 26. Nginx + Vue 生产构建

  **What to do**: `vue/vite.config.ts` + `vue/Dockerfile`（node build → nginx serve）+ `docker-compose.yml` frontend 服务 + 验证构建

  **Recommended Agent Profile**: `quick` | **Skills**: [] | **Parallelization**: YES | **Blocks**: 27 | **Blocked By**: 2, 7

  **QA Scenarios**:
  ```
  Scenario: Vue production build served by Nginx
    Tool: Bash
    Steps: npm run build → docker-compose build frontend → curl localhost/
    Expected Result: HTML 含 <div id="app">
    Evidence: .sisyphus/evidence/task-26-frontend.txt
  ```

  **Commit**: YES | `chore(frontend): add production Dockerfile with Nginx`
  Files: `vue/Dockerfile`, `vue/vite.config.ts`, `vue/nginx.conf`

- [x] 27. Docker Compose 全栈集成测试

  **What to do**: `docker-compose down -v && up -d` + 等待 healthy + 验证所有端点 + 数据持久化 + `scripts/test-fullstack.sh`

  **Recommended Agent Profile**: `unspecified-high` | **Skills**: [] | **Parallelization**: NO | **Blocks**: F1 | **Blocked By**: 18, 25, 26

  **QA Scenarios**:
  ```
  Scenario: Full stack from scratch
    Tool: Bash
    Steps: docker-compose down -v && up -d → sleep 20 → ps → curl health → curl frontend
    Expected Result: 3 容器 healthy, 200 OK
    Evidence: .sisyphus/evidence/task-27-fullstack.txt
  ```

  **Commit**: YES | `chore(deploy): add full-stack integration test`
  Files: `scripts/test-fullstack.sh`

---

## Final Verification Wave

 - [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in .sisyphus/evidence/. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

 - [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `golangci-lint` + `go test ./...`. Run `vue-tsc --noEmit` + `eslint` + `vitest run`. Review all changed files for: `interface{}`/`any`, empty catches, fmt.Println in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

 - [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state (`docker-compose down -v && docker-compose up -d`). Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration. Test edge cases: empty input, invalid date, unauthenticated access.
  Save to `.sisyphus/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

 - [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (git log/diff). Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination. Flag unaccounted changes.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

See individual task commit instructions. Each task commits independently with format: `type(scope): description`.

---

## Success Criteria

### Verification Commands
```bash
# Backend
cd src && go test ./...          # Expected: ALL PASS
cd src && go vet ./...            # Expected: no errors
cd src && go build ./cmd/         # Expected: binary created

# Frontend
cd vue && npx vitest run          # Expected: ALL PASS
cd vue && npx vue-tsc --noEmit   # Expected: no errors
cd vue && npm run build           # Expected: dist/ created

# Full Stack
docker-compose up -d              # Expected: 3 containers healthy
curl http://localhost:8080/health # Expected: {"status":"ok"}
```

### Final Checklist
- [ ] All "Must Have" present
- [ ] All "Must NOT Have" absent
- [ ] All Go tests pass
- [ ] All Vue tests pass
- [ ] docker-compose up works
- [ ] Full user flow: register → login → input birth → view chart → view fortune → view history
