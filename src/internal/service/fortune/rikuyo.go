package fortune

import (
	"fmt"
	"strings"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

// ── 日课推算结果结构 ──────────────────────────────────────────

// RikuyoResult 日课推算完整结果
type RikuyoResult struct {
	// TodayTenGod is retained internally for period frequency aggregation.
	TodayTenGod string               `json:"-"`
	TenGod      model.TenGodEvidence `json:"ten_god"`

	// 今日地支十二长生
	TwelveStage model.TwelveStageEvidence `json:"twelve_stage"`

	// 地支藏干分析
	HiddenStems []HiddenStemGod `json:"hidden_stems"`

	// 干支作用关系
	StemRelations   []StemRelation   `json:"stem_relations"`
	BranchRelations []BranchRelation `json:"branch_relations"`

	// 神煞引动
	ActivatedShenSha []ShenShaActivation `json:"activated_shen_sha"`

	// 查询日天干在查询月令中的旺相休囚死查表结果。
	SeasonalState model.SeasonalStateEvidence `json:"seasonal_state"`

	// 建除十二神
	JianChu  model.TraditionalCalendarEvidence `json:"jian_chu"`
	HuangDao model.TraditionalCalendarEvidence `json:"huang_dao"`
}

type HiddenStemGod = model.HiddenStemGod
type StemRelation = model.StemRelation
type BranchRelation = model.BranchRelation
type ShenShaActivation = model.ShenShaActivation
type DaYunInfluence = model.DaYunInfluence
type LiuNianInfluence = model.LiuNianInfluence
type SeasonalStateEvidence = model.SeasonalStateEvidence

// ── 速查表数据 ──────────────────────────────────────────────

// twelveStageTable 十二长生速查表 [日主索引][地支索引] = 长生状态
// 阳干顺行：甲丙戊庚壬
// 阴干逆转：乙丁己辛癸
var twelveStageTable = [10][12]string{
	// 甲: 长生在亥, 沐浴子, 冠带丑, 临官寅, 帝旺卯, 衰辰, 病巳, 死午, 墓未, 绝申, 胎酉, 养戌
	{"沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝", "胎", "养", "长生"},
	// 乙: 长生在午, 逆行: 沐浴巳, 冠带辰, 临官卯, 帝旺寅, 衰丑, 病子, 死亥, 墓戌, 绝酉, 胎申, 养未
	{"病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎", "绝", "墓", "死"},
	// 丙: 长生在寅, 沐浴卯, 冠带辰, 临官巳, 帝旺午, 衰未, 病申, 死酉, 墓戌, 绝亥, 胎子, 养丑
	{"胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝"},
	// 丁: 长生在酉, 逆行: 沐浴申, 冠带未, 临官午, 帝旺巳, 衰辰, 病卯, 死寅, 墓丑, 绝子, 胎亥, 养戌
	{"绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎"},
	// 戊: 长生在寅(与丙同), 沐浴卯, 冠带辰, 临官巳, 帝旺午, 衰未, 病申, 死酉, 墓戌, 绝亥, 胎子, 养丑
	{"胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病", "死", "墓", "绝"},
	// 己: 长生在酉(与丁同), 逆行
	{"绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴", "长生", "养", "胎"},
	// 庚: 长生在巳, 沐浴午, 冠带未, 临官申, 帝旺酉, 衰戌, 病亥, 死子, 墓丑, 绝寅, 胎卯, 养辰
	{"死", "墓", "绝", "胎", "养", "长生", "沐浴", "冠带", "临官", "帝旺", "衰", "病"},
	// 辛: 长生在子, 逆行: 沐浴亥, 冠带戌, 临官酉, 帝旺申, 衰未, 病午, 死巳, 墓辰, 绝卯, 胎寅, 养丑
	{"长生", "养", "胎", "绝", "墓", "死", "病", "衰", "帝旺", "临官", "冠带", "沐浴"},
	// 壬: 长生在申, 沐浴酉, 冠带戌, 临官亥, 帝旺子, 衰丑, 病寅, 死卯, 墓辰, 绝巳, 胎午, 养未
	{"帝旺", "衰", "病", "死", "墓", "绝", "胎", "养", "长生", "沐浴", "冠带", "临官"},
	// 癸: 长生在卯, 逆行: 沐浴寅, 冠带丑, 临官子, 帝旺亥, 衰戌, 病酉, 死申, 墓未, 绝午, 胎巳, 养辰
	{"临官", "冠带", "沐浴", "长生", "养", "胎", "绝", "墓", "死", "病", "衰", "帝旺"},
}

// hiddenStemMap 地支藏干表 [地支索引] = {本气, 中气, 余气}
var hiddenStemMap = [12][3]string{
	{"癸", "", ""},   // 子
	{"己", "癸", "辛"}, // 丑
	{"甲", "丙", "戊"}, // 寅
	{"乙", "", ""},   // 卯
	{"戊", "乙", "癸"}, // 辰
	{"丙", "庚", "戊"}, // 巳
	{"丁", "己", ""},  // 午
	{"己", "丁", "乙"}, // 未
	{"庚", "壬", "戊"}, // 申
	{"辛", "", ""},   // 酉
	{"戊", "辛", "丁"}, // 戌
	{"壬", "甲", ""},  // 亥
}

// stemCombineMap 天干五合 map[天干] = 合化对象
var stemCombineMap = map[string]string{
	"甲": "己", "己": "甲",
	"乙": "庚", "庚": "乙",
	"丙": "辛", "辛": "丙",
	"丁": "壬", "壬": "丁",
	"戊": "癸", "癸": "戊",
}

// combineElement 五合化五行（双向key）
var combineElement = map[string]string{
	"甲己": "土", "己甲": "土",
	"乙庚": "金", "庚乙": "金",
	"丙辛": "水", "辛丙": "水",
	"丁壬": "木", "壬丁": "木",
	"戊癸": "火", "癸戊": "火",
}

// stemClashMap 天干相冲
var stemClashMap = map[string]string{
	"甲": "庚", "庚": "甲",
	"乙": "辛", "辛": "乙",
	"丙": "壬", "壬": "丙",
	"丁": "癸", "癸": "丁",
}

// branchSixClash 地支六冲
var branchSixClash = map[string]string{
	"子": "午", "午": "子", "丑": "未", "未": "丑",
	"寅": "申", "申": "寅", "卯": "酉", "酉": "卯",
	"辰": "戌", "戌": "辰", "巳": "亥", "亥": "巳",
}

// branchSixCombine 地支六合
var branchSixCombine = map[string]string{
	"子": "丑", "丑": "子", "寅": "亥", "亥": "寅",
	"卯": "戌", "戌": "卯", "辰": "酉", "酉": "辰",
	"巳": "申", "申": "巳", "午": "未", "未": "午",
}

// branchSixHarm 地支六害
var branchSixHarm = map[string]string{
	"子": "未", "未": "子", "丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅", "卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申", "酉": "戌", "戌": "酉",
}

// branchThreePunishment 三刑
var branchThreePunishment = []struct {
	Name  string
	Group []string
	Desc  string
}{
	{"无礼之刑", []string{"子", "卯"}, "子刑卯，卯刑子，无礼之刑。主无礼犯上，是非口舌。"},
	// 经典依据：三命通会第38章——寅巳申为无恩之刑，丑戌未为恃势之刑
	{"无恩之刑", []string{"寅", "巳", "申"}, "寅刑巳，巳刑申，申刑寅，无恩之刑。主忘恩负义，辗转相克，事情反复难解。"},
	{"恃势之刑", []string{"丑", "戌", "未"}, "丑刑戌，戌刑未，未刑丑，恃势之刑。主仗势欺人，恃强凌弱，暗中损耗。"},
	{"自刑", []string{"辰", "午", "酉", "亥"}, "辰午酉亥自刑。主自我矛盾，自找麻烦。"},
}

// branchThreeCombine 地支三合
var branchThreeCombine = []struct {
	Element string
	Group   []string
}{
	{"水", []string{"申", "子", "辰"}},
	{"木", []string{"亥", "卯", "未"}},
	{"火", []string{"寅", "午", "戌"}},
	{"金", []string{"巳", "酉", "丑"}},
}

// branchThreeMeeting 地支三会
var branchThreeMeeting = []struct {
	Element string
	Group   []string
}{
	{"木", []string{"寅", "卯", "辰"}},
	{"火", []string{"巳", "午", "未"}},
	{"金", []string{"申", "酉", "戌"}},
	{"水", []string{"亥", "子", "丑"}},
}

// elementCycle 五行相生相克
var elementGeneratesMap = map[string]string{
	"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
}
var elementOvercomesMap = map[string]string{
	"木": "土", "土": "水", "水": "火", "火": "金", "金": "木",
}

var jianChuNames = []string{"建", "除", "满", "平", "定", "执", "破", "危", "成", "收", "开", "闭"}

// ── 天乙贵人速查（以日干查地支） ──────────────────────────
var tianYiGuiren = map[string][]string{
	"甲": {"丑", "未"}, "戊": {"丑", "未"}, "庚": {"丑", "未"},
	"乙": {"子", "申"}, "己": {"子", "申"},
	"丙": {"亥", "酉"}, "丁": {"亥", "酉"},
	"壬": {"卯", "巳"}, "癸": {"卯", "巳"},
	"辛": {"寅", "午"},
}

// ── 驿马速查（以日支三合局取） ──────────────────────────────
var yiMa = map[string]string{
	"寅": "申", "午": "申", "戌": "申", // 寅午戌见申
	"申": "寅", "子": "寅", "辰": "寅", // 申子辰见寅
	"巳": "亥", "酉": "亥", "丑": "亥", // 巳酉丑见亥
	"亥": "巳", "卯": "巳", "未": "巳", // 亥卯未见巳
}

// ── 桃花速查（以日支三合局取） ──────────────────────────────
var taoHua = map[string]string{
	"寅": "卯", "午": "卯", "戌": "卯", // 寅午戌见卯
	"申": "酉", "子": "酉", "辰": "酉", // 申子辰见酉
	"巳": "午", "酉": "午", "丑": "午", // 巳酉丑见午
	"亥": "子", "卯": "子", "未": "子", // 亥卯未见子
}

// ── 禄神速查（以日干查地支） ──────────────────────────────
var luShen = map[string]string{
	"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午",
	"戊": "巳", "己": "午", "庚": "申", "辛": "酉",
	"壬": "亥", "癸": "子",
}

// ── 进退气季节表 ──────────────────────────────────────────
// 五行在四季的状态：旺(当令)/相(进气)/休(退气)/囚(无气)/死
var elementSeasonState = map[string]map[string]string{
	"春": {"木": "旺", "火": "相", "水": "休", "金": "囚", "土": "死"},
	"夏": {"火": "旺", "土": "相", "木": "休", "水": "囚", "金": "死"},
	"秋": {"金": "旺", "水": "相", "土": "休", "火": "囚", "木": "死"},
	"冬": {"水": "旺", "木": "相", "金": "休", "土": "囚", "火": "死"},
}

// ── 核心计算函数 ──────────────────────────────────────────

// CalcRikuyo 日课推算主入口
func CalcRikuyo(bazi *bazipkg.BaziResult, queryDate time.Time) *RikuyoResult {
	dayGan := bazi.DayPillar.Gan

	// 获取今日干支
	qYear, qMonth, qDay := queryDate.Year(), int(queryDate.Month()), queryDate.Day()
	ec, err := getDayEightChar(qYear, qMonth, qDay)
	if err != nil {
		return &RikuyoResult{}
	}
	todayGan := ec.GetDay().GetHeavenStem().GetName()
	todayZhi := ec.GetDay().GetEarthBranch().GetName()
	monthZhi := ec.GetMonth().GetEarthBranch().GetName()

	// 步骤一：今日天干取十神，只记录结构映射。
	tenGodEvidence := observeTenGod(dayGan, todayGan)
	tenGod := tenGodEvidence.Name

	// 步骤二：今日地支取十二长生结构标签
	twelveStage := observeTwelveStage(dayGan, todayZhi)

	// 步骤三：地支藏干分析（格局感知）
	hiddenStems := calcHiddenStemGods(todayZhi, dayGan)

	// 步骤四：天干关系（今日天干 vs 命局四柱天干）
	stemRels := calcStemRelations(todayGan, bazi)

	// 步骤五：地支关系（今日地支 vs 命局四柱地支）
	branchRels := calcBranchRelations(todayZhi, bazi)

	// 步骤六：神煞引动
	shenSha := calcShenShaActivation(todayGan, todayZhi, bazi)

	// 步骤七：查询日天干在查询月令中的传统状态查表。
	seasonalState := observeSeasonalState(todayGan, monthZhi)

	// 传统日课标签只记录查表事实，不进入吉凶评分。
	jianChu := observeJianChu(monthZhi, todayZhi)
	huangDao := observeHuangDao(monthZhi, todayZhi)

	return &RikuyoResult{
		TodayTenGod:      tenGod,
		TenGod:           tenGodEvidence,
		TwelveStage:      twelveStage,
		HiddenStems:      hiddenStems,
		StemRelations:    stemRels,
		BranchRelations:  branchRels,
		ActivatedShenSha: shenSha,
		SeasonalState:    seasonalState,
		JianChu:          jianChu,
		HuangDao:         huangDao,
	}
}

func observeTenGod(referenceStem, queryStem string) model.TenGodEvidence {
	evidence := model.TenGodEvidence{
		RuleID:               "rikuyo.ten-god-day-stem-v1",
		ReferenceStem:        referenceStem,
		QueryStem:            queryStem,
		Basis:                "reference_day_stem_and_query_day_stem",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	if data.GanIndex(referenceStem) < 0 || data.GanIndex(queryStem) < 0 {
		return evidence
	}
	evidence.Name = bazipkg.ClassifyTenGod(queryStem, referenceStem, false)
	if evidence.Name == "" {
		return evidence
	}
	evidence.Status = "observed"
	return evidence
}

// ── 步骤二：十二长生结构标签 ────────────────────────────────

func observeTwelveStage(dayGan, branch string) model.TwelveStageEvidence {
	evidence := model.TwelveStageEvidence{
		RuleID:               "rikuyo.twelve-stage-v1",
		ReferenceStem:        dayGan,
		QueryBranch:          branch,
		Basis:                "reference_day_stem_and_query_day_branch",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	ganIdx := data.GanIndex(dayGan)
	zhiIdx := data.ZhiIndex(branch)
	if ganIdx < 0 || zhiIdx < 0 {
		return evidence
	}
	evidence.Name = twelveStageTable[ganIdx][zhiIdx]
	evidence.Status = "observed"
	return evidence
}

// ── 步骤三：藏干十神分析 ──────────────────────────────────

func calcHiddenStemGods(branch, dayGan string) []HiddenStemGod {
	zhiIdx := data.ZhiIndex(branch)
	if zhiIdx < 0 {
		return nil
	}

	types := []string{"本气", "中气", "余气"}
	var result []HiddenStemGod

	for i, s := range hiddenStemMap[zhiIdx] {
		if s == "" {
			continue
		}
		result = append(result, HiddenStemGod{
			RuleID:               "rikuyo.hidden-stem-ten-god-v1",
			QueryBranch:          branch,
			ReferenceStem:        dayGan,
			Stem:                 s,
			Type:                 types[i],
			Element:              data.GanElement[s],
			TenGod:               bazipkg.ClassifyTenGod(s, dayGan, false),
			Basis:                "query_branch_hidden_stem_table_and_reference_day_stem",
			Status:               "observed",
			InterpretationStatus: "not_adjudicated",
		})
	}
	return result
}

// ── 步骤四：天干关系 ──────────────────────────────────────

func calcStemRelations(todayGan string, bazi *bazipkg.BaziResult) []StemRelation {
	birthGans := []struct {
		name  string
		label string
	}{
		{bazi.YearPillar.Gan, "年干"},
		{bazi.MonthPillar.Gan, "月干"},
		{bazi.DayPillar.Gan, "日干"},
		{bazi.HourPillar.Gan, "时干"},
	}

	var rels []StemRelation
	for _, bg := range birthGans {
		if data.GanIndex(todayGan) < 0 || data.GanIndex(bg.name) < 0 {
			continue
		}

		type relationMatch struct {
			relationType, relationName, combinedElement, transformationStatus string
		}
		matches := make([]relationMatch, 0, 2)
		if stemCombineMap[todayGan] == bg.name {
			matches = append(matches, relationMatch{
				relationType: "five_combine", relationName: "五合",
				combinedElement: combineElement[todayGan+bg.name], transformationStatus: "not_adjudicated",
			})
		}
		if stemClashMap[todayGan] == bg.name {
			matches = append(matches, relationMatch{
				relationType: "clash", relationName: "相冲", transformationStatus: "not_applicable",
			})
		}

		queryElement := data.GanElement[todayGan]
		targetElement := data.GanElement[bg.name]
		elementMatch := relationMatch{transformationStatus: "not_applicable"}
		switch {
		case todayGan == bg.name:
			elementMatch.relationType, elementMatch.relationName = "same_stem", "同干"
		case queryElement == targetElement:
			elementMatch.relationType, elementMatch.relationName = "same_element", "同五行"
		case elementGeneratesMap[queryElement] == targetElement:
			elementMatch.relationType, elementMatch.relationName = "query_generates_target", "今日天干生本命天干"
		case elementGeneratesMap[targetElement] == queryElement:
			elementMatch.relationType, elementMatch.relationName = "target_generates_query", "本命天干生今日天干"
		case elementOvercomesMap[queryElement] == targetElement:
			elementMatch.relationType, elementMatch.relationName = "query_overcomes_target", "今日天干克本命天干"
		case elementOvercomesMap[targetElement] == queryElement:
			elementMatch.relationType, elementMatch.relationName = "target_overcomes_query", "本命天干克今日天干"
		}
		if elementMatch.relationType != "" {
			matches = append(matches, elementMatch)
		}

		for _, match := range matches {
			rels = append(rels, StemRelation{
				RuleID:               "rikuyo.stem-relation-v3." + match.relationType,
				QueryStem:            todayGan,
				TargetPillar:         bg.label,
				TargetStem:           bg.name,
				Type:                 match.relationType,
				Name:                 match.relationName,
				CombinedElement:      match.combinedElement,
				Basis:                "query_day_stem_and_natal_pillar_stem_all_structures",
				Status:               "observed",
				TransformationStatus: match.transformationStatus,
				InterpretationStatus: "not_adjudicated",
			})
		}
	}
	return rels
}

// ── 步骤五：地支关系 ──────────────────────────────────────

func calcBranchRelations(todayZhi string, bazi *bazipkg.BaziResult) []BranchRelation {
	birthZhis := []struct {
		name  string
		label string
	}{
		{bazi.YearPillar.Zhi, "年支"},
		{bazi.MonthPillar.Zhi, "月支"},
		{bazi.DayPillar.Zhi, "日支"},
		{bazi.HourPillar.Zhi, "时支"},
	}

	var rels []BranchRelation
	for _, bz := range birthZhis {
		if data.ZhiIndex(todayZhi) < 0 || data.ZhiIndex(bz.name) < 0 {
			continue
		}
		for _, relation := range layerBranchRelations("流日", todayZhi, bz.label, bz.name) {
			transformationStatus := "not_applicable"
			switch relation.Type {
			case "combine", "banHe", "gongHe", "banHui", "sanHe", "sanHui":
				transformationStatus = "not_adjudicated"
			}
			rels = append(rels, BranchRelation{
				RuleID:               "rikuyo.branch-relation-v3." + relation.Type,
				QueryBranch:          todayZhi,
				TargetPillar:         bz.label,
				TargetBranch:         bz.name,
				Type:                 relation.Type,
				Name:                 relation.Name,
				Basis:                "query_day_branch_and_natal_pillar_branch_all_structures",
				Status:               "observed",
				TransformationStatus: transformationStatus,
				InterpretationStatus: "not_adjudicated",
			})
		}
	}

	return rels
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// ── 步骤六：神煞引动 ──────────────────────────────────────

func calcShenShaActivation(todayGan, todayZhi string, bazi *bazipkg.BaziResult) []ShenShaActivation {
	dayGan := bazi.DayPillar.Gan
	dayZhi := bazi.DayPillar.Zhi
	yearZhi := bazi.YearPillar.Zhi
	var result []ShenShaActivation

	// 天乙贵人
	if guas, ok := tianYiGuiren[dayGan]; ok {
		for _, g := range guas {
			if todayZhi == g {
				result = append(result, newShenShaActivation("天乙贵人", fmt.Sprintf("查询日地支%s命中日干%s的天乙贵人位", todayZhi, dayGan)))
			}
		}
	}

	// 驿马、咸池均以年支或日支所属三合组查目标位。
	if references := matchingSanHeReferences(todayZhi, yiMa, yearZhi, dayZhi); len(references) > 0 {
		result = append(result, newShenShaActivation("驿马", fmt.Sprintf(
			"查询日地支%s命中%s所定驿马位", todayZhi, strings.Join(references, "、"),
		)))
	}
	if references := matchingSanHeReferences(todayZhi, taoHua, yearZhi, dayZhi); len(references) > 0 {
		result = append(result, newShenShaActivation("咸池", fmt.Sprintf(
			"查询日地支%s命中%s所定咸池位", todayZhi, strings.Join(references, "、"),
		)))
	}

	// 禄神
	if luShen[dayGan] == todayZhi {
		result = append(result, newShenShaActivation("禄神", fmt.Sprintf("查询日地支%s命中日干%s的禄神位", todayZhi, dayGan)))
	}

	return result
}

func matchingSanHeReferences(todayZhi string, targets map[string]string, yearZhi, dayZhi string) []string {
	references := make([]string, 0, 2)
	if targets[yearZhi] == todayZhi {
		references = append(references, "年支"+yearZhi)
	}
	if targets[dayZhi] == todayZhi {
		references = append(references, "日支"+dayZhi)
	}
	return references
}

func newShenShaActivation(name, activation string) ShenShaActivation {
	meta := bazipkg.LookupShenShaMeta(name)
	return ShenShaActivation{
		Name:                 name,
		RuleID:               meta.RuleID,
		Basis:                meta.Basis,
		Status:               meta.Status,
		InterpretationStatus: meta.InterpretationStatus,
		Activation:           activation,
	}
}

// ── 步骤七：大运叠加 ──────────────────────────────────────

func calcDaYunInfluence(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) DaYunInfluence {
	if len(bazi.DaYunInfo.Pillars) == 0 {
		return DaYunInfluence{Status: "unavailable", InterpretationStatus: "not_adjudicated"}
	}

	age := queryDate.Year() - birthYear
	startAge := bazi.DaYunInfo.StartAge
	idx, periodStart, periodEnd, exactStatus := exactDaYunPeriod(bazi.DaYunInfo, queryDate)
	if exactStatus == daYunBeforeStart {
		return DaYunInfluence{
			Active:               false,
			Index:                -1,
			StartAt:              periodStart,
			SelectionBasis:       "exact_start_time_and_query_time",
			Status:               "before_start",
			InterpretationStatus: "not_adjudicated",
		}
	}
	if exactStatus == daYunAfterCoveredPeriods {
		return DaYunInfluence{
			Active:               false,
			Index:                len(bazi.DaYunInfo.Pillars),
			StartAt:              periodStart,
			SelectionBasis:       "exact_start_time_and_query_time",
			Status:               "after_covered_periods",
			InterpretationStatus: "not_adjudicated",
		}
	}

	// Old or pillar-only inputs do not contain a start timestamp. Keep their
	// integer-age calculation isolated as a clearly less precise fallback.
	if exactStatus == daYunTimeUnavailable {
		if age < startAge {
			return DaYunInfluence{
				Active:               false,
				Index:                -1,
				SelectionBasis:       "integer_age_fallback",
				Status:               "before_start",
				InterpretationStatus: "not_adjudicated",
			}
		}
		idx = (age - startAge) / 10
		if idx >= len(bazi.DaYunInfo.Pillars) {
			return DaYunInfluence{
				Active:               false,
				Index:                len(bazi.DaYunInfo.Pillars),
				SelectionBasis:       "integer_age_fallback",
				Status:               "after_covered_periods",
				InterpretationStatus: "not_adjudicated",
			}
		}
	}

	pillar := bazi.DaYunInfo.Pillars[idx]
	daYunGan := pillar.Gan
	daYunZhi := pillar.Zhi
	dayGan := bazi.DayPillar.Gan

	// 计算当前大运精确的起止年龄和年份
	daYunStartAge := startAge + idx*10
	daYunEndAge := daYunStartAge + 9

	selectionBasis := "exact_start_time_and_query_time"
	if exactStatus == daYunTimeUnavailable {
		selectionBasis = "integer_age_fallback"
	}

	// 大运天干对日主的十神结构映射。
	tenGod := bazipkg.ClassifyTenGod(daYunGan, dayGan, false)

	return DaYunInfluence{
		CurrentPillar:        daYunGan + daYunZhi,
		Active:               true,
		Index:                idx,
		StartAt:              periodStart,
		EndAtExclusive:       periodEnd,
		StartAge:             daYunStartAge,
		EndAge:               daYunEndAge,
		TenGod:               tenGod,
		SelectionBasis:       selectionBasis,
		Status:               "observed",
		InterpretationStatus: "not_adjudicated",
	}
}

type daYunPeriodStatus int

const (
	daYunTimeUnavailable daYunPeriodStatus = iota
	daYunPeriodActive
	daYunBeforeStart
	daYunAfterCoveredPeriods
)

const daYunTimeLayout = "2006-01-02T15:04:05"

func exactDaYunPeriod(info bazipkg.DaYunInfo, queryDate time.Time) (int, string, string, daYunPeriodStatus) {
	if !info.Calculated || info.StartAt == "" {
		return 0, "", "", daYunTimeUnavailable
	}
	start, err := time.ParseInLocation(daYunTimeLayout, info.StartAt, queryDate.Location())
	if err != nil {
		return 0, "", "", daYunTimeUnavailable
	}
	if queryDate.Before(start) {
		return -1, start.Format(daYunTimeLayout), "", daYunBeforeStart
	}
	for idx := range info.Pillars {
		periodStart := start.AddDate(idx*10, 0, 0)
		periodEnd := start.AddDate((idx+1)*10, 0, 0)
		if queryDate.Before(periodEnd) {
			return idx, periodStart.Format(daYunTimeLayout), periodEnd.Format(daYunTimeLayout), daYunPeriodActive
		}
	}
	coveredUntil := start.AddDate(len(info.Pillars)*10, 0, 0)
	return len(info.Pillars), coveredUntil.Format(daYunTimeLayout), "", daYunAfterCoveredPeriods
}

// queryDateGan helper
func queryDateGan(queryDate time.Time) string {
	ec, err := getDayEightChar(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
	if err != nil {
		return ""
	}
	return ec.GetDay().GetHeavenStem().GetName()
}

// ── 步骤八：流年叠加 ──────────────────────────────────────

func calcLiuNianInfluence(bazi *bazipkg.BaziResult, queryDate time.Time) LiuNianInfluence {
	yearGanZhi := getYearGanZhi(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
	if len(yearGanZhi) < 2 {
		return LiuNianInfluence{Status: "unavailable", InterpretationStatus: "not_adjudicated"}
	}
	runes := []rune(yearGanZhi)
	yearGan := string(runes[0])
	yearZhi := string(runes[1])
	dayGan := bazi.DayPillar.Gan

	// 流年天干对日主的十神结构映射。
	tenGod := bazipkg.ClassifyTenGod(yearGan, dayGan, false)

	return LiuNianInfluence{
		YearPillar:           yearGan + yearZhi,
		TenGod:               tenGod,
		SelectionBasis:       "query_date_year_pillar_at_solar_term_boundary",
		Status:               "observed",
		InterpretationStatus: "not_adjudicated",
	}
}

// getYearGanZhi uses the calendar engine's exact solar-term boundary.
func getYearGanZhi(year, month, day int) string {
	ec, err := getDayEightChar(year, month, day)
	if err != nil {
		return ""
	}
	return ec.GetYear().GetName()
}

func observeSeasonalState(todayGan, monthZhi string) SeasonalStateEvidence {
	elem := data.GanElement[todayGan]
	season := monthZhiToSeason(monthZhi)
	evidence := SeasonalStateEvidence{
		RuleID:               "rikuyo.seasonal-state-v1",
		QueryStem:            todayGan,
		QueryElement:         elem,
		QueryMonthBranch:     monthZhi,
		Season:               season,
		Basis:                "query_day_stem_element_and_query_month_branch",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	if elem == "" || season == "" {
		return evidence
	}
	evidence.State = elementSeasonState[season][elem]
	if evidence.State == "" {
		return evidence
	}
	evidence.Status = "observed"
	return evidence
}

// monthZhiToSeason 以节气月地支判断季节
// 寅卯辰=春, 巳午未=夏, 申酉戌=秋, 亥子丑=冬
func monthZhiToSeason(zhi string) string {
	switch zhi {
	case "寅", "卯", "辰":
		return "春"
	case "巳", "午", "未":
		return "夏"
	case "申", "酉", "戌":
		return "秋"
	case "亥", "子", "丑":
		return "冬"
	default:
		return ""
	}
}

// solarDateToJieQiMonth 根据公历日期近似推算节气月（1-12）
// 节气月以每月节气日为界（约4-8日），此处用简化日期近似
func solarDateToJieQiMonth(month, day int) int {
	jieQiDays := []int{6, 4, 6, 5, 6, 6, 7, 7, 8, 8, 7, 7}
	if day >= jieQiDays[month-1] {
		return month
	}
	// 在当月节气日之前，属上一个节气月
	if month == 1 {
		return 12
	}
	return month - 1
}

// ── 传统日课结构标签 ──────────────────────────────────────────

func observeJianChu(monthZhi, todayZhi string) model.TraditionalCalendarEvidence {
	evidence := model.TraditionalCalendarEvidence{
		RuleID:               "rikuyo.jianchu-month-branch-v1",
		MonthBranch:          monthZhi,
		QueryBranch:          todayZhi,
		Basis:                "query_month_branch_and_query_day_branch",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	monthIdx, todayIdx := data.ZhiIndex(monthZhi), data.ZhiIndex(todayZhi)
	if monthIdx < 0 || todayIdx < 0 {
		return evidence
	}
	evidence.Name = jianChuNames[(todayIdx-monthIdx+12)%12]
	evidence.Status = "observed"
	return evidence
}

// ── 十二值神结构标签（协纪辨方书） ──────────────────────────────

// huangDaoTwelveGods 十二值神顺排顺序
var huangDaoTwelveGods = []string{
	"青龙", "明堂", "天刑", "朱雀", "金匮", "天德",
	"白虎", "玉堂", "天牢", "玄武", "司命", "勾陈",
}

func observeHuangDao(monthZhi, todayZhi string) model.TraditionalCalendarEvidence {
	evidence := model.TraditionalCalendarEvidence{
		RuleID:               "rikuyo.twelve-star.tyme4go-v2",
		MonthBranch:          monthZhi,
		QueryBranch:          todayZhi,
		Basis:                "tyme4go_sixty_cycle_day_twelve_star_formula",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	monthIdx, todayIdx := data.ZhiIndex(monthZhi), data.ZhiIndex(todayZhi)
	if monthIdx < 0 || todayIdx < 0 {
		return evidence
	}
	evidence.Name = huangDaoTwelveGods[(todayIdx+(8-monthIdx%6)*2)%12]
	evidence.Status = "observed"
	return evidence
}
