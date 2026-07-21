package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var tianXingHourGanTargets = map[string]string{
	"子": "乙", "丑": "乙", "寅": "庚", "卯": "辛", "辰": "辛", "巳": "壬",
	"午": "癸", "未": "癸", "申": "丙", "酉": "丁", "戌": "丁", "亥": "戊",
}

func TestTianXingShaExactBirthYearBranchToHourStemTable(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		if got := tianXingHourGanByYearZhi[yearBranch]; got != tianXingHourGanTargets[yearBranch] {
			t.Errorf("year branch %s 天刑时干=%s, want %s", yearBranch, got, tianXingHourGanTargets[yearBranch])
		}
	}
}

func TestTianXingShaFormalEntryMatchesOnlyHourStem(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		yearPillar := poZhaiPillarForBranch(t, yearBranch)
		target := tianXingHourGanTargets[yearBranch]
		targetPillar := yearGanExtraPillarForGan(t, target)
		for _, hourGan := range data.Gans {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   yearPillar,
				Month:  targetPillar,
				Day:    targetPillar,
				Hour:   yearGanExtraPillarForGan(t, hourGan),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			if hourGan == target {
				assertExactShenShaHitCount(t, got.Hour, "天刑煞："+target, 1)
				for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Global} {
					assertExactShenShaHitCount(t, bucket, "天刑煞："+target, 0)
				}
			} else {
				assertShenShaNameAbsentEverywhere(t, got, "天刑煞")
			}
		}
	}
}

func TestTianXingShaMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	meta := LookupShenShaMeta("天刑煞")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("天刑煞 metadata = %+v", meta)
	}
	for _, fragment := range []string{
		"只以生年支查时柱天干", "子丑乙", "申丙", "只在时干匹配时落入时柱",
		"《三命通会》PDF第122页、书内第119页", "不生成刑狱、疾病或现实事件结论",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("天刑煞 basis = %q, want %q", meta.Basis, fragment)
		}
	}
}
