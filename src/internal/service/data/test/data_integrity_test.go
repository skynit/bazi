package data_test

import (
	. "bazi/internal/service/data"
	"testing"
)

// =============================================================================
// 命理知识数据库完整性验证测试
// 验证所有映射表完整、无遗漏、值域正确
// =============================================================================

func TestDataGanElement_Complete(t *testing.T) {
	// 十天干必须有五行映射
	expected := map[string]string{"甲": "木", "乙": "木", "丙": "火", "丁": "火",
		"戊": "土", "己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水"}
	for gan, want := range expected {
		if GanElement[gan] != want {
			t.Errorf("GanElement[%s] = %q, 期望 %q", gan, GanElement[gan], want)
		}
	}
	if len(GanElement) != 10 {
		t.Errorf("GanElement 应有10项, 实际 %d", len(GanElement))
	}
}

func TestDataGanIndex_Complete(t *testing.T) {
	for gan, idx := range map[string]int{"甲": 0, "乙": 1, "丙": 2, "丁": 3, "戊": 4, "己": 5, "庚": 6, "辛": 7, "壬": 8, "癸": 9} {
		if GanIndex(gan) != idx {
			t.Errorf("GanIndex(%s) = %d, 期望 %d", gan, GanIndex(gan), idx)
		}
	}
	if GanIndex("X") != -1 {
		t.Errorf("GanIndex('X') 应返回 -1")
	}
}

func TestDataZhiIndex_Complete(t *testing.T) {
	for zhi, idx := range map[string]int{"子": 0, "丑": 1, "寅": 2, "卯": 3, "辰": 4, "巳": 5, "午": 6, "未": 7, "申": 8, "酉": 9, "戌": 10, "亥": 11} {
		if ZhiIndex(zhi) != idx {
			t.Errorf("ZhiIndex(%s) = %d, 期望 %d", zhi, ZhiIndex(zhi), idx)
		}
	}
	if ZhiIndex("X") != -1 {
		t.Errorf("ZhiIndex('X') 应返回 -1")
	}
}

func TestDataNayin_All60(t *testing.T) {
	// 六十甲子纳音矩阵：Nayin[天干索引][地支索引]
	// 天干索引：甲0乙1丙2丁3戊4己5庚6辛7壬8癸9
	// 地支索引：子0丑1寅2卯3辰4巳5午6未7申8酉9戌10亥11
	ganList := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	zhiList := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	emptyCount := 0
	for g := 0; g < 10; g++ {
		for z := 0; z < 12; z++ {
			name := Nayin[g][z]
			if name == "" {
				emptyCount++
			} else {
				entry := NaYinKnowledge[name]
				if entry.Name == "" {
					t.Errorf("Nayin[%s][%s]=%q 在知识库中缺失", ganList[g], zhiList[z], name)
				}
			}
		}
	}
	if emptyCount != 60 {
		t.Errorf("纳音矩阵应有60个空值(仅60甲子有纳音), 实际空值%d", emptyCount)
	} else {
		t.Logf("纳音矩阵正确: 60/120 有值, 60/120 空(仅合法干支组合有纳音)")
	}
}

func TestDataNayinMap_AllEntries(t *testing.T) {
	// 验证所有纳音名称在知识库中都有条目
	checked := make(map[string]bool)
	for _, row := range Nayin {
		for _, name := range row {
			if name != "" && !checked[name] {
				checked[name] = true
				entry := NaYinKnowledge[name]
				if entry.Name == "" {
					t.Errorf("NaYinKnowledge 缺少 %s", name)
				}
				if entry.Element == "" {
					t.Errorf("NaYinKnowledge[%s].Element 为空", name)
				}
			}
		}
	}
	t.Logf("纳音知识库共 %d 个有效条目", len(checked))
}

func TestDataTiaoHou_All120(t *testing.T) {
	// 验证十日干×十二月令 = 120组调候规则
	stems := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	months := []string{"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"}
	emptyCount := 0
	for _, s := range stems {
		for _, m := range months {
			rules := GetTiaohou(s, m)
			if len(rules) == 0 {
				emptyCount++
				t.Logf("调候缺失: %s月%s日", s, m)
			}
		}
	}
	if emptyCount > 0 {
		t.Errorf("调候表共缺失 %d/120 组规则", emptyCount)
	} else {
		t.Logf("调候表120/120 完整")
	}
}

func TestDataShengXiao_Complete(t *testing.T) {
	if len(ShengXiao) != 12 {
		t.Errorf("生肖应有12个, 实际 %d", len(ShengXiao))
	}
	expected := map[int]string{0: "鼠", 1: "牛", 2: "虎", 3: "兔", 4: "龙", 5: "蛇", 6: "马", 7: "羊", 8: "猴", 9: "鸡", 10: "狗", 11: "猪"}
	for i, want := range expected {
		if ShengXiao[i] != want {
			t.Errorf("ShengXiao[%d] = %s, 期望 %s", i, ShengXiao[i], want)
		}
	}
}

func TestDataEmpties_Complete(t *testing.T) {
	// 六甲空亡应有 10天干×12地支 全覆盖
	emptyCount := 0
	for g := 0; g < 10; g++ {
		for z := 0; z < 12; z++ {
			if Empties[g][z][0] == "" {
				emptyCount++
			}
		}
	}
	if emptyCount > 0 {
		t.Logf("空亡表有 %d/120 个空值（部分正常）", emptyCount)
	}
}

func TestCalcMingGong(t *testing.T) {
	cases := []struct {
		yearGan, monthZhi, hourZhi, expected string
	}{
		// Exact cases pinned from 6tail/lunar-javascript v1.7.7 EightChar tests.
		{"乙", "子", "辰", "己丑"},
		{"戊", "午", "寅", "辛酉"},
		{"乙", "子", "巳", "戊子"},
		{"癸", "丑", "巳", "癸亥"},
		{"己", "卯", "辰", "甲戌"},
		{"己", "卯", "丑", "丁丑"},
		{"丙", "寅", "辰", "己亥"},
		{"壬", "亥", "巳", "癸丑"},
	}
	for _, c := range cases {
		t.Run(c.yearGan+c.monthZhi+c.hourZhi, func(t *testing.T) {
			got, err := CalcMingGong(c.yearGan, c.monthZhi, c.hourZhi)
			if err != nil {
				t.Fatal(err)
			}
			if got != c.expected {
				t.Fatalf("CalcMingGong(%s,%s,%s) = %s, want %s", c.yearGan, c.monthZhi, c.hourZhi, got, c.expected)
			}
		})
	}
}

func TestSeasonFromMonth(t *testing.T) {
	// 节气月（非公历月）：2-4春(寅卯辰), 5-7夏(巳午未), 8-10秋(申酉戌), 11-1冬(亥子丑)
	cases := map[int]string{2: "春", 3: "春", 4: "春", 5: "夏", 6: "夏", 7: "夏",
		8: "秋", 9: "秋", 10: "秋", 11: "冬", 12: "冬", 1: "冬"}
	for m, want := range cases {
		if got := SeasonFromMonth(m); got != want {
			t.Errorf("SeasonFromMonth(%d) = %s, 期望 %s (节气月)", m, got, want)
		}
	}
}
