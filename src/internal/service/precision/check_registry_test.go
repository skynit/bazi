package precision

import (
	"strings"
	"testing"
)

func TestCheckRegistryHasStableAuditableRules(t *testing.T) {
	if len(checkRegistry) == 0 {
		t.Fatal("check registry is empty")
	}
	for id, rule := range checkRegistry {
		if strings.TrimSpace(string(id)) == "" || rule.Mode == "" {
			t.Fatalf("invalid registered rule %q: %+v", id, rule)
		}
		if rule.Publishable && (rule.Mode == compareSetMember || rule.Mode == compareRubric) {
			t.Fatalf("non-adjudicating comparison is publishable: %q %+v", id, rule)
		}
		if rule.Mode == compareTolerance && rule.Tolerance < 0 {
			t.Fatalf("negative registered tolerance: %q %+v", id, rule)
		}
	}
	hash := checkRegistryHash()
	if !strings.HasPrefix(hash, "sha256:") || len(hash) != len("sha256:")+64 {
		t.Fatalf("invalid registry hash: %q", hash)
	}
}

func TestCompareFieldCheckSetEqualityIsOrderInsensitiveButCountSensitive(t *testing.T) {
	rule := checkRule{Mode: compareSetEqual}
	matched, evaluable := compareFieldCheck(rule, fieldCheck{
		wantSet: []string{"紫微", "天府"},
		gotSet:  []string{"天府", "紫微"},
	})
	if !evaluable || !matched {
		t.Fatal("same set in a different order did not match")
	}
	matched, evaluable = compareFieldCheck(rule, fieldCheck{
		wantSet: []string{"紫微", "天府"},
		gotSet:  []string{"紫微", "天府", "天府"},
	})
	if !evaluable || matched {
		t.Fatal("duplicate set member was silently ignored")
	}
	matched, evaluable = compareFieldCheck(rule, fieldCheck{
		wantSet: []string{"紫微"},
		gotSet:  []string{" 紫微 "},
	})
	if !evaluable || matched {
		t.Fatal("set equality silently trimmed a Gold label")
	}
}

func TestCompareFieldCheckExactDoesNotNormalizeWhitespace(t *testing.T) {
	matched, evaluable := compareFieldCheck(checkRule{Mode: compareExact}, fieldCheck{want: "庚午", got: " 庚午 "})
	if !evaluable || matched {
		t.Fatal("exact comparison silently normalized whitespace")
	}
}

func TestCompareFieldCheckToleranceUsesRegisteredThreshold(t *testing.T) {
	rule := checkRule{Mode: compareTolerance, Tolerance: 0.01}
	matched, evaluable := compareFieldCheck(rule, fieldCheck{want: "1.000", got: "1.009"})
	if !evaluable || !matched {
		t.Fatal("value within tolerance did not match")
	}
	matched, evaluable = compareFieldCheck(rule, fieldCheck{want: "1.000", got: "1.011"})
	if !evaluable || matched {
		t.Fatal("value outside tolerance matched")
	}
}

func TestCompareFieldCheckRubricRequiresExternalAdjudication(t *testing.T) {
	matched, evaluable := compareFieldCheck(checkRule{Mode: compareRubric}, fieldCheck{want: "supported", got: "supported"})
	if evaluable || matched {
		t.Fatal("rubric comparison was treated as an automatic exact match")
	}
}
