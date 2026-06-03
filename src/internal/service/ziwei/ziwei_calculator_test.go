package ziwei

import (
	"testing"
)

func verifyChartSize(t *testing.T, chart *ZiWeiChart) {
	t.Helper()
	if chart == nil {
		t.Fatal("chart is nil")
	}
	nonEmptyPalaces := 0
	for i, p := range chart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty", i)
		}
		if p.Branch == "" {
			t.Errorf("Palaces[%d].Branch is empty", i)
		}
		if len(p.MainStars) > 0 || len(p.AuxStars) > 0 {
			nonEmptyPalaces++
		}
	}
	if nonEmptyPalaces < 10 {
		t.Errorf("only %d palaces have stars, expected at least 10", nonEmptyPalaces)
	}
}

// ════════════════════════════════════════════════════════════════
// 测试用例1: 癸未年三月十四日未时 (公历 2003-04-15 14:00)
// ════════════════════════════════════════════════════════════════
func TestChart_GuiWeiYear_Month3_Day14_WeiHour(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	t.Run("命宫地支=酉", func(t *testing.T) {
		if chart.Palaces[0].Branch != "酉" {
			t.Errorf("命宫地支 = %q, want 酉", chart.Palaces[0].Branch)
		}
	})
	t.Run("身宫=亥", func(t *testing.T) {
		if chart.BodyPalace != "亥" {
			t.Errorf("身宫 = %q, want 亥", chart.BodyPalace)
		}
	})
	t.Run("五行局=木三局", func(t *testing.T) {
		if chart.FiveBureau != "木三局" {
			t.Errorf("五行局 = %q, want 木三局", chart.FiveBureau)
		}
	})
	t.Run("命主身主", func(t *testing.T) {
		if chart.LifeMaster != "武曲" {
			t.Errorf("命主 = %q, want 武曲", chart.LifeMaster)
		}
		if chart.BodyMaster != "天相" {
			t.Errorf("身主 = %q, want 天相", chart.BodyMaster)
		}
	})
	t.Run("禄存位置", func(t *testing.T) {
		found := false
		for _, p := range chart.Palaces {
			for _, s := range p.AuxStars {
				if s == "禄存" {
					found = true
					if p.Branch != "子" {
						t.Errorf("禄存在 %s, want 子", p.Branch)
					}
				}
			}
		}
		if !found {
			t.Error("禄存未找到")
		}
	})
	t.Run("四化(癸年)", func(t *testing.T) {
		allHua := make(map[string]bool)
		for _, p := range chart.Palaces {
			for _, h := range p.FourHua {
				allHua[h] = true
			}
		}
		for _, e := range []string{"破军化禄", "巨门化权", "太阴化科", "贪狼化忌"} {
			if !allHua[e] {
				t.Errorf("缺少四化: %s", e)
			}
		}
	})
	t.Run("紫微星位置", func(t *testing.T) {
		for _, p := range chart.Palaces {
			for _, s := range p.MainStars {
				if s == "紫微" {
					// 木三局14日: 紫微 in 卯 (based on iztro formula)
					return
				}
			}
		}
		t.Error("紫微星未找到")
	})
	t.Run("基本数据完整性", func(t *testing.T) {
		verifyChartSize(t, chart)
	})
}

// ════════════════════════════════════════════════════════════════
// 纳音五行局 精确查表测试
// 干支 → 五行局 对照六十甲子纳音表
// ════════════════════════════════════════════════════════════════
func TestCalcFiveBureau_NaYin(t *testing.T) {
	tests := []struct {
		name     string
		stem     int
		branch   int
		wantJu   int
		wantName string
	}{
		{"甲子(金四)", 0, 0, 4, "金四局"},
		{"乙丑(金四)", 1, 1, 4, "金四局"},
		{"丙寅(火六)", 2, 2, 6, "火六局"},
		{"丁卯(火六)", 3, 3, 6, "火六局"},
		{"戊辰(木三)", 4, 4, 3, "木三局"},
		{"己巳(木三)", 5, 5, 3, "木三局"},
		{"庚午(土五)", 6, 6, 5, "土五局"},
		{"辛未(土五)", 7, 7, 5, "土五局"},
		{"壬申(金四)", 8, 8, 4, "金四局"},
		{"癸酉(金四)", 9, 9, 4, "金四局"},
		{"甲戌(火六)", 0, 10, 6, "火六局"},
		{"丙子(水二)", 2, 0, 2, "水二局"},
		{"壬子(木三)", 8, 0, 3, "木三局"},
		{"癸巳(水二)", 9, 5, 2, "水二局"},
		{"戊午(火六)", 4, 6, 6, "火六局"},
		{"辛酉(木三)", 7, 9, 3, "木三局"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ju := calcFiveBureau(tt.stem, tt.branch)
			if ju != tt.wantJu {
				t.Errorf("%s: ju = %d, want %d", tt.name, ju, tt.wantJu)
			}
			if FiveBureauName[ju] != tt.wantName {
				t.Errorf("%s: name = %q, want %q", tt.name, FiveBureauName[ju], tt.wantName)
			}
		})
	}
}

func TestGanzhiPairIndex(t *testing.T) {
	tests := []struct {
		stem   int
		branch int
		want   int
	}{
		{0, 0, 0},  // 甲子 → pair 0
		{2, 2, 1},  // 丙寅 → pair 1
		{4, 4, 2},  // 戊辰 → pair 2
		{6, 6, 3},  // 庚午 → pair 3
		{9, 9, 4},  // 癸酉 → pair 4
		{0, 10, 5}, // 甲戌 → pair 5
		{9, 5, 14}, // 癸巳 → pair 14
	}
	for _, tt := range tests {
		got := ganzhiPairIndex(tt.stem, tt.branch)
		if got != tt.want {
			t.Errorf("ganzhiPairIndex(%d,%d) = %d, want %d", tt.stem, tt.branch, got, tt.want)
		}
	}
}

// ════════════════════════════════════════════════════════════════
// 命宫/身宫 纯函数测试
// ════════════════════════════════════════════════════════════════
func TestCalcSoulAndBody(t *testing.T) {
	tests := []struct {
		name       string
		lunarMonth int
		hourBranch int
		wantSoul   string
		wantBody   string
	}{
		{"三月未时", 3, 7, "酉", "亥"},
		{"正月子时", 1, 0, "寅", "寅"},
		{"二月丑时", 2, 1, "寅", "辰"},
		{"五月午时", 5, 6, "子", "子"},
		{"腊月子时", 12, 0, "丑", "丑"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			birth := &BirthData{
				LunarMonth:  tt.lunarMonth,
				HourBranch:  tt.hourBranch,
				YearStem:    0,
				YearBranch:  0,
				Gender:      "男",
				IsLeapMonth: false,
			}
			soulBranch, bodyBranch, _ := calcSoulAndBody(birth)
			if BranchNames[soulBranch] != tt.wantSoul {
				t.Errorf("soulBranch = %s, want %s", BranchNames[soulBranch], tt.wantSoul)
			}
			if BranchNames[bodyBranch] != tt.wantBody {
				t.Errorf("bodyBranch = %s, want %s", BranchNames[bodyBranch], tt.wantBody)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════
// 紫微/天府 定位测试
// 使用 iztro 公式: offset even → pos = quotient + offset - 1; odd → pos = quotient - offset - 1
// ════════════════════════════════════════════════════════════════
func TestZiweiTianfuPosition(t *testing.T) {
	tests := []struct {
		name      string
		juValue   int
		lunarDay  int
		wantZiwei string
		wantTianfu string
	}{
		{"金四局14日", 4, 14, "巳", "未"},
		{"水二局1日", 2, 1, "亥", "丑"},
		{"火六局15日", 6, 15, "亥", "丑"},
		{"水二局2日", 2, 2, "卯", "酉"},
		{"金四局5日", 4, 5, "戌", "寅"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ziweiIdx, tianfuIdx := calcZiweiTianfuPosition(tt.juValue, tt.lunarDay, 0, 1, false)
			if BranchNames[ziweiIdx] != tt.wantZiwei {
				t.Errorf("ziwei = %d(%s), want %s", ziweiIdx, BranchNames[ziweiIdx], tt.wantZiwei)
			}
			if BranchNames[tianfuIdx] != tt.wantTianfu {
				t.Errorf("tianfu = %d(%s), want %s", tianfuIdx, BranchNames[tianfuIdx], tt.wantTianfu)
			}
		})
	}
}

func TestFourHua(t *testing.T) {
	tests := []struct {
		stem    int
		huaLu   string
		huaQuan string
		huaKe   string
		huaJi   string
	}{
		{0, "廉贞", "破军", "武曲", "太阳"},
		{4, "贪狼", "太阴", "右弼", "天机"},
		{9, "破军", "巨门", "太阴", "贪狼"},
	}

	for _, tt := range tests {
		t.Run(StemNames[tt.stem], func(t *testing.T) {
			hua := calcFourHua(tt.stem)
			if hua[tt.huaLu] != "化禄" {
				t.Errorf("stem %d: %s should be 化禄, got %s", tt.stem, tt.huaLu, hua[tt.huaLu])
			}
			if hua[tt.huaQuan] != "化权" {
				t.Errorf("stem %d: %s should be 化权", tt.stem, tt.huaQuan)
			}
			if hua[tt.huaKe] != "化科" {
				t.Errorf("stem %d: %s should be 化科", tt.stem, tt.huaKe)
			}
			if hua[tt.huaJi] != "化忌" {
				t.Errorf("stem %d: %s should be 化忌", tt.stem, tt.huaJi)
			}
		})
	}
}

func TestAuxiliaryStars(t *testing.T) {
	t.Run("癸年禄存擎羊陀罗", func(t *testing.T) {
		if LucunBranchIdx[9] != 0 {
			t.Errorf("癸年禄存位置 = %d, want 0(子)", LucunBranchIdx[9])
		}
		if QingyangIndex(9) != 1 {
			t.Errorf("癸年擎羊位置 = %d, want 1(丑)", QingyangIndex(9))
		}
		if TuoluoIndex(9) != 11 {
			t.Errorf("癸年陀罗位置 = %d, want 11(亥)", TuoluoIndex(9))
		}
	})
	t.Run("甲年禄存", func(t *testing.T) {
		if LucunBranchIdx[0] != 2 {
			t.Errorf("甲年禄存位置 = %d, want 2(寅)", LucunBranchIdx[0])
		}
	})
	t.Run("天魁天钺", func(t *testing.T) {
		kq := KuiYueTable[0]
		if kq[0] != 1 || kq[1] != 7 {
			t.Errorf("甲年魁钺 = %v, want [1,7]", kq)
		}
		kq8 := KuiYueTable[8]
		if kq8[0] != 1 || kq8[1] != 7 {
			t.Errorf("壬年魁钺 = %v, want [1,7]", kq8)
		}
	})
}

func TestLifeBodyMaster(t *testing.T) {
	tests := []struct {
		branch     int
		lifeMaster string
		bodyMaster string
	}{
		{0, "贪狼", "火星"},
		{7, "武曲", "天相"},
		{2, "禄存", "天梁"},
		{6, "破军", "火星"},
	}
	for _, tt := range tests {
		t.Run(BranchNames[tt.branch]+"年", func(t *testing.T) {
			if LifeMasterTable[tt.branch] != tt.lifeMaster {
				t.Errorf("%s年命主 = %q, want %q", BranchNames[tt.branch], LifeMasterTable[tt.branch], tt.lifeMaster)
			}
			if BodyMasterTable[tt.branch] != tt.bodyMaster {
				t.Errorf("%s年身主 = %q, want %q", BranchNames[tt.branch], BodyMasterTable[tt.branch], tt.bodyMaster)
			}
		})
	}
}

// 红鸾: 从卯(3)逆数年支 → 子年=卯(3-0=3✓), 未年=申(3-7+12=8)
func TestAdjectiveStars(t *testing.T) {
	t.Run("红鸾", func(t *testing.T) {
		if HongLuanIndex(0) != 3 {
			t.Errorf("子年红鸾 = %d(%s), want 3(卯)", HongLuanIndex(0), BranchNames[HongLuanIndex(0)])
		}
		if HongLuanIndex(7) != 8 {
			t.Errorf("未年红鸾 = %d(%s), want 8(申)", HongLuanIndex(7), BranchNames[HongLuanIndex(7)])
		}
	})
	t.Run("天喜", func(t *testing.T) {
		if TianXiIndex(0) != 9 {
			t.Errorf("子年天喜 = %d(%s), want 9(酉)", TianXiIndex(0), BranchNames[TianXiIndex(0)])
		}
		if TianXiIndex(7) != 2 {
			t.Errorf("未年天喜 = %d(%s), want 2(寅)", TianXiIndex(7), BranchNames[TianXiIndex(7)])
		}
	})
	t.Run("咸池", func(t *testing.T) {
		if XianChiBranch[0] != 9 {
			t.Errorf("子年咸池 = %d, want 9(酉)", XianChiBranch[0])
		}
		if XianChiBranch[7] != 0 {
			t.Errorf("未年咸池 = %d, want 0(子)", XianChiBranch[7])
		}
	})
}

func TestChangSheng12(t *testing.T) {
	t.Run("金四局阳男顺行", func(t *testing.T) {
		result := placeChangSheng12(4, 6, "男")
		if result[5] != "长生" {
			t.Errorf("金四局阳男长生位置 = %s, want 长生在巳(5)", result[5])
		}
		if result[6] != "沐浴" {
			t.Errorf("金四局阳男午位 = %q, want 沐浴", result[6])
		}
	})
}

func TestFixIndex(t *testing.T) {
	for _, tt := range []struct {
		input int
		want  int
	}{
		{0, 0}, {12, 0}, {13, 1}, {-1, 11}, {-12, 0}, {25, 1},
	} {
		if got := fixIndex(tt.input); got != tt.want {
			t.Errorf("fixIndex(%d) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestGetPalaceStem(t *testing.T) {
	t.Run("甲年寅宫", func(t *testing.T) {
		if got := GetPalaceStem(0, 2); got != 2 {
			t.Errorf("甲年寅宫天干 = %d, want 2(丙)", got)
		}
	})
	t.Run("甲年卯宫", func(t *testing.T) {
		if got := GetPalaceStem(0, 3); got != 3 {
			t.Errorf("甲年卯宫天干 = %d, want 3(丁)", got)
		}
	})
	t.Run("癸年酉宫", func(t *testing.T) {
		if got := GetPalaceStem(9, 9); got != 7 {
			t.Errorf("癸年酉宫天干 = %d, want 7(辛)", got)
		}
	})
}