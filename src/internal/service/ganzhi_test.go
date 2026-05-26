package service

import (
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestGanHeHua(t *testing.T) {
	tests := []struct{ gan1, gan2, want string }{
		{"甲", "己", "土"},
		{"己", "甲", "土"},
		{"乙", "庚", "金"},
		{"丙", "辛", "水"},
		{"丁", "壬", "木"},
		{"戊", "癸", "火"},
	}
	for _, tc := range tests {
		got := GanHeHua[tc.gan1+tc.gan2]
		if got != tc.want {
			t.Errorf("GanHeHua[%s%s] = %q, want %q", tc.gan1, tc.gan2, got, tc.want)
		}
	}
}

func TestGanKe(t *testing.T) {
	tests := []struct{ gan, want string }{
		{"甲", "戊"}, {"乙", "己"}, {"丙", "庚"}, {"丁", "辛"},
		{"戊", "壬"}, {"己", "癸"}, {"庚", "甲"}, {"辛", "乙"},
		{"壬", "丙"}, {"癸", "丁"},
	}
	for _, tc := range tests {
		if got := GanKe[tc.gan]; got != tc.want {
			t.Errorf("GanKe[%s] = %q, want %q", tc.gan, got, tc.want)
		}
	}
}

func TestGanSheng(t *testing.T) {
	tests := []struct{ gan, want string }{
		{"甲", "丙"}, {"乙", "丁"}, {"丙", "戊"}, {"丁", "己"},
		{"戊", "庚"}, {"己", "辛"}, {"庚", "壬"}, {"辛", "癸"},
		{"壬", "甲"}, {"癸", "乙"},
	}
	for _, tc := range tests {
		if got := GanSheng[tc.gan]; got != tc.want {
			t.Errorf("GanSheng[%s] = %q, want %q", tc.gan, got, tc.want)
		}
	}
}

func TestZhiSanHeGroups(t *testing.T) {
	groups := map[string][]string{
		"水": {"申", "子", "辰"},
		"木": {"亥", "卯", "未"},
		"火": {"寅", "午", "戌"},
		"金": {"巳", "酉", "丑"},
	}
	for _, members := range groups {
		first := ZhiSanHe[members[0]]
		if len(first) != 3 {
			t.Fatalf("ZhiSanHe[%s] len = %d, want 3", members[0], len(first))
		}
		for _, m := range members {
			if !containsInSlice(ZhiSanHe[m], members[0]) {
				t.Errorf("ZhiSanHe[%s] should contain %s", m, members[0])
			}
		}
	}
}

func TestZhiSanHuiGroups(t *testing.T) {
	groups := map[string][]string{
		"木": {"寅", "卯", "辰"},
		"火": {"巳", "午", "未"},
		"金": {"申", "酉", "戌"},
		"水": {"亥", "子", "丑"},
	}
	for _, members := range groups {
		for _, m := range members {
			if !containsInSlice(ZhiSanHui[m], members[0]) {
				t.Errorf("ZhiSanHui[%s] should contain %s", m, members[0])
			}
		}
	}
}

func TestZhiXingPairs(t *testing.T) {
	// 无礼之刑: 子卯
	if ZhiXing["子"] != "卯" || ZhiXing["卯"] != "子" {
		t.Error("无礼之刑(子卯) mismatch")
	}
	// 恃势之刑: 寅巳申
	if ZhiXing["寅"] != "巳" || ZhiXing["巳"] != "申" || ZhiXing["申"] != "寅" {
		t.Error("恃势之刑(寅巳申) mismatch")
	}
	// 无恩之刑: 丑戌未
	if ZhiXing["丑"] != "戌" || ZhiXing["戌"] != "未" || ZhiXing["未"] != "丑" {
		t.Error("无恩之刑(丑戌未) mismatch")
	}
	// 自刑
	for _, z := range []string{"辰", "午", "酉", "亥"} {
		if ZhiXing[z] != z {
			t.Errorf("自刑: ZhiXing[%s] = %q, want %q", z, ZhiXing[z], z)
		}
	}
}

func TestCalcGanZhiAnalysisSample(t *testing.T) {
	year := model.Pillar{Gan: "癸", Zhi: "未"}
	month := model.Pillar{Gan: "丙", Zhi: "辰"}
	day := model.Pillar{Gan: "戊", Zhi: "午"}
	hour := model.Pillar{Gan: "己", Zhi: "未"}

	result := CalcGanZhiAnalysis(year, month, day, hour)

	// 天干：戊癸合(年日) → "癸合戊（无情之合）" + explanation
	foundHe := false
	for _, r := range result.GanRelations {
		if r.Type == "五合" && (containsStr(r.Detail, "癸合戊") || containsStr(r.Detail, "戊合癸")) {
			foundHe = true
			break
		}
	}
	if !foundHe {
		t.Errorf("should find 戊癸合 in gan relations: %+v", result.GanRelations)
	}

	// 地支：午未合
	foundLiuHe := false
	foundChong := false
	for _, r := range result.ZhiRelations {
		if r.Type == "六合" && (containsStr(r.Detail, "午未合") || containsStr(r.Detail, "未午合")) {
			foundLiuHe = true
		}
		if r.Type == "六冲" {
			foundChong = true
		}
	}
	if !foundLiuHe {
		t.Errorf("should find 午未合 in zhi relations: %+v", result.ZhiRelations)
	}
	if foundChong {
		t.Errorf("should NOT find 六冲 in this chart: %+v", result.ZhiRelations)
	}
}

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}

func containsInSlice(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}
