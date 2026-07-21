package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestRetiredYueKongHasNoProductionPath(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"addYueKong", "\"月空\"", "getKongWangZhi(p.Month.Gan"} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired 月空 path %q", forbidden)
		}
	}
}

func TestRetiredYueKongRemainsAbsentAcrossFormerSearchSpace(t *testing.T) {
	for i := range 60 {
		month := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		for _, placedBranch := range data.Zhis {
			for _, targetIndex := range []int{0, 2, 3} {
				pillars := []model.Pillar{
					{Gan: "甲", Zhi: "子"},
					month,
					{Gan: "丙", Zhi: "寅"},
					{Gan: "戊", Zhi: "辰"},
				}
				pillars[targetIndex] = poZhaiPillarForBranch(t, placedBranch)
				got, err := CalcShenShaByPillars(ShenShaPillars{
					Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
				})
				if err != nil {
					t.Fatal(err)
				}
				assertShenShaNameAbsentEverywhere(t, got, "月空")
			}
		}
	}
}

func TestRetiredYueKongRemainsUnregistered(t *testing.T) {
	meta := LookupShenShaMeta("月空")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("月空 metadata = %+v, want unregistered/not_available", meta)
	}
}
