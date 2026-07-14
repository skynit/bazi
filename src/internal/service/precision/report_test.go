package precision

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEvaluateBaziPillarOnlyDoesNotSelfValidatePillars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pillar-only.json")
	fixture := `{
  "version":"1",
  "cases":[{
    "id":"pillar-only",
    "gender":"MALE",
    "expected":{
      "year_pillar":"庚午",
      "month_pillar":"壬午",
      "day_pillar":"癸卯",
      "hour_pillar":"丙辰",
      "day_master":"癸"
    }
  }]
}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	module := evaluateBazi(path)
	if module.Checks != 0 || module.Passed != 0 {
		t.Fatalf("pillar-only fixture must not count pillar/day-master self-checks: %+v", module)
	}
	if len(module.Warnings) == 0 || !strings.Contains(module.Warnings[0], "does not validate calendar pillar accuracy") {
		t.Fatalf("expected calendar-accuracy warning, got %v", module.Warnings)
	}
}

func TestEvaluateBaziDateFixtureChecksCalculatedPillars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "date.json")
	fixture := `{
  "version":"1",
  "cases":[{
    "id":"date-case",
    "year":1990,"month":6,"day":15,"hour":8,"minute":0,"gender":"MALE",
    "expected":{"year_pillar":"错误"}
  }]
}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	module := evaluateBazi(path)
	if module.Checks != 1 || module.Failed != 1 {
		t.Fatalf("date fixture should verify calculated pillar, got %+v", module)
	}
}
