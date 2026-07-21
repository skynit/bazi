package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var luShenExpectedProfile = map[string]string{
	"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午", "戊": "巳",
	"己": "午", "庚": "申", "辛": "酉", "壬": "亥", "癸": "子",
}

func TestLuShenProfileIsSingleSourceAcrossConsumers(t *testing.T) {
	dayStemOrder, luBranches := canonicalLuProfile()
	config := defaultBodyStrengthBonusRuleConfig()
	if config.DayStemOrder != dayStemOrder || config.LuBranches != luBranches {
		t.Fatalf("body-strength lu profile diverged: stems=%v branches=%v", config.DayStemOrder, config.LuBranches)
	}
	for i, stem := range dayStemOrder {
		want := luShenExpectedProfile[stem]
		got, ok := luBranchForStem(stem)
		if !ok || luBranches[i] != want || got != want {
			t.Errorf("stem %s canonical lu = %s/%s, want %s", stem, luBranches[i], got, want)
		}
		if got := ruleTargetsByName(dayGanShenShaRules[stem], "禄神"); len(got) != 0 {
			t.Errorf("dayGanShenShaRules still duplicates canonical lu for %s: %v", stem, got)
		}
	}
}

func TestCanonicalLuProfileReturnsIndependentValues(t *testing.T) {
	stems, branches := canonicalLuProfile()
	stems[0] = "变"
	branches[0] = "更"

	freshStems, freshBranches := canonicalLuProfile()
	if freshStems[0] != "甲" || freshBranches[0] != "寅" {
		t.Fatalf("fresh lu profile inherited mutation: %v/%v", freshStems, freshBranches)
	}
	if got, ok := luBranchForStem("甲"); !ok || got != "寅" {
		t.Fatalf("lu lookup inherited profile mutation: %q/%v", got, ok)
	}
	config := defaultBodyStrengthBonusRuleConfig()
	if config.DayStemOrder != freshStems || config.LuBranches != freshBranches {
		t.Fatalf("body-strength config inherited lu profile mutation: %+v", config)
	}
}

func TestLuShenFormalEntryTruthTableForExternalPillars(t *testing.T) {
	for _, dayGan := range data.Gans {
		for _, candidate := range data.Zhis {
			for _, targetIndex := range []int{0, 1, 3} {
				got := calcLuShenExternalFixture(t, dayGan, candidate, targetIndex)
				want := candidate == luShenExpectedProfile[dayGan]
				assertLuShenBucketTruth(t, dayGan, candidate, targetIndex, got, want)
			}
		}
	}
}

func TestLuShenDayPillarTruthAcrossSixtyCycle(t *testing.T) {
	matched := 0
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := model.Pillar{Gan: data.Gans[dayIndex%10], Zhi: data.Zhis[dayIndex%12]}
		target := luShenExpectedProfile[day.Gan]
		forbidden := map[string]bool{target: true}
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year: luShenNeutralPillar(t, forbidden, 0), Month: luShenNeutralPillar(t, forbidden, 13),
			Day: day, Hour: luShenNeutralPillar(t, forbidden, 27), Gender: model.GenderMale,
		})
		if err != nil {
			t.Fatal(err)
		}
		want := day.Zhi == target
		assertLuShenBucketTruth(t, day.Gan, day.Zhi, 2, got, want)
		if want {
			matched++
		}
	}
	if matched != 4 {
		t.Fatalf("day-pillar lu matches = %d, want 4", matched)
	}
}

func TestLuShenRepeatedTargetIsPreservedOnEveryValidPillar(t *testing.T) {
	for _, dayGan := range data.Gans {
		target := luShenExpectedProfile[dayGan]
		selfLu := sixtyCycleIndex(dayGan, target) >= 0
		day := luShenDayPillar(t, dayGan, map[string]bool{target: true})
		if selfLu {
			day = model.Pillar{Gan: dayGan, Zhi: target}
		}
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year: luShenPillarForBranch(t, target), Month: luShenPillarForBranch(t, target),
			Day: day, Hour: luShenPillarForBranch(t, target), Gender: model.GenderMale,
		})
		if err != nil {
			t.Fatal(err)
		}
		for index, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
			want := index != 2 || selfLu
			if actual := hasShenShaName(bucket, "禄神"); actual != want {
				t.Errorf("day stem %s repeated target %s bucket %d lu=%v, want %v: %+v", dayGan, target, index, actual, want, got)
			}
		}
		if hasShenShaName(got.Global, "禄神") {
			t.Errorf("day stem %s leaked lu into global bucket: %+v", dayGan, got)
		}
	}
}

func TestLuShenMetadataRecordsCanonicalConsumers(t *testing.T) {
	meta := LookupShenShaMeta("禄神")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("禄神 metadata = %+v", meta)
	}
	for _, fragment := range []string{
		"只以日干", "甲寅、乙卯、丙戊巳、丁己午、庚申、辛酉、壬亥、癸子",
		"《渊海子平》PDF第81页", "《三命通会》PDF第85-86页",
		"共同消费的唯一Profile", "不生成爵禄、财富或现实事件结论",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("禄神 basis = %q, want %q", meta.Basis, fragment)
		}
	}
}

func calcLuShenExternalFixture(t testing.TB, dayGan, candidate string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	target := luShenExpectedProfile[dayGan]
	forbidden := map[string]bool{candidate: true, target: true}
	pillars := []model.Pillar{
		luShenNeutralPillar(t, forbidden, 0),
		luShenNeutralPillar(t, forbidden, 13),
		luShenDayPillar(t, dayGan, forbidden),
		luShenNeutralPillar(t, forbidden, 27),
	}
	pillars[targetIndex] = luShenPillarForBranch(t, candidate)
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func assertLuShenBucketTruth(t testing.TB, dayGan, candidate string, targetIndex int, got ShenShaCalcResult, want bool) {
	t.Helper()
	for index, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
		actual := hasShenShaName(bucket, "禄神")
		wantInBucket := want && index == targetIndex
		if actual != wantInBucket {
			t.Errorf("day stem %s candidate %s target %d bucket %d lu=%v, want %v: %+v", dayGan, candidate, targetIndex, index, actual, wantInBucket, got)
		}
	}
}

func luShenNeutralPillar(t testing.TB, forbidden map[string]bool, offset int) model.Pillar {
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

func luShenDayPillar(t testing.TB, gan string, forbidden map[string]bool) model.Pillar {
	t.Helper()
	for _, zhi := range data.Zhis {
		if !forbidden[zhi] && sixtyCycleIndex(gan, zhi) >= 0 {
			return model.Pillar{Gan: gan, Zhi: zhi}
		}
	}
	t.Fatalf("no neutral day pillar for stem %s", gan)
	return model.Pillar{}
}

func luShenPillarForBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for _, gan := range data.Gans {
		if sixtyCycleIndex(gan, zhi) >= 0 {
			return model.Pillar{Gan: gan, Zhi: zhi}
		}
	}
	t.Fatalf("no valid pillar for branch %s", zhi)
	return model.Pillar{}
}
