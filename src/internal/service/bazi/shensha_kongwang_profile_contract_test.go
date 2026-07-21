package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var kongWangDoubleBranchProfile = [6][]string{
	{"戌", "亥"}, {"申", "酉"}, {"午", "未"},
	{"辰", "巳"}, {"寅", "卯"}, {"子", "丑"},
}

func TestKongWangExactDoubleBranchProfileAcrossSixtyCycle(t *testing.T) {
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		gan := data.Gans[dayIndex%10]
		zhi := data.Zhis[dayIndex%12]
		want := kongWangDoubleBranchProfile[dayIndex/10]
		if got := getKongWangZhi(gan, zhi); !reflect.DeepEqual(got, want) {
			t.Errorf("day %s%s kong-wang = %v, want %v", gan, zhi, got, want)
		}
	}
	if got := getKongWangZhi("甲", "丑"); got != nil {
		t.Fatalf("invalid day pillar returned kong-wang %v", got)
	}
}

func TestKongWangFormalEntryTruthTableForYearMonthAndHour(t *testing.T) {
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		for _, candidate := range data.Zhis {
			for _, targetIndex := range []int{0, 1, 3} {
				got := calcKongWangProfileFixture(t, dayIndex, candidate, targetIndex)
				want := branchInList(candidate, kongWangDoubleBranchProfile[dayIndex/10])
				assertKongWangProfileTruth(t, dayIndex, candidate, targetIndex, got, want)
			}
		}
	}
}

func TestKongWangMetadataRecordsDoubleAndSingleBranchProfiles(t *testing.T) {
	meta := LookupShenShaMeta("空亡")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("空亡 metadata = %+v", meta)
	}
	for _, fragment := range []string{
		"只以日柱完整六十甲子所属旬", "甲子旬戌亥", "甲寅旬子丑",
		"《渊海子平》PDF第105页", "《三命通会》PDF第108-110页",
		"阳日干只取阳空", "阴日干只取阴亡", "当前不缩减为单支",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("空亡 basis = %q, want %q", meta.Basis, fragment)
		}
	}
}

func calcKongWangProfileFixture(t testing.TB, dayIndex int, candidate string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	emptyBranches := kongWangDoubleBranchProfile[dayIndex/10]
	forbidden := map[string]bool{candidate: true}
	for _, branch := range emptyBranches {
		forbidden[branch] = true
	}
	pillars := []model.Pillar{
		kongWangNeutralPillar(t, forbidden, 0),
		kongWangNeutralPillar(t, forbidden, 13),
		{Gan: data.Gans[dayIndex%10], Zhi: data.Zhis[dayIndex%12]},
		kongWangNeutralPillar(t, forbidden, 27),
	}
	pillars[targetIndex] = kongWangPillarForBranch(t, candidate)
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func assertKongWangProfileTruth(t testing.TB, dayIndex int, candidate string, targetIndex int, got ShenShaCalcResult, want bool) {
	t.Helper()
	buckets := [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global}
	for index, bucket := range buckets {
		actual := hasShenShaName(bucket, "空亡")
		wantInBucket := want && index == targetIndex
		if actual != wantInBucket {
			t.Errorf("day index %d candidate %s target %d bucket %d 空亡=%v, want %v: %+v", dayIndex, candidate, targetIndex, index, actual, wantInBucket, got)
		}
	}
}

func kongWangNeutralPillar(t testing.TB, forbidden map[string]bool, offset int) model.Pillar {
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

func kongWangPillarForBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for _, gan := range data.Gans {
		if sixtyCycleIndex(gan, zhi) >= 0 {
			return model.Pillar{Gan: gan, Zhi: zhi}
		}
	}
	t.Fatalf("no valid pillar available for branch %s", zhi)
	return model.Pillar{}
}
