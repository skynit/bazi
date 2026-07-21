package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestBranchRelationGraphDistinguishesPartialAndCompleteGroups(t *testing.T) {
	tests := []struct {
		name          string
		branches      []string
		wantType      string
		wantRule      string
		wantStatus    string
		wantStructure string
		wantTransform string
		forbidType    string
		wantTarget    string
	}{
		{"三合含中神为半合", []string{"申", "子", "申", "子"}, "半合", "branch.sanhe.water.partial", "partial", "partial_structure", "not_applicable", "三合局", "水"},
		{"三合首尾为拱合", []string{"申", "辰", "午", "酉"}, "拱合", "branch.sanhe.water.partial", "partial", "partial_structure", "not_applicable", "三合局", "水"},
		{"三会两支为半会", []string{"寅", "卯", "寅", "卯"}, "半会", "branch.sanhui.wood.partial", "partial", "partial_structure", "not_applicable", "三会局", "木"},
		{"三合三支完整但成化未裁决", []string{"申", "子", "辰", "申"}, "三合局", "branch.sanhe.water", "complete", "complete_structure", "unadjudicated", "半合", "水"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			relations := buildZhiRelationGraph(testRelationPillars(tc.branches))
			got := findZhiRelationByRule(relations, tc.wantRule)
			if got == nil {
				t.Fatalf("missing %s in %+v", tc.wantRule, relations)
			}
			if got.Type != tc.wantType || got.Status != tc.wantStatus || got.StructureStatus != tc.wantStructure || got.TransformationStatus != tc.wantTransform || got.TargetElement != tc.wantTarget {
				t.Fatalf("relation = %+v", *got)
			}
			if findZhiRelationByType(relations, tc.forbidType) != nil {
				t.Fatalf("partial/complete relation leaked forbidden type %s: %+v", tc.forbidType, relations)
			}
		})
	}
}

func TestBranchRelationGraphMarksCombinationConflictsWithoutSuppressingEitherSide(t *testing.T) {
	relations := buildZhiRelationGraph(testRelationPillars([]string{"巳", "申", "巳", "申"}))
	combine := findZhiRelationByID(relations, "branch.liuhe.巳申.year-month")
	if combine == nil || combine.Status != "disputed" || combine.TransformationStatus != "disputed" {
		t.Fatalf("巳申六合 conflict = %+v", combine)
	}
	for _, want := range []string{"branch.punish.巳申.year-month", "branch.break.巳申.year-month"} {
		if !containsBranch(combine.ConflictsWith, want) {
			t.Fatalf("六合未记录冲突 %s: %+v", want, combine)
		}
	}
	if findZhiRelationByID(relations, "branch.punish.巳申.year-month") == nil || findZhiRelationByID(relations, "branch.break.巳申.year-month") == nil {
		t.Fatalf("conflicting relations were suppressed: %+v", relations)
	}
	if len(relations) < 3 || relations[0].Priority < relations[1].Priority {
		t.Fatalf("relations are not deterministically priority-sorted: %+v", relations)
	}
}

func TestBranchRelationGraphMarksCompleteMeetingDisputedByInternalHarm(t *testing.T) {
	relations := buildZhiRelationGraph(testRelationPillars([]string{"寅", "卯", "辰", "寅"}))
	meeting := findZhiRelationByRule(relations, "branch.sanhui.wood")
	if meeting == nil || meeting.StructureStatus != "complete_structure" || meeting.Status != "disputed" || meeting.TransformationStatus != "disputed" {
		t.Fatalf("寅卯辰三会 conflict = %+v", meeting)
	}
	if len(meeting.ConflictsWith) == 0 || !strings.Contains(strings.Join(meeting.DisputeReasons, ";"), "六害") {
		t.Fatalf("三会未暴露卯辰害争议: %+v", meeting)
	}
}

func TestBranchRelationGraphEmitsOneCompleteThreePunishment(t *testing.T) {
	relations := buildZhiRelationGraph(testRelationPillars([]string{"寅", "巳", "申", "寅"}))
	punishment := findZhiRelationByRule(relations, "branch.sanxing.wuen")
	if punishment == nil || punishment.Type != "三刑" || punishment.Subtype != "无恩之刑" || len(punishment.Pillars) != 4 {
		t.Fatalf("complete 三刑 = %+v", punishment)
	}
	evidence := strings.Join(punishment.Evidence, ";")
	if !strings.Contains(evidence, "《三命通会》PDF第83-84页") ||
		!strings.Contains(evidence, "《渊海子平》PDF第32页异称") ||
		!strings.Contains(punishment.Detail, "名称采用《三命通会》口径") {
		t.Fatalf("三刑名称分歧未显式输出: %+v", punishment)
	}
	for _, relation := range relations {
		if relation.Type == "相刑" && pairInGroup(relation.Branches[0], relation.Branches[1], []string{"寅", "巳", "申"}) {
			t.Fatalf("complete 三刑 was fragmented into pairwise punishment: %+v", relations)
		}
	}
}

func TestBranchPunishmentNamingDisagreementIsNotSilentlyCollapsed(t *testing.T) {
	tests := []struct {
		branches []string
		ruleID   string
		subtype  string
		sanming  string
		yuanhai  string
	}{
		{[]string{"寅", "巳", "申", "寅"}, "branch.sanxing.wuen", "无恩之刑", "寅巳申为无恩之刑", "寅巳申为恃势之刑"},
		{[]string{"丑", "戌", "未", "丑"}, "branch.sanxing.shishi", "恃势之刑", "丑戌未为恃势之刑", "丑戌未为无恩之刑"},
	}
	for _, tc := range tests {
		t.Run(tc.ruleID, func(t *testing.T) {
			relation := findZhiRelationByRule(buildZhiRelationGraph(testRelationPillars(tc.branches)), tc.ruleID)
			if relation == nil || relation.Subtype != tc.subtype {
				t.Fatalf("relation = %+v, want subtype %s", relation, tc.subtype)
			}
			evidence := strings.Join(relation.Evidence, ";")
			if !strings.Contains(evidence, tc.sanming) || !strings.Contains(evidence, tc.yuanhai) {
				t.Fatalf("naming evidence = %q", evidence)
			}
		})
	}
}

func TestValidGanZhiAnalysisRequiresExactCanonicalGraph(t *testing.T) {
	year := pillarForRelationTest("甲", "申")
	month := pillarForRelationTest("丙", "子")
	day := pillarForRelationTest("戊", "辰")
	hour := pillarForRelationTest("庚", "申")
	analysis, err := CalcGanZhiAnalysis(year, month, day, hour)
	if err != nil {
		t.Fatal(err)
	}
	if !ValidGanZhiAnalysis(analysis, year, month, day, hour) {
		t.Fatal("freshly calculated relation graph must validate")
	}

	tampered := analysis
	tampered.ZhiRelations = append([]ZhiRelation(nil), analysis.ZhiRelations...)
	tampered.ZhiRelations[0].RuleID = "tampered"
	if ValidGanZhiAnalysis(tampered, year, month, day, hour) {
		t.Fatal("relation graph with a tampered branch rule must be rejected")
	}

	incomplete := analysis
	incomplete.GanRelations = nil
	if ValidGanZhiAnalysis(incomplete, year, month, day, hour) {
		t.Fatal("relation graph with missing stem relations must be rejected")
	}
}

func TestGanZhiAnalysisRejectsInvalidSixtyCyclePillars(t *testing.T) {
	valid := [4]model.Pillar{
		pillarForRelationTest("甲", "子"), pillarForRelationTest("丙", "寅"),
		pillarForRelationTest("戊", "辰"), pillarForRelationTest("庚", "申"),
	}
	tests := []struct {
		name     string
		index    int
		invalid  model.Pillar
		position string
	}{
		{name: "empty year", index: 0, invalid: model.Pillar{}, position: "year"},
		{name: "impossible month pair", index: 1, invalid: pillarForRelationTest("甲", "丑"), position: "month"},
		{name: "invalid hour symbol", index: 3, invalid: pillarForRelationTest("甲", "无"), position: "hour"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pillars := valid
			pillars[tc.index] = tc.invalid
			analysis, err := CalcGanZhiAnalysis(pillars[0], pillars[1], pillars[2], pillars[3])
			if err == nil || !strings.Contains(err.Error(), tc.position) {
				t.Fatalf("invalid pillars returned analysis=%+v, err=%v", analysis, err)
			}
			if len(analysis.GanRelations) != 0 || len(analysis.ZhiRelations) != 0 {
				t.Fatalf("invalid pillars produced relation graph: %+v", analysis)
			}
			if ValidGanZhiAnalysis(analysis, pillars[0], pillars[1], pillars[2], pillars[3]) {
				t.Fatal("invalid pillars passed persisted relation validation")
			}
		})
	}
}

func testRelationPillars(branches []string) []branchRelationPillar {
	keys := []string{"year", "month", "day", "hour"}
	labels := []string{labelYear, labelMonth, labelDay, labelHour}
	result := make([]branchRelationPillar, len(branches))
	for i, branch := range branches {
		result[i] = branchRelationPillar{key: keys[i], label: labels[i], branch: branch}
	}
	return result
}

func pillarForRelationTest(stem, branch string) model.Pillar {
	return model.Pillar{Gan: stem, Zhi: branch}
}

func findZhiRelationByRule(relations []ZhiRelation, ruleID string) *ZhiRelation {
	for i := range relations {
		if relations[i].RuleID == ruleID {
			return &relations[i]
		}
	}
	return nil
}

func findZhiRelationByID(relations []ZhiRelation, id string) *ZhiRelation {
	for i := range relations {
		if relations[i].ID == id {
			return &relations[i]
		}
	}
	return nil
}

func findZhiRelationByType(relations []ZhiRelation, relationType string) *ZhiRelation {
	for i := range relations {
		if relations[i].Type == relationType {
			return &relations[i]
		}
	}
	return nil
}
