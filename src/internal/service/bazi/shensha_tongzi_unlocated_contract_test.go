package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestRetiredTongZiShaHasNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"isTongZiSha", "\"童子煞\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired 童子煞 path %q", forbidden)
		}
	}
}

func TestRetiredTongZiShaRemainsAbsentAcrossFormerSearchSpace(t *testing.T) {
	for _, monthZhi := range data.Zhis {
		for _, hourZhi := range data.Zhis {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   model.Pillar{Gan: "甲", Zhi: "子"},
				Month:  poZhaiPillarForBranch(t, monthZhi),
				Day:    model.Pillar{Gan: "丙", Zhi: "寅"},
				Hour:   poZhaiPillarForBranch(t, hourZhi),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertShenShaNameAbsentEverywhere(t, got, "童子煞")
		}
	}
}

func TestRetiredTongZiShaRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("童子煞")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("童子煞 metadata = %+v, want unregistered/not_available", meta)
	}
}
