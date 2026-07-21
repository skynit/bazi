package bazi

import (
	"os"
	"strings"
	"testing"
)

func TestPatternProfileProductionPathsHaveNoMutableLuOrZhuanWangGlobals(t *testing.T) {
	checks := map[string][]string{
		"lu_profile.go": {
			"var luDayStemOrder", "var luBranchOrder", "var luShenZhi",
		},
		"body_strength.go": {
			"var bodyStrengthDayStemOrder", "var bodyStrengthLuBranches",
		},
		"pattern.go": {
			"var zhuanWangProfiles", "zhuanWangProfiles[", "luShenZhi[",
		},
		"shensha.go": {"luShenZhi["},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("%s retains mutable pattern Profile path %q", path, forbidden)
			}
		}
	}

	for path, requiredValues := range map[string][]string{
		"lu_profile.go": {"func canonicalLuProfile()", "func luBranchForStem(stem string)"},
		"pattern.go":    {"func zhuanWangProfileRegistry()", "func zhuanWangProfileForElement(element string)"},
	} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredValues {
			if !strings.Contains(string(source), required) {
				t.Errorf("%s missing immutable Profile path %q", path, required)
			}
		}
	}
}

func TestLuBranchForStemRejectsUnknownInput(t *testing.T) {
	for _, stem := range []string{"", "A", "子", "甲甲"} {
		if got, ok := luBranchForStem(stem); ok || got != "" {
			t.Errorf("luBranchForStem(%q) = %q/%v, want empty/false", stem, got, ok)
		}
	}
}

func TestPatternProfileImmutabilityMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧专旺检测器读取包级可变zhuanWangProfiles",
			"建禄与专禄读取包级可变luShenZhi",
			"身强层又在启动时复制luDayStemOrder和luBranchOrder",
			"专旺Profile每次返回独立嵌套快照",
			"canonicalLuProfile和luBranchForStem纯函数",
			"格局、神煞与身强共同消费",
			"dc08ac014295b5505a9e09d963c710644593845629d0112e9a572f24905b28d8",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
