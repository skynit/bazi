package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestIncompleteBaGeShortcutFailsClosedAtFormalEntry(t *testing.T) {
	tests := []struct {
		name     string
		pillars  []model.Pillar
		monthZhi string
	}{
		{
			name: "month main qi without matching exposed stem",
			pillars: []model.Pillar{
				{Gan: "戊", Zhi: "辰"}, {Gan: "己", Zhi: "巳"},
				{Gan: "甲", Zhi: "子"}, {Gan: "庚", Zhi: "午"},
			},
			monthZhi: "巳",
		},
		{
			name: "peer month with unrelated officer stem elsewhere",
			pillars: []model.Pillar{
				{Gan: "辛", Zhi: "酉"}, {Gan: "丙", Zhi: "寅"},
				{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "辰"},
			},
			monthZhi: "寅",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.monthZhi)
			for _, candidate := range analysis.Candidates {
				if candidate.RuleID == "pattern.bage.yueling" {
					t.Fatalf("retired BaGe shortcut survived formal entry: %+v", candidate)
				}
			}
		})
	}
}

func TestIncompleteBaGeProductionPathIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"checkBaGe", "buildBaGeResult", "patternFromTenGod", "geFavorable",
			"geUnfavorable", "tenGodCategory", "pattern.bage.yueling",
			"《子平真诠》月令取格规则",
		} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains incomplete BaGe path %q", path, forbidden)
			}
		}
	}
}

func TestBaGeRetirementManifestRecordsMissingPerPatternConditions(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern.bage.yueling",
			"仅凭月支本气十神或任意其他柱透干",
			"透干、身强、扶助、制化和破格条件并不相同",
			"月支藏干与逐项十神事实继续由独立基础层输出",
			"恢复八格前须逐格建立条件 Profile 和专家 Gold",
			"《三命通会》PDF第153-220页",
			"《渊海子平》PDF第711-713页",
		} {
			if !strings.Contains(table.Description+table.Source, fragment) {
				t.Errorf("pattern description/source missing %q: %+v", fragment, table)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func baGeRetirementScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}
