package fortune

import (
	"fmt"
	"time"

	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

// ── 日课推算结果结构 ──────────────────────────────────────────

// RikuyoResult 日课推算完整结果
type RikuyoResult struct {
	// 今日天干十神
	TodayTenGod     string `json:"today_ten_god"`
	TenGodFavorable bool   `json:"ten_god_favorable"`
	TenGodDesc      string `json:"ten_god_desc"`

	// 今日地支十二长生
	TwelveStage    string `json:"twelve_stage"`
	StageFavorable bool   `json:"stage_favorable"`
	StageDesc      string `json:"stage_desc"`
	StageFlexible  string `json:"stage_flexible"` // 活法修正说明

	// 地支藏干分析
	HiddenStems []HiddenStemGod `json:"hidden_stems"`

	// 干支作用关系
	StemRelations   []StemRelation   `json:"stem_relations"`
	BranchRelations []BranchRelation `json:"branch_relations"`

	// 神煞引动
	ActivatedShenSha []ShenShaActivation `json:"activated_shen_sha"`

	// 大运流年叠加
	DaYunInfluence   DaYunInfluence   `json:"dayun_influence"`
	LiuNianInfluence LiuNianInfluence `json:"liunian_influence"`

	// 进退气
	AdvanceRetreat AdvanceRetreat `json:"advance_retreat"`

	// 用神综合影响
	YongShenImpact YongShenImpact `json:"yongshen_impact"`

	// 综合判断
	OverallVerdict string `json:"overall_verdict"`
	FavorScore     int    `json:"favor_score"`

	// 格局信息
	PatternName         string   `json:"pattern_name"`
	PatternType         string   `json:"pattern_type"`
	PatternFavorable    []string `json:"pattern_favorable"`
	PatternUnfavorable  []string `json:"pattern_unfavorable"`
}

// HiddenStemGod 地支藏干及其十神
type HiddenStemGod struct {
	Stem      string `json:"stem"`       // 藏干天干
	Type      string `json:"type"`       // 本气/中气/余气
	Element   string `json:"element"`    // 五行
	TenGod    string `json:"ten_god"`    // 十神名称
	Favorable bool   `json:"favorable"`  // 喜还是忌
}

// StemRelation 天干关系
type StemRelation struct {
	Type       string `json:"type"`       // 五合/相克/相生
	Target     string `json:"target"`     // 命局天干
	Detail     string `json:"detail"`     // 具体关系描述
	IsFavorable bool  `json:"is_favorable"`
	Note       string `json:"note"`       // 备注（如贪合忘克）
}

// BranchRelation 地支关系
type BranchRelation struct {
	Type       string `json:"type"`       // 六合/六冲/六害/三刑/三合/三会
	Target     string `json:"target"`     // 命局地支
	Detail     string `json:"detail"`
	IsFavorable bool  `json:"is_favorable"`
}

// ShenShaActivation 神煞引动
type ShenShaActivation struct {
	Name        string `json:"name"`         // 神煞名称
	Type        string `json:"type"`         // 吉神/凶煞
	Description string `json:"description"`  // 含义
	Activation  string `json:"activation"`   // 引动方式
}

// DaYunInfluence 大运影响
type DaYunInfluence struct {
	CurrentPillar string `json:"current_pillar"` // 当前大运干支
	StartAge      int    `json:"start_age"`      // 起运年龄
	EndAge        int    `json:"end_age"`        // 结束年龄
	TenGod        string `json:"ten_god"`        // 大运天干对日主的十神
	Favorable     bool   `json:"favorable"`      // 大运是否有利
	Relation      string `json:"relation"`       // 今日与大运的关系
	Score         int    `json:"score"`          // 大运对今日的影响分
	Description   string `json:"description"`
}

// LiuNianInfluence 流年影响
type LiuNianInfluence struct {
	YearPillar  string `json:"year_pillar"`  // 流年干支
	TenGod      string `json:"ten_god"`      // 流年天干对日主的十神
	Favorable   bool   `json:"favorable"`    // 流年是否有利
	Relation    string `json:"relation"`     // 今日与流年的关系
	TaiSuiRelation string `json:"tai_sui_relation"` // 太岁关系（岁伤日干/日犯岁君）
	Score       int    `json:"score"`
	Description string `json:"description"`
}

// AdvanceRetreat 进退气分析
type AdvanceRetreat struct {
	Phase       string `json:"phase"`       // 进气/当令/退气/无气
	PhaseDesc   string `json:"phase_desc"`  // 阶段描述
	Element     string `json:"element"`     // 分析的五行
	Score       int    `json:"score"`       // 进退气得分
	Description string `json:"description"` // 综合说明
}

// YongShenImpact 用神综合影响
type YongShenImpact struct {
	TiaoHouElement string `json:"tiao_hou_element"` // 调候用神五行
	TiaoHouHit     bool   `json:"tiao_hou_hit"`     // 今日是否生扶调候用神
	TongGuanElement string `json:"tong_guan_element"` // 通关用神五行
	TongGuanHit    bool   `json:"tong_guan_hit"`
	FuYiElements   []string `json:"fu_yi_elements"`  // 扶抑喜用五行
	FuYiHit        bool   `json:"fu_yi_hit"`
	Score          int    `json:"score"`
	Description    string `json:"description"`
}

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

// stageFavorability 长生状态吉凶
var stageFavorability = map[string]string{
	"长生": "吉", "冠带": "吉", "临官": "吉", "帝旺": "吉",
	"沐浴": "半吉", "胎": "半吉", "养": "半吉",
	"衰": "半凶", "病": "凶", "死": "凶", "墓": "半凶", "绝": "凶",
}

// stageDescriptions 长生状态描述
var stageDescriptions = map[string]string{
	"长生": "气始发，旭日东升。主创建、开创之事，精力渐充。",
	"沐浴": "气初生而弱，如婴儿洗浴。主落魄、酒色，半吉半凶。",
	"冠带": "气渐长，如少年加冠。主渐进成长，学业有成。",
	"临官": "气旺盛，如入仕为官。主兴盛发达，事业顺遂。",
	"帝旺": "气极盛，如日中天。主极盛巅峰，但盛极必衰。",
	"衰": "气始退，如人过壮年。主退败渐衰，力不从心。",
	"病": "气已弱，如人患病。主困顿多事，诸事不顺。",
	"死": "气已绝，如人已死。主丧祸大凶，宜守不宜进。",
	"墓": "气收藏，如入墓库。主停滞阻滞，蓄势待发。",
	"绝": "气完全消失，如断根之木。主断绝极凶，但绝处逢生。",
	"胎": "气重新凝聚，如胎儿孕育。主孕育潜伏，韬光养晦。",
	"养": "气渐充盈，如婴儿哺育。主滋养蓄势，蓄积力量。",
}

// hiddenStemMap 地支藏干表 [地支索引] = {本气, 中气, 余气}
var hiddenStemMap = [12][3]string{
	{"癸", "", ""},       // 子
	{"己", "癸", "辛"},   // 丑
	{"甲", "丙", "戊"},   // 寅
	{"乙", "", ""},       // 卯
	{"戊", "乙", "癸"},   // 辰
	{"丙", "庚", "戊"},   // 巳
	{"丁", "己", ""},     // 午
	{"己", "丁", "乙"},   // 未
	{"庚", "壬", "戊"},   // 申
	{"辛", "", ""},       // 酉
	{"戊", "辛", "丁"},   // 戌
	{"壬", "甲", ""},     // 亥
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
	{"无恩之刑", []string{"寅", "巳", "申"}, "寅刑巳，巳刑申，申刑寅，无恩之刑。主忘恩负义，恩将仇报。"},
	{"恃势之刑", []string{"丑", "戌", "未"}, "丑刑戌，戌刑未，未刑丑，恃势之刑。主仗势欺人，刑伤灾祸。"},
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

// ── 十神含义描述 ──────────────────────────────────────────
var tenGodDescriptions = map[string]string{
	"比肩": "同类相助，主竞争、合作、义气。身弱得助，身旺争财。",
	"劫财": "同类异性，主争夺、破耗、义气。身弱有帮，身旺破财。",
	"食神": "我生同性，主才艺、口福、表达。泄秀有情，利创作。",
	"伤官": "我生异性，主叛逆、创新、口舌。泄秀有力，防是非。",
	"正财": "我克异性，主正当收入、妻财。务实求财，稳定收获。",
	"偏财": "我克同性，主意外之财、投机。横财机遇，但不稳定。",
	"正官": "克我异性，主权威、责任、约束。事业正途，守规矩。",
	"七杀": "克我同性，主压力、竞争、果断。猛烈克身，须有制化。",
	"正印": "生我异性，主学问、庇护、贵人。文星高照，得长辈助。",
	"偏印": "生我同性，主偏门学问、孤僻。枭神夺食，防暗损。",
}

// ── 核心计算函数 ──────────────────────────────────────────

// CalcRikuyo 日课推算主入口
func CalcRikuyo(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) *RikuyoResult {
	dayGan := bazi.DayPillar.Gan

	// 获取有效喜忌（格局优先，扶抑兜底）
	like, _, isSpecialPattern := getEffectiveFavor(bazi)

	// 获取今日干支
	qYear, qMonth, qDay := queryDate.Year(), int(queryDate.Month()), queryDate.Day()
	ec, err := getDayEightChar(qYear, qMonth, qDay)
	if err != nil {
		return &RikuyoResult{OverallVerdict: "无法获取今日干支", FavorScore: 0}
	}
	todayGan := ec.GetDay().GetHeavenStem().GetName()
	todayZhi := ec.GetDay().GetEarthBranch().GetName()
	todayMonth := qMonth

	// 步骤一：今日天干取十神（格局感知）
	tenGod := bazipkg.ClassifyTenGod(todayGan, dayGan, false)
	tenGodFavorable := isFavorableTenGodByFavor(tenGod, like, dayGan)

	// 步骤二：今日地支取十二长生
	stage, stageFav, flexible := calcTwelveStage(dayGan, todayZhi)

	// 步骤三：地支藏干分析（格局感知）
	hiddenStems := calcHiddenStemGods(todayZhi, dayGan, like, dayGan)

	// 步骤四：天干关系（今日天干 vs 命局四柱天干）
	stemRels := calcStemRelations(todayGan, bazi)

	// 步骤五：地支关系（今日地支 vs 命局四柱地支）
	branchRels := calcBranchRelations(todayZhi, bazi)

	// 步骤六：神煞引动
	shenSha := calcShenShaActivation(todayGan, todayZhi, bazi)

	// 步骤七：大运叠加
	daYun := calcDaYunInfluence(bazi, queryDate, birthYear)

	// 步骤八：流年叠加
	liuNian := calcLiuNianInfluence(bazi, queryDate)

	// 步骤九：进退气分析
	advanceRetreat := calcAdvanceRetreat(todayGan, todayZhi, todayMonth)

	// 步骤十：用神综合判断
	yongShen := calcYongShenIntegration(bazi, todayGan, todayZhi)

	// 步骤十一：综合评分与断语
	verdict, score := calcOverallVerdict(
		tenGod, tenGodFavorable, stage, stageFav,
		hiddenStems, stemRels, branchRels,
		shenSha, daYun, liuNian,
		advanceRetreat, yongShen, isSpecialPattern, bazi,
	)

	return &RikuyoResult{
		TodayTenGod:      tenGod,
		TenGodFavorable:  tenGodFavorable,
		TenGodDesc:       tenGodDescriptions[tenGod],
		TwelveStage:      stage,
		StageFavorable:   stageFav,
		StageDesc:        stageDescriptions[stage],
		StageFlexible:    flexible,
		HiddenStems:      hiddenStems,
		StemRelations:    stemRels,
		BranchRelations:  branchRels,
		ActivatedShenSha: shenSha,
		DaYunInfluence:   daYun,
		LiuNianInfluence: liuNian,
		AdvanceRetreat:   advanceRetreat,
		YongShenImpact:   yongShen,
		OverallVerdict:   verdict,
		FavorScore:       score,
		PatternName:         bazi.PatternAnalysis.PatternName,
		PatternType:         bazi.PatternAnalysis.PatternType,
		PatternFavorable:    bazi.PatternAnalysis.FavorableElements,
		PatternUnfavorable:  bazi.PatternAnalysis.UnfavorableElements,
	}
}

// ── 步骤一：十神喜忌判断 ──────────────────────────────────

func isFavorableTenGodByStrength(tenGod string, isStrong bool) bool {
	favorableForStrong := map[string]bool{
		"正官": true, "七杀": true, "食神": true, "伤官": true,
		"正财": true, "偏财": true,
	}
	favorableForWeak := map[string]bool{
		"正印": true, "偏印": true, "比肩": true, "劫财": true,
	}
	if isStrong {
		return favorableForStrong[tenGod]
	}
	return favorableForWeak[tenGod]
}

// getEffectiveFavor 获取有效喜用五行（格局优先，扶抑兜底）
// 特殊格局（从格/专旺格）的喜忌完全覆盖普通扶抑喜忌
func getEffectiveFavor(bazi *bazipkg.BaziResult) (like, dislike []string, isSpecialPattern bool) {
	pa := bazi.PatternAnalysis
	if pa.PatternType == "特殊格局" && len(pa.FavorableElements) > 0 {
		return pa.FavorableElements, pa.UnfavorableElements, true
	}
	return bazi.BodyStrength.Like, bazi.BodyStrength.Dislike, false
}

// isFavorableElement 判断某五行是否在喜用列表中
func isFavorableElement(elem string, like []string) bool {
	for _, e := range like {
		if e == elem {
			return true
		}
	}
	return false
}

// isFavorableTenGodByFavor 格局感知的十神喜忌判断
// 通过十神对应的五行关系，判断该十神是否为喜用
func isFavorableTenGodByFavor(tenGod string, like []string, dayGan string) bool {
	dayElem := data.GanElement[dayGan]
	if dayElem == "" {
		return false
	}
	// 十神 → 对应五行关系 → 具体五行
	var targetElem string
	switch tenGod {
	case "正印", "偏印":
		targetElem = shengWo(dayElem) // 生我
	case "比肩", "劫财":
		targetElem = dayElem // 同我
	case "食神", "伤官":
		targetElem = elementGeneratesMap[dayElem] // 我生
	case "正财", "偏财":
		targetElem = elementOvercomesMap[dayElem] // 我克
	case "正官", "七杀":
		targetElem = keWoReverse(dayElem) // 克我
	}
	return isFavorableElement(targetElem, like)
}

// keWoReverse 返回克我的五行（与 keWo 逻辑一致，但在此独立定义避免跨包依赖）
func keWoReverse(elem string) string {
	return map[string]string{"木": "金", "火": "水", "土": "木", "金": "火", "水": "土"}[elem]
}

// ── 步骤二：十二长生（含活法修正） ──────────────────────────

func calcTwelveStage(dayGan, branch string) (string, bool, string) {
	ganIdx := data.GanIndex(dayGan)
	zhiIdx := data.ZhiIndex(branch)
	if ganIdx < 0 || zhiIdx < 0 {
		return "未知", false, ""
	}

	stage := twelveStageTable[ganIdx][zhiIdx]
	fav := stageFavorability[stage]
	flexible := ""

	// 活法修正：凶位检查藏干是否有救
	if fav == "凶" || fav == "半凶" {
		rescue := checkFlexibleRescue(dayGan, branch)
		if rescue != "" {
			flexible = rescue
		}
	}

	stageFav := fav == "吉"
	return stage, stageFav, flexible
}

// checkFlexibleRescue 检查活法是否有救
// 例：丙火绝于亥，亥中藏甲木（偏印），木能生火 → "虽绝有救：亥中甲木生丙火"
func checkFlexibleRescue(dayGan, branch string) string {
	dayElem := data.GanElement[dayGan]
	zhiIdx := data.ZhiIndex(branch)
	if zhiIdx < 0 {
		return ""
	}

	stems := hiddenStemMap[zhiIdx]
	for _, s := range stems {
		if s == "" {
			continue
		}
		stemElem := data.GanElement[s]
		// 检查藏干五行是否生扶日主五行
		if elementGeneratesMap[stemElem] == dayElem {
			god := bazipkg.ClassifyTenGod(s, dayGan, false)
			return fmt.Sprintf("虽%s有救：%s中%s(%s)生%s", twelveStageTable[data.GanIndex(dayGan)][zhiIdx], branch, s, god, dayGan)
		}
		// 检查藏干五行是否与日主同五行（比劫帮身）
		if stemElem == dayElem {
			god := bazipkg.ClassifyTenGod(s, dayGan, false)
			return fmt.Sprintf("虽%s有根：%s中%s(%s)帮身", twelveStageTable[data.GanIndex(dayGan)][zhiIdx], branch, s, god)
		}
	}
	return ""
}

// ── 步骤三：藏干十神分析 ──────────────────────────────────

func calcHiddenStemGods(branch, dayGan string, like []string, dayGanForFavor string) []HiddenStemGod {
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
		god := bazipkg.ClassifyTenGod(s, dayGan, false)
		fav := isFavorableTenGodByFavor(god, like, dayGanForFavor)
		result = append(result, HiddenStemGod{
			Stem:      s,
			Type:      types[i],
			Element:   data.GanElement[s],
			TenGod:    god,
			Favorable: fav,
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
		if bg.name == bazi.DayPillar.Gan {
			continue // 跳过日干自身
		}

		// 天干五合
		if stemCombineMap[todayGan] == bg.name {
			ce := combineElement[todayGan+bg.name]
			rels = append(rels, StemRelation{
				Type:        "五合",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s%s合化%s", todayGan, bg.name, ce),
				IsFavorable: true,
				Note:        checkTanHeWangKe(todayGan, bg.name, bazi),
			})
			continue
		}

		// 天干相冲
		if stemClashMap[todayGan] == bg.name {
			rels = append(rels, StemRelation{
				Type:        "相冲",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s冲%s", todayGan, bg.name),
				IsFavorable: false,
			})
			continue
		}

		// 天干相克
		todayElem := data.GanElement[todayGan]
		birthElem := data.GanElement[bg.name]
		if elementOvercomesMap[todayElem] == birthElem {
			rels = append(rels, StemRelation{
				Type:        "相克",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s克%s", todayGan, bg.name),
				IsFavorable: false,
			})
		} else if elementOvercomesMap[birthElem] == todayElem {
			rels = append(rels, StemRelation{
				Type:        "被克",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s克%s", bg.name, todayGan),
				IsFavorable: false,
			})
		} else if elementGeneratesMap[todayElem] == birthElem {
			rels = append(rels, StemRelation{
				Type:        "相生",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s生%s", todayGan, bg.name),
				IsFavorable: true,
			})
		} else if elementGeneratesMap[birthElem] == todayElem {
			rels = append(rels, StemRelation{
				Type:        "被生",
				Target:      bg.name,
				Detail:      fmt.Sprintf("%s生%s", bg.name, todayGan),
				IsFavorable: true,
			})
		}
	}
	return rels
}

// checkTanHeWangKe 检查"贪合忘克"
// 若今日天干与命局某天干相合，且该天干原为克日主的忌神，则合住后该忌神失效
func checkTanHeWangKe(todayGan, combinedGan string, bazi *bazipkg.BaziResult) string {
	dayGan := bazi.DayPillar.Gan
	dayElem := data.GanElement[dayGan]
	combinedElem := data.GanElement[combinedGan]

	// 检查被合住的天干是否克日主
	if elementOvercomesMap[combinedElem] == dayElem {
		return fmt.Sprintf("贪合忘克：%s本克%s，但%s合住%s，克力化解", combinedGan, dayGan, todayGan, combinedGan)
	}
	return ""
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
		// 六合
		if branchSixCombine[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六合",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六合", todayZhi, bz.name),
				IsFavorable: true,
			})
		}
		// 六冲
		if branchSixClash[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六冲",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六冲", todayZhi, bz.name),
				IsFavorable: false,
			})
		}
		// 六害
		if branchSixHarm[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六害",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六害", todayZhi, bz.name),
				IsFavorable: false,
			})
		}
	}

	// 三刑检测
	for _, pun := range branchThreePunishment {
		todayInGroup := contains(pun.Group, todayZhi)
		if !todayInGroup {
			continue
		}
		for _, bz := range birthZhis {
			if contains(pun.Group, bz.name) && bz.name != todayZhi {
				rels = append(rels, BranchRelation{
					Type:        "三刑",
					Target:      bz.name,
					Detail:      fmt.Sprintf("%s(%s)", pun.Name, pun.Desc),
					IsFavorable: false,
				})
			}
		}
	}

	// 三合检测
	for _, he := range branchThreeCombine {
		todayInGroup := contains(he.Group, todayZhi)
		if !todayInGroup {
			continue
		}
		for _, bz := range birthZhis {
			if contains(he.Group, bz.name) && bz.name != todayZhi {
				rels = append(rels, BranchRelation{
					Type:        "三合",
					Target:      bz.name,
					Detail:      fmt.Sprintf("%s%s三合%s局", todayZhi, bz.name, he.Element),
					IsFavorable: true,
				})
			}
		}
	}

	// 三会检测
	for _, hui := range branchThreeMeeting {
		todayInGroup := contains(hui.Group, todayZhi)
		if !todayInGroup {
			continue
		}
		for _, bz := range birthZhis {
			if contains(hui.Group, bz.name) && bz.name != todayZhi {
				rels = append(rels, BranchRelation{
					Type:        "三会",
					Target:      bz.name,
					Detail:      fmt.Sprintf("%s%s三会%s方", todayZhi, bz.name, hui.Element),
					IsFavorable: true,
				})
			}
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
	var result []ShenShaActivation

	// 天乙贵人
	if guas, ok := tianYiGuiren[dayGan]; ok {
		for _, g := range guas {
			if todayZhi == g {
				result = append(result, ShenShaActivation{
					Name:        "天乙贵人",
					Type:        "吉神",
					Description: "命中最尊贵之神，所至之处凶煞隐避。主当日有贵人相助，逢凶化吉。",
					Activation:  fmt.Sprintf("今日地支%s为%s的天乙贵人位", todayZhi, dayGan),
				})
			}
		}
	}

	// 驿马
	if yiMa[dayZhi] == todayZhi {
		result = append(result, ShenShaActivation{
			Name:        "驿马",
			Type:        "吉神",
			Description: "主出行变动、消息到来。当日有出行、变动之象。",
			Activation:  fmt.Sprintf("今日地支%s为日支%s的驿马位", todayZhi, dayZhi),
		})
	}

	// 桃花
	if taoHua[dayZhi] == todayZhi {
		result = append(result, ShenShaActivation{
			Name:        "桃花",
			Type:        "吉神",
			Description: "主异性缘、人缘佳。当日社交活跃，人缘旺盛。",
			Activation:  fmt.Sprintf("今日地支%s为日支%s的桃花位", todayZhi, dayZhi),
		})
	}

	// 禄神
	if luShen[dayGan] == todayZhi {
		result = append(result, ShenShaActivation{
			Name:        "禄神",
			Type:        "吉神",
			Description: "主俸禄、财禄。当日得禄，事业财运有助力。",
			Activation:  fmt.Sprintf("今日地支%s为%s的禄神位", todayZhi, dayGan),
		})
	}

	return result
}

// ── 步骤七：大运叠加 ──────────────────────────────────────

func calcDaYunInfluence(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) DaYunInfluence {
	if len(bazi.DaYunInfo.Pillars) == 0 {
		return DaYunInfluence{}
	}

	age := queryDate.Year() - birthYear
	startAge := bazi.DaYunInfo.StartAge

	// 确定当前大运索引
	idx := 0
	if startAge > 0 {
		idx = (age - startAge) / 10
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(bazi.DaYunInfo.Pillars) {
		idx = len(bazi.DaYunInfo.Pillars) - 1
	}

	pillar := bazi.DaYunInfo.Pillars[idx]
	daYunGan := pillar.Gan
	daYunZhi := pillar.Zhi
	dayGan := bazi.DayPillar.Gan

	// 大运天干对日主的十神（格局感知）
	tenGod := bazipkg.ClassifyTenGod(daYunGan, dayGan, false)
	like, _, _ := getEffectiveFavor(bazi)
	fav := isFavorableTenGodByFavor(tenGod, like, dayGan)

	// 分析今日干支与大运的关系
	score := 0
	desc := ""

	// 大运天干与今日天干的关系
	todayElem := data.GanElement[queryDateGan(queryDate)]
	daYunElem := data.GanElement[daYunGan]
	if elementGeneratesMap[todayElem] == daYunElem || elementGeneratesMap[daYunElem] == todayElem {
		score += 5
		desc += "今日与大运相生。"
	}
	if elementOvercomesMap[todayElem] == daYunElem {
		score -= 5
		desc += "今日克大运。"
	}

	return DaYunInfluence{
		CurrentPillar: daYunGan + daYunZhi,
		StartAge:      startAge + idx*10,
		EndAge:        startAge + (idx+1)*10 - 1,
		TenGod:        tenGod,
		Favorable:     fav,
		Relation:      desc,
		Score:         score,
		Description:   fmt.Sprintf("当前行%s运（%s），%s对日主为%s，%s", daYunGan+daYunZhi, tenGod, daYunGan, tenGod, boolToFavStr(fav)),
	}
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
	yearGanZhi := getYearGanZhi(queryDate.Year())
	if len(yearGanZhi) < 2 {
		return LiuNianInfluence{}
	}
	runes := []rune(yearGanZhi)
	yearGan := string(runes[0])
	yearZhi := string(runes[1])
	dayGan := bazi.DayPillar.Gan

	// 流年天干对日主的十神（格局感知）
	tenGod := bazipkg.ClassifyTenGod(yearGan, dayGan, false)
	like, _, _ := getEffectiveFavor(bazi)
	fav := isFavorableTenGodByFavor(tenGod, like, dayGan)

	// 岁伤日干 vs 日犯岁君
	taiSuiRel := ""
	todayGan := queryDateGan(queryDate)
	todayElem := data.GanElement[todayGan]
	yearElem := data.GanElement[yearGan]
	score := 0

	if elementOvercomesMap[yearElem] == todayElem {
		taiSuiRel = "岁伤日干：流年天干克今日天干，有祸必轻"
		score -= 3
	} else if elementOvercomesMap[todayElem] == yearElem {
		taiSuiRel = "日犯岁君：今日天干克流年天干，灾殃必重"
		score -= 8
	}

	return LiuNianInfluence{
		YearPillar:     yearGan + yearZhi,
		TenGod:         tenGod,
		Favorable:      fav,
		TaiSuiRelation: taiSuiRel,
		Score:          score,
		Description:    fmt.Sprintf("流年%s，%s对日主为%s，%s", yearGan+yearZhi, yearGan, tenGod, boolToFavStr(fav)),
	}
}

// getYearGanZhi 获取年干支（简化版，以立春为界）
func getYearGanZhi(year int) string {
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	zhis := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	ganIdx := (year - 4) % 10
	zhiIdx := (year - 4) % 12
	return gans[ganIdx] + zhis[zhiIdx]
}

// ── 步骤九：进退气分析 ──────────────────────────────────────

func calcAdvanceRetreat(todayGan, todayZhi string, month int) AdvanceRetreat {
	elem := data.GanElement[todayGan]
	season := monthToSeason(month)
	state := elementSeasonState[season][elem]

	phase := ""
	phaseDesc := ""
	score := 0

	switch state {
	case "相":
		phase = "进气"
		phaseDesc = "将来者进，朝阳初升。虽表面不显，实则后劲十足。"
		score = 8
	case "旺":
		phase = "当令"
		phaseDesc = "进而当令，日中天。表面最强，但盛极必衰。"
		score = 6
	case "休":
		phase = "退气"
		phaseDesc = "功成者退，夕阳西下。渐弱之势，力不从心。"
		score = -4
	case "囚":
		phase = "无气"
		phaseDesc = "退而无气，暗夜沉寂。最弱之时，宜守不宜进。"
		score = -8
	case "死":
		phase = "绝灭"
		phaseDesc = "气已绝灭。极弱之时，但死中藏生。"
		score = -10
	default:
		phase = "未知"
		phaseDesc = "状态不明。"
		score = 0
	}

	return AdvanceRetreat{
		Phase:       phase,
		PhaseDesc:   phaseDesc,
		Element:     elem,
		Score:       score,
		Description: fmt.Sprintf("%s在%s为%s（%s）。%s", elem, season, state, phase, phaseDesc),
	}
}

func monthToSeason(m int) string {
	switch m {
	case 2, 3, 4:
		return "春"
	case 5, 6, 7:
		return "夏"
	case 8, 9, 10:
		return "秋"
	default:
		return "冬"
	}
}

// ── 步骤十：用神综合判断 ──────────────────────────────────

func calcYongShenIntegration(bazi *bazipkg.BaziResult, todayGan, todayZhi string) YongShenImpact {
	todayElem := data.GanElement[todayGan]
	todayZhiElem := data.ZhiElement[todayZhi]
	score := 0
	desc := ""

	// 扶抑用神
	fuYiHit := false
	for _, e := range bazi.BodyStrength.Like {
		if todayElem == e || todayZhiElem == e {
			fuYiHit = true
			break
		}
	}
	if fuYiHit {
		score += 10
		desc += "今日干支生扶扶抑用神。"
	}

	// 调候用神（从命局获取）
	tiaoHouElem := ""
	if bazi.Tiaohou != nil && bazi.Tiaohou.Primary != "" {
		tiaoHouElem = data.GanElement[bazi.Tiaohou.Primary]
		if tiaoHouElem == "" {
			tiaoHouElem = bazi.Tiaohou.Primary
		}
	}
	tiaoHouHit := false
	if tiaoHouElem != "" {
		if todayElem == tiaoHouElem || todayZhiElem == tiaoHouElem {
			tiaoHouHit = true
			score += 15 // 调候优先级最高
			desc += "今日干支生扶调候用神，大吉。"
		}
	}

	// 通关用神
	tongGuanElem := bazi.TongGuan.TongGuanElement
	tongGuanHit := false
	if tongGuanElem != "" && bazi.TongGuan.HasTongGuan {
		if todayElem == tongGuanElem || todayZhiElem == tongGuanElem {
			tongGuanHit = true
			score += 10
			desc += "今日干支生扶通关用神。"
		}
	}

	return YongShenImpact{
		TiaoHouElement:  tiaoHouElem,
		TiaoHouHit:      tiaoHouHit,
		TongGuanElement: tongGuanElem,
		TongGuanHit:     tongGuanHit,
		FuYiElements:    bazi.BodyStrength.Like,
		FuYiHit:         fuYiHit,
		Score:           score,
		Description:     desc,
	}
}

// ── 步骤十一：综合评分与断语 ──────────────────────────────

func calcOverallVerdict(
	tenGod string, tenGodFav bool,
	stage string, stageFav bool,
	hiddenStems []HiddenStemGod,
	stemRels []StemRelation,
	branchRels []BranchRelation,
	shenSha []ShenShaActivation,
	daYun DaYunInfluence,
	liuNian LiuNianInfluence,
	advanceRetreat AdvanceRetreat,
	yongShen YongShenImpact,
	isSpecialPattern bool,
	bazi *bazipkg.BaziResult,
) (string, int) {

	score := 50 // 基准分

	// 天干十神喜忌 (权重25%)
	if tenGodFav {
		score += 12
	} else {
		score -= 12
	}

	// 地支长生状态 (权重15%)
	if stageFav {
		score += 8
	} else {
		score -= 8
	}

	// 藏干分析 (权重15%)
	for _, hs := range hiddenStems {
		if hs.Favorable {
			score += 3
		} else {
			score -= 3
		}
	}

	// 地支关系 (权重15%)
	for _, br := range branchRels {
		if br.IsFavorable {
			score += 4
		} else {
			score -= 4
		}
	}

	// 天干关系（含贪合忘克）
	for _, sr := range stemRels {
		if sr.Note != "" {
			score += 5 // 贪合忘克加分
		}
		if sr.IsFavorable {
			score += 3
		} else {
			score -= 3
		}
	}

	// 大运叠加 (权重15%)
	score += daYun.Score

	// 流年叠加 (权重10%)
	score += liuNian.Score

	// 进退气
	score += advanceRetreat.Score

	// 用神综合
	score += yongShen.Score

	// 神煞引动 (权重5%)
	for _, ss := range shenSha {
		if ss.Type == "吉神" {
			score += 2
		} else {
			score -= 2
		}
	}

	// 限制范围
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	// 生成综合断语
	verdict := generateVerdict(tenGod, tenGodFav, stage, stageFav, advanceRetreat.Phase, yongShen, score)

	// 特殊格局标注
	if isSpecialPattern && bazi.PatternAnalysis.PatternName != "" {
		verdict = fmt.Sprintf("【%s】喜%s忌%s。%s",
			bazi.PatternAnalysis.PatternName,
			formatElements(bazi.PatternAnalysis.FavorableElements, ""),
			formatElements(bazi.PatternAnalysis.UnfavorableElements, ""),
			verdict)
	}

	return verdict, score
}

// formatElements 格式化五行列表
func formatElements(elems []string, prefix string) string {
	if len(elems) == 0 {
		return ""
	}
	s := ""
	for _, e := range elems {
		if s != "" {
			s += ""
		}
		s += e
	}
	return s
}

func generateVerdict(tenGod string, tenGodFav bool, stage string, stageFav bool, phase string, yongShen YongShenImpact, score int) string {
	v := ""

	// 十神部分
	if tenGodFav {
		v += fmt.Sprintf("今日%s透出，喜用神当值，", tenGod)
	} else {
		v += fmt.Sprintf("今日%s透出，忌神当权，", tenGod)
	}

	// 长生部分
	if stageFav {
		v += fmt.Sprintf("日主%s得力。", stage)
	} else {
		v += fmt.Sprintf("日主%s失力。", stage)
	}

	// 进退气
	if phase == "进气" {
		v += "五行进气，后劲十足。"
	} else if phase == "退气" {
		v += "五行退气，力不从心。"
	}

	// 用神
	if yongShen.TiaoHouHit {
		v += "调候用神得力，大吉之兆。"
	}

	// 总评
	if score >= 75 {
		v += "综合判断：今日运势上佳，宜把握机遇。"
	} else if score >= 60 {
		v += "综合判断：今日运势尚可，顺势而为。"
	} else if score >= 40 {
		v += "综合判断：今日运势平平，谨慎行事。"
	} else {
		v += "综合判断：今日运势欠佳，宜守不宜进。"
	}

	return v
}

func boolToFavStr(b bool) string {
	if b {
		return "有利"
	}
	return "不利"
}
