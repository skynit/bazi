# Code Review Issues — bazi-fortune

## CRITICAL (Compile Errors)

### 1. Duplicate `ChartStore` interface
- **Files**: `src/internal/handler/fortune.go:15` and `src/internal/handler/fortune_weekly.go:17`
- **Issue**: Same `type ChartStore interface { FindByID(id uint) (*model.BirthChart, error) }` defined twice in the same package `handler`. Go does not allow duplicate type declarations in the same package.
- **Fix**: Remove from `fortune_weekly.go`. It already exists in `fortune.go`.

### 2. Missing closing brace in `testBirthChart()`
- **File**: `src/internal/handler/fortune_monthly_test.go:30-33`
- **Issue**: The function body opening `{` on line 30 is never closed. The `}` on line 33 closes only the struct literal inside `return`. Verified via ast-grep: the pattern matches through entire file remainder.
- **Fix**: Add missing `}` after line 33, before line 34.

### 3. Illegal cross-package method definition
- **File**: `src/internal/handler/fortune_monthly_test.go:35`
- **Issue**: `func (b *model.BirthChart) TableName() string` defined in package `handler`, but `BirthChart` lives in package `model`. Go prohibits defining methods on types from other packages. Additionally, this line is parsed as being inside `testBirthChart()` due to bug #2.
- **Fix**: Remove entirely (dead code — never called). If `TableName` is needed, define it in `model/birth_chart.go`.

## RUNTIME BUGS

### 4. Nil `Charts` field on `ZiWeiPeriodHandler` in main.go
- **File**: `src/cmd/main.go:77-78`
- **Issue**: `ziweiPeriodH := &handler.ZiWeiPeriodHandler{Service: ziweiSvc}` omits `Charts`. Calling `POST /api/ziwei/period` or `POST /api/ziwei/overlay` will panic with nil pointer dereference at `h.Charts.FindByID(chartID)`.
- **Fix**: Wire a `ChartStore` implementation into `ZiWeiPeriodHandler`.

## DESIGN ISSUES

### 5. `MonthlyFortuneResponse` has field named `WeeklyScore`
- **File**: `src/internal/model/dto.go:115`
- **Issue**: `WeeklyScore int \`json:"weekly_score"\`` in `MonthlyFortuneResponse` — copy-paste artifact. Should be `MonthlyScore`.
- **Fix**: Rename field and JSON tag.

### 6. Dead code: `testBirthChart()`
- **File**: `src/internal/handler/fortune_monthly_test.go:30-33`
- **Issue**: Function never called anywhere (grep confirmed). Also broken (see bug #2). Comment-only struct literal `// gorm.Model` suggests placeholder.
- **Fix**: Remove the function.

### 7. Orphaned Vue components (never imported)
- `vue/src/views/WeeklyFortuneView.vue` — fully implemented but no route, not imported by any file
- `vue/src/views/MonthlyFortuneView.vue` — fully implemented but no route, not imported by any file
- `vue/src/components/HelloWorld.vue` — default Vite template, unused
- `vue/src/components/ElementImages.vue` — unused
- `vue/src/components/AIAnalysisEntry.vue` — unused

### 8. Stub endpoint
- **File**: `src/internal/handler/ai_stub.go`
- **Issue**: Returns hardcoded `coming_soon` response. Intentional placeholder but noted.

### 9. AI slop patterns
- Over-commenting: `ziwei_chart.go:20-22` ("defensive safety" comment on JWT check already handled by middleware)
- Large comment wall: `ziwei_templates.go:3-20` (references classical texts — borderline excessive)
- `RegisterFortuneRoutes` name misleading (only registers `/weekly`, not all fortune routes)
- Duplicate routing logic between `main.go` and test-only `Register*` functions

### 10. Silenced unused variables
- **File**: `src/cmd/main.go:25-26`
- **Issue**: `_, _ = auspicious, interpreter` — variables created but genuinely unused. Comment says "will be wired when handlers are extended."
