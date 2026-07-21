package fortune

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

func TestTraditionalCalendarEvidenceUsesQueryMonthBranch(t *testing.T) {
	jianInYinMonth := observeJianChu("寅", "寅")
	jianInMaoMonth := observeJianChu("卯", "寅")
	if jianInYinMonth.Name != "建" || jianInMaoMonth.Name != "闭" {
		t.Fatalf("JianChu must rotate from query month branch: yin=%+v mao=%+v", jianInYinMonth, jianInMaoMonth)
	}
	qingLongInZiMonth := observeHuangDao("子", "申")
	qingLongInChouMonth := observeHuangDao("丑", "申")
	if qingLongInZiMonth.Name != "青龙" || qingLongInChouMonth.Name != "司命" {
		t.Fatalf("twelve officers must rotate from query month branch: zi=%+v chou=%+v", qingLongInZiMonth, qingLongInChouMonth)
	}
	for _, evidence := range []model.TraditionalCalendarEvidence{jianInYinMonth, jianInMaoMonth, qingLongInZiMonth, qingLongInChouMonth} {
		if evidence.Status != "observed" || evidence.InterpretationStatus != "not_adjudicated" || evidence.MonthBranch == "" || evidence.QueryBranch == "" {
			t.Fatalf("traditional calendar evidence metadata is incomplete: %+v", evidence)
		}
	}
}

func TestObserveHuangDaoUsesCanonicalQingLongStartBranches(t *testing.T) {
	starts := map[string]string{
		"子": "申", "午": "申",
		"丑": "戌", "未": "戌",
		"寅": "子", "申": "子",
		"卯": "寅", "酉": "寅",
		"辰": "辰", "戌": "辰",
		"巳": "午", "亥": "午",
	}
	for monthBranch, dayBranch := range starts {
		got := observeHuangDao(monthBranch, dayBranch)
		if got.Name != "青龙" || got.RuleID != "rikuyo.twelve-star.tyme4go-v2" ||
			got.Basis != "tyme4go_sixty_cycle_day_twelve_star_formula" {
			t.Errorf("month=%s day=%s twelve star = %+v, want 青龙", monthBranch, dayBranch, got)
		}
	}
}

func TestCalcRikuyoHuangDaoMatchesTymeForOrdinaryDates(t *testing.T) {
	chart, err := (&bazipkg.BaziService{}).Calculate(1990, 6, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatal(err)
	}
	for month := 1; month <= 12; month++ {
		date := time.Date(2025, time.Month(month), 15, 12, 0, 0, 0, time.UTC)
		solarDay, err := tyme.SolarDay{}.FromYmd(date.Year(), month, date.Day())
		if err != nil {
			t.Fatal(err)
		}
		want := solarDay.GetLunarDay().GetTwelveStar().GetName()
		got := CalcRikuyo(chart, date).HuangDao
		if got.Name != want {
			t.Errorf("date=%s monthBranch=%s dayBranch=%s twelve star=%s, want tyme4go %s",
				date.Format("2006-01-02"), got.MonthBranch, got.QueryBranch, got.Name, want)
		}
	}
}

func TestRikuyoJSONDoesNotReintroduceLegacyJudgments(t *testing.T) {
	payload, err := json.Marshal(RikuyoResult{
		TwelveStage: observeTwelveStage("甲", "午"),
		JianChu:     observeJianChu("未", "午"),
		HuangDao:    observeHuangDao("未", "午"),
	})
	if err != nil {
		t.Fatalf("marshal Rikuyo result: %v", err)
	}
	text := string(payload)
	for _, forbidden := range []string{
		`"stage_favorable"`, `"stage_desc"`, `"stage_flexible"`,
		`"overall_verdict"`, `"favor_score"`, `"pengzu_gan_taboo"`,
		"百事皆宜", "大事不宜", "必见灾殃",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("Rikuyo JSON contains prohibited legacy output %q: %s", forbidden, text)
		}
	}
	for _, required := range []string{`"twelve_stage"`, `"jian_chu"`, `"huang_dao"`, `"not_adjudicated"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("Rikuyo JSON missing evidence marker %q: %s", required, text)
		}
	}
}

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
		evidence := observeTwelveStage(tt.dayGan, tt.branch)
		if evidence.Name != tt.expected {
			t.Errorf("observeTwelveStage(%s, %s).Name = %s, want %s", tt.dayGan, tt.branch, evidence.Name, tt.expected)
		}
		if evidence.RuleID != "rikuyo.twelve-stage-v1" || evidence.Status != "observed" || evidence.InterpretationStatus != "not_adjudicated" {
			t.Errorf("unexpected evidence metadata for %s/%s: %+v", tt.dayGan, tt.branch, evidence)
		}
	}
}

func TestTwelveStageEvidenceUnavailableForInvalidInput(t *testing.T) {
	for _, tc := range []struct{ gan, zhi string }{{"", "子"}, {"甲", ""}, {"X", "子"}, {"甲", "X"}} {
		evidence := observeTwelveStage(tc.gan, tc.zhi)
		if evidence.Status != "unavailable" || evidence.Name != "" {
			t.Fatalf("invalid input %q/%q should be unavailable: %+v", tc.gan, tc.zhi, evidence)
		}
		if evidence.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("invalid input must remain uninterpreted: %+v", evidence)
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

func TestObserveTenGod_ReturnsStructureOnly(t *testing.T) {
	evidence := observeTenGod("甲", "丙")
	if evidence.Name != "食神" || evidence.Status != "observed" {
		t.Fatalf("unexpected ten-god evidence: %+v", evidence)
	}
	if evidence.ReferenceStem != "甲" || evidence.QueryStem != "丙" || evidence.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("ten-god evidence is not auditable: %+v", evidence)
	}
	invalid := observeTenGod("甲", "invalid")
	if invalid.Status != "unavailable" || invalid.Name != "" {
		t.Fatalf("invalid query stem must be unavailable: %+v", invalid)
	}
}

// ── 地支藏干验证 ──────────────────────────────────────────

func TestHiddenStemMap(t *testing.T) {
	tests := []struct {
		branch  string
		benQi   string
		zhongQi string
		yuQi    string
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
		a, b    string
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

func TestObserveSeasonalState(t *testing.T) {
	tests := []struct {
		todayGan  string
		monthZhi  string
		wantState string
		desc      string
	}{
		{"甲", "卯", "旺", "甲木春天=旺"},
		{"丙", "卯", "相", "丙火春天=相"},
		{"壬", "卯", "休", "壬水春天=休"},
		{"庚", "卯", "囚", "庚金春天=囚"},
		{"戊", "卯", "死", "戊土春天=死"},
		{"庚", "酉", "旺", "庚金秋天=旺"},
		{"壬", "酉", "相", "壬水秋天=相"},
		{"丙", "子", "死", "丙火冬天=死"},
	}
	for _, tt := range tests {
		result := observeSeasonalState(tt.todayGan, tt.monthZhi)
		if result.State != tt.wantState || result.Status != "observed" || result.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s: got %+v", tt.desc, result)
		}
	}
}

// ── 年干支计算验证 ──────────────────────────────────────────

func TestGetYearGanZhi(t *testing.T) {
	tests := []struct {
		year, month, day int
		want             string
	}{
		{2024, 6, 1, "甲辰"},  // 立春后
		{2025, 3, 1, "乙巳"},  // 立春后
		{2026, 1, 15, "乙巳"}, // 立春前，应属上一年乙巳
		{2026, 6, 1, "丙午"},  // 立春后
		{1990, 8, 15, "庚午"}, // 立春后
		{2000, 2, 1, "己卯"},  // 立春前，1999=己卯
		{2000, 3, 1, "庚辰"},  // 立春后
		{2025, 2, 3, "甲辰"},  // 立春前，2024=甲辰
		{2025, 2, 5, "乙巳"},  // 立春后
	}
	for _, tt := range tests {
		got := getYearGanZhi(tt.year, tt.month, tt.day)
		if got != tt.want {
			t.Errorf("getYearGanZhi(%d-%02d-%02d) = %s, want %s", tt.year, tt.month, tt.day, got, tt.want)
		}
	}
}

// ── 季节判断验证 ──────────────────────────────────────────

func TestMonthZhiToSeason(t *testing.T) {
	tests := []struct {
		zhi  string
		want string
	}{
		{"寅", "春"}, {"卯", "春"}, {"辰", "春"},
		{"巳", "夏"}, {"午", "夏"}, {"未", "夏"},
		{"申", "秋"}, {"酉", "秋"}, {"戌", "秋"},
		{"亥", "冬"}, {"子", "冬"}, {"丑", "冬"},
	}
	for _, tt := range tests {
		got := monthZhiToSeason(tt.zhi)
		if got != tt.want {
			t.Errorf("monthZhiToSeason(%s) = %s, want %s", tt.zhi, got, tt.want)
		}
	}
}

func TestSolarDateToJieQiMonth(t *testing.T) {
	tests := []struct {
		month, day int
		want       int
	}{
		{1, 5, 12},  // 小寒前属上一年子月(12月)
		{1, 7, 1},   // 小寒后属丑月(1月)
		{2, 3, 1},   // 立春前属丑月(1月)
		{2, 5, 2},   // 立春后属寅月(2月)
		{6, 5, 5},   // 芒种前属巳月(5月)
		{6, 7, 6},   // 芒种后属午月(6月)
		{12, 6, 11}, // 大雪前属亥月(11月)
		{12, 8, 12}, // 大雪后属子月(12月)
	}
	for _, tt := range tests {
		got := solarDateToJieQiMonth(tt.month, tt.day)
		if got != tt.want {
			t.Errorf("solarDateToJieQiMonth(%d, %d) = %d, want %d", tt.month, tt.day, got, tt.want)
		}
	}
}

// ── 藏干十神分析验证 ──────────────────────────────────────────

func TestCalcHiddenStemGods(t *testing.T) {
	// 寅中藏甲(本气)、丙(中气)、戊(余气)
	// 日主为庚金：甲=偏财, 丙=七杀, 戊=偏印
	stems := calcHiddenStemGods("寅", "庚")
	if len(stems) != 3 {
		t.Fatalf("寅藏干数量: got %d, want 3", len(stems))
	}
	expected := []struct {
		stem string
		typ  string
		god  string
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
		if stems[i].TenGod != exp.god || stems[i].Status != "observed" || stems[i].InterpretationStatus != "not_adjudicated" {
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

	// 日主甲木，日支午 → 咸池在卯
	// 今日地支卯 → 应引动咸池
	bazi2 := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "甲", Zhi: "午"},
	}
	sha2 := calcShenShaActivation("丙", "卯", bazi2)
	found2 := false
	for _, s := range sha2 {
		if s.Name == "咸池" {
			found2 = true
			break
		}
	}
	if !found2 {
		t.Error("甲午日遇卯支应引动咸池，但未找到")
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
// 甲木长生在亥，子水查表为沐浴；这里只验证结构标签，不解释吉凶。

func TestClassicExample1_JiaMuWeak_RenZiDay(t *testing.T) {
	// 十神验证：壬对甲 = 偏印
	god := bazipkg.ClassifyTenGod("壬", "甲", false)
	if god != "偏印" {
		t.Errorf("壬对甲十神: got %s, want 偏印", god)
	}

	// 长生验证：甲木在子 = 沐浴
	stage := observeTwelveStage("甲", "子")
	if stage.Name != "沐浴" || stage.InterpretationStatus != "not_adjudicated" {
		t.Errorf("甲木在子长生证据: got %+v, want 沐浴且未裁决", stage)
	}

}

// ── 经典示例验证：甲木身旺 + 庚午日 ──────────────────────────
// 来源：日课推算法.md 示例二
// 庚金=七杀（克我者），午火=伤官根气
// 甲木在午查表为死；这里只验证结构标签，不解释吉凶。
// 庚金克甲木为"战"，压力明显
// 综合：身旺者能扛反吉

func TestClassicExample2_JiaMuStrong_GengWuDay(t *testing.T) {
	// 十神验证：庚对甲 = 七杀
	god := bazipkg.ClassifyTenGod("庚", "甲", false)
	if god != "七杀" {
		t.Errorf("庚对甲十神: got %s, want 七杀", god)
	}

	// 长生验证：甲木在午 = 死
	stage := observeTwelveStage("甲", "午")
	if stage.Name != "死" || stage.InterpretationStatus != "not_adjudicated" {
		t.Errorf("甲木在午长生证据: got %+v, want 死且未裁决", stage)
	}

	// 天干关系：庚克甲 = 相冲（庚甲相冲）
	if stemClashMap["庚"] != "甲" {
		t.Errorf("庚甲应相冲，got stemClashMap[庚]=%s", stemClashMap["庚"])
	}
}

// ── 经典示例验证：丙火身弱 + 庚寅日 ──────────────────────────
// 来源：日课推算法.md 示例三
// 庚金=偏财（我克者），寅木=偏印根气
// 丙火在寅查表为长生；这里只验证结构标签，不解释吉凶。

func TestClassicExample3_BingHuoWeak_GengYinDay(t *testing.T) {
	// 十神验证：庚对丙 = 偏财
	god := bazipkg.ClassifyTenGod("庚", "丙", false)
	if god != "偏财" {
		t.Errorf("庚对丙十神: got %s, want 偏财", god)
	}

	// 长生验证：丙火在寅 = 长生
	stage := observeTwelveStage("丙", "寅")
	if stage.Name != "长生" || stage.InterpretationStatus != "not_adjudicated" {
		t.Errorf("丙火在寅长生证据: got %+v, want 长生且未裁决", stage)
	}

	// 寅中藏甲木（偏印）→ 木生火，对丙火有利
	stems := calcHiddenStemGods("寅", "丙")
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

func TestStemRelationFiveCombineRemainsUnadjudicated(t *testing.T) {
	bazi := &bazipkg.BaziResult{
		DayPillar:   model.Pillar{Gan: "甲", Zhi: "寅"},
		YearPillar:  model.Pillar{Gan: "庚", Zhi: "申"},
		MonthPillar: model.Pillar{Gan: "丙", Zhi: "寅"},
		HourPillar:  model.Pillar{Gan: "丁", Zhi: "卯"},
	}
	// 今日乙干，与年干庚合
	rels := calcStemRelations("乙", bazi)
	found := map[string]bool{}
	for _, r := range rels {
		if r.TargetStem == "庚" {
			found[r.Type] = true
			if r.RuleID != "rikuyo.stem-relation-v3."+r.Type ||
				r.Basis != "query_day_stem_and_natal_pillar_stem_all_structures" {
				t.Fatalf("unexpected compound stem evidence: %+v", r)
			}
			if r.Type == "five_combine" && (r.CombinedElement != "金" || r.TransformationStatus != "not_adjudicated" || r.InterpretationStatus != "not_adjudicated") {
				t.Fatalf("unexpected five-combine evidence: %+v", r)
			}
		}
	}
	for _, relationType := range []string{"five_combine", "target_overcomes_query"} {
		if !found[relationType] {
			t.Errorf("乙庚缺少%s关系: %+v", relationType, rels)
		}
	}
}

func TestStemRelationClashPreservesElementControl(t *testing.T) {
	bazi := &bazipkg.BaziResult{YearPillar: model.Pillar{Gan: "庚"}}
	rels := calcStemRelations("甲", bazi)
	found := map[string]bool{}
	for _, relation := range rels {
		if relation.TargetPillar == "年干" {
			found[relation.Type] = true
		}
	}
	for _, relationType := range []string{"clash", "target_overcomes_query"} {
		if !found[relationType] {
			t.Errorf("甲庚缺少%s关系: %+v", relationType, rels)
		}
	}
}

func TestStemRelationsIncludeNatalDayStem(t *testing.T) {
	bazi := &bazipkg.BaziResult{DayPillar: model.Pillar{Gan: "庚"}}
	rels := calcStemRelations("乙", bazi)
	found := map[string]bool{}
	for _, relation := range rels {
		if relation.TargetPillar == "日干" {
			found[relation.Type] = true
		}
	}
	for _, relationType := range []string{"five_combine", "target_overcomes_query"} {
		if !found[relationType] {
			t.Errorf("流日乙与日干庚缺少%s关系: %+v", relationType, rels)
		}
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

	// 今日午支 → 与年支子冲、与月支同支、与日支形成两支组合。
	rels := calcBranchRelations("午", bazi)
	t.Logf("午支关系数量: %d", len(rels))
	for _, r := range rels {
		t.Logf("  %s: %s", r.Type, r.Name)
	}

	// 验证子午冲
	foundClash := false
	for _, r := range rels {
		if r.Type == "clash" && r.TargetBranch == "子" {
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
	yearGanZhi := getYearGanZhi(2026, 6, 1)
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
	if elementGeneratesMap["木"] != "火" {
		t.Error("木生火")
	}
	if elementGeneratesMap["火"] != "土" {
		t.Error("火生土")
	}
	if elementGeneratesMap["土"] != "金" {
		t.Error("土生金")
	}
	if elementGeneratesMap["金"] != "水" {
		t.Error("金生水")
	}
	if elementGeneratesMap["水"] != "木" {
		t.Error("水生木")
	}

	// 相克：木→土→水→火→金→木
	if elementOvercomesMap["木"] != "土" {
		t.Error("木克土")
	}
	if elementOvercomesMap["土"] != "水" {
		t.Error("土克水")
	}
	if elementOvercomesMap["水"] != "火" {
		t.Error("水克火")
	}
	if elementOvercomesMap["火"] != "金" {
		t.Error("火克金")
	}
	if elementOvercomesMap["金"] != "木" {
		t.Error("金克木")
	}
}
