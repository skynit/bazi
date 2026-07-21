package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestJianLuHasNoShenShaProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "建禄") {
		t.Fatal("production shen-sha source still contains duplicate 建禄 output")
	}
}

func TestJianLuRemainsAbsentAcrossDayStemAndMonthBranchSpace(t *testing.T) {
	for _, dayGan := range data.Gans {
		day := jianLuDayPillar(t, dayGan)
		luZhi, _ := luBranchForStem(dayGan)
		for _, monthZhi := range data.Zhis {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   model.Pillar{Gan: "甲", Zhi: "子"},
				Month:  poZhaiPillarForBranch(t, monthZhi),
				Day:    day,
				Hour:   model.Pillar{Gan: "戊", Zhi: "辰"},
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertShenShaNameAbsentEverywhere(t, got, "建禄")
			if monthZhi == luZhi && !hasShenShaName(got.Month, "禄神") {
				t.Errorf("day stem %s month branch %s lost canonical 禄神 output: %+v", dayGan, monthZhi, got)
			}
		}
	}
}

func TestJianLuShenShaNameRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("建禄")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("建禄 metadata = %+v, want unregistered/not_available", meta)
	}
}

func jianLuDayPillar(t testing.TB, gan string) model.Pillar {
	t.Helper()
	for i := range 60 {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == gan {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle pillar for day stem %s", gan)
	return model.Pillar{}
}
