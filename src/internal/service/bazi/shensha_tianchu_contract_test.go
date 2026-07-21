package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var tianChuTargets = map[string]string{
	"甲": "巳", "乙": "午", "丙": "巳", "丁": "午", "戊": "申",
	"己": "酉", "庚": "亥", "辛": "子", "壬": "寅", "癸": "卯",
}

func TestTianChuExactDayStemTableAcrossAllBranches(t *testing.T) {
	for _, dayGan := range data.Gans {
		var gotTargets []string
		for _, rule := range dayGanShenShaRules[dayGan] {
			if rule.Name == "天厨贵人" {
				gotTargets = append(gotTargets, rule.Target)
			}
			if rule.Name == "天厨食禄" {
				t.Errorf("day stem %s still contains obsolete mixed name: %+v", dayGan, rule)
			}
		}
		wantTarget := tianChuTargets[dayGan]
		if len(gotTargets) != 1 || gotTargets[0] != wantTarget {
			t.Errorf("day stem %s 天厨贵人 targets = %v, want [%s]", dayGan, gotTargets, wantTarget)
			continue
		}
		for _, branch := range data.Zhis {
			want := branch == wantTarget
			if got := targetContainsZhi(gotTargets[0], branch); got != want {
				t.Errorf("day stem %s branch %s 天厨贵人 = %v, want %v", dayGan, branch, got, want)
			}
		}
	}
}

func TestTianChuFormalEntryAssignsTargetToActualPillar(t *testing.T) {
	for _, dayGan := range data.Gans {
		target := tianChuTargets[dayGan]
		for targetIndex := 0; targetIndex < 4; targetIndex++ {
			if targetIndex == 2 && sixtyCycleIndex(dayGan, target) < 0 {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "天厨贵人："+target)
			if hasShenShaName(got.Global, "天厨贵人") {
				t.Errorf("day stem %s target %s leaked 天厨贵人 to global: %+v", dayGan, target, got)
			}
			assertShenShaNameAbsentEverywhere(t, got, "天厨食禄")
		}
	}
}

func TestTianChuFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	for _, dayGan := range data.Gans {
		target := tianChuTargets[dayGan]
		for _, branch := range data.Zhis {
			if branch == target {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, branch, 0)
			assertShenShaNameAbsentEverywhere(t, got, "天厨贵人")
			assertShenShaNameAbsentEverywhere(t, got, "天厨食禄")
		}
	}
}

func TestTianChuCurrentProfileDoesNotUseYearStemAsRuleKey(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "子"},
		Month:  model.Pillar{Gan: "己", Zhi: "巳"},
		Day:    model.Pillar{Gan: "乙", Zhi: "丑"},
		Hour:   model.Pillar{Gan: "丙", Zhi: "戌"},
		Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertShenShaNameAbsentEverywhere(t, got, "天厨贵人")
}

func TestTianChuMetadataSeparatesObsoleteMixedName(t *testing.T) {
	meta := LookupShenShaMeta("天厨贵人")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("天厨贵人 metadata status = %+v", meta)
	}
	for _, fragment := range []string{
		"当前Profile只按日干", "逐柱落位", "甲丙巳、乙丁午、戊申、己酉、庚亥、辛子、壬寅、癸卯",
		"《渊海子平》", "PDF第76页", "年干主键口径待裁决", "PDF第91页十干食禄",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("天厨贵人 basis = %q, want %q", meta.Basis, fragment)
		}
	}

	obsolete := LookupShenShaMeta("天厨食禄")
	if obsolete.Status != "unregistered" || obsolete.InterpretationStatus != "not_available" {
		t.Errorf("obsolete 天厨食禄 metadata = %+v", obsolete)
	}
}
