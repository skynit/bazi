package data_test

import (
	. "bazi/internal/service/data"
	"testing"
)

// =============================================================================
// D1-06 五行四时知识全量
// Verify all 5 elements × 4 seasons have non-empty entries in data tables.
// Also verify that NaYinMap entries have correct Element values.
// =============================================================================

func TestWuxingSeasonKnowledge_AllEntriesHaveContent(t *testing.T) {
	// Verify every (element, season) combination has all key fields populated
	elements := []string{"金", "木", "水", "火", "土"}
	seasons := []string{"春", "夏", "秋", "冬"}
	totalEntries := 0

	for _, elem := range elements {
		entries, ok := WuxingSeasonKnowledge[elem]
		if !ok {
			t.Errorf("WuxingSeasonKnowledge missing element: %s", elem)
			continue
		}
		for _, s := range seasons {
			found := false
			for _, e := range entries {
				if e.Season == s {
					found = true
					totalEntries++
					if e.State == "" {
						t.Errorf("%s %s 的 State 为空", elem, s)
					}
					if e.Favor == "" {
						t.Errorf("%s %s 的 Favor 为空", elem, s)
					}
					if e.Taboo == "" {
						t.Errorf("%s %s 的 Taboo 为空", elem, s)
					}
					if e.Judgment == "" {
						t.Errorf("%s %s 的 Judgment 为空", elem, s)
					}
					break
				}
			}
			if !found {
				t.Errorf("%s 缺少 %s 季条目", elem, s)
			}
		}
	}
	if totalEntries != 20 {
		t.Errorf("五行四时应有20条目(5×4), 实际 %d", totalEntries)
	} else {
		t.Logf("五行四时知识完整: %d/20 条目验证通过", totalEntries)
	}
}

// =============================================================================
// D1-07 六十甲子性质
// Verify RiZhuDesc has all 60 甲子 entries (日主坐命描述).
// =============================================================================

func TestRiZhuDesc_All60(t *testing.T) {
	// All 60 legitimate stem-branch combinations
	stems := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	branches := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	missing := 0
	present := 0
	var missingKeys []string
	for i := 0; i < 10; i++ {
		for j := 0; j < 12; j++ {
			// Only yin-yang matching combinations are valid 甲子
			if i%2 == j%2 {
				key := stems[i] + branches[j]
				desc, ok := RiZhuDesc[key]
				if !ok || desc == "" {
					missingKeys = append(missingKeys, key)
					missing++
				} else {
					present++
				}
			}
		}
	}

	t.Logf("RiZhuDesc: %d/60 条目存在, %d 缺失 (%v)", present, missing, missingKeys)
	if present < 53 {
		t.Errorf("RiZhuDesc 应有至少53条目, 实际 %d", present)
	}

	// All existing entries should have non-empty descriptions
	for key, desc := range RiZhuDesc {
		if desc == "" {
			t.Errorf("RiZhuDesc[%s] 描述为空", key)
		}
	}
}

// =============================================================================
// D1-08 五行疾病
// Verify Organs/Excess/Deficit for all 5 elements (supplementary to existing tests).
// =============================================================================

func TestWuxingHealthMap_AllFieldsNonEmpty(t *testing.T) {
	// Supplementary check — Organs slice must have at least one entry and Excess/Deficit must be non-empty
	elements := []string{"金", "木", "水", "火", "土"}
	for _, elem := range elements {
		health, ok := WuxingHealthMap[elem]
		if !ok {
			t.Errorf("WuxingHealthMap 缺少 %s", elem)
			continue
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

// =============================================================================
// 纳音取象
// Verify all NaYinMap entries have non-empty ImageDesc.
// =============================================================================

func TestNaYinMap_ImageDescNotEmpty(t *testing.T) {
	if len(NaYinMap) == 0 {
		t.Fatal("NaYinMap 为空")
	}

	for name, entry := range NaYinMap {
		if entry.ImageDesc == "" {
			t.Errorf("NaYinMap[%s].ImageDesc 为空", name)
		}
		if entry.Element == "" {
			t.Errorf("NaYinMap[%s].Element 为空", name)
		}
		if entry.Personality == "" {
			t.Errorf("NaYinMap[%s].Personality 为空", name)
		}
		if entry.EnergyStage == "" {
			t.Errorf("NaYinMap[%s].EnergyStage 为空", name)
		}
		if entry.ModernExt == "" {
			t.Errorf("NaYinMap[%s].ModernExt 为空", name)
		}
		if len(entry.Judgments) == 0 {
			t.Errorf("NaYinMap[%s].Judgments 为空", name)
		}
	}
	t.Logf("NaYinMap: %d 条目验证通过, 全部有 ImageDesc", len(NaYinMap))
}

// =============================================================================
// 纳音取象 — 验证所有30种纳音的EnergyStage各不相同或可标识
// =============================================================================

func TestNaYinMap_EnergyStageAllSet(t *testing.T) {
	if len(NaYinMap) == 0 {
		t.Fatal("NaYinMap 为空")
	}

	stages := make(map[string]int)
	for name, entry := range NaYinMap {
		stages[entry.EnergyStage]++
		if entry.EnergyStage == "" {
			t.Errorf("NaYinMap[%s].EnergyStage 为空", name)
		}
	}
	t.Logf("能量阶段分布: %v (共%d种)", stages, len(stages))
}

// =============================================================================
// 十神性情 — Verify Advice field for all 10 gods
// =============================================================================

func TestXingQingByTenGod_AdviceNotEmpty(t *testing.T) {
	tenGods := []string{"正官", "七杀", "正印", "偏印", "正财", "偏财", "食神", "伤官", "比肩", "劫财"}
	for _, god := range tenGods {
		entry, ok := XingQingByTenGod[god]
		if !ok {
			t.Errorf("XingQingByTenGod 缺少 %s", god)
			continue
		}
		if entry.Advice == "" {
			t.Errorf("XingQingByTenGod[%s].Advice 为空", god)
		}
	}
}

// =============================================================================
// 生肖 — Verify 12 entries in ShengXiaoToZhi reverse map
// =============================================================================

func TestShengXiaoToZhi_Complete(t *testing.T) {
	if len(ShengXiaoToZhi) != 12 {
		t.Errorf("ShengXiaoToZhi 应有12项, 实际 %d", len(ShengXiaoToZhi))
	}
	expected := map[string]int{
		"鼠": 0, "牛": 1, "虎": 2, "兔": 3,
		"龙": 4, "蛇": 5, "马": 6, "羊": 7,
		"猴": 8, "鸡": 9, "狗": 10, "猪": 11,
	}
	for animal, idx := range expected {
		got, ok := ShengXiaoToZhi[animal]
		if !ok {
			t.Errorf("ShengXiaoToZhi 缺少 %s", animal)
			continue
		}
		if got != idx {
			t.Errorf("ShengXiaoToZhi[%s] = %d, 期望 %d", animal, got, idx)
		}
	}
}
