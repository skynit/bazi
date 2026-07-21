package ziwei

import (
	"reflect"
	"slices"
	"sort"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

// This chart is independently published by airicyu/fortel-ziweidoushu at
// 2620cc895395f9f6994abd4927e739d31015c67d (test/generateSample.test.ts).
// Fortel uses a different four-hua school for some stems and derives 命主 from
// the year branch instead of the life-palace branch, so this comparison covers
// only the shared natal structure, fourteen main stars, and common auxiliary
// stars.
func TestZiWeiChartMatchesIndependentFortelSample(t *testing.T) {
	lunarDay, err := (tyme.LunarDay{}).FromYmd(1952, 12, 15)
	if err != nil {
		t.Fatalf("convert independent lunar fixture: %v", err)
	}
	solarDay := lunarDay.GetSolarDay()
	chart, err := NewZiWeiService().CalculateChart(
		solarDay.GetYear(), solarDay.GetMonth(), solarDay.GetDay(), 3, 0, "女",
	)
	if err != nil {
		t.Fatalf("CalculateChart: %v", err)
	}

	if chart.EarthlyBranchOfSoulPalace != "亥" || chart.BodyPalace != "卯" ||
		chart.FiveBureau != "金四局" || chart.BodyMaster != "文昌" {
		t.Fatalf("independent chart structure mismatch: soul=%s body=%s bureau=%s bodyMaster=%s",
			chart.EarthlyBranchOfSoulPalace, chart.BodyPalace, chart.FiveBureau, chart.BodyMaster)
	}

	type expectedPalace struct {
		name  string
		stars []string
		aux   []string
	}
	want := map[string]expectedPalace{
		"子": {name: "父母", stars: []string{"武曲", "天府"}, aux: []string{"擎羊", "铃星"}},
		"丑": {name: "福德", stars: []string{"太阳", "太阴"}, aux: []string{"地劫"}},
		"寅": {name: "田宅", stars: []string{"贪狼"}, aux: []string{"天马"}},
		"卯": {name: "事业", stars: []string{"天机", "巨门"}, aux: []string{"天魁", "左辅"}},
		"辰": {name: "交友", stars: []string{"紫微", "天相"}, aux: []string{"火星"}},
		"巳": {name: "迁移", stars: []string{"天梁"}, aux: []string{"天钺"}},
		"午": {name: "疾厄", stars: []string{"七杀"}, aux: []string{"文曲"}},
		"未": {name: "财帛", stars: []string{}},
		"申": {name: "子女", stars: []string{"廉贞"}, aux: []string{"文昌"}},
		"酉": {name: "夫妻", stars: []string{}, aux: []string{"地空"}},
		"戌": {name: "兄弟", stars: []string{"破军"}, aux: []string{"陀罗"}},
		"亥": {name: "命宫", stars: []string{"天同"}, aux: []string{"右弼", "禄存"}},
	}

	for _, palace := range chart.Palaces {
		expected, ok := want[palace.Branch]
		if !ok {
			t.Fatalf("unexpected palace branch %q", palace.Branch)
		}
		if palace.Name != expected.name {
			t.Errorf("%s palace name = %q, want %q", palace.Branch, palace.Name, expected.name)
		}
		if got := palaceMainStars(palace); !reflect.DeepEqual(got, expected.stars) {
			t.Errorf("%s main stars = %v, want %v", palace.Branch, got, expected.stars)
		}
		gotAux := palaceAuxStars(palace)
		sort.Strings(gotAux)
		sort.Strings(expected.aux)
		if !slices.Equal(gotAux, expected.aux) {
			t.Errorf("%s auxiliary stars = %v, want %v", palace.Branch, gotAux, expected.aux)
		}
	}
}
