package bazi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestAnalyzeTiaohouAtUsesExactJieIntervalWithoutChangingCandidateOrder(t *testing.T) {
	jie, err := tyme.SolarTerm{}.FromName(2022, "惊蛰")
	if err != nil {
		t.Fatal(err)
	}
	birth := jie.GetJulianDay().GetSolarTime()
	got, err := AnalyzeTiaohouAt("甲", "卯", birth)
	if err != nil {
		t.Fatal(err)
	}
	if got.TablePrimaryCandidate != got.Rules[0].XiShen || got.TablePrimaryCandidate != "庚" ||
		got.SelectionBasis != "first_table_entry_candidate" || got.DepthAffectsSelection {
		t.Fatalf("unadjudicated depth must not change table candidate order: %+v", got)
	}
	evidence := got.DepthEvidence
	if evidence.Status != "observed" || evidence.Basis != "solar_term_jie_interval" || evidence.StartTerm != "惊蛰" || evidence.EndTerm != "清明" || evidence.ElapsedSeconds != 0 || evidence.IntervalSeconds <= 0 || evidence.Position != 0 || evidence.Phase != "前段" {
		t.Fatalf("depth evidence at Jie boundary = %+v", evidence)
	}
	if !ValidTiaohouEvidence(got, "甲", "卯") {
		t.Fatalf("tiaohou evidence contract is invalid: %+v", got)
	}
}

func TestAnalyzeTiaohouAtKeepsEarthMonthCommandProfilesSeparate(t *testing.T) {
	jie, err := tyme.SolarTerm{}.FromName(2024, "清明")
	if err != nil {
		t.Fatal(err)
	}
	birth := jie.GetJulianDay().GetSolarTime().Next(8 * secondsPerDay)
	got, err := AnalyzeTiaohouAt("甲", "辰", birth)
	if err != nil {
		t.Fatal(err)
	}

	candidates := got.DepthEvidence.MonthCommandCandidates
	if len(candidates) != 2 {
		t.Fatalf("earth-month command candidates = %+v, want two source profiles", candidates)
	}
	sanming, yuanhai := candidates[0], candidates[1]
	if sanming.Source != "《三命通会·论人元司事》" || sanming.CommandingStem != "壬" ||
		sanming.Segment != "墓库段" || sanming.SegmentStartDay != 8 || sanming.SegmentEndDay != 12 ||
		sanming.PositionDay != 9 {
		t.Fatalf("Sanming day-command candidate = %+v", sanming)
	}
	if yuanhai.Source != "《渊海子平·又论节气歌》" || yuanhai.CommandingStem != "乙" ||
		yuanhai.Segment != "余气段" || yuanhai.SegmentStartDay != 1 || yuanhai.SegmentEndDay != 9 ||
		yuanhai.PositionDay != 9 {
		t.Fatalf("Yuanhai day-command candidate = %+v", yuanhai)
	}
	if got.DepthAffectsSelection || !ValidTiaohouEvidence(got, "甲", "辰") {
		t.Fatalf("parallel day-command evidence changed selection or failed validation: %+v", got)
	}
}

func TestAnalyzeTiaohouWithoutBirthTimeMarksDepthUnavailable(t *testing.T) {
	got, err := AnalyzeTiaohou("甲", "寅")
	if err != nil {
		t.Fatal(err)
	}
	if got.DepthEvidence.Status != "unavailable" || got.DepthAffectsSelection {
		t.Fatalf("pillar-only depth must remain unavailable: %+v", got)
	}
	if !ValidTiaohouEvidence(got, "甲", "寅") {
		t.Fatalf("pillar-only tiaohou evidence contract is invalid: %+v", got)
	}
}

func TestTiaohouDoesNotPublishContradictoryLegacyJiShen(t *testing.T) {
	// 《穷通宝鉴》二月乙木明言“丙为君，癸为臣”。旧逐行表却在第一行
	// 把癸列为忌神、第二行又列为喜神；正式输出不能延续这种矛盾。
	got, err := AnalyzeTiaohou("乙", "卯")
	if err != nil {
		t.Fatal(err)
	}
	if got.RuleID != tiaohouRuleID || len(got.Rules) != 2 ||
		got.Rules[0].XiShen != "丙" || got.Rules[1].XiShen != "癸" {
		t.Fatalf("乙卯月 candidate sequence changed unexpectedly: %+v", got)
	}
	for _, rule := range got.Rules {
		if rule.JiShen != tiaohouJiShenUnadjudicated || rule.JiShenStatus != "not_adjudicated" {
			t.Fatalf("unreviewed legacy JiShen leaked into public evidence: %+v", rule)
		}
	}
	if !ValidTiaohouEvidence(got, "乙", "卯") {
		t.Fatalf("revised tiaohou evidence is invalid: %+v", got)
	}
}

func TestYiMaoTiaohouExplainsBingGuiPairing(t *testing.T) {
	// 《穷通宝鉴》PDF第53页：二月乙木“以丙为君，癸为臣”，并说明
	// “癸众用丙，丙多用癸，丙癸各得其用”，两者是配合关系。
	got, err := AnalyzeTiaohou("乙", "卯")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rules) != 2 {
		t.Fatalf("乙卯月 rules = %+v, want two candidates", got.Rules)
	}
	if !strings.Contains(got.Rules[0].SourceText, "丙为君") ||
		!strings.Contains(got.Rules[0].SourceText, "与癸水相济") ||
		!strings.Contains(got.Rules[1].SourceText, "癸为臣") {
		t.Fatalf("乙卯月未正确说明丙癸配合关系: %+v", got.Rules)
	}
	for _, rule := range got.Rules {
		if strings.Contains(rule.SourceText, "癸水不宜混") {
			t.Fatalf("乙卯月仍包含与原文矛盾的解释: %+v", rule)
		}
	}
}

func TestBingChenTiaohouMatchesClassicalPrimarySequence(t *testing.T) {
	// 《穷通宝鉴》PDF第99页：三月丙火“专用壬水”，土局“取甲为辅”，
	// 无甲时才退而用庚泄土生壬；原文另称丙火用壬为“水辅阳光”。
	got, err := AnalyzeTiaohou("丙", "辰")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"壬", "甲", "庚"}
	if len(got.Rules) != len(want) {
		t.Fatalf("丙辰月 candidates = %+v, want %v", got.Rules, want)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Errorf("丙辰月 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if got.TablePrimaryCandidate != "壬" ||
		!strings.Contains(got.Rules[0].SourceText, "水辅") ||
		strings.Contains(got.Rules[0].SourceText, "充日元之不足") {
		t.Fatalf("丙辰月主候选依据失真: %+v", got.Rules[0])
	}
}

func TestBingSummerTiaohouKeepsConditionalGuiInFifthMonth(t *testing.T) {
	// 《穷通宝鉴》PDF第112页：“丁多，兼看癸水”承接四五月；
	// 随后另明言“六月用壬，但借庚金为佐”。
	cases := []struct {
		month string
		want  []string
	}{
		{month: "午", want: []string{"壬", "庚", "癸"}},
		{month: "未", want: []string{"壬", "庚"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("丙", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("丙%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, candidate := range tc.want {
				if got.Rules[i].XiShen != candidate {
					t.Errorf("丙%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, candidate)
				}
			}
		})
	}
}

func TestJiaTiaohouCandidateOrderMatchesExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "卯", want: []string{"庚", "戊", "丁"}},
		{month: "午", want: []string{"癸", "丁", "庚"}},
		{month: "未", want: []string{"丁", "庚", "癸"}},
		{month: "申", want: []string{"丁", "庚"}},
		{month: "酉", want: []string{"丁", "丙", "庚"}},
		{month: "戌", want: []string{"丁", "壬", "癸"}},
		{month: "亥", want: []string{"庚", "丁", "丙", "戊"}},
		{month: "子", want: []string{"丁", "庚", "丙", "戊"}},
		{month: "丑", want: []string{"庚", "丁", "丙"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("甲", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("甲%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("甲%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("甲%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestYiTiaohouCandidatesCoverExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"丙", "癸"}},
		{month: "卯", want: []string{"丙", "癸"}},
		{month: "辰", want: []string{"癸", "丙"}},
		{month: "巳", want: []string{"癸", "丙", "辛", "庚"}},
		{month: "午", want: []string{"癸", "丙"}},
		{month: "未", want: []string{"癸", "丙", "庚", "辛"}},
		{month: "申", want: []string{"己", "丙", "癸"}},
		{month: "酉", want: []string{"癸", "丙", "丁"}},
		{month: "戌", want: []string{"癸", "辛"}},
		{month: "亥", want: []string{"丙", "戊", "甲", "庚"}},
		{month: "子", want: []string{"丙", "戊"}},
		{month: "丑", want: []string{"丙"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("乙", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("乙%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("乙%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("乙%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestYiWeiTiaohouKeepsMetalAsSupportForGui(t *testing.T) {
	// 《穷通宝鉴》PDF第66、67页：六月乙木“所重在癸”，总论又明言
	// “专用癸水，丙火酌用，庚辛次之”；庚辛只能作为癸水的发源之佐。
	got, err := AnalyzeTiaohou("乙", "未")
	if err != nil {
		t.Fatal(err)
	}
	if got.TablePrimaryCandidate != "癸" || len(got.Rules) != 4 ||
		got.Rules[2].XiShen != "庚" || !strings.Contains(got.Rules[2].SourceText, "生助癸水") ||
		got.Rules[3].XiShen != "辛" || !strings.Contains(got.Rules[3].SourceText, "癸水之佐") {
		t.Fatalf("六月乙木未按原文保留癸主、庚辛为佐的层级: %+v", got)
	}
}

func TestYiHaiTiaohouKeepsConditionalRemedyChain(t *testing.T) {
	// 《穷通宝鉴》PDF第81、82页：十月乙木以丙为主，随后按病药关系
	// 依次为“水旺用戊，戊多用甲，甲多用庚”，后续救应不能取代丙火。
	got, err := AnalyzeTiaohou("乙", "亥")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"丙", "戊", "甲", "庚"}
	if got.TablePrimaryCandidate != "丙" || len(got.Rules) != len(want) {
		t.Fatalf("十月乙木主候选或救应链不完整: %+v", got)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Fatalf("十月乙木 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(got.Rules[1].SourceText, "壬水过多") ||
		!strings.Contains(got.Rules[2].SourceText, "戊土过多") ||
		!strings.Contains(got.Rules[3].SourceText, "甲木过多") ||
		!strings.Contains(got.Rules[3].SourceText, "仍须兼用丙火") {
		t.Fatalf("十月乙木未保留原文病药条件: %+v", got.Rules)
	}
}

func TestChenWoodTiaohouExplainsHotLateSpring(t *testing.T) {
	jia, err := AnalyzeTiaohou("甲", "辰")
	if err != nil {
		t.Fatal(err)
	}
	if len(jia.Rules) != 2 ||
		!strings.Contains(jia.Rules[0].SourceText, "木气相竭") ||
		!strings.Contains(jia.Rules[1].SourceText, "阳盛木渴") ||
		!strings.Contains(jia.Rules[1].SourceText, "泄庚润木") {
		t.Fatalf("甲木辰月解释未体现《穷通宝鉴》的暮春裁润次序: %+v", jia.Rules)
	}
	for _, rule := range jia.Rules {
		if strings.Contains(rule.SourceText, "土湿") || strings.Contains(rule.SourceText, "湿土") {
			t.Fatalf("甲木辰月仍含与原文相反的湿土解释: %+v", jia.Rules)
		}
	}

	got, err := AnalyzeTiaohou("乙", "辰")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Rules) != 2 ||
		!strings.Contains(got.Rules[0].SourceText, "阳气愈炽") ||
		!strings.Contains(got.Rules[1].SourceText, "泄木之秀") {
		t.Fatalf("乙木辰月解释未体现《穷通宝鉴》的阳盛滋泄次序: %+v", got.Rules)
	}
	for _, rule := range got.Rules {
		if strings.Contains(rule.SourceText, "寒湿") || strings.Contains(rule.SourceText, "暖土") {
			t.Fatalf("乙木辰月仍含与原文相反的寒湿解释: %+v", got.Rules)
		}
	}
}

func TestBingTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"壬", "庚", "戊", "甲", "癸"}},
		{month: "卯", want: []string{"壬", "庚", "辛", "戊", "甲"}},
		{month: "辰", want: []string{"壬", "甲", "庚"}},
		{month: "巳", want: []string{"壬", "庚", "癸", "戊"}},
		{month: "午", want: []string{"壬", "庚", "癸"}},
		{month: "未", want: []string{"壬", "庚"}},
		{month: "申", want: []string{"壬", "戊", "甲"}},
		{month: "酉", want: []string{"壬", "癸"}},
		{month: "戌", want: []string{"甲", "壬", "癸"}},
		{month: "亥", want: []string{"甲", "戊", "庚", "己", "壬"}},
		{month: "子", want: []string{"壬", "戊", "甲", "癸"}},
		{month: "丑", want: []string{"壬", "甲"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("丙", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("丙%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("丙%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("丙%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestBingYinTiaohouKeepsJiaAndGuiConditionalFallbacks(t *testing.T) {
	// 《穷通宝鉴》PDF第90、91页：正月丙火以壬庚为正轨；壬多用戊，
	// 戊土成片时以甲制戊，支成火局而无壬时才姑用癸。
	got, err := AnalyzeTiaohou("丙", "寅")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"壬", "庚", "戊", "甲", "癸"}
	if got.TablePrimaryCandidate != "壬" || len(got.Rules) != len(want) {
		t.Fatalf("正月丙火主线或条件候选不完整: %+v", got)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Fatalf("正月丙火 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(got.Rules[3].SourceText, "戊土成片") ||
		!strings.Contains(got.Rules[3].SourceText, "甲木制戊") ||
		!strings.Contains(got.Rules[4].SourceText, "支成火局") ||
		!strings.Contains(got.Rules[4].SourceText, "无壬水") ||
		!strings.Contains(got.Rules[4].SourceText, "低于壬水") {
		t.Fatalf("正月丙火未保留甲癸候选的适用条件: %+v", got.Rules)
	}
}

func TestBingSiTiaohouUsesWuOnlyForExcessWater(t *testing.T) {
	// 《穷通宝鉴》PDF第103、104页：四月丙火仍以壬庚为正轨，无壬姑用癸；
	// 只有壬水过多，或支成水局又见一二壬透时，戊土才作为制水救应。
	got, err := AnalyzeTiaohou("丙", "巳")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"壬", "庚", "癸", "戊"}
	if got.TablePrimaryCandidate != "壬" || len(got.Rules) != len(want) {
		t.Fatalf("四月丙火主线或制水救应不完整: %+v", got)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Fatalf("四月丙火 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(got.Rules[3].SourceText, "壬水过多") ||
		!strings.Contains(got.Rules[3].SourceText, "支成水局") ||
		!strings.Contains(got.Rules[3].SourceText, "壬透") ||
		!strings.Contains(got.Rules[3].SourceText, "制水为救") {
		t.Fatalf("四月丙火戊土候选缺少原文适用条件: %+v", got.Rules[3])
	}
}

func TestBingShenTiaohouPairsJiaWithWuRemedy(t *testing.T) {
	// 《穷通宝鉴》PDF第115页：七月丙火“专壬次戊”，但用戊制壬时
	// 不能无甲，须以甲破戊并生扶已经退气的丙火。
	got, err := AnalyzeTiaohou("丙", "申")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"壬", "戊", "甲"}
	if got.TablePrimaryCandidate != "壬" || len(got.Rules) != len(want) {
		t.Fatalf("七月丙火主线或戊甲配套不完整: %+v", got)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Fatalf("七月丙火 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(got.Rules[2].SourceText, "用戊制壬") ||
		!strings.Contains(got.Rules[2].SourceText, "甲木破戊生丙") ||
		!strings.Contains(got.Rules[2].SourceText, "避免戊土") {
		t.Fatalf("七月丙火甲木候选缺少原文配套条件: %+v", got.Rules[2])
	}
}

func TestBingHaiTiaohouDistinguishesJiMixingFromWuControllingWater(t *testing.T) {
	// 《穷通宝鉴》PDF第124、125页：十月丙火水旺通常用戊制水；
	// 若壬多、有甲而无戊，则另用己土“混壬”，使水缓和后培木生火。
	got, err := AnalyzeTiaohou("丙", "亥")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"甲", "戊", "庚", "己", "壬"}
	if got.TablePrimaryCandidate != "甲" || len(got.Rules) != len(want) {
		t.Fatalf("十月丙火主线或己土混壬候选不完整: %+v", got)
	}
	for i, candidate := range want {
		if got.Rules[i].XiShen != candidate {
			t.Fatalf("十月丙火 candidate[%d] = %s, want %s", i, got.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(got.Rules[3].SourceText, "壬多有甲") ||
		!strings.Contains(got.Rules[3].SourceText, "无戊") ||
		!strings.Contains(got.Rules[3].SourceText, "己土混壬") ||
		!strings.Contains(got.Rules[3].SourceText, "培木再生丙火") {
		t.Fatalf("十月丙火己土候选缺少原文适用条件与作用: %+v", got.Rules[3])
	}
}

func TestDingTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"庚", "壬", "己"}},
		{month: "卯", want: []string{"庚", "甲", "戊"}},
		{month: "辰", want: []string{"甲", "庚", "戊", "己"}},
		{month: "巳", want: []string{"庚", "甲", "壬", "癸", "戊"}},
		{month: "午", want: []string{"壬", "庚", "癸", "甲"}},
		{month: "未", want: []string{"甲", "壬", "庚"}},
		{month: "申", want: []string{"甲", "庚", "丙", "乙", "戊", "壬"}},
		{month: "酉", want: []string{"甲", "丙", "庚", "乙"}},
		{month: "戌", want: []string{"甲", "庚"}},
		{month: "亥", want: []string{"甲", "庚", "癸", "戊"}},
		{month: "子", want: []string{"甲", "庚", "戊", "癸"}},
		{month: "丑", want: []string{"甲", "庚", "戊", "癸"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("丁", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("丁%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("丁%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("丁%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestDingSpringTiaohouKeepsConditionalEarthRemedies(t *testing.T) {
	yin, err := AnalyzeTiaohou("丁", "寅")
	if err != nil {
		t.Fatal(err)
	}
	if yin.TablePrimaryCandidate != "庚" || len(yin.Rules) != 3 ||
		yin.Rules[2].XiShen != "己" ||
		!strings.Contains(yin.Rules[2].SourceText, "壬癸并见") ||
		!strings.Contains(yin.Rules[2].SourceText, "化木不成") ||
		!strings.Contains(yin.Rules[2].SourceText, "己土制水") {
		t.Fatalf("正月丁火己土救应不完整: %+v", yin)
	}

	mao, err := AnalyzeTiaohou("丁", "卯")
	if err != nil {
		t.Fatal(err)
	}
	if mao.TablePrimaryCandidate != "庚" || len(mao.Rules) != 3 ||
		mao.Rules[2].XiShen != "戊" ||
		!strings.Contains(mao.Rules[2].SourceText, "癸水成势") ||
		!strings.Contains(mao.Rules[2].SourceText, "戊土制煞") {
		t.Fatalf("二月丁火戊土救应不完整: %+v", mao)
	}

	chen, err := AnalyzeTiaohou("丁", "辰")
	if err != nil {
		t.Fatal(err)
	}
	if chen.TablePrimaryCandidate != "甲" || len(chen.Rules) != 4 ||
		chen.Rules[2].XiShen != "戊" || chen.Rules[3].XiShen != "己" ||
		!strings.Contains(chen.Rules[2].SourceText, "支成水局") ||
		!strings.Contains(chen.Rules[2].SourceText, "壬水透干") ||
		!strings.Contains(chen.Rules[3].SourceText, "与戊土同透制煞") {
		t.Fatalf("三月丁火戊己制煞救应不完整: %+v", chen)
	}
}

func TestDingSummerTiaohouKeepsConditionalWuAndJia(t *testing.T) {
	si, err := AnalyzeTiaohou("丁", "巳")
	if err != nil {
		t.Fatal(err)
	}
	if si.TablePrimaryCandidate != "庚" || len(si.Rules) != 5 ||
		si.Rules[4].XiShen != "戊" ||
		!strings.Contains(si.Rules[4].SourceText, "有庚无甲") ||
		!strings.Contains(si.Rules[4].SourceText, "戊土透干") ||
		!strings.Contains(si.Rules[4].SourceText, "伤官生财") ||
		!strings.Contains(si.Rules[4].SourceText, "条件变格") {
		t.Fatalf("四月丁火戊土变格候选不完整: %+v", si)
	}

	wu, err := AnalyzeTiaohou("丁", "午")
	if err != nil {
		t.Fatal(err)
	}
	if wu.TablePrimaryCandidate != "壬" || len(wu.Rules) != 4 ||
		wu.Rules[3].XiShen != "甲" ||
		!strings.Contains(wu.Rules[3].SourceText, "不成火局") ||
		!strings.Contains(wu.Rules[3].SourceText, "水透干") ||
		!strings.Contains(wu.Rules[3].SourceText, "甲木引化") ||
		!strings.Contains(wu.Rules[3].SourceText, "庚劈甲") {
		t.Fatalf("五月丁火甲木条件候选不完整: %+v", wu)
	}
}

func TestDingAutumnTiaohouKeepsYiFallbackAndRenForExcessMetal(t *testing.T) {
	shen, err := AnalyzeTiaohou("丁", "申")
	if err != nil {
		t.Fatal(err)
	}
	wantShen := []string{"甲", "庚", "丙", "乙", "戊", "壬"}
	if shen.TablePrimaryCandidate != "甲" || len(shen.Rules) != len(wantShen) {
		t.Fatalf("七月丁火主线或条件候选不完整: %+v", shen)
	}
	for i, candidate := range wantShen {
		if shen.Rules[i].XiShen != candidate {
			t.Fatalf("七月丁火 candidate[%d] = %s, want %s", i, shen.Rules[i].XiShen, candidate)
		}
	}
	if !strings.Contains(shen.Rules[3].SourceText, "无甲木") ||
		!strings.Contains(shen.Rules[3].SourceText, "乙木枯草引灯") ||
		!strings.Contains(shen.Rules[3].SourceText, "配丙火晒燥") ||
		!strings.Contains(shen.Rules[5].SourceText, "庚金成势") ||
		!strings.Contains(shen.Rules[5].SourceText, "壬水泄庚") {
		t.Fatalf("七月丁火乙壬候选缺少原文适用条件: %+v", shen.Rules)
	}

	you, err := AnalyzeTiaohou("丁", "酉")
	if err != nil {
		t.Fatal(err)
	}
	if you.TablePrimaryCandidate != "甲" || len(you.Rules) != 4 ||
		you.Rules[3].XiShen != "乙" ||
		!strings.Contains(you.Rules[3].SourceText, "无甲木") ||
		!strings.Contains(you.Rules[3].SourceText, "乙木枯草引灯") ||
		!strings.Contains(you.Rules[3].SourceText, "配丙火晒燥") {
		t.Fatalf("八月丁火乙木替代候选不完整: %+v", you)
	}
}

func TestWuTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"丙", "甲", "癸"}},
		{month: "卯", want: []string{"丙", "甲", "癸"}},
		{month: "辰", want: []string{"甲", "丙", "癸"}},
		{month: "巳", want: []string{"甲", "丙", "癸"}},
		{month: "午", want: []string{"壬", "甲", "丙", "癸"}},
		{month: "未", want: []string{"癸", "丙", "甲", "壬"}},
		{month: "申", want: []string{"丙", "癸", "甲"}},
		{month: "酉", want: []string{"丙", "癸"}},
		{month: "戌", want: []string{"甲", "癸", "丙"}},
		{month: "亥", want: []string{"甲", "丙", "丁", "戊"}},
		{month: "子", want: []string{"丙", "甲"}},
		{month: "丑", want: []string{"丙", "甲", "壬"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("戊", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("戊%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("戊%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("戊%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestJiTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"丙", "癸", "戊"}},
		{month: "卯", want: []string{"甲", "癸", "丙", "壬"}},
		{month: "辰", want: []string{"丙", "癸", "甲"}},
		{month: "巳", want: []string{"癸", "丙", "辛", "壬", "庚"}},
		{month: "午", want: []string{"癸", "丙", "辛", "壬", "庚"}},
		{month: "未", want: []string{"癸", "丙", "辛", "壬", "庚"}},
		{month: "申", want: []string{"癸", "丙", "辛", "壬", "丁"}},
		{month: "酉", want: []string{"癸", "丙", "辛", "壬", "丁"}},
		{month: "戌", want: []string{"癸", "丙", "甲", "辛", "壬", "丁"}},
		{month: "亥", want: []string{"丙", "甲", "戊", "丁"}},
		{month: "子", want: []string{"丙", "甲", "丁", "戊"}},
		{month: "丑", want: []string{"丙", "甲", "丁", "戊"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("己", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("己%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("己%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("己%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestGengTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"丙", "甲", "丁", "壬", "癸"}},
		{month: "卯", want: []string{"丁", "甲", "庚", "丙"}},
		{month: "辰", want: []string{"甲", "丁", "丙", "癸", "壬"}},
		{month: "巳", want: []string{"壬", "戊", "丙", "癸", "丁"}},
		{month: "午", want: []string{"壬", "癸"}},
		{month: "未", want: []string{"丁", "甲"}},
		{month: "申", want: []string{"丁", "甲", "丙"}},
		{month: "酉", want: []string{"丁", "甲", "丙"}},
		{month: "戌", want: []string{"甲", "壬", "丙", "丁"}},
		{month: "亥", want: []string{"丁", "丙", "甲", "己", "戊"}},
		{month: "子", want: []string{"丁", "甲", "丙"}},
		{month: "丑", want: []string{"丙", "丁", "甲"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("庚", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("庚%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("庚%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("庚%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestXinTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"己", "壬", "庚", "丙"}},
		{month: "卯", want: []string{"壬", "甲", "庚", "戊", "丙"}},
		{month: "辰", want: []string{"壬", "甲"}},
		{month: "巳", want: []string{"壬", "庚", "甲", "癸"}},
		{month: "午", want: []string{"壬", "癸", "己", "庚", "甲"}},
		{month: "未", want: []string{"壬", "庚", "己", "甲"}},
		{month: "申", want: []string{"壬", "甲", "戊"}},
		{month: "酉", want: []string{"壬", "甲", "庚"}},
		{month: "戌", want: []string{"壬", "甲", "丙", "癸", "庚"}},
		{month: "亥", want: []string{"壬", "丙", "戊", "甲", "己"}},
		{month: "子", want: []string{"丙", "壬", "戊", "甲"}},
		{month: "丑", want: []string{"丙", "壬", "戊", "己", "丁", "甲"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("辛", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("辛%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("辛%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("辛%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestRenTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"庚", "丙", "戊", "甲"}},
		{month: "卯", want: []string{"戊", "辛", "庚", "壬"}},
		{month: "辰", want: []string{"甲", "庚", "癸", "丙"}},
		{month: "巳", want: []string{"壬", "辛", "庚", "癸", "甲", "戊"}},
		{month: "午", want: []string{"庚", "癸", "辛", "壬"}},
		{month: "未", want: []string{"辛", "甲", "癸", "庚", "壬"}},
		{month: "申", want: []string{"戊", "丁", "甲", "庚"}},
		{month: "酉", want: []string{"甲", "戊"}},
		{month: "戌", want: []string{"甲", "丙", "戊", "丁"}},
		{month: "亥", want: []string{"戊", "丙", "庚"}},
		{month: "子", want: []string{"戊", "丙"}},
		{month: "丑", want: []string{"丙", "丁", "甲", "戊"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("壬", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("壬%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("壬%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("壬%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestGuiTiaohouCandidatesFollowExplicitSourceSequence(t *testing.T) {
	cases := []struct {
		month string
		want  []string
	}{
		{month: "寅", want: []string{"辛", "庚", "丙", "壬"}},
		{month: "卯", want: []string{"庚", "辛", "己", "丁"}},
		{month: "辰", want: []string{"丙", "辛", "甲", "庚", "壬"}},
		{month: "巳", want: []string{"辛", "庚", "壬", "癸"}},
		{month: "午", want: []string{"庚", "辛", "壬", "癸"}},
		{month: "未", want: []string{"庚", "辛", "壬", "癸"}},
		{month: "申", want: []string{"丁", "甲"}},
		{month: "酉", want: []string{"辛", "丙", "丁"}},
		{month: "戌", want: []string{"辛", "甲", "癸", "庚", "壬"}},
		{month: "亥", want: []string{"庚", "辛", "戊", "丁", "丙"}},
		{month: "子", want: []string{"丙", "辛", "甲"}},
		{month: "丑", want: []string{"丙", "壬", "戊", "丁", "辛", "庚"}},
	}
	for _, tc := range cases {
		t.Run(tc.month, func(t *testing.T) {
			got, err := AnalyzeTiaohou("癸", tc.month)
			if err != nil {
				t.Fatal(err)
			}
			if len(got.Rules) != len(tc.want) {
				t.Fatalf("癸%s月 candidates = %+v, want %v", tc.month, got.Rules, tc.want)
			}
			for i, want := range tc.want {
				if got.Rules[i].XiShen != want {
					t.Errorf("癸%s月 candidate[%d] = %s, want %s", tc.month, i, got.Rules[i].XiShen, want)
				}
			}
			if got.TablePrimaryCandidate != tc.want[0] {
				t.Errorf("癸%s月 primary = %s, want %s", tc.month, got.TablePrimaryCandidate, tc.want[0])
			}
		})
	}
}

func TestTiaohouRulesUseStableSnakeCaseJSON(t *testing.T) {
	got, err := AnalyzeTiaohou("甲", "寅")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	text := string(encoded)
	for _, field := range []string{`"xi_shen"`, `"ji_shen":"未裁决"`, `"ji_shen_status":"not_adjudicated"`, `"source_text"`, `"depth_evidence"`, `"table_primary_candidate"`, `"selection_basis"`, `"not_adjudicated"`} {
		if !strings.Contains(text, field) {
			t.Fatalf("Tiaohou JSON missing %s: %s", field, text)
		}
	}
	for _, forbidden := range []string{`"primary_god"`, `"depth_affects_primary"`, `"depth_hint"`, `"reasons"`, `"summary"`} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("Tiaohou JSON leaked legacy field %s: %s", forbidden, text)
		}
	}
}
