package bazi

import (
	"testing"

	"bazi/internal/model"
)

func TestJiaSeDirectFourStorehouseProfileAllowsIntrinsicHiddenStems(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "戊", Zhi: "辰"}, {Gan: "己", Zhi: "丑"},
		{Gan: "戊", Zhi: "戌"}, {Gan: "己", Zhi: "未"},
	}
	got := checkZhuanWangGe(pillars)
	if got == nil || got.PatternName != "稼穑格" {
		t.Fatalf("direct four-storehouse structure = %+v", got)
	}
}

func TestJiaSeDirectFourStorehouseProfileRejectsExposedWood(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "辰"}, {Gan: "己", Zhi: "丑"},
		{Gan: "戊", Zhi: "戌"}, {Gan: "己", Zhi: "未"},
	}
	if got := checkZhuanWangGe(pillars); got != nil {
		t.Fatalf("four-storehouse structure with exposed wood matched JiaSe: %+v", got)
	}
}
