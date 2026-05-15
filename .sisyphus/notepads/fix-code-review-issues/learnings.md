# Learnings — Fix Code Review Issues

## FIX 1: Duplicate ChartStore interface
- `ChartStore` interface was defined in both `fortune.go` and `fortune_weekly.go` (same `handler` package).
- Removed from `fortune.go` since `fortune_weekly.go` has the canonical definition.
- Since they're in the same package, no import changes needed — `FortuneHandler` still references `ChartStore` from the other file.

## FIX 2: Broken testBirthChart() + TableName()
- `testBirthChart()` had no closing brace (line 30-33 only).
- `TableName()` was an invalid cross-package method definition on `model.BirthChart`.
- Both dead code — deleted.

## FIX 3: ZiWeiPeriodHandler nil-panic
- `main.go` creates `ZiWeiPeriodHandler` without `Charts` field, causing nil-panic when `lookupAndCalculate` accesses `h.Charts.FindByID()`.
- Added nil-guard in `lookupAndCalculate`: returns `errors.New("chart store not configured")` if `h.Charts == nil`.
- This is cleaner than adding mock code in `main.go`.

## FIX 4: Orphaned scaffolding files
- Deleted 6 files: HelloWorld.vue, vite.svg, vue.svg, hero.png, favicon.svg, icons.svg
- Removed empty `vue/src/assets/` directory
- Removed dangling `favicon.svg` link from `vue/index.html`
