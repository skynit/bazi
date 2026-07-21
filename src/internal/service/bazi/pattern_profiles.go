package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

type patternPillarContextProfile struct {
	PillarCount               int    `json:"pillar_count"`
	YearPillarIndex           int    `json:"year_pillar_index"`
	MonthPillarIndex          int    `json:"month_pillar_index"`
	DayPillarIndex            int    `json:"day_pillar_index"`
	HourPillarIndex           int    `json:"hour_pillar_index"`
	DeclaredMonthBranchPolicy string `json:"declared_month_branch_policy"`
}

func patternPillarContextSemanticProfile() patternPillarContextProfile {
	return patternPillarContextProfile{
		PillarCount:               4,
		YearPillarIndex:           0,
		MonthPillarIndex:          1,
		DayPillarIndex:            2,
		HourPillarIndex:           3,
		DeclaredMonthBranchPolicy: "must_equal_month_pillar_branch",
	}
}

type patternDetectorSemanticEnvelope struct {
	PillarContext       patternPillarContextProfile        `json:"pillar_context"`
	OutputNames         []string                           `json:"output_names"`
	Algorithm           patternDetectorAlgorithmProfile    `json:"algorithm"`
	BehaviorWitnesses   []patternDetectorBehaviorWitness   `json:"behavior_witnesses,omitempty"`
	MetamorphicPolicies []patternDetectorMetamorphicPolicy `json:"metamorphic_policies,omitempty"`
	BehaviorManifest    *patternDetectorBehaviorManifest   `json:"behavior_manifest,omitempty"`
	Detector            any                                `json:"detector"`
}

type patternDetectorBehaviorManifest struct {
	Scheme     string `json:"scheme"`
	Domain     string `json:"domain"`
	CaseCount  int    `json:"case_count"`
	MatchCount int    `json:"match_count"`
	SHA256     string `json:"sha256"`
}

func patternDetectorBehaviorManifestProfile(ruleID string) *patternDetectorBehaviorManifest {
	profiles := map[string]patternDetectorBehaviorManifest{
		"pattern.lu.jianlu":   {Domain: "ten_stems_x_twelve_month_branches", CaseCount: 120, MatchCount: 10, SHA256: "d51707e2f5627ade0d053f10c9885eb7406e0b55defb6ea3b7ed6ad25c91a97b"},
		"pattern.lu.yueren":   {Domain: "ten_stems_x_twelve_month_branches", CaseCount: 120, MatchCount: 5, SHA256: "cb4d3b6db008d758dbefa4730bf876c8ec6edf82a116733ed3bcd8131b0a6df3"},
		"pattern.lu.zhuanlu":  {Domain: "sixty_day_pillars", CaseCount: 60, MatchCount: 4, SHA256: "3e5fa70c2fb6805ecb5e4c0d57318242840866b2ee3b3bb7057fd6521deb4581"},
		"pattern.lu.riren":    {Domain: "sixty_day_pillars", CaseCount: 60, MatchCount: 3, SHA256: "d3258cb7609e23df271a827467f048799339f31623a47949de1c55e1930cbc1b"},
		"pattern.aux.kuigang": {Domain: "sixty_day_pillars", CaseCount: 60, MatchCount: 4, SHA256: "40f31c0dd264242bb03b8e13a1befab913faeddb75e40a4a11154e5c72f8bfe9"},
		"pattern.aux.ride":    {Domain: "sixty_day_pillars", CaseCount: 60, MatchCount: 5, SHA256: "7be77b6771f51945042d3fe7ddb0a305c8840147f89c97b84754d2dc67c65a1c"},
		"pattern.aux.jinshen": {Domain: "sixty_hour_pillars_in_four_pillar_context", CaseCount: 60, MatchCount: 3, SHA256: "be71d3d52d03a913a3e2f8010c92480c2da401397c68d82fcdd4492881b2fe09"},
		"pattern.aux.sanqi":   {Domain: "ten_stem_four_pillar_sequences", CaseCount: 10000, MatchCount: 40, SHA256: "7a9e9fb9f8a63f9bfd908c4b27187c8329e132cb63d410d371adc9b51de2348a"},
	}
	profile, ok := profiles[ruleID]
	if !ok {
		return nil
	}
	profile.Scheme = "canonical_truth_table_v1"
	return &profile
}

func patternDetectorBehaviorSHA256(ruleID string) string {
	if manifest := patternDetectorBehaviorManifestProfile(ruleID); manifest != nil {
		return manifest.SHA256
	}
	witnesses := patternDetectorBehaviorWitnessProfile(ruleID)
	policies := patternDetectorMetamorphicProfile(ruleID)
	if len(witnesses) == 0 && len(policies) == 0 {
		return ""
	}
	payload, err := json.Marshal(struct {
		Scheme    string                             `json:"scheme"`
		Witnesses []patternDetectorBehaviorWitness   `json:"witnesses"`
		Policies  []patternDetectorMetamorphicPolicy `json:"policies"`
	}{
		Scheme:    "behavior_contract_v1",
		Witnesses: witnesses,
		Policies:  policies,
	})
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type patternDetectorMetamorphicPolicy struct {
	ID        string `json:"id"`
	Transform string `json:"transform"`
	Relation  string `json:"relation"`
}

func patternDetectorMetamorphicProfile(ruleID string) []patternDetectorMetamorphicPolicy {
	switch ruleID {
	case "pattern.special.zhuanwang":
		return []patternDetectorMetamorphicPolicy{
			{ID: "zhuanwang.meta.branch_permutation", Transform: "permute_four_branch_positions_keep_day_stem_element", Relation: "preserve_output"},
			{ID: "zhuanwang.meta.repeat_required_branch", Transform: "repeat_any_required_branch_after_all_required_present", Relation: "preserve_output"},
			{ID: "zhuanwang.meta.external_nonbreaking_branch", Transform: "append_branch_outside_structure_with_nonbreaking_main_element", Relation: "preserve_output"},
			{ID: "zhuanwang.meta.external_breaking_branch", Transform: "append_branch_outside_structure_with_breaking_main_element", Relation: "reject"},
			{ID: "zhuanwang.meta.missing_required_branch", Transform: "remove_one_required_branch_and_duplicate_remaining_branches", Relation: "reject"},
		}
	case "pattern.special.liangqi":
		return []patternDetectorMetamorphicPolicy{
			{ID: "liangqi.meta.unordered_element_pair", Transform: "choose_each_unordered_pair_from_five_elements", Relation: "preserve_output"},
			{ID: "liangqi.meta.four_of_eight_positions", Transform: "assign_first_element_to_any_four_of_eight_stem_branch_positions", Relation: "preserve_output"},
			{ID: "liangqi.meta.introduce_third_element", Transform: "replace_one_first_element_position_with_element_outside_pair", Relation: "reject"},
			{ID: "liangqi.meta.unequal_five_three_split", Transform: "replace_one_first_element_position_with_second_element", Relation: "reject"},
		}
	default:
		return nil
	}
}

type patternWitnessPillar struct {
	Stem   string `json:"stem"`
	Branch string `json:"branch"`
}

type patternDetectorBehaviorWitness struct {
	ID                 string                 `json:"id"`
	BaselineID         string                 `json:"baseline_id,omitempty"`
	Mutation           string                 `json:"mutation,omitempty"`
	Pillars            []patternWitnessPillar `json:"pillars"`
	ExpectedOutputName string                 `json:"expected_output_name,omitempty"`
}

func patternDetectorBehaviorWitnessProfile(ruleID string) []patternDetectorBehaviorWitness {
	var witnesses []patternDetectorBehaviorWitness
	switch ruleID {
	case "pattern.special.zhuanwang":
		witnesses = []patternDetectorBehaviorWitness{
			{ID: "zhuanwang.east_direction", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "辰"}, {"丁", "未"}}, ExpectedOutputName: "曲直格"},
			{ID: "zhuanwang.missing_structure_branch", BaselineID: "zhuanwang.east_direction", Mutation: "day_branch 辰->午", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "未"}}},
			{ID: "zhuanwang.exposed_breaking_stem", BaselineID: "zhuanwang.east_direction", Mutation: "hour_stem 丁->辛", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "辰"}, {"辛", "未"}}},
			{ID: "zhuanwang.external_breaking_branch", BaselineID: "zhuanwang.east_direction", Mutation: "hour_branch 未->酉", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "辰"}, {"丁", "酉"}}},
			{ID: "zhuanwang.unknown_day_stem", BaselineID: "zhuanwang.east_direction", Mutation: "day_stem 甲->?", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"?", "辰"}, {"丁", "未"}}},
			{ID: "zhuanwang.unknown_other_stem", BaselineID: "zhuanwang.east_direction", Mutation: "year_stem 甲->?", Pillars: []patternWitnessPillar{{"?", "寅"}, {"丁", "卯"}, {"甲", "辰"}, {"丁", "未"}}},
			{ID: "zhuanwang.unknown_external_branch", BaselineID: "zhuanwang.east_direction", Mutation: "hour_branch 未->?", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "辰"}, {"丁", "?"}}},
			{ID: "zhuanwang.incomplete_pillars", BaselineID: "zhuanwang.east_direction", Mutation: "remove_hour_pillar", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "辰"}}},
			{ID: "zhuanwang.fire_triad", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"甲", "午"}, {"丙", "戌"}, {"乙", "未"}}, ExpectedOutputName: "炎上格"},
			{ID: "zhuanwang.fire_missing_structure", BaselineID: "zhuanwang.fire_triad", Mutation: "day_branch 戌->辰", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"甲", "午"}, {"丙", "辰"}, {"乙", "未"}}},
			{ID: "zhuanwang.fire_breaking_stem", BaselineID: "zhuanwang.fire_triad", Mutation: "hour_stem 乙->癸", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"甲", "午"}, {"丙", "戌"}, {"癸", "未"}}},
			{ID: "zhuanwang.fire_breaking_branch", BaselineID: "zhuanwang.fire_triad", Mutation: "hour_branch 未->亥", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"甲", "午"}, {"丙", "戌"}, {"乙", "亥"}}},
			{ID: "zhuanwang.four_storehouses", Pillars: []patternWitnessPillar{{"戊", "辰"}, {"己", "丑"}, {"戊", "戌"}, {"己", "未"}}, ExpectedOutputName: "稼穑格"},
			{ID: "zhuanwang.earth_missing_structure", BaselineID: "zhuanwang.four_storehouses", Mutation: "hour_branch 未->巳", Pillars: []patternWitnessPillar{{"戊", "辰"}, {"己", "丑"}, {"戊", "戌"}, {"己", "巳"}}},
			{ID: "zhuanwang.earth_breaking_stem", BaselineID: "zhuanwang.four_storehouses", Mutation: "hour_stem 己->乙", Pillars: []patternWitnessPillar{{"戊", "辰"}, {"己", "丑"}, {"戊", "戌"}, {"乙", "未"}}},
			{ID: "zhuanwang.metal_direction", Pillars: []patternWitnessPillar{{"庚", "申"}, {"辛", "酉"}, {"庚", "戌"}, {"戊", "辰"}}, ExpectedOutputName: "从革格"},
			{ID: "zhuanwang.metal_missing_structure", BaselineID: "zhuanwang.metal_direction", Mutation: "day_branch 戌->子", Pillars: []patternWitnessPillar{{"庚", "申"}, {"辛", "酉"}, {"庚", "子"}, {"戊", "辰"}}},
			{ID: "zhuanwang.metal_breaking_stem", BaselineID: "zhuanwang.metal_direction", Mutation: "hour_stem 戊->丙", Pillars: []patternWitnessPillar{{"庚", "申"}, {"辛", "酉"}, {"庚", "戌"}, {"丙", "辰"}}},
			{ID: "zhuanwang.metal_breaking_branch", BaselineID: "zhuanwang.metal_direction", Mutation: "hour_branch 辰->午", Pillars: []patternWitnessPillar{{"庚", "申"}, {"辛", "酉"}, {"庚", "戌"}, {"戊", "午"}}},
			{ID: "zhuanwang.water_direction", Pillars: []patternWitnessPillar{{"癸", "亥"}, {"壬", "子"}, {"癸", "丑"}, {"庚", "申"}}, ExpectedOutputName: "润下格"},
			{ID: "zhuanwang.water_missing_structure", BaselineID: "zhuanwang.water_direction", Mutation: "day_branch 丑->酉", Pillars: []patternWitnessPillar{{"癸", "亥"}, {"壬", "子"}, {"癸", "酉"}, {"庚", "申"}}},
			{ID: "zhuanwang.water_breaking_stem", BaselineID: "zhuanwang.water_direction", Mutation: "hour_stem 庚->戊", Pillars: []patternWitnessPillar{{"癸", "亥"}, {"壬", "子"}, {"癸", "丑"}, {"戊", "申"}}},
			{ID: "zhuanwang.water_breaking_branch", BaselineID: "zhuanwang.water_direction", Mutation: "hour_branch 申->辰", Pillars: []patternWitnessPillar{{"癸", "亥"}, {"壬", "子"}, {"癸", "丑"}, {"庚", "辰"}}},
		}
	case "pattern.special.liangqi":
		witnesses = []patternDetectorBehaviorWitness{
			{ID: "liangqi.wood_fire_balanced", Pillars: []patternWitnessPillar{{"甲", "午"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "卯"}}, ExpectedOutputName: "两气成象格"},
			{ID: "liangqi.third_element", BaselineID: "liangqi.wood_fire_balanced", Mutation: "hour_branch 卯->丑", Pillars: []patternWitnessPillar{{"甲", "午"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "丑"}}},
			{ID: "liangqi.unequal_split", BaselineID: "liangqi.wood_fire_balanced", Mutation: "year_branch 午->寅", Pillars: []patternWitnessPillar{{"甲", "寅"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "卯"}}},
			{ID: "liangqi.unknown_stem", BaselineID: "liangqi.wood_fire_balanced", Mutation: "year_stem 甲->?", Pillars: []patternWitnessPillar{{"?", "午"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "卯"}}},
			{ID: "liangqi.unknown_branch", BaselineID: "liangqi.wood_fire_balanced", Mutation: "year_branch 午->?", Pillars: []patternWitnessPillar{{"甲", "?"}, {"丁", "卯"}, {"甲", "午"}, {"丁", "卯"}}},
			{ID: "liangqi.incomplete_pillars", BaselineID: "liangqi.wood_fire_balanced", Mutation: "remove_hour_pillar", Pillars: []patternWitnessPillar{{"甲", "午"}, {"丁", "卯"}, {"甲", "午"}}},
		}
	default:
		return nil
	}
	for index := range witnesses {
		witnesses[index].Pillars = append([]patternWitnessPillar(nil), witnesses[index].Pillars...)
	}
	return witnesses
}

type patternDetectorAlgorithmProfile struct {
	Scheme       string   `json:"scheme"`
	RootFunction string   `json:"root_function"`
	Functions    []string `json:"functions"`
	ASTSHA256    string   `json:"ast_sha256"`
}

func patternDetectorAlgorithmProfileForRule(ruleID string) (patternDetectorAlgorithmProfile, bool) {
	profiles := map[string]patternDetectorAlgorithmProfile{
		"pattern.special.zhuanwang": {
			RootFunction: "checkZhuanWangGe",
			Functions:    []string{"checkZhuanWangGe", "containsAllBranches", "inStrings", "patternBranchElementProfile", "patternElementForSymbol", "patternPillarContextSemanticProfile", "patternStemElementProfile", "zhuanWangDetectorSemanticProfile", "zhuanWangProfileForElement", "zhuanWangProfileRegistry", "zhuanWangSemanticProfile"},
			ASTSHA256:    "ef1fe91027ba0e0101789f43e2bef18f3f17fd273a4913ca88d980deee66e53f",
		},
		"pattern.lu.jianlu": {
			RootFunction: "checkJianLuGe", Functions: []string{"checkJianLuGe", "luBranchForStem"},
			ASTSHA256: "a8ba8785978eecf10745ffbb81d86542f3a948ded86d569ca8ffc8ad565c603e",
		},
		"pattern.lu.yueren": {
			RootFunction: "checkYueRenGe", Functions: []string{"checkYueRenGe", "patternBranchForStem", "yangRenProfile", "yangRenZhi"},
			ASTSHA256: "3955ef0ad0405984cbc41bb2a58571f9409b1a75a51ac00be88f1278ffa6fd28",
		},
		"pattern.lu.zhuanlu": {
			RootFunction: "checkZhuanLuGe", Functions: []string{"checkZhuanLuGe", "luBranchForStem"},
			ASTSHA256: "4ac683a19f228d45b2acd2bdbe22b0c0ecd22d161f824295f6d1ca712a302067",
		},
		"pattern.lu.riren": {
			RootFunction: "checkRiRenGe", Functions: []string{"checkRiRenGe", "patternBranchForStem", "yangRenProfile", "yangRenZhi"},
			ASTSHA256: "0f124e093fbc6009e6fadb279ae9020205493dd5e4b2e2d07a4fc8bfa5a8799d",
		},
		"pattern.special.liangqi": {
			RootFunction: "checkLiangQiChengXiang", Functions: []string{"checkLiangQiChengXiang", "liangQiSemanticProfile", "patternBranchElementProfile", "patternElementForSymbol", "patternPillarContextSemanticProfile", "patternStemElementProfile"},
			ASTSHA256: "a4feed92bef94cf35624c525895a9872e1014f59aa273dd1c69cace155fc9c23",
		},
		"pattern.aux.kuigang": {
			RootFunction: "checkKuiGangGe", Functions: []string{"checkKuiGangGe", "kuiGangDayProfile", "patternStringProfileContains"},
			ASTSHA256: "552b8d2dd8cef0d7f5343638c4c07d870791a34283d50d6e7247e514334b4db6",
		},
		"pattern.aux.jinshen": {
			RootFunction: "checkJinShenHour", Functions: []string{"checkJinShenHour", "isJinShenHourPillar", "jinShenHourProfile", "jinShenSemanticProfile", "patternPillarContextSemanticProfile", "patternStringProfileContains"},
			ASTSHA256: "326f3feb4586cd0bf618737a597807f1da18b0ce70dda666c6589203993597e6",
		},
		"pattern.aux.sanqi": {
			RootFunction: "checkSanQi", Functions: []string{"checkSanQi", "classicalSanQiSequence", "patternPillarContextSemanticProfile", "patternStringProfileContains", "sanQiSemanticProfile", "sanQiSequenceProfile"},
			ASTSHA256: "69f4d63d045612b192a7276e3951e04104eb6b4eb8431aea62f6b7a712eca25c",
		},
		"pattern.aux.ride": {
			RootFunction: "checkRiDeGe", Functions: []string{"checkRiDeGe", "patternStringProfileContains", "riDeDayProfile"},
			ASTSHA256: "a5d3c9dc5ca9dcc125898be9a7305c8f1b88da9dc31f32f39e250d9c092276b2",
		},
	}
	profile, ok := profiles[ruleID]
	if !ok {
		return patternDetectorAlgorithmProfile{}, false
	}
	profile.Scheme = "go_ast_detector_closure_v1"
	profile.Functions = append([]string(nil), profile.Functions...)
	return profile, true
}

func patternDetectorAlgorithmSHA256(ruleID string) string {
	profile, ok := patternDetectorAlgorithmProfileForRule(ruleID)
	if !ok {
		return ""
	}
	return profile.ASTSHA256
}

type patternStemBranchTarget struct {
	Stem   string `json:"stem"`
	Branch string `json:"branch"`
}

func yangRenProfile() []patternStemBranchTarget {
	return []patternStemBranchTarget{
		{Stem: "甲", Branch: "卯"},
		{Stem: "丙", Branch: "午"},
		{Stem: "戊", Branch: "午"},
		{Stem: "庚", Branch: "酉"},
		{Stem: "壬", Branch: "子"},
	}
}

func kuiGangDayProfile() []string {
	return []string{"庚辰", "壬辰", "戊戌", "庚戌"}
}

func riDeDayProfile() []string {
	return []string{"甲寅", "丙辰", "戊辰", "庚辰", "壬戌"}
}

func jinShenHourProfile() []string {
	return []string{"癸酉", "己巳", "乙丑"}
}

func sanQiSequenceProfile() []string {
	return []string{"乙丙丁", "甲戊庚"}
}

type jinShenProfile struct {
	PillarCount int      `json:"pillar_count"`
	PillarIndex int      `json:"pillar_index"`
	Position    string   `json:"position"`
	Pillars     []string `json:"pillars"`
}

func jinShenSemanticProfile() jinShenProfile {
	context := patternPillarContextSemanticProfile()
	return jinShenProfile{
		PillarCount: context.PillarCount,
		PillarIndex: context.HourPillarIndex,
		Position:    "hour_pillar",
		Pillars:     jinShenHourProfile(),
	}
}

type sanQiProfile struct {
	PillarCount  int      `json:"pillar_count"`
	StemWindows  []string `json:"stem_windows"`
	WindowSize   int      `json:"window_size"`
	WindowStarts []int    `json:"window_starts"`
	Order        string   `json:"order"`
	Positions    []string `json:"positions"`
}

func sanQiSemanticProfile() sanQiProfile {
	context := patternPillarContextSemanticProfile()
	return sanQiProfile{
		PillarCount:  context.PillarCount,
		StemWindows:  sanQiSequenceProfile(),
		WindowSize:   3,
		WindowStarts: []int{0, 1},
		Order:        "adjacent_forward",
		Positions:    []string{"year_month_day", "month_day_hour"},
	}
}

type patternSymbolElementTarget struct {
	Symbol  string `json:"symbol"`
	Element string `json:"element"`
}

func patternStemElementProfile() []patternSymbolElementTarget {
	return []patternSymbolElementTarget{
		{Symbol: "甲", Element: "木"}, {Symbol: "乙", Element: "木"},
		{Symbol: "丙", Element: "火"}, {Symbol: "丁", Element: "火"},
		{Symbol: "戊", Element: "土"}, {Symbol: "己", Element: "土"},
		{Symbol: "庚", Element: "金"}, {Symbol: "辛", Element: "金"},
		{Symbol: "壬", Element: "水"}, {Symbol: "癸", Element: "水"},
	}
}

func patternBranchElementProfile() []patternSymbolElementTarget {
	return []patternSymbolElementTarget{
		{Symbol: "子", Element: "水"}, {Symbol: "丑", Element: "土"},
		{Symbol: "寅", Element: "木"}, {Symbol: "卯", Element: "木"},
		{Symbol: "辰", Element: "土"}, {Symbol: "巳", Element: "火"},
		{Symbol: "午", Element: "火"}, {Symbol: "未", Element: "土"},
		{Symbol: "申", Element: "金"}, {Symbol: "酉", Element: "金"},
		{Symbol: "戌", Element: "土"}, {Symbol: "亥", Element: "水"},
	}
}

func patternElementForSymbol(profile []patternSymbolElementTarget, symbol string) (string, bool) {
	for _, target := range profile {
		if target.Symbol == symbol {
			return target.Element, true
		}
	}
	return "", false
}

type liangQiProfile struct {
	ElementOrder          []string                     `json:"element_order"`
	StemElements          []patternSymbolElementTarget `json:"stem_elements"`
	BranchElements        []patternSymbolElementTarget `json:"branch_elements"`
	PillarCount           int                          `json:"pillar_count"`
	StemScope             string                       `json:"stem_scope"`
	BranchScope           string                       `json:"branch_scope"`
	DistinctElements      int                          `json:"distinct_elements"`
	OccurrencesPerElement int                          `json:"occurrences_per_element"`
}

func liangQiSemanticProfile() liangQiProfile {
	context := patternPillarContextSemanticProfile()
	return liangQiProfile{
		ElementOrder:          []string{"木", "火", "土", "金", "水"},
		StemElements:          patternStemElementProfile(),
		BranchElements:        patternBranchElementProfile(),
		PillarCount:           context.PillarCount,
		StemScope:             "four_heavenly_stems",
		BranchScope:           "four_branch_main_elements",
		DistinctElements:      2,
		OccurrencesPerElement: 4,
	}
}

func patternStringProfileContains(profile []string, value string) bool {
	for _, candidate := range profile {
		if candidate == value {
			return true
		}
	}
	return false
}

func patternBranchForStem(profile []patternStemBranchTarget, stem string) (string, bool) {
	for _, target := range profile {
		if target.Stem == stem {
			return target.Branch, true
		}
	}
	return "", false
}

type zhuanWangSemanticEntry struct {
	Element         string     `json:"element"`
	PatternName     string     `json:"pattern_name"`
	BreakingElement string     `json:"breaking_element"`
	Structures      [][]string `json:"structures"`
}

type zhuanWangDetectorProfile struct {
	PillarCount     int                          `json:"pillar_count"`
	DayElementScope string                       `json:"day_element_scope"`
	BranchScope     string                       `json:"branch_scope"`
	StemElements    []patternSymbolElementTarget `json:"stem_elements"`
	BranchElements  []patternSymbolElementTarget `json:"branch_elements"`
	Entries         []zhuanWangSemanticEntry     `json:"entries"`
}

func zhuanWangSemanticProfile() []zhuanWangSemanticEntry {
	registry := zhuanWangProfileRegistry()
	result := make([]zhuanWangSemanticEntry, 0, len(registry))
	for _, element := range []string{"木", "火", "土", "金", "水"} {
		profile := registry[element]
		entry := zhuanWangSemanticEntry{
			Element:         element,
			PatternName:     profile.name,
			BreakingElement: profile.breakingElement,
			Structures:      make([][]string, 0, len(profile.structures)),
		}
		for _, structure := range profile.structures {
			entry.Structures = append(entry.Structures, append([]string(nil), structure.branches...))
		}
		result = append(result, entry)
	}
	return result
}

func zhuanWangDetectorSemanticProfile() zhuanWangDetectorProfile {
	context := patternPillarContextSemanticProfile()
	return zhuanWangDetectorProfile{
		PillarCount:     context.PillarCount,
		DayElementScope: "day_stem_element",
		BranchScope:     "four_branch_main_elements",
		StemElements:    patternStemElementProfile(),
		BranchElements:  patternBranchElementProfile(),
		Entries:         zhuanWangSemanticProfile(),
	}
}

func patternDetectorOutputNames(ruleID string) []string {
	switch ruleID {
	case "pattern.special.zhuanwang":
		entries := zhuanWangSemanticProfile()
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.PatternName)
		}
		return names
	case "pattern.lu.jianlu":
		return []string{"建禄格"}
	case "pattern.lu.yueren":
		return []string{"月刃格"}
	case "pattern.lu.zhuanlu":
		return []string{"专禄格"}
	case "pattern.lu.riren":
		return []string{"日刃格"}
	case "pattern.special.liangqi":
		return []string{"两气成象格"}
	case "pattern.aux.kuigang":
		return []string{"魁罡格"}
	case "pattern.aux.jinshen":
		return []string{"金神"}
	case "pattern.aux.sanqi":
		return []string{"三奇"}
	case "pattern.aux.ride":
		return []string{"日德格"}
	default:
		return nil
	}
}

func patternDetectorSingleOutputName(ruleID string) (string, bool) {
	names := patternDetectorOutputNames(ruleID)
	if len(names) != 1 || names[0] == "" {
		return "", false
	}
	return names[0], true
}

func patternDetectorRuleSemanticProfile(ruleID string) (any, bool) {
	switch ruleID {
	case "pattern.special.zhuanwang":
		return zhuanWangDetectorSemanticProfile(), true
	case "pattern.lu.jianlu", "pattern.lu.zhuanlu":
		stems, branches := canonicalLuProfile()
		position := "month_branch"
		if ruleID == "pattern.lu.zhuanlu" {
			position = "day_branch"
		}
		return struct {
			Position string     `json:"position"`
			Stems    [10]string `json:"stems"`
			Branches [10]string `json:"branches"`
		}{position, stems, branches}, true
	case "pattern.lu.yueren", "pattern.lu.riren":
		position := "month_branch"
		if ruleID == "pattern.lu.riren" {
			position = "day_branch"
		}
		return struct {
			Position string                    `json:"position"`
			Targets  []patternStemBranchTarget `json:"targets"`
		}{position, yangRenProfile()}, true
	case "pattern.special.liangqi":
		return liangQiSemanticProfile(), true
	case "pattern.aux.kuigang":
		return struct {
			Position string   `json:"position"`
			Pillars  []string `json:"pillars"`
		}{"day_pillar", kuiGangDayProfile()}, true
	case "pattern.aux.jinshen":
		return jinShenSemanticProfile(), true
	case "pattern.aux.sanqi":
		return sanQiSemanticProfile(), true
	case "pattern.aux.ride":
		return struct {
			Position string   `json:"position"`
			Pillars  []string `json:"pillars"`
		}{"day_pillar", riDeDayProfile()}, true
	default:
		return nil, false
	}
}

func patternDetectorSemanticProfile(ruleID string) (any, bool) {
	detector, ok := patternDetectorRuleSemanticProfile(ruleID)
	outputNames := patternDetectorOutputNames(ruleID)
	algorithm, algorithmOK := patternDetectorAlgorithmProfileForRule(ruleID)
	if !ok || len(outputNames) == 0 || !algorithmOK {
		return nil, false
	}
	return patternDetectorSemanticEnvelope{
		PillarContext:       patternPillarContextSemanticProfile(),
		OutputNames:         outputNames,
		Algorithm:           algorithm,
		BehaviorWitnesses:   patternDetectorBehaviorWitnessProfile(ruleID),
		MetamorphicPolicies: patternDetectorMetamorphicProfile(ruleID),
		BehaviorManifest:    patternDetectorBehaviorManifestProfile(ruleID),
		Detector:            detector,
	}, true
}

func patternDetectorProfileSHA256(ruleID string) string {
	profile, ok := patternDetectorSemanticProfile(ruleID)
	if !ok {
		return ""
	}
	payload, err := json.Marshal(struct {
		RuleID  string `json:"rule_id"`
		Profile any    `json:"profile"`
	}{RuleID: ruleID, Profile: profile})
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}
