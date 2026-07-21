package bazi

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func TestYueLingScoreAllDayMastersAndMonthBranches(t *testing.T) {
	monthBranches := [...]string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	cases := []struct {
		dayStem string
		scores  [12]float64
	}{
		{dayStem: "甲", scores: [12]float64{2, 0.5, 3, 3, 0.5, 1, 1, 0.5, 0, 0, 0.5, 2}},
		{dayStem: "乙", scores: [12]float64{2, 0.5, 3, 3, 0.5, 1, 1, 0.5, 0, 0, 0.5, 2}},
		{dayStem: "丙", scores: [12]float64{0, 1, 2, 2, 1, 3, 3, 1, 0.5, 0.5, 1, 0}},
		{dayStem: "丁", scores: [12]float64{0, 1, 2, 2, 1, 3, 3, 1, 0.5, 0.5, 1, 0}},
		{dayStem: "戊", scores: [12]float64{0.5, 3, 0, 0, 3, 2, 2, 3, 1, 1, 3, 0.5}},
		{dayStem: "己", scores: [12]float64{0.5, 3, 0, 0, 3, 2, 2, 3, 1, 1, 3, 0.5}},
		{dayStem: "庚", scores: [12]float64{1, 2, 0.5, 0.5, 2, 0, 0, 2, 3, 3, 2, 1}},
		{dayStem: "辛", scores: [12]float64{1, 2, 0.5, 0.5, 2, 0, 0, 2, 3, 3, 2, 1}},
		{dayStem: "壬", scores: [12]float64{3, 0, 1, 1, 0, 0.5, 0.5, 0, 2, 2, 0, 3}},
		{dayStem: "癸", scores: [12]float64{3, 0, 1, 1, 0, 0.5, 0.5, 0, 2, 2, 0, 3}},
	}

	for _, tc := range cases {
		dayElem := tianGanMap[tc.dayStem].WuXing
		for monthIdx, monthBranch := range monthBranches {
			got := getYueLingScore(dayElem, monthBranch)
			if want := tc.scores[monthIdx]; got != want {
				t.Errorf("%s日主遇%s月得令分 = %.1f, want %.1f", tc.dayStem, monthBranch, got, want)
			}
		}
	}
}

func TestYueLingStateLabels(t *testing.T) {
	cases := map[float64]string{3: "旺", 2: "相", 1: "休", 0.5: "囚", 0: "死"}
	for score, want := range cases {
		if got := yueLingState(score); got != want {
			t.Errorf("yueLingState(%v) = %q, want %q", score, got, want)
		}
	}
}

func TestYueLingMatrixManifestMatchesPinnedHash(t *testing.T) {
	if got := yueLingMatrixSHA256(); got != yueLingTableSHA256 {
		t.Fatalf("yue-ling matrix hash = %q, want %q", got, yueLingTableSHA256)
	}

	want := defaultYueLingRuleConfig()
	meta := DefaultRuleMeta()
	if got := meta.BodyStrength.YueLing; got != want {
		t.Fatalf("public yue-ling manifest differs from scoring table:\n got: %+v\nwant: %+v", got, want)
	}

	var embedded RuleMeta
	if err := json.Unmarshal(ruleMetaJSON, &embedded); err != nil {
		t.Fatalf("unmarshal embedded rule manifest: %v", err)
	}
	if got := embedded.BodyStrength.YueLing; got != want {
		t.Fatalf("embedded yue-ling manifest differs from scoring table:\n got: %+v\nwant: %+v", got, want)
	}
}

func TestBodyStrengthRootAndBonusManifestsMatchRuntime(t *testing.T) {
	meta := DefaultRuleMeta()
	if got, want := meta.BodyStrength.Root, defaultBodyStrengthRootRuleConfig(); got != want {
		t.Fatalf("public root manifest differs from runtime config:\n got: %+v\nwant: %+v", got, want)
	}
	wantBonus := defaultBodyStrengthBonusRuleConfig()
	if got := meta.BodyStrength.Bonus; got != wantBonus {
		t.Fatalf("public bonus manifest differs from runtime config:\n got: %+v\nwant: %+v", got, wantBonus)
	}
	if got, want := meta.BodyStrength.Influence, defaultBodyStrengthInfluenceRuleConfig(); got != want {
		t.Fatalf("public influence manifest differs from runtime config:\n got: %+v\nwant: %+v", got, want)
	}
	if got, want := meta.BodyStrength.AdjustmentForce, defaultBodyStrengthAdjustmentForceConfig(); got != want {
		t.Fatalf("public adjustment-force manifest differs from runtime config:\n got: %+v\nwant: %+v", got, want)
	}
	if got := bodyStrengthBonusTableSHA256(wantBonus); got != bodyStrengthBonusSHA256 {
		t.Fatalf("bonus table hash = %q, want %q", got, bodyStrengthBonusSHA256)
	}

	var embedded RuleMeta
	if err := json.Unmarshal(ruleMetaJSON, &embedded); err != nil {
		t.Fatalf("unmarshal embedded rule manifest: %v", err)
	}
	if embedded.BodyStrength.Root != meta.BodyStrength.Root || embedded.BodyStrength.Bonus != meta.BodyStrength.Bonus ||
		embedded.BodyStrength.Influence != meta.BodyStrength.Influence ||
		embedded.BodyStrength.AdjustmentForce != meta.BodyStrength.AdjustmentForce {
		t.Fatal("embedded body-strength sub-profile differs from authoritative runtime config")
	}
}

func TestBodyStrengthNormalizationBaselines(t *testing.T) {
	config := defaultBodyStrengthRuleConfig()
	if got := normalizeBodyStrengthScore(config.Normalizers.ShiFormula, 0, config.Normalizers.ShiSigmoidDivisor); got != 0.5 {
		t.Fatalf("signed influence zero baseline = %.6f, want 0.5", got)
	}
	if got := normalizeBodyStrengthScore(config.Normalizers.ShengFormula, 0, config.Normalizers.ShengSigmoidDivisor); got != 0 {
		t.Fatalf("one-sided support zero baseline = %.6f, want 0", got)
	}
	low := normalizeBodyStrengthScore(config.Normalizers.ShengFormula, 1, config.Normalizers.ShengSigmoidDivisor)
	high := normalizeBodyStrengthScore(config.Normalizers.ShengFormula, 2, config.Normalizers.ShengSigmoidDivisor)
	if !(0 < low && low < high && high < 1) {
		t.Fatalf("one-sided support normalization is not bounded and monotonic: low=%.6f high=%.6f", low, high)
	}
}

func TestNoSealEvidenceContributesZeroShengScore(t *testing.T) {
	result, err := (&BaziService{}).CalculateSyntheticPillars("丙午", "丁巳", "甲戌", "辛未", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	if result.BodyStrength.ShengScore != 0 {
		t.Fatalf("sheng raw score = %.6f, want 0", result.BodyStrength.ShengScore)
	}
	for _, component := range result.BodyStrength.Components {
		if component.Key != "sheng" {
			continue
		}
		if component.RawScore != 0 || component.NormalizedScore != 0 || component.WeightedScore != 0 {
			t.Fatalf("zero seal evidence received baseline credit: %+v", component)
		}
		return
	}
	t.Fatal("missing sheng component")
}

func TestBodyStrengthEarthMonthEvidenceDisclosesWholeMonthSimplification(t *testing.T) {
	result, err := (&BaziService{}).CalculateSyntheticPillars("辛酉", "辛丑", "己酉", "丙寅", "MALE")
	if err != nil {
		t.Fatalf("CalculateSyntheticPillars failed: %v", err)
	}

	if len(result.BodyStrength.Evidence) == 0 ||
		!strings.Contains(result.BodyStrength.Evidence[0].Reason, "未按节气深浅分日司令") ||
		!strings.Contains(result.BodyStrength.Evidence[0].Reason, "整月状态仅为简化候选") {
		t.Fatalf("earth-month seasonal evidence hides the whole-month simplification: %+v", result.BodyStrength.Evidence)
	}
	foundLimitation := false
	for _, limitation := range result.BodyStrength.Limitations {
		if strings.Contains(limitation, "earth-month seasonal scoring is an unsegmented whole-month candidate") {
			foundLimitation = true
			break
		}
	}
	if !foundLimitation {
		t.Fatalf("earth-month seasonal limitation is missing: %+v", result.BodyStrength.Limitations)
	}
}

func TestBodyStrengthInfluenceProfileForJiaDayMaster(t *testing.T) {
	config := defaultBodyStrengthInfluenceRuleConfig()
	cases := []struct {
		stem, tenGod, polarity, owner string
		score                         float64
		included                      bool
	}{
		{"甲", "比肩", "support", "shi", 1.0, true},
		{"乙", "劫财", "support", "shi", 0.8, true},
		{"丙", "食神", "restrict", "shi", -0.8, true},
		{"丁", "伤官", "restrict", "shi", -0.8, true},
		{"戊", "偏财", "restrict", "shi", -0.6, true},
		{"己", "正财", "restrict", "shi", -0.6, true},
		{"庚", "七杀", "restrict", "shi", -1.2, true},
		{"辛", "正官", "restrict", "shi", -1.2, true},
		{"壬", "偏印", "support", "sheng", 0, false},
		{"癸", "正印", "support", "sheng", 0, false},
	}
	for _, tc := range cases {
		got, included := calculateBodyStrengthInfluence(config, "甲", tc.stem, 1)
		if included != tc.included || got.TenGod != tc.tenGod || got.Polarity != tc.polarity ||
			got.Owner != tc.owner || got.Score != tc.score {
			t.Errorf("甲日主见%s: got %+v/%v, want %s/%s/%s/%.1f/%v", tc.stem, got, included, tc.tenGod, tc.polarity, tc.owner, tc.score, tc.included)
		}
	}
}

func TestBodyStrengthInfluenceComponentOwnership(t *testing.T) {
	result, err := (&BaziService{}).CalculateSyntheticPillars("庚申", "辛酉", "甲辰", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	restrictSources := map[string]bool{
		"年支藏干": false,
		"月支藏干": false,
		"日支藏干": false,
		"时支藏干": false,
	}
	foundRoot := false
	foundSeal := false
	for _, item := range result.BodyStrength.Evidence {
		if item.Component == "shi" {
			if _, ok := restrictSources[item.Source]; ok {
				if item.RuleID != bodyStrengthInfluenceRuleID+".hidden-branch" || item.Polarity != "restrict" {
					t.Fatalf("hidden-branch influence ownership leaked: %+v", item)
				}
				restrictSources[item.Source] = true
			}
		}
		if item.Component == "di" && item.Source == "辰" && item.Item == "乙" {
			foundRoot = true
		}
		if item.Component == "sheng" && item.Source == "辰" && item.Item == "癸" {
			foundSeal = true
		}
		if item.Component == "shi" && (item.Item == "乙" || item.Item == "癸") {
			t.Fatalf("root or seal leaked into influence component: %+v", item)
		}
	}
	for source, found := range restrictSources {
		if !found {
			t.Errorf("missing hostile hidden-stem influence from %s", source)
		}
	}
	if !foundRoot || !foundSeal {
		t.Fatalf("missing ownership evidence: root=%v seal=%v", foundRoot, foundSeal)
	}
}

func TestBodyStrengthBonusAllProfileRules(t *testing.T) {
	config := defaultBodyStrengthBonusRuleConfig()
	for i, stem := range config.DayStemOrder {
		lu := config.LuBranches[i]
		dayTotal, dayEvidence := calculateBodyStrengthBonus(config, stem, lu, "辰")
		if dayTotal != config.Scores.DayLu || len(dayEvidence) != 1 || dayEvidence[0].RuleID != config.RuleID+".day-lu" {
			t.Errorf("%s日支%s禄规则 = %.2f/%+v", stem, lu, dayTotal, dayEvidence)
		}
		monthTotal, monthEvidence := calculateBodyStrengthBonus(config, stem, "辰", lu)
		if monthTotal != config.Scores.MonthLu || len(monthEvidence) != 1 || monthEvidence[0].RuleID != config.RuleID+".month-lu" {
			t.Errorf("%s月支%s禄规则 = %.2f/%+v", stem, lu, monthTotal, monthEvidence)
		}
	}
	for i, stem := range config.YangRenStemOrder {
		ren := config.YangRenBranches[i]
		dayTotal, dayEvidence := calculateBodyStrengthBonus(config, stem, ren, "辰")
		if dayTotal != config.Scores.DayYangRen || len(dayEvidence) != 1 || dayEvidence[0].RuleID != config.RuleID+".day-yang-ren" {
			t.Errorf("%s日支%s阳刃规则 = %.2f/%+v", stem, ren, dayTotal, dayEvidence)
		}
		monthTotal, monthEvidence := calculateBodyStrengthBonus(config, stem, "辰", ren)
		if monthTotal != config.Scores.MonthYangRen || len(monthEvidence) != 1 || monthEvidence[0].RuleID != config.RuleID+".month-yang-ren" {
			t.Errorf("%s月支%s阳刃规则 = %.2f/%+v", stem, ren, monthTotal, monthEvidence)
		}
	}
	for _, tc := range []struct{ stem, disputedBlade string }{{"乙", "辰"}, {"丁", "未"}, {"己", "未"}, {"辛", "戌"}, {"癸", "丑"}} {
		if total, evidence := calculateBodyStrengthBonus(config, tc.stem, tc.disputedBlade, tc.disputedBlade); total != 0 || len(evidence) != 0 {
			t.Errorf("阴干%s在 no_yang_ren_bonus Profile 下不应生成刃加成: %.2f/%+v", tc.stem, total, evidence)
		}
	}
}

func TestTouGanMultiplierUsesPublishedScopeAndParameters(t *testing.T) {
	findChenYi := func(t *testing.T, yearPillar string) BodyStrengthEvidence {
		t.Helper()
		result, err := (&BaziService{}).CalculateSyntheticPillars(yearPillar, "辛酉", "甲辰", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars failed: %v", err)
		}
		for _, item := range result.BodyStrength.Evidence {
			if item.Component == "di" && item.Source == "辰" && item.Item == "乙" {
				return item
			}
		}
		t.Fatal("missing 辰中乙 root evidence")
		return BodyStrengthEvidence{}
	}

	notTransparent := findChenYi(t, "庚申")
	transparent := findChenYi(t, "乙丑")
	if math.Abs(notTransparent.Score-0.45) > 1e-12 || strings.Contains(notTransparent.Reason, "透干加成") {
		t.Fatalf("non-transparent root evidence = %+v", notTransparent)
	}
	if math.Abs(transparent.Score-0.54) > 1e-12 || !strings.Contains(transparent.Reason, "all_four_heaven_stems_including_day_master") ||
		!strings.Contains(transparent.Reason, "1.20") || transparent.RuleID != bodyStrengthRootRuleID+".lookup" {
		t.Fatalf("transparent root evidence = %+v", transparent)
	}
}

func TestDayBranchRootIsOwnedOnlyByDiComponent(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.CalculateSyntheticPillars("庚申", "辛酉", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}

	foundDayRoot := false
	for _, item := range result.BodyStrength.Evidence {
		if item.Component == "di" && item.Source == "寅" && item.Item == "甲" {
			foundDayRoot = true
		}
		if item.Component == "shi" && strings.HasSuffix(item.Source, "支藏干") && item.Polarity == "support" {
			t.Fatalf("hidden-branch support was counted in both di and shi: %+v", item)
		}
	}
	if !foundDayRoot {
		t.Fatal("expected the day-branch 甲 root to remain in the di component")
	}
}

func TestChangShengStageAllDayStemsAndBranches(t *testing.T) {
	branches := [...]string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	cases := []struct {
		stem   string
		stages [12]string
	}{
		{stem: "甲", stages: [12]string{"沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝", "胎", "养", "长生"}},
		{stem: "乙", stages: [12]string{"病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎", "绝", "墓", "死"}},
		{stem: "丙", stages: [12]string{"胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝"}},
		{stem: "丁", stages: [12]string{"绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎"}},
		{stem: "戊", stages: [12]string{"胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝"}},
		{stem: "己", stages: [12]string{"绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎"}},
		{stem: "庚", stages: [12]string{"死", "墓", "绝", "胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病"}},
		{stem: "辛", stages: [12]string{"长生", "养", "胎", "绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴"}},
		{stem: "壬", stages: [12]string{"帝旺", "衰", "病", "死", "墓", "绝", "胎", "养", "长生", "沐浴", "冠带", "临官"}},
		{stem: "癸", stages: [12]string{"临官", "冠带", "沐浴", "长生", "养", "胎", "绝", "墓", "死", "病", "衰", "帝旺"}},
	}
	weights := map[string]float64{
		"长生": 1.5, "帝旺": 1.5, "临官": 1.5,
		"沐浴": 1, "冠带": 1, "衰": 1, "墓": 1,
		"胎": 0.5, "养": 0.5, "病": 0.5, "死": 0.5,
		"绝": 0,
	}

	for _, tc := range cases {
		stem, err := tyme.HeavenStem{}.FromName(tc.stem)
		if err != nil {
			t.Fatalf("parse stem %s: %v", tc.stem, err)
		}
		for branchIndex, branchName := range branches {
			branch, err := tyme.EarthBranch{}.FromName(branchName)
			if err != nil {
				t.Fatalf("parse branch %s: %v", branchName, err)
			}
			gotStage, gotWeight := changShengWeight(*stem, *branch, defaultBodyStrengthRootRuleConfig().TerrainWeights)
			wantStage := tc.stages[branchIndex]
			if gotStage != wantStage || gotWeight != weights[wantStage] {
				t.Errorf("%s日干临%s: got %s/%.1f, want %s/%.1f", tc.stem, branchName, gotStage, gotWeight, wantStage, weights[wantStage])
			}
		}
	}
}

func TestBodyStrengthBandMatchesFinalAdjustedScore(t *testing.T) {
	cases := []struct {
		name    string
		pillars [4]string
		ruleID  string
	}{
		{name: "失令不衰重水", pillars: [4]string{"壬子", "戊申", "甲午", "癸亥"}, ruleID: "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1"},
		{name: "失令不衰双水", pillars: [4]string{"辛亥", "丙申", "甲午", "癸巳"}, ruleID: "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := (&BaziService{}).CalculateSyntheticPillars(tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3], "MALE")
			if err != nil {
				t.Fatalf("CalculateFromPillars failed: %v", err)
			}
			analysis := result.BodyStrength
			if len(analysis.Adjustments) != 1 || analysis.Adjustments[0].RuleID != tc.ruleID {
				t.Fatalf("adjustment contract = %+v, want one %s", analysis.Adjustments, tc.ruleID)
			}
			want := scoreBandCandidateForBodyStrength(analysis.TotalScore)
			if analysis.ScoreBandCandidate != want {
				t.Fatalf("final score %.6f maps to %s, got %s", analysis.TotalScore, want, analysis.ScoreBandCandidate)
			}
		})
	}
}

func TestBodyStrengthDoesNotDoublePenalizeClassicalWangJiCharts(t *testing.T) {
	// 《滴天髓阐微》PDF第58、84、106页均明确称这些命局“旺之极矣”。
	// 官杀已在得势中扣分，不应再以另一套力值整体折减一次。
	cases := []struct {
		name    string
		pillars [4]string
	}{
		{name: "丙火午月癸巳时", pillars: [4]string{"丙寅", "甲午", "丙午", "癸巳"}},
		{name: "丙火午月壬辰时", pillars: [4]string{"癸丑", "戊午", "丙午", "壬辰"}},
		{name: "庚金申月丙戌时", pillars: [4]string{"壬戌", "戊申", "庚寅", "丙戌"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := (&BaziService{}).CalculateFromPillars(
				tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3], "MALE",
			)
			if err != nil {
				t.Fatalf("CalculateFromPillars failed: %v", err)
			}
			analysis := result.BodyStrength
			if len(analysis.Adjustments) != 0 {
				t.Fatalf("classical 旺极 chart received a second whole-score adjustment: %+v", analysis.Adjustments)
			}
			if analysis.ScoreBandCandidate != "身旺" && analysis.ScoreBandCandidate != "偏旺" {
				t.Fatalf("classical 旺极 chart candidate = %s (%.6f), want 旺-side candidate", analysis.ScoreBandCandidate, analysis.TotalScore)
			}
		})
	}
}

func TestCompleteSameElementBranchGroupsSupportClassicalCharts(t *testing.T) {
	// 《滴天髓阐微》分别以“会局帮身，不当弱论”“会局帮身”、
	// “支会火局”“支全金局”和“亥子丑北方一气”确认完整同气成局的扶身作用。
	cases := []struct {
		name    string
		pillars [4]string
		group   string
		floored bool
	}{
		{name: "申子辰水局扶壬", pillars: [4]string{"庚申", "戊寅", "壬子", "甲辰"}, group: "申子辰", floored: true},
		{name: "亥卯未木局扶乙", pillars: [4]string{"丁亥", "丁未", "乙亥", "己卯"}, group: "亥卯未", floored: true},
		{name: "寅午戌火局扶丙", pillars: [4]string{"乙丑", "丙戌", "丙午", "庚寅"}, group: "寅午戌", floored: true},
		{name: "巳酉丑金局扶辛", pillars: [4]string{"壬辰", "己酉", "辛丑", "癸巳"}, group: "巳酉丑"},
		{name: "亥子丑三会扶癸", pillars: [4]string{"丁亥", "壬子", "癸丑", "甲寅"}, group: "亥子丑"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := (&BaziService{}).CalculateFromPillars(
				tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3], "MALE",
			)
			if err != nil {
				t.Fatalf("CalculateFromPillars failed: %v", err)
			}
			analysis := result.BodyStrength
			foundStructure := false
			for _, item := range analysis.Evidence {
				if item.RuleID == bodyStrengthCompleteGroupRuleID+".observed" && item.Item == tc.group {
					foundStructure = true
					if item.Component != "structure" || item.Polarity != "support" || item.Score != 0 {
						t.Fatalf("complete-group evidence = %+v", item)
					}
				}
			}
			if !foundStructure {
				t.Fatalf("missing complete same-element group evidence for %s: %+v", tc.group, analysis.Evidence)
			}
			if analysis.ScoreBandCandidate == "偏弱" || analysis.ScoreBandCandidate == "身弱" {
				t.Fatalf("classical complete-group chart remained weak: %s (%.6f)", analysis.ScoreBandCandidate, analysis.TotalScore)
			}
			foundFloor := false
			for _, adjustment := range analysis.Adjustments {
				if adjustment.RuleID != bodyStrengthCompleteGroupRuleID {
					continue
				}
				foundFloor = true
				if adjustment.Before >= 0.5 || adjustment.After != 0.5 || analysis.TotalScore != 0.5 {
					t.Fatalf("complete-group neutral floor = %+v, total=%.6f", adjustment, analysis.TotalScore)
				}
			}
			if foundFloor != tc.floored {
				t.Fatalf("complete-group floor found=%v, want %v; score=%.6f adjustments=%+v", foundFloor, tc.floored, analysis.TotalScore, analysis.Adjustments)
			}
		})
	}
}

func TestIncompleteSameElementBranchGroupDoesNotTriggerStructuralSupport(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("庚戌", "戊寅", "壬子", "甲辰", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	for _, item := range result.BodyStrength.Evidence {
		if item.RuleID == bodyStrengthCompleteGroupRuleID+".observed" {
			t.Fatalf("incomplete 申子辰 structure received complete-group support: %+v", item)
		}
	}
	for _, adjustment := range result.BodyStrength.Adjustments {
		if adjustment.RuleID == bodyStrengthCompleteGroupRuleID {
			t.Fatalf("incomplete 申子辰 structure received complete-group adjustment: %+v", adjustment)
		}
	}
}

func TestShiLingAdjustmentDoesNotCountDayMasterAsExternalSupport(t *testing.T) {
	result, err := (&BaziService{}).CalculateSyntheticPillars("庚戌", "丙申", "甲午", "辛巳", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	if result.BodyStrength.LingScore > 0.5 {
		t.Fatalf("test chart must be out of season, ling score = %.2f", result.BodyStrength.LingScore)
	}
	for _, adjustment := range result.BodyStrength.Adjustments {
		if adjustment.RuleID == "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1" {
			t.Fatalf("day master alone triggered 失令不衰: %+v", adjustment)
		}
	}
}

func TestDefaultRuleMetaLoadsVersionedRuleTables(t *testing.T) {
	meta := DefaultRuleMeta()
	if meta.RuleVersion != RuleVersion {
		t.Fatalf("rule version = %q, want %q", meta.RuleVersion, RuleVersion)
	}
	if meta.School != RuleSchool {
		t.Fatalf("school = %q, want %q", meta.School, RuleSchool)
	}
	if len(meta.Tables) == 0 {
		t.Fatal("expected rule tables")
	}
	if meta.BodyStrength.Weights.Ling != 0.4 || meta.BodyStrength.Weights.Di != 0.3 {
		t.Fatalf("unexpected body strength weights: %+v", meta.BodyStrength.Weights)
	}

	var foundCalendarCore, foundZiHourPolicy, foundTenGod, foundTwelveStage, foundShenSha, foundBodyStrength, foundPatternCandidates, foundWuxingDistribution, foundInterpretationBoundaries, foundDaYunStart bool
	for _, table := range meta.Tables {
		if table.Key == "calendar_core" {
			foundCalendarCore = true
			if table.Count != 59 || table.Version != "2026-07-16.3" || len(table.Sources) != 2 {
				t.Fatalf("calendar core metadata = %+v", table)
			}
			for _, source := range table.Sources {
				if source.Repository == "" || len(source.Commit) != 40 || len(source.Files) == 0 ||
					source.License != "MIT" || source.SourceTier != "silver_external" ||
					source.ValidationStatus != "cross_checked_not_gold" {
					t.Fatalf("invalid calendar source = %+v", source)
				}
			}
		}
		if table.Key == "zi_hour_policy" {
			foundZiHourPolicy = true
			if table.Count != 2 || table.Version != "2026-07-16.2" || len(table.Sources) != 1 {
				t.Fatalf("zi-hour policy metadata = %+v", table)
			}
			source := table.Sources[0]
			if source.ID != "lunar_javascript" || source.Repository != LunarJavascriptSourceRepo ||
				source.Commit != LunarJavascriptSourceCommit || len(source.Files) != 1 ||
				source.Files["lunar.js"] != "9750324bfe1aa63c146f8c72b1143df924466c11c8a5277d7d9225c541a18aaa" {
				t.Fatalf("invalid zi-hour Silver source = %+v", source)
			}
		}
		if table.Key == "ten_god_matrix" {
			foundTenGod = true
			if table.Count != len(tenGodNames) {
				t.Fatalf("ten god count = %d, want %d", table.Count, len(tenGodNames))
			}
		}
		if table.Key == "twelve_stage" {
			foundTwelveStage = true
			if table.Count != 120 || table.Version != "2026-07-15.2" || len(table.Sources) != 1 {
				t.Fatalf("twelve-stage table metadata = %+v", table)
			}
			source := table.Sources[0]
			if source.Repository != Tyme4GoSourceRepo || source.Commit != Tyme4GoSourceCommit ||
				len(source.Files) != 2 || source.License != "MIT" || source.SourceTier != "silver_external" ||
				source.ValidationStatus != "cross_checked_not_gold" {
				t.Fatalf("invalid twelve-stage source = %+v", source)
			}
		}
		if table.Key == "shensha" {
			foundShenSha = true
			if table.Version != "2026-07-16.59" || !strings.Contains(table.Description, "太极贵人只按生年干") ||
				!strings.Contains(table.Description, "天乙贵人当前只按日干") ||
				!strings.Contains(table.Description, "删除年干重复入口和时贵别名") ||
				!strings.Contains(table.Description, "命后一辰统一使用原名破宅煞") ||
				!strings.Contains(table.Description, "暗金的煞为四仲巳") ||
				!strings.Contains(table.Description, "输入主键未裁决前不自动替换") ||
				!strings.Contains(table.Description, "死符按病符对冲保留结构表但在输出层屏蔽") ||
				!strings.Contains(table.Description, "五份固定资料没有定位当前大小耗目标表") ||
				!strings.Contains(table.Description, "旧小耗重复吊客") ||
				!strings.Contains(table.Description, "官符、病符、丧门按《三命通会》PDF第122页") ||
				!strings.Contains(table.Description, "旧白虎年支表逐项对应太岁顺行第九宫") ||
				!strings.Contains(table.Description, "灾煞统一由年支或日支所属三合组取将星对冲支") ||
				!strings.Contains(table.Description, "删除年支表的十二项副本") ||
				!strings.Contains(table.Description, "旧自缢煞错误等同年支对冲") ||
				!strings.Contains(table.Description, "只有元数据而无正式实现的岁破") ||
				!strings.Contains(table.Description, "劫煞、亡神按《三命通会》三合表") ||
				!strings.Contains(table.Description, "亡劫须参太岁") ||
				!strings.Contains(table.Description, "咸池按原名和四组三合沐浴位输出") ||
				!strings.Contains(table.Description, "时桃花、时煞、时马、时刃、时禄别名停止发布") ||
				!strings.Contains(table.Description, "六厄按生年支三合五行死位完整锁定") ||
				!strings.Contains(table.Description, "旧墓煞逐项重复华盖墓库位") ||
				!strings.Contains(table.Description, "日干七杀以完整干支入辰戌丑未墓") ||
				!strings.Contains(table.Description, "将星按年支或日支三合旺位") ||
				!strings.Contains(table.Description, "驿马按年支或日支三合首支对冲位") ||
				!strings.Contains(table.Description, "时马别名已退出") ||
				!strings.Contains(table.Description, "五份固定本地PDF未定位当前红鸾、天喜生年支十二项表") ||
				!strings.Contains(table.Description, "现两项停止发布并保持unregistered/not_available") ||
				!strings.Contains(table.Description, "旧隔角煞生年支六组相邻对表无固定文本支持") ||
				!strings.Contains(table.Description, "《渊海子平》PDF第635、730页") ||
				!strings.Contains(table.Description, "孤辰、寡宿只以生年支所在三会方") ||
				!strings.Contains(table.Description, "不是第三个孤寡煞命中名") ||
				!strings.Contains(table.Description, "元辰按《三命通会》PDF第114页修正") ||
				!strings.Contains(table.Description, "勾煞、绞煞按该书PDF第117页") ||
				!strings.Contains(table.Description, "旧合并名勾绞煞停止发布") ||
				!strings.Contains(table.Description, "可定位的暴败桃花要求子午卯酉全") ||
				!strings.Contains(table.Description, "旧金锁煞、岁驾、科名、文星、魁星年干表没有可定位封闭表依据") ||
				!strings.Contains(table.Description, "五项停止发布并保持unregistered/not_available") ||
				!strings.Contains(table.Description, "天刑煞按《三命通会》PDF第122页修正为生年支查时干") ||
				!strings.Contains(table.Description, "酉戌丁、亥戊") ||
				!strings.Contains(table.Description, "旧小时、大败、天医月支表停止发布") ||
				!strings.Contains(table.Description, "原名为生时忌见月内") ||
				!strings.Contains(table.Description, "天医只见择方、合婚或年神语境") ||
				!strings.Contains(table.Description, "神煞层停止重复发布建禄") ||
				!strings.Contains(table.Description, "日干禄位命中月支已由禄神逐柱规则完整表达") ||
				!strings.Contains(table.Description, "旧月空把月柱六甲旬空的两个地支冒名发布") ||
				!strings.Contains(table.Description, "同名月空是月支三合组查单个目标天干") ||
				!strings.Contains(table.Description, "截路空亡当前采用《渊海子平》日干查时支Profile") ||
				!strings.Contains(table.Description, "戊癸作戌亥，异表不混入") ||
				!strings.Contains(table.Description, "五份固定本地PDF均未定位童子煞名称") ||
				!strings.Contains(table.Description, "夏季目标不一致") ||
				!strings.Contains(table.Description, "月厌按十二月逆行支表锁定") ||
				!strings.Contains(table.Description, "旧表仅子月目标偶合原文、其余十一月错误") ||
				!strings.Contains(table.Description, "旧月刑、月害分别只是月支与目标支的相刑、六害关系别名") ||
				!strings.Contains(table.Description, "底层结构统一由地支关系图输出") ||
				!strings.Contains(table.Description, "旧天刑、天火、天贼、大时、兵禁、天吏六张月支单支表") ||
				!strings.Contains(table.Description, "错误月令天火单支表同步停止发布") ||
				strings.Contains(table.Description, "月令查表的天火是独立名称") ||
				!strings.Contains(table.Description, "旧致死月支十二表") ||
				!strings.Contains(table.Description, "不可观察死规则") ||
				!strings.Contains(table.Description, "高风险抑制注册表删除三个无生产调用的退役名称残留") ||
				!strings.Contains(table.Description, "只保留有固定结构表和来源合同但禁止公开输出的死符") ||
				!strings.Contains(table.Description, "羊刃与飞刃当前采用子平五阳干Profile") ||
				!strings.Contains(table.Description, "《渊海子平》PDF第119页另列十干异表") ||
				!strings.Contains(table.Description, "当前不混入乙丁己辛癸五阴目标") ||
				!strings.Contains(table.Description, "空亡当前采用日柱六甲旬双支Profile") ||
				!strings.Contains(table.Description, "另载按阴阳日干只取一支的轻重异表") ||
				!strings.Contains(table.Description, "当前不缩减为单支") ||
				!strings.Contains(table.Description, "十干禄位收敛为神煞、格局与身强共同消费的唯一Profile") ||
				!strings.Contains(table.Description, "日干综合表不再重复存储禄神目标") ||
				!strings.Contains(table.Description, "现停止发布并保持unregistered/not_available") ||
				!strings.Contains(table.Description, "当前Profile不混入申") || !strings.Contains(table.Source, "第71页") ||
				!strings.Contains(table.Source, "第73页") || !strings.Contains(table.Source, "第75页") ||
				!strings.Contains(table.Source, "第103页天德合") || !strings.Contains(table.Source, "第108页") ||
				!strings.Contains(table.Source, "第181页") || !strings.Contains(table.Source, "第182页") ||
				!strings.Contains(table.Source, "第546页") ||
				!strings.Contains(table.Source, "第666页") || !strings.Contains(table.Source, "第96页") ||
				!strings.Contains(table.Source, "第97页") || !strings.Contains(table.Source, "第109页") ||
				!strings.Contains(table.Source, "第635页") || !strings.Contains(table.Source, "PDF第121页") ||
				!strings.Contains(table.Source, "PDF第122页") || !strings.Contains(table.Source, "PDF第113页") ||
				!strings.Contains(table.Source, "PDF第117页") || !strings.Contains(table.Source, "PDF第119-120页") ||
				!strings.Contains(table.Source, "PDF第103-104页") || !strings.Contains(table.Source, "PDF第100-102页") ||
				!strings.Contains(table.Source, "PDF第105-106页") || !strings.Contains(table.Source, "PDF第76页") ||
				!strings.Contains(table.Source, "PDF第63页") || !strings.Contains(table.Source, "PDF第65页") ||
				!strings.Contains(table.Source, "PDF第67页") || !strings.Contains(table.Source, "PDF第97-98页") ||
				!strings.Contains(table.Source, "PDF第114页元辰") || !strings.Contains(table.Source, "PDF第117页勾绞") ||
				!strings.Contains(table.Source, "天刑煞与雷霆煞") || !strings.Contains(table.Source, "PDF第107页截路空亡") ||
				!strings.Contains(table.Source, "PDF第113页戊癸异表") || !strings.Contains(table.Source, "PDF第103页（书内第100页）月厌月煞") ||
				!strings.Contains(table.Source, "PDF第119页十干羊刃飞刃异表") || !strings.Contains(table.Source, "PDF第205页五阳Profile") ||
				!strings.Contains(table.Source, "《三命通会》PDF第108页及第226页五阳Profile") ||
				!strings.Contains(table.Source, "PDF第105页六甲空亡双支表") || !strings.Contains(table.Source, "PDF第108-110页双支定义及阴阳单支异表") ||
				!strings.Contains(table.Source, "《渊海子平》PDF第81页") || !strings.Contains(table.Source, "《三命通会》PDF第85-86页十干禄表") ||
				!strings.Contains(table.Source, "PDF第688页") || !strings.Contains(table.Source, "PDF第115-116页") ||
				!strings.Contains(table.Source, "PDF第77页") ||
				!strings.Contains(table.Source, "PDF第91页") ||
				!strings.Contains(table.Source, "PDF第93页") || !strings.Contains(table.Source, "PDF第227页") ||
				!strings.Contains(table.Source, "PDF第733页") ||
				!strings.Contains(table.Source, "PDF第221页") || !strings.Contains(table.Source, "PDF第111页") ||
				!strings.Contains(table.Source, "PDF第120-121页") || !strings.Contains(table.Source, "PDF第124页") ||
				!strings.Contains(table.Source, "PDF第125页") ||
				!strings.Contains(table.Source, "PDF第209页") || !strings.Contains(table.Source, "PDF第516页") ||
				!strings.Contains(table.Source, "PDF第185-186页") || !strings.Contains(table.Source, "PDF第186-187页") ||
				!strings.Contains(table.Source, "PDF第188-189页") {
				t.Fatalf("shen-sha metadata = %+v", table)
			}
		}
		if table.Key == "body_strength" {
			foundBodyStrength = true
			if table.Count != 60 || table.Version != "2026-07-18.2" {
				t.Fatalf("body-strength table metadata = %+v", table)
			}
		}
		if table.Key == "pattern_candidates" {
			foundPatternCandidates = true
			if table.Count != patternDetectorCount() || table.Version != "2026-07-17.27" ||
				!strings.Contains(table.Description, "专禄固定甲寅、乙卯、庚申、辛酉") ||
				!strings.Contains(table.Description, "与月令建禄独立并可同时命中") ||
				!strings.Contains(table.Description, "遗漏刑冲、作合、倒食、官星、日月同干、岁月同干六忌") ||
				!strings.Contains(table.Description, "现失败关闭") ||
				!strings.Contains(table.Description, "收敛为丙午、戊午、壬子三个日刃格结构") ||
				!strings.Contains(table.Description, "日刃与月刃独立并可同时命中") ||
				!strings.Contains(table.Description, "刑冲破害、会合和官杀制化只作为未裁决条件") ||
				!strings.Contains(table.Description, "pattern.lu.jianlu 与 pattern.lu.yueren") ||
				!strings.Contains(table.Description, "甲卯、丙午、戊午、庚酉、壬子五阳干Profile") ||
				!strings.Contains(table.Description, "不再共用模糊 pattern.lu.yueling") ||
				!strings.Contains(table.Description, "旧 pattern.bage.yueling") ||
				!strings.Contains(table.Description, "各格专章要求的透干、身强、扶助、制化和破格条件并不相同") ||
				!strings.Contains(table.Description, "月支藏干与逐项十神事实继续由独立基础层输出") ||
				!strings.Contains(table.Description, "旧五个 pattern.compound.* 检测器") ||
				!strings.Contains(table.Description, "子卯午酉及四库阴干") ||
				!strings.Contains(table.Description, "身强旺衰、制化力度、去财去印") ||
				!strings.Contains(table.Source, "两气双清") ||
				!strings.Contains(table.Description, "四干与四支本气八个位点") ||
				!strings.Contains(table.Description, "删除旧15%聚合分数阈值") ||
				!strings.Contains(table.Description, "曲直、炎上、从革、润下要求地支完整成方或三合局") ||
				!strings.Contains(table.Description, "稼穑要求辰戌丑未四库皆全") ||
				!strings.Contains(table.Description, "删除旧60%生扶、30%日主和10%克神分数阈值") ||
				!strings.Contains(table.Description, "旧从财、从势、从杀、从弱、从儿五个检测器") ||
				!strings.Contains(table.Description, "10%/15%生扶和60%主势") ||
				!strings.Contains(table.Description, "从势漏掉食伤并旺、不能专从一神") ||
				!strings.Contains(table.Description, "从儿错误要求身弱无根") ||
				!strings.Contains(table.Description, "删除五项注册、算法、互斥争议和统一喜忌") ||
				!strings.Contains(table.Description, "旧 checkHuaQiGe 与 checkCongHuaGe") ||
				!strings.Contains(table.Description, "30%与25%日主分数阈值") ||
				!strings.Contains(table.Description, "同盘重复发布化气格与从化格") ||
				!strings.Contains(table.Description, "删除两项注册、算法、快捷月令表和统一喜忌") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v3") ||
				!strings.Contains(table.Description, "10个存量检测器均不消费") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v4移除分数与分段门禁及快照字段") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v4 保留互斥争议状态机") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v5删除死注册参数") ||
				!strings.Contains(table.Description, "关系图的真实disputed") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v5 用未裁决的本地整数优先级") ||
				!strings.Contains(table.Description, "primary_candidate_id、selection_basis、候选role和priority") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v6删除伪主格排序") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v6 的私有patternDetection") ||
				!strings.Contains(table.Description, "Description、SubType、FavorableElements和UnfavorableElements") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v7只保留PatternName和PatternType") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v7 的candidate_id始终等于rule_id") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v8删除重复candidate_id") ||
				!strings.Contains(table.Description, "rule_id作为唯一候选身份") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v8 在每个候选重复固定状态") ||
				!strings.Contains(table.Description, "与集合级validation_status和interpretation_status完全相同") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v9删除候选级重复状态") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v9 的basis固定为local_detector_conditions_matched") ||
				!strings.Contains(table.Description, "候选进入集合本身已经证明检测条件命中") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v10删除同义basis字段") ||
				!strings.Contains(table.Description, "旧 pattern-candidate-set-v10 同时发布pattern_type与category") ||
				!strings.Contains(table.Description, "魁罡和日德被注册为辅助特征却标成特殊格局") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v11删除冲突pattern_type") ||
				!strings.Contains(table.Description, "旧金神规则ID pattern.special.jinshen") ||
				!strings.Contains(table.Description, "更名为pattern.aux.jinshen") ||
				!strings.Contains(table.Description, "10个注册规则ID前缀与category一致") ||
				!strings.Contains(table.Description, "liangQiSemanticProfile独立快照") ||
				!strings.Contains(table.Description, "3b89e62e6fe12baf7969be4a2afb35b75308b0029d55ee16c5a5d273b2c49636") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v19删除不可达的静默去重和名称次级排序") ||
				!strings.Contains(table.Description, "十干10项、十二支12项纯值Profile") ||
				!strings.Contains(table.Description, "32ffa7b00d2145bfe737b8bf0a9135f4feb10d3fc03963ca83dfc00f05cad938") ||
				!strings.Contains(table.Description, "sanQiSemanticProfile和jinShenSemanticProfile") ||
				!strings.Contains(table.Description, "881193d5290ed58d831dfdde6b57b626380b193d465c44a2d545d59b93b023f7") ||
				!strings.Contains(table.Description, "patternPillarContextSemanticProfile") ||
				!strings.Contains(table.Description, "ec2928c92f0b6e227e9f656a788fa641215a3587a0309876d7f84cee6aa08ab6") ||
				!strings.Contains(table.Description, "patternDetectorOutputNames纯Profile") ||
				!strings.Contains(table.Description, "acd631f529e51ead2c50fa1c7832149ad7d994d7137be348133ecd70de2cff1a") ||
				!strings.Contains(table.Description, "pattern-candidate-set-v24新增规范排序的detector_profiles") ||
				!strings.Contains(table.Source, "《滴天髓阐微》PDF第44-45页") ||
				!strings.Contains(table.Source, "《滴天髓阐微》PDF第173-174页") ||
				!strings.Contains(table.Source, "PDF第186-187页顺局") ||
				!strings.Contains(table.Source, "《三命通会》PDF第205页弃命从财") ||
				!strings.Contains(table.Source, "《渊海子平》PDF第566-569页") ||
				!strings.Contains(table.Source, "《滴天髓阐微》PDF第177-178页化象") ||
				!strings.Contains(table.Source, "《三命通会》PDF第72-74页论十干化气") ||
				!strings.Contains(table.Source, "《渊海子平》PDF第575-576页化气诗诀") ||
				!strings.Contains(table.Source, "《滴天髓阐微》PDF第43页") ||
				!strings.Contains(table.Source, "《三命通会》PDF第190页专禄") ||
				!strings.Contains(table.Source, "PDF第153-220页") || !strings.Contains(table.Source, "PDF第711-713页") ||
				!strings.Contains(table.Source, "PDF第226及228-230页阳刃月柱") ||
				!strings.Contains(table.Source, "PDF第230-232页建禄") ||
				!strings.Contains(table.Source, "《渊海子平》PDF第162页归禄六忌") ||
				!strings.Contains(table.Source, "PDF第217页日刃三日表") || !strings.Contains(table.Source, "《三命通会》PDF第230页日刃") {
				t.Fatalf("pattern-candidate metadata = %+v", table)
			}
		}
		if table.Key == "wuxing_distribution" {
			foundWuxingDistribution = true
			if table.Count != 5 || table.Version != "2026-07-18.1" {
				t.Fatalf("five-element distribution metadata = %+v", table)
			}
		}
		if table.Key == "interpretation_boundaries" {
			foundInterpretationBoundaries = true
			if table.Count != 5 {
				t.Fatalf("interpretation boundary count = %d, want 5", table.Count)
			}
		}
		if table.Key == "dayun_start" {
			foundDaYunStart = true
			if table.Version != "2026-07-16.3" || len(table.Sources) != 1 {
				t.Fatalf("da-yun start metadata = %+v", table)
			}
			source := table.Sources[0]
			if source.ID != "tyme4go_v1_4_2_default_child_limit" ||
				source.Repository != Tyme4GoSourceRepo || source.Commit != Tyme4GoSourceCommit ||
				len(source.Files) != 3 ||
				source.Files["tyme/DefaultChildLimitProvider.go"] != "54f84ec021962e6214edc8461d0a4ae33e3c096e4ca916d681e665316338aa7e" ||
				source.Files["tyme/AbstractChildLimitProvider.go"] != "0a92bc1f552357b15abac913bd7fee4b67e53444c11bad7d6d708a1fba56373d" ||
				source.Files["tyme/ChildLimitInfo.go"] != "09b6d0ffc74bbc09307dc9df3acf1f09382a0313c842b70c75e9584e26ddc97b" ||
				source.License != "MIT" || source.SourceTier != "silver_external" ||
				source.ValidationStatus != "cross_checked_not_gold" {
				t.Fatalf("invalid da-yun child-limit source = %+v", source)
			}
		}
	}
	if !foundCalendarCore || !foundZiHourPolicy || !foundTenGod || !foundTwelveStage || !foundShenSha || !foundBodyStrength || !foundPatternCandidates || !foundWuxingDistribution || !foundInterpretationBoundaries || !foundDaYunStart {
		t.Fatalf("missing expected rule tables: calendarCore=%v ziHour=%v tenGod=%v twelveStage=%v shenSha=%v bodyStrength=%v patternCandidates=%v wuxingDistribution=%v interpretationBoundaries=%v daYunStart=%v", foundCalendarCore, foundZiHourPolicy, foundTenGod, foundTwelveStage, foundShenSha, foundBodyStrength, foundPatternCandidates, foundWuxingDistribution, foundInterpretationBoundaries, foundDaYunStart)
	}
}

func TestDefaultRuleMetaReturnsIndependentCopy(t *testing.T) {
	meta := DefaultRuleMeta()
	if len(meta.Tables) == 0 {
		t.Fatal("expected rule tables")
	}
	meta.Tables[0].Name = "mutated"
	if len(meta.Tables[0].Sources) == 0 {
		t.Fatal("calendar core sources are missing")
	}
	meta.Tables[0].Sources[0].Files["lunar.js"] = "mutated"

	next := DefaultRuleMeta()
	if next.Tables[0].Name == "mutated" {
		t.Fatal("DefaultRuleMeta leaked caller mutation")
	}
	if next.Tables[0].Sources[0].Files["lunar.js"] == "mutated" {
		t.Fatal("DefaultRuleMeta leaked nested source-file mutation")
	}
}

func TestBodyStrengthReturnsExplainableComponents(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}

	bs := result.BodyStrength
	if bs.RuleVersion != RuleVersion || bs.School != RuleSchool {
		t.Fatalf("unexpected rule meta on body strength: %+v", bs)
	}
	if len(bs.Components) < 5 {
		t.Fatalf("expected components including bonus, got %d", len(bs.Components))
	}
	if len(bs.Evidence) == 0 {
		t.Fatal("expected body strength evidence")
	}
	if bs.Adjustments == nil || len(bs.Adjustments) != 0 {
		t.Fatalf("ordinary chart adjustments = %+v, want an empty JSON array", bs.Adjustments)
	}
	if bs.RuleID != bodyStrengthRuleID || bs.SchemaVersion != BodyStrengthSchemaVersion ||
		bs.ScoringProfile != bodyStrengthScoringProfile || bs.ScoreBandCandidate == "" ||
		bs.YueLingRuleID != yueLingRuleID || bs.YueLingProfile != yueLingProfile ||
		bs.YueLingTableSHA256 != yueLingTableSHA256 ||
		bs.ValidationStatus != "not_validated" || bs.InterpretationStatus != "not_adjudicated" ||
		bs.IsStrengthConclusion {
		t.Fatalf("body-strength evidence contract is incomplete: %+v", bs)
	}

	wantKeys := map[string]bool{"ling": false, "di": false, "shi": false, "sheng": false, "bonus": false}
	foundYueLingComponent := false
	for _, c := range bs.Components {
		if _, ok := wantKeys[c.Key]; ok {
			wantKeys[c.Key] = true
		}
		if c.Weight <= 0 {
			t.Fatalf("component %s has non-positive weight %.2f", c.Key, c.Weight)
		}
		if c.Key == "ling" && c.RuleID == yueLingRuleID {
			foundYueLingComponent = true
		}
	}
	if !foundYueLingComponent || bs.Evidence[0].RuleID != yueLingRuleID+".lookup" {
		t.Fatalf("yue-ling evidence is not bound to its stable rule ID: %+v", bs)
	}
	for key, found := range wantKeys {
		if !found {
			t.Fatalf("missing body strength component %q", key)
		}
	}

	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	if !ValidBodyStrengthEvidence(bs, pillars) {
		t.Fatalf("generated body-strength evidence did not pass strict validation: %+v", bs)
	}
	payload, err := json.Marshal(bs)
	if err != nil {
		t.Fatal(err)
	}
	var topLevel map[string]json.RawMessage
	if err := json.Unmarshal(payload, &topLevel); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"verdict", "like", "dislike", "summary"} {
		if _, ok := topLevel[forbidden]; ok || strings.Contains(string(payload), `"`+forbidden+`"`) {
			t.Fatalf("body-strength response leaked legacy field %q: %s", forbidden, payload)
		}
	}
	var decoded BodyStrengthResult
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if !ValidBodyStrengthEvidence(decoded, pillars) {
		t.Fatalf("serialized body-strength evidence did not pass strict validation: %+v", decoded)
	}
	decoded.TotalScore += 0.01
	if ValidBodyStrengthEvidence(decoded, pillars) {
		t.Fatal("tampered body-strength score must not pass strict validation")
	}
}
