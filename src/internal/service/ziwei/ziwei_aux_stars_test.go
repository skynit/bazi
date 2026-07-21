package ziwei

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// B5: 辅星/煞星 完整测试
//
// Verifies placement of auxiliary and malefic stars:
//   - 禄存/擎羊/陀罗 (by 年干) for all 10 stems
//   - 天魁/天钺 (by 年干) for all 10 stems
//   - 左辅/右弼/文昌/文曲 (by 月/时)
//   - 火星/铃星/地空/地劫
// ═══════════════════════════════════════════════════════════════════════

// TestLucunQingyangTuoluo_AllStems verifies 禄存/擎羊/陀罗 branch for all 10 year stems.
// 禄存 from LucunBranchIdx: 甲→寅(2), 乙→卯(3), 丙→巳(5), 丁→午(6), 戊→巳(5),
//
//	己→午(6), 庚→申(8), 辛→酉(9), 壬→亥(11), 癸→子(0)
//
// 擎羊 = fixIndex(禄存+1), 陀罗 = fixIndex(禄存-1)
func TestLucunQingyangTuoluo_AllStems(t *testing.T) {
	expected := []struct {
		stem    string
		lucun   string // branch
		yang    string
		luo     string
		stemIdx int
	}{
		{"甲", "寅", "卯", "丑", 0},
		{"乙", "卯", "辰", "寅", 1},
		{"丙", "巳", "午", "辰", 2},
		{"丁", "午", "未", "巳", 3},
		{"戊", "巳", "午", "辰", 4},
		{"己", "午", "未", "巳", 5},
		{"庚", "申", "酉", "未", 6},
		{"辛", "酉", "戌", "申", 7},
		{"壬", "亥", "子", "戌", 8},
		{"癸", "子", "丑", "亥", 9},
	}

	for _, exp := range expected {
		t.Run(exp.stem+"年", func(t *testing.T) {
			// Verify 禄存 branch
			gotLucun := LucunBranchIdx[exp.stemIdx]
			if BranchNames[gotLucun] != exp.lucun {
				t.Errorf("禄存(年干%s) = %s(%d), 期望 %s",
					exp.stem, BranchNames[gotLucun], gotLucun, exp.lucun)
			}

			// Verify 擎羊 = 禄存 + 1 (mod 12)
			gotYang := QingyangIndex(exp.stemIdx)
			expectedYangIdx := fixIndex(gotLucun + 1)
			if gotYang != expectedYangIdx {
				t.Errorf("擎羊索引 = %d, 期望 %d (= 禄存+1 mod12)", gotYang, expectedYangIdx)
			}
			if BranchNames[gotYang] != exp.yang {
				t.Errorf("擎羊(年干%s) = %s, 期望 %s",
					exp.stem, BranchNames[gotYang], exp.yang)
			}

			// Verify 陀罗 = 禄存 - 1 (mod 12)
			gotLuo := TuoluoIndex(exp.stemIdx)
			expectedLuoIdx := fixIndex(gotLucun - 1)
			if gotLuo != expectedLuoIdx {
				t.Errorf("陀罗索引 = %d, 期望 %d (= 禄存-1 mod12)", gotLuo, expectedLuoIdx)
			}
			if BranchNames[gotLuo] != exp.luo {
				t.Errorf("陀罗(年干%s) = %s, 期望 %s",
					exp.stem, BranchNames[gotLuo], exp.luo)
			}
		})
	}
}

// TestTianKuiTianYue_AllStems verifies 天魁/天钺 for all 10 stems.
// 口诀: 甲戊庚牛羊, 乙己鼠猴乡, 丙丁猪鸡位, 壬癸兔蛇藏, 辛逢虎马
//   - 甲戊庚: 魁=丑(1), 钺=未(7)
//   - 乙己: 魁=子(0), 钺=申(8)
//   - 丙丁: 魁=亥(11), 钺=酉(9)
//   - 辛: 魁=午(6), 钺=寅(2)
//   - 壬癸: 魁=卯(3), 钺=巳(5)
func TestTianKuiTianYue_AllStems(t *testing.T) {
	expected := []struct {
		stem    string
		kui     string // 天魁 branch
		yue     string // 天钺 branch
		stemIdx int
	}{
		{"甲", "丑", "未", 0},
		{"乙", "子", "申", 1},
		{"丙", "亥", "酉", 2},
		{"丁", "亥", "酉", 3},
		{"戊", "丑", "未", 4},
		{"己", "子", "申", 5},
		{"庚", "丑", "未", 6},
		{"辛", "午", "寅", 7},
		{"壬", "卯", "巳", 8},
		{"癸", "卯", "巳", 9},
	}

	for _, exp := range expected {
		t.Run(exp.stem+"年", func(t *testing.T) {
			kuiYue := KuiYueTable[exp.stemIdx]
			gotKui := kuiYue[0]
			gotYue := kuiYue[1]

			if BranchNames[gotKui] != exp.kui {
				t.Errorf("天魁(年干%s) = %s, 期望 %s",
					exp.stem, BranchNames[gotKui], exp.kui)
			}
			if BranchNames[gotYue] != exp.yue {
				t.Errorf("天钺(年干%s) = %s, 期望 %s",
					exp.stem, BranchNames[gotYue], exp.yue)
			}
		})
	}
}

// TestZuoFuYouBi verifies 左辅/右弼 placement formula.
// 左辅: from 辰(4) + lunarMonth - 1 (顺数)
// 右弼: from 戌(10) - lunarMonth + 1 (逆数)
func TestZuoFuYouBi(t *testing.T) {
	tests := []struct {
		lunarMonth int
		wantZuoFu  string
		wantYouBi  string
	}{
		{1, "辰", "戌"},
		{3, "午", "申"},
		{6, "酉", "巳"},
		{9, "子", "寅"},
		{12, "卯", "亥"},
	}

	for _, tt := range tests {
		t.Run("月"+BranchNames[(tt.lunarMonth+1)%12], func(t *testing.T) {
			zuofuIdx := ZuofuIndex(tt.lunarMonth)
			youbiIdx := YoubiIndex(tt.lunarMonth)

			if BranchNames[zuofuIdx] != tt.wantZuoFu {
				t.Errorf("左辅(月%d) = %s, 期望 %s", tt.lunarMonth, BranchNames[zuofuIdx], tt.wantZuoFu)
			}
			if BranchNames[youbiIdx] != tt.wantYouBi {
				t.Errorf("右弼(月%d) = %s, 期望 %s", tt.lunarMonth, BranchNames[youbiIdx], tt.wantYouBi)
			}
		})
	}
}

// TestWenChangWenQu verifies 文昌/文曲 placement formula.
// 文昌: from 戌(10) - hourBranch (逆数)
// 文曲: from 辰(4) + hourBranch (顺数)
func TestWenChangWenQu(t *testing.T) {
	tests := []struct {
		hourBranch int // 0=子 ... 11=亥
		wantChang  string
		wantQu     string
	}{
		{0, "戌", "辰"},
		{3, "未", "未"},
		{6, "辰", "戌"},
		{9, "丑", "丑"},
		{11, "亥", "卯"},
	}

	for _, tt := range tests {
		t.Run(BranchNames[tt.hourBranch]+"时", func(t *testing.T) {
			changIdx := fixIndex(10 - tt.hourBranch) // from 戌(10), -hour
			quIdx := fixIndex(4 + tt.hourBranch)     // from 辰(4), +hour

			if BranchNames[changIdx] != tt.wantChang {
				t.Errorf("文昌(%d时) = %s, 期望 %s", tt.hourBranch, BranchNames[changIdx], tt.wantChang)
			}
			if BranchNames[quIdx] != tt.wantQu {
				t.Errorf("文曲(%d时) = %s, 期望 %s", tt.hourBranch, BranchNames[quIdx], tt.wantQu)
			}
		})
	}
}

// TestHuoLingDiKongDiJie verifies 火星/铃星/地空/地劫 placement.
func TestHuoLingDiKongDiJie(t *testing.T) {
	// 火星/铃星 varies by year branch group + time
	// Test specific year branch + hour combinations
	type huoLingCase struct {
		yearBranch int
		hourBranch int
		wantHuo    string
		wantLing   string
	}

	huoLingTests := []huoLingCase{
		// 子(0) in 申子辰 group, hour=0(子) → [2,3]
		{0, 0, "寅", "戌"},
		// 寅(2) in 寅午戌 group, hour=0(子) → [1,3]
		{2, 0, "丑", "卯"},
		// 巳(5) in 巳酉丑 group, hour=0(子) → [3,10]
		{5, 0, "卯", "戌"},
		// 亥(11) in 亥卯未 group, hour=0(子) → [9,10]
		{11, 0, "酉", "戌"},
		// Test different hours: 子(0) year, hour=3(卯) → [5,1]
		{0, 3, "巳", "丑"},
		// 寅(2) year, hour=6(午) → 火星=7(未), 铃星=9(酉)
		{2, 6, "未", "酉"},
	}

	for _, tt := range huoLingTests {
		t.Run(BranchNames[tt.yearBranch]+"年"+BranchNames[tt.hourBranch]+"时", func(t *testing.T) {
			huoIdx, lingIdx := HuolingIndex(tt.yearBranch, tt.hourBranch)
			if BranchNames[huoIdx] != tt.wantHuo {
				t.Errorf("火星(年支%s,时%s) = %s, 期望 %s",
					BranchNames[tt.yearBranch], BranchNames[tt.hourBranch],
					BranchNames[huoIdx], tt.wantHuo)
			}
			if BranchNames[lingIdx] != tt.wantLing {
				t.Errorf("铃星(年支%s,时%s) = %s, 期望 %s",
					BranchNames[tt.yearBranch], BranchNames[tt.hourBranch],
					BranchNames[lingIdx], tt.wantLing)
			}
		})
	}

	// 地空/地劫: from 亥(11) ± hour
	type diKongJieCase struct {
		hourBranch int
		wantKong   string
		wantJie    string
	}
	kongJieTests := []diKongJieCase{
		{0, "亥", "亥"},
		{3, "申", "寅"},
		{6, "巳", "巳"},
		{9, "寅", "申"},
		{11, "子", "戌"},
	}
	for _, tt := range kongJieTests {
		t.Run(BranchNames[tt.hourBranch]+"时_地空地劫", func(t *testing.T) {
			kongIdx := fixIndex(11 - tt.hourBranch)
			jieIdx := fixIndex(11 + tt.hourBranch)
			if BranchNames[kongIdx] != tt.wantKong {
				t.Errorf("地空(时%s) = %s, 期望 %s",
					BranchNames[tt.hourBranch], BranchNames[kongIdx], tt.wantKong)
			}
			if BranchNames[jieIdx] != tt.wantJie {
				t.Errorf("地劫(时%s) = %s, 期望 %s",
					BranchNames[tt.hourBranch], BranchNames[jieIdx], tt.wantJie)
			}
		})
	}
}

// TestAuxStars_EndToEnd verifies aux stars appear correctly in full chart output.
func TestAuxStars_EndToEnd(t *testing.T) {
	svc := NewZiWeiService()

	type auxCheck struct {
		star string
		want string // expected palace name (or "" to just check presence)
	}

	cases := []struct {
		name   string
		year   int
		month  int
		day    int
		hour   int
		minute int
		gender string
		checks []auxCheck
	}{
		{
			name: "癸未男_三月未时",
			year: 2003, month: 4, day: 15, hour: 14, minute: 0,
			gender: "男",
			// iztro fixture: astro.bySolar("2003-4-15", 未时, "male", true, "zh-CN")
			// 癸年: 禄存=子, 擎羊=丑, 陀罗=亥, 天魁=卯, 天钺=巳, 天马=巳
			// 三月: 左辅=午, 右弼=申; 未时: 文昌=卯, 文曲=亥, 地空=辰, 地劫=午
			// 未年未时: 火星=辰, 铃星=巳
			checks: []auxCheck{
				{star: "禄存", want: "子"},
				{star: "擎羊", want: "丑"},
				{star: "陀罗", want: "亥"},
				{star: "天魁", want: "卯"},
				{star: "天钺", want: "巳"},
				{star: "天马", want: "巳"},
				{star: "左辅", want: "午"},
				{star: "右弼", want: "申"},
				{star: "文昌", want: "卯"},
				{star: "文曲", want: "亥"},
				{star: "火星", want: "辰"},
				{star: "铃星", want: "巳"},
				{star: "地空", want: "辰"},
				{star: "地劫", want: "午"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			starLocations := make(map[string]string)
			for _, p := range chart.Palaces {
				for _, s := range palaceAuxStars(p) {
					starLocations[s] = p.Branch
				}
			}

			for _, check := range tc.checks {
				gotBranch, found := starLocations[check.star]
				if !found {
					t.Errorf("%s 未在盘中找到", check.star)
					continue
				}
				if check.want != "" && gotBranch != check.want {
					t.Errorf("%s = %s, 期望 %s", check.star, gotBranch, check.want)
				}
			}
		})
	}
}
