package data

// JiaZiEntry 描述一个六十甲子组合的性质、喜忌和神煞标记。
// 数据来源：《三命通会》第7章 "释六十甲子性质吉凶"。
type JiaZiEntry struct {
	GanZhi   string   // 干支组合，如 "甲子"
	Nayin    string   // 纳音名，如 "海中金"
	Nature   string   // 性质描述，如 "宝物"
	Likes    string   // 喜用，如 "金木旺地"
	ShenSha   []string // 关联神煞标记
	Warnings []string // 注意事项
}

// JiaZiKnowledge 六十甲子性质吉凶知识库。
// 索引为干支组合字符串（如 "甲子"）。
var JiaZiKnowledge = map[string]JiaZiEntry{
	// ── 甲子 · 海中金 ──
	"甲子": {GanZhi: "甲子", Nayin: "海中金", Nature: "宝物", Likes: "喜金木旺地", ShenSha: []string{"进神", "福星"}, Warnings: []string{"平头", "悬针", "破字"}},
	"乙丑": {GanZhi: "乙丑", Nayin: "海中金", Nature: "顽矿", Likes: "喜火及南方日时", ShenSha: []string{"福星", "华盖", "正印"}, Warnings: nil},

	// ── 丙寅 · 炉中火 ──
	"丙寅": {GanZhi: "丙寅", Nayin: "炉中火", Nature: "炉炭", Likes: "喜冬及木", ShenSha: []string{"福星", "禄刑"}, Warnings: []string{"平头", "聋哑"}},
	"丁卯": {GanZhi: "丁卯", Nayin: "炉中火", Nature: "炉烟", Likes: "喜巽地及秋冬", ShenSha: nil, Warnings: []string{"平头", "截路", "悬针"}},

	// ── 戊辰 · 大林木 ──
	"戊辰": {GanZhi: "戊辰", Nayin: "大林木", Nature: "山林不材之木", Likes: "喜水", ShenSha: []string{"禄库", "华盖", "水禄马库"}, Warnings: []string{"棒杖", "伏神", "平头"}},
	"己巳": {GanZhi: "己巳", Nayin: "大林木", Nature: "山头花草", Likes: "喜春及秋", ShenSha: []string{"禄库", "八吉"}, Warnings: []string{"阙字", "曲脚"}},

	// ── 庚午 · 路旁土 ──
	"庚午": {GanZhi: "庚午", Nayin: "路旁土", Nature: "路旁干土", Likes: "喜水及春", ShenSha: []string{"福星", "官贵"}, Warnings: []string{"截路", "棒杖", "悬针"}},
	"辛未": {GanZhi: "辛未", Nayin: "路旁土", Nature: "含万宝待秋成", Likes: "喜秋及火", ShenSha: []string{"华盖"}, Warnings: []string{"悬针", "破字"}},

	// ── 壬申 · 剑锋金 ──
	"壬申": {GanZhi: "壬申", Nayin: "剑锋金", Nature: "戈戟", Likes: "大喜子午卯酉", ShenSha: []string{"平头", "大败"}, Warnings: []string{"妨害", "聋哑", "破字", "悬针"}},
	"癸酉": {GanZhi: "癸酉", Nayin: "剑锋金", Nature: "椎凿", Likes: "喜木及寅卯", ShenSha: []string{"伏神"}, Warnings: []string{"破字", "聋哑"}},

	// ── 甲戌 · 山头火 ──
	"甲戌": {GanZhi: "甲戌", Nayin: "山头火", Nature: "火所宿处", Likes: "喜春及夏", ShenSha: []string{"正印", "华盖"}, Warnings: []string{"平头", "悬针", "破字", "棒杖"}},
	"乙亥": {GanZhi: "乙亥", Nayin: "山头火", Nature: "火之热气", Likes: "喜土及夏", ShenSha: []string{"天德"}, Warnings: []string{"曲脚"}},

	// ── 丙子 · 涧下水 ──
	"丙子": {GanZhi: "丙子", Nayin: "涧下水", Nature: "江湖", Likes: "喜木及土", ShenSha: []string{"福星", "官贵"}, Warnings: []string{"平头", "聋哑", "交神", "飞刃"}},
	"丁丑": {GanZhi: "丁丑", Nayin: "涧下水", Nature: "不流清澈处", Likes: "喜金及夏", ShenSha: []string{"华盖", "进神"}, Warnings: []string{"平头", "飞刃", "阙字"}},

	// ── 戊寅 · 城头土 ──
	"戊寅": {GanZhi: "戊寅", Nayin: "城头土", Nature: "堤阜城郭", Likes: "喜木及火", ShenSha: []string{"伏神", "俸杖"}, Warnings: []string{"聋哑"}},
	"己卯": {GanZhi: "己卯", Nayin: "城头土", Nature: "破堤败城", Likes: "喜申酉及火", ShenSha: []string{"进神"}, Warnings: []string{"短夭", "九丑", "阙字", "曲脚", "悬针"}},

	// ── 庚辰 · 白蜡金 ──
	"庚辰": {GanZhi: "庚辰", Nayin: "白蜡金", Nature: "锡蜡", Likes: "喜秋及微木", ShenSha: []string{"华盖", "大败"}, Warnings: []string{"棒杖", "平头"}},
	"辛巳": {GanZhi: "辛巳", Nayin: "白蜡金", Nature: "杂沙石之金", Likes: "喜火及秋", ShenSha: []string{"天德", "福星", "官贵"}, Warnings: []string{"截路", "大败", "悬针", "曲脚"}},

	// ── 壬午 · 杨柳木 ──
	"壬午": {GanZhi: "壬午", Nayin: "杨柳木", Nature: "杨柳干节", Likes: "喜春夏", ShenSha: []string{"官贵"}, Warnings: []string{"九丑", "飞刃", "平头", "聋哑", "悬针"}},
	"癸未": {GanZhi: "癸未", Nayin: "杨柳木", Nature: "杨柳根", Likes: "喜冬及水亦宜春", ShenSha: []string{"正印", "华盖"}, Warnings: []string{"短夭", "伏神", "飞刃", "破字"}},

	// ── 甲申 · 泉中水 ──
	"甲申": {GanZhi: "甲申", Nayin: "泉中水", Nature: "甘井", Likes: "喜春及夏", ShenSha: nil, Warnings: []string{"破禄马", "截路", "平头", "破字", "悬针"}},
	"乙酉": {GanZhi: "乙酉", Nayin: "泉中水", Nature: "阴壑水", Likes: "喜东方及南", ShenSha: nil, Warnings: []string{"破禄", "短夭", "九丑", "曲脚", "破字", "聋哑"}},

	// ── 丙戌 · 屋上土 ──
	"丙戌": {GanZhi: "丙戌", Nayin: "屋上土", Nature: "堆阜", Likes: "喜春夏及水", ShenSha: []string{"天德", "华盖"}, Warnings: []string{"平头", "聋哑"}},
	"丁亥": {GanZhi: "丁亥", Nayin: "屋上土", Nature: "平原", Likes: "喜火及木", ShenSha: []string{"天乙", "福星", "官贵", "德合"}, Warnings: []string{"平头"}},

	// ── 戊子 · 霹雳火 ──
	"戊子": {GanZhi: "戊子", Nayin: "霹雳火", Nature: "雷", Likes: "喜水及春夏得土而神", ShenSha: []string{"伏神"}, Warnings: []string{"短夭", "九丑", "杖刑", "飞刃"}},
	"己丑": {GanZhi: "己丑", Nayin: "霹雳火", Nature: "电", Likes: "喜水及春夏得地而晦", ShenSha: []string{"华盖", "大败"}, Warnings: []string{"飞刃", "曲脚", "阙字"}},

	// ── 庚寅 · 松柏木 ──
	"庚寅": {GanZhi: "庚寅", Nayin: "松柏木", Nature: "松柏干节", Likes: "喜秋冬", ShenSha: nil, Warnings: []string{"破禄马", "相刑", "杖刑", "聋哑"}},
	"辛卯": {GanZhi: "辛卯", Nayin: "松柏木", Nature: "松柏之根", Likes: "喜水土及春", ShenSha: nil, Warnings: []string{"破禄", "交神", "九丑", "悬针"}},

	// ── 壬辰 · 长流水 ──
	"壬辰": {GanZhi: "壬辰", Nayin: "长流水", Nature: "龙水", Likes: "喜雷电及春夏", ShenSha: []string{"正印", "天德", "水禄马库"}, Warnings: []string{"退神", "平头", "聋哑"}},
	"癸巳": {GanZhi: "癸巳", Nayin: "长流水", Nature: "不息入海", Likes: "喜亥子乃变化", ShenSha: []string{"天乙", "官贵", "德合", "伏马"}, Warnings: []string{"破字", "曲脚"}},

	// ── 甲午 · 砂中金 ──
	"甲午": {GanZhi: "甲午", Nayin: "砂中金", Nature: "百炼精金", Likes: "喜水木土", ShenSha: []string{"进神", "德合"}, Warnings: []string{"平头", "破字", "悬针"}},
	"乙未": {GanZhi: "乙未", Nayin: "砂中金", Nature: "炉炭余金", Likes: "喜大火及土", ShenSha: []string{"华盖"}, Warnings: []string{"截路", "曲脚", "破字"}},

	// ── 丙申 · 山下火 ──
	"丙申": {GanZhi: "丙申", Nayin: "山下火", Nature: "白茅野烧", Likes: "喜秋冬及木", ShenSha: nil, Warnings: []string{"平头", "聋哑", "大败", "破字", "悬针"}},
	"丁酉": {GanZhi: "丁酉", Nayin: "山下火", Nature: "无形之火", Likes: "喜辰戌丑未", ShenSha: []string{"天乙", "喜神"}, Warnings: []string{"平头", "破字", "聋哑", "大败"}},

	// ── 戊戌 · 平地木 ──
	"戊戌": {GanZhi: "戊戌", Nayin: "平地木", Nature: "蒿艾枯者", Likes: "喜火及春夏", ShenSha: []string{"华盖", "大败", "八专"}, Warnings: []string{"杖刑", "截路"}},
	"己亥": {GanZhi: "己亥", Nayin: "平地木", Nature: "蒿艾茅", Likes: "喜水及春夏", ShenSha: nil, Warnings: []string{"阙字", "曲脚"}},

	// ── 庚子 · 壁上土 ──
	"庚子": {GanZhi: "庚子", Nayin: "壁上土", Nature: "屋宇", Likes: "喜木及金", ShenSha: []string{"木德合"}, Warnings: []string{"杖刑"}},
	"辛丑": {GanZhi: "辛丑", Nayin: "壁上土", Nature: "坟墓", Likes: "喜木及火与春", ShenSha: []string{"华盖"}, Warnings: []string{"悬针", "阙字"}},

	// ── 壬寅 · 金箔金 ──
	"壬寅": {GanZhi: "壬寅", Nayin: "金箔金", Nature: "华饰", Likes: "喜木及微火", ShenSha: nil, Warnings: []string{"截路", "平头", "聋哑"}},
	"癸卯": {GanZhi: "癸卯", Nayin: "金箔金", Nature: "环钮钤铎", Likes: "喜盛火及秋", ShenSha: []string{"贵人"}, Warnings: []string{"破字", "悬针"}},

	// ── 甲辰 · 覆灯火 ──
	"甲辰": {GanZhi: "甲辰", Nayin: "覆灯火", Nature: "灯", Likes: "喜夜及水，恶昼", ShenSha: []string{"华盖", "大败"}, Warnings: []string{"平头", "破字", "悬针"}},
	"乙巳": {GanZhi: "乙巳", Nayin: "覆灯火", Nature: "灯光", Likes: "喜申酉及秋", ShenSha: []string{"正禄马", "大败"}, Warnings: []string{"曲脚", "阙字"}},

	// ── 丙午 · 天河水 ──
	"丙午": {GanZhi: "丙午", Nayin: "天河水", Nature: "月轮", Likes: "喜夜及秋水旺", ShenSha: []string{"喜神", "羊刃", "交神"}, Warnings: []string{"平头", "聋哑", "悬针"}},
	"丁未": {GanZhi: "丁未", Nayin: "天河水", Nature: "火光", Likes: "同上", ShenSha: []string{"华盖", "羊刃", "退神", "八专"}, Warnings: []string{"平头", "破字"}},

	// ── 戊申 · 大驿土 ──
	"戊申": {GanZhi: "戊申", Nayin: "大驿土", Nature: "秋间田地", Likes: "喜申酉及火", ShenSha: []string{"福星", "伏马"}, Warnings: []string{"杖刑", "破字", "悬针"}},
	"己酉": {GanZhi: "己酉", Nayin: "大驿土", Nature: "秋间禾稼", Likes: "喜申酉及冬", ShenSha: []string{"退神", "截路"}, Warnings: []string{"九丑", "阙字", "曲脚", "破字", "聋哑"}},

	// ── 庚戌 · 钗钏金 ──
	"庚戌": {GanZhi: "庚戌", Nayin: "钗钏金", Nature: "刃剑之余", Likes: "喜微火及木", ShenSha: []string{"华盖"}, Warnings: []string{"杖刑"}},
	"辛亥": {GanZhi: "辛亥", Nayin: "钗钏金", Nature: "钟鼎实物", Likes: "喜木火及土", ShenSha: []string{"正禄马"}, Warnings: []string{"悬针"}},

	// ── 壬子 · 桑柘木 ──
	"壬子": {GanZhi: "壬子", Nayin: "桑柘木", Nature: "伤水多之木", Likes: "喜火土及夏", ShenSha: []string{"羊刃", "九丑"}, Warnings: []string{"平头", "聋哑"}},
	"癸丑": {GanZhi: "癸丑", Nayin: "桑柘木", Nature: "伤水少之木", Likes: "喜金水及秋", ShenSha: []string{"华盖", "福星", "八专"}, Warnings: []string{"破字", "阙字", "羊刃"}},

	// ── 甲寅 · 大溪水 ──
	"甲寅": {GanZhi: "甲寅", Nayin: "大溪水", Nature: "雨", Likes: "喜夏及火", ShenSha: []string{"正禄马", "福神", "八专"}, Warnings: []string{"平头", "破字", "悬针", "聋哑"}},
	"乙卯": {GanZhi: "乙卯", Nayin: "大溪水", Nature: "露", Likes: "喜水及火", ShenSha: []string{"建禄", "喜神", "八专"}, Warnings: []string{"九刃", "曲脚", "悬针"}},

	// ── 丙辰 · 砂中土 ──
	"丙辰": {GanZhi: "丙辰", Nayin: "砂中土", Nature: "堤岸", Likes: "喜金及木", ShenSha: []string{"禄库", "正印", "华盖"}, Warnings: []string{"截路", "平头", "聋哑"}},
	"丁巳": {GanZhi: "丁巳", Nayin: "砂中土", Nature: "沮洳", Likes: "喜火及西北", ShenSha: []string{"禄库"}, Warnings: []string{"平头", "阙字", "曲脚"}},

	// ── 戊午 · 天上火 ──
	"戊午": {GanZhi: "戊午", Nayin: "天上火", Nature: "日轮", Likes: "冬暖夏畏", ShenSha: []string{"伏神", "羊刃", "九丑"}, Warnings: []string{"棒杖", "悬针"}},
	"己未": {GanZhi: "己未", Nayin: "天上火", Nature: "日光", Likes: "忌夜亦畏四者", ShenSha: []string{"福星", "华盖", "羊刃"}, Warnings: []string{"阙字", "曲脚", "破字"}},

	// ── 庚申 · 石榴木 ──
	"庚申": {GanZhi: "庚申", Nayin: "石榴木", Nature: "榴花", Likes: "喜夏不宜秋冬", ShenSha: []string{"建禄马", "八专"}, Warnings: []string{"杖刑", "破字", "悬针"}},
	"辛酉": {GanZhi: "辛酉", Nayin: "石榴木", Nature: "榴子", Likes: "喜秋及夏", ShenSha: []string{"建禄", "交神", "九丑", "八专"}, Warnings: []string{"悬针", "聋哑"}},

	// ── 壬戌 · 大海水 ──
	"壬戌": {GanZhi: "壬戌", Nayin: "大海水", Nature: "海", Likes: "喜春夏及木", ShenSha: []string{"华盖", "退神"}, Warnings: []string{"平头", "聋哑", "杖刑"}},
	"癸亥": {GanZhi: "癸亥", Nayin: "大海水", Nature: "百川", Likes: "喜金土火", ShenSha: []string{"伏马", "大败"}, Warnings: []string{"破字", "截路"}},
}
