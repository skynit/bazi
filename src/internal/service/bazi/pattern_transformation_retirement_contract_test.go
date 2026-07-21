package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

var retiredTransformationRuleIDs = []string{
	"pattern.transform.huaqi",
	"pattern.transform.conghua",
}

func TestIncompleteTransformationShortcutsFailClosedAtFormalEntry(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "己", Zhi: "丑"},
		{Gan: "甲", Zhi: "午"}, {Gan: "庚", Zhi: "午"},
	}
	for _, scores := range []map[string]int{
		{"木": 10, "火": 20, "土": 50, "金": 15, "水": 5},
		{"木": 27, "火": 10, "土": 40, "金": 13, "水": 10},
	} {
		analysis := AnalyzePatternExtended(pillars, "丑")
		for _, candidate := range analysis.Candidates {
			for _, retiredID := range retiredTransformationRuleIDs {
				if candidate.RuleID == retiredID {
					t.Fatalf("retired transformation shortcut survived formal entry for scores %v: %+v", scores, candidate)
				}
			}
		}
	}
}

func TestIncompleteTransformationProductionPathsAreAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go", "wuxing.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range append([]string{
			"checkHuaQiGe", "checkCongHuaGe", "isHuaQiWang",
			"favorHuaQi", "totalScore", "《三命通会》化气格规则（本地条件化）",
			"《三命通会》从化格规则（本地条件化）",
		}, retiredTransformationRuleIDs...) {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains transformation shortcut %q", path, forbidden)
			}
		}
	}
}

func TestTransformationRetirementPreservesFiveCombineFacts(t *testing.T) {
	pillars := []ganRelationPillar{
		{key: "year", label: labelYear, stem: "丙", branch: "子"},
		{key: "month", label: labelMonth, stem: "己", branch: "丑"},
		{key: "day", label: labelDay, stem: "甲", branch: "午"},
		{key: "hour", label: labelHour, stem: "庚", branch: "午"},
	}
	got := findGanRelation(buildGanRelationGraph(pillars), "五合", "stem.combine.甲己")
	if got == nil || got.TargetElement != "土" || got.StructureStatus != "complete_structure" ||
		got.TransformationStatus != "unadjudicated" || got.TransformationEvidence == nil {
		t.Fatalf("five-combine evidence was not preserved: %+v", got)
	}
}

func TestTransformationRetirementClassicalPDFHashes(t *testing.T) {
	wants := map[string]string{
		"library/滴天髓阐微.pdf": "65c67d88421319fccbba23bce88d61d4ace288a7913edd6a10ebf3143e72a48b",
		"library/三命通会.pdf":  "63eb2a85036ebbd360b815a58780edd242bda0fd1a5faaac7413e00c5f726d47",
		"library/渊海子平.pdf":  "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5",
	}
	for path, want := range wants {
		raw, err := os.ReadFile("../../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("%s SHA-256 = %s, want %s", path, got, want)
		}
	}
}

func TestTransformationRetirementManifestRecordsAlgorithmDefects(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"《滴天髓阐微》PDF第177-178页化象",
			"《三命通会》PDF第72-74页论十干化气",
			"《渊海子平》PDF第575-576页化气诗诀",
			"旧 checkHuaQiGe 与 checkCongHuaGe",
			"30%与25%日主分数阈值",
			"同盘重复发布化气格与从化格",
			"只检查月干或时干",
			"次月、辰时、妒合、得辰",
			"删除两项注册、算法、快捷月令表和统一喜忌",
			"天干五合结构与unadjudicated成化证据继续保留",
			"classical_text_local / text_located_not_expert_gold",
		} {
			if !strings.Contains(table.Source+table.Description, fragment) {
				t.Errorf("pattern description/source missing %q: %+v", fragment, table)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
