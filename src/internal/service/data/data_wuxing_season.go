package data

// WuxingSeasonEntry 描述某个五行在特定季节的表现特征。
// 数据来源：《三命通会》第65-69章 "论五行时地分野吉凶"。
type WuxingSeasonEntry struct {
	Element string // 五行
	Season  string // 季节：春/夏/秋/冬 或 正月/二月 等
	State   string // 旺衰状态
	Climate string // 晴雨喜忌
	Favor   string // 喜用
	Taboo   string // 忌
	Judgment string // 综合断语
	Regions  string // 地域差异（徐扬/荆梁豫/冀雍/兖青）
}

// WuxingSeasonKnowledge 五行四时论知识库。
var WuxingSeasonKnowledge = map[string][]WuxingSeasonEntry{
	// ── 金 ──
	"金": {
		{
			Element: "金", Season: "春", State: "绝地（寅月绝，卯辰月胎养）",
			Climate: "晴天为吉，阴雨则气机阻滞", Favor: "宜土生扶（金气将绝，得土方有生发）", Taboo: "忌阴雨、忌水泄金气",
			Judgment: "正月见木为财但过旺，技艺学识扬名；见火有再婚之象；火土俱全富贵非凡。二三月见土生意绵延富寿。",
			Regions: "荆梁豫人逢木多富，徐扬人干支土多则贵显。",
		},
		{
			Element: "金", Season: "夏", State: "极柔弱（火势炽盛）",
			Climate: "晴天烈日烤熔，雨天水润", Favor: "均宜见土", Taboo: "忌火（火炎金烁贫夭）",
			Judgment: "见土可出将入相、翰林显贵。见火贫夭者多。见木为财，荆梁豫人富贵。逢水孤寒，与火土同行则富贵安康。",
			Regions: "荆梁豫人见木富贵。行运喜土木火，最忌秋金肃杀。",
		},
		{
			Element: "金", Season: "秋", State: "当令得势（申酉月），性格刚强",
			Climate: "顽金无火不能成器", Favor: "需火锻造（玉带金鱼之贵）", Taboo: "见土则彩光被埋（虽富不达，多孤独）",
			Judgment: "七八月：需火制锋芒。见水金清水澈文翰清贵。水火皆无主夭。见木为财徐扬人富贵。九月：金气稍退，夜生遇火奇贵。",
			Regions: "徐扬人见木贵。兖青人逢土富贵多。行运喜东南忌西北。",
		},
		{
			Element: "金", Season: "冬", State: "伏藏畏寒，无生发之意",
			Climate: "晴明则金清水秀，雨雪则水冷金寒", Favor: "必须得火融寒", Taboo: "忌见水（愈寒）、忌寒土",
			Judgment: "十月见火主伤残。子丑月喜火温暖。徐扬人无火亦喜逢土得火方贵。遇水春冬贫寒，干支有火土者福寿。",
			Regions: "徐扬人火土俱全福寿。冀雍人有土无火孤贫。行运喜东南忌西北。",
		},
	},

	// ── 木 ──
	"木": {
		{
			Element: "木", Season: "春", State: "旺盛（阳气生发）",
			Climate: "", Favor: "喜火土", Taboo: "忌金重克木",
			Judgment: "春木旺盛，得火土配合富贵。见金则木被克伤。",
			Regions: "",
		},
		{
			Element: "木", Season: "夏", State: "旺盛但易燥",
			Climate: "", Favor: "喜水滋润", Taboo: "忌火过旺（木焚）",
			Judgment: "夏木需要水来滋润，水火既济为贵。无水则枯。",
			Regions: "",
		},
		{
			Element: "木", Season: "秋", State: "凋零肃杀",
			Climate: "", Favor: "喜金修剪成材", Taboo: "忌金过重伤木",
			Judgment: "秋木凋零，有金修剪成材，金过重则伤。",
			Regions: "",
		},
		{
			Element: "木", Season: "冬", State: "寒气旺盛",
			Climate: "", Favor: "喜火温暖", Taboo: "忌水过旺木浮",
			Judgment: "冬木需要火来温暖调候，水火配合为贵。",
			Regions: "",
		},
	},

	// ── 水 ──
	"水": {
		{
			Element: "水", Season: "春", State: "看似至弱（病寅死卯墓辰）",
			Climate: "晴则春水溶溶，雨则汪洋泛滥", Favor: "正月见火冰融冻释；二三月见土有堤防", Taboo: "忌土克（贫寒）、忌无火木冷",
			Judgment: "正月见火富贵雍容，得金相助。二三月浩然无边际，见土堤防则富贵。见火水火相刑灾讼不免。",
			Regions: "徐扬人正月得金最佳。兖青人二月忌木泄气。",
		},
		{
			Element: "水", Season: "夏", State: "失令干涸",
			Climate: "忌晴喜雨", Favor: "喜土止而不流（福气深厚）", Taboo: "忌与火同行（火盛土燥水涸）",
			Judgment: "初夏：得土止流福厚，不宜与火同行。五六月：喜土同行，阴雨生时主富贵文章。见火有涸水之嫌轻重不一。",
			Regions: "徐扬人干支无金水者疾夭。荆梁豫人逢金吉。",
		},
		{
			Element: "水", Season: "秋", State: "清秀澄澈（正能滋实万物）",
			Climate: "晴则清澈无瑕，雨则潦水浑浊", Favor: "金为母曜子母相生（文章清贵）", Taboo: "不宜与火同行（徐扬火多者贫夭）",
			Judgment: "七月滋实万物，不宜与火同行。金母适当其时文章清贵。八月九月遇令星福寿难量。金火同垣功名显赫。",
			Regions: "徐扬豫人干支土多亦困滞终身。行运利西北东南失宜。",
		},
		{
			Element: "水", Season: "冬", State: "司令当旺寒气严凝",
			Climate: "雨则冰凝，晴则冻释", Favor: "冬三月俱喜火以温之", Taboo: "忌无火（水冷金寒反贫薄）",
			Judgment: "喜火富贵无虞。见金子母相生。丑月生者贵显。见木水寒木冻俱无生意贫夭。",
			Regions: "徐扬人金紫玉堂之贵。冀雍人虽相生反贫薄。行运喜南方东方次之。",
		},
	},

	// ── 火 ──
	"火": {
		{
			Element: "火", Season: "春", State: "初生（温和但无旺势）",
			Climate: "", Favor: "喜木生火、喜水济火", Taboo: "忌金水过重",
			Judgment: "春火温和，需木生扶。水火相济为贵。",
			Regions: "",
		},
		{
			Element: "火", Season: "夏", State: "得令极旺",
			Climate: "", Favor: "喜水调候（水火既济）", Taboo: "忌土过旺晦火无光",
			Judgment: "夏火极旺，必须有水来济。无水则火炎土燥贫夭。",
			Regions: "",
		},
		{
			Element: "火", Season: "秋", State: "渐衰收藏",
			Climate: "", Favor: "喜木扶助", Taboo: "忌金水过旺",
			Judgment: "秋火渐衰，有木扶助则能持续发光。",
			Regions: "",
		},
		{
			Element: "火", Season: "冬", State: "极弱畏寒",
			Climate: "", Favor: "喜木火相助", Taboo: "忌水过旺（水火不交）",
			Judgment: "冬火微弱，必须有木火相助方能调候。",
			Regions: "",
		},
	},

	// ── 土 ──
	"土": {
		{
			Element: "土", Season: "春", State: "被木所克",
			Climate: "", Favor: "喜火生土", Taboo: "忌木过旺克土",
			Judgment: "春土被木克，需火来化解木克之力。",
			Regions: "",
		},
		{
			Element: "土", Season: "夏", State: "厚实燥热",
			Climate: "", Favor: "喜水润土", Taboo: "忌火过旺燥土",
			Judgment: "夏土燥热，得水润则万物生长。",
			Regions: "",
		},
		{
			Element: "土", Season: "秋", State: "金泄土气",
			Climate: "", Favor: "喜火补土", Taboo: "忌金过旺泄土",
			Judgment: "秋土被金泄气，需火来补土之力。",
			Regions: "",
		},
		{
			Element: "土", Season: "冬", State: "寒冷凝固",
			Climate: "", Favor: "喜火温暖", Taboo: "忌水寒土冻",
			Judgment: "冬土寒冷，得火温暖方能生养。",
			Regions: "",
		},
	},
}

// SeasonFromMonth maps birth month (1-12) to season key.
// 按中国传统节气/地支划分与公历对应：寅卯辰≈公历2-4月=春, 巳午未≈公历5-7月=夏,
// 申酉戌≈公历8-10月=秋, 亥子丑≈公历11-1月=冬.
func SeasonFromMonth(m int) string {
	switch m {
	case 2, 3, 4:
		return "春"
	case 5, 6, 7:
		return "夏"
	case 8, 9, 10:
		return "秋"
	default: // 11, 12, 1
		return "冬"
	}
}
