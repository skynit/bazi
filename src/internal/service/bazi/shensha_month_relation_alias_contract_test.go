package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var retiredMonthRelationTargets = map[string]map[string]string{
	"月刑": {
		"子": "卯", "丑": "戌", "寅": "巳", "卯": "子", "辰": "辰", "巳": "申",
		"午": "午", "未": "丑", "申": "寅", "酉": "酉", "戌": "未", "亥": "亥",
	},
	"月害": {
		"子": "未", "丑": "午", "寅": "巳", "卯": "辰", "辰": "卯", "巳": "寅",
		"午": "丑", "未": "子", "申": "亥", "酉": "戌", "戌": "酉", "亥": "申",
	},
}

func TestRetiredMonthRelationAliasesHaveNoShenShaProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"月刑", "月害"} {
		if strings.Contains(string(source), name) {
			t.Errorf("production shen-sha source still contains relation alias %s", name)
		}
	}
}

func TestRetiredMonthRelationAliasesRemainAbsentAcrossFormerSearchSpace(t *testing.T) {
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
			for name := range retiredMonthRelationTargets {
				assertShenShaNameAbsentEverywhere(t, got, name)
			}
		}
	}
}

func TestRetiredMonthRelationFactsRemainInBranchRelationGraph(t *testing.T) {
	for name, targets := range retiredMonthRelationTargets {
		wantType := "相刑"
		if name == "月害" {
			wantType = "六害"
		}
		for _, monthZhi := range data.Zhis {
			target := targets[monthZhi]
			relations := buildZhiRelationGraph([]branchRelationPillar{
				{key: "month", label: "月柱", branch: monthZhi},
				{key: "hour", label: "时柱", branch: target},
			})
			if !hasBranchRelationType(relations, wantType) {
				t.Errorf("%s %s/%s missing canonical %s relation: %+v", name, monthZhi, target, wantType, relations)
			}
		}
	}
}

func TestRetiredMonthRelationAliasesRemainUnregistered(t *testing.T) {
	for name := range retiredMonthRelationTargets {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func hasBranchRelationType(relations []ZhiRelation, want string) bool {
	for _, relation := range relations {
		if relation.Type == want {
			return true
		}
	}
	return false
}
