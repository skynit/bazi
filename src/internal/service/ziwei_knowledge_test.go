package service

import (
	"strings"
	"testing"
)

func TestComputeSanfangSizheng(t *testing.T) {
	tests := []struct {
		palaceIdx int
		want      [3]int
	}{
		{0, [3]int{6, 4, 8}},   // 命宫 -> 对宫 遷移(6), 三合 僕役(4), 官祿(8)
		{1, [3]int{7, 5, 9}},   // 兄弟宮 -> 对宫 田宅(7), 財帛(5), 僕役(9)
		{6, [3]int{0, 10, 2}},  // 遷移宮 -> 对宫 命宮(0), 官祿(10), 兄弟(2)
		{11, [3]int{5, 3, 7}},  // 父母宮 -> 对宫 疾厄(5), 夫妻(3), 田宅(7)
	}

	for _, tt := range tests {
		got := ComputeSanfangSizheng(tt.palaceIdx)
		if got != tt.want {
			t.Errorf("ComputeSanfangSizheng(%d) = %v, want %v", tt.palaceIdx, got, tt.want)
		}
	}
}

func TestGetPalaceSanfang(t *testing.T) {
	sf := GetPalaceSanfang(0)
	if sf.Opposite != "遷移宮" {
		t.Errorf("Opposite = %v, want 遷移宮", sf.Opposite)
	}
	if sf.Trine1 != "財帛宮" {
		t.Errorf("Trine1 = %v, want 財帛宮", sf.Trine1)
	}
	if sf.Trine2 != "官祿宮" {
		t.Errorf("Trine2 = %v, want 官祿宮", sf.Trine2)
	}
}

func TestGetStarBrightness(t *testing.T) {
	tests := []struct {
		star      string
		brightness string
		wantEmpty bool
	}{
		{"紫微", "庙", false},
		{"紫微", "陷", false},
		{"紫微", "unknown", false}, // falls back to 平
		{"贪狼", "庙", false},
		{"天梁", "旺", false},
	}

	for _, tt := range tests {
		got := GetStarBrightness(tt.star, tt.brightness)
		if tt.wantEmpty && got != "" {
			t.Errorf("GetStarBrightness(%s, %s) = %v, want empty", tt.star, tt.brightness, got)
		}
		if !tt.wantEmpty && got == "" {
			t.Errorf("GetStarBrightness(%s, %s) = empty, want non-empty", tt.star, tt.brightness)
		}
	}
}

func TestSihuaConstants(t *testing.T) {
	// Verify SI_HUA_TABLE has 10 entries (one per stem)
	if len(SI_HUA_TABLE) != 10 {
		t.Errorf("SI_HUA_TABLE has %d entries, want 10", len(SI_HUA_TABLE))
	}

	// Each stem should have exactly 4 transformations
	for stem, stars := range SI_HUA_TABLE {
		if len(stars) != 4 {
			t.Errorf("SI_HUA_TABLE[%d] has %d stars, want 4", stem, len(stars))
		}
	}
}

func TestPatternCheckersNotEmpty(t *testing.T) {
	if len(patternCheckers) < 35 {
		t.Errorf("patternCheckers has %d entries, want at least 35", len(patternCheckers))
	}
}

func TestDetectSelfMutagens(t *testing.T) {
	// nil chart returns nil
	result := DetectSelfMutagens(nil)
	if result != nil {
		t.Errorf("DetectSelfMutagens(nil) = %v, want nil", result)
	}
}

func TestSelfMutagenResultStruct(t *testing.T) {
	// Verify SelfMutagenResult struct fields
	r := SelfMutagenResult{
		Palace:  "命宮",
		Star:    "紫微",
		HuaType: "化禄",
		Effect:  "紫微化禄，此星在本宫自化禄，财运与事业有天生加持，但需防贪心",
		IsSelf:  true,
	}
	if r.Palace != "命宮" {
		t.Errorf("Palace = %v, want 命宮", r.Palace)
	}
	if r.Star != "紫微" {
		t.Errorf("Star = %v, want 紫微", r.Star)
	}
	if r.HuaType != "化禄" {
		t.Errorf("HuaType = %v, want 化禄", r.HuaType)
	}
	if r.Effect == "" {
		t.Error("Effect should not be empty")
	}
	if !r.IsSelf {
		t.Error("IsSelf should be true")
	}
}

func TestBuildSelfMutagenEffect(t *testing.T) {
	tests := []struct {
		star     string
		huaType  string
		palace   string
		wantCont string
	}{
		{"紫微", "化禄", "命宮", "紫微化禄"},
		{"天机", "化权", "兄弟宮", "天机化权"},
		{"太阳", "化科", "夫妻宮", "太阳化科"},
		{"武曲", "化忌", "財帛宮", "武曲化忌"},
	}
	for _, tt := range tests {
		got := buildSelfMutagenEffect(tt.star, tt.huaType, tt.palace)
		if !strings.Contains(got, tt.wantCont) {
			t.Errorf("buildSelfMutagenEffect(%s,%s,%s) = %v, want containing %s", tt.star, tt.huaType, tt.palace, got, tt.wantCont)
		}
	}
}

func TestComputeAdjectiveStars(t *testing.T) {
	// nil chart returns nil
	result := ComputeAdjectiveStars(nil)
	if result != nil {
		t.Errorf("ComputeAdjectiveStars(nil) = %v, want nil", result)
	}
}

func TestAdjectiveStarTablesExist(t *testing.T) {
	// Verify the lookup tables are populated
	if len(XIANCHI_TABLE) != 12 {
		t.Errorf("XIANCHI_TABLE has %d entries, want 12", len(XIANCHI_TABLE))
	}
	if len(HUAGAI_TABLE) != 12 {
		t.Errorf("HUAGAI_TABLE has %d entries, want 12", len(HUAGAI_TABLE))
	}
	if len(POUSUI_TABLE) != 12 {
		t.Errorf("POUSUI_TABLE has %d entries, want 12", len(POUSUI_TABLE))
	}
	if len(FEILIAN_TABLE) != 12 {
		t.Errorf("FEILIAN_TABLE has %d entries, want 12", len(FEILIAN_TABLE))
	}
	if len(YINSHIA_TABLE) != 12 {
		t.Errorf("YINSHIA_TABLE has %d entries, want 12", len(YINSHIA_TABLE))
	}
}

func TestComputeTwelveShen(t *testing.T) {
	result := ComputeTwelveShen(nil)
	if result != [12]struct{ Changsheng, Boshi, Jiangqian, Suiqian string }{} {
		t.Errorf("ComputeTwelveShen(nil) should return empty array")
	}
}

func TestTwelveShenConstants(t *testing.T) {
	if len(CHANGSHENG_12) != 12 {
		t.Errorf("CHANGSHENG_12 has %d entries, want 12", len(CHANGSHENG_12))
	}
	if len(BOSHI_12) != 12 {
		t.Errorf("BOSHI_12 has %d entries, want 12", len(BOSHI_12))
	}
	if len(JIANG_QIAN_12) != 12 {
		t.Errorf("JIANG_QIAN_12 has %d entries, want 12", len(JIANG_QIAN_12))
	}
	if len(SUI_QIAN_12) != 12 {
		t.Errorf("SUI_QIAN_12 has %d entries, want 12", len(SUI_QIAN_12))
	}
}

func TestSihuaChainItemExtendedFields(t *testing.T) {
	item := SihuaChainItem{
		FromStar:     "紫微",
		ToPalace:     "命宮",
		FromPalace:   "命宮",
		Effect:       "紫微化禄飞入命宮",
		ChainDepth:   1,
		StarAffinity: 2,
		MutagenType:  "self",
		FlyDirection: "same_palace",
		IsSelfMutagen: true,
	}
	if item.MutagenType != "self" {
		t.Errorf("MutagenType = %v, want self", item.MutagenType)
	}
	if item.FlyDirection != "same_palace" {
		t.Errorf("FlyDirection = %v, want same_palace", item.FlyDirection)
	}
	if !item.IsSelfMutagen {
		t.Error("IsSelfMutagen should be true")
	}
}

func TestSihuaChainResultExtendedFields(t *testing.T) {
	result := SihuaChainResult{
		HuaLu:   []SihuaChainItem{},
		HuaQuan: []SihuaChainItem{},
		HuaKe:   []SihuaChainItem{},
		HuaJi:   []SihuaChainItem{},
		TotalChainDepth: 5,
		KeyMutagens: []string{"紫微", "天机"},
	}
	if result.TotalChainDepth != 5 {
		t.Errorf("TotalChainDepth = %v, want 5", result.TotalChainDepth)
	}
	if len(result.KeyMutagens) != 2 {
		t.Errorf("KeyMutagens has %d entries, want 2", len(result.KeyMutagens))
	}
}

func TestLiuYaoStars(t *testing.T) {
	// Test that LIU_YAO_STARS has 10 entries
	if len(LIU_YAO_STARS) != 10 {
		t.Errorf("LIU_YAO_STARS has %d entries, want 10", len(LIU_YAO_STARS))
	}
	// Verify expected stars
	expected := []string{"天魁", "天钺", "文昌", "文曲", "禄存", "擎羊", "陀罗", "天马", "红鸾", "天喜"}
	for i, star := range LIU_YAO_STARS {
		if star != expected[i] {
			t.Errorf("LIU_YAO_STARS[%d] = %v, want %v", i, star, expected[i])
		}
	}
}

func TestComputeLiuNianStars(t *testing.T) {
	// nil chart returns empty array
	result := computeLiuNianStars(nil, 2025)
	for i := 0; i < 12; i++ {
		if len(result[i]) != 0 {
			t.Errorf("computeLiuNianStars(nil)[%d] = %v, want empty", i, result[i])
		}
	}
}

func TestComputeLiuYueStars(t *testing.T) {
	// nil chart returns empty array
	result := computeLiuYueStars(nil, 5)
	for i := 0; i < 12; i++ {
		if len(result[i]) != 0 {
			t.Errorf("computeLiuYueStars(nil)[%d] = %v, want empty", i, result[i])
		}
	}
}

func TestComputeLiuRiStars(t *testing.T) {
	// nil chart returns empty array
	result := computeLiuRiStars(nil, 15)
	for i := 0; i < 12; i++ {
		if len(result[i]) != 0 {
			t.Errorf("computeLiuRiStars(nil)[%d] = %v, want empty", i, result[i])
		}
	}
}

func TestStarPluginInterface(t *testing.T) {
	// Test that StarPlugin interface is defined correctly
	var plugin StarPlugin = nil // this won't compile if interface is wrong
	_ = plugin
}

func TestAlgorithmType(t *testing.T) {
	if AlgorithmFullBook != 0 {
		t.Errorf("AlgorithmFullBook = %v, want 0", AlgorithmFullBook)
	}
	if AlgorithmZhongZhou != 1 {
		t.Errorf("AlgorithmZhongZhou = %v, want 1", AlgorithmZhongZhou)
	}
}

func TestZhongZhouMingGong(t *testing.T) {
	// Test 中州派 formula: (yearBranch + 1) % 12
	// branchOrder: ["寅","卯","辰","巳","午","未","申","酉","戌","亥","子","丑"]
	// branch indices: 0=寅, 1=卯, 2=辰, 3=巳, 4=午, 5=未, 6=申, 7=酉, 8=戌, 9=亥, 10=子, 11=丑
	tests := []struct {
		yearBranch int
		want       string
	}{
		{0, "卯"},   // 寅(0) + 1 = 卯(1)
		{2, "巳"},   // 辰(2) + 1 = 巳(3)
		{9, "子"},   // 亥(9) + 1 = 子(10)
		{10, "丑"},  // 子(10) + 1 = 丑(11)
	}
	for _, tt := range tests {
		got := computeMingGongBranchZhongZhou(tt.yearBranch)
		if got != tt.want {
			t.Errorf("computeMingGongBranchZhongZhou(%d) = %v, want %v", tt.yearBranch, got, tt.want)
		}
	}
}