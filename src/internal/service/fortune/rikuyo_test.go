package fortune

import (
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

// ── 十二长生速查表验证 ──────────────────────────────────────────
// 来源：每日干支与用神的关系速查.md 表二

func TestTwelveStageTable(t *testing.T) {
	tests := []struct {
		dayGan   string
		branch   string
		expected string
	}{
		// 甲木：长生在亥, 沐浴子, 冠带丑, 临官寅, 帝旺卯, 衰辰, 病巳, 死午, 墓未, 绝申, 胎酉, 养戌
		{"甲", "亥", "长生"}, {"甲", "子", "沐浴"}, {"甲", "丑", "冠带"},
		{"甲", "寅", "临官"}, {"甲", "卯", "帝旺"}, {"甲", "辰", "衰"},
		{"甲", "巳", "病"}, {"甲", "午", "死"}, {"甲", "未", "墓"},
		{"甲", "申", "绝"}, {"甲", "酉", "胎"}, {"甲", "戌", "养"},
		// 乙木（阴干逆转）：长生在午
		{"乙", "午", "长生"}, {"乙", "巳", "沐浴"}, {"乙", "辰", "冠带"},
		{"乙", "卯", "临官"}, {"乙", "寅", "帝旺"}, {"乙", "丑", "衰"},
		{"乙", "子", "病"}, {"乙", "亥", "死"}, {"乙", "戌", "墓"},
		{"乙", "酉", "绝"}, {"乙", "申", "胎"}, {"乙", "未", "养"},
		// 丙火：长生在寅
		{"丙", "寅", "长生"}, {"丙", "卯", "沐浴"}, {"丙", "辰", "冠带"},
		{"丙", "巳", "临官"}, {"丙", "午", "帝旺"}, {"丙", "未", "衰"},
		{"丙", "申", "病"}, {"丙", "酉", "死"}, {"丙", "戌", "墓"},
		{"丙", "亥", "绝"}, {"丙", "子", "胎"}, {"丙", "丑", "养"},
		// 庚金：长生在巳
		{"庚", "巳", "长生"}, {"庚", "午", "沐浴"}, {"庚", "未", "冠带"},
		{"庚", "申", "临官"}, {"庚", "酉", "帝旺"}, {"庚", "戌", "衰"},
		{"庚", "亥", "病"}, {"庚", "子", "死"}, {"庚", "丑", "墓"},
		{"庚", "寅", "绝"}, {"庚", "卯", "胎"}, {"庚", "辰", "养"},
		// 壬水：长生在申
		{"壬", "申", "长生"}, {"壬", "酉", "沐浴"}, {"壬", "戌", "冠带"},
		{"壬", "亥", "临官"}, {"壬", "子", "帝旺"}, {"壬", "丑", "衰"},
		{"壬", "寅", "病"}, {"壬", "卯", "死"}, {"壬", "辰", "墓"},
		{"壬", "巳", "绝"}, {"壬", "午", "胎"}, {"壬", "未", "养"},
	}
	for _, tt := range tests {
		stage, _, _ := calcTwelveStage(tt.dayGan, tt.branch)
		if stage != tt.expected {
			t.Errorf("calcTwelveStage(%s, %s) = %s, want %s", tt.dayGan, tt.branch, stage, tt.expected)
		}
	}
}

// ── 十二长生吉凶判断验证 ──────────────────────────────────────────

func TestTwelveStageFavorability(t *testing.T) {
	tests := []struct {
		dayGan  string
		branch  string
		favWant bool
		desc    string
	}{
		// 吉位：长生/冠带/临官/帝旺
		{"甲", "亥", true, "甲木长生在亥=吉"},
		{"甲", "丑", true, "甲木冠带在丑=吉"},
		{"甲", "寅", true, "甲木临官在寅=吉"},
		{"甲", "卯", true, "甲木帝旺在卯=吉"},
		// 半吉：沐浴/胎/养
		{"甲", "子", false, "甲木沐浴在子=半吉→favorable=false"}, // 半吉返回false
		{"甲", "酉", false, "甲木胎在酉=半吉→favorable=false"},
		{"甲", "戌", false, "甲木养在戌=半吉→favorable=false"},
		// 凶位：衰/病/死/墓/绝
		{"甲", "辰", false, "甲木衰在辰=半凶"},
		{"甲", "巳", false, "甲木病在巳=凶"},
		{"甲", "午", false, "甲木死在午=凶"},
		{"甲", "未", false, "甲木墓在未=半凶"},
		{"甲", "申", false, "甲木绝在申=凶"},
	}
	for _, tt := range tests {
		_, fav, _ := calcTwelveStage(tt.dayGan, tt.branch)
		if fav != tt.favWant {
			t.Errorf("%s: got favorable=%v, want %v", tt.desc, fav, tt.favWant)
		}
	}
}

// ── 活法修正验证 ──────────────────────────────────────────
// 丙火绝于亥，亥中藏甲木（偏印），木能生火 → 有救

func TestFlexibleRescue(t *testing.T) {
	tests := []struct {
		dayGan    string
		branch    string
		wantRescue bool
		desc      string
	}{
		{"丙", "亥", true, "丙火绝于亥，亥藏甲木生火→有救"},
		{"庚", "丑", true, "庚金墓于丑，丑藏己土生金→有救"},
		{"壬", "巳", true, "壬水绝于巳，巳藏庚金生水→有救"},
		{"甲", "午", false, "甲木死于午，午藏丁己无木水→无救"},
	}
	for _, tt := range tests {
		rescue := checkFlexibleRescue(tt.dayGan, tt.branch)
		got := rescue != ""
		if got != tt.wantRescue {
			t.Errorf("%s: got rescue=%v (\"%s\"), want %v", tt.desc, got, rescue, tt.wantRescue)
		}
	}
}

// ── 十神分类验证 ──────────────────────────────────────────
// 来源：每日干支与用神的关系速查.md 表一

func TestTenGodClassification(t *testing.T) {
	tests := []struct {
		todayGan string
		dayGan   string
		expected string
	}{
		// 日主甲：甲比肩, 乙劫财, 丙食神, 丁伤官, 戊偏财, 己正财, 庚七杀, 辛正官, 壬偏印, 癸正印
		{"甲", "甲", "比肩"}, {"乙", "甲", "劫财"}, {"丙", "甲", "食神"},
		{"丁", "甲", "伤官"}, {"戊", "甲", "偏财"}, {"己", "甲", "正财"},
		{"庚", "甲", "七杀"}, {"辛", "甲", "正官"}, {"壬", "甲", "偏印"},
		{"癸", "甲", "正印"},
		// 日主丙：甲偏印, 乙正印, 丙比肩, 丁劫财, 戊食神, 己伤官, 庚偏财, 辛正财, 壬七杀, 癸正官
		{"甲", "丙", "偏印"}, {"乙", "丙", "正印"}, {"丙", "丙", "比肩"},
		{"丁", "丙", "劫财"}, {"戊", "丙", "食神"}, {"己", "丙", "伤官"},
		{"庚", "丙", "偏财"}, {"辛", "丙", "正财"}, {"壬", "丙", "七杀"},
		{"癸", "丙", "正官"},
		// 日主庚：甲偏财, 乙正财, 丙七杀, 丁正官, 戊偏印, 己正印, 庚比肩, 辛劫财, 壬食神, 癸伤官
		{"甲", "庚", "偏财"}, {"乙", "庚", "正财"}, {"丙", "庚", "七杀"},
		{"丁", "庚", "正官"}, {"戊", "庚", "偏印"}, {"己", "庚", "正印"},
		{"庚", "庚", "比肩"}, {"辛", "庚", "劫财"}, {"壬", "庚", "食神"},
		{"癸", "庚", "伤官"},
	}
	for _, tt := range tests {
		got := bazipkg.ClassifyTenGod(tt.todayGan, tt.dayGan, false)
		if got != tt.expected {
			t.Errorf("ClassifyTenGod(%s, %s) = %s, want %s", tt.todayGan, tt.dayGan, got, tt.expected)
		}
	}
}

// ── 十神喜忌判断验证 ──────────────────────────────────────────

func TestTenGodFavorability(t *testing.T) {
	tests := []struct {
		tenGod  string
		isStrong bool
		want    bool
		desc    string
	}{
		// 身弱喜生扶：正印/偏印/比肩/劫财为喜
		{"正印", false, true, "身弱正印=喜"},
		{"偏印", false, true, "身弱偏印=喜"},
		{"比肩", false, true, "身弱比肩=喜"},
		{"劫财", false, true, "身弱劫财=喜"},
		{"食神", false, false, "身弱食神=忌"},
		{"伤官", false, false, "身弱伤官=忌"},
		{"正财", false, false, "身弱正财=忌"},
		{"七杀", false, false, "身弱七杀=忌"},
		// 身旺喜克泄耗：正官/七杀/食神/伤官/正财/偏财为喜
		{"正官", true, true, "身旺正官=喜"},
		{"七杀", true, true, "身旺七杀=喜"},
		{"食神", true, true, "身旺食神=喜"},
		{"伤官", true, true, "身旺伤官=喜"},
		{"正财", true, true, "身旺正财=喜"},
		{"偏财", true, true, "身旺偏财=喜"},
		{"正印", true, false, "身旺正印=忌"},
		{"比肩", true, false, "身旺比肩=忌"},
	}
	for _, tt := range tests {
		got := isFavorableTenGodByStrength(tt.tenGod, tt.isStrong)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
		}
	}
}

// ── 地支藏干验证 ──────────────────────────────────────────

func TestHiddenStemMap(t *testing.T) {
	tests := []struct {
		branch   string
		benQi    string
		zhongQi  string
		yuQi     string
	}{
		{"子", "癸", "", ""},
		{"丑", "己", "癸", "辛"},
		{"寅", "甲", "丙", "戊"},
		{"卯", "乙", "", ""},
		{"辰", "戊", "乙", "癸"},
		{"巳", "丙", "庚", "戊"},
		{"午", "丁", "己", ""},
		{"未", "己", "丁", "乙"},
		{"申", "庚", "壬", "戊"},
		{"酉", "辛", "", ""},
		{"戌", "戊", "辛", "丁"},
		{"亥", "壬", "甲", ""},
	}
	for _, tt := range tests {
		idx := data.ZhiIndex(tt.branch)
		stems := hiddenStemMap[idx]
		if stems[0] != tt.benQi {
			t.Errorf("%s本气: got %s, want %s", tt.branch, stems[0], tt.benQi)
		}
		if stems[1] != tt.zhongQi {
			t.Errorf("%s中气: got %s, want %s", tt.branch, stems[1], tt.zhongQi)
		}
		if stems[2] != tt.yuQi {
			t.Errorf("%s余气: got %s, want %s", tt.branch, stems[2], tt.yuQi)
		}
	}
}

// ── 天干五合验证 ──────────────────────────────────────────

func TestStemCombine(t *testing.T) {
	tests := []struct {
		a, b   string
		element string
	}{
		{"甲", "己", "土"}, {"己", "甲", "土"},
		{"乙", "庚", "金"}, {"庚", "乙", "金"},
		{"丙", "辛", "水"}, {"辛", "丙", "水"},
		{"丁", "壬", "木"}, {"壬", "丁", "木"},
		{"戊", "癸", "火"}, {"癸", "戊", "火"},
	}
	for _, tt := range tests {
		if stemCombineMap[tt.a] != tt.b {
			t.Errorf("stemCombineMap[%s] = %s, want %s", tt.a, stemCombineMap[tt.a], tt.b)
		}
		ce := combineElement[tt.a+tt.b]
		if ce != tt.element {
			t.Errorf("combineElement[%s%s] = %s, want %s", tt.a, tt.b, ce, tt.element)
		}
	}
}

// ── 天干相冲验证 ──────────────────────────────────────────

func TestStemClash(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"甲", "庚"}, {"庚", "甲"},
		{"乙", "辛"}, {"辛", "乙"},
		{"丙", "壬"}, {"壬", "丙"},
		{"丁", "癸"}, {"癸", "丁"},
	}
	for _, tt := range tests {
		if stemClashMap[tt.a] != tt.b {
			t.Errorf("stemClashMap[%s] = %s, want %s", tt.a, stemClashMap[tt.a], tt.b)
		}
	}
}

// ── 地支六冲验证 ──────────────────────────────────────────

func TestBranchSixClash(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"子", "午"}, {"午", "子"},
		{"丑", "未"}, {"未", "丑"},
		{"寅", "申"}, {"申", "寅"},
		{"卯", "酉"}, {"酉", "卯"},
		{"辰", "戌"}, {"戌", "辰"},
		{"巳", "亥"}, {"亥", "巳"},
	}
	for _, tt := range tests {
		if branchSixClash[tt.a] != tt.b {
			t.Errorf("branchSixClash[%s] = %s, want %s", tt.a, branchSixClash[tt.a], tt.b)
		}
	}
}

// ── 地支六合验证 ──────────────────────────────────────────

func TestBranchSixCombine(t *testing.T) {
	tests := []struct {
		a, b string
	}{
		{"子", "丑"}, {"丑", "子"},
		{"寅", "亥"}, {"亥", "寅"},
		{"卯", "戌"}, {"戌", "卯"},
		{"辰", "酉"}, {"酉", "辰"},
		{"巳", "申"}, {"申", "巳"},
		{"午", "未"}, {"未", "午"},
	}
	for _, tt := range tests {
		if branchSixCombine[tt.a] != tt.b {
			t.Errorf("branchSixCombine[%s] = %s, want %s", tt.a, branchSixCombine[tt.a], tt.b)
		}
	}
}

// ── 天乙贵人验证 ──────────────────────────────────────────
// 口诀：甲戊庚牛羊，乙己鼠猴乡，丙丁猪鸡位，壬癸兔蛇藏，辛逢虎马乡

func TestTianYiGuiren(t *testing.T) {
	tests := []struct {
		dayGan string
		want   []string
	}{
		{"甲", []string{"丑", "未"}},
		{"戊", []string{"丑", "未"}},
		{"庚", []string{"丑", "未"}},
		{"乙", []string{"子", "申"}},
		{"己", []string{"子", "申"}},
		{"丙", []string{"亥", "酉"}},
		{"丁", []string{"亥", "酉"}},
		{"壬", []string{"卯", "巳"}},
		{"癸", []string{"卯", "巳"}},
		{"辛", []string{"寅", "午"}},
	}
	for _, tt := range tests {
		got := tianYiGuiren[tt.dayGan]
		if len(got) != len(tt.want) {
			t.Errorf("天乙贵人[%s] len=%d, want %d", tt.dayGan, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("天乙贵人[%s][%d] = %s, want %s", tt.dayGan, i, got[i], tt.want[i])
			}
		}
	}
}

// ── 驿马验证 ──────────────────────────────────────────
// 寅午戌见申，申子辰见寅，巳酉丑见亥，亥卯未见巳

func TestYiMa(t *testing.T) {
	tests := []struct {
		dayZhi string
		want   string
	}{
		{"寅", "申"}, {"午", "申"}, {"戌", "申"},
		{"申", "寅"}, {"子", "寅"}, {"辰", "寅"},
		{"巳", "亥"}, {"酉", "亥"}, {"丑", "亥"},
		{"亥", "巳"}, {"卯", "巳"}, {"未", "巳"},
	}
	for _, tt := range tests {
		if yiMa[tt.dayZhi] != tt.want {
			t.Errorf("驿马[%s] = %s, want %s", tt.dayZhi, yiMa[tt.dayZhi], tt.want)
		}
	}
}

// ── 桃花验证 ──────────────────────────────────────────

func TestTaoHua(t *testing.T) {
	tests := []struct {
		dayZhi string
		want   string
	}{
		{"寅", "卯"}, {"午", "卯"}, {"戌", "卯"},
		{"申", "酉"}, {"子", "酉"}, {"辰", "酉"},
		{"巳", "午"}, {"酉", "午"}, {"丑", "午"},
		{"亥", "子"}, {"卯", "子"}, {"未", "子"},
	}
	for _, tt := range tests {
		if taoHua[tt.dayZhi] != tt.want {
			t.Errorf("桃花[%s] = %s, want %s", tt.dayZhi, taoHua[tt.dayZhi], tt.want)
		}
	}
}

// ── 禄神验证 ──────────────────────────────────────────

func TestLuShen(t *testing.T) {
	tests := []struct {
		dayGan string
		want   string
	}{
		{"甲", "寅"}, {"乙", "卯"}, {"丙", "巳"}, {"丁", "午"},
		{"戊", "巳"}, {"己", "午"}, {"庚", "申"}, {"辛", "酉"},
		{"壬", "亥"}, {"癸", "子"},
	}
	for _, tt := range tests {
		if luShen[tt.dayGan] != tt.want {
			t.Errorf("禄神[%s] = %s, want %s", tt.dayGan, luShen[tt.dayGan], tt.want)
		}
	}
}

// ── 进退气验证 ──────────────────────────────────────────
// 春：木旺/火相/水休/金囚/土死

func TestAdvanceRetreat(t *testing.T) {
	tests := []struct {
		todayGan string
		month    int
		wantPhase string
		desc     string
	}{
		{"甲", 3, "当令", "甲木春天=旺=当令"},
		{"丙", 3, "进气", "丙火春天=相=进气"},
		{"壬", 3, "退气", "壬水春天=休=退气"},
		{"庚", 3, "无气", "庚金春天=囚=无气"},
		{"戊", 3, "绝灭", "戊土春天=死=绝灭"},
		{"庚", 8, "当令", "庚金秋天=旺=当令"},
		{"壬", 8, "进气", "壬水秋天=相=进气"},
		{"丙", 12, "绝灭", "丙火冬天=死=绝灭"},
	}
	for _, tt := range tests {
		result := calcAdvanceRetreat(tt.todayGan, "", tt.month)
		if result.Phase != tt.wantPhase {
			t.Errorf("%s: got phase=%s, want %s", tt.desc, result.Phase, tt.wantPhase)
		}
	}
}

// ── 年干支计算验证 ──────────────────────────────────────────

func TestGetYearGanZhi(t *testing.T) {
	tests := []struct {
		year int
		want string
	}{
		{2024, "甲辰"},
		{2025, "乙巳"},
		{2026, "丙午"},
		{1990, "庚午"},
		{2000, "庚辰"},
	}
	for _, tt := range tests {
		got := getYearGanZhi(tt.year)
		if got != tt.want {
			t.Errorf("getYearGanZhi(%d) = %s, want %s", tt.year, got, tt.want)
		}
	}
}

// ── 季节判断验证 ──────────────────────────────────────────

func TestMonthToSeason(t *testing.T) {
	tests := []struct {
		month int
		want  string
	}{
		{1, "冬"}, {2, "春"}, {3, "春"}, {4, "春"},
		{5, "夏"}, {6, "夏"}, {7, "夏"},
		{8, "秋"}, {9, "秋"}, {10, "秋"},
		{11, "冬"}, {12, "冬"},
	}
	for _, tt := range tests {
		got := monthToSeason(tt.month)
		if got != tt.want {
			t.Errorf("monthToSeason(%d) = %s, want %s", tt.month, got, tt.want)
		}
	}
}

// ── 藏干十神分析验证 ──────────────────────────────────────────

func TestCalcHiddenStemGods(t *testing.T) {
	// 寅中藏甲(本气)、丙(中气)、戊(余气)
	// 日主为庚金：甲=偏财, 丙=七杀, 戊=偏印
	stems := calcHiddenStemGods("寅", "庚", []string{"土", "水"}, "庚")
	if len(stems) != 3 {
		t.Fatalf("寅藏干数量: got %d, want 3", len(stems))
	}
	expected := []struct {
		stem   string
		typ    string
		god    string
	}{
		{"甲", "本气", "偏财"},
		{"丙", "中气", "七杀"},
		{"戊", "余气", "偏印"},
	}
	for i, exp := range expected {
		if stems[i].Stem != exp.stem {
			t.Errorf("寅藏干[%d] stem: got %s, want %s", i, stems[i].Stem, exp.stem)
		}
		if stems[i].Type != exp.typ {
			t.Errorf("寅藏干[%d] type: got %s, want %s", i, stems[i].Type, exp.typ)
		}
		if stems[i].TenGod != exp.god {
			t.Errorf("寅藏干[%d] ten_god: got %s, want %s", i, stems[i].TenGod, exp.god)
		}
	}
}

// ── 神煞引动验证 ──────────────────────────────────────────

func TestShenShaActivation(t *testing.T) {
	// 日主甲木，日支丑 → 天乙贵人在丑未
	// 今日地支丑 → 应引动天乙贵人
	bazi := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "丑"},
	}
	sha := calcShenShaActivation("丙", "丑", bazi)
	found := false
	for _, s := range sha {
		if s.Name == "天乙贵人" {
			found = true
			break
		}
	}
	if !found {
		t.Error("甲日主遇丑支应引动天乙贵人，但未找到")
	}

	// 日主甲木，日支午 → 桃花在卯
	// 今日地支卯 → 应引动桃花
	bazi2 := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "午"},
	}
	sha2 := calcShenShaActivation("丙", "卯", bazi2)
	found2 := false
	for _, s := range sha2 {
		if s.Name == "桃花" {
			found2 = true
			break
		}
	}
	if !found2 {
		t.Error("甲午日遇卯支应引动桃花，但未找到")
	}

	// 日主甲木，日支寅 → 禄神在寅
	// 今日地支寅 → 应引动禄神
	bazi3 := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "寅"},
	}
	sha3 := calcShenShaActivation("丙", "寅", bazi3)
	found3 := false
	for _, s := range sha3 {
		if s.Name == "禄神" {
			found3 = true
			break
		}
	}
	if !found3 {
		t.Error("甲寅日遇寅支应引动禄神，但未找到")
	}
}

// ── 经典示例验证：甲木身弱 + 壬子日 ──────────────────────────
// 来源：日课推算法.md 示例一
// 壬水=偏印（生我者），子水=偏印根气 → 印绶生身，利于用神
// 甲木长生在亥，子水为沐浴 → 半吉
// 综合：吉

func TestClassicExample1_JiaMuWeak_RenZiDay(t *testing.T) {
	// 十神验证：壬对甲 = 偏印
	god := bazipkg.ClassifyTenGod("壬", "甲", false)
	if god != "偏印" {
		t.Errorf("壬对甲十神: got %s, want 偏印", god)
	}

	// 长生验证：甲木在子 = 沐浴
	stage, fav, _ := calcTwelveStage("甲", "子")
	if stage != "沐浴" {
		t.Errorf("甲木在子长生: got %s, want 沐浴", stage)
	}
	if fav {
		t.Error("沐浴应为半吉(favorable=false)")
	}

	// 偏印对身弱=喜
	favGod := isFavorableTenGodByStrength("偏印", false)
	if !favGod {
		t.Error("身弱偏印应为喜")
	}
}

// ── 经典示例验证：甲木身旺 + 庚午日 ──────────────────────────
// 来源：日课推算法.md 示例二
// 庚金=七杀（克我者），午火=伤官根气
// 甲木死在午 → 凶位
// 庚金克甲木为"战"，压力明显
// 综合：身旺者能扛反吉

func TestClassicExample2_JiaMuStrong_GengWuDay(t *testing.T) {
	// 十神验证：庚对甲 = 七杀
	god := bazipkg.ClassifyTenGod("庚", "甲", false)
	if god != "七杀" {
		t.Errorf("庚对甲十神: got %s, want 七杀", god)
	}

	// 长生验证：甲木在午 = 死
	stage, _, _ := calcTwelveStage("甲", "午")
	if stage != "死" {
		t.Errorf("甲木在午长生: got %s, want 死", stage)
	}

	// 七杀对身旺=喜
	favGod := isFavorableTenGodByStrength("七杀", true)
	if !favGod {
		t.Error("身旺七杀应为喜")
	}

	// 天干关系：庚克甲 = 相冲（庚甲相冲）
	if stemClashMap["庚"] != "甲" {
		t.Errorf("庚甲应相冲，got stemClashMap[庚]=%s", stemClashMap["庚"])
	}
}

// ── 经典示例验证：丙火身弱 + 庚寅日 ──────────────────────────
// 来源：日课推算法.md 示例三
// 庚金=偏财（我克者），寅木=偏印根气
// 丙火长生在寅 → 大吉
// 寅木生丙火 → 好
// 综合：有惊无险

func TestClassicExample3_BingHuoWeak_GengYinDay(t *testing.T) {
	// 十神验证：庚对丙 = 偏财
	god := bazipkg.ClassifyTenGod("庚", "丙", false)
	if god != "偏财" {
		t.Errorf("庚对丙十神: got %s, want 偏财", god)
	}

	// 长生验证：丙火在寅 = 长生
	stage, fav, _ := calcTwelveStage("丙", "寅")
	if stage != "长生" {
		t.Errorf("丙火在寅长生: got %s, want 长生", stage)
	}
	if !fav {
		t.Error("长生应为吉(favorable=true)")
	}

	// 寅中藏甲木（偏印）→ 木生火，对丙火有利
	stems := calcHiddenStemGods("寅", "丙", []string{"木", "火"}, "丙")
	if len(stems) == 0 {
		t.Fatal("寅应有藏干")
	}
	if stems[0].Stem != "甲" {
		t.Errorf("寅本气: got %s, want 甲", stems[0].Stem)
	}
	if stems[0].TenGod != "偏印" {
		t.Errorf("甲对丙: got %s, want 偏印", stems[0].TenGod)
	}
}

// ── 贪合忘克验证 ──────────────────────────────────────────
// 今日乙庚合，庚金本克甲木，但乙合住庚 → 贪合忘克

func TestTanHeWangKe(t *testing.T) {
	bazi := &bazipkg.BaziResult{
		DayPillar:   model.Pillar{Gan: "甲", Zhi: "寅"},
		YearPillar:  model.Pillar{Gan: "庚", Zhi: "申"},
		MonthPillar: model.Pillar{Gan: "丙", Zhi: "寅"},
		HourPillar:  model.Pillar{Gan: "丁", Zhi: "卯"},
	}
	// 今日乙干，与年干庚合
	rels := calcStemRelations("乙", bazi)
	found := false
	for _, r := range rels {
		if r.Type == "五合" && r.Target == "庚" {
			found = true
			if r.Note == "" {
				t.Error("乙庚合应检测到贪合忘克（庚克甲），但Note为空")
			}
			t.Logf("贪合忘克: %s", r.Note)
			break
		}
	}
	if !found {
		t.Error("乙与庚应形成五合关系")
	}
}

// ── 地支关系综合验证 ──────────────────────────────────────────

func TestBranchRelationsComprehensive(t *testing.T) {
	bazi := &bazipkg.BaziResult{
		YearPillar:  model.Pillar{Gan: "甲", Zhi: "子"},
		MonthPillar: model.Pillar{Gan: "丙", Zhi: "午"},
		DayPillar:   model.Pillar{Gan: "庚", Zhi: "寅"},
		HourPillar:  model.Pillar{Gan: "戊", Zhi: "卯"},
	}

	// 今日午支 → 与年支子冲，与月支午自刑(午午)，与日支寅三合(寅午戌)
	rels := calcBranchRelations("午", bazi)
	t.Logf("午支关系数量: %d", len(rels))
	for _, r := range rels {
		t.Logf("  %s: %s (favorable=%v)", r.Type, r.Detail, r.IsFavorable)
	}

	// 验证子午冲
	foundClash := false
	for _, r := range rels {
		if r.Type == "六冲" && r.Target == "子" {
			foundClash = true
			break
		}
	}
	if !foundClash {
		t.Error("午与子应形成六冲关系")
	}
}

// ── 岁伤日干 vs 日犯岁君方向验证 ──────────────────────────────────
// 岁伤日干：流年克今日 → 轻
// 日犯岁君：今日克流年 → 重

func TestTaiSuiDirection(t *testing.T) {
	// 2026丙午年
	yearGanZhi := getYearGanZhi(2026)
	if yearGanZhi != "丙午" {
		t.Fatalf("2026年干支: got %s, want 丙午", yearGanZhi)
	}

	// 岁伤日干：丙(火)克庚(金) → 轻
	yearElem := data.GanElement["丙"]
	todayElem := data.GanElement["庚"]
	// 火克金 → 岁伤日干
	if elementOvercomesMap[yearElem] != todayElem {
		t.Error("丙火应克庚金（岁伤日干）")
	}

	// 日犯岁君：庚(金)克甲(木) → 重
	todayElem2 := data.GanElement["庚"]
	yearElem2 := data.GanElement["甲"]
	if elementOvercomesMap[todayElem2] != yearElem2 {
		t.Error("庚金应克甲木（日犯岁君）")
	}
}

// ── 五行相生相克循环验证 ──────────────────────────────────────────

func TestElementCycle(t *testing.T) {
	// 相生：木→火→土→金→水→木
	if elementGeneratesMap["木"] != "火" { t.Error("木生火") }
	if elementGeneratesMap["火"] != "土" { t.Error("火生土") }
	if elementGeneratesMap["土"] != "金" { t.Error("土生金") }
	if elementGeneratesMap["金"] != "水" { t.Error("金生水") }
	if elementGeneratesMap["水"] != "木" { t.Error("水生木") }

	// 相克：木→土→水→火→金→木
	if elementOvercomesMap["木"] != "土" { t.Error("木克土") }
	if elementOvercomesMap["土"] != "水" { t.Error("土克水") }
	if elementOvercomesMap["水"] != "火" { t.Error("水克火") }
	if elementOvercomesMap["火"] != "金" { t.Error("火克金") }
	if elementOvercomesMap["金"] != "木" { t.Error("金克木") }
}

// ── 综合断语评分范围验证 ──────────────────────────────────────────

func TestOverallVerdictScoreRange(t *testing.T) {
	// 极端有利情况
	v1, s1 := calcOverallVerdict(
		"正印", true, "长生", true,
		[]HiddenStemGod{{Favorable: true}, {Favorable: true}},
		[]StemRelation{{IsFavorable: true}},
		[]BranchRelation{{IsFavorable: true}, {IsFavorable: true}},
		[]ShenShaActivation{{Type: "吉神"}},
		DaYunInfluence{Score: 5},
		LiuNianInfluence{Score: 3},
		AdvanceRetreat{Score: 8},
		YongShenImpact{Score: 15},
		false, nil,
	)
	t.Logf("极端有利: score=%d, verdict=%s", s1, v1)
	if s1 < 60 || s1 > 100 {
		t.Errorf("极端有利评分应在60-100之间，got %d", s1)
	}

	// 极端不利情况
	v2, s2 := calcOverallVerdict(
		"七杀", false, "绝", false,
		[]HiddenStemGod{{Favorable: false}, {Favorable: false}},
		[]StemRelation{{IsFavorable: false}},
		[]BranchRelation{{IsFavorable: false}, {IsFavorable: false}},
		[]ShenShaActivation{{Type: "凶煞"}},
		DaYunInfluence{Score: -5},
		LiuNianInfluence{Score: -8},
		AdvanceRetreat{Score: -8},
		YongShenImpact{Score: 0},
		false, nil,
	)
	t.Logf("极端不利: score=%d, verdict=%s", s2, v2)
	if s2 < 0 || s2 > 40 {
		t.Errorf("极端不利评分应在0-40之间，got %d", s2)
	}
}

// ── 格局喜忌覆盖测试 ──────────────────────────────────────────

func TestSpecialPatternFavorOverride(t *testing.T) {
	// 稼穑格：戊土日主，辰月，全局土旺
	// 格局喜忌：喜火土，忌木水
	// 身旺喜忌：喜木金水，忌火土
	// 验证：getEffectiveFavor 应返回格局喜忌而非身旺喜忌

	t.Run("getEffectiveFavor returns pattern favor for special pattern", func(t *testing.T) {
		bazi := &bazipkg.BaziResult{
			DayPillar: model.Pillar{Gan: "戊", Zhi: "辰"},
			BodyStrength: bazipkg.BodyStrengthResult{
				Verdict: "身旺",
				Like:    []string{"木", "金", "水"},   // 身旺喜克泄耗
				Dislike: []string{"火", "土"},         // 身旺忌生扶
			},
			PatternAnalysis: bazipkg.PatternAnalysis{
				PatternType:         "特殊格局",
				PatternName:         "稼穑格（从强格）",
				FavorableElements:   []string{"火", "土"}, // 从强格喜生扶
				UnfavorableElements: []string{"木", "水"}, // 从强格忌克破
			},
		}

		like, dislike, isSpecial := getEffectiveFavor(bazi)

		if !isSpecial {
			t.Error("应识别为特殊格局")
		}
		if !isFavorableElement("火", like) {
			t.Error("稼穑格应喜火，got dislike")
		}
		if !isFavorableElement("土", like) {
			t.Error("稼穑格应喜土，got dislike")
		}
		if isFavorableElement("木", like) {
			t.Error("稼穑格应忌木，got like")
		}
		if isFavorableElement("水", like) {
			t.Error("稼穑格应忌水，got like")
		}
		_ = dislike
	})

	t.Run("getEffectiveFavor falls back to body strength for normal pattern", func(t *testing.T) {
		bazi := &bazipkg.BaziResult{
			DayPillar: model.Pillar{Gan: "甲", Zhi: "寅"},
			BodyStrength: bazipkg.BodyStrengthResult{
				Verdict: "身旺",
				Like:    []string{"金", "水", "土"},
				Dislike: []string{"木", "火"},
			},
			PatternAnalysis: bazipkg.PatternAnalysis{
				PatternType:       "正格",
				PatternName:       "正官格",
				FavorableElements: []string{"金", "水"},
			},
		}

		like, _, isSpecial := getEffectiveFavor(bazi)

		if isSpecial {
			t.Error("正格不应识别为特殊格局")
		}
		if !isFavorableElement("金", like) {
			t.Error("身旺应喜金")
		}
		if !isFavorableElement("水", like) {
			t.Error("身旺应喜水")
		}
	})

	t.Run("isFavorableTenGodByFavor for special pattern", func(t *testing.T) {
		// 稼穑格戊土：喜火土，忌木水
		like := []string{"火", "土"}
		dayGan := "戊"

		// 正印（火，生我）→ 喜
		if !isFavorableTenGodByFavor("正印", like, dayGan) {
			t.Error("稼穑格戊土，正印(火)应为喜")
		}
		// 偏印（火）→ 喜
		if !isFavorableTenGodByFavor("偏印", like, dayGan) {
			t.Error("稼穑格戊土，偏印(火)应为喜")
		}
		// 比肩（土，同我）→ 喜
		if !isFavorableTenGodByFavor("比肩", like, dayGan) {
			t.Error("稼穑格戊土，比肩(土)应为喜")
		}
		// 劫财（土）→ 喜
		if !isFavorableTenGodByFavor("劫财", like, dayGan) {
			t.Error("稼穑格戊土，劫财(土)应为喜")
		}
		// 七杀（木，克我）→ 忌
		if isFavorableTenGodByFavor("七杀", like, dayGan) {
			t.Error("稼穑格戊土，七杀(木)应为忌")
		}
		// 正官（木）→ 忌
		if isFavorableTenGodByFavor("正官", like, dayGan) {
			t.Error("稼穑格戊土，正官(木)应为忌")
		}
		// 正财（水，我克）→ 忌
		if isFavorableTenGodByFavor("正财", like, dayGan) {
			t.Error("稼穑格戊土，正财(水)应为忌")
		}
		// 偏财（水）→ 忌
		if isFavorableTenGodByFavor("偏财", like, dayGan) {
			t.Error("稼穑格戊土，偏财(水)应为忌")
		}
		// 食神（金，我生）→ 忌（不在喜用列表中）
		if isFavorableTenGodByFavor("食神", like, dayGan) {
			t.Error("稼穑格戊土，食神(金)应为忌")
		}
	})

	t.Run("calcOverallVerdict includes pattern name for special pattern", func(t *testing.T) {
		bazi := &bazipkg.BaziResult{
			DayPillar: model.Pillar{Gan: "戊", Zhi: "辰"},
			PatternAnalysis: bazipkg.PatternAnalysis{
				PatternType:         "特殊格局",
				PatternName:         "稼穑格（从强格）",
				FavorableElements:   []string{"火", "土"},
				UnfavorableElements: []string{"木", "水"},
			},
		}

		v, _ := calcOverallVerdict(
			"正印", true, "长生", true,
			[]HiddenStemGod{{Favorable: true}},
			nil, nil, nil,
			DaYunInfluence{Score: 5},
			LiuNianInfluence{Score: 3},
			AdvanceRetreat{Score: 8},
			YongShenImpact{Score: 15},
			true, bazi,
		)

		if !strContains(v, "稼穑格") {
			t.Errorf("断语应包含格局名称，got: %s", v)
		}
		if !strContains(v, "喜") || !strContains(v, "忌") {
			t.Logf("断语: %s", v)
		}
	})
}

func strContains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
