package bazi

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestAnalyzePatternCandidatesDoesNotCreateFallbackAfterRetirements(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "辰"},
		{Gan: "己", Zhi: "巳"},
		{Gan: "甲", Zhi: "子"},
		{Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "巳")

	if analysis.SchemaVersion != PatternSchemaVersion {
		t.Fatalf("schema version = %q", analysis.SchemaVersion)
	}
	if analysis.RuleID != "bazi.pattern-candidate-set-v34" ||
		analysis.DetectorProfile != PatternDetectorProfile ||
		analysis.DetectorCount != patternDetectorCount() ||
		analysis.ValidationStatus != "not_validated" ||
		analysis.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("pattern contract metadata is incomplete: %+v", analysis)
	}
	wantInputs := patternInputSnapshot(pillars, "巳")
	if !reflect.DeepEqual(analysis.Inputs, wantInputs) {
		t.Fatalf("inputs = %+v, want %+v", analysis.Inputs, wantInputs)
	}
	if len(analysis.Candidates) != 0 || analysis.Status != "observed_without_structural_candidate" {
		t.Fatalf("retired shortcuts created candidates or fallback: %+v", analysis)
	}
	if !ValidPatternAnalysis(analysis, pillars, "巳") {
		t.Fatalf("generated pattern analysis did not pass strict validation: %+v", analysis)
	}
}

func TestPatternAnalysisJSONExposesOnlyCandidateEvidence(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"},
		{Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"},
		{Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(payload, &topLevel); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"pattern_name", "pattern_type", "description", "favorable_elements", "unfavorable_elements", "sub_type",
		"has_dispute", "dispute_reasons", "selection_basis", "primary_candidate_id", "candidate_id",
	} {
		if _, ok := topLevel[forbidden]; ok {
			t.Fatalf("top-level pattern response leaked legacy field %q: %s", forbidden, payload)
		}
	}
	for _, forbidden := range []string{
		"\"description\"", "\"favorable_elements\"", "\"unfavorable_elements\"", "\"sub_type\"",
		"\"element_scores\"", "\"body_strength_score_band_candidate\"",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("pattern response leaked legacy field %s: %s", forbidden, payload)
		}
	}
	var candidatePayloads []map[string]json.RawMessage
	if err := json.Unmarshal(topLevel["candidates"], &candidatePayloads); err != nil {
		t.Fatalf("decode candidate payloads: %v", err)
	}
	allowedCandidateFields := map[string]bool{
		"rule_id": true, "pattern_name": true,
		"category": true, "source": true,
	}
	for _, candidate := range candidatePayloads {
		for field := range candidate {
			if !allowedCandidateFields[field] {
				t.Fatalf("pattern candidate leaked non-contract field %q: %s", field, payload)
			}
		}
	}

	var decoded PatternAnalysis
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if !ValidPatternAnalysis(decoded, pillars, "寅") {
		t.Fatalf("serialized pattern response did not pass strict validation: %+v", decoded)
	}
	tampered := decoded
	tampered.Candidates = append([]PatternCandidate(nil), decoded.Candidates...)
	tampered.Candidates[0].PatternName = "篡改格"
	if ValidPatternAnalysis(tampered, pillars, "巳") {
		t.Fatal("tampered candidate must not pass strict validation")
	}
}

func TestAnalyzePatternCandidatesKeepsAuxiliaryFeatureWithoutFallback(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "癸", Zhi: "亥"},
		{Gan: "甲", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"},
		{Gan: "丁", Zhi: "卯"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")

	findPatternCandidate(t, analysis.Candidates, "日德格")
	for _, candidate := range analysis.Candidates {
		if candidate.RuleID == "pattern.normal.fallback" {
			t.Fatalf("absence of another detector match created fallback candidate: %+v", candidate)
		}
	}
}

func TestMonthCommandPatternEvidenceDistinguishesExposureStrength(t *testing.T) {
	cases := []struct {
		name           string
		pillars        []model.Pillar
		month          string
		hiddenStem     string
		candidate      string
		exposureStatus string
		special        string
	}{
		{
			name: "午中己土透年干为伤官候选",
			pillars: []model.Pillar{
				{Gan: "己", Zhi: "巳"}, {Gan: "庚", Zhi: "午"},
				{Gan: "丙", Zhi: "午"}, {Gan: "甲", Zhi: "午"},
			},
			month: "午", hiddenStem: "己", candidate: "伤官格",
			exposureStatus: "exact_hidden_stem_exposed", special: "月刃",
		},
		{
			name: "卯中乙木透月干为正印候选",
			pillars: []model.Pillar{
				{Gan: "癸", Zhi: "未"}, {Gan: "乙", Zhi: "卯"},
				{Gan: "丙", Zhi: "午"}, {Gan: "辛", Zhi: "卯"},
			},
			month: "卯", hiddenStem: "乙", candidate: "正印格",
			exposureStatus: "exact_hidden_stem_exposed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.month)
			var found *MonthCommandPatternEvidence
			for index := range analysis.MonthCommandEvidence {
				if analysis.MonthCommandEvidence[index].HiddenStem == tc.hiddenStem {
					found = &analysis.MonthCommandEvidence[index]
					break
				}
			}
			if found == nil {
				t.Fatalf("month-command evidence for %s not found: %+v", tc.hiddenStem, analysis.MonthCommandEvidence)
			}
			if found.ExposureStatus != tc.exposureStatus || found.MonthSpecialStructure != tc.special ||
				found.IsEstablishedPattern || found.InterpretationStatus != "pattern_candidate_not_adjudicated" {
				t.Fatalf("month-command evidence = %+v", *found)
			}
			if !containsPatternString(found.CandidateNames, tc.candidate) {
				t.Fatalf("candidate %s missing from %+v", tc.candidate, *found)
			}
		})
	}
}

func TestMonthCommandPatternEvidenceRequiresExactHiddenStem(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "壬", Zhi: "申"}, {Gan: "癸", Zhi: "巳"},
		{Gan: "甲", Zhi: "子"}, {Gan: "乙", Zhi: "丑"},
	}
	analysis := AnalyzePatternExtended(pillars, "巳")
	if len(analysis.MonthCommandEvidence) != 0 {
		t.Fatalf("unexposed month hidden stems created pattern evidence: %+v", analysis.MonthCommandEvidence)
	}

	counterpartOnly := []model.Pillar{
		{Gan: "辛", Zhi: "卯"}, {Gan: "丙", Zhi: "申"},
		{Gan: "癸", Zhi: "卯"}, {Gan: "甲", Zhi: "寅"},
	}
	analysis = AnalyzePatternExtended(counterpartOnly, "申")
	for _, evidence := range analysis.MonthCommandEvidence {
		if evidence.HiddenStem == "庚" {
			t.Fatalf("辛金 must not be treated as 申中庚金透干: %+v", evidence)
		}
	}
}

func containsPatternString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestAnalyzePatternCandidatesRejectsInvalidInputsWithoutFallback(t *testing.T) {
	validPillars := []model.Pillar{
		{Gan: "丙", Zhi: "辰"},
		{Gan: "己", Zhi: "巳"},
		{Gan: "甲", Zhi: "子"},
		{Gan: "戊", Zhi: "辰"},
	}
	tests := []struct {
		name    string
		pillars []model.Pillar
		month   string
	}{
		{name: "missing pillars", pillars: nil, month: "巳"},
		{name: "invalid sixty cycle", pillars: []model.Pillar{{Gan: "甲", Zhi: "丑"}, validPillars[1], validPillars[2], validPillars[3]}, month: "巳"},
		{name: "month mismatch", pillars: validPillars, month: "午"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.month)
			if analysis.Status != "invalid_input" || analysis.ValidationStatus != "invalid_input" || len(analysis.Candidates) != 0 {
				t.Fatalf("invalid input produced pattern evidence: %+v", analysis)
			}
			if ValidPatternAnalysis(analysis, tc.pillars, tc.month) {
				t.Fatal("invalid input passed persisted pattern validation")
			}
		})
	}
}

func findPatternCandidate(t *testing.T, candidates []PatternCandidate, name string) PatternCandidate {
	t.Helper()
	for _, candidate := range candidates {
		if candidate.PatternName == name {
			return candidate
		}
	}
	t.Fatalf("candidate %q not found in %+v", name, candidates)
	return PatternCandidate{}
}
