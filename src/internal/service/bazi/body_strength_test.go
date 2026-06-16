package bazi

import "testing"

func TestDefaultRuleMetaLoadsVersionedRuleTables(t *testing.T) {
	meta := DefaultRuleMeta()
	if meta.RuleVersion != RuleVersion {
		t.Fatalf("rule version = %q, want %q", meta.RuleVersion, RuleVersion)
	}
	if meta.School != RuleSchool {
		t.Fatalf("school = %q, want %q", meta.School, RuleSchool)
	}
	if len(meta.Tables) == 0 {
		t.Fatal("expected rule tables")
	}
	if meta.BodyStrength.Weights.Ling != 0.4 || meta.BodyStrength.Weights.Di != 0.3 {
		t.Fatalf("unexpected body strength weights: %+v", meta.BodyStrength.Weights)
	}

	var foundTenGod, foundBodyStrength bool
	for _, table := range meta.Tables {
		if table.Key == "ten_god_matrix" {
			foundTenGod = true
			if table.Count != len(tenGodNames) {
				t.Fatalf("ten god count = %d, want %d", table.Count, len(tenGodNames))
			}
		}
		if table.Key == "body_strength" {
			foundBodyStrength = true
		}
	}
	if !foundTenGod || !foundBodyStrength {
		t.Fatalf("missing expected rule tables: tenGod=%v bodyStrength=%v", foundTenGod, foundBodyStrength)
	}
}

func TestDefaultRuleMetaReturnsIndependentCopy(t *testing.T) {
	meta := DefaultRuleMeta()
	if len(meta.Tables) == 0 {
		t.Fatal("expected rule tables")
	}
	meta.Tables[0].Name = "mutated"

	next := DefaultRuleMeta()
	if next.Tables[0].Name == "mutated" {
		t.Fatal("DefaultRuleMeta leaked caller mutation")
	}
}

func TestBodyStrengthReturnsExplainableComponents(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}

	bs := result.BodyStrength
	if bs.RuleVersion != RuleVersion || bs.School != RuleSchool {
		t.Fatalf("unexpected rule meta on body strength: %+v", bs)
	}
	if len(bs.Components) < 5 {
		t.Fatalf("expected components including bonus, got %d", len(bs.Components))
	}
	if len(bs.Evidence) == 0 {
		t.Fatal("expected body strength evidence")
	}
	if bs.Summary == "" {
		t.Fatal("expected summary")
	}

	wantKeys := map[string]bool{"ling": false, "di": false, "shi": false, "sheng": false, "bonus": false}
	for _, c := range bs.Components {
		if _, ok := wantKeys[c.Key]; ok {
			wantKeys[c.Key] = true
		}
		if c.Weight <= 0 {
			t.Fatalf("component %s has non-positive weight %.2f", c.Key, c.Weight)
		}
	}
	for key, found := range wantKeys {
		if !found {
			t.Fatalf("missing body strength component %q", key)
		}
	}
}
