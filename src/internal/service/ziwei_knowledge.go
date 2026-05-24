package service

// ──────────────────── ZiWei Knowledge Enrichment Layer ────────────────────
//
// All constants and algorithms are derived from classical ZiWei Dou Shu texts
// and the reference implementation at https://github.com/Renhuai123/ziwei-doushu.
// No AI/LLM generation. No external runtime dependencies.
//
// This layer enriches the output of ziwei-zenith with local computation:
//   - 四化飞星 chain analysis
//   - 40+ local pattern detectors
//   - 三方四正 computation
//   - 合盘 (heming) analysis
//   - Template-based palace readings
// ─────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"strings"

	"github.com/kaecer68/ziwei-zenith/pkg/basis"
)

// ──────────────────── Constants ────────────────────

// SI_HUA_TABLE maps each of the 10 heavenly stems to its four transformations.
// Order: [禄, 权, 科, 忌]
var SI_HUA_TABLE = map[int][]string{
	0: {"廉贞", "破军", "武曲", "太阳"}, // 甲
	1: {"天机", "天梁", "紫微", "太阴"}, // 乙
	2: {"天梁", "紫微", "天机", "太阳"}, // 丙
	3: {"太阴", "武曲", "天同", "天机"}, // 丁
	4: {"天同", "太阳", "武曲", "廉贞"}, // 戊
	5: {"太阳", "天同", "天梁", "巨门"}, // 己
	6: {"武曲", "天相", "太阳", "天机"}, // 庚
	7: {"天相", "天同", "武曲", "紫微"}, // 辛
	8: {"天梁", "天机", "天同", "太阳"}, // 壬
	9: {"紫微", "天梁", "天机", "太阴"}, // 癸
}

// STAR_BRIGHTNESS maps star → brightness level → description.
var STAR_BRIGHTNESS = map[string]map[string]string{
	"紫微": {"庙": "紫微帝星入庙，气势非凡，具领导才能与威严气度。", "旺": "紫微旺势，气度雍容，具统御之才。", "得": "紫微得地，根基稳固，具备一定领导能力。", "利": "紫微利位，气势稍减，宜借助团队力量。", "平": "紫微平势，帝王之气不显，需靠自身努力。", "陷": "紫微陷落，有志难伸，多怀才不遇之感。", "不": "紫微失辉，气运低迷，行事受阻。"},
	"天府": {"庙": "天府入庙，库藏丰厚，主一生财禄充足。", "旺": "天府旺势，财库充盈，善理财富。", "得": "天府得地，衣食无忧，有一定理财观念。", "利": "天府利位，财运平稳，宜保守理财。", "平": "天府平势，库藏之力不显，财来财去。", "陷": "天府陷落，库门大开，钱财易散难聚。", "不": "天府失库，财运不济，收入不稳定。"},
	"天机": {"庙": "天机入庙，智慧超群，心思敏捷，善谋划。", "旺": "天机旺势，聪明机智，学习能力强。", "得": "天机得地，头脑清晰，具备分析策划才能。", "利": "天机利位，思维敏锐但格局有限。", "平": "天机平势，智慧表现平稳。", "陷": "天机陷落，思路混乱，容易钻牛角尖。", "不": "天机失智，思绪紊乱，反應迟钝。"},
	"太阳": {"庙": "太阳入庙，光芒万丈，性格光明磊落，热情开朗。", "旺": "太阳旺势，热忱积极，乐于助人。", "得": "太阳得地，光明正直，有一定人缘。", "利": "太阳利位，光芒稍敛，性格较内敛。", "平": "太阳平势，行事低调，不喜张扬。", "陷": "太阳陷落，光芒被遮蔽，性格变消沉或急躁。", "不": "太阳失辉，运势低迷，难发挥优势。"},
	"武曲": {"庙": "武曲入庙，刚毅果决，理财能力出众。", "旺": "武曲旺势，做事干脆利落，工作效率极高。", "得": "武曲得地，做事踏实可靠，理财稳健保守。", "利": "武曲利位，执行力尚可，但格局稍小。", "平": "武曲平势，刚毅之气减弱，做事较犹豫。", "陷": "武曲陷落，刚强之气化为固执，易与人冲突。", "不": "武曲失势，决断力不足，做事缺乏魄力。"},
	"天同": {"庙": "天同入庙，性情温和，知足常乐，天生具有福气。", "旺": "天同旺势，心态平和，生活安逸。", "得": "天同得地，福气不错，生活衣食无忧。", "利": "天同利位，福缘尚可，需加执行力和主动性。", "平": "天同平势，福气平常，凡事靠自己努力。", "陷": "天同陷落，性情变消极懒散，缺乏进取心。", "不": "天同失福，运势较差，容易感到疲惫无力。"},
	"廉贞": {"庙": "廉贞入庙，刚柔并济，执法力强，处事圆融。", "旺": "廉贞旺势，能力出众，做事有条理。", "得": "廉贞得地，做事认真负责，有一定管理能力。", "利": "廉贞利位，能力尚可但范围有限。", "平": "廉贞平势，表現平稳，无特别突出之处。", "陷": "廉贞陷落，性格变偏执或多疑，易与人摩擦。", "不": "廉贞失势，自律能力下降，容易放纵或消极。"},
	"贪狼": {"庙": "贪狼入庙，多才多艺，魅力四射，社交能力极强。", "旺": "贪狼旺势，交际手腕灵活，善经营人脉。", "得": "贪狼得地，有一定才艺和社交能力。", "利": "贪狼利位，欲望心重，需学会知足。", "平": "贪狼平势，才华表现一般，社交能力尚可。", "陷": "贪狼陷落，容易沉迷于物质享受或不良嗜好。", "不": "贪狼失守，贪欲过盛，易招惹麻烦。"},
	"巨门": {"庙": "巨门入庙，口才犀利，分析能力极强，善辩论谈判。", "旺": "巨门旺势，表达能力强，善发现问题本质。", "得": "巨门得地，思维清晰，有较好分析和表达能力。", "平": "巨门平势，口才和分析能力尚可。", "陷": "巨门陷落，言语容易招惹是非，须谨言慎行。"},
	"天相": {"庙": "天相入庙，为人正直，善辅助他人，是优秀幕僚。", "旺": "天相旺势，处事稳当，有条不紊。", "得": "天相得地，做事认真负责，能胜任辅助性工作。", "平": "天相平势，辅助能力一般。", "陷": "天相陷落，服务精神减弱，容易感到委屈不平衡。"},
	"天梁": {"庙": "天梁入庙，心地善良，乐于助人，有长者之风。", "旺": "天梁旺势，性格稳重，有责任感，能为他人遮风挡雨。", "得": "天梁得地，心地不错，有一定福报。", "利": "天梁利位，福报稍弱，付出与收获难平衡。", "平": "天梁平势，福报平常，宜多积德行善。", "陷": "天梁陷落，容易受累于他人之事。", "不": "天梁失荫，福德有亏，孤寒难免。"},
	"七杀": {"庙": "七杀入庙，胆识过人，勇于开拓，具有极强竞争力。", "旺": "七杀旺势，行动力强，做事果断不拖沓。", "得": "七杀得地，有一定魄力，能够应对挑战。", "利": "七杀利位，竞争压力大，宜顺势而行。", "平": "七杀平势，将星之气不显，魄力稍欠。", "陷": "七杀陷落，容易冲动误事，需克制急躁。", "不": "七杀失威，胆识不足，畏缩难伸。"},
	"破军": {"庙": "破军入庙，敢作敢为，具有强大破坏力和创造力。", "旺": "破军旺势，行动力强，不畏困难，勇于挑战现状。", "得": "破军得地，有一定开拓精神。", "利": "破军利位，变革有阻，需借势而行。", "平": "破军平势，变革动力不足。", "陷": "破军陷落，容易因冲动而导致损失。", "不": "破军失力，破旧无力，损耗过重。"},
	"太阴": {"庙": "太阴入庙，温柔贤淑，心思细腻，具有艺术和审美天赋。", "旺": "太阴旺势，性情温和，做事细心周到。", "得": "太阴得地，性格柔和，有一定审美能力。", "利": "太阴利位，情感细腻易生忧郁，宜乐观行事。", "平": "太阴平势，能力表现平稳。", "陷": "太阴陷落，情绪容易波动，判断力下降。", "不": "太阴失辉，运势晦暗，谋事难成。"},
}

// STAR_BRIGHTNESS_AUX provides brightness descriptions for auxiliary (辅助) and unlucky (煞星) stars.
// Covers 6 lucky stars (左辅, 右弼, 文昌, 文曲, 天魁, 天钺) and 6 unlucky stars (擎羊, 陀罗, 火星, 铃星, 地空, 地劫).
var STAR_BRIGHTNESS_AUX = map[string]map[string]string{
	"左辅": {"庙": "左辅入庙，辅佐有力，贵人运强。", "旺": "左辅旺势，行事顺遂，得人相助。", "得": "左辅得地，有一定助力，人缘不错。", "利": "左辅利位，助力有限，需自身努力。", "平": "左辅平势，贵人运平常。", "陷": "左辅陷落，助力薄弱，行事多阻。", "不": "左辅失助，孤立无援。"},
	"右弼": {"庙": "右弼入庙，辅弼得力，机遇佳。", "旺": "右弼旺势，机缘深厚，行事顺畅。", "得": "右弼得地，机遇不错，有所收获。", "利": "右弼利位，机遇稍纵即逝，宜把握。", "平": "右弼平势，机遇平常。", "陷": "右弼陷落，机遇难逢，机缘有亏。", "不": "右弼失机，运势低迷。"},
	"文昌": {"庙": "文昌入庙，文采飞扬，学业出众。", "旺": "文昌旺势，聪明好学，成绩优异。", "得": "文昌得地，文学天赋，有一定学业成就。", "利": "文昌利位，学业需努力才能维持。", "平": "文昌平势，学业表现平常。", "陷": "文昌陷落，学业有阻，不利考试。", "不": "文昌失辉，学业难成。"},
	"文曲": {"庙": "文曲入庙，才艺超群，表达能力强。", "旺": "文曲旺势，多才多艺，文艺方面表现突出。", "得": "文曲得地，有一定文艺才华。", "利": "文曲利位，才艺展示需借助时机。", "平": "文曲平势，才艺表现平常。", "陷": "文曲陷落，才艺难发挥，学术受阻。", "不": "文曲失彩，才华不显。"},
	"天魁": {"庙": "天魁入庙，贵人鼎力，机遇从天而降。", "旺": "天魁旺势，贵人相助，机会连连。", "得": "天魁得地，有一定机遇，得贵人提携。", "利": "天魁利位，机遇需主动把握。", "平": "天魁平势，机遇平平。", "陷": "天魁陷落，机遇难寻，贵人无力。", "不": "天魁失机，运势不佳。"},
	"天钺": {"庙": "天钺入庙，机遇独特，创新能力强。", "旺": "天钺旺势，敢于突破，机会佳。", "得": "天钺得地，有独特机遇，可有所作为。", "利": "天钺利位，机遇与风险并存。", "平": "天钺平势，机遇平常。", "陷": "天钺陷落，机遇变风险，行事需谨慎。", "不": "天钺失势，机会流失。"},
	"擎羊": {"庙": "擎羊入庙，冲劲十足，敢于冒险。", "旺": "擎羊旺势，行动力强，果断敢冲。", "得": "擎羊得地，有一定冲劲，可有所作为。", "利": "擎羊利位，冲动有风险，宜稳。", "平": "擎羊平势，冲劲平常。", "陷": "擎羊陷落，冲动误事，易有血光之灾。", "不": "擎羊失制，灾祸易生。"},
	"陀罗": {"庙": "陀罗入庙，恒心毅力，耐久力强。", "旺": "陀罗旺势，意志坚定，不轻言放弃。", "得": "陀罗得地，有一定耐力，做事能坚持。", "利": "陀罗利位，耐力有极限，需适度调节。", "平": "陀罗平势，耐力平常。", "陷": "陀罗陷落，意志动摇，易生拖延。", "不": "陀罗失持，毅力不足。"},
	"火星": {"庙": "火星入庙，爆发力强，行动迅猛。", "旺": "火星旺势，冲劲十足，敢闯敢拼。", "得": "火星得地，有一定爆发力，能成就大事。", "利": "火星利位，爆发力有时限，宜把握时机。", "平": "火星平势，爆发力平常。", "陷": "火星陷落，爆发力受阻，行事急躁易败。", "不": "火星失焰，动力不足。"},
	"铃星": {"庙": "铃星入庙，细腻而有力，深藏不露。", "旺": "铃星旺势，爆发力强且持久，後劲十足。", "得": "铃星得地，有一定耐力与爆发力。", "利": "铃星利位，後劲需防后继乏力。", "平": "铃星平势，耐力平常。", "陷": "铃星陷落，耐力不足，行事易中途而废。", "不": "铃星失鸣，动力难续。"},
	"地空": {"庙": "地空入庙，空灵智慧，精神力量强。", "旺": "地空旺势，思维独特，有创新精神。", "得": "地空得地，有一定精神力，行事独特。", "利": "地空利位，精神力有波动，宜平衡。", "平": "地空平势，创意平常。", "陷": "地空陷落，精神空乏，行事难落实。", "不": "地空失空，空想误事。"},
	"地劫": {"庙": "地劫入庙，破坏力强，具改革精神。", "旺": "地劫旺势，破旧立新，行动力强。", "得": "地劫得地，有一定魄力，可打破常规。", "利": "地劫利位，破坏需有方向，忌盲目。", "平": "地劫平势，魄力平常。", "陷": "地劫陷落，破坏力失控，损人损己。", "不": "地劫失势，破局无力。"},
}

// LUCUN_TABLE maps year stem index to the branch where 禄存 is located.
var LUCUN_TABLE = map[int]int{
	0: 2,  // 甲 → 寅
	1: 3,  // 乙 → 卯
	2: 4,  // 丙 → 辰
	3: 5,  // 丁 → 巳
	4: 6,  // 戊 → 午
	5: 7,  // 己 → 未
	6: 8,  // 庚 → 申
	7: 9,  // 辛 → 酉
	8: 10, // 壬 → 戌
	9: 11, // 癸 → 亥
}

// TIANMA_TABLE maps the three-combination group to the branch where 天马 is.
var TIANMA_TABLE = map[string]int{
	"寅午戌": 10, // 马在申
	"申子辰": 4,  // 马在寅
	"巳酉丑": 7,  // 马在亥
	"亥卯未": 1,  // 马在巳
}

// ──────────────────── Palace order (clockwise from 命宫) ────────────────────
var PALACE_NAMES = []string{
	"命宮", "兄弟宮", "夫妻宮", "子女宮",
	"財帛宮", "疾厄宮", "遷移宮", "僕役宮",
	"官祿宮", "田宅宮", "福德宮", "父母宮",
}

// ──────────────────── Sihua Chain Analysis ────────────────────

// SihuaChainResult holds the result of a full Sihua chain analysis.
type SihuaChainResult struct {
	HuaLu           []SihuaChainItem `json:"hua_lu"`
	HuaQuan        []SihuaChainItem `json:"hua_quan"`
	HuaKe          []SihuaChainItem `json:"hua_ke"`
	HuaJi          []SihuaChainItem `json:"hua_ji"`
	TotalChainDepth int              `json:"total_chain_depth"`
	KeyMutagens    []string         `json:"key_mutagens"`
}

// SelfMutagenResult holds a self-mutagen occurrence.
type SelfMutagenResult struct {
	Palace  string `json:"palace"`
	Star    string `json:"star"`
	HuaType string `json:"hua_type"`
	Effect  string `json:"effect"`
	IsSelf  bool   `json:"is_self"`
}

// SihuaChainItem represents one star in a sihua chain.
type SihuaChainItem struct {
	FromStar     string `json:"from_star"`
	ToPalace     string `json:"to_palace"`
	FromPalace   string `json:"from_palace"`
	Effect       string `json:"effect"`
	ChainDepth   int    `json:"chain_depth"`
	StarAffinity int    `json:"star_affinity"`
	MutagenType  string `json:"mutagen_type"`   // "self" | "regular"
	FlyDirection string `json:"fly_direction"`  // "same_palace" | "to_palace"
	IsSelfMutagen bool  `json:"is_self_mutagen"`
}

// AnalyzeSihuaChain performs full sihua chain analysis on a chart.
// It finds all incoming palaces (not just where stars land) and builds the chain.
func AnalyzeSihuaChain(chart *ZiWeiChart) *SihuaChainResult {
	if chart == nil || chart.engineChart == nil {
		return nil
	}

	ec := chart.engineChart
	yearStem := int(ec.YearPillar.Stem)
	huaStars, ok := SI_HUA_TABLE[yearStem]
	if !ok {
		return nil
	}

	huaLuStar := huaStars[0]
	huaQuanStar := huaStars[1]
	huaKeStar := huaStars[2]
	huaJiStar := huaStars[3]

	// Build star -> branch map
	starToBranch := make(map[string]int)
	branchOfPalace := make(map[int]int) // branch -> palace index
	for b := 0; b < 12; b++ {
		if p, ok := ec.Palaces[basis.Branch(b)]; ok {
			branchOfPalace[int(b)] = int(p)
		}
		for _, s := range ec.Stars[basis.Branch(b)] {
			starToBranch[s.String()] = int(b)
		}
	}

	result := &SihuaChainResult{}

	// Helper to build chain for a given transformation star
	buildChain := func(starName string, huaType string) []SihuaChainItem {
		var items []SihuaChainItem
		if starName == "" {
			return items
		}

		fromBranch, found := starToBranch[starName]
		if !found {
			return items
		}

		toPalaceIdx := branchOfPalace[fromBranch]
		if toPalaceIdx < 0 || toPalaceIdx >= 12 {
			return items
		}

		fromPalaceIdx := -1
		// Find where this star originally came from (its "home" palace)
		// For sihua, the star transforms in place; we look at the palace it's in
		fromPalaceIdx = toPalaceIdx

		effect := buildSihuaEffect(starName, huaType, PALACE_NAMES[toPalaceIdx])
		chainDepth := computeChainDepth(chart, toPalaceIdx, huaType)
		starAffinity := computeStarAffinity(chart, toPalaceIdx)
		mutagenType, flyDirection, isSelf := computeSelfMutagen(chart, starName, fromPalaceIdx, toPalaceIdx)

		items = append(items, SihuaChainItem{
			FromStar:     starName,
			ToPalace:     PALACE_NAMES[toPalaceIdx],
			FromPalace:   PALACE_NAMES[fromPalaceIdx],
			Effect:       effect,
			ChainDepth:   chainDepth,
			StarAffinity: starAffinity,
			MutagenType:  mutagenType,
			FlyDirection: flyDirection,
			IsSelfMutagen: isSelf,
		})

		// Also find if any other palace's transformed star also flies to same palace (incoming)
		findIncomingPalaces(chart, starName, huaType, toPalaceIdx, &items)

		return items
	}

	result.HuaLu = buildChain(huaLuStar, "化禄")
	result.HuaQuan = buildChain(huaQuanStar, "化权")
	result.HuaKe = buildChain(huaKeStar, "化科")
	result.HuaJi = buildChain(huaJiStar, "化忌")

	// Compute TotalChainDepth and KeyMutagens
	totalDepth := 0
	mutagenCounts := make(map[string]int)
	for _, items := range [][]SihuaChainItem{result.HuaLu, result.HuaQuan, result.HuaKe, result.HuaJi} {
		for _, item := range items {
			totalDepth += item.ChainDepth
			if item.IsSelfMutagen {
				mutagenCounts[item.FromStar]++
			}
		}
	}
	result.TotalChainDepth = totalDepth
	for star, count := range mutagenCounts {
		if count >= 2 {
			result.KeyMutagens = append(result.KeyMutagens, star)
		}
	}

	return result
}

// computeChainDepth counts how many sequential sihua triggers exist in the same palace.
func computeChainDepth(chart *ZiWeiChart, palaceIdx int, huaType string) int {
	if chart == nil {
		return 0
	}
	count := 0
	palaceName := PALACE_NAMES[palaceIdx]
	for _, p := range chart.Palaces {
		if p.Name == palaceName {
			for _, star := range p.FourHua {
				if strings.Contains(star, huaType) {
					count++
				}
			}
		}
	}
	return count
}

// computeStarAffinity counts auxiliary/support stars in the same palace.
func computeStarAffinity(chart *ZiWeiChart, palaceIdx int) int {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return 0
	}
	return len(chart.Palaces[palaceIdx].AuxStars)
}

// findIncomingPalaces traces all palaces that send a transformation to the given palace.
func findIncomingPalaces(chart *ZiWeiChart, starName string, huaType string, targetPalaceIdx int, items *[]SihuaChainItem) {
	if chart == nil {
		return
	}
	// In the reference sihua system, each star transforms only once (倪师天纪)
	// So incoming is just where this star currently resides
	// But we check if another star of same hua type also targets this palace
	for i, p := range chart.Palaces {
		if i == targetPalaceIdx {
			continue
		}
		for _, tStar := range p.FourHua {
			if strings.Contains(tStar, starName) && strings.Contains(tStar, huaType) {
				*items = append(*items, SihuaChainItem{
					FromStar:   starName,
					ToPalace:   PALACE_NAMES[targetPalaceIdx],
					FromPalace: p.Name,
					Effect:     buildSihuaEffect(starName, huaType, PALACE_NAMES[targetPalaceIdx]),
					ChainDepth: 0,
				})
			}
		}
	}
}

// buildSihuaEffect generates a rule-based effect description.
func buildSihuaEffect(star, huaType, palace string) string {
	effects := map[string]string{
		"命宮":  "直接影响个人运势与性格",
		"兄弟宮": "影响兄弟姐妹关系与助力",
		"夫妻宮": "影响婚姻感情与配偶关系",
		"子女宮": "影响子女缘分与下属关系",
		"財帛宮": "影响财运与金钱进出",
		"疾厄宮": "影响身体健康状况",
		"遷移宮": "影响外出运程与社会形象",
		"僕役宮": "影响朋友与部属关系",
		"官祿宮": "影响事业运程与工作成就",
		"田宅宮": "影响房产运程与家庭环境",
		"福德宮": "影响精神享受与内心世界",
		"父母宮": "影响父母缘分与长辈助力",
	}
	if desc, ok := effects[palace]; ok {
		return fmt.Sprintf("%s%s飞入%s，%s", star, huaType, palace, desc)
	}
	return fmt.Sprintf("%s%s飞入%s", star, huaType, palace)
}

// computeSelfMutagen checks if a transformed star stays in the same palace (self-mutagen)
// or flies to a different palace (regular sihua).
func computeSelfMutagen(chart *ZiWeiChart, fromStar string, fromPalaceIdx, toPalaceIdx int) (mutagenType string, flyDirection string, isSelf bool) {
	if fromPalaceIdx == toPalaceIdx {
		return "self", "same_palace", true
	}
	return "regular", "to_palace", false
}

// DetectSelfMutagens finds all self-mutagen occurrences in a chart.
// Algorithm: for each palace, for each mainStar, check if the same star appears
// as a transformed star (FourHua) in the SAME palace — that indicates self-mutagen.
// Regular sihua: a star transforms and FLIES to a DIFFERENT palace.
// Self-mutagen: a star transforms but STAYS in its own palace.
func DetectSelfMutagens(chart *ZiWeiChart) []SelfMutagenResult {
	if chart == nil {
		return nil
	}
	var results []SelfMutagenResult
	for _, palace := range chart.Palaces {
		for _, mainStar := range palace.MainStars {
			for _, transformed := range palace.FourHua {
				if strings.Contains(transformed, mainStar) {
					huaType := ""
					if strings.Contains(transformed, "化禄") {
						huaType = "化禄"
					} else if strings.Contains(transformed, "化权") {
						huaType = "化权"
					} else if strings.Contains(transformed, "化科") {
						huaType = "化科"
					} else if strings.Contains(transformed, "化忌") {
						huaType = "化忌"
					}
					if huaType != "" {
						results = append(results, SelfMutagenResult{
							Palace:  palace.Name,
							Star:    mainStar,
							HuaType: huaType,
							Effect:  buildSelfMutagenEffect(mainStar, huaType, palace.Name),
							IsSelf:  true,
						})
					}
				}
			}
		}
	}
	return results
}

func buildSelfMutagenEffect(star, huaType, palace string) string {
	effect := map[string]string{
		"化禄": "此星在本宫自化禄，财运与事业有天生加持，但需防贪心",
		"化权": "此星在本宫自化权，行事果断有魄力，但易刚愎自用",
		"化科": "此星在本宫自化科，智慧与名声有自然彰显，但宜防虚浮",
		"化忌": "此星在本宫自化忌，易有执着与障碍，需防自我中心",
	}
	if e, ok := effect[huaType]; ok {
		return fmt.Sprintf("%s%s，%s", star, huaType, e)
	}
	return fmt.Sprintf("%s%s", star, huaType)
}

// ──────────────────── Pattern Detection ────────────────────

// patternChecker is a function that checks for a specific pattern.
type patternChecker func(chart *ZiWeiChart) (bool, string)

// DetectLocalPatterns detects fortune patterns using local rules (not engine).
func DetectLocalPatterns(chart *ZiWeiChart) []string {
	if chart == nil {
		return nil
	}

	var detected []string

	for _, pc := range patternCheckers {
		if present, _ := pc.checker(chart); present {
			detected = append(detected, pc.name)
		}
	}

	return detected
}


// patternCheckers is a list of pattern checker functions with their names.
var patternCheckers = []struct {
	name    string
	checker patternChecker
}{
	{"紫府同宫", checkZiFuTongGong},
	{"紫破同宫", checkZiPoTongGong},
	{"紫相拱照", checkZiXiangGongZhao},
	{"杀破狼格", checkShaPoLang},
	{"机月同梁格", checkJiYueTongLiang},
	{"紫武廉府", checkZiWuLianFu},
	{"府相朝垣", checkFuXiangChaoYuan},
	{"日月拱照", checkRiYueGongZhao},
	{"日月反背", checkRiYueFanBei},
	{"禄马交驰", checkLuMaJiaoChi},
	{"天马拱命", checkTianMaGongMing},
	{"火贪格", checkHuoTanGe},
	{"铃贪格", checkLingTanGe},
	{"擎羊入命", checkQingYangRuMing},
	{"陀罗入命", checkTuoLuoRuMing},
	{"空宫", checkKongGong},
	{"桃花犯主", checkTaoHuaFanZhu},
	{"水木清华", checkShuiMuQingHua},
	{"土金相生", checkTuJinXiangSheng},
	{"日月并明", checkRiYueBingMing},
	{"极向离明", checkJiXiangLiMing},
	{"石中隐玉", checkShiZhongYinYu},
	{"文桂文华", checkWenGuiWenHua},
	{"天府守垣", checkTianFuShouYuan},
	{"寿星入庙", checkShouXingRuMiao},
	{"马头带剑", checkMaTouDaiJian},
	{"君子在野", checkJunZiZaiYe},
	{"巨日同宫", checkJuRiTongGong},
	{"科名会禄", checkKeMingHuiLu},
	{"财印夹马", checkCaiYinJiaMa},
	{"三奇嘉会", checkSanQiJiaHui},
	{"天乙同宫", checkTianYiTongGong},
	{"七杀朝斗", checkQiShaChaoDou},
	{"武贪格", checkWuTanGe},
	{"廉府双星", checkLianFuShuangXing},
	{"同梁双星", checkTongLiangShuangXing},
	{"日月夹命", checkRiYueJiaMing},
	{"辅弼夹印", checkFuBiJiaYin},
}


func starInPalace(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	for _, main := range chart.Palaces[palaceIdx].MainStars {
		for _, s := range starNames {
			if main == s {
				return true
			}
		}
	}
	return false
}

func starInPalaceByName(chart *ZiWeiChart, palaceName string, starNames []string) bool {
	if chart == nil {
		return false
	}
	for i, p := range chart.Palaces {
		if p.Name == palaceName {
			return starInPalace(chart, i, starNames)
		}
	}
	return false
}

func auxStarInPalace(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	for _, aux := range chart.Palaces[palaceIdx].AuxStars {
		for _, s := range starNames {
			if aux == s {
				return true
			}
		}
	}
	return false
}

func starInSamePalace(chart *ZiWeiChart, palaceIdx int, star1, star2 string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	p := chart.Palaces[palaceIdx]
	has1 := false
	has2 := false
	for _, s := range p.MainStars {
		if s == star1 {
			has1 = true
		}
		if s == star2 {
			has2 = true
		}
	}
	return has1 && has2
}

func hasBrightness(chart *ZiWeiChart, palaceIdx int, starName string, brightness []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	b := chart.Palaces[palaceIdx].Brightness
	if b == nil {
		return false
	}
	starBright, ok := b[starName]
	if !ok {
		return false
	}
	for _, br := range brightness {
		if starBright == br {
			return true
		}
	}
	return false
}

func hasTrine(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	trine1 := (palaceIdx + 4) % 12
	trine2 := (palaceIdx + 8) % 12
	return starInPalace(chart, trine1, starNames) || starInPalace(chart, trine2, starNames)
}

func hasOpposition(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	opposite := (palaceIdx + 6) % 12
	return starInPalace(chart, opposite, starNames)
}

// Pattern checkers

func checkZiFuTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := 0; i < 12; i++ {
		if starInPalace(chart, i, []string{"紫微", "天府"}) {
			return true, "紫府同宫"
		}
	}
	return false, ""
}

func checkZiPoTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInSamePalace(chart, i, "紫微", "破军") {
			return true, "紫破同宫"
		}
	}
	return false, ""
}

func checkZiXiangGongZhao(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// Find 紫微 and 天相 in trine/opposition relationship
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"紫微"}) {
			if hasTrine(chart, i, []string{"天相"}) || hasOpposition(chart, i, []string{"天相"}) {
				return true, "紫相拱照"
			}
		}
	}
	return false, ""
}

func checkShaPoLang(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasSha := starInPalace(chart, i, []string{"七杀"})
		hasPo := starInPalace(chart, i, []string{"破军"})
		hasTan := starInPalace(chart, i, []string{"贪狼"})
		if (hasSha && hasPo) || (hasSha && hasTan) || (hasPo && hasTan) {
			return true, "杀破狼格"
		}
	}
	// Also check opposition (三方四正)
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"七杀", "破军", "贪狼"}) {
			count := 0
			for _, s := range chart.Palaces[i].MainStars {
				if s == "七杀" || s == "破军" || s == "贪狼" {
					count++
				}
			}
			if count >= 2 {
				return true, "杀破狼格"
			}
			if hasTrine(chart, i, []string{"七杀", "破军", "贪狼"}) || hasOpposition(chart, i, []string{"七杀", "破军", "贪狼"}) {
				return true, "杀破狼格"
			}
		}
	}
	return false, ""
}

func checkJiYueTongLiang(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 天机、太阴、天同、天梁 in trine
	for i := range chart.Palaces {
		stars := chart.Palaces[i].MainStars
		count := 0
		for _, s := range stars {
			if s == "天机" || s == "太阴" || s == "天同" || s == "天梁" {
				count++
			}
		}
		if count >= 2 {
			trine1 := (i + 4) % 12
			trine2 := (i + 8) % 12
			tCount := 0
			for _, s := range chart.Palaces[trine1].MainStars {
				if s == "天机" || s == "太阴" || s == "天同" || s == "天梁" {
					tCount++
				}
			}
			for _, s := range chart.Palaces[trine2].MainStars {
				if s == "天机" || s == "太阴" || s == "天同" || s == "天梁" {
					tCount++
				}
			}
			if tCount >= 2 {
				return true, "机月同梁格"
			}
		}
	}
	return false, ""
}

func checkZiWuLianFu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		stars := chart.Palaces[i].MainStars
		count := 0
		for _, s := range stars {
			if s == "紫微" || s == "武曲" || s == "廉贞" || s == "天府" {
				count++
			}
		}
		if count >= 3 {
			return true, "紫武廉府"
		}
	}
	return false, ""
}

func checkFuXiangChaoYuan(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天府"}) {
			if hasTrine(chart, i, []string{"天相"}) || hasOpposition(chart, i, []string{"天相"}) {
				return true, "府相朝垣"
			}
		}
	}
	return false, ""
}

func checkRiYueGongZhao(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"太阳"}) {
			if hasTrine(chart, i, []string{"太阴"}) || hasOpposition(chart, i, []string{"太阴"}) {
				return true, "日月拱照"
			}
		}
	}
	return false, ""
}

func checkRiYueFanBei(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 日月反背: 太阳太阴在迁移宫且太阳在陷地
	for i, p := range chart.Palaces {
		if p.Name == "遷移宮" {
			if starInPalace(chart, i, []string{"太阳", "太阴"}) {
				if hasBrightness(chart, i, "太阳", []string{"陷", "不"}) {
					return true, "日月反背"
				}
			}
		}
	}
	return false, ""
}

func checkLuMaJiaoChi(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasLu := auxStarInPalace(chart, i, []string{"禄存"})
		hasMa := auxStarInPalace(chart, i, []string{"天马"})
		if hasLu && hasMa {
			return true, "禄马交驰"
		}
		if hasLu && hasOpposition(chart, i, []string{"天马"}) {
			return true, "禄马交驰"
		}
	}
	return false, ""
}

func checkTianMaGongMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 天马在命宫三方四正
	for i, p := range chart.Palaces {
		if p.Name == "命宮" || hasTrine(chart, i, []string{"天马"}) || hasOpposition(chart, i, []string{"天马"}) {
			for j := range chart.Palaces {
				if auxStarInPalace(chart, j, []string{"天马"}) {
					return true, "天马拱命"
				}
			}
		}
	}
	return false, ""
}

func checkHuoTanGe(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 火星/铃星 + 贪狼同宫
	for i := range chart.Palaces {
		hasHuo := auxStarInPalace(chart, i, []string{"火星"})
		hasTan := starInPalace(chart, i, []string{"贪狼"})
		if hasHuo && hasTan {
			return true, "火贪格"
		}
	}
	return false, ""
}

func checkLingTanGe(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasLing := auxStarInPalace(chart, i, []string{"铃星"})
		hasTan := starInPalace(chart, i, []string{"贪狼"})
		if hasLing && hasTan {
			return true, "铃贪格"
		}
	}
	return false, ""
}

func checkQingYangRuMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i, p := range chart.Palaces {
		if p.Name == "命宮" {
			if auxStarInPalace(chart, i, []string{"擎羊"}) {
				return true, "擎羊入命"
			}
		}
	}
	return false, ""
}

func checkTuoLuoRuMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := 0; i < 12; i++ {
		if chart.Palaces[i].Name == "命宮" {
			if auxStarInPalace(chart, i, []string{"陀罗"}) {
				return true, "陀罗入命"
			}
		}
	}
	return false, ""
}

func checkKongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := 0; i < 12; i++ {
		p := chart.Palaces[i]
		if p.Name == "命宮" {
			if len(p.MainStars) == 0 {
				return true, "空宫"
			}
		}
	}
	return false, ""
}

func checkTaoHuaFanZhu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if len(chart.Palaces[i].MainStars) > 0 {
			count := 0
			for _, aux := range chart.Palaces[i].AuxStars {
				if aux == "左辅" || aux == "右弼" || aux == "文昌" || aux == "文曲" {
					count++
				}
			}
			if count >= 2 {
				for _, main := range chart.Palaces[i].MainStars {
					if main == "贪狼" || main == "紫微" || main == "太阴" {
						return true, "桃花犯主"
					}
				}
			}
		}
	}
	return false, ""
}

func checkShuiMuQingHua(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天机"}) && starInPalace(chart, i, []string{"太阴"}) {
			return true, "水木清华"
		}
	}
	return false, ""
}

func checkTuJinXiangSheng(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天府"}) && starInPalace(chart, i, []string{"武曲"}) {
			return true, "土金相生"
		}
	}
	return false, ""
}

func checkRiYueBingMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	yangIdx := -1
	yinIdx := -1
	for i := 0; i < 12; i++ {
		if starInPalace(chart, i, []string{"太阳"}) {
			yangIdx = i
		}
		if starInPalace(chart, i, []string{"太阴"}) {
			yinIdx = i
		}
	}
	if yangIdx >= 0 && yinIdx >= 0 {
		return true, "日月并明"
	}
	return false, ""
}

func checkJiXiangLiMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 紫微在午宫坐命
	for i := 0; i < 12; i++ {
		p := chart.Palaces[i]
		if p.Name == "命宮" && starInPalace(chart, i, []string{"紫微"}) {
			// 午宫 is index 6 (遷移宮 is 命宮+6)
			// Check if 命宮 is 午 (we'd need branch info)
			// For simplicity, check brightness 庙 in 迁移宫
			for j := 0; j < 12; j++ {
				p2 := chart.Palaces[j]
				if p2.Name == "遷移宮" && starInPalace(chart, j, []string{"紫微"}) {
					if hasBrightness(chart, j, "紫微", []string{"庙"}) {
						return true, "极向离明"
					}
				}
			}
		}
	}
	return false, ""
}

func checkShiZhongYinYu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"巨门"}) {
			if hasBrightness(chart, i, "巨门", []string{"庙", "旺", "得"}) {
				// Check if has 化禄 or 化权 in same palace
				for _, t := range chart.Palaces[i].FourHua {
					if strings.Contains(t, "化禄") || strings.Contains(t, "化权") {
						return true, "石中隐玉"
					}
				}
			}
		}
	}
	return false, ""
}

func checkWenGuiWenHua(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasChang := auxStarInPalace(chart, i, []string{"文昌"})
		hasQu := auxStarInPalace(chart, i, []string{"文曲"})
		if hasChang || hasQu {
			if hasTrine(chart, i, []string{"文昌", "文曲"}) || hasOpposition(chart, i, []string{"文昌", "文曲"}) {
				return true, "文桂文华"
			}
		}
	}
	return false, ""
}

func checkTianFuShouYuan(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := 0; i < 12; i++ {
		if starInPalace(chart, i, []string{"天府"}) && chart.Palaces[i].Name == "田宅宮" {
			return true, "天府守垣"
		}
	}
	return false, ""
}

func checkShouXingRuMiao(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天梁"}) {
			if hasBrightness(chart, i, "天梁", []string{"庙", "旺"}) {
				return true, "寿星入庙"
			}
		}
	}
	return false, ""
}

func checkMaTouDaiJian(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasMa := auxStarInPalace(chart, i, []string{"天马"})
		hasYang := auxStarInPalace(chart, i, []string{"擎羊"})
		if hasMa && hasYang {
			return true, "马头带剑"
		}
	}
	return false, ""
}

func checkJunZiZaiYe(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 君子在野: 紫微在亥宫且不加吉
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"紫微"}) {
			b := chart.Palaces[i].Brightness
			if br, ok := b["紫微"]; ok && (br == "陷" || br == "不") {
				count := 0
				for _, aux := range chart.Palaces[i].AuxStars {
					if aux == "左辅" || aux == "右弼" || aux == "天魁" || aux == "天钺" {
						count++
					}
				}
				if count == 0 {
					return true, "君子在野"
				}
			}
		}
	}
	return false, ""
}

func checkJuRiTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"巨门"}) && starInPalace(chart, i, []string{"太阳"}) {
			return true, "巨日同宫"
		}
	}
	return false, ""
}

func checkKeMingHuiLu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		for _, t := range chart.Palaces[i].FourHua {
			if strings.Contains(t, "化科") || strings.Contains(t, "化禄") {
				if starInPalace(chart, i, []string{"天梁"}) || starInPalace(chart, i, []string{"天府"}) {
					return true, "科名会禄"
				}
			}
		}
	}
	return false, ""
}

func checkCaiYinJiaMa(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasCai := auxStarInPalace(chart, i, []string{"禄存"})
		hasYin := auxStarInPalace(chart, i, []string{"右弼"})
		hasMa := auxStarInPalace(chart, i, []string{"天马"})
		if hasCai && hasYin && hasMa {
			return true, "财印夹马"
		}
	}
	return false, ""
}

func checkSanQiJiaHui(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 三奇: 化禄、化权、化科同时出现在命宫三方四正
	huaTypes := make(map[string]int)
	for i := range chart.Palaces {
		for _, t := range chart.Palaces[i].FourHua {
			if strings.Contains(t, "化禄") {
				huaTypes["化禄"]++
			}
			if strings.Contains(t, "化权") {
				huaTypes["化权"]++
			}
			if strings.Contains(t, "化科") {
				huaTypes["化科"]++
			}
		}
	}
	if huaTypes["化禄"] > 0 && huaTypes["化权"] > 0 && huaTypes["化科"] > 0 {
		return true, "三奇嘉会"
	}
	return false, ""
}

func checkTianYiTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天魁"}) && starInPalace(chart, i, []string{"天梁"}) {
			return true, "天乙同宫"
		}
	}
	return false, ""
}

func checkQiShaChaoDou(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"七杀"}) {
			if hasBrightness(chart, i, "七杀", []string{"庙", "旺"}) {
				return true, "七杀朝斗"
			}
		}
	}
	return false, ""
}

func checkWuTanGe(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"武曲"}) && starInPalace(chart, i, []string{"贪狼"}) {
			return true, "武贪格"
		}
	}
	return false, ""
}

func checkLianFuShuangXing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"廉贞", "天府"}) {
			return true, "廉府双星"
		}
	}
	return false, ""
}

func checkTongLiangShuangXing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天同", "天梁"}) {
			return true, "同梁双星"
		}
	}
	return false, ""
}

func checkRiYueJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 太阳太阴夹命宫
	for i, p := range chart.Palaces {
		if p.Name == "命宮" {
			yangIdx := (i + 11) % 12 // 前一宫
			yinIdx := (i + 1) % 12    // 后一宫
			if starInPalace(chart, yangIdx, []string{"太阳"}) && starInPalace(chart, yinIdx, []string{"太阴"}) {
				return true, "日月夹命"
			}
		}
	}
	return false, ""
}

func checkFuBiJiaYin(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 左辅右弼夹命宫
	for i, p := range chart.Palaces {
		if p.Name == "命宮" {
			idx1 := (i + 11) % 12
			idx2 := (i + 1) % 12
			hasZuo := auxStarInPalace(chart, idx1, []string{"左辅"}) || auxStarInPalace(chart, idx2, []string{"左辅"})
			hasYou := auxStarInPalace(chart, idx1, []string{"右弼"}) || auxStarInPalace(chart, idx2, []string{"右弼"})
			if hasZuo && hasYou {
				return true, "辅弼夹印"
			}
		}
	}
	return false, ""
}

// ──────────────────── Sanfang Sizheng Computation ────────────────────

// ComputeSanfangSizheng computes the 三方四正 for a given palace index.
// Returns [3]int: {oppositeIndex, trine1Index, trine2Index}
func ComputeSanfangSizheng(palaceIdx int) [3]int {
	opposite := (palaceIdx + 6) % 12
	trine1 := (palaceIdx + 4) % 12
	trine2 := (palaceIdx + 8) % 12
	return [3]int{opposite, trine1, trine2}
}

// SanfangSizhengResult holds the sanfang analysis for one palace.
type SanfangSizhengResult struct {
	Opposite string `json:"opposite"` // 对宫 palace name
	Trine1   string `json:"trine1"`   // 三合宫 1
	Trine2   string `json:"trine2"`   // 三合宫 2
}

// GetPalaceSanfang returns the sanfang sizheng result for a palace index.
func GetPalaceSanfang(palaceIdx int) *SanfangSizhengResult {
	sf := ComputeSanfangSizheng(palaceIdx)
	return &SanfangSizhengResult{
		Opposite: PALACE_NAMES[sf[0]],
		Trine1:   PALACE_NAMES[sf[1]],
		Trine2:   PALACE_NAMES[sf[2]],
	}
}

// EnhancedSanfangResult holds detailed sanfang analysis including star energy from SiHua.
type EnhancedSanfangResult struct {
	Opposite       string `json:"opposite"`        // 对宫 palace name
	OppositeIdx    int    `json:"opposite_idx"`    // 对宫 palace index
	Trine1         string `json:"trine1"`          // 三合宫 1
	Trine1Idx      int    `json:"trine1_idx"`       // 三合宫 1 index
	Trine2         string `json:"trine2"`          // 三合宫 2
	Trine2Idx      int    `json:"trine2_idx"`      // 三合宫 2 index
	OppositeSihua  string `json:"opposite_sihua"`  // 四化能量冲宫描述
	TrineSihua     string `json:"trine_sihua"`      // 三合拱照四化描述
}

// GetEnhancedSanfang returns detailed sanfang analysis with SiHua interaction descriptions.
func GetEnhancedSanfang(chart *ZiWeiChart, palaceIdx int) *EnhancedSanfangResult {
	sf := ComputeSanfangSizheng(palaceIdx)
	result := &EnhancedSanfangResult{
		Opposite:    PALACE_NAMES[sf[0]],
		OppositeIdx: sf[0],
		Trine1:      PALACE_NAMES[sf[1]],
		Trine1Idx:   sf[1],
		Trine2:      PALACE_NAMES[sf[2]],
		Trine2Idx:   sf[2],
	}

	if chart == nil {
		return result
	}

	var oppSihua, trineSihua []string

	// Check if any SiHua stars in opposite palace affect current palace
	oppPalace := chart.Palaces[sf[0]]
	curPalace := chart.Palaces[palaceIdx]

	for _, t := range oppPalace.FourHua {
		for _, curStar := range curPalace.MainStars {
			if strings.Contains(t, curStar) {
				oppSihua = append(oppSihua, t+"照"+PALACE_NAMES[palaceIdx])
			}
		}
	}

	// Check trine palaces for SiHua energy
	for _, triIdx := range []int{sf[1], sf[2]} {
		triPalace := chart.Palaces[triIdx]
		for _, t := range triPalace.FourHua {
			for _, curStar := range curPalace.MainStars {
				if strings.Contains(t, curStar) {
					trineSihua = append(trineSihua, t+"拱"+PALACE_NAMES[palaceIdx])
				}
			}
		}
	}

	if len(oppSihua) > 0 {
		result.OppositeSihua = strings.Join(oppSihua, "、")
	}
	if len(trineSihua) > 0 {
		result.TrineSihua = strings.Join(trineSihua, "、")
	}

	return result
}

// ──────────────────── Star Brightness Fallback ────────────────────

// GetStarBrightness returns the brightness description for a star+brightness combination.
// Falls back to "平" if the combination is unknown.
func GetStarBrightness(star, brightness string) string {
	if starMap, ok := STAR_BRIGHTNESS[star]; ok {
		if desc, ok := starMap[brightness]; ok {
			return desc
		}
	}
	// Fallback
	if starMap, ok := STAR_BRIGHTNESS[star]; ok {
		if desc, ok := starMap["平"]; ok {
			return desc
		}
	}
	// Try auxiliary stars
	if auxMap, ok := STAR_BRIGHTNESS_AUX[star]; ok {
		if desc, ok := auxMap[brightness]; ok {
			return desc
		}
		if desc, ok := auxMap["平"]; ok {
			return desc
		}
	}
	return "此星曜在此宫位亮度普通，运势平稳。"
}

// ──────────────────── Heming (合盘) Analysis ────────────────────

// HemingResult holds the result of a two-chart compatibility analysis.
type HemingResult struct {
	OverallScore     int                   `json:"overall_score"`      // 0-100
	YuanFenType      string                `json:"yuanfen_type"`       // 帝旺格/长生格/墓绝格/胎养格/普通格
	MarriageTiming   string                `json:"marriage_timing"`     // 婚期判断 guidance
	ShuangGongLianCan []ShuangGongScore     `json:"shuang_gong_lian_can"` // 双宫联参 scores
	WuBuXingYun      []WuBuScore           `json:"wu_bu_xing_yun"`      // 合盘五步行运
	Confidence       float64               `json:"confidence"`         // 0.0-1.0 how definitive
	Notes            string                `json:"notes"`              // additional notes
}

// ShuangGongScore holds a dual-palace comparison score.
type ShuangGongScore struct {
	Palace     string `json:"palace"`      // palace name
	ChartAScore int    `json:"chart_a_score"` // 0-100
	ChartBScore int    `json:"chart_b_score"` // 0-100
	Harmony     string `json:"harmony"`     // 和谐/冲突/中等
}

// WuBuScore holds one step of the 5-step heming method.
type WuBuScore struct {
	Step        string `json:"step"`         // step name
	Description string `json:"description"`   // what's being evaluated
	Score       int    `json:"score"`         // 0-100
	Interpretation string `json:"interpretation"` // 解释
}

// AnalyzeHeming performs a full compatibility analysis between two charts.
func AnalyzeHeming(chartA, chartB *ZiWeiChart) *HemingResult {
	if chartA == nil || chartB == nil {
		return nil
	}

	result := &HemingResult{
		ShuangGongLianCan: make([]ShuangGongScore, 0),
		WuBuXingYun:       make([]WuBuScore, 0),
		Confidence:        0.85,
	}

	// 双宫联参: compare same-name palaces across two charts
	shuangGongPalaces := []string{"命宮", "夫妻宮", "福德宮", "官祿宮", "財帛宮"}
	for _, palaceName := range shuangGongPalaces {
		scoreA := computePalaceScore(chartA, palaceName)
		scoreB := computePalaceScore(chartB, palaceName)
		harmony := computeHarmony(scoreA, scoreB)
		result.ShuangGongLianCan = append(result.ShuangGongLianCan, ShuangGongScore{
			Palace:     palaceName,
			ChartAScore: scoreA,
			ChartBScore: scoreB,
			Harmony:    harmony,
		})
	}

	// 合盘五步行运
	fiveSteps := []string{"命宫互看", "夫妻宫对冲", "福德宫共鸣", "官禄宫合作", "财帛宫流通"}
	for i, step := range fiveSteps {
		score := computeFiveStepScore(chartA, chartB, i)
		result.WuBuXingYun = append(result.WuBuXingYun, WuBuScore{
			Step:          step,
			Description:  getFiveStepDesc(i),
			Score:         score,
			Interpretation: getFiveStepInterp(score),
		})
	}

	// 缘分类型
	result.YuanFenType = classifyYuanFen(chartA, chartB)

	// 婚期判断
	result.MarriageTiming = estimateMarriageTiming(chartA, chartB)

	// Overall score: weighted average
	totalScore := 0
	for _, sg := range result.ShuangGongLianCan {
		totalScore += (sg.ChartAScore + sg.ChartBScore) / 2
	}
	for _, wb := range result.WuBuXingYun {
		totalScore += wb.Score
	}
	count := len(result.ShuangGongLianCan) + len(result.WuBuXingYun)
	if count > 0 {
		result.OverallScore = totalScore / count
	}

	// Cap at 100
	if result.OverallScore > 100 {
		result.OverallScore = 100
	}

	return result
}

func computePalaceScore(chart *ZiWeiChart, palaceName string) int {
	if chart == nil {
		return 50
	}
	for _, p := range chart.Palaces {
		if p.Name == palaceName {
			score := 50
			// Main star quality (庙/旺 = +20, 得/利 = +10, 平 = 0, 陷/不 = -10)
			for _, main := range p.MainStars {
				if b, ok := p.Brightness[main]; ok {
					switch b {
					case "庙", "旺":
						score += 15
					case "得", "利":
						score += 8
					case "平":
						score += 0
					case "陷", "不":
						score -= 10
					}
				}
			}
			// Auxiliary stars add bonus
			score += min(len(p.AuxStars)*3, 15)
			// Fourhua presence
			score += min(len(p.FourHua)*5, 20)
			return min(max(score, 0), 100)
		}
	}
	return 50
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func computeHarmony(scoreA, scoreB int) string {
	diff := abs(scoreA - scoreB)
	if diff <= 15 {
		return "和谐"
	}
	if diff <= 30 {
		return "中等"
	}
	return "冲突"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func computeFiveStepScore(chartA, chartB *ZiWeiChart, step int) int {
	if chartA == nil || chartB == nil {
		return 50
	}
	switch step {
	case 0: // 命宫互看 - compare life palaces
		scoreA := computePalaceScore(chartA, "命宮")
		scoreB := computePalaceScore(chartB, "命宮")
		return (scoreA + scoreB) / 2
	case 1: // 夫妻宫对冲 - check opposition
		scoreA := computePalaceScore(chartA, "夫妻宮")
		scoreB := computePalaceScore(chartB, "夫妻宮")
		return (scoreA + scoreB) / 2
	case 2: // 福德宫共鸣
		scoreA := computePalaceScore(chartA, "福德宮")
		scoreB := computePalaceScore(chartB, "福德宮")
		return (scoreA + scoreB) / 2
	case 3: // 官禄宫合作
		scoreA := computePalaceScore(chartA, "官祿宮")
		scoreB := computePalaceScore(chartB, "官祿宮")
		return (scoreA + scoreB) / 2
	case 4: // 财帛宫流通
		scoreA := computePalaceScore(chartA, "財帛宮")
		scoreB := computePalaceScore(chartB, "財帛宮")
		return (scoreA + scoreB) / 2
	}
	return 50
}

func getFiveStepDesc(step int) string {
	descs := []string{
		"分析双方命宫的星曜配置与亮度，配合度越高则事业与人生观越契合",
		"检查夫妻宫的对宫关系，对宫相冲则婚姻需多磨合",
		"比较双方福德宫，精神层面的共鸣程度影响长期关系稳定",
		"分析官禄宫的合作潜力，适合共同创业或事业合作",
		"评估财帛宫的财运流通，财星相生则经济互补",
	}
	if step >= 0 && step < len(descs) {
		return descs[step]
	}
	return ""
}

func getFiveStepInterp(score int) string {
	if score >= 80 {
		return "非常匹配"
	}
	if score >= 60 {
		return "较好匹配"
	}
	if score >= 40 {
		return "中等匹配"
	}
	return "需多加磨合"
}

func classifyYuanFen(chartA, chartB *ZiWeiChart) string {
	if chartA == nil || chartB == nil {
		return "普通格"
	}
	// 缘分类型 based on stem compatibility
	stemA := int(chartA.birthInfo.YearPillar.Stem)
	stemB := int(chartB.birthInfo.YearPillar.Stem)

	// Same stem = 帝旺格 (strongest)
	if stemA == stemB {
		return "帝旺格"
	}
	// 相生相合 = 长生格
	if (stemA+1)%10 == stemB || (stemB+1)%10 == stemA {
		return "长生格"
	}
	// 相克 = 墓绝格
	if (stemA+5)%10 == stemB || (stemB+5)%10 == stemA {
		return "墓绝格"
	}
	// 其他 = 普通格
	return "普通格"
}

func estimateMarriageTiming(chartA, chartB *ZiWeiChart) string {
	if chartA == nil || chartB == nil {
		return "需结合双方大限与流年综合判断"
	}
	// Simple heuristic: check if both have 化禄 in 夫妻宫 or similar patterns
	hasLuA := false
	hasLuB := false
	for _, p := range chartA.Palaces {
		if p.Name == "夫妻宮" {
			for _, t := range p.FourHua {
				if strings.Contains(t, "化禄") {
					hasLuA = true
				}
			}
		}
	}
	for _, p := range chartB.Palaces {
		if p.Name == "夫妻宮" {
			for _, t := range p.FourHua {
				if strings.Contains(t, "化禄") {
					hasLuB = true
				}
			}
		}
	}
	if hasLuA && hasLuB {
		return "双化禄入夫妻宫，早婚之命，宜在25-30岁间成婚"
	}
	if hasLuA || hasLuB {
		return "单化禄入夫妻，婚期适中，约在28-35岁间"
	}
	return "婚期需综合大限与流年判断，通常在30-40岁间有较好时机"
}

// ──────────────────── Palace Interpretation ────────────────────

// PalaceReading holds the full interpretation for one palace.
type PalaceReading struct {
	MainStarAnalysis    string   `json:"main_star_analysis"`
	AuxStarInfluence    string   `json:"aux_star_influence"`
	SihuaInfluence      string   `json:"sihua_influence"`
	SanfangAnalysis     string   `json:"sanfang_analysis"`
	PatternNotes        string   `json:"pattern_notes"`
	Brightness          string   `json:"brightness"`
}

// GetPalaceReading returns a full template-based reading for a palace.
func GetPalaceReading(chart *ZiWeiChart, palaceIdx int) *PalaceReading {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return nil
	}

	p := chart.Palaces[palaceIdx]
	reading := &PalaceReading{}

	// Main star analysis
	if len(p.MainStars) > 0 {
		mainStar := p.MainStars[0]
		brightness := "平"
		if p.Brightness != nil {
			if br, ok := p.Brightness[mainStar]; ok {
				brightness = br
			}
		}
		reading.MainStarAnalysis = GetStarBrightness(mainStar, brightness)
		reading.Brightness = brightness
	}

	// Auxiliary star influence
	if len(p.AuxStars) > 0 {
		influence := buildAuxStarInfluence(p.AuxStars)
		reading.AuxStarInfluence = influence
	}

	// Sihua influence
	if len(p.FourHua) > 0 {
		influence := buildSihuaInfluence(p.FourHua)
		reading.SihuaInfluence = influence
	}

	// Sanfang analysis - enhanced with SiHua interaction
	sf := GetPalaceSanfang(palaceIdx)
	enhanced := GetEnhancedSanfang(chart, palaceIdx)

	base := fmt.Sprintf("对宫%s，三合%s与%s，形成三方四正格局。",
		sf.Opposite, sf.Trine1, sf.Trine2)

	// Add SiHua interaction details
	var extra []string
	if enhanced.OppositeSihua != "" {
		extra = append(extra, enhanced.OppositeSihua)
	}
	if enhanced.TrineSihua != "" {
		extra = append(extra, enhanced.TrineSihua)
	}
	if len(extra) > 0 {
		reading.SanfangAnalysis = base + " " + strings.Join(extra, "，") + "。"
	} else {
		reading.SanfangAnalysis = base
	}

	// Pattern notes - check if any pattern affects this palace
	var notes []string
	if present, name := checkZiFuTongGong(chart); present && starInPalace(chart, palaceIdx, []string{"紫微", "天府"}) {
		notes = append(notes, name)
	}
	if present, name := checkShaPoLang(chart); present && starInPalace(chart, palaceIdx, []string{"七杀", "破军", "贪狼"}) {
		notes = append(notes, name)
	}
	if len(notes) > 0 {
		reading.PatternNotes = strings.Join(notes, "、") + "格局在此宫显现"
	}

	return reading
}

func buildAuxStarInfluence(auxStars []string) string {
	if len(auxStars) == 0 {
		return ""
	}
	var parts []string
	for _, star := range auxStars {
		if desc := getAuxStarDesc(star); desc != "" {
			parts = append(parts, desc)
		}
	}
	return strings.Join(parts, "；")
}

func getAuxStarDesc(star string) string {
	descMap := map[string]string{
		"左辅": "左輔星入此宮，有貴人相助，為人正直可靠，處事穩重。",
		"右弼": "右弼星入此宮，得同儕或晚輩之助，善於配合他人，團隊協作能力強。",
		"文昌": "文昌星入此宮，才華出眾，學業優異，思維敏捷，能言善辯。",
		"文曲": "文曲星入此宮，才藝出眾，具有藝術天賦和審美眼光，口才流利。",
		"天魁": "天魁星入此宮，為天乙貴人，一生多得貴人提攜，遇事逢凶化吉。",
		"天钺": "天鉞星入此宮，為玉堂貴人，多得好友和同事之助，具有號召力。",
		"擎羊": "擎羊星入此宮，性格剛烈好勝，行事衝動，須注意控制脾氣。",
		"陀罗": "陀羅星入此宮，做事拖延猶豫，容易錯失良機，須培養果斷的習慣。",
		"火星": "火星入此宮，脾氣急躁，行事衝動，容易因一時之氣而誤事。",
		"铃星": "鈴星入此宮，內心多憂慮煩惱，容易鑽牛角尖，須學會釋放壓力。",
		"地空": "地空星入此宮，想法天馬行空，有時脫離現實，須務實面對人生。",
		"地劫": "地劫星入此宮，人生多波折起伏，須做好風險防範和應急準備。",
		"禄存": "祿存入此宮，財祿豐厚，衣食無憂，善於累積財富。",
		"天马": "天馬入此宮，奔波勞碌，一生多變動和遷移，適合需要流動性的工作。",
	}
	if desc, ok := descMap[star]; ok {
		return desc
	}
	return ""
}

func buildSihuaInfluence(fourHua []string) string {
	if len(fourHua) == 0 {
		return ""
	}
	var parts []string
	for _, t := range fourHua {
		if desc := getFourHuaDesc(t); desc != "" {
			parts = append(parts, desc)
		}
	}
	return strings.Join(parts, "；")
}

func getFourHuaDesc(star string) string {
	huaTypeMap := map[string]int{"化禄": 0, "化权": 1, "化科": 2, "化忌": 3}
	// Parse star name to find the transformation type
	for huaIdx, stars := range SI_HUA_TABLE {
		for _, s := range stars {
			if strings.Contains(star, s) {
				return getHuaDesc(s, huaIdx)
			}
		}
	}
	// Direct match from fourHua strings
	for huaStr, huaIdx := range huaTypeMap {
		if strings.Contains(star, huaStr) {
			// Extract star name
			starName := strings.TrimPrefix(star, huaStr)
			return getHuaDesc(starName, huaIdx)
		}
	}
	return star // fallback
}

func getHuaDesc(starName string, huaType int) string {
	// Get description from fourHua templates if available
	// For now, return a simple description
	huaNames := []string{"化禄", "化权", "化科", "化忌"}
	if huaType >= 0 && huaType < 4 {
		return fmt.Sprintf("%s%s，表示%s", starName, huaNames[huaType], getHuaEffectSummary(huaType))
	}
	return starName
}

func getHuaEffectSummary(huaType int) string {
	effects := []string{
		"财运与事业顺利发展",
		"权力与执行力增强",
		"学业与名声提升",
		"需注意是非与调整",
	}
	if huaType >= 0 && huaType < len(effects) {
		return effects[huaType]
	}
	return ""
}

// ──────────────────── Adjective Stars (杂耀星) Computation ────────────────────

// ComputeAdjectiveStars computes all adjective star (杂耀星) placements across 12 palaces.
// Returns a map of palace index → list of adjective star names.
func ComputeAdjectiveStars(chart *ZiWeiChart) map[int][]string {
	if chart == nil {
		return nil
	}
	result := make(map[int][]string)
	for i := 0; i < 12; i++ {
		result[i] = []string{}
	}
	if chart.engineChart == nil {
		return result
	}

	yearBranchIdx := int(chart.engineChart.YearPillar.Branch) % 12
	monthIdx := int(chart.engineChart.MonthPillar.Branch) % 12

	// 红鸾: start=卯(1), direction=-1, steps=yearBranchIndex
	// target = (start - steps + 12) % 12 = (1 - yearBranchIdx + 12) % 12
	hongluanIdx := (1 - yearBranchIdx + 12) % 12
	result[hongluanIdx] = append(result[hongluanIdx], "红鸾")

	// 天喜: 对宫 of 红鸾
	tianxiIdx := (hongluanIdx + 6) % 12
	result[tianxiIdx] = append(result[tianxiIdx], "天喜")

	// 咸池: per-year branch lookup
	xianchiIdx := XIANCHI_TABLE[yearBranchIdx]
	result[xianchiIdx] = append(result[xianchiIdx], "咸池")

	// 华盖: per-year branch lookup
	huagaiIdx := HUAGAI_TABLE[yearBranchIdx]
	result[huagaiIdx] = append(result[huagaiIdx], "华盖")

	// 天姚: start=丑(1), direction=+1, steps=monthIndex
	// target = (1 + monthIdx) % 12
	tianyaoIdx := (1 + monthIdx) % 12
	result[tianyaoIdx] = append(result[tianyaoIdx], "天姚")

	// 天刑: start=酉(9), direction=+1, steps=monthIndex
	// target = (9 + monthIdx) % 12
	tianxingIdx := (9 + monthIdx) % 12
	result[tianxingIdx] = append(result[tianxingIdx], "天刑")

	// 阴煞: per-month lookup
	yinshaIdx := YINSHIA_TABLE[monthIdx]
	result[yinshaIdx] = append(result[yinshaIdx], "阴煞")

	// 破碎: 3-cycle per year branch
	pousuiIdx := POUSUI_TABLE[yearBranchIdx]
	result[pousuiIdx] = append(result[pousuiIdx], "破碎")

	// 飞廉: per-year branch lookup
	feilianIdx := FEILIAN_TABLE[yearBranchIdx]
	result[feilianIdx] = append(result[feilianIdx], "飞廉")

	// 龙池: start=辰(4), direction=+1, steps=yearBranchIndex
	// target = (4 + yearBranchIdx) % 12
	longchiIdx := (4 + yearBranchIdx) % 12
	result[longchiIdx] = append(result[longchiIdx], "龙池")

	// 凤阁: start=戌(10), direction=-1, steps=yearBranchIndex
	// target = (10 - yearBranchIdx + 12) % 12
	fenggeIdx := (10 - yearBranchIdx + 12) % 12
	result[fenggeIdx] = append(result[fenggeIdx], "凤阁")

	// 天空: start=(yearBranch+1), direction=+1, steps=1
	// target = (yearBranchIdx + 1 + 1) % 12 = (yearBranchIdx + 2) % 12
	tiankongIdx := (yearBranchIdx + 2) % 12
	result[tiankongIdx] = append(result[tiankongIdx], "天空")

	return result
}

// ComputeTwelveShen computes all 4 twelve-shen systems for each palace.
// Uses yearBranch and fiveElementClass from chart.birthInfo.
// Direction rule: 阳男/阴女 = forward (+i), 阴男/阳女 = backward (-i)
// Five element class derived from yearStem: 甲(1)→3(木), 乙(2)→3(木), 丙(3)→6(火), 丁(4)→6(火),
// 戊(5)→5(土), 己(6)→5(土), 庚(7)→4(金), 辛(8)→4(金), 壬(9)→2(水), 癸(10)→2(水)
func ComputeTwelveShen(chart *ZiWeiChart) [12]struct {
	Changsheng, Boshi, Jiangqian, Suiqian string
} {
	if chart == nil || chart.engineChart == nil {
		return [12]struct{ Changsheng, Boshi, Jiangqian, Suiqian string }{}
	}

	result := [12]struct{ Changsheng, Boshi, Jiangqian, Suiqian string }{}

	// Derive five element class from year stem
	yearStem := int(chart.birthInfo.YearPillar.Stem) % 10
	fiveElementClass := map[int]int{0: 3, 1: 3, 2: 6, 3: 6, 4: 5, 5: 5, 6: 4, 7: 4, 8: 2, 9: 2}[yearStem]

	// Direction: based on year branch parity vs gender
	yearBranchIdx := int(chart.engineChart.YearPillar.Branch) % 12
	gender := chart.birthInfo.Sex
	yearBranchYang := yearBranchIdx % 2 // 0=阳, 1=阴
	genderYang := func() int {
		if gender == basis.SexMale {
			return 0
		}
		return 1
	}()
	direction := 1
	if yearBranchYang == genderYang {
		direction = -1 // 阴男/阳女 = backward
	}

	// CHANGSHENG_12: start from CHANGSHENG_START_TABLE[fiveElementClass], apply direction
	changshengStart := CHANGSHENG_START_TABLE[fiveElementClass]
	for i := 0; i < 12; i++ {
		idx := (changshengStart + direction*i + 12) % 12
		result[idx].Changsheng = CHANGSHENG_12[i]
	}

	// BOSHI_12: start from 禄存 position (from LUCUN_TABLE[yearStem])
	luIdx := LUCUN_TABLE[yearStem]
	for i := 0; i < 12; i++ {
		idx := (luIdx + direction*i + 12) % 12
		result[idx].Boshi = BOSHI_12[i]
	}

	// JIANG_QIAN_12: start from year branch group
	yearBranchGroup := getYearBranchGroup(yearBranchIdx)
	jiangStart := JIANG_QIAN_START_TABLE[yearBranchGroup]
	for i := 0; i < 12; i++ {
		idx := (jiangStart + i) % 12
		result[idx].Jiangqian = JIANG_QIAN_12[i]
	}

	// SUI_QIAN_12: start from year branch itself, forward
	for i := 0; i < 12; i++ {
		idx := (yearBranchIdx + i) % 12
		result[idx].Suiqian = SUI_QIAN_12[i]
	}

	return result
}

func getYearBranchGroup(branchIdx int) string {
	groups := map[int]string{
		0:  "寅午戌",
		1:  "寅午戌",
		4:  "寅午戌",
		10: "寅午戌",
		7:  "申子辰",
		8:  "申子辰",
		11: "申子辰",
		3:  "申子辰",
		2:  "巳酉丑",
		6:  "巳酉丑",
		9:  "巳酉丑",
		5:  "巳酉丑",
	}
	if g, ok := groups[branchIdx]; ok {
		return g
	}
	return "寅午戌" // fallback
}