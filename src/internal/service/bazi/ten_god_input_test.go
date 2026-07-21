package bazi

import "testing"

func TestClassifyTenGodRejectsInvalidStems(t *testing.T) {
	tests := []struct {
		name, stem, day string
	}{
		{"both empty", "", ""},
		{"both unknown", "Z", "W"},
		{"invalid observed stem", "X", "甲"},
		{"invalid day stem", "甲", "Y"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyTenGod(tc.stem, tc.day, false); got != "" {
				t.Fatalf("ClassifyTenGod(%q, %q) = %q, want empty", tc.stem, tc.day, got)
			}
		})
	}
}

func TestClassifyTenGodKeepsValidRelations(t *testing.T) {
	tests := []struct {
		stem, day, want string
	}{
		{"甲", "甲", "比肩"},
		{"乙", "甲", "劫财"},
		{"丙", "甲", "食神"},
		{"辛", "甲", "正官"},
		{"己", "甲", "正财"},
	}
	for _, tc := range tests {
		if got := ClassifyTenGod(tc.stem, tc.day, false); got != tc.want {
			t.Errorf("ClassifyTenGod(%s, %s) = %s, want %s", tc.stem, tc.day, got, tc.want)
		}
	}
}
