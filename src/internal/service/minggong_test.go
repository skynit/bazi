package service

import (
	"strings"
	"testing"
)

func TestCalcMingGongBasic(t *testing.T) {
	tests := []struct {
		name       string
		yearGan    string
		monthZhi   string
		hourZhi    string
		wantGanZhi string
	}{
		{"癸未年辰月未时", "癸", "辰", "未", "丁巳"},
		{"甲年寅月子时", "甲", "寅", "子", "丙子"},
		{"甲年寅月午时", "甲", "寅", "午", "庚午"},
		{"丙年辰月辰时", "丙", "辰", "辰", "庚寅"},
		{"戊年午月子时", "戊", "午", "子", "庚申"},
		{"壬年丑月亥时", "壬", "丑", "亥", "壬子"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CalcMingGong(tt.yearGan, tt.monthZhi, tt.hourZhi)
			if err != nil {
				t.Fatalf("CalcMingGong(%s, %s, %s) error: %v", tt.yearGan, tt.monthZhi, tt.hourZhi, err)
			}
			if got != tt.wantGanZhi {
				t.Errorf("CalcMingGong(%s, %s, %s) = %q, want %q", tt.yearGan, tt.monthZhi, tt.hourZhi, got, tt.wantGanZhi)
			}
		})
	}
}

func TestCalcMingGongInvalidParams(t *testing.T) {
	tests := []struct {
		name     string
		yearGan  string
		monthZhi string
		hourZhi  string
		wantErr  string
	}{
		{"invalid year gan", "Z", "寅", "子", "无效年干"},
		{"invalid month zhi", "甲", "猫", "子", "无效月支"},
		{"invalid hour zhi", "甲", "寅", "狗", "无效时支"},
		{"empty year gan", "", "寅", "子", "无效年干"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalcMingGong(tt.yearGan, tt.monthZhi, tt.hourZhi)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestGetMingGongShenSha(t *testing.T) {
	// 命宫神煞按地支固定对应
	tests := []struct {
		zhi         string
		wantShenSha string
	}{
		{"子", "天贵"},
		{"丑", "天厄"},
		{"寅", "天权"},
		{"卯", "天赦"},
		{"辰", "天如"},
		{"巳", "天文"},
		{"午", "天福"},
		{"未", "天驿"},
		{"申", "天孤"},
		{"酉", "天秘"},
		{"戌", "天艺"},
		{"亥", "天寿"},
	}

	for _, tt := range tests {
		t.Run(tt.zhi, func(t *testing.T) {
			got := GetMingGongShenSha(tt.zhi)
			if got != tt.wantShenSha {
				t.Errorf("GetMingGongShenSha(%q) = %q, want %q", tt.zhi, got, tt.wantShenSha)
			}
		})
	}
}

func TestGetMingGongShenShaInvalid(t *testing.T) {
	if got := GetMingGongShenSha("xx"); got != "" {
		t.Errorf("GetMingGongShenSha invalid = %q, want empty", got)
	}
}

func TestBuildMingGongDetail(t *testing.T) {
	detail := BuildMingGongDetail("丙子")

	if detail.GanZhi != "丙子" {
		t.Errorf("GanZhi = %q, want 丙子", detail.GanZhi)
	}
	if detail.Gan != "丙" {
		t.Errorf("Gan = %q, want 丙", detail.Gan)
	}
	if detail.Zhi != "子" {
		t.Errorf("Zhi = %q, want 子", detail.Zhi)
	}
	// 子 → 天贵
	if detail.ShenSha != "天贵" {
		t.Errorf("ShenSha = %q, want 天贵", detail.ShenSha)
	}
	if detail.ShenShaDesc == "" {
		t.Error("ShenShaDesc should not be empty")
	}
	if detail.ZhiDetail == "" {
		t.Error("ZhiDetail should not be empty")
	}
}

func TestBuildMingGongDetailInvalid(t *testing.T) {
	detail := BuildMingGongDetail("")
	if detail.GanZhi != "" {
		t.Errorf("GanZhi = %q, want empty", detail.GanZhi)
	}
}

func TestCalcMingGongFromBaziService(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2003, 4, 15, 13, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result.MingGong.GanZhi == "" {
		t.Error("MingGong.GanZhi should not be empty")
	}
	if result.MingGong.Gan == "" {
		t.Error("MingGong.Gan should not be empty")
	}
	if result.MingGong.Zhi == "" {
		t.Error("MingGong.Zhi should not be empty")
	}
	if result.MingGong.ShenSha == "" {
		t.Error("MingGong.ShenSha should not be empty")
	}

	// 癸未年 丙辰月 戊午日 己未时 → 命宫丁巳
	expected := "丁巳"
	if result.MingGong.GanZhi != expected {
		t.Errorf("MingGong.GanZhi = %q, want %q", result.MingGong.GanZhi, expected)
	}
	// 巳 → 天文
	if result.MingGong.ShenSha != "天文" {
		t.Errorf("MingGong.ShenSha = %q, want 天文", result.MingGong.ShenSha)
	}

	t.Logf("MingGong: %+v", result.MingGong)
}

func TestMingGongShenShaByZhiComplete(t *testing.T) {
	// Verify all 12 branches have a shensha assigned
	for i, zhi := range Zhis {
		if MingGongShenShaByZhi[i] == "" {
			t.Errorf("MingGongShenShaByZhi[%d] (branch %s) is empty", i, zhi)
		}
	}
}

func TestMingGongShenShaDescComplete(t *testing.T) {
	allShenSha := []string{
		"天贵", "天厄", "天权", "天赦",
		"天如", "天文", "天福", "天驿",
		"天孤", "天秘", "天艺", "天寿",
	}
	for _, name := range allShenSha {
		if desc := MingGongShenShaDesc[name]; desc == "" {
			t.Errorf("MingGongShenShaDesc[%q] is empty", name)
		}
	}
}

func TestMingGongZhiDetailComplete(t *testing.T) {
	for _, zhi := range Zhis {
		if detail := MingGongZhiDetail[zhi]; detail == "" {
			t.Errorf("MingGongZhiDetail[%q] is empty", zhi)
		}
	}
}
