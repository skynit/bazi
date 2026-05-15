## Learnings — Fortune Lookup Data Rules

### Go package structure
- Module: `bazi` (go.mod)
- Package: `service` for business logic services
- Models exist in `bazi/internal/model`: `AuspiciousRule`, `ActivityCatalog`

### Lookup table mappings (all 10 stems, 12 branches covered)
- Lucky colors: grouped by five elements (木火土金水 → 甲乙/丙丁/戊己/庚辛/壬癸)
- Lucky numbers: same five-element grouping (3,8 / 2,7 / 5,0 / 4,9 / 1,6)
- Wealth direction: 甲乙→东北, 丙丁→正西, 戊己→正北, 庚辛→正东, 壬癸→正南
- Clash zodiacs: six mutual clash pairs (六冲) — each branch maps to the branch 6 positions away in the 12-cycle
- Auspicious hours: 2-3 double-hour ranges per branch
- Activities: 25 宜 and 25 忌 activities

### Go environment
- `go` binary not installed in this environment; verification is manual code review only
- Build and test execution must happen in Docker or with Go toolchain available
