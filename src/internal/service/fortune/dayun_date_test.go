package fortune

import (
	"testing"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestCalcDaYunInfluenceUsesExactStartDate(t *testing.T) {
	chart := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "子"},
		DaYunInfo: bazipkg.DaYunInfo{
			Calculated: true,
			StartAge:   8,
			StartAt:    "2030-12-12T07:19:00",
			Pillars: []model.Pillar{
				{Gan: "乙", Zhi: "丑"},
				{Gan: "丙", Zhi: "寅"},
			},
		},
	}
	location := time.FixedZone("CST", 8*3600)

	beforeStart := time.Date(2030, 12, 12, 7, 18, 59, 0, location)
	before := calcDaYunInfluence(chart, beforeStart, 2022)
	if before.Active || before.CurrentPillar != "" {
		t.Fatalf("before start influence = %+v, want inactive", before)
	}
	if before.StartAt != "2030-12-12T07:19:00" {
		t.Errorf("before start_at = %q", before.StartAt)
	}

	atStart := time.Date(2030, 12, 12, 7, 19, 0, 0, location)
	first := calcDaYunInfluence(chart, atStart, 2022)
	if !first.Active || first.Index != 0 || first.CurrentPillar != "乙丑" {
		t.Fatalf("at start influence = %+v, want first decade", first)
	}
	if first.StartAt != "2030-12-12T07:19:00" || first.EndAtExclusive != "2040-12-12T07:19:00" {
		t.Errorf("first period bounds = %q..%q", first.StartAt, first.EndAtExclusive)
	}
	if first.SelectionBasis != "exact_start_time_and_query_time" || first.Status != "observed" || first.InterpretationStatus != "not_adjudicated" {
		t.Errorf("first period evidence metadata = %+v", first)
	}

	beforeSecond := time.Date(2040, 12, 12, 7, 18, 59, 0, location)
	stillFirst := calcDaYunInfluence(chart, beforeSecond, 2022)
	if !stillFirst.Active || stillFirst.Index != 0 || stillFirst.CurrentPillar != "乙丑" {
		t.Fatalf("before second decade influence = %+v, want first decade", stillFirst)
	}

	atSecond := time.Date(2040, 12, 12, 7, 19, 0, 0, location)
	second := calcDaYunInfluence(chart, atSecond, 2022)
	if !second.Active || second.Index != 1 || second.CurrentPillar != "丙寅" {
		t.Fatalf("at second decade influence = %+v, want second decade", second)
	}
	if second.StartAt != "2040-12-12T07:19:00" || second.EndAtExclusive != "2050-12-12T07:19:00" {
		t.Errorf("second period bounds = %q..%q", second.StartAt, second.EndAtExclusive)
	}
	if second.SelectionBasis != "exact_start_time_and_query_time" || second.Status != "observed" || second.InterpretationStatus != "not_adjudicated" {
		t.Errorf("second period evidence metadata = %+v", second)
	}

	afterCovered := time.Date(2050, 12, 12, 7, 19, 0, 0, location)
	after := calcDaYunInfluence(chart, afterCovered, 2022)
	if after.Active || after.CurrentPillar != "" || after.Index != len(chart.DaYunInfo.Pillars) {
		t.Fatalf("after covered periods influence = %+v, want inactive", after)
	}
}

func TestCalcDaYunInfluenceIntegerFallbackRespectsCoverageBounds(t *testing.T) {
	chart := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "子"},
		DaYunInfo: bazipkg.DaYunInfo{
			StartAge: 8,
			Pillars: []model.Pillar{
				{Gan: "乙", Zhi: "丑"},
				{Gan: "丙", Zhi: "寅"},
			},
		},
	}

	before := calcDaYunInfluence(chart, time.Date(2029, 1, 1, 0, 0, 0, 0, time.UTC), 2022)
	if before.Active || before.Index != -1 || before.Status != "before_start" || before.SelectionBasis != "integer_age_fallback" {
		t.Fatalf("integer fallback before start = %+v", before)
	}
	first := calcDaYunInfluence(chart, time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC), 2022)
	if !first.Active || first.Index != 0 || first.CurrentPillar != "乙丑" || first.Status != "observed" || first.SelectionBasis != "integer_age_fallback" {
		t.Fatalf("integer fallback first period = %+v", first)
	}
	second := calcDaYunInfluence(chart, time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC), 2022)
	if !second.Active || second.Index != 1 || second.CurrentPillar != "丙寅" {
		t.Fatalf("integer fallback second period = %+v", second)
	}
	after := calcDaYunInfluence(chart, time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC), 2022)
	if after.Active || after.Index != 2 || after.Status != "after_covered_periods" || after.SelectionBasis != "integer_age_fallback" {
		t.Fatalf("integer fallback after covered periods = %+v", after)
	}
}
