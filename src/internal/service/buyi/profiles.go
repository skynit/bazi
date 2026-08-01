package buyi

type trigramProfile struct {
	Name    string
	Symbol  string
	Quality string
	Excess  string
}

var trigramProfiles = map[rune]trigramProfile{
	'天': {Name: "乾", Symbol: "天", Quality: "主动、创造与自律", Excess: "用力过猛"},
	'地': {Name: "坤", Symbol: "地", Quality: "承载、配合与落实", Excess: "一味承受"},
	'雷': {Name: "震", Symbol: "雷", Quality: "启动、回应与行动", Excess: "仓促反应"},
	'风': {Name: "巽", Symbol: "风", Quality: "顺势进入、沟通与持续影响", Excess: "反复摇摆"},
	'水': {Name: "坎", Symbol: "水", Quality: "识别风险、反复核实与守住原则", Excess: "被风险牵着走"},
	'火': {Name: "离", Symbol: "火", Quality: "辨明、呈现与有所依附", Excess: "只看表面"},
	'山': {Name: "艮", Symbol: "山", Quality: "停止、界限与稳定", Excess: "停滞僵化"},
	'泽': {Name: "兑", Symbol: "泽", Quality: "交流、协商与共同感受", Excess: "只求和气而回避问题"},
}

// hexagramThemes are neutral reading themes, not predictions or outcome labels.
var hexagramThemes = map[int]string{
	1:  "主动开创，同时保持自律",
	2:  "承载配合，把想法落实为长期行动",
	3:  "起步条件尚乱，先建立次序",
	4:  "经验不足时先求证、再行动",
	5:  "条件未齐，等待时继续准备",
	6:  "面对分歧，先厘清事实与边界",
	7:  "依靠组织、分工与共同纪律",
	8:  "建立联结前确认共同基础",
	9:  "用小步积累代替一次强推",
	10: "谨慎履行，尊重规则与分寸",
	11: "保持上下沟通与动态平衡",
	12: "沟通受阻时保存力量、调整路径",
	13: "公开沟通，在差异中寻找共同目标",
	14: "资源增加后更要妥善管理",
	15: "降低姿态，克制占有与居功",
	16: "行动前先完成准备，避免只凭兴奋",
	17: "顺应变化，同时保留自己的判断",
	18: "修复积累已久的问题，追查根因",
	19: "主动接近与支持，也为后续变化准备",
	20: "先观察全局，再检视自己的位置",
	21: "正面处理障碍，用清楚规则推进",
	22: "改善表达形式，但不让形式代替内容",
	23: "结构正在松动，先保护关键部分",
	24: "回到原点，从一个小步骤恢复",
	25: "依据事实行动，减少主观预设",
	26: "积蓄能力，在成熟前克制出手",
	27: "审视正在吸收、提供和重复的内容",
	28: "负担超过结构承受力，需要重新分配",
	29: "风险反复出现时守住原则、逐项核实",
	30: "看清依赖关系，不被表象带走",
	31: "留意相互影响，尊重真实回应",
	32: "用稳定节奏维持长期投入",
	33: "适时后退，为判断和转向保留余地",
	34: "力量增长时克制使用方式",
	35: "逐步推进，让成果获得看见与确认",
	36: "外部不利时保护判断与核心能力",
	37: "先整理内部秩序、角色和责任",
	38: "承认分歧存在，从可合作的小处开始",
	39: "正面路径受阻时绕行、求助或暂停",
	40: "解除已有压力，及时处理遗留事项",
	41: "主动减省，把资源集中到重点",
	42: "在互惠基础上增加投入与支持",
	43: "清楚表态，以公开方式完成决断",
	44: "面对突然出现的影响，先判断边界",
	45: "汇聚人员与资源前先建立共识",
	46: "依靠积累稳步上升，不急于跨级",
	47: "受限时节省资源，维持关键承诺",
	48: "维护共同依赖的基础与长期能力",
	49: "条件成熟后再推动结构性改变",
	50: "更新承载方式，把资源转化为新用途",
	51: "突发变化时先稳定，再决定回应",
	52: "在应当停止的位置守住界限",
	53: "循序渐进，让关系和能力逐步稳定",
	54: "先认清位置、条件与彼此期待",
	55: "信息资源充足时防止判断过载",
	56: "处于过渡环境时谨慎、自持并尊重边界",
	57: "柔和进入、持续沟通，并以明确边界避免反复摇摆",
	58: "通过真诚交流形成共同理解",
	59: "疏通隔阂，让分散的信息重新连接",
	60: "设定可执行的尺度，避免约束过度",
	61: "用内在一致与诚实建立信任",
	62: "小事可以推进，大动作需要克制",
	63: "看似完成后继续检查和维护",
	64: "尚未完成时整理条件、谨慎收尾",
}

func trigramsFor(hexagram Hexagram) (trigramProfile, trigramProfile, bool) {
	runes := []rune(hexagram.Name)
	if len(runes) < 3 {
		return trigramProfile{}, trigramProfile{}, false
	}

	if runes[1] == '为' {
		trigram, ok := trigramProfiles[runes[len(runes)-1]]
		return trigram, trigram, ok
	}

	upper, upperOK := trigramProfiles[runes[0]]
	lower, lowerOK := trigramProfiles[runes[1]]
	return upper, lower, upperOK && lowerOK
}

func themeFor(hexagram Hexagram) string {
	if theme := hexagramThemes[hexagram.Number]; theme != "" {
		return theme
	}
	return "从事实、条件与下一步三个层面重新审视问题"
}
