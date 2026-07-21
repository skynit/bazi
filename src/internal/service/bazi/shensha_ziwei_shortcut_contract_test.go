package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestUnsupportedBaziZiWeiShortcutIsAbsentFromProductionSource(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"addZiWeiXing", "紫微星"} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production shen-sha source still contains unsupported shortcut %q", forbidden)
		}
	}
}

func TestYearDerivedJiangXingIsNotDuplicatedAsBaziZiWei(t *testing.T) {
	for _, yearZhi := range data.Zhis {
		t.Run(yearZhi, func(t *testing.T) {
			target := sanHeShenShaRules[yearZhi].Jiang
			if target == "" {
				t.Fatalf("year branch %s has no 将星 target", yearZhi)
			}
			pillars := ShenShaPillars{
				Year:   firstSixtyCyclePillarWithBranch(t, yearZhi),
				Month:  model.Pillar{Gan: "丙", Zhi: "寅"},
				Day:    firstSixtyCyclePillarWithBranch(t, target),
				Hour:   model.Pillar{Gan: "甲", Zhi: "子"},
				Gender: "MALE",
			}
			got, err := CalcShenShaByPillars(pillars)
			if err != nil {
				t.Fatal(err)
			}
			assertExactShenShaInBucket(t, "day", got.Day, "将星："+target)
			for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
				if hasShenShaName(bucket, "紫微星") {
					t.Fatalf("year-derived 将星 was duplicated as 八字紫微星: %+v", got)
				}
			}
		})
	}
}

func TestRemovedBaziZiWeiShortcutRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("紫微星")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("removed 八字紫微星 metadata = %+v", meta)
	}
}

func firstSixtyCyclePillarWithBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		if data.Zhis[i%12] == zhi {
			return model.Pillar{Gan: data.Gans[i%10], Zhi: zhi}
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %q", zhi)
	return model.Pillar{}
}
