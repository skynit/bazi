package ziwei

import (
	"strings"
	"testing"
)

func TestParseGanZhiNameCoversSexagenaryCycle(t *testing.T) {
	for index := 0; index < 60; index++ {
		name := StemNames[index%len(StemNames)] + BranchNames[index%len(BranchNames)]
		stem, branch, err := parseGanZhiName(name)
		if err != nil || stem != index%10 || branch != index%12 {
			t.Fatalf("parse %s = %d/%d err=%v", name, stem, branch, err)
		}
	}
	stem, branch, err := parseGanZhiName(" 甲子 ")
	if err != nil || stem != 0 || branch != 0 {
		t.Fatalf("trimmed ganzhi = %d/%d err=%v", stem, branch, err)
	}
}

func TestParseGanZhiNameRejectsMalformedOrImpossiblePairs(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "", want: "exactly two"},
		{name: "甲", want: "exactly two"},
		{name: "甲子附", want: "exactly two"},
		{name: "天子", want: "unknown heavenly stem"},
		{name: "甲天", want: "unknown earthly branch"},
		{name: "甲丑", want: "parity mismatch"},
		{name: "乙子", want: "parity mismatch"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseGanZhiName(tc.name); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("parse %q error = %v, want containing %q", tc.name, err, tc.want)
			}
			input := ZiWeiDerivationInput{PeriodGanZhi: tc.name}
			if _, _, ok := derivationStemBranch(input); ok {
				t.Fatalf("derivation parser accepted %q", tc.name)
			}
		})
	}
}

func TestBuildBirthDataHourPillarMatchesAllCivilHours(t *testing.T) {
	for hour := 0; hour < 24; hour++ {
		birth, err := buildBirthData(2003, 4, 15, hour, 0, "男")
		if err != nil {
			t.Fatalf("hour %d: buildBirthData: %v", hour, err)
		}
		wantBranch := ((hour + 1) / 2) % len(BranchNames)
		if birth.HourBranch != wantBranch || birth.HourStem < 0 || birth.HourStem >= len(StemNames) ||
			birth.HourStem%2 != birth.HourBranch%2 {
			t.Fatalf("hour %d pillar = %s%s indexes=%d/%d, want branch %d", hour,
				StemNames[birth.HourStem], BranchNames[birth.HourBranch], birth.HourStem, birth.HourBranch, wantBranch)
		}
	}
}
