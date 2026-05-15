# Learnings - Bazi Fortune Frontend

## Design System (from style.css)
- Tailwind theme tokens: bazi-red (#C41E3A), bazi-blue (#2B3A42), bazi-paper (#F5F0E8), bazi-ink (#1A1A1A)
- Element Plus for UI components
- Five-element colors: 木=green, 火=red, 土=yellow, 金=amber/gold, 水=blue

## Architecture
- Pinia setup stores (composition API style) in `stores/`
- Axios client with interceptors in `api/client.ts`
- Lazy-loaded route components via `() => import(...)`
- Chinese locale throughout (labels, messages, validation)

## Router
- `/` → HomeView
- `/login` → LoginView
- `/register` → RegisterView
- `/chart/:id` → ChartView (id="new" shows BirthInputForm)
- `/fortune` → FortuneView (query param: chart_id)
- `/history` → HistoryView

## BaziChart Data Structure
- chart object has: year_pillar, month_pillar, day_pillar, hour_pillar
- Each pillar: { gan: string, zhi: string }
- Gan element lookup map for 10 heavenly stems
- Zhi element lookup map for 12 earthly branches
- Six clashes (六冲): 子午, 丑未, 寅申, 卯酉, 辰戌, 巳亥
- Six harmonies (六合): 子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未

## API Endpoints Used
- POST /api/auth/login → { token, user }
- POST /api/auth/register → { token, user }
- GET /api/auth/me → { user }
- POST /api/chart → { id, chart } (body: calendar_type, year, month, day, shichen, gender)
- GET /api/chart/:id → { chart }
- GET /api/charts → { charts[] }
- POST /api/fortune → fortune data

## Vue Component Implementation (Tasks 34, 38, 39, 41, 42)

### Design System
- Tailwind CSS v4 with `@theme` directive in `style.css`
- Color tokens: `--color-bazi-red` (#C41E3A), `--color-bazi-blue` (#2B3A42), `--color-bazi-paper` (#F5F0E8), `--color-bazi-ink` (#1A1A1A)
- Element Plus UI library integrated via `app.use(ElementPlus)`
- All components use `<script setup lang="ts">` SFC pattern
- Axios client at `@/api/client` with JWT interceptors

### Component Patterns
- Props-based data flow (no stores for components)
- Tailwind utility classes + scoped CSS for component styles
- CSS variables (`var(--color-bazi-*)`) used for design token colors
- Element Plus `ElDialog` used for modals
- Transition components used for expand/collapse animations

### ZiWei Chart Grid Layout
- Traditional square chart: 4-column CSS Grid
- Row 1: 巳 午 未 申
- Row 2: 辰 [center spans 2×2] 酉
- Row 3: 卯 [center continues] 戌
- Row 4: 寅 丑 子 亥
- Center displays 命主/身主/五行局 using `grid-row: span 2; grid-column: span 2`

### Star Brightness Color Map
- 庙=#C41E3A (bazi-red), 旺=#FF8C00, 得=#DAA520, 利=#228B22, 平=#808080, 不=#87CEEB, 陷=#191970
- Dark backgrounds (陷/利/庙) use white text; others use dark text

### API Endpoints Used
- POST /api/ziwei/chart — get ZiWei chart data
- POST /api/ziwei/period — get period data (dayun/liunian/liuyue/liuri/sihua)
- POST /api/ziwei/overlay — get liunian overlay for a specific year

## Wave 5 Integration (Tasks 25-27)

### Nginx Configuration
- Two identical nginx configs: `nginx/default.conf` (dev/reference) and `vue/nginx.conf` (Docker build)
- Both proxy `/api/` → `http://backend:8080` and serve SPA with `try_files $uri $uri/ /index.html`
- Vue Dockerfile copies `vue/nginx.conf` to `/etc/nginx/conf.d/default.conf`

### Vue Import Chain
- `main.ts` → imports `App.vue` + `router` + Pinia + ElementPlus
- `App.vue` → uses `<router-view />` for SPA routing
- `router/index.ts` → lazy-loads all 7 views (Home, Login, Register, Chart, Fortune, History, ZiWei)
- Views import components: ChartView imports BaziChart+BirthInputForm, FortuneView imports DailyFortune, ZiWeiView imports ZiWeiChart+ZiWeiInterpretation+ZiWeiOverlay
- `stores/auth.ts` imports `api/client.ts` (axios with JWT interceptors, baseURL: `/api`)

### Docker Compose
- 3 services: db (MySQL 8.0), backend (Go), frontend (Vue/Nginx)
- Healthchecks: db uses mysqladmin ping, backend uses curl health endpoint
- backend depends_on db with `condition: service_healthy`
- Named volume: mysql_data
- Shared network: bazi-network (bridge)

### Scripts
- `scripts/test-e2e.sh` — full E2E flow: health → register/login → create chart → get chart → fortune → history
- `scripts/docker-verify.sh` — structural check of docker-compose.yml (services, healthchecks, volumes, networks)
