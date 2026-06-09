package ziwei

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// B1: 命宫/身宫 精确测试
//
// Computes SoulPalace (命宫) and BodyPalace (身宫) branches and verifies
// against the actual implementation in calcSoulAndBody().
//
// The implementation formula:
//   monthBranch = (2 + LunarMonth - 1) % 12    // 寅(2)起正月
//   soulBranch  = fixIndex(monthBranch - HourBranch)  // 逆数生时
//   bodyBranch  = fixIndex(monthBranch + HourBranch)  // 顺数生时
//
// Also verifies SoulGan via 五虎遁: GetPalaceStem(yearStem, soulBranch).
// ═══════════════════════════════════════════════════════════════════════

// soulBodyTestCase holds parameters for a soul/body palace formula verification.
type soulBodyTestCase struct {
	name           string
	lunarMonth     int // 1-12
	hourBranch     int // 0=子 ... 11=亥
	yearStem       int // 0=甲 ... 9=癸
	wantSoulBranch string
	wantBodyBranch string
	wantSoulStem   string
}

func TestSoulBodyPalace_MultipleCombinations(t *testing.T) {
	tests := []soulBodyTestCase{
		// ── 15+ combinations across different months and hours ──
		// Formula: monthBranch = (2 + lunarMonth - 1) % 12
		//          soulBranch = fixIndex(monthBranch - hourBranch)
		//          bodyBranch = fixIndex(monthBranch + hourBranch)
		//          soulStem = GetPalaceStem(yearStem, soulBranch)

		// month=3(三月), hour=7(未/14:00) → monthBranch=4(辰), soul=9(酉), body=11(亥)
		// 癸年(9): TigerRule[9]=0(甲), offset=(9-2)=7, stem=(0+7)%10=7=辛(辛)
		{name: "三月未时_癸年", lunarMonth: 3, hourBranch: 7, yearStem: 9,
			wantSoulBranch: "酉", wantBodyBranch: "亥", wantSoulStem: "辛"},
		// month=1(正月), hour=0(子) → monthBranch=2(寅), soul=2(寅), body=2(寅)
		// 甲年(0): TigerRule[0]=2(丙), offset=0, stem=2=丙
		{name: "正月子时", lunarMonth: 1, hourBranch: 0, yearStem: 0,
			wantSoulBranch: "寅", wantBodyBranch: "寅", wantSoulStem: "丙"},
		// month=2(二月), hour=1(丑) → monthBranch=3(卯), soul=2(寅), body=4(辰)
		{name: "二月丑时", lunarMonth: 2, hourBranch: 1, yearStem: 0,
			wantSoulBranch: "寅", wantBodyBranch: "辰", wantSoulStem: "丙"},
		// month=5(五月), hour=6(午) → monthBranch=6(午), soul=0(子), body=0(子)
		// TigerRule[0]=2(丙), offset=(0-2+12)%12=10, stem=(2+10)%10=2=丙
		{name: "五月午时", lunarMonth: 5, hourBranch: 6, yearStem: 0,
			wantSoulBranch: "子", wantBodyBranch: "子", wantSoulStem: "丙"},
		// month=12(腊月), hour=0(子) → monthBranch=1(丑), soul=1(丑), body=1(丑)
		// TigerRule[0]=2(丙), offset=(1-2+12)%12=11, stem=(2+11)%10=3=丁
		{name: "腊月子时", lunarMonth: 12, hourBranch: 0, yearStem: 0,
			wantSoulBranch: "丑", wantBodyBranch: "丑", wantSoulStem: "丁"},
		// month=1(正月), hour=6(午) → monthBranch=2(寅), soul=8(申), body=8(申)
		// 癸年(9): TigerRule[9]=0(甲), offset=(8-2)=6, stem=(0+6)%10=6=庚
		{name: "正月午时", lunarMonth: 1, hourBranch: 6, yearStem: 9,
			wantSoulBranch: "申", wantBodyBranch: "申", wantSoulStem: "庚"},
		// month=6(六月), hour=4(辰/8:00) → monthBranch=7(未), soul=3(卯), body=11(亥)
		// 甲年(0): offset=(3-2)=1, stem=(2+1)%10=3=丁
		{name: "六月辰时", lunarMonth: 6, hourBranch: 4, yearStem: 0,
			wantSoulBranch: "卯", wantBodyBranch: "亥", wantSoulStem: "丁"},
		// month=8(八月), hour=2(寅/4:00) → monthBranch=9(酉), soul=7(未), body=11(亥)
		// 庚年(6): TigerRule[6]=4(戊), offset=(7-2)=5, stem=(4+5)%10=9=癸
		{name: "八月寅时", lunarMonth: 8, hourBranch: 2, yearStem: 6,
			wantSoulBranch: "未", wantBodyBranch: "亥", wantSoulStem: "癸"},
		// month=10(十月), hour=5(巳/10:00) → monthBranch=11(亥), soul=6(午), body=4(辰)
		// 丙年(2): TigerRule[2]=6(庚), offset=(6-2)=4, stem=(6+4)%10=0=甲
		{name: "十月巳时", lunarMonth: 10, hourBranch: 5, yearStem: 2,
			wantSoulBranch: "午", wantBodyBranch: "辰", wantSoulStem: "甲"},
		// month=4(四月), hour=9(酉/18:00) → monthBranch=5(巳), soul=8(申), body=2(寅)
		// 戊年(4): TigerRule[4]=0(甲), offset=(8-2)=6, stem=(0+6)%10=6=庚
		{name: "四月酉时", lunarMonth: 4, hourBranch: 9, yearStem: 4,
			wantSoulBranch: "申", wantBodyBranch: "寅", wantSoulStem: "庚"},
		// month=7(七月), hour=10(戌/20:00) → monthBranch=8(申), soul=10(戌), body=6(午)
		// 壬年(8): TigerRule[8]=8(壬), offset=(10-2)=8, stem=(8+8)%10=6=庚
		{name: "七月戌时", lunarMonth: 7, hourBranch: 10, yearStem: 8,
			wantSoulBranch: "戌", wantBodyBranch: "午", wantSoulStem: "庚"},
		// month=11(十一月), hour=11(亥/22:00) → monthBranch=0(子), soul=1(丑), body=11(亥)
		// 甲年(0): offset=(1-2+12)%12=11, stem=(2+11)%10=3=丁
		{name: "十一月亥时", lunarMonth: 11, hourBranch: 11, yearStem: 0,
			wantSoulBranch: "丑", wantBodyBranch: "亥", wantSoulStem: "丁"},
		// month=9(九月), hour=3(卯/6:00) → monthBranch=10(戌), soul=7(未), body=1(丑)
		// 辛年(7): TigerRule[7]=6(庚), offset=(7-2)=5, stem=(6+5)%10=1=乙
		{name: "九月卯时", lunarMonth: 9, hourBranch: 3, yearStem: 7,
			wantSoulBranch: "未", wantBodyBranch: "丑", wantSoulStem: "乙"},
		// month=3(三月), hour=8(申/16:00) → monthBranch=4(辰), soul=8(申), body=0(子)
		// 癸年(9): offset=(8-2)=6, stem=(0+6)%10=6=庚
		{name: "三月申时", lunarMonth: 3, hourBranch: 8, yearStem: 9,
			wantSoulBranch: "申", wantBodyBranch: "子", wantSoulStem: "庚"},
		// month=6(六月), hour=7(未/14:00) → monthBranch=7(未), soul=0(子), body=2(寅)
		// 己年(5): TigerRule[5]=2(丙), offset=(0-2+12)%12=10, stem=(2+10)%10=2=丙
		{name: "六月未时", lunarMonth: 6, hourBranch: 7, yearStem: 5,
			wantSoulBranch: "子", wantBodyBranch: "寅", wantSoulStem: "丙"},
		// month=2(二月), hour=8(申/16:00) → monthBranch=3(卯), soul=7(未), body=11(亥)
		// 乙年(1): TigerRule[1]=4(戊), offset=(7-2)=5, stem=(4+5)%10=9=癸
		{name: "二月申时", lunarMonth: 2, hourBranch: 8, yearStem: 1,
			wantSoulBranch: "未", wantBodyBranch: "亥", wantSoulStem: "癸"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			birth := &BirthData{
				LunarMonth:  tt.lunarMonth,
				HourBranch:  tt.hourBranch,
				YearStem:    tt.yearStem,
				YearBranch:  0, // not used for soul/body calc
				Gender:      "男",
				IsLeapMonth: false,
			}
			soulBranch, bodyBranch, soulStem := calcSoulAndBody(birth)

			// Verify soul palace branch
			if BranchNames[soulBranch] != tt.wantSoulBranch {
				t.Errorf("命宫地支 = %s(索引%d), 期望 %s",
					BranchNames[soulBranch], soulBranch, tt.wantSoulBranch)
			}

			// Verify body palace branch
			if BranchNames[bodyBranch] != tt.wantBodyBranch {
				t.Errorf("身宫地支 = %s(索引%d), 期望 %s",
					BranchNames[bodyBranch], bodyBranch, tt.wantBodyBranch)
			}

			// Verify soul palace heavenly stem via 五虎遁
			if StemNames[soulStem] != tt.wantSoulStem {
				t.Errorf("命宫天干 = %s(索引%d), 期望 %s",
					StemNames[soulStem], soulStem, tt.wantSoulStem)
			}
		})
	}
}

// TestSoulBodyChart_EndToEnd verifies soul/body palace through full CalculateChart.
func TestSoulBodyChart_EndToEnd(t *testing.T) {
	svc := NewZiWeiService()

	type e2eCase struct {
		name       string
		year       int
		month      int
		day        int
		hour       int
		minute     int
		gender     string
		wantSoulEB string // earthly branch of soul palace
		wantBody   string // body palace name
	}

	cases := []e2eCase{
		// 癸未年三月十四日未时 (2003-04-15 14:00) → 三月未时, 命宫酉, 身宫亥
		{name: "癸未男_三月未时", year: 2003, month: 4, day: 15, hour: 14, minute: 0, gender: "男",
			wantSoulEB: "酉", wantBody: "亥"},
		// 甲子年正月初一子时 (1984-02-02 0:00) → 正月子时, 命宫寅, 身宫寅
		{name: "甲子女_正月子时", year: 1984, month: 2, day: 2, hour: 0, minute: 0, gender: "女",
			wantSoulEB: "寅", wantBody: "寅"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			// Check soul palace (命宫 = Palaces[0])
			if chart.Palaces[0].Branch != tc.wantSoulEB {
				t.Errorf("命宫地支(Chart) = %s, 期望 %s", chart.Palaces[0].Branch, tc.wantSoulEB)
			}
			if chart.EarthlyBranchOfSoulPalace != tc.wantSoulEB {
				t.Errorf("EarthlyBranchOfSoulPalace = %s, 期望 %s",
					chart.EarthlyBranchOfSoulPalace, tc.wantSoulEB)
			}

			// Check body palace
			if chart.BodyPalace != tc.wantBody {
				t.Errorf("身宫(Chart) = %s, 期望 %s", chart.BodyPalace, tc.wantBody)
			}
			if chart.EarthlyBranchOfBodyPalace != tc.wantBody {
				t.Errorf("EarthlyBranchOfBodyPalace = %s, 期望 %s",
					chart.EarthlyBranchOfBodyPalace, tc.wantBody)
			}

			// Verify is_body_palace flag
			bodyFound := false
			for _, p := range chart.Palaces {
				if p.IsBodyPalace {
					bodyFound = true
					if p.Branch != tc.wantBody {
						t.Errorf("身宫标注在 %s, 期望 %s", p.Branch, tc.wantBody)
					}
				}
			}
			if !bodyFound {
				t.Error("未找到身宫标注的宫位")
			}
		})
	}
}

// TestSoulStem_WuHuDun verifies 五虎遁 derivation for all 10 year stems.
// For each year stem, the soul palace stem is computed via 五虎遁 from 寅宫.
func TestSoulStem_WuHuDun(t *testing.T) {
	// 五虎遁口诀: 甲己之年丙作首, 乙庚之岁戊为头, 丙辛必定寻庚起,
	//            丁壬壬位顺行流, 若问戊癸何方发, 甲寅之上好追求.
	// For year stem X, the stem at 寅 is TigerRule[X].
	// For any palace at branch B: Stem = (TigerRule[X] + (B - 2 + 12) % 12) % 10
	tests := []struct {
		yearStem        int
		soulBranch      int    // soul palace branch index
		expectedSoulStem string
	}{
		{0, 2, "丙"},  // 甲年, 命在寅→丙 (寅TigerRule=2, offset=0, (2+0)%10=2)
		{0, 9, "癸"},  // 甲年, 命在酉→癸 (offset=(9-2)=7, (2+7)%10=9=癸)
		{0, 0, "丙"},  // 甲年, 命在子→丙 (offset=(0-2+12)%12=10, (2+10)%10=2=丙)
		{1, 2, "戊"},  // 乙年, 命在寅→戊 (TigerRule=4, offset=0, (4+0)%10=4=戊)
		{9, 9, "辛"},  // 癸年, 命在酉→辛 (TigerRule=0, offset=7, (0+7)%10=7=辛)
		{9, 2, "甲"},  // 癸年, 命在寅→甲 (TigerRule=0, offset=0, 0=甲)
		{0, 1, "丁"},  // 甲年, 命在丑→丁 (offset=(1-2+12)%12=11, (2+11)%10=3=丁)
		{6, 7, "癸"},  // 庚年, 命在未→癸 (TigerRule=4(戊), offset=5, (4+5)%10=9=癸)
		{4, 8, "庚"},  // 戊年, 命在申→庚 (TigerRule=0(甲), offset=6, (0+6)%10=6=庚)
	}

	for _, tt := range tests {
		t.Run(StemNames[tt.yearStem]+"年_"+BranchNames[tt.soulBranch]+"宫", func(t *testing.T) {
			got := GetPalaceStem(tt.yearStem, tt.soulBranch)
			if StemNames[got] != tt.expectedSoulStem {
				t.Errorf("GetPalaceStem(%d,%d) = %d(%s), 期望 %s",
					tt.yearStem, tt.soulBranch, got, StemNames[got], tt.expectedSoulStem)
			}
		})
	}
}

// TestSoulBodyFormula_Implementation verifies the exact formula implementation
// against direct formula computation using the derived monthBranch.
func TestSoulBodyFormula_Implementation(t *testing.T) {
	// For each combination, verify: 
	//   monthBranch = (2 + lunarMonth - 1) % 12
	//   soulBranch  = fixIndex(monthBranch - hourBranch)
	//   bodyBranch  = fixIndex(monthBranch + hourBranch)
	//
	// The user's reference formula (in 1-based 子=1...亥=12):
	//   SoulBranch = (14 - monthIdx - hourIdx) % 12
	//   BodyBranch = (14 - monthIdx - hourIdx + 2) % 12
	// where monthIdx = month number (1-12), hourIdx = hour branch (1-based 子=1...亥=12)

	combinations := []struct {
		lunarMonth int
		hourBranch int // 0-based
	}{
		{1, 0}, {1, 3}, {1, 6}, {1, 9},
		{2, 1}, {2, 4}, {2, 7}, {2, 10},
		{3, 2}, {3, 5}, {3, 8}, {3, 11},
		{4, 0}, {4, 3}, {4, 6}, {4, 9},
		{5, 1}, {5, 4}, {5, 7}, {5, 10},
		{6, 2}, {6, 5}, {6, 8}, {6, 11},
		{7, 0}, {7, 3}, {7, 6}, {7, 9},
		{8, 1}, {8, 4}, {8, 7}, {8, 10},
		{9, 2}, {9, 5}, {9, 8}, {9, 11},
		{10, 0}, {10, 3}, {10, 6}, {10, 9},
		{11, 1}, {11, 4}, {11, 7}, {11, 10},
		{12, 2}, {12, 5}, {12, 8}, {12, 11},
	}

	for _, c := range combinations {
		name := BranchNames[(c.lunarMonth+1)%12] + "月" + BranchNames[c.hourBranch] + "时"
		t.Run(name, func(t *testing.T) {
			birth := &BirthData{
				LunarMonth:  c.lunarMonth,
				HourBranch:  c.hourBranch,
				YearStem:    0,
				YearBranch:  0,
				Gender:      "男",
				IsLeapMonth: false,
			}
			soul, body, _ := calcSoulAndBody(birth)

			// Verify: monthBranch = (2 + lunarMonth - 1) % 12
			monthBranch := (2 + c.lunarMonth - 1) % 12
			expectedSoul := fixIndex(monthBranch - c.hourBranch)
			expectedBody := fixIndex(monthBranch + c.hourBranch)

			if soul != expectedSoul {
				t.Errorf("命宫=%d(%s), 期望=%d(%s)", soul, BranchNames[soul],
					expectedSoul, BranchNames[expectedSoul])
			}
			if body != expectedBody {
				t.Errorf("身宫=%d(%s), 期望=%d(%s)", body, BranchNames[body],
					expectedBody, BranchNames[expectedBody])
			}
		})
	}
}

// TestLeapMonth_Adjustment verifies 闰月 adjustment for soul/body palace.
func TestLeapMonth_Adjustment(t *testing.T) {
	tests := []struct {
		name          string
		lunarMonth    int
		lunarDay      int
		hourBranch    int
		isLeapMonth   bool
		wantSoul      string
		wantBody      string
	}{
		// Non-leap month: no adjustment
		{name: "非闰月三月未时", lunarMonth: 3, lunarDay: 14, hourBranch: 7, isLeapMonth: false, wantSoul: "酉", wantBody: "亥"},
		// Leap month, day < 16: no adjustment
		{name: "闰月三月十四未时", lunarMonth: 3, lunarDay: 14, hourBranch: 7, isLeapMonth: true, wantSoul: "酉", wantBody: "亥"},
		// Leap month, day >= 16: adjust
		{name: "闰月三月十六未时", lunarMonth: 3, lunarDay: 16, hourBranch: 7, isLeapMonth: true, wantSoul: "戌", wantBody: "子"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			birth := &BirthData{
				LunarMonth:  tt.lunarMonth,
				LunarDay:    tt.lunarDay,
				HourBranch:  tt.hourBranch,
				YearStem:    9,
				YearBranch:  0,
				Gender:      "男",
				IsLeapMonth: tt.isLeapMonth,
			}
			soul, body, _ := calcSoulAndBody(birth)

			if BranchNames[soul] != tt.wantSoul {
				t.Errorf("命宫地支 = %s, 期望 %s", BranchNames[soul], tt.wantSoul)
			}
			if BranchNames[body] != tt.wantBody {
				t.Errorf("身宫地支 = %s, 期望 %s", BranchNames[body], tt.wantBody)
			}
		})
	}
}
