package service

// 五行生克与喜忌公共工具层。
// 主气表的真理源是 tyme.MAIN（参考 bazi.go:380-391），此处缓存为查表以避免对 tyme 的强依赖；
// 如果 tyme 主气定义发生变化，需要同步更新本表。

// ZhiMainQi 地支主气（本气）映射。
var ZhiMainQi = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木",
	"辰": "土", "巳": "火", "午": "火", "未": "土",
	"申": "金", "酉": "金", "戌": "土", "亥": "水",
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

// favorHuaQi 返回化气格的喜用：生扶化神（印 + 比劫）。
func favorHuaQi(huaQi string) []string {
	return []string{shengWo(huaQi), huaQi}
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
