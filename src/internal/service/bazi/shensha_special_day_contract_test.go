package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestShiEDaBaiExactTenDayTableAcrossSixtyCycle(t *testing.T) {
	wants := map[string]bool{
		"甲辰": true, "乙巳": true, "壬申": true, "丙申": true, "丁亥": true,
		"庚辰": true, "戊戌": true, "癸亥": true, "辛巳": true, "乙丑": true,
	}
	matched := 0
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		want := wants[day]
		if got := specialDayContainsName(day, "十恶大败"); got != want {
			t.Errorf("special day %s 十恶大败 = %v, want %v", day, got, want)
		}
		result := calcSpecialDayFixture(t, day)
		if got := hasShenShaName(result.Day, "十恶大败"); got != want {
			t.Errorf("formal day %s 十恶大败 = %v, want %v: %+v", day, got, want, result)
		}
		for _, bucket := range [][]string{result.Year, result.Month, result.Hour, result.Global} {
			if hasShenShaName(bucket, "十恶大败") {
				t.Errorf("day %s leaked 十恶大败 outside day bucket: %+v", day, result)
			}
		}
		if want {
			matched++
		}
	}
	if matched != 10 {
		t.Fatalf("十恶大败 matched days = %d, want 10", matched)
	}
}

func TestShiEDaBaiUsesYiChouInsteadOfJiChou(t *testing.T) {
	yiChou := calcSpecialDayFixture(t, "乙丑")
	if !hasShenShaName(yiChou.Day, "十恶大败") {
		t.Fatalf("乙丑 lacks 十恶大败: %+v", yiChou)
	}
	jiChou := calcSpecialDayFixture(t, "己丑")
	if hasShenShaName(jiChou.Day, "十恶大败") {
		t.Fatalf("己丑 still has 十恶大败: %+v", jiChou)
	}
}

func TestSpecialDayTableDoesNotDuplicateYangRenOrCiGuan(t *testing.T) {
	for day, names := range specialDayShenShaRules {
		for _, name := range names {
			if name == "羊刃" || name == "词馆" {
				t.Errorf("special day %s still injects canonical day-gan rule %s", day, name)
			}
		}
	}
}

func TestCanonicalYangRenDayHitsAcrossSixtyCycle(t *testing.T) {
	wants := map[string]bool{"丙午": true, "戊午": true, "壬子": true}
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		result := calcSpecialDayFixture(t, day)
		if got, want := hasShenShaName(result.Day, "羊刃"), wants[day]; got != want {
			t.Errorf("day %s 羊刃 = %v, want %v: %+v", day, got, want, result)
		}
	}
	for _, falsePositive := range []string{"己未", "癸丑"} {
		result := calcSpecialDayFixture(t, falsePositive)
		if hasShenShaName(result.Day, "羊刃") {
			t.Errorf("yin-stem day %s still produced 羊刃: %+v", falsePositive, result)
		}
	}
}

func TestFormerDayStemCiGuanShortcutIsAbsent(t *testing.T) {
	for _, day := range []string{"甲寅", "乙卯", "庚申", "辛酉"} {
		result := calcSpecialDayFixture(t, day)
		if hasShenShaName(result.Day, "词馆") {
			t.Errorf("former day-stem shortcut still marks %s as 词馆: %+v", day, result)
		}
	}
}

func TestShiEDaBaiMetadataHasLocatedEvidence(t *testing.T) {
	meta := LookupShenShaMeta("十恶大败")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("十恶大败 metadata = %+v", meta)
	}
	for _, citation := range []string{"PDF第111页", "PDF第120-121页", "乙丑", "十日", "年日细分", "另一口径"} {
		if !strings.Contains(meta.Basis, citation) {
			t.Errorf("十恶大败 basis = %q, want %q", meta.Basis, citation)
		}
	}
}

func specialDayContainsName(day, want string) bool {
	for _, name := range specialDayShenShaRules[day] {
		if name == want {
			return true
		}
	}
	return false
}

func calcSpecialDayFixture(t testing.TB, day string) ShenShaCalcResult {
	t.Helper()
	result, err := CalcShenShaByPillars(ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "辰"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
		Day: parseGongTestPillar(t, day), Hour: model.Pillar{Gan: "庚", Zhi: "午"}, Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
