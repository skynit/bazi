package bazi

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestPatternDetectorSemanticProfilesAreComplete(t *testing.T) {
	expected := map[string]string{
		"pattern.special.zhuanwang": "daedf9a821d62349eadcaf07757cc678f86ff67f3adb76e966d3d8d3685cd57c",
		"pattern.lu.jianlu":         "29314b503c9d6b5dcd248121d9abb189ae5304e2984648299be0580bb04b5322",
		"pattern.lu.yueren":         "80c34b3028f8a9cb8b8dab004ee89fc9355ed2b8feb5b451c1113fe41d58f97e",
		"pattern.lu.zhuanlu":        "94f7305566788f872cbcceadd7aa62e15ab41d42421869026c509798ae256b2b",
		"pattern.lu.riren":          "bb3a422fa07b0661abac8c7925aaa039262f4a1a6e8cde0b1c6d5d20aa847983",
		"pattern.special.liangqi":   "20ed815f5411b2add63e4d98931f8f775b7af1b4f98a2b41f91a6b9feae9c610",
		"pattern.aux.kuigang":       "5473b6110b3cb1bfbc2570f000c6e84da7d1ad5a6738ef1f1ca462d4c40d8e02",
		"pattern.aux.jinshen":       "c09c584d0ae7e0b55e6167cae417f6f85c9f1a514978ff0640d9dde9c128f63e",
		"pattern.aux.sanqi":         "b54db26c03be56842109528860f258280937d5dafbaf5b5c5dd61cfbd6956028",
		"pattern.aux.ride":          "d6624103cfe639446158c5dfa6909f4979dc1b995a1da1b245acf1621431ab02",
	}
	seen := make(map[string]string, patternDetectorCount())
	for _, detector := range patternDetectorRegistry() {
		profile, ok := patternDetectorSemanticProfile(detector.ruleID)
		if !ok || profile == nil {
			t.Errorf("detector %s has no semantic Profile", detector.ruleID)
			continue
		}
		if len(detector.profileSHA256) != 64 || detector.profileSHA256 != patternDetectorProfileSHA256(detector.ruleID) {
			t.Errorf("detector %s profile SHA-256 = %q", detector.ruleID, detector.profileSHA256)
		}
		if detector.profileSHA256 != expected[detector.ruleID] {
			t.Errorf("detector %s profile SHA-256 = %s, want %s", detector.ruleID, detector.profileSHA256, expected[detector.ruleID])
		}
		if formerRuleID, duplicate := seen[detector.profileSHA256]; duplicate {
			t.Errorf("detectors %s and %s share profile SHA-256 %s", formerRuleID, detector.ruleID, detector.profileSHA256)
		}
		seen[detector.profileSHA256] = detector.ruleID
	}
	if len(seen) != patternDetectorCount() {
		t.Fatalf("semantic Profile hash count = %d, want %d", len(seen), patternDetectorCount())
	}
}

func TestPatternDetectorClosedProfilesReturnIndependentValues(t *testing.T) {
	yangRen := yangRenProfile()
	kuiGang := kuiGangDayProfile()
	riDe := riDeDayProfile()
	jinShen := jinShenHourProfile()
	sanQi := sanQiSequenceProfile()
	jinShenSemantic := jinShenSemanticProfile()
	sanQiSemantic := sanQiSemanticProfile()
	liangQi := liangQiSemanticProfile()
	zhuanWang := zhuanWangDetectorSemanticProfile()
	stemElements := patternStemElementProfile()
	branchElements := patternBranchElementProfile()
	pillarContext := patternPillarContextSemanticProfile()
	algorithm, ok := patternDetectorAlgorithmProfileForRule("pattern.special.zhuanwang")
	if !ok {
		t.Fatal("zhuan-wang algorithm Profile missing")
	}
	yangRen[0] = patternStemBranchTarget{Stem: "变", Branch: "更"}
	kuiGang[0], riDe[0], jinShen[0], sanQi[0] = "变", "更", "改", "写"
	liangQi.ElementOrder[0] = "变"
	liangQi.StemElements[0] = patternSymbolElementTarget{Symbol: "变", Element: "更"}
	zhuanWang.BranchElements[0] = patternSymbolElementTarget{Symbol: "改", Element: "写"}
	jinShenSemantic.Pillars[0] = "变更"
	sanQiSemantic.StemWindows[0], sanQiSemantic.WindowStarts[0], sanQiSemantic.Positions[0] = "改写", 9, "变更"
	stemElements[0], branchElements[0] = patternSymbolElementTarget{}, patternSymbolElementTarget{}
	pillarContext.DayPillarIndex = 0
	algorithm.Functions[0] = "changed"

	if fresh := yangRenProfile(); fresh[0].Stem != "甲" || fresh[0].Branch != "卯" {
		t.Fatalf("fresh yang-ren Profile inherited mutation: %+v", fresh)
	}
	if fresh := kuiGangDayProfile(); fresh[0] != "庚辰" {
		t.Fatalf("fresh kui-gang Profile inherited mutation: %v", fresh)
	}
	if fresh := riDeDayProfile(); fresh[0] != "甲寅" {
		t.Fatalf("fresh ri-de Profile inherited mutation: %v", fresh)
	}
	if fresh := jinShenHourProfile(); fresh[0] != "癸酉" {
		t.Fatalf("fresh jin-shen Profile inherited mutation: %v", fresh)
	}
	if fresh := sanQiSequenceProfile(); fresh[0] != "乙丙丁" {
		t.Fatalf("fresh san-qi Profile inherited mutation: %v", fresh)
	}
	if fresh := jinShenSemanticProfile(); fresh.PillarCount != 4 || fresh.PillarIndex != 3 || fresh.Pillars[0] != "癸酉" {
		t.Fatalf("fresh jin-shen semantic Profile inherited mutation: %+v", fresh)
	}
	if fresh := sanQiSemanticProfile(); fresh.PillarCount != 4 || fresh.WindowSize != 3 ||
		fresh.WindowStarts[0] != 0 || fresh.StemWindows[0] != "乙丙丁" || fresh.Positions[0] != "year_month_day" {
		t.Fatalf("fresh san-qi semantic Profile inherited mutation: %+v", fresh)
	}
	if fresh := liangQiSemanticProfile(); fresh.ElementOrder[0] != "木" || fresh.PillarCount != 4 ||
		fresh.DistinctElements != 2 || fresh.OccurrencesPerElement != 4 || fresh.StemElements[0].Symbol != "甲" {
		t.Fatalf("fresh liang-qi Profile inherited mutation: %+v", fresh)
	}
	if fresh := zhuanWangDetectorSemanticProfile(); fresh.PillarCount != 4 || fresh.BranchElements[0].Symbol != "子" {
		t.Fatalf("fresh zhuan-wang detector Profile inherited mutation: %+v", fresh)
	}
	if fresh := patternStemElementProfile(); fresh[0] != (patternSymbolElementTarget{Symbol: "甲", Element: "木"}) {
		t.Fatalf("fresh stem-element Profile inherited mutation: %+v", fresh)
	}
	if fresh := patternBranchElementProfile(); fresh[0] != (patternSymbolElementTarget{Symbol: "子", Element: "水"}) {
		t.Fatalf("fresh branch-element Profile inherited mutation: %+v", fresh)
	}
	if fresh := patternPillarContextSemanticProfile(); fresh.DayPillarIndex != 2 || fresh.MonthPillarIndex != 1 {
		t.Fatalf("fresh pillar-context Profile inherited mutation: %+v", fresh)
	}
	if fresh, ok := patternDetectorAlgorithmProfileForRule("pattern.special.zhuanwang"); !ok || fresh.Functions[0] != "checkZhuanWangGe" {
		t.Fatalf("fresh algorithm Profile inherited mutation: %+v/%v", fresh, ok)
	}
}

func TestPatternPillarContextProfileIsCompleteAndFailClosed(t *testing.T) {
	want := patternPillarContextProfile{
		PillarCount:               4,
		YearPillarIndex:           0,
		MonthPillarIndex:          1,
		DayPillarIndex:            2,
		HourPillarIndex:           3,
		DeclaredMonthBranchPolicy: "must_equal_month_pillar_branch",
	}
	if got := patternPillarContextSemanticProfile(); !reflect.DeepEqual(got, want) || !validPatternPillarContextProfile(got) {
		t.Fatalf("pillar-context Profile = %+v, want valid %+v", got, want)
	}
	mutations := []func(*patternPillarContextProfile){
		func(profile *patternPillarContextProfile) { profile.PillarCount = 5 },
		func(profile *patternPillarContextProfile) { profile.YearPillarIndex = -1 },
		func(profile *patternPillarContextProfile) { profile.MonthPillarIndex = 4 },
		func(profile *patternPillarContextProfile) { profile.DayPillarIndex = profile.MonthPillarIndex },
		func(profile *patternPillarContextProfile) { profile.HourPillarIndex = profile.DayPillarIndex },
		func(profile *patternPillarContextProfile) { profile.DeclaredMonthBranchPolicy = "changed" },
	}
	for index, mutate := range mutations {
		profile := want
		mutate(&profile)
		if validPatternPillarContextProfile(profile) {
			t.Errorf("invalid pillar-context mutation %d passed: %+v", index, profile)
		}
	}
}

func TestEveryDetectorSemanticHashBindsSharedPillarContext(t *testing.T) {
	want := patternPillarContextSemanticProfile()
	for _, detector := range patternDetectorRegistry() {
		profile, ok := patternDetectorSemanticProfile(detector.ruleID)
		if !ok {
			t.Fatalf("detector %s has no semantic Profile", detector.ruleID)
		}
		envelope, ok := profile.(patternDetectorSemanticEnvelope)
		algorithm, algorithmOK := patternDetectorAlgorithmProfileForRule(detector.ruleID)
		if !ok || !algorithmOK || !reflect.DeepEqual(envelope.PillarContext, want) ||
			!reflect.DeepEqual(envelope.OutputNames, detector.outputNames) || envelope.Detector == nil {
			t.Errorf("detector %s semantic envelope = %#v", detector.ruleID, profile)
		}
		if !reflect.DeepEqual(envelope.Algorithm, algorithm) {
			t.Errorf("detector %s algorithm envelope = %#v, want %#v", detector.ruleID, envelope.Algorithm, algorithm)
		}
	}
}

func TestPatternDetectorOutputNameProfilesAreCompleteAndIndependent(t *testing.T) {
	expected := map[string][]string{
		"pattern.special.zhuanwang": {"曲直格", "炎上格", "稼穑格", "从革格", "润下格"},
		"pattern.lu.jianlu":         {"建禄格"},
		"pattern.lu.yueren":         {"月刃格"},
		"pattern.lu.zhuanlu":        {"专禄格"},
		"pattern.lu.riren":          {"日刃格"},
		"pattern.special.liangqi":   {"两气成象格"},
		"pattern.aux.kuigang":       {"魁罡格"},
		"pattern.aux.jinshen":       {"金神"},
		"pattern.aux.sanqi":         {"三奇"},
		"pattern.aux.ride":          {"日德格"},
	}
	for ruleID, want := range expected {
		got := patternDetectorOutputNames(ruleID)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("detector %s output names = %v, want %v", ruleID, got, want)
			continue
		}
		got[0] = "变更"
		if fresh := patternDetectorOutputNames(ruleID); !reflect.DeepEqual(fresh, want) {
			t.Errorf("detector %s output names inherited mutation: %v", ruleID, fresh)
		}
	}
	if names := patternDetectorOutputNames("pattern.unknown"); names != nil {
		t.Errorf("unknown detector output names = %v, want nil", names)
	}
}

func TestPatternElementProfilesCoverCanonicalStemsAndBranches(t *testing.T) {
	wantStems := []patternSymbolElementTarget{
		{Symbol: "甲", Element: "木"}, {Symbol: "乙", Element: "木"},
		{Symbol: "丙", Element: "火"}, {Symbol: "丁", Element: "火"},
		{Symbol: "戊", Element: "土"}, {Symbol: "己", Element: "土"},
		{Symbol: "庚", Element: "金"}, {Symbol: "辛", Element: "金"},
		{Symbol: "壬", Element: "水"}, {Symbol: "癸", Element: "水"},
	}
	wantBranches := []patternSymbolElementTarget{
		{Symbol: "子", Element: "水"}, {Symbol: "丑", Element: "土"},
		{Symbol: "寅", Element: "木"}, {Symbol: "卯", Element: "木"},
		{Symbol: "辰", Element: "土"}, {Symbol: "巳", Element: "火"},
		{Symbol: "午", Element: "火"}, {Symbol: "未", Element: "土"},
		{Symbol: "申", Element: "金"}, {Symbol: "酉", Element: "金"},
		{Symbol: "戌", Element: "土"}, {Symbol: "亥", Element: "水"},
	}
	if got := patternStemElementProfile(); !reflect.DeepEqual(got, wantStems) {
		t.Fatalf("stem-element Profile = %+v, want %+v", got, wantStems)
	}
	if got := patternBranchElementProfile(); !reflect.DeepEqual(got, wantBranches) {
		t.Fatalf("branch-element Profile = %+v, want %+v", got, wantBranches)
	}
	for _, unknown := range []string{"", "A", "甲甲", "木"} {
		if element, ok := patternElementForSymbol(wantStems, unknown); ok || element != "" {
			t.Errorf("unknown symbol %q resolved to %q/%v", unknown, element, ok)
		}
	}
}

func TestPatternDetectorsIgnoreMutableDataElementMaps(t *testing.T) {
	oldGan, hadGan := data.GanElement["甲"]
	oldBranch, hadBranch := data.ZhiElement["卯"]
	defer func() {
		if hadGan {
			data.GanElement["甲"] = oldGan
		} else {
			delete(data.GanElement, "甲")
		}
		if hadBranch {
			data.ZhiElement["卯"] = oldBranch
		} else {
			delete(data.ZhiElement, "卯")
		}
	}()
	data.GanElement["甲"] = "金"
	data.ZhiElement["卯"] = "金"

	zhuanWangPillars := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
	}
	if got := checkZhuanWangGe(zhuanWangPillars); got == nil || got.PatternName != "曲直格" {
		t.Fatalf("mutable data element maps changed zhuan-wang detection: %+v", got)
	}
	liangQiPillars := []model.Pillar{
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
	}
	if got := checkLiangQiChengXiang(liangQiPillars); got == nil || got.PatternName != "两气成象格" {
		t.Fatalf("mutable data element maps changed liang-qi detection: %+v", got)
	}
}

func TestPatternDetectorsConsumeSharedSemanticProfiles(t *testing.T) {
	checks := map[string][]string{
		"pattern.go": {
			"map[string]bool{\"庚辰\"", "map[string]bool{\"甲寅\"",
			"map[string]string{\"甲\": \"卯\"",
			"for _, element := range []string{\"木\", \"火\", \"土\", \"金\", \"水\"}",
			"data.GanElement", "data.ZhiElement",
		},
		"shensha.go": {
			"seq == \"乙丙丁\"", "seq == \"甲戊庚\"", "case \"癸酉\", \"己巳\", \"乙丑\"",
			"i+2 < len(gans)", "gans[i:i+3]",
		},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("%s retains detector-local semantic Profile %q", path, forbidden)
			}
		}
	}

	patternSource, err := os.ReadFile("pattern.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"semanticProfile := zhuanWangDetectorSemanticProfile()",
		"contextProfile := patternPillarContextSemanticProfile()",
		"len(pillars) != semanticProfile.PillarCount",
		"pillars[contextProfile.DayPillarIndex]",
		"patternElementForSymbol(semanticProfile.StemElements",
		"patternElementForSymbol(semanticProfile.BranchElements",
		"profile := liangQiSemanticProfile()",
		"len(pillars) != profile.PillarCount",
		"len(counts) != profile.DistinctElements",
		"range profile.ElementOrder",
		"count != profile.OccurrencesPerElement",
		"patternElementForSymbol(profile.StemElements",
		"patternElementForSymbol(profile.BranchElements",
	} {
		if !strings.Contains(string(patternSource), required) {
			t.Errorf("pattern.go does not consume liang-qi Profile field %q", required)
		}
	}
	for path, requiredValues := range map[string][]string{
		"pattern.go": {
			"jinShenProfile := jinShenSemanticProfile()",
			"len(pillars) != jinShenProfile.PillarCount",
			"pillars[jinShenProfile.PillarIndex]",
		},
		"shensha.go": {
			"profile := sanQiSemanticProfile()",
			"len(gans) != profile.PillarCount",
			"range profile.WindowStarts",
			"start + profile.WindowSize",
			"profile.StemWindows",
			"jinShenSemanticProfile().Pillars",
		},
	} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredValues {
			if !strings.Contains(string(source), required) {
				t.Errorf("%s does not consume auxiliary semantic Profile field %q", path, required)
			}
		}
	}
	for path, forbiddenValues := range map[string][]string{
		"pattern_candidates.go": {"pillars[1]", "pillars[2]"},
		"pattern.go":            {"pillars[2]", `PatternName: "`},
	} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("%s retains detector-context pillar hardcode %q", path, forbidden)
			}
		}
	}
	for _, required := range []string{
		"contextProfile := patternPillarContextSemanticProfile()",
		"validPatternPillarContextProfile(contextProfile)",
		"pillars[contextProfile.MonthPillarIndex].Zhi",
		"pillars[contextProfile.DayPillarIndex].Gan",
		"pillars[contextProfile.DayPillarIndex].Zhi",
	} {
		source, err := os.ReadFile("pattern_candidates.go")
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(source), required) {
			t.Errorf("pattern_candidates.go does not consume pillar-context Profile field %q", required)
		}
	}
}

func TestPatternDetectorSemanticProfileMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if count := strings.Count(table.Description, "旧三奇语义Profile记录窗口大小和两个位置"); count != 1 {
			t.Errorf("san-qi/jin-shen metadata statement count = %d, want 1", count)
		}
		if count := strings.Count(table.Description, "旧10条逐规则摘要未绑定共享四柱位置上下文"); count != 1 {
			t.Errorf("pillar-context metadata statement count = %d, want 1", count)
		}
		if count := strings.Count(table.Description, "旧多数检测器把PatternName作为局部字面量返回"); count != 1 {
			t.Errorf("output-name metadata statement count = %d, want 1", count)
		}
		for _, fragment := range []string{
			"旧清单摘要只散列implementation版本名",
			"封闭表、柱位范围和算法参数",
			"pattern-candidate-set-v17集中10个运行时语义Profile",
			"每条注册项计算profile_sha256",
			"实现共同消费这些纯Profile",
			"ee21f7c8438031bd64f284f9691d16934d53930dacfa3661bafc5874e3fe4a8f",
			"逐规则摘要由合同分别锁定",
			"旧两气语义Profile记录了五行顺序和计数参数",
			"liangQiSemanticProfile独立快照",
			"49d0edcdc94b96ef1d351b44d653541ad964942aefdef33b9b37dde8ee254c07",
			"3b89e62e6fe12baf7969be4a2afb35b75308b0029d55ee16c5a5d273b2c49636",
			"命中语义、古籍来源和证据等级不变",
			"旧专旺与两气检测器直接读取data.GanElement和data.ZhiElement",
			"十干10项、十二支12项纯值Profile",
			"专旺与两气的运行时和语义摘要共同消费映射及四柱数",
			"d4e2a3250ea362c239982cb9c5ea6ccc62b69ba0e0f3198a551e277b6b0e8073",
			"03514ea7676c03bcc79bff1616f73552cbc07cfa7fdd5121eebd8f1e26db4543",
			"32ffa7b00d2145bfe737b8bf0a9135f4feb10d3fc03963ca83dfc00f05cad938",
			"旧三奇语义Profile记录窗口大小和两个位置",
			"sanQiSemanticProfile和jinShenSemanticProfile",
			"共同消费四柱数、窗口大小、窗口起点、顺序表、时柱索引和三时表",
			"79627a02c955fc510bfe283954991a51a83b7e04653f11aea6304af10941125b",
			"004184a1c8e70f481240225d637eba2d1ce6d0ccb09f9db1e05a1199f186909a",
			"881193d5290ed58d831dfdde6b57b626380b193d465c44a2d545d59b93b023f7",
			"旧10条逐规则摘要未绑定共享四柱位置上下文",
			"patternPillarContextSemanticProfile",
			"年/月/日/时索引0/1/2/3",
			"同一上下文封装进全部10条逐规则摘要",
			"ec2928c92f0b6e227e9f656a788fa641215a3587a0309876d7f84cee6aa08ab6",
			"旧多数检测器把PatternName作为局部字面量返回",
			"patternDetectorOutputNames纯Profile",
			"检测器构造、注册器允许集合、逐规则envelope和总清单共同消费",
			"acd631f529e51ead2c50fa1c7832149ad7d994d7137be348133ecd70de2cff1a",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
