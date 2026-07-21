package ziwei

import "testing"

func TestChartMatchesProfileRejectsRehashedInvalidStructure(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart: %v", err)
	}
	if !svc.ChartMatchesProfile(base, DefaultProfileID) {
		t.Fatal("fresh chart must satisfy the published structure contract")
	}
	replayed := roundTripProjectionFixture(t, base)
	if !svc.ChartMatchesProfile(replayed, DefaultProfileID) {
		t.Fatal("JSON-replayed chart must satisfy the published structure contract")
	}

	tests := []struct {
		name   string
		mutate func(*ZiWeiChart)
	}{
		{name: "palace order", mutate: func(chart *ZiWeiChart) {
			chart.Palaces[0], chart.Palaces[1] = chart.Palaces[1], chart.Palaces[0]
		}},
		{name: "palace stem", mutate: func(chart *ZiWeiChart) {
			chart.Palaces[0].HeavenlyStem = nextStem(chart.Palaces[0].HeavenlyStem)
		}},
		{name: "body marker", mutate: func(chart *ZiWeiChart) {
			for i := range chart.Palaces {
				chart.Palaces[i].IsBodyPalace = false
			}
		}},
		{name: "life master", mutate: func(chart *ZiWeiChart) {
			chart.LifeMaster = "篡改命主"
		}},
		{name: "five bureau", mutate: func(chart *ZiWeiChart) {
			chart.FiveBureau = "篡改五行局"
		}},
		{name: "major star position", mutate: movePublishedMajorStar},
		{name: "star brightness", mutate: func(chart *ZiWeiChart) {
			palaceIdx, starIdx := findPublishedStar(chart, "紫微")
			chart.Palaces[palaceIdx].Stars[starIdx].Brightness = "篡改亮度"
		}},
		{name: "four hua", mutate: func(chart *ZiWeiChart) {
			for i := range chart.Palaces {
				if len(chart.Palaces[i].FourHua) == 0 {
					continue
				}
				chart.Palaces[i].FourHua[0] = "篡改化曜"
				return
			}
			panic("fixture has no four hua")
		}},
		{name: "adjective star", mutate: func(chart *ZiWeiChart) {
			chart.Palaces[0].AdjectiveStars = append(chart.Palaces[0].AdjectiveStars, "篡改杂曜")
		}},
		{name: "twelve shen", mutate: func(chart *ZiWeiChart) {
			chart.Palaces[0].Boshi12 = "篡改博士十二神"
		}},
		{name: "sanfang", mutate: func(chart *ZiWeiChart) {
			chart.SanfangSizheng[0].Opposite = "篡改对宫"
		}},
		{name: "patterns", mutate: func(chart *ZiWeiChart) {
			chart.Patterns = append(chart.Patterns, "篡改格局")
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := roundTripProjectionFixture(t, base)
			tc.mutate(chart)
			restampZiWeiChartContentHash(t, chart)
			if !validChartContentHash(chart) {
				t.Fatal("fixture must have a valid recomputed content hash")
			}
			if svc.ChartMatchesProfile(chart, DefaultProfileID) ||
				svc.ChartMatchesInputProfile(chart, DefaultProfileID, 2003, 4, 15, 14, 0, "男") {
				t.Fatal("rehashed structurally invalid chart must not match the current profile")
			}
		})
	}
}

func TestPublishedChartConsumersRejectRehashedInvalidStructure(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CalculateChart(1984, 2, 15, 8, 0, "女")
	if err != nil {
		t.Fatal(err)
	}
	dayun := svc.CalculateDayun(base)
	if dayun == nil || BuildDayunAnalysis(base, dayun, 23) == nil || svc.AnalyzeFlyingStars(base) == nil ||
		svc.AnalyzeSihuaChain(base) == nil || svc.GetPalaceReading(base, 0) == nil ||
		svc.BuildQueryView(base) == nil || svc.AnalyzeHeming(base, second) == nil {
		t.Fatal("valid published charts must be accepted by all structural consumers")
	}
	forgedDayun := append(Dayun(nil), dayun...)
	forgedDayun[0].StartAge++
	if BuildDayunAnalysis(base, forgedDayun, 23) != nil {
		t.Fatal("dayun analysis accepted stages that do not match the authenticated chart")
	}
	liunian := svc.CalculateLiunian(base, 2026)
	interpreter := NewPeriodInterpreterFromChart(base)
	otherInterpreter := NewPeriodInterpreterFromChart(second)
	if liunian == nil || interpreter == nil || interpreter.AnalyzeLiunian(liunian, 2026) == nil {
		t.Fatal("valid chart-bound period interpreter rejected its own derivation")
	}
	if otherInterpreter == nil || otherInterpreter.AnalyzeLiunian(liunian, 2026) != nil {
		t.Fatal("period interpreter accepted a derivation bound to another natal chart")
	}

	invalid := roundTripProjectionFixture(t, base)
	movePublishedMajorStar(invalid)
	restampZiWeiChartContentHash(t, invalid)
	if svc.DetectLocalPatterns(invalid) != nil || svc.AnalyzeFlyingStars(invalid) != nil ||
		svc.CalculateDayun(invalid) != nil || svc.AnalyzeSihuaChain(invalid) != nil ||
		svc.GetPalaceReading(invalid, 0) != nil || svc.BuildQueryView(invalid) != nil ||
		svc.AnalyzeHeming(invalid, second) != nil || svc.AnalyzeSelfMutagen(invalid) != nil ||
		BuildDayunAnalysis(invalid, nil, 23) != nil || NewPeriodInterpreterFromChart(invalid) != nil {
		t.Fatal("rehashed invalid chart reached a published structural consumer")
	}
}

func nextStem(stem string) string {
	idx, ok := StemIndex[stem]
	if !ok {
		return "甲"
	}
	return StemNames[(idx+1)%len(StemNames)]
}

func findPublishedStar(chart *ZiWeiChart, name string) (palaceIdx, starIdx int) {
	for i := range chart.Palaces {
		for j := range chart.Palaces[i].Stars {
			if chart.Palaces[i].Stars[j].Name == name {
				return i, j
			}
		}
	}
	panic("published star not found: " + name)
}

func movePublishedMajorStar(chart *ZiWeiChart) {
	from, starIdx := findPublishedStar(chart, "紫微")
	star := chart.Palaces[from].Stars[starIdx]
	chart.Palaces[from].Stars = append(
		chart.Palaces[from].Stars[:starIdx],
		chart.Palaces[from].Stars[starIdx+1:]...,
	)
	to := (from + 1) % len(chart.Palaces)
	chart.Palaces[to].Stars = append(chart.Palaces[to].Stars, star)
}
