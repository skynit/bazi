package bazi

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternDetectorRegistryIsTheSingleManifest(t *testing.T) {
	detectors := patternDetectorRegistry()
	if len(detectors) != 10 || patternDetectorCount() != len(detectors) {
		t.Fatalf("pattern detector registry/count = %d/%d, want 10", len(detectors), patternDetectorCount())
	}
	seen := make(map[string]struct{}, len(detectors))
	for index, detector := range detectors {
		if detector.ruleID == "" || detector.source == "" || len(detector.outputNames) == 0 ||
			detector.algorithmSHA256 == "" || detector.behaviorSHA256 == "" || detector.profileSHA256 == "" || detector.detect == nil {
			t.Errorf("detector %d is incomplete: %+v", index, detector)
		}
		if detector.category != patternCategoryStructural && detector.category != patternCategoryAuxiliary {
			t.Errorf("detector %s has invalid category %q", detector.ruleID, detector.category)
		}
		if _, exists := seen[detector.ruleID]; exists {
			t.Errorf("duplicate detector rule ID %q", detector.ruleID)
		}
		seen[detector.ruleID] = struct{}{}
	}
}

func TestPatternDetectorRegistryReturnsIndependentSnapshots(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	want := AnalyzePatternExtended(pillars, "寅")

	mutated := patternDetectorRegistry()
	mutated[0].ruleID = "mutated"
	mutated[0].source = "mutated"
	mutated[0].category = "mutated"
	mutated[0].algorithmSHA256 = "mutated"
	mutated[0].behaviorSHA256 = "mutated"
	mutated[0].profileSHA256 = "mutated"
	mutated[0].outputNames[0] = "mutated"
	mutated[0].detect = nil

	fresh := patternDetectorRegistry()
	if fresh[0].ruleID == "mutated" || fresh[0].source == "mutated" ||
		fresh[0].category == "mutated" || fresh[0].algorithmSHA256 == "mutated" || fresh[0].behaviorSHA256 == "mutated" || fresh[0].profileSHA256 == "mutated" ||
		fresh[0].outputNames[0] == "mutated" || fresh[0].detect == nil {
		t.Fatalf("fresh detector registry inherited mutation: %+v", fresh[0])
	}
	if got := AnalyzePatternExtended(pillars, "寅"); !reflect.DeepEqual(got, want) {
		t.Fatalf("registry snapshot mutation changed analysis:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestPatternDetectorRegistryReplacesManualCountAndAddCalls(t *testing.T) {
	source, err := os.ReadFile("pattern_candidates.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []*regexp.Regexp{
		regexp.MustCompile(`patternDetectorCount\s*=\s*10`),
		regexp.MustCompile(`\badd\s*\(\s*check`),
		regexp.MustCompile(`\bvar\s+patternDetectorRegistry\b`),
	} {
		if forbidden.Match(source) {
			t.Errorf("pattern candidate source still contains manual registry path %q", forbidden)
		}
	}
	for _, required := range []string{
		"func patternDetectorRegistry() []patternDetectorDefinition",
		"detectors := patternDetectorRegistry()",
		"detectorCount := len(detectors)",
		"for _, detector := range detectors",
	} {
		if !strings.Contains(string(source), required) {
			t.Errorf("pattern candidate source missing registry contract %q", required)
		}
	}
}

func TestPatternDetectorRegistryManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧patternDetectorRegistry是包级可变var",
			"同包代码可永久替换规则身份、来源、分类或调用函数",
			"pattern-candidate-set-v14改为每次返回独立注册表快照",
			"执行和detector_count绑定同一局部快照",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
