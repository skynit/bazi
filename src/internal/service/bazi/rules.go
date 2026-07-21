package bazi

import (
	_ "embed"
	"encoding/json"
	"reflect"
	"strings"
	"sync"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

const (
	// RuleVersion identifies the deterministic rule tables and scoring
	// weights used by the BaZi and fortune APIs.
	RuleVersion = "bazi-rules-2026-07-17.27"
	// RuleSchool names the primary school of interpretation for the current
	// deterministic rule set.
	RuleSchool = "子平八字-扶抑调候-v2"
)

type RuleMeta = model.RuleMeta
type RuleTableMeta = model.RuleTableMeta
type RuleSourceMeta = model.RuleSourceMeta
type BodyStrengthRuleConfig = model.BodyStrengthRuleConfig
type BodyStrengthWeights = model.BodyStrengthWeights
type BodyStrengthNormalizers = model.BodyStrengthNormalizers
type BodyStrengthAdjustmentThresholds = model.BodyStrengthAdjustmentThresholds
type YueLingRuleConfig = model.YueLingRuleConfig
type YueLingScoreState = model.YueLingScoreState
type BodyStrengthRootRuleConfig = model.BodyStrengthRootRuleConfig
type BodyStrengthHideStemWeights = model.BodyStrengthHideStemWeights
type BodyStrengthTerrainWeights = model.BodyStrengthTerrainWeights
type BodyStrengthBonusRuleConfig = model.BodyStrengthBonusRuleConfig
type BodyStrengthBonusScores = model.BodyStrengthBonusScores
type BodyStrengthInfluenceRuleConfig = model.BodyStrengthInfluenceRuleConfig
type BodyStrengthAdjustmentForceConfig = model.BodyStrengthAdjustmentForceConfig

//go:embed rules/rule_meta.json
var ruleMetaJSON []byte

var (
	ruleMetaOnce sync.Once
	ruleMeta     RuleMeta
)

// DefaultRuleMeta returns the public rule manifest for API responses.
func DefaultRuleMeta() RuleMeta {
	ruleMetaOnce.Do(func() {
		if err := json.Unmarshal(ruleMetaJSON, &ruleMeta); err != nil {
			ruleMeta = fallbackRuleMeta()
		}
		applyRuleMetaDefaults(&ruleMeta)
		applyRuleTableCounts(&ruleMeta)
		applyRuleTableSources(&ruleMeta)
	})
	return cloneRuleMeta(ruleMeta)
}

// ValidRuleMeta accepts historical version labels while requiring every table,
// source fingerprint, and scoring parameter to match the authoritative manifest.
func ValidRuleMeta(meta RuleMeta) bool {
	if meta.RuleVersion == "" || meta.School == "" {
		return false
	}
	want := DefaultRuleMeta()
	want.RuleVersion = meta.RuleVersion
	want.School = meta.School
	return reflect.DeepEqual(meta, want)
}

func cloneRuleMeta(meta RuleMeta) RuleMeta {
	if len(meta.Tables) > 0 {
		meta.Tables = append([]RuleTableMeta(nil), meta.Tables...)
		for i := range meta.Tables {
			if len(meta.Tables[i].Sources) == 0 {
				continue
			}
			meta.Tables[i].Sources = append([]RuleSourceMeta(nil), meta.Tables[i].Sources...)
			for j := range meta.Tables[i].Sources {
				meta.Tables[i].Sources[j].Files = cloneStringMap(meta.Tables[i].Sources[j].Files)
			}
		}
	}
	return meta
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func defaultBodyStrengthRuleConfig() BodyStrengthRuleConfig {
	return BodyStrengthRuleConfig{
		Weights: BodyStrengthWeights{
			Ling:  0.40,
			Di:    0.30,
			Shi:   0.20,
			Sheng: 0.10,
			Bonus: 1.00,
		},
		Normalizers: BodyStrengthNormalizers{
			Ling:                3.0,
			Di:                  7.0,
			ShiSigmoidDivisor:   1.5,
			ShengSigmoidDivisor: 1.5,
			ShiFormula:          bodyStrengthSignedNormalizationFormula,
			ShengFormula:        bodyStrengthSupportNormalizationFormula,
		},
		AdjustmentThresholds: BodyStrengthAdjustmentThresholds{
			ShiLingSupportForce: 5.0,
			ShiLingBlendSelf:    0.75,
			ShiLingBlendNeutral: 0.25,
		},
		YueLing:         defaultYueLingRuleConfig(),
		Root:            defaultBodyStrengthRootRuleConfig(),
		Bonus:           defaultBodyStrengthBonusRuleConfig(),
		Influence:       defaultBodyStrengthInfluenceRuleConfig(),
		AdjustmentForce: defaultBodyStrengthAdjustmentForceConfig(),
	}
}

func defaultYueLingRuleConfig() YueLingRuleConfig {
	return YueLingRuleConfig{
		RuleID:           yueLingRuleID,
		Profile:          yueLingProfile,
		HashBasis:        yueLingHashBasis,
		DayElementOrder:  yueLingDayElementOrder,
		MonthBranchOrder: yueLingMonthBranchOrder,
		Scores:           yueLingMatrix,
		ScoreStates: [5]YueLingScoreState{
			{State: "旺", Score: 3},
			{State: "相", Score: 2},
			{State: "休", Score: 1},
			{State: "囚", Score: 0.5},
			{State: "死", Score: 0},
		},
		TableSHA256:      verifiedYueLingMatrixSHA256(),
		EarthMonthPolicy: "simplified_earth_prosperous_in_chou_chen_wei_xu",
		ValidationStatus: "engineering_complete_expert_gold_pending",
	}
}

func defaultBodyStrengthRootRuleConfig() BodyStrengthRootRuleConfig {
	return BodyStrengthRootRuleConfig{
		RuleID:  bodyStrengthRootRuleID,
		Profile: bodyStrengthRootProfile,
		HideStemWeights: BodyStrengthHideStemWeights{
			Main: 0.6, Middle: 0.3, Residual: 0.1,
		},
		TerrainWeights: BodyStrengthTerrainWeights{
			ChangSheng: 1.5, MuYu: 1, GuanDai: 1, LinGuan: 1.5,
			DiWang: 1.5, Shuai: 1, Bing: 0.5, Si: 0.5,
			Mu: 1, Jue: 0, Tai: 0.5, Yang: 0.5,
		},
		RootMultiplier:   1.5,
		TouGanMultiplier: 1.2,
		TouGanScope:      "all_four_heaven_stems_including_day_master",
		ValidationStatus: "not_validated",
	}
}

func defaultBodyStrengthBonusRuleConfig() BodyStrengthBonusRuleConfig {
	dayStemOrder, luBranches := canonicalLuProfile()
	config := BodyStrengthBonusRuleConfig{
		RuleID:           bodyStrengthBonusRuleID,
		Profile:          bodyStrengthBonusProfile,
		HashBasis:        bodyStrengthBonusHashBasis,
		DayStemOrder:     dayStemOrder,
		LuBranches:       luBranches,
		YangRenStemOrder: bodyStrengthYangRenStemOrder,
		YangRenBranches:  bodyStrengthYangRenBranches,
		Scores: BodyStrengthBonusScores{
			DayLu: 0.08, MonthLu: 0.06, DayYangRen: 0.07, MonthYangRen: 0.05,
		},
		YinStemBladePolicy: "no_yang_ren_bonus",
		ValidationStatus:   "not_validated",
	}
	config.TableSHA256 = verifiedBodyStrengthBonusSHA256(config)
	return config
}

func defaultBodyStrengthInfluenceRuleConfig() BodyStrengthInfluenceRuleConfig {
	return BodyStrengthInfluenceRuleConfig{
		RuleID:                     bodyStrengthInfluenceRuleID,
		Profile:                    bodyStrengthInfluenceProfile,
		VisibleStemScope:           "year_month_hour_heaven_stems_excluding_day_master",
		SamePolarityPeerWeight:     1.0,
		OppositePolarityPeerWeight: 0.8,
		OfficerKillerWeight:        1.2,
		OutputWeight:               0.8,
		WealthWeight:               0.6,
		HiddenBranchScope:          "all_four_branches_restrict_only",
		HiddenBranchMultiplier:     1.5,
		SameElementRootOwnership:   "di",
		SealOwnership:              "sheng",
		ValidationStatus:           "not_validated",
	}
}

func defaultBodyStrengthAdjustmentForceConfig() BodyStrengthAdjustmentForceConfig {
	return BodyStrengthAdjustmentForceConfig{
		RuleID:                 bodyStrengthAdjustmentForceRuleID,
		Profile:                bodyStrengthAdjustmentForceProfile,
		StemForce:              5.0,
		HiddenStemMultiplier:   3.0,
		HiddenStemWeightSource: "root.hide_stem_weights",
		ShiLingSupportBasis:    "other_visible_stems_and_all_hidden_stems_seal_or_same_element_excluding_day_master",
		NeutralTarget:          0.5,
		ValidationStatus:       "not_validated",
	}
}

func bodyStrengthRuleConfig() BodyStrengthRuleConfig {
	return DefaultRuleMeta().BodyStrength
}

func applyRuleMetaDefaults(meta *RuleMeta) {
	if meta.RuleVersion == "" {
		meta.RuleVersion = RuleVersion
	}
	if meta.School == "" {
		meta.School = RuleSchool
	}
	if meta.BodyStrength.Weights.Ling == 0 {
		meta.BodyStrength = defaultBodyStrengthRuleConfig()
		return
	}
	if meta.BodyStrength.Weights.Bonus == 0 {
		meta.BodyStrength.Weights.Bonus = 1.0
	}
	if meta.BodyStrength.Normalizers.Ling == 0 {
		meta.BodyStrength.Normalizers.Ling = 3.0
	}
	if meta.BodyStrength.Normalizers.Di == 0 {
		meta.BodyStrength.Normalizers.Di = 7.0
	}
	if meta.BodyStrength.Normalizers.ShiSigmoidDivisor == 0 {
		meta.BodyStrength.Normalizers.ShiSigmoidDivisor = 1.5
	}
	if meta.BodyStrength.Normalizers.ShengSigmoidDivisor == 0 {
		meta.BodyStrength.Normalizers.ShengSigmoidDivisor = 1.5
	}
	if meta.BodyStrength.Normalizers.ShiFormula == "" {
		meta.BodyStrength.Normalizers.ShiFormula = bodyStrengthSignedNormalizationFormula
	}
	if meta.BodyStrength.Normalizers.ShengFormula == "" {
		meta.BodyStrength.Normalizers.ShengFormula = bodyStrengthSupportNormalizationFormula
	}
	if meta.BodyStrength.AdjustmentThresholds.ShiLingSupportForce == 0 {
		meta.BodyStrength.AdjustmentThresholds = defaultBodyStrengthRuleConfig().AdjustmentThresholds
	}
	if meta.BodyStrength.YueLing.RuleID == "" {
		meta.BodyStrength.YueLing = defaultYueLingRuleConfig()
	}
	if meta.BodyStrength.Root.RuleID == "" {
		meta.BodyStrength.Root = defaultBodyStrengthRootRuleConfig()
	}
	if meta.BodyStrength.Bonus.RuleID == "" {
		meta.BodyStrength.Bonus = defaultBodyStrengthBonusRuleConfig()
	}
	if meta.BodyStrength.Influence.RuleID == "" {
		meta.BodyStrength.Influence = defaultBodyStrengthInfluenceRuleConfig()
	}
	if meta.BodyStrength.AdjustmentForce.RuleID == "" {
		meta.BodyStrength.AdjustmentForce = defaultBodyStrengthAdjustmentForceConfig()
	}
	for i := range meta.Tables {
		if meta.Tables[i].Key == "shensha" {
			meta.Tables[i].Version = "2026-07-16.59"
			meta.Tables[i].Description = strings.ReplaceAll(meta.Tables[i].Description,
				"旧天火煞因原文要求五位且水神口径未裁决，而正式入口只有四柱，现停止发布；月令查表的天火是独立名称。",
				"旧天火煞因原文要求五位且水神口径未裁决，而正式入口只有四柱，现停止发布；错误月令天火单支表同步停止发布。")
			meta.Tables[i].Source += "；《渊海子平》PDF第107页截路空亡、《三命通会》PDF第113页戊癸异表"
			meta.Tables[i].Source += "；《三命通会》PDF第103页（书内第100页）月厌月煞原页影像"
			meta.Tables[i].Source += "；《渊海子平》PDF第119页十干羊刃飞刃异表、PDF第205页五阳Profile；《三命通会》PDF第108页及第226页五阳Profile"
			meta.Tables[i].Source += "；《渊海子平》PDF第105页六甲空亡双支表；《三命通会》PDF第108-110页双支定义及阴阳单支异表"
			meta.Tables[i].Source += "；《渊海子平》PDF第81页、《三命通会》PDF第85-86页十干禄表"
			meta.Tables[i].Description += " 神煞层停止重复发布建禄：日干禄位命中月支已由禄神逐柱规则完整表达，《三命通会》月令建禄属于取格语境，继续由独立格局和身强证据层处理；旧神煞名称建禄保持unregistered/not_available。"
			meta.Tables[i].Description += " 旧月空把月柱六甲旬空的两个地支冒名发布；《三命通会》同名月空是月支三合组查单个目标天干的不同表，名称、输出类型和公式均不一致，现删除旧旁路并保持unregistered/not_available；候选表的原字与落柱口径待独立Profile裁决。"
			meta.Tables[i].Description += " 截路空亡当前采用《渊海子平》日干查时支Profile：甲己申酉、乙庚午未、丙辛辰巳、丁壬寅卯、戊癸子丑，只落时柱；《三命通会》前四组相同而戊癸作戌亥，异表不混入。"
			meta.Tables[i].Description += " 五份固定本地PDF均未定位童子煞名称或旧月支四季查时支公式；旧实现又缺少常见候选口诀的纳音五行分支且夏季目标不一致，不能拼接补表，现删除并保持unregistered/not_available。"
			meta.Tables[i].Description += " 月厌按十二月逆行支表锁定；月煞按《三命通会》四组三合表修正，旧表仅子月目标偶合原文、其余十一月错误。两项当前均以月支为主键逐柱匹配目标支，只记录结构。"
			meta.Tables[i].Description += " 旧月刑、月害分别只是月支与目标支的相刑、六害关系别名，五份固定资料未定位为独立神煞表；现停止神煞发布并保持unregistered/not_available，底层结构统一由地支关系图输出相刑、三刑和六害。"
			meta.Tables[i].Description += " 旧天刑、天火、天贼、大时、兵禁、天吏六张月支单支表均无同名封闭表依据；天刑和天火还分别与已定位的天刑煞、天火煞在主键、目标类型及复合条件上冲突，六项停止发布并保持unregistered/not_available。"
			meta.Tables[i].Description += " 旧致死月支十二表在五份固定资料中无同名公式，且全部命中原本被高风险门禁无条件丢弃，属于不可观察死规则；现删除表成员与抑制名单残留并保持unregistered/not_available。"
			meta.Tables[i].Description += " 高风险抑制注册表删除三个无生产调用的退役名称残留，只保留有固定结构表和来源合同但禁止公开输出的死符；退役名称保持unregistered/not_available且不进入公开结果。"
			meta.Tables[i].Description += " 羊刃与飞刃当前采用子平五阳干Profile：甲卯酉、丙午子、戊午子、庚酉卯、壬子午，只以日干为主键逐柱匹配前后两个目标；《渊海子平》PDF第119页另列十干异表，当前不混入乙丁己辛癸五阴目标。"
			meta.Tables[i].Description += " 空亡当前采用日柱六甲旬双支Profile：甲子戌亥、甲戌申酉、甲申午未、甲午辰巳、甲辰寅卯、甲寅子丑，逐柱匹配年、月、时支；《三命通会》另载按阴阳日干只取一支的轻重异表，当前不缩减为单支。"
			meta.Tables[i].Description += " 十干禄位收敛为神煞、格局与身强共同消费的唯一Profile：甲寅、乙卯、丙戊巳、丁己午、庚申、辛酉、壬亥、癸子；神煞只以日干为主键逐柱匹配，日干综合表不再重复存储禄神目标。"
		}
		if meta.Tables[i].Key == "pattern_candidates" {
			meta.Tables[i].Version = "2026-07-17.27"
			meta.Tables[i].Source = strings.ReplaceAll(meta.Tables[i].Source, "从化专旺、特殊", "特殊")
			meta.Tables[i].Source = strings.ReplaceAll(meta.Tables[i].Source, "从化、特殊", "特殊")
			meta.Tables[i].Description = strings.ReplaceAll(meta.Tables[i].Description, "四者都不自动成为主格。", "四者只作为辅助特征 category，不参与结构格局裁决。")
			meta.Tables[i].Description = strings.ReplaceAll(meta.Tables[i].Description, "三奇与金神同样只作为辅助结构，不自动成为主格。", "三奇与金神同样只作为辅助特征 category，不参与结构格局裁决。")
			if !strings.Contains(meta.Tables[i].Source, "《滴天髓阐微》PDF第43页") {
				meta.Tables[i].Source += "；《滴天髓阐微》PDF第43页两气双清、生克十局各半且不可夹杂"
			}
			if !strings.Contains(meta.Tables[i].Source, "《滴天髓阐微》PDF第44-45页") {
				meta.Tables[i].Source += "；《滴天髓阐微》PDF第44-45页独象方局全、四库全且不杂克神"
			}
			if !strings.Contains(meta.Tables[i].Source, "《滴天髓阐微》PDF第173-174页") {
				meta.Tables[i].Source += "；《滴天髓阐微》PDF第173-174页从象、PDF第186-187页顺局；《三命通会》PDF第205页弃命从财；《渊海子平》PDF第566-569页弃命从杀论及从财从杀诗诀"
			}
			if !strings.Contains(meta.Tables[i].Source, "《滴天髓阐微》PDF第177-178页化象") {
				meta.Tables[i].Source += "；《滴天髓阐微》PDF第177-178页化象；《三命通会》PDF第72-74页论十干化气；《渊海子平》PDF第575-576页化气诗诀"
			}
			for _, contract := range []struct {
				fragment  string
				statement string
			}{
				{fragment: "透干、身强、扶助、制化和破格条件并不相同", statement: "固定各格专章要求的透干、身强、扶助、制化和破格条件并不相同。"},
				{fragment: "恢复八格前须逐格建立条件 Profile 和专家 Gold", statement: "恢复八格前须逐格建立条件 Profile 和专家 Gold。"},
				{fragment: "旧两神成像检测器仅以聚合分数两项达到15%", statement: "旧两神成像检测器仅以聚合分数两项达到15%且差值不超过总分四分之一就声称全局仅两种五行，允许其余三项合计大量夹杂，名称也不合原文。"},
				{fragment: "四干与四支本气八个位点", statement: "当前两气成象格只按四干与四支本气八个位点，要求恰好两种五行且各占四位，删除旧15%聚合分数阈值和通关喜用推断。"},
				{fragment: "藏干杂气、月令旺衰、顺逆取用和行运破局仍未裁决", statement: "藏干杂气、月令旺衰、顺逆取用和行运破局仍未裁决。"},
				{fragment: "曲直、炎上、从革、润下要求地支完整成方或三合局", statement: "曲直、炎上、从革、润下要求地支完整成方或三合局，稼穑要求辰戌丑未四库皆全。"},
				{fragment: "删除旧60%生扶、30%日主和10%克神分数阈值", statement: "删除旧60%生扶、30%日主和10%克神分数阈值及统一喜忌。"},
				{fragment: "转化成局、藏干杂气、得时旺衰、引通取用和行运破局仍未裁决", statement: "转化成局、藏干杂气、得时旺衰、引通取用和行运破局仍未裁决。"},
				{fragment: "旧从财、从势、从杀、从弱、从儿五个检测器", statement: "旧从财、从势、从杀、从弱、从儿五个检测器以10%/15%生扶和60%主势等本地分数阈值发布候选；从财、从杀只要当令或透干即命中，不满足满局专从，从弱没有独立可执行古籍Profile，从势漏掉食伤并旺、不能专从一神等原文条件，从儿错误要求身弱无根并允许印官杂入，与原文明言不论身强弱、比劫仍可生食伤、食伤在提纲、又见财且忌印官相反。现删除五项注册、算法、互斥争议和统一喜忌；十神、藏干与五行事实继续保留。恢复前须分别建立成破条件Profile和专家Gold，证据等级固定为classical_text_local / text_located_not_expert_gold。"},
				{fragment: "旧 checkHuaQiGe 与 checkCongHuaGe", statement: "旧 checkHuaQiGe 与 checkCongHuaGe 判断同一日干五合，却以无固定来源的30%与25%日主分数阈值区分化气与从化，低于25%时同盘重复发布化气格与从化格。两者只检查月干或时干，遗漏古籍允许的年干配合；把五行三合支直接当月令快捷表，未完整处理各组次月、辰时、妒合、得辰、旺衰虚实及逐组克破条件。现删除两项注册、算法、快捷月令表和统一喜忌；天干五合结构与unadjudicated成化证据继续保留。恢复前须先选择固定化象Profile并建立真假化、柱位、得令得时、辰引、妒合克破和专家Gold；证据等级固定为classical_text_local / text_located_not_expert_gold。"},
				{fragment: "旧 pattern-candidate-set-v3", statement: "旧 pattern-candidate-set-v3 在所有比例检测器退役后仍强制校验并快照五行分数和身强分段，但10个存量检测器均不消费这两项，导致相同四柱因无关分数缺键、负值或分段标签变化而错误进入invalid_input。pattern-candidate-set-v4移除分数与分段门禁及快照字段，权威输入只保留四柱与月支；函数签名与持久化验证只接收四柱和月支。"},
				{fragment: "旧 pattern-candidate-set-v4 保留互斥争议状态机", statement: "旧 pattern-candidate-set-v4 保留互斥争议状态机，但10个注册器的exclusiveGroup全部为空，patternStatusDisputed与复合格局类别均不可达，has_dispute、dispute_reasons、候选status和dispute_reason只能固定为空或candidate。pattern-candidate-set-v5删除死注册参数、汇总算法和四个公开字段，并同步解释层与前端消费者；天干与地支关系图的真实disputed、ConflictsWith和DisputeReasons合同不受影响。"},
				{fragment: "旧 pattern-candidate-set-v5 用未裁决的本地整数优先级", statement: "旧 pattern-candidate-set-v5 用未裁决的本地整数优先级给候选分配主格、兼格和显示首项，但这些数值既不是古籍封闭表，也未由训练数据或专家Gold产生；primary_candidate_id、selection_basis、候选role和priority会把稳定展示顺序误读为唯一格局裁决。pattern-candidate-set-v6删除伪主格排序及相关公开字段，按规则ID与名称稳定排序只用于确定性序列化；结构格局与辅助特征继续由category区分，全部命中仍完整保留。"},
				{fragment: "旧 pattern-candidate-set-v6 的私有patternDetection", statement: "旧 pattern-candidate-set-v6 的私有patternDetection仍携带无人消费的Description、SubType、FavorableElements和UnfavorableElements；正式候选汇总只读取名称与类型，其中专禄残留统一喜忌还会为未来误接入留下伪解释入口。pattern-candidate-set-v7只保留PatternName和PatternType，删除解释、喜忌、两气关系文本helper和专旺子结构展示名；封闭输入真值、正式规则来源和全部候选命中保持不变。"},
				{fragment: "旧 pattern-candidate-set-v7 的candidate_id始终等于rule_id", statement: "旧 pattern-candidate-set-v7 的candidate_id始终等于rule_id，没有独立生成、选择或引用语义；双身份字段只增加序列化、缓存和篡改漂移面。pattern-candidate-set-v8删除重复candidate_id，以非空且唯一的rule_id作为唯一候选身份；出生时间不确定性候选的哈希candidate_id属于独立选择合同，不受影响。"},
				{fragment: "旧 pattern-candidate-set-v8 在每个候选重复固定状态", statement: "旧 pattern-candidate-set-v8 在每个候选重复固定状态not_validated/not_adjudicated，与集合级validation_status和interpretation_status完全相同，且没有任何逐候选裁决分支；双层状态允许历史快照形成互相矛盾的语义。pattern-candidate-set-v9删除候选级重复状态，以集合级状态作为唯一裁决边界；候选只保留规则身份、名称、类型、类别、来源和命中依据。"},
				{fragment: "旧 pattern-candidate-set-v9 的basis固定为local_detector_conditions_matched", statement: "旧 pattern-candidate-set-v9 的basis固定为local_detector_conditions_matched，没有任何其他取值；候选进入集合本身已经证明检测条件命中，重复字段只增加缓存漂移面。pattern-candidate-set-v10删除同义basis字段，rule_id和source继续提供可审计身份与来源；其他领域具有独立取值语义的basis不受影响。"},
				{fragment: "旧 pattern-candidate-set-v10 同时发布pattern_type与category", statement: "旧 pattern-candidate-set-v10 同时发布pattern_type与category两套分类：魁罡和日德被注册为辅助特征却标成特殊格局，结构角色与传统类型在同一候选自相矛盾，且没有专家Gold裁决这套正格/特殊格局分类。pattern-candidate-set-v11删除冲突pattern_type，以category作为唯一候选分类；结构格局与辅助特征由注册器明确赋值。"},
				{fragment: "旧金神规则ID pattern.special.jinshen", statement: "旧金神规则ID pattern.special.jinshen 的special前缀表示结构格局，category却是辅助特征，导致稳定身份与唯一分类冲突。现更名为pattern.aux.jinshen，并以正式入口矩阵证明10个注册规则ID前缀与category一致；旧ID不再发布。"},
				{fragment: "旧实现手写patternDetectorCount=10", statement: "旧实现手写patternDetectorCount=10，规则身份、来源、分类和调用分散在10次add，新增或删除检测器时容易形成数量、元数据与实际执行路径漂移。pattern-candidate-set-v13建立单一patternDetectorRegistry，检测器定义和执行统一遍历，detector_count由注册表长度派生。"},
				{fragment: "旧patternDetectorRegistry是包级可变var", statement: "旧patternDetectorRegistry是包级可变var，同包代码可永久替换规则身份、来源、分类或调用函数，而初始化时捕获的计数仍保持不变，可能让同进程后续结果随调用顺序漂移。pattern-candidate-set-v14改为每次返回独立注册表快照；分析入口只创建一次快照，执行和detector_count绑定同一局部快照，外部修改不能污染后续调用。"},
				{fragment: "旧detector_profile只是人工版本标签", statement: "旧detector_profile只是人工版本标签，未携带注册规则身份、来源、分类和实现映射的机器摘要；清单内容改变而忘记升级版本时，历史响应无法识别漂移。pattern-candidate-set-v15新增detector_manifest_sha256，对按rule_id规范排序的规则身份、来源、category和implementation结构化清单计算SHA-256；合法与非法结果都从单次执行快照派生同一摘要，持久化重算逐字段验证。"},
				{fragment: "旧专旺检测器读取包级可变zhuanWangProfiles", statement: "旧专旺检测器读取包级可变zhuanWangProfiles map，建禄与专禄读取包级可变luShenZhi map，身强层又在启动时复制luDayStemOrder和luBranchOrder；同一进程修改后会让格局、神煞和身强消费不同真值，且实现名摘要不变。pattern-candidate-set-v16删除这些可变全局：专旺Profile每次返回独立嵌套快照，禄位以canonicalLuProfile和luBranchForStem纯函数供格局、神煞与身强共同消费；三个检测器的implementation标识升级并把detector_manifest_sha256更新为dc08ac014295b5505a9e09d963c710644593845629d0112e9a572f24905b28d8。"},
				{fragment: "旧清单摘要只散列implementation版本名", statement: "旧清单摘要只散列implementation版本名，不散列检测器实际使用的封闭表、柱位范围和算法参数；规则值改变但实现名不变时仍无法识别漂移。pattern-candidate-set-v17集中10个运行时语义Profile，为每条注册项计算profile_sha256并纳入规范清单；专旺、禄位、羊刃、两气、魁罡、金神、三奇与日德实现共同消费这些纯Profile。当前detector_manifest_sha256为ee21f7c8438031bd64f284f9691d16934d53930dacfa3661bafc5874e3fe4a8f，逐规则摘要由合同分别锁定。"},
				{fragment: "旧两气语义Profile记录了五行顺序和计数参数", statement: "旧两气语义Profile记录了五行顺序和计数参数，但运行时仍在checkLiangQiChengXiang局部重复木火土金水顺序、四柱、两种五行和各四位常量；只改局部实现时profile_sha256不会变化。pattern-candidate-set-v18新增liangQiSemanticProfile独立快照，摘要与运行时共同消费柱数、五行顺序、种类数和每类位点数；两气逐规则摘要更新为49d0edcdc94b96ef1d351b44d653541ad964942aefdef33b9b37dde8ee254c07，总detector_manifest_sha256更新为3b89e62e6fe12baf7969be4a2afb35b75308b0029d55ee16c5a5d273b2c49636。命中语义、古籍来源和证据等级不变。"},
				{fragment: "旧最终汇总器以rule_id与PatternName拼接复合去重键", statement: "旧最终汇总器以rule_id与PatternName拼接复合去重键，并在rule_id相同时按名称次级排序；单次遍历唯一规则注册表不会产生完全重复项，而同一rule_id配不同名称时复合键反而保留两个冲突身份。pattern-candidate-set-v19删除不可达的静默去重和名称次级排序，候选只按唯一rule_id确定序列化顺序并直接映射；注册表合同继续失败检查空值和重复rule_id。正式候选集合、检测器清单与detector_manifest_sha256均不变。"},
				{fragment: "旧专旺与两气检测器直接读取data.GanElement和data.ZhiElement", statement: "旧专旺与两气检测器直接读取data.GanElement和data.ZhiElement包级可变map，十干和十二支五行映射既未进入语义Profile摘要，也可被同进程其他代码污染而改变格局命中。pattern-candidate-set-v20建立十干10项、十二支12项纯值Profile，专旺与两气的运行时和语义摘要共同消费映射及四柱数；未知符号失败关闭。专旺profile_sha256更新为d4e2a3250ea362c239982cb9c5ea6ccc62b69ba0e0f3198a551e277b6b0e8073，两气更新为03514ea7676c03bcc79bff1616f73552cbc07cfa7fdd5121eebd8f1e26db4543，总detector_manifest_sha256更新为32ffa7b00d2145bfe737b8bf0a9135f4feb10d3fc03963ca83dfc00f05cad938。合法输入命中与古籍边界不变。"},
				{fragment: "旧三奇语义Profile记录窗口大小和两个位置", statement: "旧三奇语义Profile记录窗口大小和两个位置，但classicalSanQiSequence仍硬编码i+2与i:i+3并接受任意长度滑窗；旧金神Profile只写hour_pillar，checkJinShenHour另行硬编码四柱与索引3。pattern-candidate-set-v21建立sanQiSemanticProfile和jinShenSemanticProfile，神煞与格局共同消费四柱数、窗口大小、窗口起点、顺序表、时柱索引和三时表；非四柱输入及越界Profile失败关闭。三奇profile_sha256更新为79627a02c955fc510bfe283954991a51a83b7e04653f11aea6304af10941125b，金神更新为004184a1c8e70f481240225d637eba2d1ce6d0ccb09f9db1e05a1199f186909a，总detector_manifest_sha256更新为881193d5290ed58d831dfdde6b57b626380b193d465c44a2d545d59b93b023f7。正式四柱命中与未裁决边界不变。"},
				{fragment: "旧10条逐规则摘要未绑定共享四柱位置上下文", statement: "旧10条逐规则摘要未绑定共享四柱位置上下文，入口在摘要之外硬编码pillars[1]月支、pillars[2]日干日支和月支一致性；柱位提取改变时规则profile_sha256仍可不变。pattern-candidate-set-v22建立patternPillarContextSemanticProfile，固定四柱数、年/月/日/时索引0/1/2/3及声明月支必须等于月柱支；输入验证、检测上下文和专旺日主提取共同消费，并把同一上下文封装进全部10条逐规则摘要。索引越界、重复、错误柱数或策略均失败关闭，总detector_manifest_sha256更新为ec2928c92f0b6e227e9f656a788fa641215a3587a0309876d7f84cee6aa08ab6。合法输入候选与古籍边界不变。"},
				{fragment: "旧多数检测器把PatternName作为局部字面量返回", statement: "旧多数检测器把PatternName作为局部字面量返回，逐规则语义Profile和总清单都未绑定公开候选名称；名称改变时rule_id、profile_sha256和detector_manifest_sha256仍可保持不变。pattern-candidate-set-v23建立patternDetectorOutputNames纯Profile：专旺固定曲直、炎上、稼穑、从革、润下五名，其余9条各固定唯一名称；检测器构造、注册器允许集合、逐规则envelope和总清单共同消费。空名称、未知规则或检测返回不在允许集合时失败关闭，生产源码不再保留PatternName字面量。全部10条逐规则摘要重新固定，总detector_manifest_sha256更新为acd631f529e51ead2c50fa1c7832149ad7d994d7137be348133ecd70de2cff1a。候选名称和值不变。"},
				{fragment: "旧响应只公开不透明的detector_manifest_sha256总摘要", statement: "旧响应只公开不透明的detector_manifest_sha256总摘要；历史结果发生变化时，无法仅凭响应定位10条规则中哪条Profile改变，未命中检测器也没有逐规则摘要证据。pattern-candidate-set-v24新增规范排序的detector_profiles，只发布rule_id与profile_sha256；合法与非法结果都从同一次注册表快照生成10项清单，持久化重算逐字段校验并拒绝篡改。清单不重复来源、分类、implementation或输出名称，总detector_manifest_sha256保持acd631f529e51ead2c50fa1c7832149ad7d994d7137be348133ecd70de2cff1a。候选与检测器语义不变。"},
				{fragment: "旧逐规则Profile仍以人工implementation字符串代表算法实现", statement: "旧逐规则Profile仍以人工implementation字符串代表算法实现；检测根函数或同包helper的布尔组合、遍历与失败关闭逻辑变化时，只要人工标签未更新，摘要就无法识别漂移。pattern-candidate-set-v25删除注册表人工implementation字段，建立go_ast_detector_closure_v1算法Profile：每条规则固定根函数、规范排序的同包调用闭包与规范化Go AST SHA-256，fixedPatternDetection作为已由输出名Profile独立约束的边界；构建期合同从生产源码重算闭包和摘要，运行时不读取源码。10条逐规则摘要全部更新，总detector_manifest_sha256更新为bbb80d8b291e81264f6894933b422c45ad40940138d49e58bcd8ad8b98d1048f。候选命中、封闭表、古籍来源和未裁决边界不变。"},
				{fragment: "专旺与两气复杂检测器缺少可序列化的单原子输入变异见证", statement: "专旺与两气复杂检测器缺少可序列化的单原子输入变异见证，现有正反例不能机器证明每个拒绝条件只由一个输入差异触发。pattern-candidate-set-v26在两条语义envelope中新增behavior_witnesses，固定专旺五格基线成立盘、两气基线成立盘与单字段或单柱删除变异；合同验证每个负例只距基线一个输入原子并覆盖结构缺失、天干克神、结构外克神、未知干支、柱数、第三气及非四四均分。专旺profile_sha256更新为1a9de9f4e8ba6fb1907df1e9b31f80d00510f05b4e678a8644becba90c5aed55，两气更新为42e00bf6dedb3018a3f14bd62af5e89716c1e9b7efe030b13d97ea06357116b4，总detector_manifest_sha256更新为a151430bffe88b3a9e9b9f27b872bfbf231970280529cdda7ae88791a7a336ec；其余8条逐规则摘要和10条算法AST摘要不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "专旺检测仍缺少地支多重集与柱位排列的形变合同", statement: "专旺检测仍缺少地支多重集与柱位排列的形变合同，固定正反例不能证明同一组方局在换柱、重复必需支或加入任一结构外支后保持一致。pattern-candidate-set-v27新增metamorphic_policies，固定排列保持、必需支重复保持、结构外非克神保持、结构外克神拒绝及缺失必需支拒绝五种关系；合同对9组方局、三合局与四库结构生成全部合法六十甲子排列，验证1632个保持命中与552个拒绝盘，覆盖重复必需支、缺失必需支和全部结构外地支。专旺profile_sha256更新为daedf9a821d62349eadcaf07757cc678f86ff67f3adb76e966d3d8d3685cd57c，总detector_manifest_sha256更新为4d9bf34f68fdc078b70fc22fa1ed67d2270d78526b46759b3adec6307eaf2130；其余9条逐规则摘要、29项单原子见证和10条算法AST摘要不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "两气成象仍缺少十种无序五行对与八个位点的对称形变合同", statement: "两气成象仍缺少十种无序五行对与八个位点的对称形变合同，木火固定盘不能证明算法对其他五行对和干支位置完全对称。pattern-candidate-set-v28为两气新增四项独立metamorphic_policies，穷举10种无序五行对、每对70种四选四位置组合，使用各五行阳干阳支构造合法六十甲子，验证700个保持命中与1400个拒绝盘；每个成立盘分别引入第三气与5:3非均分并必须拒绝。两气profile_sha256更新为20ed815f5411b2add63e4d98931f8f775b7af1b4f98a2b41f91a6b9feae9c610，总detector_manifest_sha256更新为a32a72454b8dc2d4b9d12430dc65e2c54f6114ee40eab2350585f787ec6ecc28；其余9条逐规则摘要、专旺五项形变策略、29项单原子见证和10条算法AST摘要不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "八个简单表型检测器的穷举结果仍分散在测试中", statement: "八个简单表型检测器的穷举结果仍分散在测试中，逐规则摘要虽绑定表值却不能直接标识完整输入到输出真值表。pattern-candidate-set-v29新增canonical_truth_table_v1 behavior_manifest，固定输入域、案例数、命中数与行为清单SHA-256：建禄与月刃各120例，专禄、日刃、魁罡、日德四个日柱表及金神各60例，三奇10000例。构建期合同通过注册表实际检测函数重算全部10540行并验证8条摘要互不重复。8条简单规则profile_sha256全部更新，总detector_manifest_sha256更新为e4d5d633ab5139f0f2fb51b28639154c8e229e5e1ccb456223f95720fb5d8619；专旺与两气摘要、两组形变策略、29项单原子见证及10条算法AST摘要不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "公开detector_profiles仍只有profile_sha256", statement: "公开detector_profiles仍只有profile_sha256，历史差异不能直接区分算法源码变化、有限域行为变化和其他语义Profile变化。pattern-candidate-set-v30为每条规则同时公开algorithm_sha256、behavior_sha256与profile_sha256：算法层来自规范化Go AST调用闭包，简单行为层来自canonical_truth_table_v1，专旺与两气行为层来自behavior_contract_v1见证和形变策略摘要。合法与非法结果继续从同一快照生成规范排序10项，三层任一篡改都由持久化重算拒绝。10条profile_sha256保持不变，总detector_manifest_sha256更新为6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "三层摘要虽已公开，但仍只提供原始散列", statement: "三层摘要虽已公开，但仍只提供原始散列；客户端各自解释差异时可能把behavior_sha256不变误写成传统结论不变。pattern-candidate-set-v31新增layered_detector_digest_delta_v1变更分类合同，按rule_id对齐并仅枚举detector_added、detector_removed、algorithm_digest_changed、behavior_evidence_digest_changed、semantic_profile_digest_changed与layered_digests_unchanged；空ID、重复ID或非64位小写十六进制摘要统一失败关闭。layered_digests_unchanged只表示三条工程证据摘要相同，行为证据范围明确为simple_full_truth_table_complex_partial_contract，推断边界固定为digest_evidence_only，不证明传统正确性、完整四柱行为或现实预测等价。10条三层摘要与总detector_manifest_sha256 6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8保持不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "三层差异仍只存在于响应快照和人工版本说明中", statement: "三层差异仍只存在于响应快照和人工版本说明中，版本升级时无法机器证明说明与前后摘要一致。pattern-candidate-set-v32新增pattern_detector_profile_migration_ledger_v1，规范存储一份去重摘要集、v30、v31与v32三份版本快照、一个预期变化集和两条连续迁移记录；每条预期变化都由layered_detector_digest_delta_v1重算，链尾必须与当前引擎、规则、schema、detector profile、总manifest和10条逐规则摘要一致。响应发布账本引用与规范SHA-256 47d8ce51013f556366d069c0d9c83d5d239099c68c3888e3846184b4f78feae1，引用任一字段篡改由持久化重算拒绝。账本只证明版本快照与摘要差异说明一致，不证明传统正确性、完整行为或预测有效性。10条三层摘要与总detector_manifest_sha256 6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8保持不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "迁移账本虽有整体摘要，但没有逐项前项链接", statement: "迁移账本虽有整体摘要，但没有逐项前项链接，正常追加与重写旧迁移都会只表现为总摘要改变。pattern-candidate-set-v33升级为pattern_detector_profile_migration_ledger_v2并新增pattern_detector_profile_migration_chain_v1：首项previous_migration_sha256固定为64个零，后续项必须引用前项migration_sha256；每项摘要绑定前后完整快照、解析后的逐规则摘要和预期分类，不只散列引用ID。账本现含v30至v33四份快照和三条连续迁移，链头固定为ea922b348fa81df44a70ece07f84b30fc5d8b50d2958e0012219d353ea5de2aa，规范账本SHA-256更新为0c9258b2d186ee641df7455469fb3797a6f2f32cc931c553909fd15e0ab1be2f；响应同步发布链scheme和链头并拒绝任一字段篡改。该合同只证明追加链和工程证据未被静默重写，不证明传统正确性、完整行为或预测有效性。10条三层摘要与总detector_manifest_sha256 6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8保持不变。候选命中、封闭表与未裁决边界不变。"},
				{fragment: "本地摘要链仍可与账本和常量在同一次修改中共同重写", statement: "本地摘要链仍可与账本和常量在同一次修改中共同重写，仅靠嵌入式证据无法形成独立发布交叉检查。pattern-candidate-set-v34新增pattern_detector_profile_release_anchor_v1：根目录release/pattern-detector-profile-anchor.json不嵌入生产二进制，独立固定当前五项版本、总manifest、逐规则摘要集SHA-256 b00a44f1659cc578d1f6fcb321a19aa479e06c5c8ddcfb62682506a7fde0d1e1、账本SHA-256 a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825及链头07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd。repository_ci_cross_check_v1只运行internal/service/bazi锚合同并上传该文件为CI artifact；响应发布路径与规范SHA-256 ebd6323f28715695aa3c4ee9038e74d261c9fa34b422037266c4b097e3086a2e。trust_boundary固定为unsigned_repository_ci_artifact，明确它不是签名tag、透明日志或外部时间戳，仍不能证明整个仓库未被共同重写。该锚只提高工程发布交叉验证，不证明传统正确性、完整行为或预测有效性。10条三层摘要与总detector_manifest_sha256 6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8保持不变。候选命中、封闭表与未裁决边界不变。"},
			} {
				if !strings.Contains(meta.Tables[i].Description, contract.fragment) {
					meta.Tables[i].Description += " " + contract.statement
				}
			}
		}
	}
}

func applyRuleTableCounts(meta *RuleMeta) {
	for i := range meta.Tables {
		switch meta.Tables[i].Key {
		case "ten_god_matrix":
			meta.Tables[i].Count = len(tenGodNames)
		case "ten_god_occurrence":
			meta.Tables[i].Count = len(tenGodNames)
		case "hidden_stems":
			meta.Tables[i].Count = len(data.Zhis)
		case "nayin":
			meta.Tables[i].Count = len(data.NaYinMap)
		case "twelve_stage":
			meta.Tables[i].Count = 10 * 12
		case "tiaohou":
			meta.Tables[i].Count = countTiaohouRules()
		case "pattern_candidates":
			meta.Tables[i].Count = patternDetectorCount()
		case "wuxing_distribution":
			meta.Tables[i].Count = 5
		case "interpretation_boundaries":
			meta.Tables[i].Count = 5
		case "zi_hour_policy":
			meta.Tables[i].Count = 2
		case "calendar_core":
			meta.Tables[i].Count = 59
		case "month_season":
			meta.Tables[i].Count = 12
		case "body_strength":
			meta.Tables[i].Count = len(yueLingDayElementOrder) * len(yueLingMonthBranchOrder)
		}
	}
}

func applyRuleTableSources(meta *RuleMeta) {
	for i := range meta.Tables {
		switch meta.Tables[i].Key {
		case "calendar_core":
			meta.Tables[i].Sources = baziExternalSilverSources()
		case "zi_hour_policy":
			meta.Tables[i].Sources = []RuleSourceMeta{lunarJavascriptSilverSource()}
		case "twelve_stage":
			meta.Tables[i].Sources = tyme4goTerrainSource()
		case "dayun_start":
			meta.Tables[i].Sources = tyme4goDefaultChildLimitSource()
		case "shensha":
			meta.Tables[i].Sources = shenShaClassicalSources()
		}
	}
}

func fallbackRuleMeta() RuleMeta {
	meta := RuleMeta{
		RuleVersion:  RuleVersion,
		School:       RuleSchool,
		BodyStrength: defaultBodyStrengthRuleConfig(),
		Tables: []RuleTableMeta{
			{Key: "calendar_core", Name: "八字基础排盘", Version: "2026-07-16.3", School: "子平四柱与命宫", Source: "tyme4go v1.4.2 + 固定双外部 Silver 差分", Description: "验证四柱、命宫、纳音、大运顺逆和大运柱；普通日期要求双源共识，十二节令、跨世纪立春和两种晚子时口径按秒验证，上游争议案例隔离且不进入准确率分母。"},
			{Key: "month_season", Name: "月令季节事实", Version: "2026-07-15.1", School: "子平月令", Source: "月柱地支寅起十二月序", Description: "仅按月柱地支返回传统月序和春夏秋冬归属，不生成旺衰、吉凶或人生结果断语。"},
			{Key: "ten_god_matrix", Name: "十神矩阵", Version: "2026-06-16.1", School: "子平十神", Source: "五行生克与阴阳同异", Description: "以日干为主，按五行生克与阴阳同异推导十神。"},
			{Key: "ten_god_occurrence", Name: "十神等权计次", Version: "2026-07-15.2", School: "子平十神结构观察", Source: "三处非日主透干与四支全部藏干", Description: "从四柱重算透干与藏干计次、比例、排序和并列最高项；不使用藏干深浅或月令权重，不推断性格、职业、财富、关系或事件概率。"},
			{Key: "stem_relations", Name: "天干关系图", Version: "2026-07-18.1", School: "天干五合四冲生克", Source: "《三命通会》卷十一天干交差相畏；lunar-java 0194eb4574f33ab056fe7cac62a9d8bf24272478 CHONG_GAN_4；天干五合与五行生克规则", Description: "完整覆盖十干比和、生克、五合及甲庚、乙辛、丙壬、丁癸四冲；四冲与五行相克方向分别记录，五合结构和成化状态分离，并返回月令、透干、根气证据及争合/妒合冲突。"},
			{Key: "hidden_stems", Name: "地支藏干", Version: "2026-07-15.2", School: "子平藏干", Source: "tyme4go MAIN/MIDDLE/RESIDUAL", Description: "按四柱地支重算本气、中气、余气及逐柱明细，历史快照必须逐字段一致。"},
			{Key: "nayin", Name: "六十甲子纳音", Version: "2026-07-15.2", School: "纳音五行", Source: "六十甲子纳音映射表 + 外部 Silver 差分", Description: "只返回柱干支对应的纳音名称与五行，不生成性格、运势、家庭或人生结果断语。"},
			{Key: "twelve_stage", Name: "十二长生", Version: "2026-07-15.2", School: "长生十二宫", Source: "tyme4go v1.4.2 HeavenStem.GetTerrain 十干阳顺阴逆表", Description: "按十日干分别计算十二地支长生阶段；身强得地只消费该结构阶段的本地权重，传统裁决和权重参数仍未通过 Gold。"},
			{Key: "branch_relations", Name: "地支关系图", Version: "2026-07-15.2", School: "地支作用", Source: "三命通会、协纪辨方书常用关系表", Description: "作为唯一公开关系合同，统一输出伏吟、六合、六冲、六害、三刑、六破、半合、拱合、半会、三合和三会；结构成立与五行成化分离，显式保留冲突关系，并要求历史快照能由四柱完整重算。"},
			{Key: "shensha", Name: "神煞规则", Version: "2026-07-16.59", School: "子平神煞", Source: "《渊海子平》《三命通会》已定位神煞原文与五份本地资料负向检索", Description: "权威入口要求四个合法六十甲子和规范性别。十恶大败只取古籍十个日柱；九丑、八专、孤鸾和阴差阳错分别固定为古籍9/8/8/12日表。日德与魁罡固定为古籍5/4日表；福德秀气当前取《三命通会》五阴干配巳酉丑的15日完整表。红艳煞按日干固定十干目标并逐柱落位。正学堂与正词馆改由生年纳音五行查长生/临官完整干支，删除日干只查地支的混合旁路。天厨贵人当前Profile只取日干查支表，旧天厨食禄混名删除，十干食禄保持为不同未实现规则。旧文昌贵人表既无固定来源又不等于《三命通会》文昌贵或文星贵两张表，现停止发布，三套口径等待独立Profile裁决。国印贵人改按年干禄宫第九位查支；金舆按日干禄前二辰公式锁定。旧福星贵人日干单支表既丢失完整干支又与年论、多目标版本冲突，现停止发布等待独立Profile裁决。天官贵人改按生年干查支，修正壬寅、癸午并删除日干旧名旁路。太极贵人只按生年干查甲乙子午、丙丁卯酉、戊己辰戌丑未、庚辛寅亥、壬癸巳申，删除日干旁路并记录《三命通会》戊己申位差异。天乙贵人当前只按日干采用《三命通会》甲戊庚丑未等聚合表，删除年干重复入口和时贵别名；《渊海子平》庚位冲突、昼夜/节令分治及年时互换贵保持为待裁决 Profile。年支命后一辰统一使用原名破宅煞；旧宅煞、年/月两套无固定公式的飞廉及错误等同病符的的煞停止发布，暗金的煞另待主键裁决。死符按病符对冲保留结构表但在输出层屏蔽，吊客按命后二辰逐柱落位；无固定目标表的大耗、小耗停止发布。官符、病符、丧门按《三命通会》PDF第122页分别锁定太岁前五辰、太岁后一辰和命前二辰；病符只保留结构名称，不生成健康结论。旧白虎年支表在五份固定资料中无可定位公式，现停止发布。灾煞统一由年支或日支三合组的将星对冲公式产生，删除年支硬表副本；错误等同年支对冲且被高风险屏蔽的自缢煞内部路径删除，只有元数据而无实现的岁破同步失败关闭。劫煞、亡神依据三合表及太岁原文补齐年支主键，与日支结果并取；桃花煞别名统一输出原名咸池，时桃花、时煞、时马、时刃、时禄五个重复别名删除。六厄按生年支三合五行死位锁定；旧墓煞逐项复制华盖墓库位，但古籍同名规则实际要求日干七杀以完整干支入墓，现删除错误别名。将星按年日三合旺位、驿马按年日三合首支对冲位补齐封闭表与可定位元数据；删除时马别名后只保留规范名称逐柱输出。旧流霞、血刃、血忌年干表在固定资料中无可定位查法；外部同名材料分别属于未治理文本、大六壬月煞或六爻月煞，现停止发布，禁止跨术数拼表。羊刃只由日干目标表产生。未定位到名称或公式的六秀日、十灵日及阳煞/阴煞日表不进入正式输出；古籍阴阳煞不由这些旧表冒名实现。其他已定位规则按嵌入清单执行，未登记命中显式标记 unregistered/not_available。"},
			{Key: "tiaohou", Name: "调候表规则候选", Version: "2026-07-18.1", School: "穷通宝鉴", Source: "日干十二月调候表；《穷通宝鉴》二月乙木‘丙为君，癸为臣’配伍原文", Description: "以日干、月令返回全部表规则及表首候选；日期输入的节令区间秒数必须与标准化出生时刻重算一致，表序和深浅均不裁决唯一用神。旧逐行JiShen字段会把同组佐神同时列为喜忌，正式输出统一标记未裁决，不再生成未经复核的忌神或现实吉凶。"},
			{Key: "body_strength", Name: "身强本地评分证据", Version: "2026-07-18.2", School: "扶抑法工程 Profile", Source: "月令、长生、通根透干、禄刃、完整同气三合三会、Influence、归一化与失令生扶 Profile", Description: "得势使用 centered_logistic_v1，得生使用 zero_origin_logistic_v1，零印星证据不再获得 0.5 归一化基线；克泄耗只在 Influence 组件计分一次。完整同气三合、三会只在加权结果仍落弱侧时复用现有 0.5 中和基准，不新增权重、不宣称成化；依据《滴天髓阐微》水木火金四类成局扶身命例，仍等待专家 Gold 裁决。"},
			{Key: "pattern_candidates", Name: "格局规则候选", Version: "2026-07-16.11", School: "子平格局", Source: "从化专旺、特殊日柱与金神时柱检测器；《三命通会》PDF第153-220页各格专章及《渊海子平》PDF第711-713页月令诸格条件用于旧八格和复合格局快捷候选退役审计；《三命通会》PDF第190页专禄、PDF第226及228-230页阳刃月柱、PDF第230-232页建禄；《渊海子平》PDF第162页归禄六忌、PDF第217页日刃三日表；《三命通会》PDF第230页日刃", Description: "严格要求四个合法六十甲子、月支一致、完整非负五行分数和规范身强分段；不以未命中或非法输入生成普通格局。日德辅助候选固定为甲寅、丙辰、戊辰、庚辰、壬戌，魁罡辅助候选固定为庚辰、壬辰、戊戌、庚戌，不发布无来源的水土/土金分类。三奇与金神同样只作为辅助结构，不自动成为主格。专禄固定甲寅、乙卯、庚申、辛酉四个合法自坐禄日，与月令建禄独立并可同时命中；旧日禄归时只检查时支禄位和透干正官，遗漏刑冲、作合、倒食、官星、日月同干、岁月同干六忌的大部分条件，现失败关闭。旧羊刃格快捷候选更名并收敛为丙午、戊午、壬子三个日刃格结构，日刃与月刃独立并可同时命中；刑冲破害、会合和官杀制化只作为未裁决条件。旧月令联合检测器拆分为 pattern.lu.jianlu 与 pattern.lu.yueren：建禄固定十干禄位落月支，月刃只取甲卯、丙午、戊午、庚酉、壬子五阳干Profile；不再共用模糊 pattern.lu.yueling，财官印食、透藏强弱、刑冲会合和官杀制化只作为未裁决条件。旧 pattern.bage.yueling 仅凭月支本气十神或任意其他柱透干就宣称八格成格，现失败关闭。旧五个 pattern.compound.* 检测器把食神制杀、伤官配印、财滋杀旺、食神生财、正官佩印简化为同柱或相邻十神配对；其本气 helper 丢失子卯午酉及四库阴干的十神正偏，邻接 helper 还允许两个十神来自同一侧。各格专章要求的身强旺衰、制化力度、去财去印、伤官伤尽和枭夺食等条件并不相同，现删除五项注册、算法、位置 helper 和统一喜忌。月支藏干与逐项十神事实继续由独立基础层输出；恢复八格或复合格局前须逐格建立条件 Profile 和专家 Gold。全部现存候选仍未经 Gold 验证。"},
			{Key: "wuxing_distribution", Name: "五行计分观察", Version: "2026-07-18.1", School: "四柱五行计分", Source: "天干与地支藏干原始计分", Description: "从四柱完整重算原始分数、天干/藏干明细与缺失/低分观察；只报告结构事实，不从聚合分数推导流通、通关、喜用神或现实建议。"},
			{Key: "interpretation_boundaries", Name: "解释边界", Version: "2026-07-15.1", School: "证据治理", Source: "命理输出与医学、寿命及人身风险事实的边界", Description: "不生成健康字段、器官疾病映射、寿命断语或直接死亡/自伤标签；神煞只保留受控的传统规则命中。"},
			{Key: "zi_hour_policy", Name: "晚子时换日口径", Version: "2026-07-16.2", School: "子平日界", Source: "tyme4go Provider + 固定 lunar-javascript sect 1/2 Silver", Description: "显式选择 late_zi_next_day 或 late_zi_same_day；以相邻秒外部 Silver 锁定 23:00 分离、时柱一致和民用午夜收敛，不修改全局 Provider。"},
			{Key: "true_solar_time", Name: "真太阳时", Version: "2026-07-18.2", School: "地方视太阳时", Source: trueSolarSource, Description: "按出生瞬间时区偏移、经度平太阳时修正与 USNO J2000 均时差公式计算到秒；公开适用范围为 1800–2200 年。候选区间使用6秒工程量：约1角分坐标精度折算4秒，另计UTC/UT1差异1秒和秒级取整1秒；该值不是严格天文误差上界，经度输入误差另计。"},
			{Key: "dayun_start", Name: "大运起运", Version: "2026-07-16.3", School: "阳男阴女顺、阴男阳女逆", Source: "固定 tyme4go v1.4.2 DefaultChildLimitProvider", Description: "显式调用固定默认童限 Profile，不读取可变进程全局 Provider；输出顺逆依据、前后节令、精确时差、起运日期及岁月日时分，历史快照必须与标准化出生时刻和性别的完整重算结果一致。"},
			{Key: "fortune_layers", Name: "运势分层", Version: "2026-07-15.1", School: "大运流年流月小运叠加", Source: "命局、查询时刻与日期级大运边界", Description: "按精确起运时刻划分每十年大运，并叠加流年、流月、小运；只返回柱、边界、十神、关系与神煞等结构证据，不从单柱五行生成增强、减弱、泄秀或现实影响。"},
		},
	}
	for i := range meta.Tables {
		if meta.Tables[i].Key == "shensha" {
			meta.Tables[i].Version = "2026-07-16.59"
			meta.Tables[i].Description += " 五份固定本地PDF未定位当前红鸾、天喜生年支十二项表；《三命通会》同名天喜是春戊、夏丑、秋辰、冬未四季表，红鸾天印是完整干支格，《渊海子平》仅有红鸾吉兆描述，均不能证明旧目标表，现两项停止发布并保持unregistered/not_available。"
			meta.Tables[i].Description += " 旧隔角煞生年支六组相邻对表无固定文本支持；《三命通会》PDF第118页把隔角置于孤辰寡宿与方位语境，《渊海子平》PDF第635、730页分别要求孤寡临角位或举丑寅日时，均不能证明旧主键与封闭表，现停止发布并保持unregistered/not_available。"
			meta.Tables[i].Description += " 孤辰、寡宿只以生年支所在三会方分别取进前一辰、退后一辰并逐柱落位，删除元数据中的日支主键误述；古籍孤寡双全是两项并见后的解释条件，不是第三个孤寡煞命中名，旧聚合别名停止发布并保持unregistered/not_available。"
			meta.Tables[i].Description += " 元辰按《三命通会》PDF第114页修正为阳男阴女取六冲支顺行一辰、阴男阳女取逆行一辰；勾煞、绞煞按该书PDF第117页分别取生年支正反三辰并保留角色，旧合并名勾绞煞停止发布。《渊海子平》PDF第635页不含性别分组的异表不混入当前Profile。旧勾绞双见派生暴败煞没有文本支持；可定位的暴败桃花要求子午卯酉全，故暴败煞路径删除并保持unregistered/not_available。"
			meta.Tables[i].Description += " 旧金锁煞、岁驾、科名、文星、魁星年干表没有可定位封闭表依据；岁驾原文属于生年或太岁禄马条件，魁星另有完整干支和甲辰旬异表，均不能证明旧映射，五项停止发布并保持unregistered/not_available。天刑煞按《三命通会》PDF第122页修正为生年支查时干子丑乙、寅庚、卯辰辛、巳壬、午未癸、申丙、酉戌丁、亥戊，只在时柱落位。"
			meta.Tables[i].Description += " 旧小时、大败、天医月支表停止发布并保持unregistered/not_available。《五行精纪》只定位到与旧小时数值相同但原名为生时忌见月内的规则，不能借异名恢复；大败固定材料属于十恶大败、驿马干支或女命年配月等不同口径；天医只见择方、合婚或年神语境，均未证明旧月支查目标支表。"
		}
	}
	return meta
}

func countTiaohouRules() int {
	total := 0
	for i := range data.TiaohouData {
		for j := range data.TiaohouData[i] {
			total += len(data.TiaohouData[i][j])
		}
	}
	return total
}
