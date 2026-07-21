package bazi

import (
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestCalculateAtUsesSecondAtJieBoundary(t *testing.T) {
	jie, err := tyme.SolarTerm{}.FromName(2022, "惊蛰")
	if err != nil {
		t.Fatalf("create solar term: %v", err)
	}
	at := jie.GetJulianDay().GetSolarTime()
	before := at.Next(-1)

	service := &BaziService{}
	beforeResult, err := service.CalculateAt(
		before.GetYear(), before.GetMonth(), before.GetDay(),
		before.GetHour(), before.GetMinute(), before.GetSecond(), "FEMALE",
	)
	if err != nil {
		t.Fatalf("CalculateAt before boundary: %v", err)
	}
	atResult, err := service.CalculateAt(
		at.GetYear(), at.GetMonth(), at.GetDay(),
		at.GetHour(), at.GetMinute(), at.GetSecond(), "FEMALE",
	)
	if err != nil {
		t.Fatalf("CalculateAt at boundary: %v", err)
	}
	if beforeResult.MonthPillar == atResult.MonthPillar {
		t.Fatalf("month pillar did not change at exact jie second: before=%+v at=%+v", beforeResult.MonthPillar, atResult.MonthPillar)
	}
	if atResult.DaYunInfo.ReferenceDeltaSeconds != 0 {
		t.Fatalf("reference delta at exact jie = %d, want 0", atResult.DaYunInfo.ReferenceDeltaSeconds)
	}
}

func TestCalculateAtSecondAffectsDaYunStart(t *testing.T) {
	service := &BaziService{}
	first, err := service.CalculateAt(2022, 3, 9, 20, 51, 0, "MALE")
	if err != nil {
		t.Fatalf("CalculateAt first: %v", err)
	}
	second, err := service.CalculateAt(2022, 3, 9, 20, 51, 1, "MALE")
	if err != nil {
		t.Fatalf("CalculateAt second: %v", err)
	}
	if first.DaYunInfo.StartAt == second.DaYunInfo.StartAt {
		t.Fatalf("start_at ignored birth second: both=%q", first.DaYunInfo.StartAt)
	}
}
