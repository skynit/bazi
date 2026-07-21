package bazi

import (
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestUnlocatedBloodShenShaAreAbsentFromStemRuleTables(t *testing.T) {
	for _, name := range []string{"流霞", "血刃", "血忌"} {
		for _, gan := range data.Gans {
			if got := ruleTargetsByName(yearGanShenShaRules[gan], name); len(got) != 0 {
				t.Errorf("year stem %s still publishes %s targets: %v", gan, name, got)
			}
			if got := ruleTargetsByName(dayGanShenShaRules[gan], name); len(got) != 0 {
				t.Errorf("day stem %s still publishes %s targets: %v", gan, name, got)
			}
		}
	}
}

func TestUnlocatedBloodShenShaFailClosedAcrossLegacySearchSpace(t *testing.T) {
	for _, name := range []string{"流霞", "血刃", "血忌"} {
		for _, yearGan := range data.Gans {
			year := bloodRuleYearPillar(t, yearGan)
			for _, branch := range data.Zhis {
				got, err := CalcShenShaByPillars(ShenShaPillars{
					Year:   year,
					Month:  bloodRulePillarForBranch(t, branch),
					Day:    model.Pillar{Gan: "戊", Zhi: "辰"},
					Hour:   model.Pillar{Gan: "庚", Zhi: "申"},
					Gender: model.GenderMale,
				})
				if err != nil {
					t.Fatalf("%s/%s/%s fixture rejected: %v", name, yearGan, branch, err)
				}
				assertShenShaNameAbsentEverywhere(t, got, name)
			}
		}
	}
}

func TestUnlocatedBloodShenShaMetadataFailsClosed(t *testing.T) {
	for _, name := range []string{"流霞", "血刃", "血忌"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func bloodRuleYearPillar(t testing.TB, gan string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == gan {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle year pillar for stem %q", gan)
	return model.Pillar{}
}

func bloodRulePillarForBranch(t testing.TB, branch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi == branch {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %q", branch)
	return model.Pillar{}
}
