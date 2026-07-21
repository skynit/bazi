package bazi

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestShenShaMetadataContainsOnlyRuleEvidence(t *testing.T) {
	names := make([]string, 0, len(shenShaBasisCatalog))
	for name := range shenShaBasisCatalog {
		names = append(names, name)
	}

	wantKeys := []string{"basis", "interpretation_status", "name", "rule_id", "status"}
	for _, name := range names {
		meta := LookupShenShaMeta(name)
		if meta.Name != name || meta.RuleID == "" || meta.Basis == "" ||
			meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("invalid factual metadata for %q: %+v", name, meta)
		}
		payload, err := json.Marshal(meta)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatal(err)
		}
		gotKeys := make([]string, 0, len(fields))
		for key := range fields {
			gotKeys = append(gotKeys, key)
		}
		sort.Strings(gotKeys)
		if !reflect.DeepEqual(gotKeys, wantKeys) {
			t.Fatalf("metadata fields for %q = %v, want %v: %s", name, gotKeys, wantKeys, payload)
		}
	}

	unregistered := LookupShenShaMeta("未收录测试名")
	if unregistered.Status != "unregistered" || unregistered.InterpretationStatus != "not_available" ||
		unregistered.Basis != "未登记可审计查法依据" {
		t.Fatalf("unregistered shen-sha was disguised as rule evidence: %+v", unregistered)
	}
}

func TestCalculatedShenShaMetadataHasNoRegistrationGaps(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	details := append([]ShenShaMeta{}, result.DayShenShaDetails...)
	details = append(details, result.GlobalShenShaDetails...)
	for _, pillar := range result.ShenShaByPillar {
		details = append(details, pillar.Details...)
	}
	registeredCount, unregisteredCount := 0, 0
	for _, detail := range details {
		_, registered := shenShaBasisCatalog[detail.Name]
		if registered {
			registeredCount++
			if detail.Status != "observed" || detail.InterpretationStatus != "not_adjudicated" || detail.Basis == "未登记可审计查法依据" {
				t.Fatalf("registered shen-sha metadata = %+v", detail)
			}
			continue
		}
		unregisteredCount++
		if detail.Status != "unregistered" || detail.InterpretationStatus != "not_available" || detail.Basis != "未登记可审计查法依据" {
			t.Fatalf("unregistered shen-sha metadata = %+v", detail)
		}
	}
	if registeredCount == 0 || unregisteredCount != 0 {
		t.Fatalf("calculated shen-sha metadata registration coverage: registered=%d unregistered=%d", registeredCount, unregisteredCount)
	}

	items := []string{"攀鞍：丑", "未收录测试名：子"}
	contract := BuildShenShaDetails(items)
	if !ValidShenShaDetails(items, contract) {
		t.Fatal("canonical mixed registration contract must validate")
	}
	contract[1].Status = "observed"
	if ValidShenShaDetails(items, contract) {
		t.Fatal("unregistered rule status tampering passed validation")
	}
}

func TestPillarShenShaUsesNaturalOrderWithoutRoleRanking(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	wantOrder := []string{"year", "month", "day", "hour"}
	if len(result.ShenShaByPillar) != len(wantOrder) {
		t.Fatalf("pillar groups = %d, want %d", len(result.ShenShaByPillar), len(wantOrder))
	}
	for i, want := range wantOrder {
		if result.ShenShaByPillar[i].Pillar != want {
			t.Fatalf("pillar group %d = %q, want %q", i, result.ShenShaByPillar[i].Pillar, want)
		}
		payload, err := json.Marshal(result.ShenShaByPillar[i])
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]any
		if err := json.Unmarshal(payload, &fields); err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"priority", "role"} {
			if _, ok := fields[forbidden]; ok {
				t.Fatalf("pillar group leaked %q: %s", forbidden, payload)
			}
		}
	}
}
