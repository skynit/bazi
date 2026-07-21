package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestShenShaEntrypointAcceptsEverySixtyCyclePillar(t *testing.T) {
	base := ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
		Day: model.Pillar{Gan: "戊", Zhi: "辰"}, Hour: model.Pillar{Gan: "庚", Zhi: "申"},
		Gender: model.GenderMale,
	}
	for i := 0; i < 60; i++ {
		input := base
		input.Year = model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if _, err := CalcShenShaByPillars(input); err != nil {
			t.Errorf("cycle %d %s%s rejected: %v", i, input.Year.Gan, input.Year.Zhi, err)
		}
	}
	base.Gender = model.GenderFemale
	if _, err := CalcShenShaByPillars(base); err != nil {
		t.Fatalf("canonical female input rejected: %v", err)
	}
}

func TestShenShaEntrypointRejectsInvalidPillarsAndGender(t *testing.T) {
	valid := ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
		Day: model.Pillar{Gan: "戊", Zhi: "辰"}, Hour: model.Pillar{Gan: "庚", Zhi: "申"},
		Gender: model.GenderMale,
	}
	tests := []struct {
		name      string
		mutate    func(*ShenShaPillars)
		wantError string
	}{
		{name: "unknown year stem", mutate: func(p *ShenShaPillars) { p.Year.Gan = "X" }, wantError: "year pillar"},
		{name: "invalid year cycle pair", mutate: func(p *ShenShaPillars) { p.Year = model.Pillar{Gan: "甲", Zhi: "丑"} }, wantError: "year pillar"},
		{name: "empty month", mutate: func(p *ShenShaPillars) { p.Month = model.Pillar{} }, wantError: "month pillar"},
		{name: "invalid day cycle pair", mutate: func(p *ShenShaPillars) { p.Day = model.Pillar{Gan: "乙", Zhi: "子"} }, wantError: "day pillar"},
		{name: "unknown hour branch", mutate: func(p *ShenShaPillars) { p.Hour.Zhi = "X" }, wantError: "hour pillar"},
		{name: "unknown gender", mutate: func(p *ShenShaPillars) { p.Gender = "UNKNOWN" }, wantError: "gender"},
		{name: "unnormalized gender", mutate: func(p *ShenShaPillars) { p.Gender = "male" }, wantError: "gender"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.mutate(&input)
			got, err := CalcShenShaByPillars(input)
			if err == nil || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("result=%+v error=%v, want error containing %q", got, err, tc.wantError)
			}
			if len(got.Year) != 0 || len(got.Month) != 0 || len(got.Day) != 0 || len(got.Hour) != 0 || len(got.Global) != 0 {
				t.Fatalf("invalid input returned partial shen-sha result: %+v", got)
			}
		})
	}
}
