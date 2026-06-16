package fortune

import (
	"fmt"
	"time"

	"bazi/internal/model"
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
	PatternName        string   `json:"pattern_name"`
	PatternType        string   `json:"pattern_type"`
	PatternFavorable   []string `json:"pattern_favorable"`
	PatternUnfavorable []string `json:"pattern_unfavorable"`

	// 建除十二神
	JianChuName  string `json:"jian_chu_name"`
	JianChuFavor string `json:"jian_chu_favor"`
	JianChuDesc  string `json:"jian_chu_desc"`

	// 彭祖百忌
	PengzuGanTaboo string `json:"pengzu_gan_taboo"`
	PengzuZhiTaboo string `json:"pengzu_zhi_taboo"`

	// 黄道黑道日
	HuangDaoName      string `json:"huang_dao_name"`
	HuangDaoFavorable bool   `json:"huang_dao_favorable"`
	HuangDaoDesc      string `json:"huang_dao_desc"`

	// 日课综合建议
	DayClassSummary string `json:"day_class_summary"`
	DayClassAdvice  string `json:"day_class_advice"`
}

type HiddenStemGod = model.HiddenStemGod
type StemRelation = model.StemRelation
type BranchRelation = model.BranchRelation
type ShenShaActivation = model.ShenShaActivation
type DaYunInfluence = model.DaYunInfluence
type LiuNianInfluence = model.LiuNianInfluence
type AdvanceRetreat = model.AdvanceRetreat
type YongShenImpact = model.YongShenImpact

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
	"衰":  "气始退，如人过壮年。主退败渐衰，力不从心。",
	"病":  "气已弱，如人患病。主困顿多事，诸事不顺。",
	"死":  "气已绝，如人已死。主丧祸大凶，宜守不宜进。",
	"墓":  "气收藏，如入墓库。主停滞阻滞，蓄势待发。",
	"绝":  "气完全消失，如断根之木。主断绝极凶，但绝处逢生。",
	"胎":  "气重新凝聚，如胎儿孕育。主孕育潜伏，韬光养晦。",
	"养":  "气渐充盈，如婴儿哺育。主滋养蓄势，蓄积力量。",
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

// ── 黄道黑道日（协纪辨方书） ──────────────────────────────────
// 黄道六神：青龙、明堂、金匮、天德、玉堂、司命 → 吉
// 黑道六神：天刑、朱雀、白虎、天牢、玄武、勾陈 → 凶
// 以月支起青龙，顺排十二建星
// 建除十二神：建、除、满、平、定、执、破、危、成、收、开、闭
var jianChuTable = map[string][]string{
	// 月支 → [建,除,满,平,定,执,破,危,成,收,开,闭] 对应日子地支
	"寅": {"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"},
	"卯": {"卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑", "寅"},
	"辰": {"辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑", "寅", "卯"},
	"巳": {"巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰"},
	"午": {"午", "未", "申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳"},
	"未": {"未", "申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午"},
	"申": {"申", "酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未"},
	"酉": {"酉", "戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申"},
	"戌": {"戌", "亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉"},
	"亥": {"亥", "子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌"},
	"子": {"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"},
	"丑": {"丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子"},
}

// jianChuFavorability 建除十二神吉凶
var jianChuFavorability = map[string]string{
	"建": "吉", "除": "吉", "满": "吉", "平": "平",
	"定": "吉", "执": "吉", "破": "凶", "危": "凶",
	"成": "吉", "收": "吉", "开": "吉", "闭": "凶",
}

// jianChuDesc 建除十二神描述（协纪辨方书）
var jianChuDesc = map[string]string{
	"建": "万物生育之始，宜出行、上任、拜谒",
	"除": "除旧布新，宜治病、解除、清洁",
	"满": "丰收圆满，宜祭祀、祈福、开张",
	"平": "平顺无奇，宜修造、安葬，诸事平常",
	"定": "安定稳固，宜签约、交易、安床",
	"执": "执持果断，宜捕捉、执法、施工",
	"破": "破败损耗，大事不宜，仅宜破屋坏垣",
	"危": "高处危险，宜祭祀安床，忌登高冒险",
	"成": "成就圆满，宜开业、嫁娶、签约",
	"收": "收藏收获，宜纳财、收债、入库",
	"开": "开通顺利，宜开业、出行、求医",
	"闭": "关闭不通，宜收藏静守，忌开业出行",
}

// ── 彭祖百忌日（协纪辨方书） ──────────────────────────────────
// 以天干取前半句，以地支取后半句
var pengzuGanTaboo = map[string]string{
	"甲": "甲不开仓财物耗散",
	"乙": "乙不栽植千株不长",
	"丙": "丙不修灶必见灾殃",
	"丁": "丁不剃头头必生疮",
	"戊": "戊不受田田主不祥",
	"己": "己不破券二比并亡",
	"庚": "庚不经络织机虚张",
	"辛": "辛不合酱主人不尝",
	"壬": "壬不汲水更难提防",
	"癸": "癸不词讼理弱敌强",
}

var pengzuZhiTaboo = map[string]string{
	"子": "子不问卜自惹祸殃",
	"丑": "丑不冠带主不还乡",
	"寅": "寅不祭祀神鬼不尝",
	"卯": "卯不穿井水泉不香",
	"辰": "辰不哭泣必主重丧",
	"巳": "巳不远行财物伏藏",
	"午": "午不苫盖屋主更张",
	"未": "未不服药毒气入肠",
	"申": "申不安床鬼祟入房",
	"酉": "酉不宴客醉坐颠狂",
	"戌": "戌不吃犬作怪上床",
	"亥": "亥不嫁娶不利新郎",
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
	like, dislike, isSpecialPattern := getEffectiveFavor(bazi)

	// 获取今日干支
	qYear, qMonth, qDay := queryDate.Year(), int(queryDate.Month()), queryDate.Day()
	ec, err := getDayEightChar(qYear, qMonth, qDay)
	if err != nil {
		return &RikuyoResult{OverallVerdict: "无法获取今日干支", FavorScore: 0}
	}
	todayGan := ec.GetDay().GetHeavenStem().GetName()
	todayZhi := ec.GetDay().GetEarthBranch().GetName()
	monthZhi := ec.GetMonth().GetEarthBranch().GetName()

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
	branchRels := calcBranchRelations(todayZhi, bazi, like, dislike)

	// 步骤六：神煞引动
	shenSha := calcShenShaActivation(todayGan, todayZhi, bazi)

	// 步骤七：大运叠加
	daYun := calcDaYunInfluence(bazi, queryDate, birthYear)

	// 步骤八：流年叠加
	liuNian := calcLiuNianInfluence(bazi, queryDate)

	// 步骤九：进退气分析
	advanceRetreat := calcAdvanceRetreat(todayGan, todayZhi, monthZhi)

	// 步骤十：用神综合判断
	yongShen := calcYongShenIntegration(bazi, todayGan, todayZhi)

	// 步骤十一：综合评分与断语
	verdict, score := calcOverallVerdict(
		tenGod, tenGodFavorable, stage, stageFav,
		hiddenStems, stemRels, branchRels,
		shenSha, daYun, liuNian,
		advanceRetreat, yongShen, isSpecialPattern, bazi,
	)

	// 步骤十二：建除十二神
	jianChuName, jianChuFav, jianChuD := calcJianChu(bazi.MonthPillar.Zhi, todayZhi)

	// 步骤十三：彭祖百忌
	pengzuGan := pengzuGanTaboo[todayGan]
	pengzuZhi := pengzuZhiTaboo[todayZhi]

	// 步骤十四：黄道黑道日
	huangDaoName, huangDaoFav, huangDaoDesc := calcHuangDaoHeiDao(monthZhi, todayZhi)

	// 步骤十五：日课综合判断
	dayClassSummary, dayClassAdvice := calcDayClassSummary(jianChuName, jianChuFav, huangDaoName, huangDaoFav, pengzuGan, pengzuZhi)

	return &RikuyoResult{
		TodayTenGod:        tenGod,
		TenGodFavorable:    tenGodFavorable,
		TenGodDesc:         tenGodDescriptions[tenGod],
		TwelveStage:        stage,
		StageFavorable:     stageFav,
		StageDesc:          stageDescriptions[stage],
		StageFlexible:      flexible,
		HiddenStems:        hiddenStems,
		StemRelations:      stemRels,
		BranchRelations:    branchRels,
		ActivatedShenSha:   shenSha,
		DaYunInfluence:     daYun,
		LiuNianInfluence:   liuNian,
		AdvanceRetreat:     advanceRetreat,
		YongShenImpact:     yongShen,
		OverallVerdict:     verdict,
		FavorScore:         score,
		PatternName:        bazi.PatternAnalysis.PatternName,
		PatternType:        bazi.PatternAnalysis.PatternType,
		PatternFavorable:   bazi.PatternAnalysis.FavorableElements,
		PatternUnfavorable: bazi.PatternAnalysis.UnfavorableElements,
		JianChuName:        jianChuName,
		JianChuFavor:       jianChuFav,
		JianChuDesc:        jianChuD,
		PengzuGanTaboo:     pengzuGan,
		PengzuZhiTaboo:     pengzuZhi,
		HuangDaoName:       huangDaoName,
		HuangDaoFavorable:  huangDaoFav,
		HuangDaoDesc:       huangDaoDesc,
		DayClassSummary:    dayClassSummary,
		DayClassAdvice:     dayClassAdvice,
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
		targetElem = bazipkg.ShengWo(dayElem) // 生我
	case "比肩", "劫财":
		targetElem = dayElem // 同我
	case "食神", "伤官":
		targetElem = elementGeneratesMap[dayElem] // 我生
	case "正财", "偏财":
		targetElem = elementOvercomesMap[dayElem] // 我克
	case "正官", "七杀":
		targetElem = bazipkg.KeWo(dayElem) // 克我
	}
	return isFavorableElement(targetElem, like)
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

func calcBranchRelations(todayZhi string, bazi *bazipkg.BaziResult, like, dislike []string) []BranchRelation {
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
		// 六合：合住忌神为吉（忌神被合住不为害），合住喜神为凶（喜神被合住失去作用）
		if branchSixCombine[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六合",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六合", todayZhi, bz.name),
				IsFavorable: !isFavorableElement(data.ZhiElement[bz.name], like),
			})
		}
		// 六冲：冲去忌神为吉（忌神被冲走），冲去喜神为凶（喜神被冲散）
		if branchSixClash[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六冲",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六冲", todayZhi, bz.name),
				IsFavorable: isFavorableElement(data.ZhiElement[bz.name], dislike),
			})
		}
		// 六害：害到忌神为吉，害到喜神为凶
		if branchSixHarm[todayZhi] == bz.name {
			rels = append(rels, BranchRelation{
				Type:        "六害",
				Target:      bz.name,
				Detail:      fmt.Sprintf("%s%s六害", todayZhi, bz.name),
				IsFavorable: !isFavorableElement(data.ZhiElement[bz.name], like),
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
				result = append(result, newShenShaActivation("天乙贵人", "吉神", "命中最尊贵之神，所至之处凶煞隐避。主当日有贵人相助，逢凶化吉。", fmt.Sprintf("今日地支%s为%s的天乙贵人位", todayZhi, dayGan)))
			}
		}
	}

	// 驿马
	if yiMa[dayZhi] == todayZhi {
		result = append(result, newShenShaActivation("驿马", "吉神", "主出行变动、消息到来。当日有出行、变动之象。", fmt.Sprintf("今日地支%s为日支%s的驿马位", todayZhi, dayZhi)))
	}

	// 桃花
	if taoHua[dayZhi] == todayZhi {
		result = append(result, newShenShaActivation("桃花", "吉神", "主异性缘、人缘佳。当日社交活跃，人缘旺盛。", fmt.Sprintf("今日地支%s为日支%s的桃花位", todayZhi, dayZhi)))
	}

	// 禄神
	if luShen[dayGan] == todayZhi {
		result = append(result, newShenShaActivation("禄神", "吉神", "主俸禄、财禄。当日得禄，事业财运有助力。", fmt.Sprintf("今日地支%s为%s的禄神位", todayZhi, dayGan)))
	}

	return result
}

func newShenShaActivation(name, typ, description, activation string) ShenShaActivation {
	meta := bazipkg.LookupShenShaMeta(name)
	return ShenShaActivation{
		Name:        name,
		Type:        typ,
		Category:    meta.Category,
		Polarity:    meta.Polarity,
		Priority:    meta.Priority,
		Description: description,
		Activation:  activation,
	}
}

// ── 步骤七：大运叠加 ──────────────────────────────────────

func calcDaYunInfluence(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) DaYunInfluence {
	if len(bazi.DaYunInfo.Pillars) == 0 {
		return DaYunInfluence{}
	}

	age := queryDate.Year() - birthYear
	startAge := bazi.DaYunInfo.StartAge

	// 确定当前大运索引（精确计算，处理起运年龄非整十年的情况）
	// 大运每步10年，从起运年龄开始：第0步 = [startAge, startAge+9]，第1步 = [startAge+10, startAge+19]...
	idx := 0
	if age >= startAge {
		idx = (age - startAge) / 10
	} else {
		idx = 0 // 未起运，取第一步大运
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

	// 计算当前大运精确的起止年龄和年份
	daYunStartAge := startAge + idx*10
	daYunEndAge := daYunStartAge + 9

	// 大运天干对日主的十神（格局感知）
	tenGod := bazipkg.ClassifyTenGod(daYunGan, dayGan, false)
	like, _, _ := getEffectiveFavor(bazi)
	fav := isFavorableTenGodByFavor(tenGod, like, dayGan)

	// 分析今日干支与大运的关系
	score := 0
	desc := ""

	// 大运天干与今日天干的关系（优先检查天干五合）
	todayGan := queryDateGan(queryDate)
	todayElem := data.GanElement[todayGan]
	daYunElem := data.GanElement[daYunGan]

	if stemCombineMap[todayGan] == daYunGan {
		// 天干五合：气机交融，正面关系
		score += 8
		desc += fmt.Sprintf("%s%s合化%s。", todayGan, daYunGan, combineElement[todayGan+daYunGan])
	} else if elementGeneratesMap[todayElem] == daYunElem || elementGeneratesMap[daYunElem] == todayElem {
		score += 5
		desc += "今日与大运相生。"
	} else if elementOvercomesMap[todayElem] == daYunElem {
		score -= 5
		desc += "今日克大运。"
	} else if elementOvercomesMap[daYunElem] == todayElem {
		score -= 3
		desc += "大运克今日。"
	}

	return DaYunInfluence{
		CurrentPillar: daYunGan + daYunZhi,
		StartAge:      daYunStartAge,
		EndAge:        daYunEndAge,
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
	yearGanZhi := getYearGanZhi(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
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

// getYearGanZhi 获取年干支（以立春为界）
// 立春前仍属上一年干支
func getYearGanZhi(year, month, day int) string {
	y := year
	// 立春约在2月4日，立春前用上一年干支
	if month < 2 || (month == 2 && day < 4) {
		y = year - 1
	}
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	zhis := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	ganIdx := (y - 4) % 10
	zhiIdx := (y - 4) % 12
	return gans[ganIdx] + zhis[zhiIdx]
}

// ── 步骤九：进退气分析 ──────────────────────────────────────

func calcAdvanceRetreat(todayGan, todayZhi string, monthZhi string) AdvanceRetreat {
	elem := data.GanElement[todayGan]
	season := monthZhiToSeason(monthZhi)
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
	default: // 亥, 子, 丑
		return "冬"
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

// ── 步骤十二：建除十二神 ──────────────────────────────────────

// calcJianChu 计算建除十二神
// 以月支为起点，今日地支与月支的距离决定建除值
func calcJianChu(monthZhi, todayZhi string) (name, favor, desc string) {
	table, ok := jianChuTable[monthZhi]
	if !ok {
		return "未知", "平", ""
	}
	zhiOrder := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	monthIdx, todayIdx := -1, -1
	for i, z := range zhiOrder {
		if z == monthZhi {
			monthIdx = i
		}
		if z == todayZhi {
			todayIdx = i
		}
	}
	if monthIdx < 0 || todayIdx < 0 {
		return "未知", "平", ""
	}
	// 月支本身为"建"日，顺推
	offset := (todayIdx - monthIdx + 12) % 12
	jianChu := table[offset]
	return jianChu, jianChuFavorability[jianChu], jianChuDesc[jianChu]
}

// ── 黄道黑道日（协纪辨方书） ──────────────────────────────────

// huangDaoTwelveGods 十二值神顺排顺序
var huangDaoTwelveGods = []string{
	"青龙", "明堂", "天刑", "朱雀", "金匮", "天德",
	"白虎", "玉堂", "天牢", "玄武", "司命", "勾陈",
}

// huangDaoFavorableMap 黄道六神（吉）vs 黑道六神（凶）
var huangDaoFavorableMap = map[string]bool{
	"青龙": true, "明堂": true, "金匮": true, "天德": true, "玉堂": true, "司命": true,
	"天刑": false, "朱雀": false, "白虎": false, "天牢": false, "玄武": false, "勾陈": false,
}

// huangDaoDescMap 十二值神描述
var huangDaoDescMap = map[string]string{
	"青龙": "黄道吉神，主喜庆、贵人、财利。所值之日，百事皆宜。",
	"明堂": "黄道吉神，主明达、公开、顺畅。利公开事务、文书契约。",
	"天刑": "黑道凶神，主刑伤、意外。忌诉讼、出行、手术。",
	"朱雀": "黑道凶神，主口舌、是非、文书纠纷。忌签约、诉讼。",
	"金匮": "黄道吉神，主财富、收藏、积蓄。利纳财、储蓄、收账。",
	"天德": "黄道吉神，主德行、化解、贵人。百事皆宜，逢凶化吉。",
	"白虎": "黑道凶神，主血光、伤灾、丧事。忌动刀、出行、见血。",
	"玉堂": "黄道吉神，主吉庆、文昌、学业。利考试、文书、求名。",
	"天牢": "黑道凶神，主囚禁、束缚、不自由。忌诉讼、出行。",
	"玄武": "黑道凶神，主盗窃、暗昧、欺诈。防失财、小人、被骗。",
	"司命": "黄道吉神，主寿命、健康、平安。利求医、祈福、安床。",
	"勾陈": "黑道凶神，主田土、牵连、拖沓。忌动土、搬家、签约。",
}

// huangDaoStartIdx 月支 → 青龙起始地支索引
// 子午月从寅起，丑未月从卯起，寅申月从辰起，卯酉月从巳起，辰戌月从午起，巳亥月从未起
var huangDaoStartIdx = map[string]int{
	"子": 2, "午": 2, // 寅(2)起青龙
	"丑": 3, "未": 3, // 卯(3)起青龙
	"寅": 4, "申": 4, // 辰(4)起青龙
	"卯": 5, "酉": 5, // 巳(5)起青龙
	"辰": 6, "戌": 6, // 午(6)起青龙
	"巳": 7, "亥": 7, // 未(7)起青龙
}

// calcHuangDaoHeiDao 计算黄道黑道日
// 以月支定青龙起点，顺排十二值神，看今日地支落在哪个神
func calcHuangDaoHeiDao(monthZhi, todayZhi string) (name string, favorable bool, desc string) {
	startIdx, ok := huangDaoStartIdx[monthZhi]
	if !ok {
		return "未知", false, ""
	}
	todayIdx := data.ZhiIndex(todayZhi)
	if todayIdx < 0 {
		return "未知", false, ""
	}
	offset := (todayIdx - startIdx + 12) % 12
	name = huangDaoTwelveGods[offset]
	favorable = huangDaoFavorableMap[name]
	desc = huangDaoDescMap[name]
	return
}

// ── 日课综合判断 ──────────────────────────────────────────

// calcDayClassSummary 综合建除十二神、黄道黑道日、彭祖百忌，给出日课综合建议
func calcDayClassSummary(jianChuName, jianChuFav, huangDaoName string, huangDaoFav bool, pengzuGan, pengzuZhi string) (summary, advice string) {
	score := 0
	var goodPoints, badPoints []string

	// 建除十二神
	switch jianChuFav {
	case "吉":
		score += 2
		goodPoints = append(goodPoints, "建除"+jianChuName+"为吉")
	case "凶":
		score -= 2
		badPoints = append(badPoints, "建除"+jianChuName+"为凶")
	}

	// 黄道黑道
	if huangDaoFav {
		score += 2
		goodPoints = append(goodPoints, huangDaoName+"为黄道吉神")
	} else {
		score -= 2
		badPoints = append(badPoints, huangDaoName+"为黑道凶神")
	}

	// 生成综合断语
	if score >= 3 {
		summary = "日课大吉：黄道" + huangDaoName + "、建除" + jianChuName + "皆吉，诸事皆宜。"
		advice = "今日日课吉利，宜行大事。忌：" + pengzuGan + "；" + pengzuZhi
	} else if score >= 1 {
		summary = "日课偏吉：黄道" + huangDaoName + "、建除" + jianChuName + "尚可，宜谨慎行事。"
		advice = "今日日课尚可，宜择吉而行。忌：" + pengzuGan + "；" + pengzuZhi
	} else if score >= -1 {
		summary = "日课平平：黄道" + huangDaoName + "、建除" + jianChuName + "参半，宜守不宜攻。"
		advice = "今日日课平平，宜谨慎行事。忌：" + pengzuGan + "；" + pengzuZhi
	} else {
		summary = "日课欠佳：黄道" + huangDaoName + "、建除" + jianChuName + "皆凶，大事不宜。"
		advice = "今日日课不利，宜守不宜进，大事不宜。忌：" + pengzuGan + "；" + pengzuZhi
	}

	return
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
