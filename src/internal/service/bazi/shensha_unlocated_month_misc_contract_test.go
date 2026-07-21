package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var retiredUnlocatedMonthNames = []string{"天刑", "天火", "天贼", "大时", "兵禁", "天吏"}

func TestUnlocatedMonthMiscRulesHaveNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range retiredUnlocatedMonthNames {
		if strings.Contains(string(source), "\""+name+"\"") {
			t.Errorf("production source still contains unlocated month rule %s", name)
		}
	}
}

func TestUnlocatedMonthMiscRulesRemainAbsentAcrossFormerSearchSpace(t *testing.T) {
	for _, monthZhi := range data.Zhis {
		for _, placedBranch := range data.Zhis {
			for _, targetIndex := range []int{0, 2, 3} {
				pillars := []model.Pillar{
					{Gan: "甲", Zhi: "子"},
					poZhaiPillarForBranch(t, monthZhi),
					{Gan: "丙", Zhi: "寅"},
					{Gan: "戊", Zhi: "辰"},
				}
				pillars[targetIndex] = poZhaiPillarForBranch(t, placedBranch)
				got, err := CalcShenShaByPillars(ShenShaPillars{
					Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
				})
				if err != nil {
					t.Fatal(err)
				}
				for _, name := range retiredUnlocatedMonthNames {
					assertShenShaNameAbsentEverywhere(t, got, name)
				}
			}
		}
	}
}

func TestUnlocatedMonthMiscRulesRemainUnregistered(t *testing.T) {
	for _, name := range retiredUnlocatedMonthNames {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
	if meta := LookupShenShaMeta("天刑煞"); meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Errorf("located 天刑煞 profile was damaged: %+v", meta)
	}
}
