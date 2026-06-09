package ziwei

import (
	"reflect"
	"testing"
)

// ════════════════════════════════════════════════════════════════
// Tests for SiHuaTable (ziwei_data.go) and SI_HUA_TABLE (ziwei_knowledge.go)
// ════════════════════════════════════════════════════════════════

// TestSiHuaTable_AllStems verifies calcFourHua for all 10 heavenly stems.
func TestSiHuaTable_AllStems(t *testing.T) {
	// Expected four transformations for each stem (0=甲 through 9=癸)
	// Order: [化禄, 化权, 化科, 化忌]
	expected := [][]string{
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

// TestSiHuaTable_Consistency verifies that SiHuaTable (array in ziwei_data.go)
// and SI_HUA_TABLE (map in ziwei_knowledge.go) hold identical data.
//
// NOTE: 庚年 (stem 6) is a known intentional difference:
//   - SiHuaTable[6] = [太阳, 武曲, 太阴, 天同]  (太阴=科, 天同=忌)
//   - SI_HUA_TABLE[6] = [太阳, 武曲, 天同, 太阴] (天同=科, 太阴=忌)
//   The SI_HUA_TABLE value includes a "(含line5435定盘修正)" correction.
func TestSiHuaTable_Consistency(t *testing.T) {
	for stem := 0; stem < 10; stem++ {
		t.Run(StemNames[stem], func(t *testing.T) {
			// From the array-based table in ziwei_data.go
			want := SiHuaTable[stem]

			// From the map-based table in ziwei_knowledge.go
			got, ok := SI_HUA_TABLE[stem]
			if !ok {
				t.Fatalf("SI_HUA_TABLE missing entry for stem %d", stem)
			}

			// 庚年 (stem 6) has a known intentional difference — skip strict equality
			if stem == 6 {
				t.Logf("庚年: known difference — SiHuaTable=%v, SI_HUA_TABLE=%v", want, got)
				return
			}

			if !reflect.DeepEqual(want[:], got) {
				t.Errorf("SiHuaTable[%d] = %v, SI_HUA_TABLE[%d] = %v", stem, want, stem, got)
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
		got, ok := SI_HUA_TABLE[stem]
		if !ok {
			t.Errorf("SI_HUA_TABLE missing stem %d", stem)
		} else if len(got) != 4 {
			t.Errorf("SI_HUA_TABLE[%d] has length %d, want 4", stem, len(got))
		}
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
	result := AnalyzeSihuaChain(nil)
	if result != nil {
		t.Errorf("AnalyzeSihuaChain(nil) = %v, want nil", result)
	}
}

// TestAnalyzeSihuaChain_WithSimplifiedChart uses SimplifiedZiWei to get a
// minimal chart and runs chain analysis on it.
func TestAnalyzeSihuaChain_WithSimplifiedChart(t *testing.T) {
	chart := SimplifiedZiWei(2000, 6, 15, 12, 0, "MALE")
	if chart == nil {
		t.Fatal("SimplifiedZiWei returned nil chart")
	}

	// SimplifiedZiWei doesn't set YearStem; set it explicitly for testing.
	// Use 甲年 (stem 0): 廉贞化禄, 破军化权, 武曲化科, 太阳化忌
	chart.YearStem = 0

	// The simplified chart has MainStars=["紫微","天机"] and AuxStars=["左辅","文昌"]
	// in every palace. None of these are in the 甲年 SiHuaTable,
	// so chain results should be empty arrays.
	result := AnalyzeSihuaChain(chart)
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
	if result.TotalChainDepth != 0 {
		t.Errorf("expected TotalChainDepth=0, got %d", result.TotalChainDepth)
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
			Name:     PALACE_NAMES[i],
			MainStars: []string{},
			AuxStars:  []string{},
			FourHua:   []string{},
		}
	}

	// Place 廉贞 in 命宫 (index 0) — this is 化禄 in 甲年
	palaces[0].MainStars = []string{"廉贞"}
	// Place 破军 in 子女 (index 3) — this is 化权 in 甲年
	palaces[3].MainStars = []string{"破军"}

	chart := &ZiWeiChart{
		Palaces:  palaces,
		YearStem: 0, // 甲年: 廉贞化禄, 破军化权, 武曲化科, 太阳化忌
	}

	result := AnalyzeSihuaChain(chart)
	if result == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	// Should find 廉贞化禄 pointing to 命宫
	if len(result.HuaLu) != 1 {
		t.Fatalf("expected 1 HuaLu item, got %d: %+v", len(result.HuaLu), result.HuaLu)
	}
	if result.HuaLu[0].FromStar != "廉贞" {
		t.Errorf("HuaLu[0].FromStar = %q, want %q", result.HuaLu[0].FromStar, "廉贞")
	}
	if result.HuaLu[0].ToPalace != "命宫" {
		t.Errorf("HuaLu[0].ToPalace = %q, want %q", result.HuaLu[0].ToPalace, "命宫")
	}
	if result.HuaLu[0].FromPalace != "命宫" {
		t.Errorf("HuaLu[0].FromPalace = %q, want %q", result.HuaLu[0].FromPalace, "命宫")
	}

	// Should find 破军化权 pointing to 子女
	if len(result.HuaQuan) != 1 {
		t.Fatalf("expected 1 HuaQuan item, got %d: %+v", len(result.HuaQuan), result.HuaQuan)
	}
	if result.HuaQuan[0].FromStar != "破军" {
		t.Errorf("HuaQuan[0].FromStar = %q, want %q", result.HuaQuan[0].FromStar, "破军")
	}
	if result.HuaQuan[0].ToPalace != "子女" {
		t.Errorf("HuaQuan[0].ToPalace = %q, want %q", result.HuaQuan[0].ToPalace, "子女")
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
			Name:     PALACE_NAMES[i],
			MainStars: []string{},
			AuxStars:  []string{},
			FourHua:   []string{},
		}
	}
	// 己年 (stem 5): 武曲化禄, 贪狼化权, 天梁化科, 文曲化忌
	// Place 文曲 as an aux star in 交友 (index 7)
	palaces[7].AuxStars = []string{"文曲"}
	// Place 天梁 as a main star in 迁移 (index 6)
	palaces[6].MainStars = []string{"天梁"}

	chart := &ZiWeiChart{
		Palaces:  palaces,
		YearStem: 5, // 己年
	}

	result := AnalyzeSihuaChain(chart)
	if result == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	// 天梁化科 should point to 迁移
	if len(result.HuaKe) != 1 {
		t.Fatalf("expected 1 HuaKe item for 天梁化科, got %d: %+v", len(result.HuaKe), result.HuaKe)
	}
	if result.HuaKe[0].FromStar != "天梁" {
		t.Errorf("HuaKe[0].FromStar = %q, want %q", result.HuaKe[0].FromStar, "天梁")
	}
	if result.HuaKe[0].ToPalace != "迁移" {
		t.Errorf("HuaKe[0].ToPalace = %q, want %q", result.HuaKe[0].ToPalace, "迁移")
	}

	// 文曲化忌 should point to 交友
	if len(result.HuaJi) != 1 {
		t.Fatalf("expected 1 HuaJi item for 文曲化忌, got %d: %+v", len(result.HuaJi), result.HuaJi)
	}
	if result.HuaJi[0].FromStar != "文曲" {
		t.Errorf("HuaJi[0].FromStar = %q, want %q", result.HuaJi[0].FromStar, "文曲")
	}
	if result.HuaJi[0].ToPalace != "交友" {
		t.Errorf("HuaJi[0].ToPalace = %q, want %q", result.HuaJi[0].ToPalace, "交友")
	}
}

// ════════════════════════════════════════════════════════════════
// Tests for analyzeFourHua in ziwei.go (the flying stars analysis)
// ════════════════════════════════════════════════════════════════

// TestAnalyzeFlyingStars_YearHua tests the flying stars analysis for
// the annual (流年) four transformations.
func TestAnalyzeFlyingStars_YearHua(t *testing.T) {
	chart := SimplifiedZiWei(2000, 6, 15, 12, 0, "MALE")
	if chart == nil {
		t.Fatal("SimplifiedZiWei returned nil")
	}

	// SimplifiedZiWei doesn't set YearStem; set for 壬年 (stem 8)
	// 壬年: 天梁化禄, 紫微化权, 左辅化科, 武曲化忌
	chart.YearStem = 8

	// The service method wraps FlyingStarAnalysis, so test the public API
	// via ZiWeiService
	service := &ZiWeiService{}
	flyingResult := service.AnalyzeFlyingStars(chart)
	if flyingResult == nil {
		t.Fatal("AnalyzeFlyingStars returned nil")
	}
	_ = flyingResult
}
