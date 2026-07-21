package ziwei

import (
	"reflect"
	"testing"
)

// ════════════════════════════════════════════════════════════════
// Tests for the single authoritative SiHuaTable in ziwei_data.go.
// ════════════════════════════════════════════════════════════════

// TestSiHuaTable_AllStems verifies calcFourHua for all 10 heavenly stems.
func TestSiHuaTable_AllStems(t *testing.T) {
	// Expected four transformations for each stem (0=甲 through 9=癸)
	// Order: [化禄, 化权, 化科, 化忌]
	expected := [10][4]string{
		{"廉贞", "破军", "武曲", "太阳"}, // 甲
		{"天机", "天梁", "紫微", "太阴"}, // 乙
		{"天同", "天机", "文昌", "廉贞"}, // 丙
		{"太阴", "天同", "天机", "巨门"}, // 丁
		{"贪狼", "太阴", "右弼", "天机"}, // 戊
		{"武曲", "贪狼", "天梁", "文曲"}, // 己
		{"太阳", "武曲", "太阴", "天同"}, // 庚
		{"巨门", "太阳", "文曲", "文昌"}, // 辛
		{"天梁", "紫微", "左辅", "武曲"}, // 壬
		{"破军", "巨门", "太阴", "贪狼"}, // 癸
	}
	if SiHuaTable != expected {
		t.Fatalf("SiHuaTable = %v, want %v", SiHuaTable, expected)
	}

	labels := []string{"化禄", "化权", "化科", "化忌"}

	for stem := 0; stem < 10; stem++ {
		t.Run(StemNames[stem], func(t *testing.T) {
			hua := calcFourHua(stem)
			if len(hua) != 4 {
				t.Fatalf("calcFourHua(%d) returned %d entries, want 4", stem, len(hua))
			}
			for i, wantStar := range expected[stem] {
				gotLabel := hua[wantStar]
				if gotLabel != labels[i] {
					t.Errorf("calcFourHua(%d)[%s] = %q, want %q", stem, wantStar, gotLabel, labels[i])
				}
			}
		})
	}
}

func TestSiHuaTable_SourcePinned(t *testing.T) {
	if SiHuaRuleID != "ziwei.sihua.ten-stem.iztro-v1" ||
		SiHuaSourceRepo != "https://github.com/SylarLong/iztro" ||
		SiHuaSourceCommit != "2dfe3ecb41d725b2bea1084bbdfe4dd655e37b13" ||
		SiHuaSourcePath != "src/data/heavenlyStems.ts" {
		t.Fatalf("four-hua source metadata is not pinned: %s %s %s %s", SiHuaRuleID, SiHuaSourceRepo, SiHuaSourceCommit, SiHuaSourcePath)
	}
}

func TestSiHuaTable_AllConsumersAgree(t *testing.T) {
	chart := chartContainingAllFourHuaStars()
	for stem := 0; stem < 10; stem++ {
		t.Run(StemNames[stem], func(t *testing.T) {
			want := SiHuaTable[stem]
			chart.Palaces[0].HeavenlyStem = StemNames[stem]
			stampPublishedFourHua(chart, stem)

			chain := analyzeSihuaChain(chart)
			if chain == nil || len(chain.HuaLu) == 0 || len(chain.HuaQuan) == 0 || len(chain.HuaKe) == 0 || len(chain.HuaJi) == 0 {
				t.Fatalf("AnalyzeSihuaChain(%s) omitted a transformation: %+v", StemNames[stem], chain)
			}
			if got := [4]string{chain.HuaLu[0].TransformedStar, chain.HuaQuan[0].TransformedStar, chain.HuaKe[0].TransformedStar, chain.HuaJi[0].TransformedStar}; got != want {
				t.Errorf("chain = %v, want %v", got, want)
			}

			flying := buildFlyingStarAnalysisFromChart(chart)
			if flying == nil || len(flying.HuaLu) == 0 || len(flying.HuaQuan) == 0 || len(flying.HuaKe) == 0 || len(flying.HuaJi) == 0 {
				t.Fatalf("flying-star analysis omitted a transformation: %+v", flying)
			}
			if got := [4]string{flying.HuaLu[0].TransformedStar, flying.HuaQuan[0].TransformedStar, flying.HuaKe[0].TransformedStar, flying.HuaJi[0].TransformedStar}; got != want {
				t.Errorf("flying = %v, want %v", got, want)
			}

			overlay := buildLiunianFourHuaTriggers(chart, stem)
			if len(overlay) != 4 {
				t.Fatalf("overlay transformations = %d, want 4: %+v", len(overlay), overlay)
			}
			if got := [4]string{overlay[0].Star, overlay[1].Star, overlay[2].Star, overlay[3].Star}; got != want {
				t.Errorf("overlay = %v, want %v", got, want)
			}
		})
	}
}

// TestSiHuaTable_Length verifies all entries have exactly 4 items.
func TestSiHuaTable_Length(t *testing.T) {
	for stem := 0; stem < 10; stem++ {
		if len(SiHuaTable[stem]) != 4 {
			t.Errorf("SiHuaTable[%d] has length %d, want 4", stem, len(SiHuaTable[stem]))
		}
	}
}

func chartContainingAllFourHuaStars() *ZiWeiChart {
	stars := []string{
		"廉贞", "破军", "武曲", "太阳", "天机", "天梁", "紫微", "太阴",
		"天同", "文昌", "巨门", "贪狼", "右弼", "文曲", "左辅",
	}
	chart := &ZiWeiChart{}
	for i := range chart.Palaces {
		chart.Palaces[i].Name = PALACE_NAMES[i]
		chart.Palaces[i].Branch = BranchNames[i]
	}
	for i, star := range stars {
		idx := i % len(chart.Palaces)
		starType := "major"
		if star == "文昌" || star == "右弼" || star == "文曲" || star == "左辅" {
			starType = "soft"
		}
		chart.Palaces[idx].Stars = append(chart.Palaces[idx].Stars, StarOutput{Name: star, Type: starType, Scope: "origin"})
	}
	return chart
}

func stampPublishedFourHua(chart *ZiWeiChart, stem int) {
	for i := range chart.Palaces {
		chart.Palaces[i].FourHua = nil
	}
	starToPalace := buildStarPalaceIndex(chart)
	for huaIdx, star := range SiHuaTable[stem] {
		palaceIdx := starToPalace[star]
		chart.Palaces[palaceIdx].FourHua = append(chart.Palaces[palaceIdx].FourHua, star+SiHuaLabels[huaIdx])
	}
}

// TestSiHuaLabels verifies SiHuaLabels contains the correct 4 labels.
func TestSiHuaLabels(t *testing.T) {
	want := [4]string{"化禄", "化权", "化科", "化忌"}
	if SiHuaLabels != want {
		t.Errorf("SiHuaLabels = %v, want %v", SiHuaLabels, want)
	}
}

// ════════════════════════════════════════════════════════════════
// Tests for getFourHuaInPalace
// ════════════════════════════════════════════════════════════════

// TestGetFourHuaInPalace verifies that the correct four hua strings are
// extracted from a palace's major and auxiliary stars.
func TestGetFourHuaInPalace(t *testing.T) {
	tests := []struct {
		name   string
		major  []StarInfo
		aux    []StarInfo
		huaMap map[string]string
		want   []string
	}{
		{
			name:   "no matching stars",
			major:  []StarInfo{{Name: "紫微"}, {Name: "天府"}},
			aux:    []StarInfo{{Name: "左辅"}},
			huaMap: map[string]string{"廉贞": "化禄", "破军": "化权"},
			want:   []string{},
		},
		{
			name:   "one major star matches",
			major:  []StarInfo{{Name: "廉贞"}, {Name: "天府"}},
			aux:    []StarInfo{{Name: "左辅"}},
			huaMap: map[string]string{"廉贞": "化禄", "破军": "化权"},
			want:   []string{"廉贞化禄"},
		},
		{
			name:   "major and aux both match",
			major:  []StarInfo{{Name: "廉贞"}, {Name: "天机"}},
			aux:    []StarInfo{{Name: "文昌"}},
			huaMap: map[string]string{"廉贞": "化禄", "文昌": "化科", "天机": "化权"},
			want:   []string{"廉贞化禄", "天机化权", "文昌化科"},
		},
		{
			name:   "empty inputs",
			major:  []StarInfo{},
			aux:    []StarInfo{},
			huaMap: map[string]string{"廉贞": "化禄"},
			want:   []string{},
		},
		{
			name:   "aux only matches",
			major:  []StarInfo{{Name: "紫微"}},
			aux:    []StarInfo{{Name: "右弼"}, {Name: "文曲"}},
			huaMap: map[string]string{"右弼": "化科"},
			want:   []string{"右弼化科"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getFourHuaInPalace(tt.major, tt.aux, tt.huaMap)

			// Use slice equality check that handles nil vs empty slice correctly
			if !stringSlicesEqual(got, tt.want) {
				t.Errorf("getFourHuaInPalace() = %v (len=%d, nil=%v), want %v (len=%d, nil=%v)",
					got, len(got), got == nil, tt.want, len(tt.want), tt.want == nil)
			}
		})
	}
}

// stringSlicesEqual compares two string slices, treating nil and empty slices as equal.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ════════════════════════════════════════════════════════════════
// Tests for applyFourHua
// ════════════════════════════════════════════════════════════════

// TestApplyFourHua verifies the full-palace four hua application.
func TestApplyFourHua(t *testing.T) {
	// Create known star placements for all 12 palaces
	var majorStars [12][]StarInfo
	var auxStars [12][]StarInfo

	// Palace 0: major=廉贞 (化禄 in 甲年), aux=none
	majorStars[0] = []StarInfo{{Name: "廉贞"}}
	// Palace 1: major=破军 (化权 in 甲年)
	majorStars[1] = []StarInfo{{Name: "破军"}}
	// Palace 2: major=武曲 (化科 in 甲年), aux=文昌 (not transformed in 甲年)
	majorStars[2] = []StarInfo{{Name: "武曲"}}
	auxStars[2] = []StarInfo{{Name: "文昌"}}
	// Palace 3: major=太阳 (化忌 in 甲年)
	majorStars[3] = []StarInfo{{Name: "太阳"}}
	// Palace 5: major=紫微 (not transformed in 甲年)
	majorStars[5] = []StarInfo{{Name: "紫微"}}
	// Palace 7: aux=文曲 (not transformed in 甲年)
	auxStars[7] = []StarInfo{{Name: "文曲"}}

	// Apply 甲年 (stem 0) transformations
	result := applyFourHua(majorStars, auxStars, 0)

	// Check that Mutagen was set correctly on stars
	if majorStars[0][0].Mutagen != "化禄" {
		t.Errorf("majorStars[0][0].Mutagen = %q, want %q", majorStars[0][0].Mutagen, "化禄")
	}
	if majorStars[1][0].Mutagen != "化权" {
		t.Errorf("majorStars[1][0].Mutagen = %q, want %q", majorStars[1][0].Mutagen, "化权")
	}
	if majorStars[2][0].Mutagen != "化科" {
		t.Errorf("majorStars[2][0].Mutagen = %q, want %q", majorStars[2][0].Mutagen, "化科")
	}
	if majorStars[3][0].Mutagen != "化忌" {
		t.Errorf("majorStars[3][0].Mutagen = %q, want %q", majorStars[3][0].Mutagen, "化忌")
	}
	// Stars not in SiHuaTable should NOT have Mutagen set
	if majorStars[5][0].Mutagen != "" {
		t.Errorf("majorStars[5][0] (紫微) should have empty Mutagen, got %q", majorStars[5][0].Mutagen)
	}
	if auxStars[7][0].Mutagen != "" {
		t.Errorf("auxStars[7][0] (文曲) should have empty Mutagen in 甲年, got %q", auxStars[7][0].Mutagen)
	}

	// Check the result per-palace: each palace gets a slice of "StarName化X" strings
	wantResult := func() [12][]string {
		var r [12][]string
		r[0] = []string{"廉贞化禄"}
		r[1] = []string{"破军化权"}
		r[2] = []string{"武曲化科"}
		r[3] = []string{"太阳化忌"}
		return r
	}()
	if !reflect.DeepEqual(result, wantResult) {
		t.Errorf("applyFourHua() result = %v, want %v", result, wantResult)
	}
}

// ════════════════════════════════════════════════════════════════
// Tests for SihuaChainResult / AnalyzeSihuaChain
// ════════════════════════════════════════════════════════════════

// TestAnalyzeSihuaChain_NilChart verifies nil handling.
func TestAnalyzeSihuaChain_NilChart(t *testing.T) {
	result := analyzeSihuaChain(nil)
	if result != nil {
		t.Errorf("AnalyzeSihuaChain(nil) = %v, want nil", result)
	}
}

// TestAnalyzeSihuaChain_WithEmptySyntheticChart verifies the pure projection
// helper against an explicitly synthetic chart with no palace stems.
func TestAnalyzeSihuaChain_WithEmptySyntheticChart(t *testing.T) {
	chart := &ZiWeiChart{}
	result := analyzeSihuaChain(chart)
	if result == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	if len(result.HuaLu) != 0 {
		t.Errorf("expected 0 HuaLu items, got %d: %v", len(result.HuaLu), result.HuaLu)
	}
	if len(result.HuaQuan) != 0 {
		t.Errorf("expected 0 HuaQuan items, got %d: %v", len(result.HuaQuan), result.HuaQuan)
	}
	if len(result.HuaKe) != 0 {
		t.Errorf("expected 0 HuaKe items, got %d: %v", len(result.HuaKe), result.HuaKe)
	}
	if len(result.HuaJi) != 0 {
		t.Errorf("expected 0 HuaJi items, got %d: %v", len(result.HuaJi), result.HuaJi)
	}
	assertSihuaProjectionSemantics(t, result.SihuaProjectionSemantics)
	if result.AnalysisKind != sihuaDirectFlightAnalysisKind {
		t.Errorf("analysis_kind = %q, want %q", result.AnalysisKind, sihuaDirectFlightAnalysisKind)
	}
}

// TestAnalyzeSihuaChain_WithKnownTransformations creates a chart where we
// know exactly which stars are transformed and verifies they're found.
func TestAnalyzeSihuaChain_WithKnownTransformations(t *testing.T) {
	// Build a chart where palace 0 (命宫) contains 廉贞 (化禄 in 甲年)
	// and palace 3 (子女) contains 破军 (化权 in 甲年)
	var palaces [12]PalaceInfo
	for i := range palaces {
		palaces[i] = PalaceInfo{
			Name:         PALACE_NAMES[i],
			HeavenlyStem: "",
			Stars:        []StarOutput{},
			FourHua:      []string{},
		}
	}
	palaces[0].HeavenlyStem = "甲"

	// Place 廉贞 in 命宫 (index 0) — this is 化禄 in 甲年
	palaces[0].Stars = []StarOutput{{Name: "廉贞", Type: "major", Scope: "origin"}}
	// Place 破军 in 子女 (index 3) — this is 化权 in 甲年
	palaces[3].Stars = []StarOutput{{Name: "破军", Type: "major", Scope: "origin"}}

	chart := &ZiWeiChart{
		Palaces: palaces,
	}

	result := analyzeSihuaChain(chart)
	if result == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	// Should find 廉贞化禄 pointing to 命宫
	if len(result.HuaLu) != 1 {
		t.Fatalf("expected 1 HuaLu item, got %d: %+v", len(result.HuaLu), result.HuaLu)
	}
	if result.HuaLu[0].TransformedStar != "廉贞" {
		t.Errorf("HuaLu[0].TransformedStar = %q, want %q", result.HuaLu[0].TransformedStar, "廉贞")
	}
	if result.HuaLu[0].TargetPalace != "命宫" {
		t.Errorf("HuaLu[0].TargetPalace = %q, want %q", result.HuaLu[0].TargetPalace, "命宫")
	}
	if result.HuaLu[0].SourcePalace != "命宫" {
		t.Errorf("HuaLu[0].SourcePalace = %q, want %q", result.HuaLu[0].SourcePalace, "命宫")
	}
	if !result.HuaLu[0].IsSelfMutagen {
		t.Errorf("HuaLu[0] should be a palace-stem self mutagen: %+v", result.HuaLu[0])
	}

	// Should find 破军化权 pointing to 子女
	if len(result.HuaQuan) != 1 {
		t.Fatalf("expected 1 HuaQuan item, got %d: %+v", len(result.HuaQuan), result.HuaQuan)
	}
	if result.HuaQuan[0].TransformedStar != "破军" {
		t.Errorf("HuaQuan[0].TransformedStar = %q, want %q", result.HuaQuan[0].TransformedStar, "破军")
	}
	if result.HuaQuan[0].TargetPalace != "子女" {
		t.Errorf("HuaQuan[0].TargetPalace = %q, want %q", result.HuaQuan[0].TargetPalace, "子女")
	}
	if result.HuaQuan[0].SourcePalace != "命宫" || result.HuaQuan[0].IsSelfMutagen {
		t.Errorf("HuaQuan[0] should fly from 命宫 to 子女: %+v", result.HuaQuan[0])
	}

	// 武曲化科 and 太阳化忌 should exist in the table but have no chart stars → empty results
	if len(result.HuaKe) != 0 {
		t.Errorf("expected 0 HuaKe items (武曲 not placed), got %d", len(result.HuaKe))
	}
	if len(result.HuaJi) != 0 {
		t.Errorf("expected 0 HuaJi items (太阳 not placed), got %d", len(result.HuaJi))
	}
}

// TestAnalyzeSihuaChain_WithAuxStars tests chain analysis finding hua in aux stars.
func TestAnalyzeSihuaChain_WithAuxStars(t *testing.T) {
	var palaces [12]PalaceInfo
	for i := range palaces {
		palaces[i] = PalaceInfo{
			Name:         PALACE_NAMES[i],
			HeavenlyStem: "",
			Stars:        []StarOutput{},
			FourHua:      []string{},
		}
	}
	palaces[0].HeavenlyStem = "己"
	// 己年 (stem 5): 武曲化禄, 贪狼化权, 天梁化科, 文曲化忌
	// Place 文曲 as an aux star in 交友 (index 7)
	palaces[7].Stars = []StarOutput{{Name: "文曲", Type: "soft", Scope: "origin"}}
	// Place 天梁 as a main star in 迁移 (index 6)
	palaces[6].Stars = []StarOutput{{Name: "天梁", Type: "major", Scope: "origin"}}

	chart := &ZiWeiChart{
		Palaces: palaces,
	}

	result := analyzeSihuaChain(chart)
	if result == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	// 天梁化科 should point to 迁移
	if len(result.HuaKe) != 1 {
		t.Fatalf("expected 1 HuaKe item for 天梁化科, got %d: %+v", len(result.HuaKe), result.HuaKe)
	}
	if result.HuaKe[0].TransformedStar != "天梁" {
		t.Errorf("HuaKe[0].TransformedStar = %q, want %q", result.HuaKe[0].TransformedStar, "天梁")
	}
	if result.HuaKe[0].TargetPalace != "迁移" {
		t.Errorf("HuaKe[0].TargetPalace = %q, want %q", result.HuaKe[0].TargetPalace, "迁移")
	}

	// 文曲化忌 should point to 交友
	if len(result.HuaJi) != 1 {
		t.Fatalf("expected 1 HuaJi item for 文曲化忌, got %d: %+v", len(result.HuaJi), result.HuaJi)
	}
	if result.HuaJi[0].TransformedStar != "文曲" {
		t.Errorf("HuaJi[0].TransformedStar = %q, want %q", result.HuaJi[0].TransformedStar, "文曲")
	}
	if result.HuaJi[0].TargetPalace != "交友" {
		t.Errorf("HuaJi[0].TargetPalace = %q, want %q", result.HuaJi[0].TargetPalace, "交友")
	}
}

// ════════════════════════════════════════════════════════════════
// Tests for analyzeFourHua in ziwei.go (the flying stars analysis)
// ════════════════════════════════════════════════════════════════

// TestAnalyzeFlyingStars_YearHua tests the flying stars analysis for
// the annual (流年) four transformations.
func TestAnalyzeFlyingStars_YearHua(t *testing.T) {
	service := NewZiWeiService()
	chart, err := service.CalculateChart(2000, 6, 15, 12, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart: %v", err)
	}
	flyingResult := service.AnalyzeFlyingStars(chart)
	if flyingResult == nil {
		t.Fatal("AnalyzeFlyingStars returned nil")
	}
	_ = flyingResult
}
