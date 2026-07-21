package bazi_test

import (
	"testing"

	"bazi/internal/service/data"
)

// 60-cycle 甲子 in order: gan index = i % 10, zhi index = i % 12
var gans = [10]string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var zhis = [12]string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// Expected NaYin names for all 60 甲子 combinations, ordered by the 60-cycle (渊海子平).
var expectedNayin60 = []string{
	// 1-12
	"海中金", "海中金", // 甲子, 乙丑
	"炉中火", "炉中火", // 丙寅, 丁卯
	"大林木", "大林木", // 戊辰, 己巳
	"路旁土", "路旁土", // 庚午, 辛未
	"剑锋金", "剑锋金", // 壬申, 癸酉
	"山头火", "山头火", // 甲戌, 乙亥
	// 13-24
	"涧下水", "涧下水", // 丙子, 丁丑
	"城头土", "城头土", // 戊寅, 己卯
	"白蜡金", "白蜡金", // 庚辰, 辛巳
	"杨柳木", "杨柳木", // 壬午, 癸未
	"井泉水", "井泉水", // 甲申, 乙酉
	"屋上土", "屋上土", // 丙戌, 丁亥
	// 25-36
	"霹雳火", "霹雳火", // 戊子, 己丑
	"松柏木", "松柏木", // 庚寅, 辛卯
	"长流水", "长流水", // 壬辰, 癸巳
	"砂中金", "砂中金", // 甲午, 乙未
	"山下火", "山下火", // 丙申, 丁酉
	"平地木", "平地木", // 戊戌, 己亥
	// 37-48
	"壁上土", "壁上土", // 庚子, 辛丑
	"金箔金", "金箔金", // 壬寅, 癸卯
	"覆灯火", "覆灯火", // 甲辰, 乙巳
	"天河水", "天河水", // 丙午, 丁未
	"大驿土", "大驿土", // 戊申, 己酉
	"钗钏金", "钗钏金", // 庚戌, 辛亥
	// 49-60
	"桑柘木", "桑柘木", // 壬子, 癸丑
	"大溪水", "大溪水", // 甲寅, 乙卯
	"砂中土", "砂中土", // 丙辰, 丁巳
	"天上火", "天上火", // 戊午, 己未
	"石榴木", "石榴木", // 庚申, 辛酉
	"大海水", "大海水", // 壬戌, 癸亥
}

// Expected element for each of the 30 NaYin types.
// 金: 海中金, 金箔金, 白蜡金, 砂中金, 剑锋金, 钗钏金
// 木: 大林木, 杨柳木, 松柏木, 桑柘木, 石榴木, 平地木
// 水: 涧下水, 大溪水, 长流水, 天河水, 井泉水, 大海水
// 火: 炉中火, 霹雳火, 天上火, 山下火, 山头火, 覆灯火
// 土: 路旁土, 壁上土, 城头土, 砂中土, 大驿土, 屋上土
var expectedElementMap = map[string]string{
	// 金 (6)
	"海中金": "金",
	"金箔金": "金",
	"白蜡金": "金",
	"砂中金": "金",
	"剑锋金": "金",
	"钗钏金": "金",
	// 木 (6)
	"大林木": "木",
	"杨柳木": "木",
	"松柏木": "木",
	"桑柘木": "木",
	"石榴木": "木",
	"平地木": "木",
	// 水 (6)
	"涧下水": "水",
	"大溪水": "水",
	"长流水": "水",
	"天河水": "水",
	"井泉水": "水",
	"大海水": "水",
	// 火 (6)
	"炉中火": "火",
	"霹雳火": "火",
	"天上火": "火",
	"山下火": "火",
	"山头火": "火",
	"覆灯火": "火",
	// 土 (6)
	"路旁土": "土",
	"壁上土": "土",
	"城头土": "土",
	"砂中土": "土",
	"大驿土": "土",
	"屋上土": "土",
}

// TestNayin60Cycle verifies all 60 甲子 combinations produce correct NaYin names
// by directly checking data.Nayin[ganIdx][zhiIdx] against the classic 渊海子平 table.
func TestNayin60Cycle(t *testing.T) {
	for i, expected := range expectedNayin60 {
		gan := gans[i%10]
		zhi := zhis[i%12]
		ganIdx := data.GanIndex(gan)
		zhiIdx := data.ZhiIndex(zhi)
		actual := data.Nayin[ganIdx][zhiIdx]

		if actual != expected {
			t.Fatalf("60甲子 cycle position %d: %s%s → expected %q, got %q",
				i+1, gan, zhi, expected, actual)
		}
	}
}

// TestNayinMapExistence verifies NaYinMap has exactly 30 entries
// and that it covers all expected NaYin names.
func TestNayinMapExistence(t *testing.T) {
	if len(data.NaYinMap) != 30 {
		t.Fatalf("NaYinMap has %d entries, expected exactly 30", len(data.NaYinMap))
	}

	for name := range expectedElementMap {
		entry, ok := data.NaYinMap[name]
		if !ok {
			t.Fatalf("NaYinMap missing entry for %q", name)
		}
		if entry.Name != name {
			t.Fatalf("NaYinMap[%q].Name = %q, expected %q", name, entry.Name, name)
		}
		if entry.Element != expectedElementMap[name] {
			t.Fatalf("NaYinMap[%q].Element = %q, expected %q", name, entry.Element, expectedElementMap[name])
		}
		if len(entry.StemBranches) != 2 {
			t.Fatalf("NaYinMap[%q].StemBranches has %d entries, expected exactly 2",
				name, len(entry.StemBranches))
		}
	}

	// Ensure no extra entries exist in NaYinMap beyond the 30 expected types
	for name := range data.NaYinMap {
		if _, ok := expectedElementMap[name]; !ok {
			t.Fatalf("NaYinMap contains unexpected entry %q", name)
		}
	}
}
