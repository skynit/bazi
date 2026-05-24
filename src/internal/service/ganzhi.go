package service

import "bazi/internal/model"

// ============================================================
// 天干新增数据（已有：Gans, GanElement, ganHe 在 data_gans.go / shensha.go）
// ============================================================

// GanYinYang 天干阴阳
var GanYinYang = map[string]string{
	"甲": "阳", "乙": "阴", "丙": "阳", "丁": "阴",
	"戊": "阳", "己": "阴", "庚": "阳", "辛": "阴",
	"壬": "阳", "癸": "阴",
}

// GanHeHua 天干五合化气
var GanHeHua = map[string]string{
	"甲己": "土", "己甲": "土",
	"乙庚": "金", "庚乙": "金",
	"丙辛": "水", "辛丙": "水",
	"丁壬": "木", "壬丁": "木",
	"戊癸": "火", "癸戊": "火",
}

// GanKe 天干相克（我克者）
var GanKe = map[string]string{
	"甲": "戊", "乙": "己", "丙": "庚", "丁": "辛",
	"戊": "壬", "己": "癸", "庚": "甲", "辛": "乙",
	"壬": "丙", "癸": "丁",
}

// GanSheng 天干相生（我生者）
var GanSheng = map[string]string{
	"甲": "丙", "乙": "丁", "丙": "戊", "丁": "己",
	"戊": "庚", "己": "辛", "庚": "壬", "辛": "癸",
	"壬": "甲", "癸": "乙",
}

// ============================================================
// 地支新增数据（已有：Zhis, ZhiElement, zhiLiuHe, zhiLiuChong 在 data_gans.go / shensha.go）
// ============================================================

// ZhiYinYang 地支阴阳
var ZhiYinYang = map[string]string{
	"子": "阳", "丑": "阴", "寅": "阳", "卯": "阴", "辰": "阳", "巳": "阴",
	"午": "阳", "未": "阴", "申": "阳", "酉": "阴", "戌": "阳", "亥": "阴",
}

// ZhiWuXing 地支五行
var ZhiWuXing = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木", "辰": "土", "巳": "火",
	"午": "火", "未": "土", "申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// ZhiSanHe 地支三合局
var ZhiSanHe = map[string][]string{
	"申": {"申", "子", "辰"}, "子": {"申", "子", "辰"}, "辰": {"申", "子", "辰"},
	"亥": {"亥", "卯", "未"}, "卯": {"亥", "卯", "未"}, "未": {"亥", "卯", "未"},
	"寅": {"寅", "午", "戌"}, "午": {"寅", "午", "戌"}, "戌": {"寅", "午", "戌"},
	"巳": {"巳", "酉", "丑"}, "酉": {"巳", "酉", "丑"}, "丑": {"巳", "酉", "丑"},
}

// ZhiSanHui 地支三会
var ZhiSanHui = map[string][]string{
	"寅": {"寅", "卯", "辰"}, "卯": {"寅", "卯", "辰"}, "辰": {"寅", "卯", "辰"},
	"巳": {"巳", "午", "未"}, "午": {"巳", "午", "未"}, "未": {"巳", "午", "未"},
	"申": {"申", "酉", "戌"}, "酉": {"申", "酉", "戌"}, "戌": {"申", "酉", "戌"},
	"亥": {"亥", "子", "丑"}, "子": {"亥", "子", "丑"}, "丑": {"亥", "子", "丑"},
}

// ZhiLiuHai 地支六害
var ZhiLiuHai = map[string]string{
	"子": "未", "未": "子", "丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅", "卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申", "酉": "戌", "戌": "酉",
}

// ZhiXing 地支相刑
var ZhiXing = map[string]string{
	"子": "卯", "卯": "子",
	"寅": "巳", "巳": "申", "申": "寅",
	"丑": "戌", "戌": "未", "未": "丑",
	"辰": "辰", "午": "午", "酉": "酉", "亥": "亥",
}

// ============================================================
// 天干地支综合分析
// ============================================================

// GanRelation 天干关系
type GanRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
	Detail  string `json:"detail"`
}

// ZhiRelation 地支关系
type ZhiRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
	Detail  string `json:"detail"`
}

// GanZhiAnalysis 干支综合分析结果
type GanZhiAnalysis struct {
	GanRelations []GanRelation `json:"gan_relations"`
	ZhiRelations []ZhiRelation `json:"zhi_relations"`
}

const (
	labelYear  = "年柱"
	labelMonth = "月柱"
	labelDay   = "日柱"
	labelHour  = "时柱"
)

// CalcGanZhiAnalysis 计算四柱之间的天干和地支关系
func CalcGanZhiAnalysis(year, month, day, hour model.Pillar) GanZhiAnalysis {
	var result GanZhiAnalysis

	pillars := []struct {
		label string
		gan   string
		zhi   string
	}{
		{labelYear, year.Gan, year.Zhi},
		{labelMonth, month.Gan, month.Zhi},
		{labelDay, day.Gan, day.Zhi},
		{labelHour, hour.Gan, hour.Zhi},
	}

	// 6 对两两组合
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			// 天干关系
			result.GanRelations = append(result.GanRelations,
				calcGanPairRelations(pillars[i].label, pillars[i].gan,
					pillars[j].label, pillars[j].gan)...)

			// 地支关系
			result.ZhiRelations = append(result.ZhiRelations,
				calcZhiPairRelations(pillars[i].label, pillars[i].zhi,
					pillars[j].label, pillars[j].zhi)...)
		}
	}

	return result
}

func calcGanPairRelations(label1, gan1, label2, gan2 string) []GanRelation {
	var rels []GanRelation

	// 五合
	if ganHe[gan1] == gan2 {
		pair := gan1 + gan2
		hua := GanHeHua[pair]
		rels = append(rels, GanRelation{label1, label2, "五合", gan1 + gan2 + "合" + hua})
	}

	// 相克
	if GanKe[gan1] == gan2 {
		rels = append(rels, GanRelation{label1, label2, "相克", gan1 + "克" + gan2})
	} else if GanKe[gan2] == gan1 {
		rels = append(rels, GanRelation{label1, label2, "相克", gan2 + "克" + gan1})
	}

	// 相生
	if GanSheng[gan1] == gan2 {
		rels = append(rels, GanRelation{label1, label2, "相生", gan1 + "生" + gan2})
	} else if GanSheng[gan2] == gan1 {
		rels = append(rels, GanRelation{label1, label2, "相生", gan2 + "生" + gan1})
	}

	return rels
}

func calcZhiPairRelations(label1, zhi1, label2, zhi2 string) []ZhiRelation {
	var rels []ZhiRelation

	// 六冲
	if zhiLiuChong[zhi1] == zhi2 {
		rels = append(rels, ZhiRelation{label1, label2, "六冲", zhi1 + zhi2 + "冲"})
	}

	// 六合
	if zhiLiuHe[zhi1] == zhi2 {
		rels = append(rels, ZhiRelation{label1, label2, "六合", zhi1 + zhi2 + "合"})
	}

	// 六害
	if ZhiLiuHai[zhi1] == zhi2 {
		rels = append(rels, ZhiRelation{label1, label2, "六害", zhi1 + zhi2 + "害"})
	}

	// 相刑
	if xingTarget, ok := ZhiXing[zhi1]; ok && xingTarget == zhi2 {
		var xingName string
		switch {
		case (zhi1 == "子" && zhi2 == "卯") || (zhi1 == "卯" && zhi2 == "子"):
			xingName = "无礼之刑"
		case (zhi1 == "寅" && zhi2 == "巳") || (zhi1 == "巳" && zhi2 == "申") || (zhi1 == "申" && zhi2 == "寅"):
			xingName = "恃势之刑"
		case (zhi1 == "丑" && zhi2 == "戌") || (zhi1 == "戌" && zhi2 == "未") || (zhi1 == "未" && zhi2 == "丑"):
			xingName = "无恩之刑"
		default:
			xingName = "自刑"
		}
		rels = append(rels, ZhiRelation{label1, label2, "相刑", zhi1 + zhi2 + "刑（" + xingName + "）"})
	}

	// 三会（同在三会方）
	if group := ZhiSanHui[zhi1]; group != nil {
		for _, z := range group {
			if z == zhi2 {
				rels = append(rels, ZhiRelation{label1, label2, "三会", zhi1 + zhi2 + "会"})
				break
			}
		}
	}

	return rels
}
