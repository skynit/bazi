package bazi_test

import (
	"strings"
	"testing"

	. "bazi/internal/service/bazi"
)

func TestLegacyExtendedBaziFixtureRemainsQuarantined(t *testing.T) {
	data, err := loadTestData("../../testdata/classical_cases_extended.json")
	if err != nil {
		t.Fatal(err)
	}
	assertLegacyFixtureQuarantined(t, data.Metadata)
	if len(data.Cases) == 0 {
		t.Fatal("legacy pillar fixture unexpectedly contains no cases")
	}

	service := &BaziService{}
	evaluated := 0
	ids := make(map[string]bool, len(data.Cases))
	for index, tc := range data.Cases {
		if tc.ID == "" || ids[tc.ID] || tc.LegacyID == "" || tc.SourceIndex != index+1 || strings.TrimSpace(tc.LegacyAnnotations.Description) == "" {
			t.Fatalf("extended fixture provenance is incomplete at index %d: %+v", index, tc)
		}
		ids[tc.ID] = true
		pillars := []string{tc.Expected.YearPillar, tc.Expected.MonthPillar, tc.Expected.DayPillar, tc.Expected.HourPillar}
		complete := true
		for i := range pillars {
			pillars[i], complete = canonicalLegacyPillar(pillars[i])
			if !complete {
				break
			}
		}
		if !complete {
			continue
		}
		gender := tc.Gender
		if gender == "" {
			gender = "MALE"
		}
		result, err := service.CalculateSyntheticPillars(pillars[0], pillars[1], pillars[2], pillars[3], gender)
		if err != nil {
			t.Fatalf("Bronze pillar smoke calculation %s failed: %v", tc.ID, err)
		}
		if result.Tiaohou == nil || result.Tiaohou.DepthEvidence.Status != "unavailable" {
			t.Fatalf("pillar-only calculation %s must not invent month-depth evidence", tc.ID)
		}
		evaluated++
	}
	if evaluated == 0 {
		t.Fatal("legacy pillar fixture contains no complete smoke input")
	}
}

func canonicalLegacyPillar(value string) (string, bool) {
	value = strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	return value, len([]rune(value)) == 2
}
