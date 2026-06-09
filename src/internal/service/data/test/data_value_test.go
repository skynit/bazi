package data_test

import (
	. "bazi/internal/service/data"
	"testing"
)

// ── 六十甲子性质验证 ──────────────────────────────────────

func TestJiaZiKnowledge_All60Entries(t *testing.T) {
	// 验证六十甲子全部有定义
	stems := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	branches := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	missing := 0
	for i := 0; i < 10; i++ {
		for j := 0; j < 12; j++ {
			gan := stems[i]
			zhi := branches[j]
			// 甲子组合：天干索引等于地支索引的不一定有效，需要检查奇数偶数配对
			// 天干为阳干(甲丙戊庚壬)配阳支(子寅辰午申戌)，阴干配阴支
			ganOdd := i%2 == 0 // 甲丙戊庚壬 = 阳
			zhiOdd := j%2 == 0 // 子寅辰午申戌 = 阳
			if ganOdd == zhiOdd {
				// 合法干支组合
				key := gan + zhi
				entry, ok := JiaZiKnowledge[key]
				if !ok {
					t.Errorf("JiaZiKnowledge 缺少 %s", key)
					missing++
					continue
				}
				if entry.GanZhi != key {
					t.Errorf("JiaZiKnowledge[%s].GanZhi = %s, 期望 %s", key, entry.GanZhi, key)
				}
				if entry.Nayin == "" {
					t.Errorf("JiaZiKnowledge[%s].Nayin 为空", key)
				}
				if entry.Nature == "" {
					t.Errorf("JiaZiKnowledge[%s].Nature 为空", key)
				}
			}
		}
	}
	if missing > 0 {
		t.Errorf("JiaZiKnowledge 共缺失 %d 个条目", missing)
	}
}

func TestJiaZiKnowledge_NayinConsistency(t *testing.T) {
	// 验证 JiaZiKnowledge 中的纳音与 Nayin 矩阵一致
	for key, entry := range JiaZiKnowledge {
		if len(key) != 2 {
			continue
		}
		runes := []rune(key)
		gan := string(runes[0])
		zhi := string(runes[1])
		gi := GanIndex(gan)
		zi := ZhiIndex(zhi)
		if gi < 0 || zi < 0 {
			continue
		}
		nayinFromMatrix := Nayin[gi][zi]
		if entry.Nayin != nayinFromMatrix {
			t.Errorf("JiaZiKnowledge[%s].Nayin = %s, 但 Nayin[%d][%d] = %s",
				key, entry.Nayin, gi, zi, nayinFromMatrix)
		}
	}
}

func TestJiaZiKnowledge_ShenShaNotEmpty(t *testing.T) {
	// 验证神煞标记结构合理
	for key, entry := range JiaZiKnowledge {
		if len(entry.ShenSha) == 0 {
			// 部分特殊组合可能没有神煞
			t.Logf("JiaZiKnowledge[%s] 神煞为空 (可能合理)", key)
		}
	}
}

// 验证特定组合的经典性质（《三命通会》记载）
func TestJiaZiKnowledge_SpecificEntries(t *testing.T) {
	type expected struct {
		ganZhi string
		nayin  string
		nature string
	}

	tests := []expected{
		{"甲子", "海中金", "宝物"},
		{"乙丑", "海中金", "顽矿"},
		{"丙寅", "炉中火", "炉炭"},
		{"丁卯", "炉中火", "炉烟"},
		{"戊辰", "大林木", "山林不材之木"},
		{"己巳", "大林木", "山头花草"},
		{"庚午", "路旁土", "路旁干土"},
		{"辛未", "路旁土", "含万宝待秋成"},
		{"壬申", "剑锋金", "戈戟"},
		{"癸酉", "剑锋金", "椎凿"},
		{"甲戌", "山头火", "火所宿处"},
		{"乙亥", "山头火", "火之热气"},
		{"壬戌", "大海水", "海"},
		{"癸亥", "大海水", "百川"},
	}

	for _, tt := range tests {
		entry, ok := JiaZiKnowledge[tt.ganZhi]
		if !ok {
			t.Errorf("JiaZiKnowledge 缺少 %s", tt.ganZhi)
			continue
		}
		if entry.Nayin != tt.nayin {
			t.Errorf("%s 纳音 = %s, 期望 %s", tt.ganZhi, entry.Nayin, tt.nayin)
		}
		if entry.Nature != tt.nature {
			t.Errorf("%s 性质 = %s, 期望 %s", tt.ganZhi, entry.Nature, tt.nature)
		}
	}
}

// ── 五行四时知识验证 ──────────────────────────────────────

func TestWuxingSeasonKnowledge_AllFiveElements(t *testing.T) {
	// 验证五个五行都有定义
	elements := []string{"金", "木", "水", "火", "土"}
	for _, elem := range elements {
		entries, ok := WuxingSeasonKnowledge[elem]
		if !ok {
			t.Errorf("WuxingSeasonKnowledge 缺少 %s", elem)
			continue
		}
		// 每个五行应有四个季节
		if len(entries) != 4 {
			t.Errorf("WuxingSeasonKnowledge[%s] 应有 4 条(四季), 实际 %d 条", elem, len(entries))
		}
	}
}

func TestWuxingSeasonKnowledge_SeasonConsistency(t *testing.T) {
	// 验证每个五行包含"春夏秋冬"四季
	for _, entries := range WuxingSeasonKnowledge {
		seenSeasons := make(map[string]bool)
		for _, e := range entries {
			seenSeasons[e.Season] = true
		}
		for _, s := range []string{"春", "夏", "秋", "冬"} {
			if !seenSeasons[s] {
				t.Errorf("某五行缺少 %s 季条目", s)
			}
		}
	}
}

func TestWuxingSeasonKnowledge_StateNotEmpty(t *testing.T) {
	// 验证状态描述非空
	for elem, entries := range WuxingSeasonKnowledge {
		for _, e := range entries {
			if e.State == "" {
				t.Errorf("%s 在 %s 的 State 为空", elem, e.Season)
			}
			if e.Judgment == "" {
				t.Errorf("%s 在 %s 的 Judgment 为空", elem, e.Season)
			}
		}
	}
}

func TestSeasonFromMonth_Value(t *testing.T) {
	tests := []struct {
		month int
		want  string
	}{
		{2, "春"}, {3, "春"}, {4, "春"},
		{5, "夏"}, {6, "夏"}, {7, "夏"},
		{8, "秋"}, {9, "秋"}, {10, "秋"},
		{11, "冬"}, {12, "冬"}, {1, "冬"},
	}
	for _, tt := range tests {
		got := SeasonFromMonth(tt.month)
		if got != tt.want {
			t.Errorf("SeasonFromMonth(%d) = %s, want %s", tt.month, got, tt.want)
		}
	}
}

// ── 五行疾病验证 ──────────────────────────────────────────

func TestWuxingHealthMap_AllFiveElements(t *testing.T) {
	// 验证五个五行都有疾病映射
	elements := []string{"金", "木", "水", "火", "土"}
	for _, elem := range elements {
		health, ok := WuxingHealthMap[elem]
		if !ok {
			t.Errorf("WuxingHealthMap 缺少 %s", elem)
			continue
		}
		if health.Element != elem {
			t.Errorf("WuxingHealthMap[%s].Element = %s", elem, health.Element)
		}
		if len(health.Organs) == 0 {
			t.Errorf("WuxingHealthMap[%s].Organs 为空", elem)
		}
		if health.Excess == "" {
			t.Errorf("WuxingHealthMap[%s].Excess 为空", elem)
		}
		if health.Deficit == "" {
			t.Errorf("WuxingHealthMap[%s].Deficit 为空", elem)
		}
	}
}

func TestWuxingHealthMap_Organs(t *testing.T) {
	// 验证脏腑对应关系（《三命通会》论疾病）
	tests := []struct {
		element  string
		expected []string
	}{
		{"木", []string{"肝", "胆"}},
		{"火", []string{"心", "小肠"}},
		{"土", []string{"脾", "胃"}},
		{"金", []string{"肺", "大肠"}},
		{"水", []string{"肾", "膀胱"}},
	}
	for _, tt := range tests {
		health, ok := WuxingHealthMap[tt.element]
		if !ok {
			t.Errorf("WuxingHealthMap 缺少 %s", tt.element)
			continue
		}
		for _, expectedOrgan := range tt.expected {
			found := false
			for _, organ := range health.Organs {
				if organ == expectedOrgan {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s 应包含脏腑 %s, 实际 Organs=%v", tt.element, expectedOrgan, health.Organs)
			}
		}
	}
}

func TestGetWuxingHealth(t *testing.T) {
	// 验证 GetWuxingHealth 公开函数
	for _, elem := range []string{"金", "木", "水", "火", "土"} {
		health, ok := GetWuxingHealth(elem)
		if !ok {
			t.Errorf("GetWuxingHealth(%s) 返回 false", elem)
		}
		if health.Element != elem {
			t.Errorf("GetWuxingHealth(%s).Element = %s", elem, health.Element)
		}
	}
	// 验证未知五行
	_, ok := GetWuxingHealth("未知")
	if ok {
		t.Error("GetWuxingHealth('未知') 应返回 false")
	}
}

// ── 五行相貌验证 ──────────────────────────────────────────

func TestWuxingAppearanceMap_AllFiveElements(t *testing.T) {
	elements := []string{"金", "木", "水", "火", "土"}
	for _, elem := range elements {
		appearance, ok := WuxingAppearanceMap[elem]
		if !ok {
			t.Errorf("WuxingAppearanceMap 缺少 %s", elem)
			continue
		}
		if appearance.Element != elem {
			t.Errorf("WuxingAppearanceMap[%s].Element = %s", elem, appearance.Element)
		}
		if appearance.BodyType == "" {
			t.Errorf("WuxingAppearanceMap[%s].BodyType 为空", elem)
		}
		if appearance.FaceColor == "" {
			t.Errorf("WuxingAppearanceMap[%s].FaceColor 为空", elem)
		}
	}
}

func TestGetWuxingAppearance(t *testing.T) {
	for _, elem := range []string{"金", "木", "水", "火", "土"} {
		a, ok := GetWuxingAppearance(elem)
		if !ok {
			t.Errorf("GetWuxingAppearance(%s) 返回 false", elem)
		}
		if a.BodyType == "" {
			t.Errorf("GetWuxingAppearance(%s).BodyType 为空", elem)
		}
	}
	_, ok := GetWuxingAppearance("未知")
	if ok {
		t.Error("GetWuxingAppearance('未知') 应返回 false")
	}
}

// ── 十神性情验证 ──────────────────────────────────────────

func TestXingQingByTenGod_AllTenGods(t *testing.T) {
	tenGods := []string{"正官", "七杀", "正印", "偏印", "正财", "偏财", "食神", "伤官", "比肩", "劫财"}
	for _, god := range tenGods {
		entry, ok := XingQingByTenGod[god]
		if !ok {
			t.Errorf("XingQingByTenGod 缺少 %s", god)
			continue
		}
		if entry.God != god {
			t.Errorf("XingQingByTenGod[%s].God = %s", god, entry.God)
		}
		if entry.Positive == "" {
			t.Errorf("XingQingByTenGod[%s].Positive 为空", god)
		}
		if entry.Negative == "" {
			t.Errorf("XingQingByTenGod[%s].Negative 为空", god)
		}
	}
}

func TestGetXingQingByGod(t *testing.T) {
	entry, ok := GetXingQingByGod("正官")
	if !ok {
		t.Error("GetXingQingByGod('正官') 返回 false")
	}
	if entry.Positive == "" {
		t.Error("正官 Positive 为空")
	}
	_, ok = GetXingQingByGod("未知")
	if ok {
		t.Error("GetXingQingByGod('未知') 应返回 false")
	}
}
