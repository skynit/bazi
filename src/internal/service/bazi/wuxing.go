package bazi

// 五行生克与喜忌公共工具层。
// 主气表的真理源是 tyme.MAIN（参考 bazi.go:380-391），此处缓存为查表以避免对 tyme 的强依赖；
// 如果 tyme 主气定义发生变化，需要同步更新本表。

// zhiMainQi 地支主气（本气）映射。
var zhiMainQi = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木",
	"辰": "土", "巳": "火", "午": "火", "未": "土",
	"申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// zhiAllElements 地支藏干所含五行集合（含本气、中气、余气）。
var zhiAllElements = map[string][]string{
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
	return zhiMainQi[zhi]
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
