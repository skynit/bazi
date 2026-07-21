package ziwei

import "testing"

func TestUnsupportedFourHuaCombinationPatternsAreNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{{Name: "武曲", Type: "major"}, {Name: "天同", Type: "major"}}
	chart.Palaces[0].FourHua = []string{"武曲化科", "天同化忌"}
	chart.Palaces[6].Stars = []StarOutput{{Name: "天机", Type: "major"}, {Name: "廉贞", Type: "major"}}
	chart.Palaces[6].FourHua = []string{"天机化禄", "廉贞化忌"}

	for _, pattern := range DetectLocalPatterns(chart) {
		switch pattern {
		case "半空折翅格", "科忌同宫", "禄忌同宫":
			t.Fatalf("未经固定来源支持的四化组合仍被发布为格局: %s", pattern)
		}
	}
	if !palaceHasHua(chart, 0, "化科") || !palaceHasHua(chart, 0, "化忌") ||
		!palaceHasHua(chart, 6, "化禄") || !palaceHasHua(chart, 6, "化忌") {
		t.Fatal("撤下格局标签时丢失了原始四化落宫结构")
	}
}

func TestSingleToughStarInLifePalaceIsNotPublishedAsPattern(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{
		{Name: "擎羊", Type: "tough"},
		{Name: "陀罗", Type: "tough"},
	}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "擎羊入命" || pattern == "陀罗入命" {
			t.Fatalf("单颗煞星坐命仍被升级为整盘格局: %s", pattern)
		}
	}
	aux := palaceAuxStars(chart.Palaces[0])
	hasQingYang, hasTuoLuo := false, false
	for _, star := range aux {
		hasQingYang = hasQingYang || star == "擎羊"
		hasTuoLuo = hasTuoLuo || star == "陀罗"
	}
	if !hasQingYang || !hasTuoLuo {
		t.Fatalf("撤下格局标签时丢失了命宫原始煞星: %v", aux)
	}
	if _, exists := buildPatternTemplates()["擎羊入命"]; exists {
		t.Fatal("已撤下的擎羊入命仍保留解释模板")
	}
	if _, exists := buildPatternTemplates()["陀羅入命"]; exists {
		t.Fatal("已撤下的陀罗入命仍保留解释模板")
	}
}

func TestShaPoLangDoesNotPromoteBodyOnlyFormation(t *testing.T) {
	bodyChart := chartWithLifePalaceAtZi()
	bodyChart.Palaces[1].IsBodyPalace = true
	bodyChart.Palaces[1].Stars = []StarOutput{{Name: "七杀", Type: "major"}}
	bodyChart.Palaces[5].Stars = []StarOutput{{Name: "破军", Type: "major"}}
	bodyChart.Palaces[9].Stars = []StarOutput{{Name: "贪狼", Type: "major"}}

	if matched, _ := checkShaPoLang(bodyChart); matched {
		t.Fatal("仅身宫三方齐会杀破狼不得自动发布整盘杀破狼格")
	}
}

func TestUnsupportedMainStarPairAliasesAreNotPublished(t *testing.T) {
	tests := []struct {
		name       string
		stars      []StarOutput
		oldPattern string
	}{
		{name: "紫微破军", stars: []StarOutput{{Name: "紫微", Type: "major"}, {Name: "破军", Type: "major"}}, oldPattern: "紫破同宫"},
		{name: "廉贞天府", stars: []StarOutput{{Name: "廉贞", Type: "major"}, {Name: "天府", Type: "major"}}, oldPattern: "廉府双星"},
		{name: "廉贞七杀庙旺", stars: []StarOutput{{Name: "廉贞", Type: "major"}, {Name: "七杀", Type: "major", Brightness: "庙"}}, oldPattern: "廉贞七杀同宫庙旺"},
		{name: "廉贞七杀落陷", stars: []StarOutput{{Name: "廉贞", Type: "major"}, {Name: "七杀", Type: "major", Brightness: "陷"}}, oldPattern: "廉贞七杀同宫落陷"},
		{name: "廉贞破军", stars: []StarOutput{{Name: "廉贞", Type: "major"}, {Name: "破军", Type: "major"}}, oldPattern: "廉贞破军同宫"},
		{name: "廉贞贪狼落陷", stars: []StarOutput{{Name: "廉贞", Type: "major"}, {Name: "贪狼", Type: "major", Brightness: "陷"}}, oldPattern: "廉贞贪狼同宫落陷"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			chart.Palaces[1].Stars = tc.stars
			for _, pattern := range DetectLocalPatterns(chart) {
				if pattern == tc.oldPattern {
					t.Fatalf("无固定来源的主星对仍被发布成整盘格局: %s", pattern)
				}
			}
			if len(chart.Palaces[1].Stars) != len(tc.stars) {
				t.Fatal("撤下格局别名时丢失了原始主星落宫")
			}
		})
	}
}

func TestTianTongTianLiangUsesSourcedCanonicalName(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Branch = "寅"
	chart.Palaces[0].Stars = []StarOutput{{Name: "天同", Type: "major"}, {Name: "天梁", Type: "major"}}
	hasCanonical := false
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "天同天梁格" {
			hasCanonical = true
		}
		if pattern == "同梁双星" {
			t.Fatal("仍发布非固定来源名称同梁双星")
		}
	}
	if !hasCanonical {
		t.Fatal("天同、天梁同坐寅宫命宫未发布天同天梁格")
	}

	nonLife := chartWithLifePalaceAtZi()
	nonLife.Palaces[2].Stars = []StarOutput{{Name: "天同", Type: "major"}, {Name: "天梁", Type: "major"}}
	if matched, _ := checkTianTongTianLiang(nonLife); matched {
		t.Fatal("天同、天梁只在非命宫同宫，不得发布整盘天同天梁格")
	}
}

func TestJuRiTongGongRequiresYinOrShenLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Branch = "申"
	positive.Palaces[0].Stars = []StarOutput{{Name: "巨门", Type: "major"}, {Name: "太阳", Type: "major"}}
	if matched, name := checkJuRiTongGong(positive); !matched || name != "巨日同宫" {
		t.Fatalf("巨门太阳同坐申宫命宫未命中: matched=%v name=%q", matched, name)
	}

	nonLife := chartWithLifePalaceAtZi()
	nonLife.Palaces[2].Stars = []StarOutput{{Name: "巨门", Type: "major"}, {Name: "太阳", Type: "major"}}
	if matched, _ := checkJuRiTongGong(nonLife); matched {
		t.Fatal("巨门太阳只在非命宫同宫，不得发布整盘巨日同宫")
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = []StarOutput{{Name: "巨门", Type: "major"}, {Name: "太阳", Type: "major"}}
	if matched, _ := checkJuRiTongGong(wrongBranch); matched {
		t.Fatal("巨门太阳同坐子宫命宫不符合寅申巨日结构")
	}
}

func TestFuXiangChaoYuanRequiresFuInCareerAndXiangInWealth(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[8].Name = "事业"
	positive.Palaces[8].Stars = []StarOutput{{Name: "天府", Type: "major"}}
	positive.Palaces[4].Name = "财帛"
	positive.Palaces[4].Stars = []StarOutput{{Name: "天相", Type: "major"}}
	if matched, name := checkFuXiangChaoYuan(positive); !matched || name != "府相朝垣" {
		t.Fatalf("天府守事业、天相守财帛未命中府相朝垣: matched=%v name=%q", matched, name)
	}

	reversed := chartWithLifePalaceAtZi()
	reversed.Palaces[8].Name = "事业"
	reversed.Palaces[8].Stars = []StarOutput{{Name: "天相", Type: "major"}}
	reversed.Palaces[4].Name = "财帛"
	reversed.Palaces[4].Stars = []StarOutput{{Name: "天府", Type: "major"}}
	if matched, _ := checkFuXiangChaoYuan(reversed); matched {
		t.Fatal("天府、天相宫位对调不得误报府相朝垣")
	}

	genericSanfang := chartWithLifePalaceAtZi()
	genericSanfang.Palaces[8].Name = "事业"
	genericSanfang.Palaces[4].Name = "财帛"
	genericSanfang.Palaces[0].Stars = []StarOutput{{Name: "天府", Type: "major"}}
	genericSanfang.Palaces[6].Stars = []StarOutput{{Name: "天相", Type: "major"}}
	if matched, _ := checkFuXiangChaoYuan(genericSanfang); matched {
		t.Fatal("天府、天相只在任意命宫三方位置不得误报府相朝垣")
	}
}

func TestZiFuTongGongRequiresYinOrShenLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Branch = "寅"
	positive.Palaces[0].Stars = []StarOutput{{Name: "紫微", Type: "major"}, {Name: "天府", Type: "major"}}
	if matched, name := checkZiFuTongGong(positive); !matched || name != "紫府同宫" {
		t.Fatalf("紫微天府同坐寅宫命宫未命中: matched=%v name=%q", matched, name)
	}

	nonLife := chartWithLifePalaceAtZi()
	nonLife.Palaces[2].Stars = []StarOutput{{Name: "紫微", Type: "major"}, {Name: "天府", Type: "major"}}
	if matched, _ := checkZiFuTongGong(nonLife); matched {
		t.Fatal("紫微天府只在非命宫同宫，不得发布整盘紫府同宫格")
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = []StarOutput{{Name: "紫微", Type: "major"}, {Name: "天府", Type: "major"}}
	if matched, _ := checkZiFuTongGong(wrongBranch); matched {
		t.Fatal("紫微天府同坐子宫命宫不符合寅申紫府同宫结构")
	}
}

func TestShiZhongYinYuRequiresJuMenInZiOrWuLifePalace(t *testing.T) {
	positive := &ZiWeiChart{}
	positive.Palaces[0] = PalaceInfo{
		Name: "命宫", Branch: "子",
		Stars: []StarOutput{{Name: "巨门", Type: "major", Brightness: "旺"}},
	}
	if matched, name := checkShiZhongYinYu(positive); !matched || name != "石中隐玉" {
		t.Fatalf("巨门子宫坐命未命中石中隐玉: matched=%v name=%q", matched, name)
	}

	wrongPalace := &ZiWeiChart{}
	wrongPalace.Palaces[0] = PalaceInfo{Name: "命宫", Branch: "寅"}
	wrongPalace.Palaces[1] = PalaceInfo{
		Name: "事业", Branch: "子",
		Stars:   []StarOutput{{Name: "巨门", Type: "major", Brightness: "旺"}},
		FourHua: []string{"其他星化禄"},
	}
	if matched, _ := checkShiZhongYinYu(wrongPalace); matched {
		t.Fatal("巨门不在命宫却误报石中隐玉")
	}

	wrongBranch := &ZiWeiChart{}
	wrongBranch.Palaces[0] = PalaceInfo{
		Name: "命宫", Branch: "寅",
		Stars:   []StarOutput{{Name: "巨门", Type: "major", Brightness: "旺"}},
		FourHua: []string{"巨门化权"},
	}
	if matched, _ := checkShiZhongYinYu(wrongBranch); matched {
		t.Fatal("巨门坐命但不在子午却误报石中隐玉")
	}
}

func TestKeQuanShuangHuiRequiresMajorTransformationsInLifeSanfang(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[4].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	positive.Palaces[4].FourHua = []string{"武曲化科"}
	positive.Palaces[8].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	positive.Palaces[8].FourHua = []string{"天梁化权"}
	if matched, name := checkKeQuanShuangHui(positive); !matched || name != "科权双会" {
		t.Fatalf("命宫三方四正见主星科权未命中: matched=%v name=%q", matched, name)
	}

	outsideLifeSanfang := chartWithLifePalaceAtZi()
	outsideLifeSanfang.Palaces[1].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	outsideLifeSanfang.Palaces[1].FourHua = []string{"武曲化科"}
	outsideLifeSanfang.Palaces[5].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	outsideLifeSanfang.Palaces[5].FourHua = []string{"天梁化权"}
	if matched, _ := checkKeQuanShuangHui(outsideLifeSanfang); matched {
		t.Fatal("科权只在命宫三方四正之外却误报科权双会")
	}

	auxiliaryTransformation := chartWithLifePalaceAtZi()
	auxiliaryTransformation.Palaces[4].Stars = []StarOutput{{Name: "右弼", Type: "soft"}}
	auxiliaryTransformation.Palaces[4].FourHua = []string{"右弼化科"}
	auxiliaryTransformation.Palaces[8].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	auxiliaryTransformation.Palaces[8].FourHua = []string{"天梁化权"}
	if matched, _ := checkKeQuanShuangHui(auxiliaryTransformation); matched {
		t.Fatal("辅星化科被误当作主星科权双会")
	}
}

func TestSanQiJiaHuiUsesCanonicalNameWithoutDuplicateAlias(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	chart.Palaces[0].FourHua = []string{"武曲化禄"}
	chart.Palaces[4].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	chart.Palaces[4].FourHua = []string{"天梁化权"}
	chart.Palaces[8].Stars = []StarOutput{{Name: "天同", Type: "major"}}
	chart.Palaces[8].FourHua = []string{"天同化科"}

	chart.Patterns = DetectLocalPatterns(chart)
	counts := map[string]int{}
	for _, pattern := range chart.Patterns {
		counts[pattern]++
	}
	if counts["三奇加会"] != 1 || counts["科权双会"] != 1 {
		t.Fatalf("规范四化格局名称未各发布一次: %v", chart.Patterns)
	}
	for _, retired := range []string{"三奇嘉会", "科权相逢", "科权禄三会"} {
		if counts[retired] != 0 {
			t.Fatalf("旧名称或重复格局仍在发布: %s patterns=%v", retired, chart.Patterns)
		}
	}
	details := buildPatternDetailsForPalace(chart, 0)
	if !patternDetailsContain(details, "三奇加会") || !patternDetailsContain(details, "科权双会") {
		t.Fatalf("命宫格局依据未同步规范名称: %#v", details)
	}
}

func chartWithLifePalaceAtZi() *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i, branch := range BranchNames {
		chart.Palaces[i].Branch = branch
	}
	chart.Palaces[0].Name = "命宫"
	return chart
}

func TestJiYueTongLiangRequiresYinOrShenLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Name = ""
	positive.Palaces[2].Name = "命宫"
	positive.Palaces[2].Stars = []StarOutput{{Name: "天机", Type: "major"}}
	positive.Palaces[6].Stars = []StarOutput{{Name: "天同", Type: "major"}}
	positive.Palaces[8].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	positive.Palaces[10].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	if matched, name := checkJiYueTongLiang(positive); !matched || name != "机月同梁格" {
		t.Fatalf("寅宫命宫四星齐会应成机月同梁格: matched=%v name=%q", matched, name)
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = []StarOutput{{Name: "天机", Type: "major"}}
	wrongBranch.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	wrongBranch.Palaces[6].Stars = []StarOutput{{Name: "天同", Type: "major"}}
	wrongBranch.Palaces[8].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	if matched, _ := checkJiYueTongLiang(wrongBranch); matched {
		t.Fatal("子宫命宫即使四星齐会也不得误报只在寅申成立的机月同梁格")
	}
}

func TestUnsupportedTianMaGongMingIsNotPublished(t *testing.T) {
	tianMaTrinesLife := chartWithLifePalaceAtZi()
	tianMaTrinesLife.Palaces[4].Stars = []StarOutput{{Name: "天马", Type: "tianma"}}
	for _, pattern := range DetectLocalPatterns(tianMaTrinesLife) {
		if pattern == "天马拱命" {
			t.Fatal("仅有天马位于命宫三方，不得发布无固定来源的天马拱命")
		}
	}
	if _, exists := buildPatternTemplates()["天馬拱命"]; exists {
		t.Fatal("已撤下的天马拱命仍保留解释模板")
	}
}

func TestLuMaPatternsReadAuxiliaryStarsAcrossIntendedPalaces(t *testing.T) {
	luMaSamePalace := chartWithLifePalaceAtZi()
	luMaSamePalace.Palaces[3].Stars = []StarOutput{
		{Name: "禄存", Type: "lucun"},
		{Name: "天马", Type: "tianma"},
	}
	if matched, name := checkLuMaJiaoChi(luMaSamePalace); !matched || name != "禄马交驰" {
		t.Fatalf("辅星禄存与天马同宫未命中禄马交驰: matched=%v name=%q", matched, name)
	}

	luMaOpposition := chartWithLifePalaceAtZi()
	luMaOpposition.Palaces[0].Stars = []StarOutput{{Name: "禄存", Type: "lucun"}}
	luMaOpposition.Palaces[6].Stars = []StarOutput{{Name: "天马", Type: "tianma"}}
	if matched, _ := checkLuMaJiaoChi(luMaOpposition); matched {
		t.Fatal("禄存与天马仅在对宫不得误报同宫口径的禄马交驰")
	}
}

func TestMaTouDaiJianRequiresQingYangInWuLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Branch = "午"
	positive.Palaces[0].Stars = []StarOutput{{Name: "擎羊", Type: "tough"}}
	if matched, name := checkMaTouDaiJian(positive); !matched || name != "马头带箭" {
		t.Fatalf("擎羊在午宫坐命未命中马头带箭: matched=%v name=%q", matched, name)
	}

	oldWrongRule := chartWithLifePalaceAtZi()
	oldWrongRule.Palaces[1].Stars = []StarOutput{
		{Name: "擎羊", Type: "tough"},
		{Name: "天马", Type: "tianma"},
	}
	if matched, _ := checkMaTouDaiJian(oldWrongRule); matched {
		t.Fatal("天马、擎羊在非命宫同宫不得冒充马头带箭")
	}

	for _, pattern := range DetectLocalPatterns(oldWrongRule) {
		if pattern == "马头带剑" {
			t.Fatal("仍发布错误名称马头带剑")
		}
	}
}

func TestQiShaChaoDouUsesBranchSpecificOppositeStars(t *testing.T) {
	tests := []struct {
		name          string
		lifeBranch    string
		oppositeStars []StarOutput
	}{
		{
			name:       "七杀在寅对宫紫府",
			lifeBranch: "寅",
			oppositeStars: []StarOutput{
				{Name: "紫微", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
		{
			name:       "七杀在申对宫紫府",
			lifeBranch: "申",
			oppositeStars: []StarOutput{
				{Name: "紫微", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
		{
			name:       "七杀在子对宫武府",
			lifeBranch: "子",
			oppositeStars: []StarOutput{
				{Name: "武曲", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
		{
			name:       "七杀在午对宫武府",
			lifeBranch: "午",
			oppositeStars: []StarOutput{
				{Name: "武曲", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			lifeIdx := BranchIndex[tc.lifeBranch]
			chart.Palaces[0].Name = ""
			chart.Palaces[lifeIdx].Name = "命宫"
			chart.Palaces[lifeIdx].Stars = []StarOutput{{Name: "七杀", Type: "major"}}
			oppositeIdx := fixIndex(lifeIdx + 6)
			chart.Palaces[oppositeIdx].Stars = tc.oppositeStars

			if matched, name := checkQiShaChaoDou(chart); !matched || name != "七杀朝斗" {
				t.Fatalf("合法七杀朝斗未命中: matched=%v name=%q", matched, name)
			}
		})
	}

	for _, tc := range []struct {
		name          string
		lifeBranch    string
		oppositeStars []StarOutput
	}{
		{
			name:       "子命误用紫府",
			lifeBranch: "子",
			oppositeStars: []StarOutput{
				{Name: "紫微", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
		{
			name:       "寅命误用武府",
			lifeBranch: "寅",
			oppositeStars: []StarOutput{
				{Name: "武曲", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			lifeIdx := BranchIndex[tc.lifeBranch]
			chart.Palaces[0].Name = ""
			chart.Palaces[lifeIdx].Name = "命宫"
			chart.Palaces[lifeIdx].Stars = []StarOutput{{Name: "七杀", Type: "major"}}
			oppositeIdx := fixIndex(lifeIdx + 6)
			chart.Palaces[oppositeIdx].Stars = tc.oppositeStars

			if matched, _ := checkQiShaChaoDou(chart); matched {
				t.Fatal("错误对宫星组不得误报七杀朝斗")
			}
		})
	}
}

func TestYangTuoJiaJiRequiresHuaJiInLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	positive.Palaces[0].FourHua = []string{"武曲化忌"}
	positive.Palaces[11].Stars = []StarOutput{{Name: "擎羊", Type: "tough"}}
	positive.Palaces[1].Stars = []StarOutput{{Name: "陀罗", Type: "tough"}}
	if matched, name := checkYangTuoJiaJi(positive); !matched || name != "羊陀夹忌" {
		t.Fatalf("化忌坐命且羊陀分居两侧未命中羊陀夹忌: matched=%v name=%q", matched, name)
	}

	lucunOnly := chartWithLifePalaceAtZi()
	lucunOnly.Palaces[0].Stars = []StarOutput{{Name: "禄存", Type: "lucun"}}
	lucunOnly.Palaces[11].Stars = []StarOutput{{Name: "擎羊", Type: "tough"}}
	lucunOnly.Palaces[1].Stars = []StarOutput{{Name: "陀罗", Type: "tough"}}
	if matched, _ := checkYangTuoJiaJi(lucunOnly); matched {
		t.Fatal("禄存坐命天然被羊陀夹但没有化忌，不得误报羊陀夹忌")
	}

	for _, pattern := range DetectLocalPatterns(lucunOnly) {
		if pattern == "羊陀夹杀" {
			t.Fatal("仍发布缺少七杀条件的旧名称羊陀夹杀")
		}
	}
}

func TestUnsupportedKeMingHuiLuIsNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{{Name: "天梁", Type: "major"}}
	chart.Palaces[0].FourHua = []string{"天梁化科"}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "科名会禄" {
			t.Fatal("无来源的科名会禄检测器仍在发布")
		}
	}
}

func TestUnsupportedTaoHuaFanZhuIsNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{
		{Name: "紫微", Type: "major"},
		{Name: "左辅", Type: "soft"},
		{Name: "右弼", Type: "soft"},
	}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "桃花犯主" {
			t.Fatal("无可靠可执行条件的桃花犯主检测器仍在发布")
		}
	}
	if _, exists := buildPatternTemplates()["桃花犯主"]; exists {
		t.Fatal("已撤下的桃花犯主仍保留解释模板")
	}
}

func TestRiYueGongZhaoUsesSourcedBranchStructures(t *testing.T) {
	tests := []struct {
		name       string
		lifeBranch string
		sunBranch  string
		moonBranch string
	}{
		{name: "日巳月酉命丑", lifeBranch: "丑", sunBranch: "巳", moonBranch: "酉"},
		{name: "日卯月亥命未", lifeBranch: "未", sunBranch: "卯", moonBranch: "亥"},
		{name: "日月同未命丑", lifeBranch: "丑", sunBranch: "未", moonBranch: "未"},
		{name: "日辰月戌命辰", lifeBranch: "辰", sunBranch: "辰", moonBranch: "戌"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			chart.Palaces[0].Name = ""
			chart.Palaces[BranchIndex[tc.lifeBranch]].Name = "命宫"
			chart.Palaces[BranchIndex[tc.sunBranch]].Stars = append(
				chart.Palaces[BranchIndex[tc.sunBranch]].Stars,
				StarOutput{Name: "太阳", Type: "major"},
			)
			chart.Palaces[BranchIndex[tc.moonBranch]].Stars = append(
				chart.Palaces[BranchIndex[tc.moonBranch]].Stars,
				StarOutput{Name: "太阴", Type: "major"},
			)
			if matched, name := checkRiYueGongZhao(chart); !matched || name != "日月拱照" {
				t.Fatalf("固定日月拱照结构未命中: matched=%v name=%q", matched, name)
			}
		})
	}

	genericTrine := chartWithLifePalaceAtZi()
	genericTrine.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major"}}
	genericTrine.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	if matched, _ := checkRiYueGongZhao(genericTrine); matched {
		t.Fatal("任意日月三合不得泛化为日月拱照")
	}
}

func TestUnsupportedZiXiangGongZhaoIsNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{{Name: "紫微", Type: "major"}}
	chart.Palaces[4].Stars = []StarOutput{{Name: "天相", Type: "major"}}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "紫相拱照" {
			t.Fatal("无固定来源独立规则的紫相拱照仍在发布")
		}
	}
	if _, exists := buildPatternTemplates()["紫相拱照"]; exists {
		t.Fatal("已撤下的紫相拱照仍保留解释模板")
	}
}

func TestUnsupportedElementalCombinationLabelsAreNotPublished(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		stars   []StarOutput
	}{
		{
			name:    "天机太阴不得改名为水木清华",
			pattern: "水木清华",
			stars: []StarOutput{
				{Name: "天机", Type: "major"},
				{Name: "太阴", Type: "major"},
			},
		},
		{
			name:    "武曲天府不得改名为土金相生",
			pattern: "土金相生",
			stars: []StarOutput{
				{Name: "武曲", Type: "major"},
				{Name: "天府", Type: "major"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			chart.Palaces[0].Stars = tc.stars
			for _, pattern := range DetectLocalPatterns(chart) {
				if pattern == tc.pattern {
					t.Fatalf("无固定来源的%s仍在发布", tc.pattern)
				}
			}
			if _, exists := buildPatternTemplates()[tc.pattern]; exists {
				t.Fatalf("已撤下的%s仍保留解释模板", tc.pattern)
			}
		})
	}
}

func TestRiYueBingMingRequiresBrightSunMoonInLifeSanfang(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "庙"}}
	positive.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "旺"}}
	if matched, name := checkRiYueBingMing(positive); !matched || name != "日月并明" {
		t.Fatalf("命宫三方见庙旺日月未命中: matched=%v name=%q", matched, name)
	}

	dim := chartWithLifePalaceAtZi()
	dim.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "陷"}}
	dim.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "旺"}}
	if matched, _ := checkRiYueBingMing(dim); matched {
		t.Fatal("太阳落陷不得误报日月并明")
	}

	outsideLifeSanfang := chartWithLifePalaceAtZi()
	outsideLifeSanfang.Palaces[1].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "庙"}}
	outsideLifeSanfang.Palaces[5].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "旺"}}
	if matched, _ := checkRiYueBingMing(outsideLifeSanfang); matched {
		t.Fatal("庙旺日月只在命宫三方之外不得误报日月并明")
	}
}

func TestRiYueFanBeiRequiresDimSunMoonInLifeSanfang(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "陷"}}
	positive.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "不"}}
	if matched, name := checkRiYueFanBei(positive); !matched || name != "日月反背" {
		t.Fatalf("命宫三方见落陷日月未命中: matched=%v name=%q", matched, name)
	}

	oneDim := chartWithLifePalaceAtZi()
	oneDim.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "陷"}}
	oneDim.Palaces[4].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "旺"}}
	if matched, _ := checkRiYueFanBei(oneDim); matched {
		t.Fatal("仅一颗日月落陷不得误报日月反背")
	}

	outsideLifeSanfang := chartWithLifePalaceAtZi()
	outsideLifeSanfang.Palaces[1].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "陷"}}
	outsideLifeSanfang.Palaces[5].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "不"}}
	if matched, _ := checkRiYueFanBei(outsideLifeSanfang); matched {
		t.Fatal("落陷日月只在命宫三方之外不得误报日月反背")
	}
}

func TestWuTanRequiresWuQuTanLangTogetherInChouOrWeiLifeBodyPalace(t *testing.T) {
	lifePalace := chartWithLifePalaceAtZi()
	lifePalace.Palaces[0].Branch = "丑"
	lifePalace.Palaces[0].Stars = []StarOutput{
		{Name: "武曲", Type: "major"},
		{Name: "贪狼", Type: "major"},
	}
	if matched, name := checkWuTanGe(lifePalace); !matched || name != "武贪格" {
		t.Fatalf("武贪同坐丑宫命宫未命中: matched=%v name=%q", matched, name)
	}

	bodyPalace := chartWithLifePalaceAtZi()
	bodyPalace.Palaces[2].Branch = "未"
	bodyPalace.Palaces[2].IsBodyPalace = true
	bodyPalace.Palaces[2].Stars = []StarOutput{
		{Name: "武曲", Type: "major"},
		{Name: "贪狼", Type: "major"},
	}
	if matched, name := checkWuTanGe(bodyPalace); !matched || name != "武贪格" {
		t.Fatalf("武贪同坐未宫身宫未命中: matched=%v name=%q", matched, name)
	}

	opposite := chartWithLifePalaceAtZi()
	opposite.Palaces[0].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	opposite.Palaces[6].Stars = []StarOutput{{Name: "贪狼", Type: "major"}}
	if matched, _ := checkWuTanGe(opposite); matched {
		t.Fatal("武贪对宫不得误报武贪格")
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = []StarOutput{
		{Name: "武曲", Type: "major"},
		{Name: "贪狼", Type: "major"},
	}
	if matched, _ := checkWuTanGe(wrongBranch); matched {
		t.Fatal("武贪同坐子宫命宫不符合丑未武贪结构")
	}

	outsideLifeBody := chartWithLifePalaceAtZi()
	outsideLifeBody.Palaces[1].Branch = "丑"
	outsideLifeBody.Palaces[1].Stars = []StarOutput{
		{Name: "武曲", Type: "major"},
		{Name: "贪狼", Type: "major"},
	}
	if matched, _ := checkWuTanGe(outsideLifeBody); matched {
		t.Fatal("武贪同坐丑宫但不在身命宫不得发布整盘武贪格")
	}
}

func TestHuoLingTanRequiresAuxStarRelationToLifeTanLang(t *testing.T) {
	samePalace := chartWithLifePalaceAtZi()
	samePalace.Palaces[0].Stars = []StarOutput{
		{Name: "贪狼", Type: "major"},
		{Name: "火星", Type: "tough"},
	}
	if matched, name := checkHuoTanGe(samePalace); !matched || name != "火贪格" {
		t.Fatalf("火贪同宫会命未命中: matched=%v name=%q", matched, name)
	}

	fourTombTrine := chartWithLifePalaceAtZi()
	fourTombTrine.Palaces[0].Name = ""
	fourTombTrine.Palaces[4].Name = "命宫"
	fourTombTrine.Palaces[4].Stars = []StarOutput{{Name: "贪狼", Type: "major"}}
	fourTombTrine.Palaces[8].Stars = []StarOutput{{Name: "铃星", Type: "tough"}}
	if matched, name := checkLingTanGe(fourTombTrine); !matched || name != "铃贪格" {
		t.Fatalf("贪狼在辰宫受铃星三合未命中: matched=%v name=%q", matched, name)
	}

	nonFourTombTrine := chartWithLifePalaceAtZi()
	nonFourTombTrine.Palaces[0].Stars = []StarOutput{{Name: "贪狼", Type: "major"}}
	nonFourTombTrine.Palaces[4].Stars = []StarOutput{{Name: "铃星", Type: "tough"}}
	if matched, _ := checkLingTanGe(nonFourTombTrine); matched {
		t.Fatal("贪狼不在四墓宫且铃星只三合会照，不得泛化为铃贪格")
	}

	fallenFireTan := chartWithLifePalaceAtZi()
	fallenFireTan.Palaces[0].Branch = "巳"
	fallenFireTan.Palaces[0].Stars = []StarOutput{
		{Name: "贪狼", Type: "major"},
		{Name: "火星", Type: "tough"},
	}
	if matched, _ := checkHuoTanGe(fallenFireTan); matched {
		t.Fatal("贪狼落巳宫即使同见火星也不符合火贪成格宫位")
	}

	outsideLifeSanfang := chartWithLifePalaceAtZi()
	outsideLifeSanfang.Palaces[1].Stars = []StarOutput{
		{Name: "贪狼", Type: "major"},
		{Name: "火星", Type: "tough"},
	}
	if matched, _ := checkHuoTanGe(outsideLifeSanfang); matched {
		t.Fatal("火贪同宫但贪狼不在命宫三方却误报")
	}
}

func TestQuanLuShengFengRequiresBrightTransformationsInLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{
		{Name: "武曲", Type: "major", Brightness: "庙"},
		{Name: "贪狼", Type: "major", Brightness: "旺"},
	}
	positive.Palaces[0].FourHua = []string{"武曲化权", "贪狼化禄"}
	if matched, name := checkQuanLuShengFeng(positive); !matched || name != "权禄生逢" {
		t.Fatalf("命宫庙旺权禄未命中: matched=%v name=%q", matched, name)
	}

	dim := chartWithLifePalaceAtZi()
	dim.Palaces[0].Stars = []StarOutput{
		{Name: "武曲", Type: "major", Brightness: "陷"},
		{Name: "贪狼", Type: "major", Brightness: "旺"},
	}
	dim.Palaces[0].FourHua = []string{"武曲化权", "贪狼化禄"}
	if matched, _ := checkQuanLuShengFeng(dim); matched {
		t.Fatal("化权星落陷却误报权禄生逢")
	}

	outsideLifePalace := chartWithLifePalaceAtZi()
	outsideLifePalace.Palaces[1].Stars = positive.Palaces[0].Stars
	outsideLifePalace.Palaces[1].FourHua = positive.Palaces[0].FourHua
	if matched, _ := checkQuanLuShengFeng(outsideLifePalace); matched {
		t.Fatal("权禄只在命宫之外却误报权禄生逢")
	}
}

func TestRiYueJiaMingRequiresSunAndMoonOnOppositeSides(t *testing.T) {
	for _, tc := range []struct {
		name        string
		left, right string
	}{
		{name: "太阳左太阴右", left: "太阳", right: "太阴"},
		{name: "太阴左太阳右", left: "太阴", right: "太阳"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			chart.Palaces[11].Stars = []StarOutput{{Name: tc.left, Type: "major"}}
			chart.Palaces[1].Stars = []StarOutput{{Name: tc.right, Type: "major"}}
			if matched, name := checkRiYueJiaMing(chart); !matched || name != "日月夹命" {
				t.Fatalf("日月分居命宫两侧未命中: matched=%v name=%q", matched, name)
			}
		})
	}

	sameSide := chartWithLifePalaceAtZi()
	sameSide.Palaces[11].Stars = []StarOutput{
		{Name: "太阳", Type: "major"},
		{Name: "太阴", Type: "major"},
	}
	if matched, _ := checkRiYueJiaMing(sameSide); matched {
		t.Fatal("日月同在命宫一侧不得误报日月夹命")
	}
}

func TestFuBiJiaMingRequiresStarsOnOppositeSides(t *testing.T) {
	for _, tc := range []struct {
		name        string
		left, right string
	}{
		{name: "左辅左右弼右", left: "左辅", right: "右弼"},
		{name: "右弼左左辅右", left: "右弼", right: "左辅"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chart := chartWithLifePalaceAtZi()
			chart.Palaces[11].Stars = []StarOutput{{Name: tc.left, Type: "soft"}}
			chart.Palaces[1].Stars = []StarOutput{{Name: tc.right, Type: "soft"}}
			if matched, name := checkFuBiJiaMing(chart); !matched || name != "辅弼夹命" {
				t.Fatalf("辅弼分居命宫两侧未命中: matched=%v name=%q", matched, name)
			}
			for _, pattern := range DetectLocalPatterns(chart) {
				if pattern == "辅弼夹印" {
					t.Fatal("仍发布错误名称辅弼夹印")
				}
			}
		})
	}

	sameSide := chartWithLifePalaceAtZi()
	sameSide.Palaces[11].Stars = []StarOutput{
		{Name: "左辅", Type: "soft"},
		{Name: "右弼", Type: "soft"},
	}
	if matched, _ := checkFuBiJiaMing(sameSide); matched {
		t.Fatal("辅弼同在命宫一侧不得误报辅弼夹命")
	}
}

func TestLuMaPeiYinRequiresTianXiangLuCunAndTianMaTogether(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{
		{Name: "天相", Type: "major"},
		{Name: "禄存", Type: "lucun"},
		{Name: "天马", Type: "horse"},
	}
	if matched, name := checkLuMaPeiYin(positive); !matched || name != "禄马佩印" {
		t.Fatalf("天相禄存天马同宫未命中禄马佩印: matched=%v name=%q", matched, name)
	}

	oldWrongCombination := chartWithLifePalaceAtZi()
	oldWrongCombination.Palaces[0].Stars = []StarOutput{
		{Name: "右弼", Type: "soft"},
		{Name: "禄存", Type: "lucun"},
		{Name: "天马", Type: "horse"},
	}
	if matched, _ := checkLuMaPeiYin(oldWrongCombination); matched {
		t.Fatal("右弼禄存天马同宫不得冒充禄马佩印")
	}
}

func TestUnsupportedHighFrequencyPatternAliasesAreNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Stars = []StarOutput{
		{Name: "天魁", Type: "soft"},
		{Name: "天梁", Type: "major"},
		{Name: "右弼", Type: "soft"},
		{Name: "禄存", Type: "lucun"},
		{Name: "天马", Type: "horse"},
	}
	chart.Palaces[1].Name = "田宅"
	chart.Palaces[1].Stars = []StarOutput{{Name: "天府", Type: "major"}}

	for _, pattern := range DetectLocalPatterns(chart) {
		switch pattern {
		case "财印夹马", "天府守垣", "天乙同宫":
			t.Fatalf("无固定来源的高频格局仍在发布: %s", pattern)
		}
	}
	if _, exists := buildPatternTemplates()["天府守垣"]; exists {
		t.Fatal("已撤下的天府守垣仍保留解释模板")
	}
}

func TestRiYueJiaCaiRequiresWuQuBetweenSunAndMoon(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Stars = []StarOutput{{Name: "武曲", Type: "major"}}
	positive.Palaces[11].Stars = []StarOutput{{Name: "太阳", Type: "major"}}
	positive.Palaces[1].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	if matched, name := checkRiYueJiaCai(positive); !matched || name != "日月夹财" {
		t.Fatalf("武曲守命且日月来夹未命中: matched=%v name=%q", matched, name)
	}

	missingWuQu := chartWithLifePalaceAtZi()
	missingWuQu.Palaces[11].Stars = positive.Palaces[11].Stars
	missingWuQu.Palaces[1].Stars = positive.Palaces[1].Stars
	if matched, _ := checkRiYueJiaCai(missingWuQu); matched {
		t.Fatal("命宫无武曲不得误报日月夹财")
	}

	wealthPalaceClamped := chartWithLifePalaceAtZi()
	wealthPalaceClamped.Palaces[4].Name = "财帛"
	wealthPalaceClamped.Palaces[4].Stars = []StarOutput{{Name: "天府", Type: "major"}}
	wealthPalaceClamped.Palaces[3].Stars = []StarOutput{{Name: "太阳", Type: "major"}}
	wealthPalaceClamped.Palaces[5].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	if matched, name := checkRiYueJiaCai(wealthPalaceClamped); !matched || name != "日月夹财" {
		t.Fatalf("财帛宫受日月夹未命中日月夹财: matched=%v name=%q", matched, name)
	}
	wealthPalaceClamped.Patterns = DetectLocalPatterns(wealthPalaceClamped)
	if details := buildPatternDetailsForPalace(wealthPalaceClamped, 4); !patternDetailsContain(details, "日月夹财") {
		t.Fatalf("财帛宫受日月夹但宫位解读缺少依据: %#v", details)
	}
}

func TestKuiYueJiaMingUsesCanonicalNameAndOppositeSides(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[11].Stars = []StarOutput{{Name: "天魁", Type: "soft"}}
	positive.Palaces[1].Stars = []StarOutput{{Name: "天钺", Type: "soft"}}
	if matched, name := checkKuiYueJiaMing(positive); !matched || name != "魁钺夹命" {
		t.Fatalf("魁钺分居命宫两侧未命中魁钺夹命: matched=%v name=%q", matched, name)
	}
	for _, pattern := range DetectLocalPatterns(positive) {
		if pattern == "魁钺夹贵" {
			t.Fatal("仍发布非固定来源名称魁钺夹贵")
		}
	}

	sameSide := chartWithLifePalaceAtZi()
	sameSide.Palaces[11].Stars = []StarOutput{
		{Name: "天魁", Type: "soft"},
		{Name: "天钺", Type: "soft"},
	}
	if matched, _ := checkKuiYueJiaMing(sameSide); matched {
		t.Fatal("魁钺同在命宫一侧不得误报魁钺夹命")
	}
}

func TestZiFuJiaMingRequiresJiYueLifePalaceInYinOrShen(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Name = ""
	positive.Palaces[2].Name = "命宫"
	positive.Palaces[2].Stars = []StarOutput{
		{Name: "天机", Type: "major"},
		{Name: "太阴", Type: "major"},
	}
	positive.Palaces[1].Stars = []StarOutput{{Name: "天府", Type: "major"}}
	positive.Palaces[3].Stars = []StarOutput{{Name: "紫微", Type: "major"}}
	if matched, name := checkZiFuJiaMing(positive); !matched || name != "紫府夹命" {
		t.Fatalf("寅宫机月坐命且紫府分居两侧未命中: matched=%v name=%q", matched, name)
	}
	for _, pattern := range DetectLocalPatterns(positive) {
		if pattern == "紫府夹权" {
			t.Fatal("没有化权条件的错误标签紫府夹权仍在发布")
		}
	}

	missingJiYue := chartWithLifePalaceAtZi()
	missingJiYue.Palaces[0].Name = ""
	missingJiYue.Palaces[2].Name = "命宫"
	missingJiYue.Palaces[1].Stars = positive.Palaces[1].Stars
	missingJiYue.Palaces[3].Stars = positive.Palaces[3].Stars
	if matched, _ := checkZiFuJiaMing(missingJiYue); matched {
		t.Fatal("命宫没有天机太阴不得误报紫府夹命")
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = positive.Palaces[2].Stars
	wrongBranch.Palaces[11].Stars = positive.Palaces[1].Stars
	wrongBranch.Palaces[1].Stars = positive.Palaces[3].Stars
	if matched, _ := checkZiFuJiaMing(wrongBranch); matched {
		t.Fatal("子宫机月坐命不得误报只在寅申成立的紫府夹命")
	}
}

func TestJiXiangLiMingRequiresNoShaInLifeSanfang(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Name = ""
	positive.Palaces[6].Name = "命宫"
	positive.Palaces[6].Stars = []StarOutput{{Name: "紫微", Type: "major", Brightness: "旺"}}
	if matched, name := checkJiXiangLiMing(positive); !matched || name != "极向离明" {
		t.Fatalf("紫微午宫坐命且三方四正无煞未命中: matched=%v name=%q", matched, name)
	}

	shaInLifeSanfang := chartWithLifePalaceAtZi()
	shaInLifeSanfang.Palaces[0].Name = ""
	shaInLifeSanfang.Palaces[6].Name = "命宫"
	shaInLifeSanfang.Palaces[6].Stars = []StarOutput{{Name: "紫微", Type: "major", Brightness: "旺"}}
	shaInLifeSanfang.Palaces[2].Stars = []StarOutput{{Name: "火星", Type: "tough"}}
	if matched, _ := checkJiXiangLiMing(shaInLifeSanfang); matched {
		t.Fatal("命宫三方四正见火星不得误报极向离明")
	}

	shaOutsideLifeSanfang := chartWithLifePalaceAtZi()
	shaOutsideLifeSanfang.Palaces[0].Name = ""
	shaOutsideLifeSanfang.Palaces[6].Name = "命宫"
	shaOutsideLifeSanfang.Palaces[6].Stars = []StarOutput{{Name: "紫微", Type: "major", Brightness: "旺"}}
	shaOutsideLifeSanfang.Palaces[1].Stars = []StarOutput{{Name: "擎羊", Type: "tough"}}
	if matched, name := checkJiXiangLiMing(shaOutsideLifeSanfang); !matched || name != "极向离明" {
		t.Fatalf("三方四正之外见煞不应阻断极向离明: matched=%v name=%q", matched, name)
	}
}

func TestYueLangTianMenRequiresTaiYinInHaiLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Branch = "亥"
	positive.Palaces[0].Stars = []StarOutput{{Name: "太阴", Type: "major", Brightness: "庙"}}
	if matched, name := checkYueLangTianMen(positive); !matched || name != "月朗天门" {
		t.Fatalf("太阴守亥命未命中月朗天门: matched=%v name=%q", matched, name)
	}

	oppositeOnly := chartWithLifePalaceAtZi()
	oppositeOnly.Palaces[0].Branch = "亥"
	oppositeOnly.Palaces[5].Stars = []StarOutput{{Name: "太阴", Type: "major"}}
	if matched, _ := checkYueLangTianMen(oppositeOnly); matched {
		t.Fatal("太阴只在巳宫对照不得误报月朗天门")
	}
}

func TestRiZhaoLeiMenRequiresTaiYangInMaoLifePalace(t *testing.T) {
	positive := chartWithLifePalaceAtZi()
	positive.Palaces[0].Branch = "卯"
	positive.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "庙"}}
	if matched, name := checkRiZhaoLeiMen(positive); !matched || name != "日照雷门" {
		t.Fatalf("太阳守卯命未命中日照雷门: matched=%v name=%q", matched, name)
	}

	wrongBranch := chartWithLifePalaceAtZi()
	wrongBranch.Palaces[0].Stars = []StarOutput{{Name: "太阳", Type: "major", Brightness: "陷"}}
	if matched, _ := checkRiZhaoLeiMen(wrongBranch); matched {
		t.Fatal("太阳守子命不得误报日照雷门")
	}
}

func TestUnsupportedJunZiZaiYeShortcutIsNotPublished(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[11].Stars = []StarOutput{{Name: "紫微", Type: "major", Brightness: "陷"}}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "君子在野" {
			t.Fatal("紫微在亥落陷的无来源捷径仍发布君子在野")
		}
	}
}
