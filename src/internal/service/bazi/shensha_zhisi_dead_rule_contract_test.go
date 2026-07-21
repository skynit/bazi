package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestZhiSiDeadRuleHasNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "致死") {
		t.Fatal("production source still contains dead 致死 rule or suppression entry")
	}
}

func TestZhiSiRemainsAbsentAcrossFormerMonthSearchSpace(t *testing.T) {
	for _, monthZhi := range data.Zhis {
		for _, placedBranch := range data.Zhis {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   model.Pillar{Gan: "甲", Zhi: "子"},
				Month:  poZhaiPillarForBranch(t, monthZhi),
				Day:    model.Pillar{Gan: "丙", Zhi: "寅"},
				Hour:   poZhaiPillarForBranch(t, placedBranch),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertShenShaNameAbsentEverywhere(t, got, "致死")
		}
	}
}

func TestZhiSiRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("致死")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("致死 metadata = %+v, want unregistered/not_available", meta)
	}
}
