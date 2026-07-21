package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

var retiredCompoundRuleIDs = []string{
	"pattern.compound.shishen_zhisha",
	"pattern.compound.shangguan_peiyin",
	"pattern.compound.caizi_shawang",
	"pattern.compound.shishen_shengcai",
	"pattern.compound.zhengguan_peiyin",
}

func TestIncompleteCompoundShortcutsFailClosedAtFormalEntry(t *testing.T) {
	tests := []struct {
		name     string
		pillars  []model.Pillar
		monthZhi string
	}{
		{
			name: "former food-generates-wealth shortcut",
			pillars: []model.Pillar{
				{Gan: "丙", Zhi: "辰"}, {Gan: "己", Zhi: "巳"},
				{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "辰"},
			},
			monthZhi: "巳",
		},
		{
			name: "former simultaneous compound shortcuts",
			pillars: []model.Pillar{
				{Gan: "庚", Zhi: "戌"}, {Gan: "丁", Zhi: "巳"},
				{Gan: "甲", Zhi: "辰"}, {Gan: "戊", Zhi: "申"},
			},
			monthZhi: "巳",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.monthZhi)
			for _, candidate := range analysis.Candidates {
				for _, retiredID := range retiredCompoundRuleIDs {
					if candidate.RuleID == retiredID {
						t.Fatalf("retired compound shortcut survived formal entry: %+v", candidate)
					}
				}
			}
		})
	}
}

func TestIncompleteCompoundProductionPathsAreAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range append([]string{
			"checkCompoundGeCandidates", "compoundRuleID", "buildCompound", "samePillarPairs",
			"hasAdjacentPair", "stemFromQi", "monthZhiBenQiMap",
			"《子平真诠》相神与成格关系",
		}, retiredCompoundRuleIDs...) {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains compound shortcut %q", path, forbidden)
			}
		}
	}
}

func TestCompoundRetirementPreservesUnderlyingTenGodFacts(t *testing.T) {
	for _, tc := range []struct {
		seen, day, want string
	}{
		{seen: "丙", day: "甲", want: "食神"},
		{seen: "戊", day: "甲", want: "偏财"},
		{seen: "庚", day: "甲", want: "七杀"},
		{seen: "癸", day: "甲", want: "正印"},
		{seen: "辛", day: "甲", want: "正官"},
	} {
		if got := ClassifyTenGod(tc.seen, tc.day, false); got != tc.want {
			t.Errorf("ClassifyTenGod(%s, %s) = %s, want %s", tc.seen, tc.day, got, tc.want)
		}
	}
}

func TestCompoundRetirementManifestRecordsAlgorithmDefects(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧五个 pattern.compound.* 检测器",
			"同柱或相邻十神配对",
			"子卯午酉及四库阴干",
			"邻接 helper 允许两个十神来自同一侧",
			"身强旺衰、制化力度、去财去印、伤官伤尽、枭夺食",
			"删除五项注册、算法、位置 helper 和统一喜忌",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func compoundRetirementScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}
