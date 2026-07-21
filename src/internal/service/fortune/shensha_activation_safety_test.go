package fortune

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestShenShaActivationIsUnscoredRuleEvidence(t *testing.T) {
	chart := &bazipkg.BaziResult{DayPillar: model.Pillar{Gan: "甲", Zhi: "午"}}
	items := calcShenShaActivation("丙", "卯", chart)
	if len(items) == 0 {
		t.Fatal("expected a rule activation")
	}
	wantKeys := []string{"activation", "basis", "interpretation_status", "name", "rule_id", "status"}
	for _, item := range items {
		if item.Status != "observed" || item.InterpretationStatus != "not_adjudicated" ||
			item.RuleID == "" || item.Basis == "" || item.Activation == "" {
			t.Fatalf("invalid activation evidence: %+v", item)
		}
		payload, err := json.Marshal(item)
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
			t.Fatalf("activation fields = %v, want %v: %s", gotKeys, wantKeys, payload)
		}
		for _, forbidden := range []string{"吉神", "凶煞", "逢凶化吉", "人缘旺盛", "事业财运有助力"} {
			if strings.Contains(string(payload), forbidden) {
				t.Fatalf("activation inferred %q: %s", forbidden, payload)
			}
		}
	}
}

func TestShenShaActivationUsesYearAndDaySanHeKeys(t *testing.T) {
	chart := &bazipkg.BaziResult{
		YearPillar: model.Pillar{Zhi: "子"},
		DayPillar:  model.Pillar{Gan: "甲", Zhi: "午"},
	}
	for _, tc := range []struct {
		name, queryBranch, wantName, wantReference string
	}{
		{name: "年支驿马", queryBranch: "寅", wantName: "驿马", wantReference: "年支子"},
		{name: "日支驿马", queryBranch: "申", wantName: "驿马", wantReference: "日支午"},
		{name: "年支咸池", queryBranch: "酉", wantName: "咸池", wantReference: "年支子"},
		{name: "日支咸池", queryBranch: "卯", wantName: "咸池", wantReference: "日支午"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := calcShenShaActivation("丙", tc.queryBranch, chart)
			matches := 0
			for _, item := range items {
				if item.Name == "桃花" {
					t.Fatalf("运势神煞仍输出已停用别名: %+v", item)
				}
				if item.Name != tc.wantName {
					continue
				}
				matches++
				if !strings.Contains(item.Activation, tc.wantReference) || item.Status != "observed" ||
					item.InterpretationStatus != "not_adjudicated" {
					t.Fatalf("%s激活依据不完整: %+v", tc.wantName, item)
				}
			}
			if matches != 1 {
				t.Fatalf("%s命中数 = %d, want 1: %+v", tc.wantName, matches, items)
			}
		})
	}
}

func TestShenShaActivationMergesSameSanHeGroupReferences(t *testing.T) {
	chart := &bazipkg.BaziResult{
		YearPillar: model.Pillar{Zhi: "子"},
		DayPillar:  model.Pillar{Gan: "甲", Zhi: "辰"},
	}
	items := calcShenShaActivation("丙", "寅", chart)
	matches := 0
	for _, item := range items {
		if item.Name != "驿马" {
			continue
		}
		matches++
		if !strings.Contains(item.Activation, "年支子") || !strings.Contains(item.Activation, "日支辰") {
			t.Fatalf("同组三合主键未合并: %+v", item)
		}
	}
	if matches != 1 {
		t.Fatalf("同组三合驿马输出数 = %d, want 1: %+v", matches, items)
	}
}
