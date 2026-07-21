package bazi

import (
	"reflect"
	"testing"

	"bazi/internal/service/data"
)

func TestRetiredWenChangTableIsAbsentFromDayStemRules(t *testing.T) {
	for dayGan, rules := range dayGanShenShaRules {
		for _, rule := range rules {
			if rule.Name == "文昌贵人" {
				t.Errorf("day stem %s still publishes unadjudicated 文昌贵人 rule: %+v", dayGan, rule)
			}
		}
	}
}

func TestRetiredWenChangTableNeverReachesFormalResultBuckets(t *testing.T) {
	retiredTargets := map[string]string{
		"甲": "巳", "乙": "午", "丙": "申", "丁": "酉", "戊": "申",
		"己": "酉", "庚": "亥", "辛": "子", "壬": "寅", "癸": "卯",
	}
	for _, dayGan := range data.Gans {
		for _, branch := range data.Zhis {
			got := calcHongYanFixture(t, dayGan, retiredTargets[dayGan], branch, 0)
			assertShenShaNameAbsentEverywhere(t, got, "文昌贵人")
		}
	}
}

func TestLocatedSanMingWenChangVariantsDifferFromRetiredTable(t *testing.T) {
	retired := map[string]string{
		"甲": "巳", "乙": "午", "丙": "申", "丁": "酉", "戊": "申",
		"己": "酉", "庚": "亥", "辛": "子", "壬": "寅", "癸": "卯",
	}
	wenChangGui := map[string]string{
		"甲": "巳", "乙": "亥", "丙": "戌", "丁": "辰", "戊": "申",
		"己": "午", "庚": "寅", "辛": "未", "壬": "卯", "癸": "丑",
	}
	wenXingGui := map[string]string{
		"甲": "午", "乙": "巳", "丙": "申", "丁": "酉", "戊": "申",
		"己": "酉", "庚": "戌", "辛": "亥", "壬": "寅", "癸": "卯",
	}
	if reflect.DeepEqual(retired, wenChangGui) || reflect.DeepEqual(retired, wenXingGui) || reflect.DeepEqual(wenChangGui, wenXingGui) {
		t.Fatal("retired 文昌贵人, 三命通会文昌贵, and 三命通会文星贵 tables must remain distinct")
	}
}

func TestRetiredWenChangMetadataIsUnavailable(t *testing.T) {
	meta := LookupShenShaMeta("文昌贵人")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("retired 文昌贵人 metadata = %+v", meta)
	}
}
