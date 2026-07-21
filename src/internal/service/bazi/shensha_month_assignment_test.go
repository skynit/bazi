package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var retiredMonthExtraNames = []string{"小时", "大败", "天医"}

func TestRetiredMonthExtraTablesHaveNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"monthXiaoShi", "monthDaBai", "tianYiByMonthZhi", "addMonthZhiExtra",
		"\"小时\"", "\"大败\"", "\"天医\"",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired month path %q", forbidden)
		}
	}
}

func TestRetiredMonthExtrasRemainAbsentAcrossFormerSearchSpace(t *testing.T) {
	for _, monthBranch := range data.Zhis {
		for _, placedBranch := range data.Zhis {
			for _, targetIndex := range []int{0, 2, 3} {
				pillars := []model.Pillar{
					{Gan: "甲", Zhi: "子"},
					poZhaiPillarForBranch(t, monthBranch),
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
				for _, name := range retiredMonthExtraNames {
					assertShenShaNameAbsentEverywhere(t, got, name)
				}
			}
		}
	}
}

func TestRetiredMonthExtrasRemainUnregistered(t *testing.T) {
	for _, name := range retiredMonthExtraNames {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}
