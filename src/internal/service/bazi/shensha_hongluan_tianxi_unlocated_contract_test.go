package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestHongLuanAndTianXiUnlocatedTablesHaveNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"var hongLuan =", "hongLuan[", "var tianXi =", "tianXi[", "\"红鸾\"", "\"天喜\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains unlocated rule path %q", forbidden)
		}
	}
}

func TestHongLuanAndTianXiRemainAbsentAcrossFormerYearBranchSearchSpace(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for _, placedBranch := range data.Zhis {
			got := calcPoZhaiFixture(t, yearBranch, placedBranch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "红鸾")
			assertShenShaNameAbsentEverywhere(t, got, "天喜")
		}
	}
}

func TestHongLuanAndTianXiRemainUnregistered(t *testing.T) {
	for _, name := range []string{"红鸾", "天喜"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}
