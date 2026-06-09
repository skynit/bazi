package ziwei

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// B8-02: 三方四正 计算测试
//
// For each palace, verify its 三方四正 (三合 + 对宫):
//   - 对宫(opposite): palaceIdx + 6 (mod 12)
//   - 三合(trine): palaceIdx + 4 and palaceIdx + 8 (mod 12)
//
// 12 palaces × verify correct 3 harmony + opposite palace
// ═══════════════════════════════════════════════════════════════════════

// SanfangSizhengExpected holds the expected 三方四正 for one palace.
type SanfangSizhengExpected struct {
	PalaceName    string
	OppositeName  string // 对宫
	Trine1Name    string // 三合1
	Trine2Name    string // 三合2
}

// TestComputeSanfangSizheng_AllPalaces verifies all 12 palaces.
func TestComputeSanfangSizheng_AllPalaces(t *testing.T) {
	// 12 palaces in ZIWEI_PALACE_NAMES order: 命宫,兄弟,夫妻,子女,财帛,疾厄,迁移,交友,事业,田宅,福德,父母
	//
	// 三合 pattern:
	//   命宫(0)  ↔  迁移(4)  ↔  财帛(8)  ↔  (back to 命宫)
	//   兄弟(1)  ↔  交友(5)  ↔  疾厄(9)  ↔  (back to 兄弟)
	//   夫妻(2)  ↔  事业(6)  ↔  田宅(10) ↔  (back to 夫妻)
	//   子女(3)  ↔  福德(7)  ↔  父母(11) ↔  (back to 子女)
	//
	// 对宫:
	//   命宫(0) ↔ 迁移(6)
	//   兄弟(1) ↔ 交友(7)
	//   夫妻(2) ↔ 事业(8)
	//   子女(3) ↔ 福德(9)
	//   财帛(4) ↔ 父母(10)
	//   疾厄(5) ↔ 田宅(11)
	//   迁移(6) ↔ 命宫(0)
	//   ...

	expected := []SanfangSizhengExpected{
		// Format: self → opposite, trine1, trine2
		// Formula: opposite=(i+6)%12, trine1=(i+4)%12, trine2=(i+8)%12
		{PalaceName: "命宫", OppositeName: "迁移", Trine1Name: "财帛", Trine2Name: "事业"},
		{PalaceName: "兄弟", OppositeName: "交友", Trine1Name: "疾厄", Trine2Name: "田宅"},
		{PalaceName: "夫妻", OppositeName: "事业", Trine1Name: "迁移", Trine2Name: "福德"},
		{PalaceName: "子女", OppositeName: "田宅", Trine1Name: "交友", Trine2Name: "父母"},
		{PalaceName: "财帛", OppositeName: "福德", Trine1Name: "事业", Trine2Name: "命宫"},
		{PalaceName: "疾厄", OppositeName: "父母", Trine1Name: "田宅", Trine2Name: "兄弟"},
		{PalaceName: "迁移", OppositeName: "命宫", Trine1Name: "福德", Trine2Name: "夫妻"},
		{PalaceName: "交友", OppositeName: "兄弟", Trine1Name: "父母", Trine2Name: "子女"},
		{PalaceName: "事业", OppositeName: "夫妻", Trine1Name: "命宫", Trine2Name: "财帛"},
		{PalaceName: "田宅", OppositeName: "子女", Trine1Name: "兄弟", Trine2Name: "疾厄"},
		{PalaceName: "福德", OppositeName: "财帛", Trine1Name: "夫妻", Trine2Name: "迁移"},
		{PalaceName: "父母", OppositeName: "疾厄", Trine1Name: "子女", Trine2Name: "交友"},
	}

	for _, exp := range expected {
		// Find palace index
		palaceIdx := -1
		for i, name := range ZIWEI_PALACE_NAMES {
			if name == exp.PalaceName {
				palaceIdx = i
				break
			}
		}
		if palaceIdx < 0 {
			t.Fatalf("Palace %q not found in ZIWEI_PALACE_NAMES", exp.PalaceName)
		}

		t.Run(exp.PalaceName, func(t *testing.T) {
			// Test pure index computation
			sf := ComputeSanfangSizheng(palaceIdx)
			// sf[0] = opposite, sf[1] = trine1, sf[2] = trine2

			oppositeName := ZIWEI_PALACE_NAMES[sf[0]]
			trine1Name := ZIWEI_PALACE_NAMES[sf[1]]
			trine2Name := ZIWEI_PALACE_NAMES[sf[2]]

			// Verify opposite
			if oppositeName != exp.OppositeName {
				t.Errorf("对宫 = %s(%d), 期望 %s(%d)",
					oppositeName, sf[0], exp.OppositeName,
					findPalaceIndexInNames(exp.OppositeName))
			}

			// Verify trine palaces: trine1 and trine2 should be the two palaces at +4 and +8
			expectedTrine1 := (palaceIdx + 4) % 12
			expectedTrine2 := (palaceIdx + 8) % 12

			// Trine1 should be +4
			if sf[1] != expectedTrine1 {
				t.Errorf("三合1 = %s(%d), 期望 %s(%d)",
					trine1Name, sf[1],
					ZIWEI_PALACE_NAMES[expectedTrine1], expectedTrine1)
			}
			// Trine2 should be +8
			if sf[2] != expectedTrine2 {
				t.Errorf("三合2 = %s(%d), 期望 %s(%d)",
					trine2Name, sf[2],
					ZIWEI_PALACE_NAMES[expectedTrine2], expectedTrine2)
			}

			// Verify that all 4 indices (self + opposite + 2 trine) cover all unique positions
			// when grouped by earthly branch triplicity
			all := []int{palaceIdx, sf[0], sf[1], sf[2]}
			seen := make(map[int]bool)
			for _, idx := range all {
				seen[idx] = true
			}
			if len(seen) != 4 {
				t.Errorf("三方四正索引不唯一: %v", all)
			}

			// Verify classic rule: 三方 = 4 steps apart each
			// The difference between self and each trine should be 4 mod 12
			diff1 := fixIndex(sf[1] - palaceIdx)
			diff2 := fixIndex(sf[2] - palaceIdx)
			if diff1 != 4 && diff1 != 8 {
				t.Errorf("三合1距离 = %d, 期望 4 或 8", diff1)
			}
			if diff2 != 4 && diff2 != 8 {
				t.Errorf("三合2距离 = %d, 期望 4 或 8", diff2)
			}
			if diff1+diff2 != 12 {
				t.Errorf("两三合距离和 = %d, 期望 12", diff1+diff2)
			}

			// Verify opposite is exactly 6 steps away
			oppDiff := fixIndex(sf[0] - palaceIdx)
			if oppDiff != 6 {
				t.Errorf("对宫距离 = %d, 期望 6", oppDiff)
			}
		})
	}
}

// TestSanfangSizheng_Formula verifies the math formula directly.
func TestSanfangSizheng_Formula(t *testing.T) {
	for palaceIdx := 0; palaceIdx < 12; palaceIdx++ {
		t.Run(ZIWEI_PALACE_NAMES[palaceIdx], func(t *testing.T) {
			// Formula: opposite = (i+6)%12, trine1 = (i+4)%12, trine2 = (i+8)%12
			wantOpposite := (palaceIdx + 6) % 12
			wantTrine1 := (palaceIdx + 4) % 12
			wantTrine2 := (palaceIdx + 8) % 12

			sf := ComputeSanfangSizheng(palaceIdx)

			if sf[0] != wantOpposite {
				t.Errorf("对宫索引 = %d, 公式期望 %d (i+6 mod12)", sf[0], wantOpposite)
			}
			if sf[1] != wantTrine1 {
				t.Errorf("三合1索引 = %d, 公式期望 %d (i+4 mod12)", sf[1], wantTrine1)
			}
			if sf[2] != wantTrine2 {
				t.Errorf("三合2索引 = %d, 公式期望 %d (i+8 mod12)", sf[2], wantTrine2)
			}
		})
	}
}

// TestSanfangSizheng_ChartIntegration verifies the chart has SanfangSizheng populated.
func TestSanfangSizheng_ChartIntegration(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Verify chart.SanfangSizheng is populated for all 12 palaces
	for i := 0; i < 12; i++ {
		sf := chart.SanfangSizheng[i]
		if sf.Opposite == "" {
			t.Errorf("Palace[%d](%s): Opposite is empty", i, ZIWEI_PALACE_NAMES[i])
		}
		if sf.Trine1 == "" {
			t.Errorf("Palace[%d](%s): Trine1 is empty", i, ZIWEI_PALACE_NAMES[i])
		}
		if sf.Trine2 == "" {
			t.Errorf("Palace[%d](%s): Trine2 is empty", i, ZIWEI_PALACE_NAMES[i])
		}

		// Verify names match palace names
		expected := ComputeSanfangSizheng(i)
		if sf.Opposite != ZIWEI_PALACE_NAMES[expected[0]] {
			t.Errorf("Palace[%d]: Opposite = %s, 期望 %s",
				i, sf.Opposite, ZIWEI_PALACE_NAMES[expected[0]])
		}
		if sf.Trine1 != ZIWEI_PALACE_NAMES[expected[1]] {
			t.Errorf("Palace[%d]: Trine1 = %s, 期望 %s",
				i, sf.Trine1, ZIWEI_PALACE_NAMES[expected[1]])
		}
		if sf.Trine2 != ZIWEI_PALACE_NAMES[expected[2]] {
			t.Errorf("Palace[%d]: Trine2 = %s, 期望 %s",
				i, sf.Trine2, ZIWEI_PALACE_NAMES[expected[2]])
		}
	}
}

// TestGetPalaceSanfang verifies GetPalaceSanfang returns correct results.
func TestGetPalaceSanfang(t *testing.T) {
	for i := 0; i < 12; i++ {
		t.Run(ZIWEI_PALACE_NAMES[i], func(t *testing.T) {
			result := GetPalaceSanfang(i)
			expected := ComputeSanfangSizheng(i)

			if result.Opposite != ZIWEI_PALACE_NAMES[expected[0]] {
				t.Errorf("Opposite = %s, 期望 %s", result.Opposite, ZIWEI_PALACE_NAMES[expected[0]])
			}
			if result.Trine1 != ZIWEI_PALACE_NAMES[expected[1]] {
				t.Errorf("Trine1 = %s, 期望 %s", result.Trine1, ZIWEI_PALACE_NAMES[expected[1]])
			}
			if result.Trine2 != ZIWEI_PALACE_NAMES[expected[2]] {
				t.Errorf("Trine2 = %s, 期望 %s", result.Trine2, ZIWEI_PALACE_NAMES[expected[2]])
			}
		})
	}
}

// TestSanfangSizheng_Symmetry verifies the symmetry property:
// Palace A's opposite = Palace B means Palace B's opposite = Palace A.
// Trine relationship is also symmetric (if A trines B, then B trines A).
func TestSanfangSizheng_Symmetry(t *testing.T) {
	for i := 0; i < 12; i++ {
		sf := ComputeSanfangSizheng(i)

		// Opposite symmetry
		oppositeSF := ComputeSanfangSizheng(sf[0])
		if oppositeSF[0] != i {
			t.Errorf("%s的对宫是%s, 但%s的对宫是%s",
				ZIWEI_PALACE_NAMES[i], ZIWEI_PALACE_NAMES[sf[0]],
				ZIWEI_PALACE_NAMES[sf[0]], ZIWEI_PALACE_NAMES[oppositeSF[0]])
		}

		// Trine symmetry: if j is in i's trine, then i is in j's trine
		for _, triIdx := range []int{sf[1], sf[2]} {
			triSF := ComputeSanfangSizheng(triIdx)
			if triSF[1] != i && triSF[2] != i {
				t.Errorf("%s的三合是%s, 但%s的三合不含%s",
					ZIWEI_PALACE_NAMES[i], ZIWEI_PALACE_NAMES[triIdx],
					ZIWEI_PALACE_NAMES[triIdx], ZIWEI_PALACE_NAMES[i])
			}
		}
	}
}

// Helper: find palace index by name
func findPalaceIndexInNames(name string) int {
	for i, n := range ZIWEI_PALACE_NAMES {
		if n == name {
			return i
		}
	}
	return -1
}
