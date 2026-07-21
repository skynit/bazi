package bazi

import (
	"testing"

	"bazi/internal/model"
)

func TestBaZhuanMatchesHourPillar(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "壬", Zhi: "辰"},
		Month:  model.Pillar{Gan: "甲", Zhi: "寅"},
		Day:    model.Pillar{Gan: "戊", Zhi: "子"},
		Hour:   model.Pillar{Gan: "甲", Zhi: "寅"},
		Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !shenShaItemsContain(got.Hour, "八专") {
		t.Fatalf("甲寅时应命中八专: %+v", got.Hour)
	}
	if shenShaItemsContain(got.Day, "八专") {
		t.Fatalf("戊子日不应命中八专: %+v", got.Day)
	}
}

func shenShaItemsContain(items []string, want string) bool {
	for _, item := range items {
		if shenShaName(item) == want {
			return true
		}
	}
	return false
}
