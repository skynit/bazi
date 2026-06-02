package bazi

// 五行生克与喜忌公共工具层。
// 主气表的真理源是 tyme.MAIN（参考 bazi.go:380-391），此处缓存为查表以避免对 tyme 的强依赖；
// 如果 tyme 主气定义发生变化，需要同步更新本表。

// ZhiMainQi 地支主气（本气）映射。
var ZhiMainQi = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木",
	"辰": "土", "巳": "火", "午": "火", "未": "土",
	"申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// ZhiAllElements 地支藏干所含五行集合（含本气、中气、余气）。
var ZhiAllElements = map[string][]string{
	"子": {"水"},
	"丑": {"土", "水", "金"},
	"寅": {"木", "火", "土"},
	"卯": {"木"},
	"辰": {"土", "木", "水"},
	"巳": {"火", "金", "土"},
	"午": {"火", "土"},
	"未": {"土", "火", "木"},
	"申": {"金", "水", "土"},
	"酉": {"金"},
	"戌": {"土", "金", "火"},
	"亥": {"水", "木"},
}

// mainQi 返回地支本气五行；未知地支返回空字符串。
func mainQi(zhi string) string {
	return ZhiMainQi[zhi]
}

// shengWo 返回生我者（印星五行）。
func shengWo(elem string) string {
	return map[string]string{"木": "水", "火": "木", "土": "火", "金": "土", "水": "金"}[elem]
}

// woSheng 返回我生者（食伤五行）。
func woSheng(elem string) string {
	return map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}[elem]
}

// keWuXing 返回我克者（财五行）。等价于 woKe。
func keWuXing(elem string) string {
	return map[string]string{"木": "土", "火": "金", "土": "水", "金": "木", "水": "火"}[elem]
}

// woKe 是 keWuXing 的语义化别名。
func woKe(elem string) string { return keWuXing(elem) }

// keWo 返回克我者（官杀五行）。
func keWo(elem string) string {
	return map[string]string{"木": "金", "火": "水", "土": "木", "金": "火", "水": "土"}[elem]
}

// ShengWo 是 shengWo 的导出版本，供其他包调用。
func ShengWo(elem string) string { return shengWo(elem) }

// WoSheng 是 woSheng 的导出版本，供其他包调用。
func WoSheng(elem string) string { return woSheng(elem) }

// WoKe 是 woKe 的导出版本，供其他包调用。
func WoKe(elem string) string { return woKe(elem) }

// KeWo 是 keWo 的导出版本，供其他包调用。
func KeWo(elem string) string { return keWo(elem) }

// favorHuaQi 返回化气格的喜用：生扶化神及化神所生（印 + 比劫 + 食伤）。
func favorHuaQi(huaQi string) []string {
	return []string{shengWo(huaQi), huaQi, woSheng(huaQi)}
}

// tongGuan 返回两神成像格的通关五行（生二者之一的五行）。
func tongGuan(a, b string) string {
	if shengWo(b) == a {
		return b
	}
	if shengWo(a) == b {
		return a
	}
	if keWuXing(a) == b {
		return woSheng(a)
	}
	if keWuXing(b) == a {
		return woSheng(b)
	}
	return ""
}

// totalScore 累加五行得分。
func totalScore(scores map[string]int) int {
	sum := 0
	for _, s := range scores {
		sum += s
	}
	return sum
}

// absInt 返回 int 绝对值。
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// inStrings 检查目标是否出现在列表中。
func inStrings(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

// computeFavorByDayElem 基于日主五行返回喜忌列表。
// congRuo=false 时：身旺型喜忌反转由调用方按 verdict 决定；本函数返回 (生扶, 克泄耗)。
// congRuo=true  时：从弱格 like=克泄耗, dislike=生扶。
//
// 调用约定：
//   like, dislike := computeFavorByDayElem(dayElem, true)
//   // 从弱：like 给克泄耗，dislike 给生扶
//
//   like, dislike := computeFavorByDayElem(dayElem, false)
//   // like = [印, 比劫], dislike = [官杀, 食伤, 财]
//   // 身旺时调用方对调即可
func computeFavorByDayElem(dayElem string, congRuo bool) (like, dislike []string) {
	support := []string{shengWo(dayElem), dayElem}                  // 印 + 比劫
	restrict := []string{keWo(dayElem), woSheng(dayElem), woKe(dayElem)} // 官杀 + 食伤 + 财
	if congRuo {
		return restrict, support
	}
	return support, restrict
}
