package fortune

import (
	"testing"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestBuildFortuneLayersWithDefaultCalculators(t *testing.T) {
	svc := &bazipkg.BaziService{}
	chart, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	chart.DaYunInfo.Pillars = []model.Pillar{{Gan: "乙", Zhi: "卯"}}

	queryDate := time.Date(2026, 6, 16, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	layers := BuildFortuneLayers(chart, queryDate, 1990)
	if layers.RuleVersion != bazipkg.RuleVersion || layers.School != bazipkg.RuleSchool {
		t.Fatalf("unexpected layer rule meta: %+v", layers)
	}
	checks := map[string]string{
		"dayun":   layers.DaYun.Key,
		"liunian": layers.LiuNian.Key,
		"liuyue":  layers.LiuYue.Key,
		"xiaoyun": layers.XiaoYun.Key,
	}
	for want, got := range checks {
		if got != want {
			t.Fatalf("layer key = %q, want %q", got, want)
		}
	}
	if layers.LiuNian.Year != 2026 {
		t.Fatalf("liunian year = %d, want 2026", layers.LiuNian.Year)
	}
	if len(layers.LiuNian.ShenShaDetails) != len(layers.LiuNian.ActivatedShenSha) {
		t.Fatalf("liunian shensha details/name count mismatch: %+v", layers.LiuNian)
	}
}
