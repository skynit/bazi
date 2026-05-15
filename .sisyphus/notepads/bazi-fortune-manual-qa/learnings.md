# Manual QA Review — bazi-fortune

## Structure
- Go: 43 source files under `src/`
- Vue/TS: 24 source files under `vue/src/`
- Total LOC (Go + Vue/TS): 10,953

## Test Coverage

| Package | Files | Tested | Untested | Notes |
|---------|-------|--------|----------|-------|
| handler/ | 9 | 9 | 0 | Full coverage ✅ |
| service/ | 7 | 5 | 2 | auspicious_data (179 lines), ziwei_templates (676 lines) — hardcoded data stores |
| config/ | 1 | 1 | 0 | ✅ |
| middleware/ | 1 | 1 | 0 | ✅ |
| model/ | 7 | 0 | 7 | Pure DTOs (10-141 lines each), 268 total lines |

## Infrastructure
- docker-compose.yml: Well-structured (version 3.8, 3 services, healthchecks, depends_on, networks, volumes) ✅
- No TODOs or FIXMEs anywhere in codebase ✅

## Vue Router
- 7 routes defined matching 7 active views
- 2 orphan views exist on disk but not in router: MonthlyFortuneView.vue, WeeklyFortuneView.vue
- These views are never imported anywhere in vue/src/
