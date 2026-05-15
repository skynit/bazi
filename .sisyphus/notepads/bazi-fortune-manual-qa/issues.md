# Manual QA Issues

## Minor Issues
1. **Orphan Vue views**: `MonthlyFortuneView.vue` and `WeeklyFortuneView.vue` exist in `vue/src/views/` but are not referenced in the router nor imported anywhere. Either dead code or unfinished feature. 2 files, ~unknown LOC.
2. **Untested service data stores**: `auspicious_data.go` (179 lines) and `ziwei_templates.go` (676 lines) lack unit tests. Both are hardcoded lookup tables — low risk for correctness bugs but no regression protection.
3. **Model layer untested**: 7 DTO files (268 total lines) have no tests. Acceptable for pure data structs.

## Non-Issues (Verified)
- All handler endpoints are tested ✅
- All service logic (non-data) is tested ✅
- config, middleware fully tested ✅
- docker-compose is valid and complete ✅
- No TODO/FIXME debt ✅
