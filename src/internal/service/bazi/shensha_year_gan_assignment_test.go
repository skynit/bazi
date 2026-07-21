package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var retiredYearGanExtraNames = []string{"金锁煞", "岁驾", "科名", "文星", "魁星"}

func TestRetiredYearGanExtraTablesHaveNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"jinSuoShaTarget", "keMingByGan", "wenXingByGan", "kuiXingByGan",
		"\"金锁煞\"", "\"岁驾\"", "\"科名\"", "\"文星\"", "\"魁星\"",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired year-stem path %q", forbidden)
		}
	}
}

func TestRetiredYearGanExtrasRemainAbsentAcrossFormerSearchSpace(t *testing.T) {
	for _, yearGan := range data.Gans {
		yearPillar := yearGanExtraPillarForGan(t, yearGan)
		for _, placedBranch := range data.Zhis {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   yearPillar,
				Month:  poZhaiPillarForBranch(t, placedBranch),
				Day:    model.Pillar{Gan: "甲", Zhi: "子"},
				Hour:   model.Pillar{Gan: "丙", Zhi: "寅"},
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range retiredYearGanExtraNames {
				assertShenShaNameAbsentEverywhere(t, got, name)
			}
		}
	}
}

func TestRetiredYearGanExtrasRemainUnregistered(t *testing.T) {
	for _, name := range retiredYearGanExtraNames {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func yearGanExtraPillarForGan(t testing.TB, gan string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == gan {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle pillar for stem %s", gan)
	return model.Pillar{}
}

func assertExactShenShaInBucket(t testing.TB, bucket string, items []string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !containsExactShenSha(items, want) {
			t.Errorf("%s shen-sha = %v, want %s", bucket, items, want)
		}
	}
}
