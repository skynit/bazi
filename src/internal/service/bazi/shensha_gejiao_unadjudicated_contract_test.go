package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestGeJiaoUnadjudicatedTableHasNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"geJiaoPair", "\"隔角煞\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains unadjudicated 隔角煞 path %q", forbidden)
		}
	}
}

func TestGeJiaoRemainsAbsentAcrossFormerYearBranchSearchSpace(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for _, placedBranch := range data.Zhis {
			got := calcPoZhaiFixture(t, yearBranch, placedBranch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "隔角煞")
		}
	}
}

func TestGeJiaoRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("隔角煞")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("隔角煞 metadata = %+v, want unregistered/not_available", meta)
	}
}
