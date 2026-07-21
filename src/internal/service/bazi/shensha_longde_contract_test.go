package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestUnsupportedLongDeShortcutIsAbsentFromProductionSource(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"addLongDe", "龙德"} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production shen-sha source still contains unsupported shortcut %q", forbidden)
		}
	}
}

func TestRemovedLongDeMonthGateWasTautological(t *testing.T) {
	for i, target := range data.MonthShenMap[data.TianDe] {
		if target == "" {
			t.Fatalf("month %s has empty 天德 target; old gate was not tautological", data.Zhis[i])
		}
	}
}

func TestSameSanHeYearAndDayDoNotCreateLongDe(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pillars ShenShaPillars
	}{
		{
			name: "申子辰",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
				Day: model.Pillar{Gan: "壬", Zhi: "辰"}, Hour: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: "MALE",
			},
		},
		{
			name: "寅午戌",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "丙", Zhi: "寅"}, Month: model.Pillar{Gan: "戊", Zhi: "寅"},
				Day: model.Pillar{Gan: "甲", Zhi: "午"}, Hour: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: "FEMALE",
			},
		},
		{
			name: "亥卯未",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "乙", Zhi: "亥"}, Month: model.Pillar{Gan: "戊", Zhi: "寅"},
				Day: model.Pillar{Gan: "丁", Zhi: "卯"}, Hour: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: "MALE",
			},
		},
		{
			name: "巳酉丑",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "乙", Zhi: "巳"}, Month: model.Pillar{Gan: "戊", Zhi: "寅"},
				Day: model.Pillar{Gan: "乙", Zhi: "丑"}, Hour: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: "FEMALE",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CalcShenShaByPillars(tc.pillars)
			if err != nil {
				t.Fatal(err)
			}
			for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
				if hasShenShaName(bucket, "龙德") {
					t.Fatalf("same-trine year/day produced unsupported 龙德: %+v", got)
				}
			}
		})
	}
}

func TestRemovedLongDeRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("龙德")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("removed 龙德 metadata = %+v", meta)
	}
}
