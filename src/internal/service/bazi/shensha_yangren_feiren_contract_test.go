package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var yangRenFiveYangTargets = map[string]string{
	"甲": "卯", "丙": "午", "戊": "午", "庚": "酉", "壬": "子",
}

var feiRenFiveYangTargets = map[string]string{
	"甲": "酉", "丙": "子", "戊": "子", "庚": "卯", "壬": "午",
}

func TestYangRenAndFeiRenExactFiveYangProfile(t *testing.T) {
	if len(feiRenByGan) != len(feiRenFiveYangTargets) {
		t.Fatalf("feiRenByGan = %v, want five-yang profile", feiRenByGan)
	}
	for _, gan := range data.Gans {
		wantYangRen := yangRenFiveYangTargets[gan]
		gotYangRen := yangRenRuleTargets(dayGanShenShaRules[gan])
		if wantYangRen == "" {
			if len(gotYangRen) != 0 {
				t.Errorf("yin stem %s 羊刃 targets = %v, want none", gan, gotYangRen)
			}
		} else if len(gotYangRen) != 1 || gotYangRen[0] != wantYangRen {
			t.Errorf("stem %s 羊刃 targets = %v, want [%s]", gan, gotYangRen, wantYangRen)
		}

		wantFeiRen := feiRenFiveYangTargets[gan]
		if gotFeiRen := feiRenByGan[gan]; gotFeiRen != wantFeiRen {
			t.Errorf("stem %s 飞刃 target = %q, want %q", gan, gotFeiRen, wantFeiRen)
		}
		if wantYangRen != "" && zhiLiuChong[wantYangRen] != wantFeiRen {
			t.Errorf("stem %s 羊刃=%s 飞刃=%s, want exact clash", gan, wantYangRen, wantFeiRen)
		}
	}
}

func TestYangRenAndFeiRenFormalEntryTruthTable(t *testing.T) {
	for _, dayGan := range data.Gans {
		for _, candidate := range data.Zhis {
			got := calcYangRenFeiRenFixture(t, dayGan, candidate)
			assertYangRenFeiRenBucketTruth(t, dayGan, candidate, got)
		}
	}
}

func TestYangRenAndFeiRenMetadataRecordsProfileConflict(t *testing.T) {
	checks := map[string][]string{
		"羊刃": {"只以日干", "甲卯、丙午、戊午、庚酉、壬子", "PDF第205页", "《三命通会》PDF第108页及第226页", "PDF第119页另列十干羊刃表", "不混入五阴目标"},
		"飞刃": {"五阳干羊刃目标的六冲支", "甲酉、丙子、戊子、庚卯、壬午", "PDF第205页", "《三命通会》PDF第108页及第226页", "PDF第119页另列十干飞刃表", "不混入五阴目标"},
	}
	for name, fragments := range checks {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("%s metadata = %+v", name, meta)
		}
		for _, fragment := range fragments {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}
}

func yangRenRuleTargets(rules []shenShaRule) []string {
	targets := make([]string, 0, 1)
	for _, rule := range rules {
		if rule.Name == "羊刃" {
			targets = append(targets, rule.Target)
		}
	}
	return targets
}

func calcYangRenFeiRenFixture(t testing.TB, dayGan, candidate string) ShenShaCalcResult {
	t.Helper()
	forbidden := map[string]bool{
		candidate:                      true,
		yangRenFiveYangTargets[dayGan]: true,
		feiRenFiveYangTargets[dayGan]:  true,
	}
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   yangRenNeutralPillar(t, forbidden, 0),
		Month:  yangRenNeutralPillar(t, forbidden, 11),
		Day:    yangRenDayPillar(t, dayGan, forbidden),
		Hour:   yangRenPillarForBranch(t, candidate),
		Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func assertYangRenFeiRenBucketTruth(t testing.TB, dayGan, candidate string, got ShenShaCalcResult) {
	t.Helper()
	wantHour := map[string]bool{
		"羊刃": candidate == yangRenFiveYangTargets[dayGan] && yangRenFiveYangTargets[dayGan] != "",
		"飞刃": candidate == feiRenFiveYangTargets[dayGan] && feiRenFiveYangTargets[dayGan] != "",
	}
	for name, want := range wantHour {
		if actual := hasShenShaName(got.Hour, name); actual != want {
			t.Errorf("day stem %s candidate %s hour %s = %v, want %v: %+v", dayGan, candidate, name, actual, want, got)
		}
		for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Global} {
			if hasShenShaName(bucket, name) {
				t.Errorf("day stem %s candidate %s leaked %s outside hour: %+v", dayGan, candidate, name, got)
			}
		}
	}
}

func yangRenNeutralPillar(t testing.TB, forbidden map[string]bool, offset int) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		index := (i + offset) % 60
		zhi := data.Zhis[index%12]
		if !forbidden[zhi] {
			return model.Pillar{Gan: data.Gans[index%10], Zhi: zhi}
		}
	}
	t.Fatal("no neutral pillar available")
	return model.Pillar{}
}

func yangRenDayPillar(t testing.TB, gan string, forbidden map[string]bool) model.Pillar {
	t.Helper()
	for _, zhi := range data.Zhis {
		if !forbidden[zhi] && sixtyCycleIndex(gan, zhi) >= 0 {
			return model.Pillar{Gan: gan, Zhi: zhi}
		}
	}
	t.Fatalf("no neutral day pillar available for stem %s", gan)
	return model.Pillar{}
}

func yangRenPillarForBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for _, gan := range data.Gans {
		if sixtyCycleIndex(gan, zhi) >= 0 {
			return model.Pillar{Gan: gan, Zhi: zhi}
		}
	}
	t.Fatalf("no valid pillar available for branch %s", zhi)
	return model.Pillar{}
}
