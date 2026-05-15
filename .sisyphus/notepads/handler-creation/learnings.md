# Learnings

## Handler creation patterns
- All handlers follow the same structure: struct with dependencies (stores, services), handler methods, route registration function
- JWT middleware pattern: `middleware.InitJWT("test-secret")` in test setup, `middleware.AuthMiddleware()` on routes
- `c.Get("userID")` retrieves user ID from JWT context (set by middleware)
- `ChartStore` interface (FindByID) is defined in both `fortune.go` and `fortune_weekly.go` — identical definitions are fine in Go
- Mock stores in tests: simple structs implementing store interfaces with in-memory maps
- httptest pattern: `httptest.NewRequest`, `.Header.Set("Authorization", "Bearer "+token)`, `gin.SetMode(gin.TestMode)`
- LSP "No active builds contain" warning is workspace config issue, not a code error — `go` binary not available in this environment

## ZiWei service methods
- `CalculateDayun(chart)` → `Dayun` ([]DayunStage) — no extra params needed
- `CalculateLiunian(chart, year)` → `*ZiWeiChart` — needs target year
- `CalculateLiuyue(chart, lunarMonth)` → `*ZiWeiChart`
- `CalculateLiuri(chart, lunarDay)` → `*ZiWeiChart`
- All methods handle nil chart gracefully (return nil)
- `CalculateChart` accepts gender as "男"/"女" directly
