package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestHongYanExactDayStemTableAcrossAllBranches(t *testing.T) {
	wants := map[string]string{
		"甲": "午", "乙": "午", "丙": "寅", "丁": "未", "戊": "子",
		"己": "辰", "庚": "戌", "辛": "酉", "壬": "巳", "癸": "申",
	}
	for _, dayGan := range data.Gans {
		gotTargets := make([]string, 0, 1)
		for _, rule := range dayGanShenShaRules[dayGan] {
			if rule.Name == "红艳煞" {
				gotTargets = append(gotTargets, rule.Target)
			}
		}
		if len(gotTargets) != 1 || gotTargets[0] != wants[dayGan] {
			t.Errorf("day stem %s 红艳煞 targets = %v, want [%s]", dayGan, gotTargets, wants[dayGan])
		}
		for _, branch := range data.Zhis {
			want := branch == wants[dayGan]
			if got := targetContainsZhi(gotTargets[0], branch); got != want {
				t.Errorf("day stem %s branch %s 红艳煞 = %v, want %v", dayGan, branch, got, want)
			}
		}
	}
}

func TestHongYanFormalEntryAssignsTargetToActualPillar(t *testing.T) {
	wants := map[string]string{
		"甲": "午", "乙": "午", "丙": "寅", "丁": "未", "戊": "子",
		"己": "辰", "庚": "戌", "辛": "酉", "壬": "巳", "癸": "申",
	}
	for _, dayGan := range data.Gans {
		target := wants[dayGan]
		for targetIndex := 0; targetIndex < 4; targetIndex++ {
			if targetIndex == 2 && sixtyCycleIndex(dayGan, target) < 0 {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "红艳煞："+target)
			if hasShenShaName(got.Global, "红艳煞") {
				t.Errorf("day stem %s target %s leaked 红艳煞 to global: %+v", dayGan, target, got)
			}
		}
	}
}

func TestHongYanFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	wants := map[string]string{
		"甲": "午", "乙": "午", "丙": "寅", "丁": "未", "戊": "子",
		"己": "辰", "庚": "戌", "辛": "酉", "壬": "巳", "癸": "申",
	}
	for _, dayGan := range data.Gans {
		target := wants[dayGan]
		for _, branch := range data.Zhis {
			if branch == target {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, branch, 0)
			assertShenShaNameAbsentEverywhere(t, got, "红艳煞")
		}
	}
}

func TestHongYanFormerWrongTargetsAreNegative(t *testing.T) {
	for _, tc := range []struct {
		dayGan string
		wrong  string
		want   string
	}{
		{dayGan: "乙", wrong: "申", want: "午"},
		{dayGan: "戊", wrong: "辰", want: "子"},
		{dayGan: "壬", wrong: "子", want: "巳"},
	} {
		wrong := calcHongYanFixture(t, tc.dayGan, tc.want, tc.wrong, 0)
		assertShenShaNameAbsentEverywhere(t, wrong, "红艳煞")
		positive := calcHongYanFixture(t, tc.dayGan, tc.want, tc.want, 0)
		assertOnlyPillarBucketHas(t, positive, 0, "红艳煞："+tc.want)
	}
}

func TestHongYanMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	meta := LookupShenShaMeta("红艳煞")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("红艳煞 metadata status = %+v", meta)
	}
	for _, citation := range []string{
		"甲乙午、丙寅、丁未、戊子、己辰、庚戌、辛酉、壬巳、癸申",
		"逐柱落位", "《三命通会》", "PDF第125页", "书内第122页",
	} {
		if !strings.Contains(meta.Basis, citation) {
			t.Errorf("红艳煞 basis = %q, want %q", meta.Basis, citation)
		}
	}
}

func calcHongYanFixture(t testing.TB, dayGan, target, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	pillars := []model.Pillar{
		hongYanNeutralPillar(t, target, 0),
		hongYanNeutralPillar(t, target, 10),
		hongYanDayPillar(t, dayGan, target),
		hongYanNeutralPillar(t, target, 20),
	}
	if targetIndex == 2 {
		pillars[2] = model.Pillar{Gan: dayGan, Zhi: placedBranch}
		if sixtyCycleIndex(pillars[2].Gan, pillars[2].Zhi) < 0 {
			t.Fatalf("invalid day pillar fixture %s%s", dayGan, placedBranch)
		}
	} else {
		pillars[targetIndex] = hongYanPillarForBranch(t, placedBranch)
	}
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func hongYanDayPillar(t testing.TB, dayGan, avoidBranch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == dayGan && pillar.Zhi != avoidBranch {
			return pillar
		}
	}
	t.Fatalf("no day pillar for stem %s avoiding branch %s", dayGan, avoidBranch)
	return model.Pillar{}
}

func hongYanNeutralPillar(t testing.TB, avoidBranch string, start int) model.Pillar {
	t.Helper()
	for offset := 0; offset < 60; offset++ {
		i := (start + offset) % 60
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi != avoidBranch {
			return pillar
		}
	}
	t.Fatalf("no neutral pillar avoiding branch %s", avoidBranch)
	return model.Pillar{}
}

func hongYanPillarForBranch(t testing.TB, branch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi == branch {
			return pillar
		}
	}
	t.Fatalf("no pillar for branch %s", branch)
	return model.Pillar{}
}
