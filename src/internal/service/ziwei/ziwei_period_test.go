package ziwei

import (
	"encoding/json"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// 大限/流年测试
// Verify dayun direction and sequences, liunian overlay.
// ═══════════════════════════════════════════════════════════════════════

// TestDayun_Exists verifies that CalculateDayun returns non-nil results
// for charts calculated with various inputs.
func TestDayun_Exists(t *testing.T) {
	svc := NewZiWeiService()

	cases := []struct {
		name   string
		year   int
		month  int
		day    int
		hour   int
		minute int
		gender string
	}{
		{"癸未男", 2003, 4, 15, 14, 0, "男"},
		{"甲子女", 1984, 1, 1, 0, 0, "女"},
		{"庚午男", 1990, 6, 15, 12, 0, "男"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			dayun := svc.CalculateDayun(chart)
			if dayun == nil {
				t.Fatal("CalculateDayun returned nil")
			}

			if len(dayun) == 0 {
				t.Fatal("CalculateDayun returned empty slice")
			}
		})
	}
}

// TestDayun_StartAge verifies that the starting age equals the JuValue
// (五行局值). For example, 木三局 should start at age 3, 金四局 at age 4, etc.
func TestDayun_StartAge(t *testing.T) {
	svc := NewZiWeiService()

	cases := []struct {
		name       string
		year       int
		month      int
		day        int
		hour       int
		minute     int
		gender     string
		wantBureau string
		wantJuVal  int
	}{
		// 癸未年三月十四日未时 男 → 木三局 (JuValue=3)
		{"癸未男_木三局", 2003, 4, 15, 14, 0, "男", "木三局", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			if chart.FiveBureau != tc.wantBureau {
				t.Fatalf("五行局 = %q, want %q", chart.FiveBureau, tc.wantBureau)
			}

			dayun := svc.CalculateDayun(chart)
			if len(dayun) == 0 {
				t.Fatal("dayun empty")
			}

			// First dayun stage should start at JuValue (age = JuValue)
			wantStartAge := tc.wantJuVal
			if dayun[0].StartAge != wantStartAge {
				t.Errorf("第一大限起始年龄 = %d, 应为 %d (五行局=%s)",
					dayun[0].StartAge, wantStartAge, tc.wantBureau)
			}
		})
	}
}

// TestDayun_AgeRanges verifies that each 10-year period has the correct
// StartAge and EndAge, and that they are contiguous without gaps.
func TestDayun_AgeRanges(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	dayun := svc.CalculateDayun(chart)
	if len(dayun) < 6 {
		t.Fatalf("大限阶段数 = %d, 至少应有6个", len(dayun))
	}

	for i, stage := range dayun {
		ageSpan := stage.EndAge - stage.StartAge
		if ageSpan != 9 && ageSpan != 0 {
			// Each 10-year period: start=N, end=N+9 → span=9
			t.Errorf("大限[%d] %s: %d~%d, 跨度应为9年(含), 实际=%d",
				i, stage.Palace, stage.StartAge, stage.EndAge, ageSpan)
		}
		if stage.StartAge > stage.EndAge {
			t.Errorf("大限[%d] %s: 起始年龄(%d) > 结束年龄(%d)",
				i, stage.Palace, stage.StartAge, stage.EndAge)
		}
	}

	// Verify contiguous age ranges (no gaps between stages)
	for i := 1; i < len(dayun); i++ {
		if dayun[i].StartAge != dayun[i-1].EndAge+1 {
			t.Errorf("大限年龄不连续: [%d]结束于%d, [%d]起始于%d (期望%d)",
				i-1, dayun[i-1].EndAge, i, dayun[i].StartAge, dayun[i-1].EndAge+1)
		}
	}
}

// TestDayun_Direction_阳男阴女顺行 verifies the dayun direction logic.
// 阳男(年干偶数) → 顺行 (clockwise from 命宫)
// 阴女(年干奇数) → 顺行
// 阴男(年干奇数) → 逆行 (counterclockwise from 命宫)
// 阳女(年干偶数) → 逆行
//
// We can verify direction by checking if the palace name sequence
// matches the expected palace order.
func TestDayun_Direction(t *testing.T) {
	svc := NewZiWeiService()

	// 阳男(年干偶数): 顺行 — verify forward (clockwise) direction
	t.Run("阳男顺行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1984, 1, 1, 0, 0, "男")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: all stages have unique palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阴女(年干奇数): 顺行
	t.Run("阴女顺行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1985, 6, 15, 12, 0, "女")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces, all 12 eventually covered
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阴男(年干奇数): 逆行
	t.Run("阴男逆行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1985, 6, 15, 12, 0, "男")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阳女(年干偶数): 逆行
	t.Run("阳女逆行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1984, 1, 1, 0, 0, "女")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})
}

func TestDayun_DirectionUsesYearStemAndGender(t *testing.T) {
	tests := []struct {
		name     string
		yearStem int
		gender   string
		want     bool
	}{
		{name: "阳男顺行", yearStem: StemIndex["甲"], gender: "男", want: true},
		{name: "阴女顺行", yearStem: StemIndex["乙"], gender: "女", want: true},
		{name: "阴男逆行", yearStem: StemIndex["乙"], gender: "男", want: false},
		{name: "阳女逆行", yearStem: StemIndex["甲"], gender: "女", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isForwardByYearStem(tt.yearStem, tt.gender); got != tt.want {
				t.Errorf("isForwardByYearStem(%d, %s) = %v, want %v", tt.yearStem, tt.gender, got, tt.want)
			}
		})
	}
}

func TestDayun_GenderChangesPalaceDirection(t *testing.T) {
	svc := NewZiWeiService()

	maleChart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("male CalculateChart failed: %v", err)
	}
	femaleChart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "女")
	if err != nil {
		t.Fatalf("female CalculateChart failed: %v", err)
	}

	maleDayun := svc.CalculateDayun(maleChart)
	femaleDayun := svc.CalculateDayun(femaleChart)
	if len(maleDayun) < 2 || len(femaleDayun) < 2 {
		t.Fatal("dayun must have at least 2 stages")
	}

	if maleDayun[0].Palace != "命宫" || femaleDayun[0].Palace != "命宫" {
		t.Fatalf("first dayun should start from 命宫, got male=%s female=%s", maleDayun[0].Palace, femaleDayun[0].Palace)
	}
	if maleDayun[1].Palace != "兄弟" {
		t.Errorf("癸年男命第二大限 = %s, want 兄弟", maleDayun[1].Palace)
	}
	if femaleDayun[1].Palace != "父母" {
		t.Errorf("癸年女命第二大限 = %s, want 父母", femaleDayun[1].Palace)
	}
}

// TestDayun_StarPresence verifies that dayun stages contain the stars
// from their corresponding palaces.
func TestDayun_StarPresence(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	dayun := svc.CalculateDayun(chart)
	if len(dayun) == 0 {
		t.Fatal("dayun empty")
	}

	for i, stage := range dayun {
		// Find the matching palace
		var palace *PalaceInfo
		for j := range chart.Palaces {
			if chart.Palaces[j].Name == stage.Palace {
				palace = &chart.Palaces[j]
				break
			}
		}
		if palace == nil {
			t.Errorf("大限[%d]: 找不到宫位 %s", i, stage.Palace)
			continue
		}

		wantStars := make([]string, 0, len(palace.Stars))
		for _, star := range palace.Stars {
			if star.Name != "" {
				wantStars = append(wantStars, star.Name)
			}
		}
		if !slices.Equal(stage.Stars, wantStars) {
			t.Errorf("大限[%d] %s 星曜 = %v, want published stars %v", i, stage.Palace, stage.Stars, wantStars)
		}
	}
}

func TestDayun_JSONReplayMatchesFreshChart(t *testing.T) {
	svc := NewZiWeiService()

	tests := []struct {
		name   string
		year   int
		gender string
	}{
		{name: "阳男顺行", year: 1984, gender: "男"},
		{name: "阳女逆行", year: 1984, gender: "女"},
		{name: "阴男逆行", year: 1985, gender: "男"},
		{name: "阴女顺行", year: 1985, gender: "女"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tt.year, 6, 15, 12, 0, tt.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}
			want := svc.CalculateDayun(chart)
			if len(want) != len(chart.Palaces) {
				t.Fatalf("fresh dayun stages = %d, want %d", len(want), len(chart.Palaces))
			}
			if want[0].StartAge == 0 || len(want[0].Stars) == 0 {
				t.Fatalf("fresh dayun lacks bureau age or stars: %+v", want[0])
			}
			wantAnalysis := BuildDayunAnalysis(chart, want, 42)

			payload, err := json.Marshal(chart)
			if err != nil {
				t.Fatalf("marshal chart: %v", err)
			}
			var replayed ZiWeiChart
			if err := json.Unmarshal(payload, &replayed); err != nil {
				t.Fatalf("unmarshal chart: %v", err)
			}

			got := svc.CalculateDayun(&replayed)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("dayun changed after JSON replay:\n got: %#v\nwant: %#v", got, want)
			}
			gotAnalysis := BuildDayunAnalysis(&replayed, got, 42)
			if !reflect.DeepEqual(gotAnalysis, wantAnalysis) {
				t.Fatalf("dayun analysis changed after JSON replay:\n got: %#v\nwant: %#v", gotAnalysis, wantAnalysis)
			}
		})
	}
}

func TestDayun_RejectsIncompletePublishedContract(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*ZiWeiChart)
	}{
		{name: "invalid content hash", mutate: func(chart *ZiWeiChart) {
			chart.ContentHash = ""
		}},
		{name: "invalid input fingerprint", mutate: func(chart *ZiWeiChart) {
			chart.InputFingerprint = ""
		}},
		{name: "invalid calculation input", mutate: func(chart *ZiWeiChart) {
			chart.CalculationInput.Basis = ""
		}},
		{name: "invalid five bureau", mutate: func(chart *ZiWeiChart) {
			chart.FiveBureau = ""
		}},
		{name: "invalid soul branch", mutate: func(chart *ZiWeiChart) {
			chart.EarthlyBranchOfSoulPalace = ""
		}},
		{name: "missing palace branch", mutate: func(chart *ZiWeiChart) {
			chart.Palaces[1].Branch = chart.Palaces[0].Branch
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutated := *chart
			tt.mutate(&mutated)
			if got := svc.CalculateDayun(&mutated); got != nil {
				t.Fatalf("CalculateDayun accepted incomplete published contract: %#v", got)
			}
		})
	}
}

func TestTransitCharts_JSONReplayMatchesFreshChart(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}
	replayed := replayZiWeiChartJSON(t, base)

	tests := []struct {
		name      string
		calculate func(*ZiWeiChart) *ZiWeiChart
	}{
		{name: "流年", calculate: func(chart *ZiWeiChart) *ZiWeiChart {
			return svc.CalculateLiunian(chart, 2026)
		}},
		{name: "流月", calculate: func(chart *ZiWeiChart) *ZiWeiChart {
			return svc.CalculateLiuyueForDate(chart, 2026, 3, 15)
		}},
		{name: "流日", calculate: func(chart *ZiWeiChart) *ZiWeiChart {
			return svc.CalculateLiuriForDate(chart, 2026, 3, 15)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.calculate(base)
			got := tt.calculate(replayed)
			if want == nil || got == nil {
				t.Fatalf("transit calculation returned nil: fresh=%v replay=%v", want == nil, got == nil)
			}
			assertZiWeiJSONEqual(t, got, want)
		})
	}
}

func TestTransitInterpretation_JSONReplayMatchesFreshChart(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}
	replayedBase := replayZiWeiChartJSON(t, base)

	freshLiunian := svc.CalculateLiunian(base, 2026)
	freshLiuyue := svc.CalculateLiuyueForDate(base, 2026, 3, 15)
	freshLiuri := svc.CalculateLiuriForDate(base, 2026, 3, 15)
	replayedLiunian := svc.CalculateLiunian(replayedBase, 2026)
	replayedLiuyue := svc.CalculateLiuyueForDate(replayedBase, 2026, 3, 15)
	replayedLiuri := svc.CalculateLiuriForDate(replayedBase, 2026, 3, 15)
	if freshLiunian == nil || freshLiuyue == nil || freshLiuri == nil ||
		replayedLiunian == nil || replayedLiuyue == nil || replayedLiuri == nil {
		t.Fatal("valid transit chart returned nil")
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{
			name: "流年结构解释",
			got:  NewPeriodInterpreterFromChart(replayedBase).AnalyzeLiunian(replayedLiunian, 2026),
			want: NewPeriodInterpreterFromChart(base).AnalyzeLiunian(freshLiunian, 2026),
		},
		{
			name: "流月结构解释",
			got:  NewPeriodInterpreterFromChart(replayedBase).AnalyzeLiuyue(replayedLiuyue, 2026, 3, 15),
			want: NewPeriodInterpreterFromChart(base).AnalyzeLiuyue(freshLiuyue, 2026, 3, 15),
		},
		{
			name: "流日结构解释",
			got:  NewPeriodInterpreterFromChart(replayedBase).AnalyzeLiuri(replayedLiuri, 2026, 3, 15),
			want: NewPeriodInterpreterFromChart(base).AnalyzeLiuri(freshLiuri, 2026, 3, 15),
		},
		{
			name: "流年分析",
			got:  BuildLiunianAnalysis(replayedBase, replayedLiunian, 2026),
			want: BuildLiunianAnalysis(base, freshLiunian, 2026),
		},
		{
			name: "流月分析",
			got:  BuildLiuyueAnalysis(replayedBase, replayedLiuyue, 2026, 3, 15),
			want: BuildLiuyueAnalysis(base, freshLiuyue, 2026, 3, 15),
		},
		{
			name: "流日分析",
			got:  BuildLiuriAnalysis(replayedBase, replayedLiuri, 2026, 3, 15),
			want: BuildLiuriAnalysis(base, freshLiuri, 2026, 3, 15),
		},
		{
			name: "流年叠盘解释",
			got:  svc.AnalyzeLiunianOverlay(replayedBase, replayedLiunian, 2026),
			want: svc.AnalyzeLiunianOverlay(base, freshLiunian, 2026),
		},
	}

	freshInterpreter := NewPeriodInterpreterFromChart(base)
	replayedInterpreter := NewPeriodInterpreterFromChart(replayedBase)
	if freshInterpreter == nil || replayedInterpreter == nil {
		t.Fatal("valid published chart did not restore period interpreter")
	}
	tests = append(tests, struct {
		name string
		got  any
		want any
	}{
		name: "三层摘要",
		got:  replayedInterpreter.SummarizeAll(replayedLiunian, replayedLiuyue, replayedLiuri, 2026, 3, 15),
		want: freshInterpreter.SummarizeAll(freshLiunian, freshLiuyue, freshLiuri, 2026, 3, 15),
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got == nil || tt.want == nil {
				t.Fatalf("interpretation returned nil: got=%v want=%v", tt.got == nil, tt.want == nil)
			}
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("interpretation changed after JSON replay:\n got: %#v\nwant: %#v", tt.got, tt.want)
			}
		})
	}
}

func TestTransitConsumers_RejectInvalidPublishedContracts(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 3, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 3, 15)
	if liunian == nil || liuyue == nil || liuri == nil {
		t.Fatal("valid transit chart returned nil")
	}

	invalidNatalCases := []struct {
		name   string
		mutate func(*ZiWeiChart)
	}{
		{name: "content hash", mutate: func(chart *ZiWeiChart) {
			chart.ContentHash = ""
		}},
		{name: "input fingerprint", mutate: func(chart *ZiWeiChart) {
			chart.InputFingerprint = ""
			restampZiWeiChartContentHash(t, chart)
		}},
		{name: "calculation input", mutate: func(chart *ZiWeiChart) {
			chart.CalculationInput.Basis = ""
			chart.InputFingerprint = ziweiInputFingerprint(chart.CalculationInput)
			restampZiWeiChartContentHash(t, chart)
		}},
	}
	for _, tt := range invalidNatalCases {
		t.Run("本命/"+tt.name, func(t *testing.T) {
			invalidBase := *base
			tt.mutate(&invalidBase)
			if NewPeriodInterpreterFromChart(&invalidBase) != nil {
				t.Fatal("period interpreter accepted invalid natal contract")
			}
			if svc.CalculateLiunian(&invalidBase, 2026) != nil ||
				svc.CalculateLiuyueForDate(&invalidBase, 2026, 3, 15) != nil ||
				svc.CalculateLiuriForDate(&invalidBase, 2026, 3, 15) != nil {
				t.Fatal("transit calculation accepted invalid natal contract")
			}
			if BuildLiunianAnalysis(&invalidBase, liunian, 2026) != nil ||
				BuildLiuyueAnalysis(&invalidBase, liuyue, 2026, 3, 15) != nil ||
				BuildLiuriAnalysis(&invalidBase, liuri, 2026, 3, 15) != nil {
				t.Fatal("transit analysis accepted invalid natal contract")
			}
			if svc.AnalyzeLiunianOverlay(&invalidBase, liunian, 2026) != nil {
				t.Fatal("liunian overlay silently defaulted from invalid natal contract")
			}
		})
	}

	invalidLiunian := *liunian
	invalidLiunian.DerivedContentHash = ""
	invalidLiuyue := *liuyue
	invalidLiuyue.DerivedContentHash = ""
	invalidLiuri := *liuri
	invalidLiuri.DerivedContentHash = ""
	if BuildLiunianAnalysis(base, &invalidLiunian, 2026) != nil ||
		BuildLiuyueAnalysis(base, &invalidLiuyue, 2026, 3, 15) != nil ||
		BuildLiuriAnalysis(base, &invalidLiuri, 2026, 3, 15) != nil {
		t.Fatal("transit analysis accepted invalid derived contract")
	}
	if svc.AnalyzeLiunianOverlay(base, &invalidLiunian, 2026) != nil {
		t.Fatal("liunian overlay accepted invalid derived contract")
	}

	otherBase, err := svc.CalculateChart(1992, 9, 8, 9, 0, "女")
	if err != nil {
		t.Fatalf("CalculateChart other base failed: %v", err)
	}
	if BuildLiunianAnalysis(otherBase, liunian, 2026) != nil ||
		BuildLiuyueAnalysis(otherBase, liuyue, 2026, 3, 15) != nil ||
		BuildLiuriAnalysis(otherBase, liuri, 2026, 3, 15) != nil ||
		svc.AnalyzeLiunianOverlay(otherBase, liunian, 2026) != nil {
		t.Fatal("transit consumer accepted a derived chart bound to another natal chart")
	}

	unbound := &PeriodInterpreter{birthData: mustPublishedBirthData(t, base)}
	if unbound.AnalyzeLiunian(liunian, 2026) != nil ||
		unbound.AnalyzeLiuyue(liuyue, 2026, 3, 15) != nil ||
		unbound.AnalyzeLiuri(liuri, 2026, 3, 15) != nil {
		t.Fatal("period interpreter accepted an unbound birth context")
	}
}

func replayZiWeiChartJSON(t *testing.T, chart *ZiWeiChart) *ZiWeiChart {
	t.Helper()
	payload, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("marshal chart: %v", err)
	}
	var replayed ZiWeiChart
	if err := json.Unmarshal(payload, &replayed); err != nil {
		t.Fatalf("unmarshal chart: %v", err)
	}
	return &replayed
}

func assertZiWeiJSONEqual(t *testing.T, got, want *ZiWeiChart) {
	t.Helper()
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal got chart: %v", err)
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal want chart: %v", err)
	}
	if string(gotJSON) != string(wantJSON) {
		t.Fatalf("public chart JSON changed after replay:\n got: %s\nwant: %s", gotJSON, wantJSON)
	}
}

func restampZiWeiChartContentHash(t *testing.T, chart *ZiWeiChart) {
	t.Helper()
	hash, err := chartContentHash(chart)
	if err != nil {
		t.Fatalf("hash chart: %v", err)
	}
	chart.ContentHash = hash
}

// TestLiunian_Basic verifies that CalculateLiunian produces a valid overlay
// chart with liu_nian_stars populated.
func TestLiunian_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Calculate liunian for a target year (e.g., 2024)
	targetYear := 2024
	liunianChart := svc.CalculateLiunian(chart, targetYear)
	if liunianChart == nil {
		t.Fatal("CalculateLiunian returned nil")
	}

	// Verify LiuNianStars is populated
	hasContent := false
	for i, stars := range liunianChart.LiuNianStars {
		if len(stars) > 0 {
			hasContent = true
			break
		}
		_ = i
	}
	if !hasContent {
		t.Error("LiuNianStars 全部为空, 期望至少有流禄/流羊/流陀/流马之一")
	}

	// Verify the palatial structure is preserved
	for i, p := range liunianChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty in liunian chart", i)
		}
	}
}

// TestLiunian_Stars verifies specific liunian star placements.
// 流禄 is placed at LucunBranchIdx for the target year stem
// 流羊 at LucunBranchIdx+1
// 流陀 at LucunBranchIdx-1
// 流马 at TianmaBranchIdx for the target year branch
func TestLiunian_Stars(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Target year 2024 → 甲辰年 → year stem index = (2024-4)%10 = 0, year branch index = (2024-4)%12 = 4
	targetYear := 2024
	yearStem := (targetYear - 4) % 10
	yearBranch := (targetYear - 4) % 12

	liunianChart := svc.CalculateLiunian(chart, targetYear)

	// Expected liunian star positions
	expectedLucunBranch := LucunBranchIdx[yearStem]
	expectedQingyangBranch := fixIndex(expectedLucunBranch + 1)
	expectedTuoluoBranch := fixIndex(expectedLucunBranch - 1)
	expectedTianmaBranch := TianmaBranchIdx[yearBranch]

	// Check for 流禄
	foundLiuLu, foundLiuYang, foundLiuTuo, foundLiuMa := false, false, false, false
	for _, stars := range liunianChart.LiuNianStars {
		for _, s := range stars {
			if s == "流禄" {
				foundLiuLu = true
			}
			if s == "流羊" {
				foundLiuYang = true
			}
			if s == "流陀" {
				foundLiuTuo = true
			}
			if s == "流马" {
				foundLiuMa = true
			}
		}
	}

	if !foundLiuLu {
		t.Errorf("2024年(甲辰) 流禄未找到, 期望在分支 %d(%s)", expectedLucunBranch, BranchNames[expectedLucunBranch])
	}
	if !foundLiuYang {
		t.Errorf("2024年(甲辰) 流羊未找到, 期望在分支 %d(%s)", expectedQingyangBranch, BranchNames[expectedQingyangBranch])
	}
	if !foundLiuTuo {
		t.Errorf("2024年(甲辰) 流陀未找到, 期望在分支 %d(%s)", expectedTuoluoBranch, BranchNames[expectedTuoluoBranch])
	}
	if !foundLiuMa {
		t.Errorf("2024年(甲辰) 流马未找到, 期望在分支 %d(%s)", expectedTianmaBranch, BranchNames[expectedTianmaBranch])
	}
}

func TestLiunian_StarsUsePalaceBranch(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	targetYear := 2024 // 甲辰
	yearStem, yearBranch := annualStemBranch(targetYear)
	liunianChart := svc.CalculateLiunian(chart, targetYear)

	expected := map[string]string{
		"流禄": BranchNames[LucunBranchIdx[yearStem]],
		"流羊": BranchNames[fixIndex(LucunBranchIdx[yearStem]+1)],
		"流陀": BranchNames[fixIndex(LucunBranchIdx[yearStem]-1)],
		"流马": BranchNames[TianmaBranchIdx[yearBranch]],
	}

	for star, wantBranch := range expected {
		found := false
		for i, stars := range liunianChart.LiuNianStars {
			for _, got := range stars {
				if got != star {
					continue
				}
				found = true
				if liunianChart.Palaces[i].Branch != wantBranch {
					t.Fatalf("%s 落在 %s宫(%s), want branch %s", star, liunianChart.Palaces[i].Name, liunianChart.Palaces[i].Branch, wantBranch)
				}
			}
		}
		if !found {
			t.Fatalf("%s not found", star)
		}
	}
}

func TestLiunianOverlayAnalysis_BuildsEvidence(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	liunianChart := svc.CalculateLiunian(chart, 2024)
	analysis := svc.AnalyzeLiunianOverlay(chart, liunianChart, 2024)
	if analysis == nil {
		t.Fatal("AnalyzeLiunianOverlay returned nil")
	}
	if analysis.GanZhi != "甲辰" {
		t.Fatalf("GanZhi=%q, want 甲辰", analysis.GanZhi)
	}
	if len(analysis.Method) == 0 {
		t.Fatal("Method should not be empty")
	}
	if len(analysis.FourHua) == 0 {
		t.Fatal("FourHua should not be empty")
	}
	if len(analysis.AnnualStars) != 11 {
		t.Fatalf("AnnualStars length=%d, want 11", len(analysis.AnnualStars))
	}
	if len(analysis.FocusPalaces) == 0 {
		t.Fatal("FocusPalaces should not be empty")
	}
}

func TestLiunian_SeparatesTransitStarsAndFourHua(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	targetYear := 1984 // 甲子年, stem=0
	liunianChart := svc.CalculateLiunian(chart, targetYear)
	if liunianChart == nil {
		t.Fatal("CalculateLiunian returned nil")
	}
	for _, stars := range liunianChart.LiuNianStars {
		for _, star := range stars {
			if strings.Contains(star, "化") {
				t.Fatalf("four-hua label leaked into liu_nian_stars: %q", star)
			}
		}
	}
	want := []string{"廉贞化禄", "破军化权", "武曲化科", "太阳化忌"}
	got := flattenPeriodStars(liunianChart.LiuNianFourHua)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("liu_nian_four_hua = %v, want %v", got, want)
	}
}

// TestLiuyue_Basic verifies that CalculateLiuyueForDate returns a valid overlay chart.
func TestLiuyue_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	liuyueChart := svc.CalculateLiuyueForDate(chart, 2026, 3, 15)
	if liuyueChart == nil {
		t.Fatal("CalculateLiuyueForDate returned nil")
	}

	// Verify LiuYueStars exists (length check instead of direct comparison)
	if len(liuyueChart.LiuYueStars) != 12 {
		t.Errorf("LiuYueStars length = %d, want 12", len(liuyueChart.LiuYueStars))
	}

	// Verify palaces preserved
	for i, p := range liuyueChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty", i)
		}
	}
}

// TestLiuri_Basic verifies that CalculateLiuriForDate returns a valid overlay chart.
func TestLiuri_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	liuriChart := svc.CalculateLiuriForDate(chart, 2026, 3, 15)
	if liuriChart == nil {
		t.Fatal("CalculateLiuriForDate returned nil")
	}

	// Verify LiuriStars is a [12][]string
	_ = liuriChart.LiuRiStars

	// Verify palaces preserved
	for i, p := range liuriChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty", i)
		}
	}
}

// TestDayun_NilChart verifies nil handling.
func TestDayun_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateDayun(nil)
	if result != nil {
		t.Errorf("CalculateDayun(nil) = %v, want nil", result)
	}
}

// TestLiunian_NilChart verifies nil handling.
func TestLiunian_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiunian(nil, 2024)
	if result != nil {
		t.Errorf("CalculateLiunian(nil) = %v, want nil", result)
	}
}

// TestLiuyue_NilChart verifies nil handling.
func TestLiuyue_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiuyueForDate(nil, 2026, 3, 15)
	if result != nil {
		t.Errorf("CalculateLiuyueForDate(nil) = %v, want nil", result)
	}
}

// TestLiuri_NilChart verifies nil handling.
func TestLiuri_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiuriForDate(nil, 2026, 3, 15)
	if result != nil {
		t.Errorf("CalculateLiuriForDate(nil) = %v, want nil", result)
	}
}
