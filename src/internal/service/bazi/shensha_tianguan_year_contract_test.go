package bazi

import (
	"strings"
	"testing"

	"bazi/internal/service/data"
)

var tianGuanTargets = map[string]string{
	"甲": "未", "乙": "辰", "丙": "巳", "丁": "酉", "戊": "戌",
	"己": "卯", "庚": "亥", "辛": "申", "壬": "寅", "癸": "午",
}

func TestTianGuanExactTableUsesYearStemOnly(t *testing.T) {
	for _, yearGan := range data.Gans {
		want := tianGuanTargets[yearGan]
		got := ruleTargetsByName(yearGanShenShaRules[yearGan], "天官贵人")
		if len(got) != 1 || got[0] != want {
			t.Errorf("year stem %s 天官贵人 targets = %v, want [%s]", yearGan, got, want)
		}
		if got := ruleTargetsByName(dayGanShenShaRules[yearGan], "天官贵人"); len(got) != 0 {
			t.Errorf("day stem %s still publishes 天官贵人 shortcut: %v", yearGan, got)
		}
		for _, rules := range [][]shenShaRule{yearGanShenShaRules[yearGan], dayGanShenShaRules[yearGan]} {
			if got := ruleTargetsByName(rules, "天官"); len(got) != 0 {
				t.Errorf("stem %s still publishes obsolete 天官 name: %v", yearGan, got)
			}
		}
	}
}

func TestTianGuanFormalEntryAssignsYearStemTargetToActualPillar(t *testing.T) {
	for _, yearGan := range data.Gans {
		target := tianGuanTargets[yearGan]
		for targetIndex := 0; targetIndex < 4; targetIndex++ {
			if targetIndex == 0 && sixtyCycleIndex(yearGan, target) < 0 {
				continue
			}
			got := calcGuoYinFixture(t, yearGan, target, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "天官贵人："+target)
			if hasShenShaName(got.Global, "天官贵人") {
				t.Errorf("year stem %s target %s leaked 天官贵人 to global: %+v", yearGan, target, got)
			}
			assertShenShaNameAbsentEverywhere(t, got, "天官")
		}
	}
}

func TestTianGuanFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	for _, yearGan := range data.Gans {
		target := tianGuanTargets[yearGan]
		for _, branch := range data.Zhis {
			if branch == target {
				continue
			}
			got := calcGuoYinFixture(t, yearGan, target, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "天官贵人")
			assertShenShaNameAbsentEverywhere(t, got, "天官")
		}
	}
}

func TestTianGuanFormerRenGuiTargetsAreNegative(t *testing.T) {
	for _, tc := range []struct {
		yearGan string
		wrong   string
		want    string
	}{
		{yearGan: "壬", wrong: "丑", want: "寅"},
		{yearGan: "癸", wrong: "子", want: "午"},
	} {
		wrong := calcGuoYinFixture(t, tc.yearGan, tc.want, tc.wrong, 1)
		assertShenShaNameAbsentEverywhere(t, wrong, "天官贵人")
		positive := calcGuoYinFixture(t, tc.yearGan, tc.want, tc.want, 1)
		assertOnlyPillarBucketHas(t, positive, 1, "天官贵人："+tc.want)
	}
}

func TestTianGuanMetadataSeparatesObsoleteShortName(t *testing.T) {
	meta := LookupShenShaMeta("天官贵人")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("天官贵人 metadata status = %+v", meta)
	}
	for _, fragment := range []string{
		"生年干", "逐柱落位", "甲未、乙辰、丙巳、丁酉、戊戌、己卯、庚亥、辛申、壬寅、癸午",
		"《渊海子平》", "PDF第65页", "其法以生年干论",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("天官贵人 basis = %q, want %q", meta.Basis, fragment)
		}
	}
	obsolete := LookupShenShaMeta("天官")
	if obsolete.Status != "unregistered" || obsolete.InterpretationStatus != "not_available" {
		t.Errorf("obsolete 天官 metadata = %+v", obsolete)
	}
}
