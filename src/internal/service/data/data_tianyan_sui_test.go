package data

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFindMissingElementsReportsFactsWithoutRemedyConclusion(t *testing.T) {
	tests := []struct {
		name        string
		scores      map[string]int
		wantMissing []string
		wantWeak    []string
	}{
		{
			name:        "缺木不自动补木或补水",
			scores:      map[string]int{"木": 0, "火": 12, "土": 18, "金": 14, "水": 9},
			wantMissing: []string{"木"},
			wantWeak:    []string{},
		},
		{
			name:        "五行均有得分",
			scores:      map[string]int{"木": 5, "火": 6, "土": 7, "金": 8, "水": 9},
			wantMissing: []string{},
			wantWeak:    []string{},
		},
		{
			name:        "偏低不等于缺失",
			scores:      map[string]int{"木": 1, "火": 4, "土": 5, "金": 6, "水": 7},
			wantMissing: []string{},
			wantWeak:    []string{"木", "火"},
		},
		{
			name:        "输出按固定五行顺序",
			scores:      map[string]int{"水": 0, "金": 2, "土": 0, "火": 3, "木": 0},
			wantMissing: []string{"木", "土", "水"},
			wantWeak:    []string{"火", "金"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FindMissingElements(tc.scores)
			if got.Status != "observed" || got.RuleID != "wuxing.raw-score-presence-v1" {
				t.Fatalf("unstable observation metadata: %+v", got)
			}
			if !reflect.DeepEqual(got.MissingElements, tc.wantMissing) || !reflect.DeepEqual(got.WeakElements, tc.wantWeak) {
				t.Fatalf("FindMissingElements() missing=%v weak=%v", got.MissingElements, got.WeakElements)
			}
			if got.MissingCount != len(tc.wantMissing) || got.IsYongshenConclusion || got.RemedyStatus != "not_adjudicated" {
				t.Fatalf("distribution was promoted to an adjudicated conclusion: %+v", got)
			}
			if got.MissingElements == nil || got.WeakElements == nil || got.Scores == nil {
				t.Fatalf("JSON collections must remain non-null: %+v", got)
			}
			if !reflect.DeepEqual(got.Scores, tc.scores) {
				t.Fatalf("normalized scores = %v, want %v", got.Scores, tc.scores)
			}
			payload, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(payload), "remedy_elements") || strings.Contains(string(payload), "remedy_advice") || strings.Contains(string(payload), "severity") {
				t.Fatalf("legacy remedy fields leaked into JSON: %s", payload)
			}
		})
	}
}
