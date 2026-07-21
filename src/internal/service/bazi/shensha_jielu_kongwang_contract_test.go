package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestJieLuKongWangUsesYuanHaiDayStemHourBranchProfile(t *testing.T) {
	want := map[string][]string{
		"甲": {"申", "酉"}, "己": {"申", "酉"},
		"乙": {"午", "未"}, "庚": {"午", "未"},
		"丙": {"辰", "巳"}, "辛": {"辰", "巳"},
		"丁": {"寅", "卯"}, "壬": {"寅", "卯"},
		"戊": {"子", "丑"}, "癸": {"子", "丑"},
	}
	if !reflect.DeepEqual(jieLuKongWangByDayGan, want) {
		t.Fatalf("截路空亡 table = %+v, want YuanHai profile %+v", jieLuKongWangByDayGan, want)
	}
}

func TestJieLuKongWangFormalEntryCoversTenByTwelveTruthTable(t *testing.T) {
	for _, dayGan := range data.Gans {
		for _, hourZhi := range data.Zhis {
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   model.Pillar{Gan: "甲", Zhi: "子"},
				Month:  model.Pillar{Gan: "乙", Zhi: "丑"},
				Day:    jianLuDayPillar(t, dayGan),
				Hour:   poZhaiPillarForBranch(t, hourZhi),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			wantHit := branchInList(hourZhi, jieLuKongWangByDayGan[dayGan])
			if gotHit := hasShenShaName(got.Hour, "截路空亡"); gotHit != wantHit {
				t.Errorf("day stem %s hour branch %s hit = %v, want %v: %+v", dayGan, hourZhi, gotHit, wantHit, got)
			}
			for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Global} {
				if hasShenShaName(bucket, "截路空亡") {
					t.Errorf("day stem %s hour branch %s leaked 截路空亡 outside hour: %+v", dayGan, hourZhi, got)
				}
			}
		}
	}
}

func TestJieLuKongWangMetadataExposesProfileConflict(t *testing.T) {
	meta := LookupShenShaMeta("截路空亡")
	for _, want := range []string{"《渊海子平》PDF第107页", "戊癸子丑", "《三命通会》PDF第113页", "戊癸作戌亥", "当前不混表"} {
		if !strings.Contains(meta.Basis, want) {
			t.Errorf("截路空亡 metadata basis %q missing %q", meta.Basis, want)
		}
	}
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("截路空亡 metadata = %+v, want observed/not_adjudicated", meta)
	}
}
