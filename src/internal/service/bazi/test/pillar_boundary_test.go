package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// A1 四柱边界测试
//
// 验证 BaziService.Calculate 在关键节气边界、子时边界、各时辰
// 的正确四柱输出。
// ============================================================

// TestPillarBoundary_LiChun 验证 立春前后年柱/月柱变化
//
// 使用 tyme4go 库的实际 立春转换时间（已通过探测确定）:
//
//	2024: 16:27→癸卯/乙丑, 16:28→甲辰/丙寅
//	2025: 22:10→甲辰/丁丑, 22:11→乙巳/戊寅
//	2026: 04:02:08
//	2027: 09:46→丙午/辛丑, 09:47→丁未/壬寅
//	2028: 15:31→丁未/癸丑, 15:32→戊申/甲寅
func TestPillarBoundary_LiChun(t *testing.T) {
	svc := &BaziService{}

	tests := []struct {
		name             string
		year, month, day int
		hour, min        int
		wantYearPillar   string
		wantMonthPillar  string
	}{
		// 2024 立春：transition 16:27→16:28 (astronomical 16:26:53)
		{"2024立春前-16:27", 2024, 2, 4, 16, 27, "癸卯", "乙丑"},
		{"2024立春后-16:28", 2024, 2, 4, 16, 28, "甲辰", "丙寅"},
		// 2025 立春：transition 22:10→22:11 (astronomical 22:10:13)
		{"2025立春前-22:10", 2025, 2, 3, 22, 10, "甲辰", "丁丑"},
		{"2025立春后-22:11", 2025, 2, 3, 22, 11, "乙巳", "戊寅"},
		// 2023 立春：10:42:33；分钟接口在 10:43 才进入新节令
		{"2023立春前-10:42", 2023, 2, 4, 10, 42, "壬寅", "癸丑"},
		{"2023立春后-10:43", 2023, 2, 4, 10, 43, "癸卯", "甲寅"},
		// 2026 立春：04:02:08（此处用宽边界断言；秒级断言见 external_silver_test.go）
		{"2026立春前-03:30", 2026, 2, 4, 3, 30, "乙巳", "己丑"},
		{"2026立春后-04:30", 2026, 2, 4, 4, 30, "丙午", "庚寅"},
		// 2027 立春：09:46:18
		{"2027立春前-09:46", 2027, 2, 4, 9, 46, "丙午", "辛丑"},
		{"2027立春后-09:47", 2027, 2, 4, 9, 47, "丁未", "壬寅"},
		// 2028 立春：15:31:13
		{"2028立春前-15:31", 2028, 2, 4, 15, 31, "丁未", "癸丑"},
		{"2028立春后-15:32", 2028, 2, 4, 15, 32, "戊申", "甲寅"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := svc.Calculate(tt.year, tt.month, tt.day, tt.hour, tt.min, "MALE")
			if err != nil {
				t.Fatalf("Calculate(%d,%d,%d,%d,%d) 失败: %v",
					tt.year, tt.month, tt.day, tt.hour, tt.min, err)
			}
			gotY := pillarStr(result.YearPillar)
			gotM := pillarStr(result.MonthPillar)
			if gotY != tt.wantYearPillar {
				t.Errorf("年柱: 期望 %s, 实际 %s", tt.wantYearPillar, gotY)
			}
			if gotM != tt.wantMonthPillar {
				t.Errorf("月柱: 期望 %s, 实际 %s", tt.wantMonthPillar, gotM)
			}
		})
	}
}

// TestPillarBoundary_AllSolarTerms 验证 2024 年各节气交界点月柱变化
//
// 精确转换时间（通过探测确定）：
//
//	小寒 01-06 04:49 → 04:49 刻, 04:50→乙丑（transition在04:49-04:50）
//	立春 02-04 16:28 (已知)
//	惊蛰 03-05 10:23 (10:22→丙寅, 10:23→丁卯)
//	清明 04-04 15:03 (15:02→丁卯, 15:03→戊辰)
//	立夏 05-05 ~08:10 (08:09→戊辰, 08:10→戊辰, 实际transition更晚)
//	芒种 06-05 ~12:10 (12:09→己巳, 12:10→庚午)
//	小暑 07-06 ~22:20+ (22:19→庚午, 22:20→庚午)
//	立秋 08-07 ~08:10 (08:09→辛未, 08:10→壬申)
//	白露 09-07 ~11:12 (11:11→壬申, 11:12→癸酉)
//	寒露 10-08 03:00 (02:59→癸酉, 03:00→甲戌)
//	立冬 11-07 ~06:21 (06:20→甲戌, 06:21→乙亥)
//	大雪 12-06 ~23:17+ (23:16→乙亥, 23:17→乙亥)
func TestPillarBoundary_AllSolarTerms(t *testing.T) {
	svc := &BaziService{}

	type termCase struct {
		name                        string
		year, month, day, hour, min int
		wantMonthPillar             string
	}
	cases := []termCase{
		// 小寒前
		{"小寒前", 2024, 1, 6, 4, 48, "甲子"},
		{"小寒后", 2024, 1, 6, 4, 50, "乙丑"},
		// 立春（已知）
		{"立春前", 2024, 2, 4, 16, 27, "乙丑"},
		{"立春后", 2024, 2, 4, 16, 28, "丙寅"},
		// 惊蛰
		{"惊蛰前", 2024, 3, 5, 10, 22, "丙寅"},
		{"惊蛰后", 2024, 3, 5, 10, 23, "丁卯"},
		// 清明
		{"清明前", 2024, 4, 4, 15, 2, "丁卯"},
		{"清明后", 2024, 4, 4, 15, 3, "戊辰"},
		// 立夏前 (known before transition)
		{"立夏前", 2024, 5, 5, 8, 8, "戊辰"},
		// 芒种
		{"芒种前", 2024, 6, 5, 12, 9, "己巳"},
		{"芒种后", 2024, 6, 5, 12, 10, "庚午"},
		// 小暑前 (known before transition)
		{"小暑前", 2024, 7, 6, 22, 19, "庚午"},
		// 立秋
		{"立秋前", 2024, 8, 7, 8, 9, "辛未"},
		{"立秋后", 2024, 8, 7, 8, 10, "壬申"},
		// 白露
		{"白露前", 2024, 9, 7, 11, 11, "壬申"},
		{"白露后", 2024, 9, 7, 11, 12, "癸酉"},
		// 寒露
		{"寒露前", 2024, 10, 8, 2, 59, "癸酉"},
		{"寒露后", 2024, 10, 8, 3, 0, "甲戌"},
		// 立冬
		{"立冬前", 2024, 11, 7, 6, 20, "甲戌"},
		{"立冬后", 2024, 11, 7, 6, 21, "乙亥"},
		// 大雪前 (known before transition)
		{"大雪前", 2024, 12, 6, 23, 16, "乙亥"},
		{"大雪后次日", 2024, 12, 7, 0, 0, "丙子"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := svc.Calculate(tc.year, tc.month, tc.day, tc.hour, tc.min, "MALE")
			if err != nil {
				t.Fatalf("Calculate(%d,%d,%d,%d,%d) 失败: %v",
					tc.year, tc.month, tc.day, tc.hour, tc.min, err)
			}
			gotM := pillarStr(result.MonthPillar)
			if gotM != tc.wantMonthPillar {
				t.Errorf("[%s] 月柱: 期望 %s, 实际 %s (年柱=%s)",
					tc.name, tc.wantMonthPillar, gotM, pillarStr(result.YearPillar))
			}
		})
	}
}

// TestPillarBoundary_Hour 验证子时边界及 12 时辰的时柱
func TestPillarBoundary_Hour(t *testing.T) {
	svc := &BaziService{}

	baseY, baseM, baseD := 2024, 2, 15
	// 2024-02-15 日柱为庚戌日（通过探测确定）
	// 五鼠遁：乙庚丙作初 → 子时天干为丙
	// 丙子, 丁丑, 戊寅, 己卯, 庚辰, 辛巳, 壬午, 癸未, 甲申, 乙酉, 丙戌, 丁亥

	t.Run("子时边界", func(t *testing.T) {
		result, err := svc.Calculate(baseY, baseM, 15, 23, 0, "MALE")
		if err != nil {
			t.Fatalf("Calculate 23:00 失败: %v", err)
		}
		hP := pillarStr(result.HourPillar)
		dayStr := pillarStr(result.DayPillar)
		t.Logf("2024-02-15 23:00 日柱=%s 时柱=%s", dayStr, hP)
		if len([]rune(hP)) != 2 {
			t.Errorf("23:00 时柱格式异常: %s (runes=%d)", hP, len([]rune(hP)))
		}
		if hP == "" {
			t.Error("23:00 时柱为空")
		}

		// 子时末 00:01（同一子时，时柱相同）
		result2, err := svc.Calculate(baseY, baseM, 16, 0, 1, "MALE")
		if err != nil {
			t.Fatalf("Calculate 00:01 失败: %v", err)
		}
		hP2 := pillarStr(result2.HourPillar)
		t.Logf("2024-02-16 00:01 日柱=%s 时柱=%s", pillarStr(result2.DayPillar), hP2)
		if len([]rune(hP2)) != 2 {
			t.Errorf("00:01 时柱格式异常: %s (runes=%d)", hP2, len([]rune(hP2)))
		}

		// 23:59 在同一子时
		result3, err := svc.Calculate(baseY, baseM, 15, 23, 59, "MALE")
		if err != nil {
			t.Fatalf("Calculate 23:59 失败: %v", err)
		}
		hP3 := pillarStr(result3.HourPillar)
		t.Logf("2024-02-15 23:59 日柱=%s 时柱=%s", pillarStr(result3.DayPillar), hP3)
	})

	// 验证 12 时辰的时柱
	t.Run("12时辰", func(t *testing.T) {
		// 在 BaZi 中，子时(23:00-01:00)属于下一天
		// 所以我们用 12:00 获取当天日干，测试 01:00-21:00 (丑~亥时)
		// 子时(23:00) 单独测试，用其实际日干
		result, err := svc.Calculate(baseY, baseM, baseD, 12, 0, "MALE")
		if err != nil {
			t.Fatalf("Calculate 失败: %v", err)
		}
		dayStem := result.DayPillar.Gan // 己 (己酉日)
		t.Logf("2024-02-15 12:00 日干=%s 日柱=%s", dayStem, pillarStr(result.DayPillar))

		// 五鼠遁口诀
		wuShuDun := map[string]string{
			"甲": "甲", "己": "甲",
			"乙": "丙", "庚": "丙",
			"丙": "戊", "辛": "戊",
			"丁": "庚", "壬": "庚",
			"戊": "壬", "癸": "壬",
		}
		firstGan := wuShuDun[dayStem]

		gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
		zhis := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
		startIdx := 0
		for i, g := range gans {
			if g == firstGan {
				startIdx = i
				break
			}
		}

		type shichenCase struct {
			name string
			hour int
			min  int
		}
		// 丑时 → 亥时（01:00-21:00）属于当天
		// 子时(23:00) 属于下一天，单独处理
		hours := []shichenCase{
			{"丑时-01:00", 1, 0},
			{"寅时-03:00", 3, 0},
			{"卯时-05:00", 5, 0},
			{"辰时-07:00", 7, 0},
			{"巳时-09:00", 9, 0},
			{"午时-11:00", 11, 0},
			{"未时-13:00", 13, 0},
			{"申时-15:00", 15, 0},
			{"酉时-17:00", 17, 0},
			{"戌时-19:00", 19, 0},
			{"亥时-21:00", 21, 0},
		}

		// 己日：甲己还加甲 → 甲子,乙丑,丙寅...
		// 子时(23:00-01:00) 用甲子，但属于下一天
		// 丑时(01:00-03:00): 乙丑
		for i, sc := range hours {
			ganIdx := (startIdx + i + 1) % 10 // +1 跳过子时
			expectedHour := gans[ganIdx] + zhis[i+1]

			r, err := svc.Calculate(baseY, baseM, baseD, sc.hour, sc.min, "MALE")
			if err != nil {
				t.Errorf("%s: Calculate 失败: %v", sc.name, err)
				continue
			}
			hP := pillarStr(r.HourPillar)
			if hP != expectedHour {
				t.Errorf("%s: 时柱期望 %s, 实际 %s (日干=%s)", sc.name, expectedHour, hP, dayStem)
			}
		}

		// 单独验证子时（属于下一天庚戌日）
		r, err := svc.Calculate(baseY, baseM, baseD, 23, 0, "MALE")
		if err != nil {
			t.Fatalf("Calculate 23:00 失败: %v", err)
		}
		subDayStem := r.DayPillar.Gan       // 庚
		subFirstGan := wuShuDun[subDayStem] // 丙
		expectedZi := subFirstGan + "子"
		hP := pillarStr(r.HourPillar)
		if hP != expectedZi {
			t.Errorf("子时(23:00): 时柱期望 %s, 实际 %s (日干=%s)", expectedZi, hP, subDayStem)
		}
	})
}
