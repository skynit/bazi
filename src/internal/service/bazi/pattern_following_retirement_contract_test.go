package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

var retiredFollowingRuleIDs = []string{
	"pattern.following.congshi",
	"pattern.following.congcai",
	"pattern.following.congsha",
	"pattern.following.conger",
	"pattern.following.congruo",
}

func TestIncompleteFollowingShortcutsFailClosedAtFormalEntry(t *testing.T) {
	tests := []struct {
		name     string
		pillars  []model.Pillar
		monthZhi string
		scores   map[string]int
	}{
		{
			name: "former overlapping following shortcuts",
			pillars: []model.Pillar{
				{Gan: "己", Zhi: "丑"}, {Gan: "庚", Zhi: "申"},
				{Gan: "甲", Zhi: "午"}, {Gan: "辛", Zhi: "酉"},
			},
			monthZhi: "申",
			scores:   map[string]int{"木": 2, "火": 6, "土": 45, "金": 45, "水": 2},
		},
		{
			name: "former aggregate from-child shortcut",
			pillars: []model.Pillar{
				{Gan: "庚", Zhi: "申"}, {Gan: "辛", Zhi: "酉"},
				{Gan: "戊", Zhi: "子"}, {Gan: "壬", Zhi: "子"},
			},
			monthZhi: "酉",
			scores:   map[string]int{"木": 2, "火": 2, "土": 5, "金": 61, "水": 30},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.monthZhi)
			for _, candidate := range analysis.Candidates {
				for _, retiredID := range retiredFollowingRuleIDs {
					if candidate.RuleID == retiredID {
						t.Fatalf("retired following shortcut survived formal entry: %+v", candidate)
					}
				}
			}
		})
	}
}

func TestIncompleteFollowingProductionPathsAreAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go", "wuxing.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range append([]string{
			"checkCongRuoGe", "checkCongCaiGe", "checkCongShiGe",
			"checkQiMingCongShaGe", "checkCongErGe", "computeFavorByDayElem",
			"从格规则：财官并旺且日主无根（本地条件化）",
			"《滴天髓》从儿规则（本地条件化）",
		}, retiredFollowingRuleIDs...) {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains following shortcut %q", path, forbidden)
			}
		}
	}
}

func TestFollowingRetirementPreservesTenGodAndHiddenStemFacts(t *testing.T) {
	for _, tc := range []struct {
		seen, day, want string
	}{
		{seen: "甲", day: "甲", want: "比肩"},
		{seen: "癸", day: "甲", want: "正印"},
		{seen: "丙", day: "甲", want: "食神"},
		{seen: "戊", day: "甲", want: "偏财"},
		{seen: "庚", day: "甲", want: "七杀"},
	} {
		if got := ClassifyTenGod(tc.seen, tc.day, false); got != tc.want {
			t.Errorf("ClassifyTenGod(%s, %s) = %s, want %s", tc.seen, tc.day, got, tc.want)
		}
	}
	if got, want := zhiAllElements["申"], []string{"金", "水", "土"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("申 hidden elements = %v, want %v", got, want)
	}
}

func TestFollowingRetirementClassicalPDFHashes(t *testing.T) {
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

func TestFollowingRetirementManifestRecordsClassicalConflicts(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"《滴天髓阐微》PDF第173-174页从象",
			"PDF第186-187页顺局",
			"《三命通会》PDF第205页弃命从财",
			"《渊海子平》PDF第566-569页",
			"旧从财、从势、从杀、从弱、从儿五个检测器",
			"10%/15%生扶和60%主势",
			"从势漏掉食伤并旺、不能专从一神",
			"从儿错误要求身弱无根",
			"删除五项注册、算法、互斥争议和统一喜忌",
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
