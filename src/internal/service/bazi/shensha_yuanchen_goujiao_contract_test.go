package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestYuanChenGouJiaoClosedFormulaForAllYearBranchesAndGroups(t *testing.T) {
	groups := []struct {
		name    string
		gan     string
		gender  string
		forward bool
	}{
		{name: "阳男", gan: "甲", gender: model.GenderMale, forward: true},
		{name: "阴女", gan: "乙", gender: model.GenderFemale, forward: true},
		{name: "阴男", gan: "乙", gender: model.GenderMale, forward: false},
		{name: "阳女", gan: "甲", gender: model.GenderFemale, forward: false},
	}

	for yearIndex, yearBranch := range data.Zhis {
		for _, group := range groups {
			yuan, gou, jiao := genderBasedShenShaTargets(group.gan, yearBranch, group.gender)
			chongIndex := (yearIndex + 6) % 12
			wantYuan, wantGou, wantJiao := "", "", ""
			if group.forward {
				wantYuan = data.Zhis[(chongIndex+1)%12]
				wantGou = data.Zhis[(yearIndex+3)%12]
				wantJiao = data.Zhis[(yearIndex-3+12)%12]
			} else {
				wantYuan = data.Zhis[(chongIndex-1+12)%12]
				wantGou = data.Zhis[(yearIndex-3+12)%12]
				wantJiao = data.Zhis[(yearIndex+3)%12]
			}
			if yuan != wantYuan || gou != wantGou || jiao != wantJiao {
				t.Errorf("%s %s targets=(%s,%s,%s), want (%s,%s,%s)", group.name, yearBranch, yuan, gou, jiao, wantYuan, wantGou, wantJiao)
			}
		}
	}
}

func TestYuanChenGouJiaoFormalEntryForAllYearBranchesAndGenders(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		yearPillar := poZhaiPillarForBranch(t, yearBranch)
		for _, gender := range []string{model.GenderMale, model.GenderFemale} {
			yuan, gou, jiao := genderBasedShenShaTargets(yearPillar.Gan, yearBranch, gender)
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   yearPillar,
				Month:  poZhaiPillarForBranch(t, yuan),
				Day:    poZhaiPillarForBranch(t, gou),
				Hour:   poZhaiPillarForBranch(t, jiao),
				Gender: gender,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertGenderShenShaOnlyInBucket(t, got, 1, "元辰："+yuan)
			assertGenderShenShaOnlyInBucket(t, got, 2, "勾煞："+gou)
			assertGenderShenShaOnlyInBucket(t, got, 3, "绞煞："+jiao)
			assertShenShaNameAbsentEverywhere(t, got, "勾绞煞")
			assertShenShaNameAbsentEverywhere(t, got, "暴败煞")
		}
	}
}

func TestYuanChenGouJiaoAttachToEveryRepeatedTargetPillar(t *testing.T) {
	pillars := ShenShaPillars{Year: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: model.GenderMale}
	for name, target := range map[string]string{"元辰": "未", "勾煞": "卯", "绞煞": "酉"} {
		var got ShenShaCalcResult
		addGenderBasedShenSha(pillars, []string{"子", target, target, target}, &got)
		assertExactShenShaHitCount(t, got.Year, name+"："+target, 0)
		for _, bucket := range [][]string{got.Month, got.Day, got.Hour} {
			assertExactShenShaHitCount(t, bucket, name+"："+target, 1)
		}
	}
}

func TestRetiredGouJiaoAndBaoBaiNamesHaveNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"\"勾绞煞\"", "\"暴败煞\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired path %s", forbidden)
		}
	}
	for _, name := range []string{"勾绞煞", "暴败煞"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func TestYuanChenGouJiaoMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	want := map[string][]string{
		"元辰": {"只以生年支六冲位", "阳男阴女", "甲子取未", "乙丑男取午", "《三命通会》PDF第114页", "不生成健康、灾祸或现实事件结论"},
		"勾煞": {"顺行三辰", "逆行三辰", "《三命通会》PDF第117页", "《渊海子平》PDF第635页", "当前Profile不混表"},
		"绞煞": {"逆行三辰", "顺行三辰", "《三命通会》PDF第117页", "《渊海子平》PDF第635页", "当前Profile不混表"},
	}
	for name, fragments := range want {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v", name, meta)
		}
		for _, fragment := range fragments {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}
}

func assertGenderShenShaOnlyInBucket(t testing.TB, got ShenShaCalcResult, wantIndex int, want string) {
	t.Helper()
	for index, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
		wantCount := 0
		if index == wantIndex {
			wantCount = 1
		}
		assertExactShenShaHitCount(t, bucket, want, wantCount)
	}
}
