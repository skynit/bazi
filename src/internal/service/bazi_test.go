package service

import (
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestCalculateMale1990(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result.YearPillar.Gan == "" || result.YearPillar.Zhi == "" {
		t.Errorf("YearPillar empty: %+v", result.YearPillar)
	}
	if result.MonthPillar.Gan == "" || result.MonthPillar.Zhi == "" {
		t.Errorf("MonthPillar empty: %+v", result.MonthPillar)
	}
	if result.DayPillar.Gan == "" || result.DayPillar.Zhi == "" {
		t.Errorf("DayPillar empty: %+v", result.DayPillar)
	}
	if result.HourPillar.Gan == "" || result.HourPillar.Zhi == "" {
		t.Errorf("HourPillar empty: %+v", result.HourPillar)
	}

	if len(result.FiveElements) != 5 {
		t.Errorf("FiveElements has %d keys, want 5: %v", len(result.FiveElements), result.FiveElements)
	}
	for _, elem := range []string{"木", "火", "土", "金", "水"} {
		if _, ok := result.FiveElements[elem]; !ok {
			t.Errorf("FiveElements missing key %q", elem)
		}
	}

	t.Logf("Pillars: Y=%s%s M=%s%s D=%s%s H=%s%s",
		result.YearPillar.Gan, result.YearPillar.Zhi,
		result.MonthPillar.Gan, result.MonthPillar.Zhi,
		result.DayPillar.Gan, result.DayPillar.Zhi,
		result.HourPillar.Gan, result.HourPillar.Zhi)
	t.Logf("FiveElements: %v", result.FiveElements)
	t.Logf("TenGods: %v", result.TenGods)
}

func TestCalculateFemale2000(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2000, 6, 1, 12, 0, "FEMALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result.DaYunInfo.StartAge <= 0 {
		t.Errorf("DaYun StartAge is %d, expected > 0", result.DaYunInfo.StartAge)
	}
	if len(result.DaYunInfo.Pillars) != 8 {
		t.Errorf("DaYun has %d pillars, want 8", len(result.DaYunInfo.Pillars))
	}

	t.Logf("DaYun: direction=%s startAge=%d pillars=%v",
		result.DaYunInfo.Direction, result.DaYunInfo.StartAge, result.DaYunInfo.Pillars)
	t.Logf("NaYin: %v", result.NaYin)
	t.Logf("HiddenStems: %v", result.HiddenStems)
}

func TestTenGodProportionSimpleCounting20030415WeiHour(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2003, 4, 15, 13, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if got := result.YearPillar.Gan + result.YearPillar.Zhi; got != "癸未" {
		t.Fatalf("year pillar = %s, want 癸未", got)
	}
	if got := result.MonthPillar.Gan + result.MonthPillar.Zhi; got != "丙辰" {
		t.Fatalf("month pillar = %s, want 丙辰", got)
	}
	if got := result.DayPillar.Gan + result.DayPillar.Zhi; got != "戊午" {
		t.Fatalf("day pillar = %s, want 戊午", got)
	}
	if got := result.HourPillar.Gan + result.HourPillar.Zhi; got != "己未" {
		t.Fatalf("hour pillar = %s, want 己未", got)
	}

	counts := make(map[string]int)
	percents := make(map[string]float64)
	for _, r := range result.TenGodProportion {
		counts[r.Name] = r.Count
		percents[r.Name] = r.Percent
	}

	expectedCounts := map[string]int{
		"劫财": 4,
		"正印": 3,
		"正官": 3,
		"正财": 2,
		"比肩": 1,
		"偏印": 1,
		"七杀": 0,
		"偏财": 0,
		"食神": 0,
		"伤官": 0,
	}
	for name, want := range expectedCounts {
		if got := counts[name]; got != want {
			t.Errorf("%s count = %d, want %d", name, got, want)
		}
	}

	expectedPercents := map[string]float64{
		"劫财": 28.57,
		"正印": 21.43,
		"正官": 21.43,
		"正财": 14.29,
		"比肩": 7.14,
		"偏印": 7.14,
	}
	for name, want := range expectedPercents {
		if got := percents[name]; got != want {
			t.Errorf("%s percent = %.2f, want %.2f", name, got, want)
		}
	}
}

func TestCalcDayShenShaWuWu(t *testing.T) {
	got := calcDayShenSha("戊", "午")
	for _, want := range []string{
		"禄神：巳", "金舆：未", "红艳煞：辰", "羊刃：午", "天厨食禄：申",
		"将星：午", "华盖：戌", "驿马：申", "劫煞：亥", "亡神：巳",
		"桃花：卯", "九丑日", "孤鸾煞", "阴差阳错", "六秀日",
	} {
		if !containsShenSha(got, want) {
			t.Fatalf("calcDayShenSha(戊, 午) missing %q in %#v", want, got)
		}
	}
}

func TestCalcDayShenShaSpecialDayMergesDuplicateDefinitions(t *testing.T) {
	for _, tc := range []struct {
		gan  string
		zhi  string
		want []string
	}{
		{gan: "庚", zhi: "辰", want: []string{"魁罡", "日德"}},
		{gan: "壬", zhi: "辰", want: []string{"魁罡", "阴差阳错"}},
	} {
		got := calcDayShenSha(tc.gan, tc.zhi)
		for _, want := range tc.want {
			if !containsShenSha(got, want) {
				t.Fatalf("calcDayShenSha(%s, %s) missing %q in %#v", tc.gan, tc.zhi, want, got)
			}
		}
	}
}

func containsShenSha(values []string, wantPrefix string) bool {
	for _, value := range values {
		if value == wantPrefix || strings.HasPrefix(value, wantPrefix+"：") || strings.HasPrefix(value, wantPrefix+"｜") {
			return true
		}
	}
	return false
}

func TestShenShaByPillarOrderAndItems(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2003, 4, 15, 13, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}
	byPillar := result.ShenShaByPillar
	if len(byPillar) != 4 {
		t.Fatalf("shen_sha_by_pillar has %d entries, want 4", len(byPillar))
	}
	for i, want := range []struct {
		pillar   string
		label    string
		priority int
	}{
		{"day", "日柱", 1},
		{"year", "年柱", 2},
		{"month", "月柱", 3},
		{"hour", "时柱", 4},
	} {
		if byPillar[i].Pillar != want.pillar || byPillar[i].Label != want.label || byPillar[i].Priority != want.priority {
			t.Fatalf("shen_sha_by_pillar[%d]: pillar=%s label=%s priority=%d, want pillar=%s label=%s priority=%d",
				i, byPillar[i].Pillar, byPillar[i].Label, byPillar[i].Priority, want.pillar, want.label, want.priority)
		}
	}
	if result.ShenShaSummary == nil {
		t.Fatal("shen_sha_summary is nil")
	}
	if len(result.ShenShaSummary.Description) < 5 {
		t.Fatalf("shen_sha_summary.description has %d items, want >=5", len(result.ShenShaSummary.Description))
	}
}

func TestCalcShenShaSampleHasAllCategories(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2003, 4, 15, 13, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}
	groups := map[string][]string{}
	for _, group := range result.ShenShaByPillar {
		groups[group.Pillar] = group.Items
	}
	if len(groups["day"]) == 0 || len(groups["hour"]) == 0 {
		t.Fatalf("priority shensha groups are unexpectedly empty: %#v", result.ShenShaByPillar)
	}
	for _, want := range []string{"羊刃：午", "病符：午", "的煞：午", "九丑日", "孤鸾煞", "阴差阳错", "六秀日"} {
		if !containsShenSha(groups["day"], want) {
			t.Fatalf("day shensha missing %q in %#v", want, groups["day"])
		}
	}
	if !containsShenSha(groups["hour"], "时贵：未") {
		t.Fatalf("hour shensha missing 时贵：未 in %#v", groups["hour"])
	}
}

func TestShenShaGuiWeiBingChenWuWuJiWeiMatchesManualVerification(t *testing.T) {
	got := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "癸", Zhi: "未"},
		Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	assertShenShaNames(t, "year", got.Year, []string{"华盖", "六厄", "墓煞", "德秀贵人", "金舆", "天乙贵人"})
	assertShenShaNames(t, "month", got.Month, []string{"月刑", "红艳煞", "太极贵人", "寡宿"})
	assertShenShaNames(t, "day", got.Day, []string{"病符", "的煞", "德秀贵人", "飞廉", "天火", "羊刃", "将星", "九丑日", "孤鸾煞", "阴差阳错", "六秀日", "十灵日"})
	assertShenShaNames(t, "hour", got.Hour, []string{"华盖", "德秀贵人", "金舆", "天乙贵人", "时贵"})
	assertShenShaNames(t, "global", got.Global, []string{"雷霆煞"})
}

func assertShenShaNames(t *testing.T, label string, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s shensha = %#v, want names %#v", label, got, want)
	}
	for i, item := range got {
		if name := shenShaName(item); name != want[i] {
			t.Fatalf("%s shensha[%d] = %s from %q, want %s", label, i, name, item, want[i])
		}
	}
}

func TestTianDeUsesMonthBranchMapping(t *testing.T) {
	got := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "癸", Zhi: "未"},
		Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	for _, group := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
		if containsShenSha(group, "天德贵人") || containsShenSha(group, "月德贵人") {
			t.Fatalf("癸未 丙辰 戊午 己未 should not have 天德/月德: %#v", got)
		}
	}

	withRen := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "子"},
		Month: model.Pillar{Gan: "壬", Zhi: "辰"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	if !containsShenSha(withRen.Month, "天德贵人：壬") {
		t.Fatalf("辰月见壬 should have 天德贵人：壬 in month shensha: %#v", withRen.Month)
	}
}

func TestShenShaPriorityDedupesAcrossCategories(t *testing.T) {
	got := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "午"},
		Month: model.Pillar{Gan: "丙", Zhi: "午"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "午"},
	})
	// Verify each pillar internally deduplicates
	for _, group := range []struct {
		name  string
		items []string
	}{
		{"day", got.Day},
		{"year", got.Year},
		{"month", got.Month},
		{"hour", got.Hour},
		{"global", got.Global},
	} {
		seen := map[string]bool{}
		for _, item := range group.items {
			name := shenShaName(item)
			if seen[name] {
				t.Fatalf("shensha %s appears twice within %s", name, group.name)
			}
			seen[name] = true
		}
	}
	// Day should contain dayGan shensha
	if !containsShenSha(got.Day, "羊刃") {
		t.Fatalf("day should keep 羊刃 in %#v", got.Day)
	}
	// Cross-pillar same-name shensha is now allowed (e.g. 年贵 and 日贵 are independent)
}

func TestKongWangBySixtyCycle(t *testing.T) {
	for _, tc := range []struct {
		gan  string
		zhi  string
		want []string
	}{
		{"甲", "子", []string{"戌", "亥"}},
		{"甲", "戌", []string{"申", "酉"}},
		{"甲", "申", []string{"午", "未"}},
		{"甲", "午", []string{"辰", "巳"}},
		{"甲", "辰", []string{"寅", "卯"}},
		{"甲", "寅", []string{"子", "丑"}},
	} {
		got := getKongWangZhi(tc.gan, tc.zhi)
		if len(got) != 2 || got[0] != tc.want[0] || got[1] != tc.want[1] {
			t.Fatalf("getKongWangZhi(%s%s) = %#v, want %#v", tc.gan, tc.zhi, got, tc.want)
		}
	}
}

func TestTianChuUsesShiShenLu(t *testing.T) {
	want := map[string]string{"甲": "巳", "乙": "午", "丙": "巳", "丁": "午", "戊": "申", "己": "酉", "庚": "亥", "辛": "子", "壬": "寅", "癸": "卯"}
	for gan, zhi := range want {
		got := calcDayShenSha(gan, "子")
		if !containsShenSha(got, "天厨食禄："+zhi) {
			t.Fatalf("%s day missing 天厨食禄：%s in %#v", gan, zhi, got)
		}
	}
}

func TestGlobalShenShaSequences(t *testing.T) {
	for _, tc := range []struct {
		pillars shenShaPillars
		want    string
	}{
		{shenShaPillars{Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "戊", Zhi: "寅"}, Day: model.Pillar{Gan: "庚", Zhi: "辰"}, Hour: model.Pillar{Gan: "癸", Zhi: "巳"}}, "天三奇：甲戊庚"},
		{shenShaPillars{Year: model.Pillar{Gan: "癸", Zhi: "子"}, Month: model.Pillar{Gan: "乙", Zhi: "寅"}, Day: model.Pillar{Gan: "丙", Zhi: "辰"}, Hour: model.Pillar{Gan: "丁", Zhi: "巳"}}, "地三奇：乙丙丁"},
		{shenShaPillars{Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "壬", Zhi: "寅"}, Day: model.Pillar{Gan: "癸", Zhi: "辰"}, Hour: model.Pillar{Gan: "辛", Zhi: "巳"}}, "人三奇：壬癸辛"},
	} {
		got := calcShenShaByPillars(tc.pillars)
		if !containsShenSha(got.Global, tc.want) {
			t.Fatalf("global shensha missing %q in %#v", tc.want, got.Global)
		}
	}
}

func TestDaYunGenderDiff(t *testing.T) {
	svc := &BaziService{}
	resultM, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate MALE failed: %v", err)
	}
	resultF, err := svc.Calculate(1990, 1, 15, 8, 0, "FEMALE")
	if err != nil {
		t.Fatalf("Calculate FEMALE failed: %v", err)
	}

	if resultM.DaYunInfo.Direction == resultF.DaYunInfo.Direction {
		t.Errorf("Expected different DaYun direction for MALE vs FEMALE, both are %q",
			resultM.DaYunInfo.Direction)
	}

	t.Logf("MALE: direction=%s startAge=%d firstPillar=%s%s",
		resultM.DaYunInfo.Direction, resultM.DaYunInfo.StartAge,
		resultM.DaYunInfo.Pillars[0].Gan, resultM.DaYunInfo.Pillars[0].Zhi)
	t.Logf("FEMALE: direction=%s startAge=%d firstPillar=%s%s",
		resultF.DaYunInfo.Direction, resultF.DaYunInfo.StartAge,
		resultF.DaYunInfo.Pillars[0].Gan, resultF.DaYunInfo.Pillars[0].Zhi)
}

func TestGuChenGuSuBySanHui(t *testing.T) {
	tests := []struct {
		yearZhi    string
		guChenWant string
		guSuWant   string
	}{
		{"寅", "巳", "丑"}, {"卯", "巳", "丑"}, {"辰", "巳", "丑"},
		{"巳", "申", "辰"}, {"午", "申", "辰"}, {"未", "申", "辰"},
		{"申", "亥", "未"}, {"酉", "亥", "未"}, {"戌", "亥", "未"},
		{"亥", "寅", "戌"}, {"子", "寅", "戌"}, {"丑", "寅", "戌"},
	}
	for _, tc := range tests {
		if g := guChenBySanHui[tc.yearZhi]; g != tc.guChenWant {
			t.Errorf("guChenBySanHui[%s] = %s, want %s", tc.yearZhi, g, tc.guChenWant)
		}
		if g := guSuBySanHui[tc.yearZhi]; g != tc.guSuWant {
			t.Errorf("guSuBySanHui[%s] = %s, want %s", tc.yearZhi, g, tc.guSuWant)
		}
	}
}

func TestTenLingDayStandard(t *testing.T) {
	want := map[string]bool{
		"甲辰": true, "乙亥": true, "丙辰": true, "丁酉": true, "戊午": true,
		"庚戌": true, "庚寅": true, "辛亥": true, "壬寅": true, "癸未": true,
	}
	// Verify all standard 十灵日 are present
	for dayPillar := range want {
		found := false
		for _, name := range specialDayShenShaRules[dayPillar] {
			if name == "十灵日" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("十灵日 %s missing from specialDayShenShaRules", dayPillar)
		}
	}
	// Verify old wrong entries are removed
	badEntries := []string{"乙卯", "庚辰", "壬申", "丁巳", "己巳", "辛未", "丙寅", "己丑", "癸亥"}
	for _, dayPillar := range badEntries {
		for _, name := range specialDayShenShaRules[dayPillar] {
			if name == "十灵日" {
				t.Errorf("十灵日 %s should NOT be in specialDayShenShaRules", dayPillar)
			}
		}
	}
}

func TestSixXiuDayComplete(t *testing.T) {
	want := map[string]bool{
		"丙子": true, "丁丑": true, "戊子": true, "戊午": true, "己丑": true,
		"丙午": true, "丁未": true, "己未": true, "辛酉": true,
	}
	for dayPillar := range want {
		found := false
		for _, name := range specialDayShenShaRules[dayPillar] {
			if name == "六秀日" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("六秀日 %s missing from specialDayShenShaRules", dayPillar)
		}
	}
}

func TestShenShaPerPillarAttribution(t *testing.T) {
	got := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "癸", Zhi: "未"},
		Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	// Day-gan 戊 rules: 天乙贵人丑未 → should match Year(未) and Hour(未), NOT Day(午)
	for _, item := range got.Year {
		if name := shenShaName(item); name == "天乙贵人" {
			goto yearOk
		}
	}
	t.Fatalf("Year should have 天乙贵人 because Day Gan 戊 贵人=丑未 matches Year Zhi=未: %#v", got.Year)
yearOk:
	for _, item := range got.Hour {
		if name := shenShaName(item); name == "天乙贵人" {
			goto hourOk
		}
	}
	t.Fatalf("Hour should have 天乙贵人 because Day Gan 戊 贵人=丑未 matches Hour Zhi=未: %#v", got.Hour)
hourOk:
	// Day should NOT have 天乙贵人 (Day Zhi=午 doesn't match 丑未)
	for _, item := range got.Day {
		if name := shenShaName(item); name == "天乙贵人" {
			t.Fatalf("Day should NOT have 天乙贵人 (Day Zhi=午 doesn't match 丑未): %#v", got.Day)
		}
	}
	// 金舆 for 戊=未 → should be in Year and Hour, not Day
	if !containsShenSha(got.Year, "金舆") {
		t.Fatalf("Year should have 金舆：未: %#v", got.Year)
	}
	if !containsShenSha(got.Hour, "金舆") {
		t.Fatalf("Hour should have 金舆：未: %#v", got.Hour)
	}
	if containsShenSha(got.Day, "金舆") {
		t.Fatalf("Day should NOT have 金舆：未: %#v", got.Day)
	}
	// 时贵 should appear (dayGan 戊 天乙贵人 丑未, Hour Zhi=未 matches)
	if !containsShenSha(got.Hour, "时贵") {
		t.Fatalf("Hour should have 时贵：未: %#v", got.Hour)
	}
}

func TestYueLingMatrixCorrectValues(t *testing.T) {
	// Verify that 休(生我)=1 and 死(我克)=0 are in the right positions for each day element
	tests := []struct {
		dayElem  string
		expected [5]float64 // 木火土金水 order for month branch
	}{
		{"木", [5]float64{3, 2, 0, 0, 1}}, // 旺木 相火 死土 囚金 休水
		{"火", [5]float64{1, 3, 2, 0, 0}}, // 休木 旺火 相土 死金 囚水
		{"土", [5]float64{0, 1, 3, 2, 0}}, // 囚木 休火 旺土 相金 死水
		{"金", [5]float64{0, 0, 1, 3, 2}}, // 死木 囚火 休土 旺金 相水
		{"水", [5]float64{2, 0, 0, 1, 3}}, // 相木 死火 囚土 休金 旺水
	}
	for _, tc := range tests {
		for mi, monthElem := range []string{"木", "火", "土", "金", "水"} {
			got := getYueLingScore(tc.dayElem, monthElem)
			want := tc.expected[mi]
			if got != want {
				t.Errorf("%s日主在%s月: 得令分=%.0f, want %.0f", tc.dayElem, monthElem, got, want)
			}
		}
	}
}

func TestCalcBodyStrengthLikeDislikeDynamic(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE") // 庚午 戊寅 己亥 戊辰
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}
	bs := result.BodyStrength
	dayGan := result.DayPillar.Gan // 己 → 土日主
	_ = dayGan
	// 己土日主, 身旺or身弱 depends on the actual computation
	// Verify like/dislike are 5 elements total (no empty, no duplicates)
	allElems := map[string]bool{"木": false, "火": false, "土": false, "金": false, "水": false}
	for _, e := range bs.Like {
		allElems[e] = true
	}
	for _, e := range bs.Dislike {
		allElems[e] = true
	}
	for elem, found := range allElems {
		if !found {
			t.Errorf("element %s not in like or dislike: like=%v dislike=%v", elem, bs.Like, bs.Dislike)
		}
	}
	if len(bs.Like)+len(bs.Dislike) != 5 {
		t.Errorf("like+dislike should cover all 5 elements: like=%v dislike=%v", bs.Like, bs.Dislike)
	}
	t.Logf("己土日主(1990-01-15): verdict=%s like=%v dislike=%v total=%.2f ling=%.2f di=%.2f shi=%.2f sheng=%.2f",
		bs.Verdict, bs.Like, bs.Dislike, bs.TotalScore, bs.LingScore, bs.DiScore, bs.ShiScore, bs.ShengScore)
}

func TestCalcBodyStrengthVerdictThreshold(t *testing.T) {
	svc := &BaziService{}

	for _, tc := range []struct {
		name   string
		year   int
		month  int
		day    int
		hour   int
		gender string
	}{
		// 1990-01-15: 己巳年 丁丑月 庚辰日 庚辰时 (金日主, 丑月=土, 土生金=休=1)
		{"金日主_丑月_1990", 1990, 1, 15, 8, "MALE"},
		// 2000-06-01: 庚辰年 壬午月 庚寅日 壬午时 (金日主, 午月=火, 火克金=囚=0)
		{"金日主_午月_2000", 2000, 6, 1, 12, "FEMALE"},
		// 2003-04-15: 癸未年 丙辰月 戊午日 己未时 (土日主, 辰月=土, 同=旺=3)
		{"土日主_辰月_2003", 2003, 4, 15, 13, "MALE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := svc.Calculate(tc.year, tc.month, tc.day, tc.hour, 0, tc.gender)
			if err != nil {
				t.Fatalf("Calculate failed: %v", err)
			}
			bs := result.BodyStrength
			if bs.Verdict != "身旺" && bs.Verdict != "身弱" {
				t.Errorf("verdict should be 身旺 or 身弱, got %q", bs.Verdict)
			}
			if bs.TotalScore < 0 {
				t.Errorf("total score should be >= 0, got %.2f", bs.TotalScore)
			}
			t.Logf("%s: dayGan=%s verdict=%s total=%.2f ling=%.2f di=%.2f shi=%.2f sheng=%.2f like=%v dislike=%v",
				tc.name, result.DayPillar.Gan, bs.Verdict, bs.TotalScore,
				bs.LingScore, bs.DiScore, bs.ShiScore, bs.ShengScore,
				bs.Like, bs.Dislike)
		})
	}
}

func TestCalcBodyStrengthIndividualScores(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2003, 4, 15, 13, 0, "MALE") // 癸未 丙辰 戊午 己未
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}
	bs := result.BodyStrength

	// 戊土日主生于辰月(土月) → 同我=旺=3 → 得令分应≥2
	if bs.LingScore < 2 {
		t.Errorf("戊土日主在辰(土)月, 得令分应≥2, got %.2f", bs.LingScore)
	}
	// 得分应≥0 (四柱地支中有比劫/印星藏干: 午藏丁(火生土=印), 未藏己(土=比), 辰藏戊(土=比))
	if bs.DiScore <= 0 {
		t.Errorf("戊土在辰月应有点得地分, got %.2f", bs.DiScore)
	}
	// 得生分: 天干地支藏干中丙(火生土=印)多处 → 应>0
	if bs.ShengScore <= 0 {
		t.Errorf("戊土日主有丙火印星, 得生分应>0, got %.2f", bs.ShengScore)
	}
	t.Logf("戊土日主(2003-04-15): verdict=%s total=%.2f ling=%.2f di=%.2f shi=%.2f sheng=%.2f like=%v dislike=%v",
		bs.Verdict, bs.TotalScore, bs.LingScore, bs.DiScore, bs.ShiScore, bs.ShengScore, bs.Like, bs.Dislike)
}

func TestYueKongByMonthXunKong(t *testing.T) {
	got := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "癸", Zhi: "未"},
		Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	// 丙辰月 → sixtyCycleIndex → 丙辰 is i=52, 52/10=5 → 空亡 子丑
	// 子丑 not in any pillar branch → no 月空 written
	for _, pillar := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
		if containsShenSha(pillar, "月空") {
			t.Fatalf("月空 should not appear for 丙辰月 (空亡子丑 not matched): %#v", pillar)
		}
	}

	// Test with 甲子月 (空亡戌亥): if any pillar has 戌 or 亥, 月空 should appear
	got2 := calcShenShaByPillars(shenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "戌"},
		Month: model.Pillar{Gan: "甲", Zhi: "子"},
		Day:   model.Pillar{Gan: "戊", Zhi: "午"},
		Hour:  model.Pillar{Gan: "己", Zhi: "未"},
	})
	// 甲子月 → sixtyCycleIndex=0, 0/10=0 → 空亡戌亥, Year=戌 matches
	if !containsShenSha(got2.Year, "月空") {
		t.Fatalf("Year should have 月空: %#v", got2.Year)
	}
}
