package bazi

import (
	_ "embed"
	"encoding/json"
	"sync"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

const (
	// RuleVersion identifies the deterministic rule tables and scoring
	// weights used by the BaZi and fortune APIs.
	RuleVersion = "bazi-rules-2026-06-16"
	// RuleSchool names the primary school of interpretation for the current
	// deterministic rule set.
	RuleSchool = "子平八字-扶抑调候-v1"
)

type RuleMeta = model.RuleMeta
type RuleTableMeta = model.RuleTableMeta
type BodyStrengthRuleConfig = model.BodyStrengthRuleConfig
type BodyStrengthWeights = model.BodyStrengthWeights
type BodyStrengthNormalizers = model.BodyStrengthNormalizers
type BodyStrengthAdjustmentThresholds = model.BodyStrengthAdjustmentThresholds

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
	})
	return cloneRuleMeta(ruleMeta)
}

func cloneRuleMeta(meta RuleMeta) RuleMeta {
	if len(meta.Tables) > 0 {
		meta.Tables = append([]RuleTableMeta(nil), meta.Tables...)
	}
	return meta
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
		},
		AdjustmentThresholds: BodyStrengthAdjustmentThresholds{
			DeLingRestrictRatio: 0.4,
			DeLingMultiplier:    0.75,
			ShiLingSupportForce: 5.0,
			ShiLingBlendSelf:    0.75,
			ShiLingBlendNeutral: 0.25,
		},
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
	if meta.BodyStrength.AdjustmentThresholds.DeLingRestrictRatio == 0 {
		meta.BodyStrength.AdjustmentThresholds = defaultBodyStrengthRuleConfig().AdjustmentThresholds
	}
}

func applyRuleTableCounts(meta *RuleMeta) {
	for i := range meta.Tables {
		switch meta.Tables[i].Key {
		case "ten_god_matrix":
			meta.Tables[i].Count = len(tenGodNames)
		case "hidden_stems":
			meta.Tables[i].Count = len(data.Zhis)
		case "nayin":
			meta.Tables[i].Count = len(data.NaYinMap)
		case "twelve_stage":
			meta.Tables[i].Count = 12
		case "tiaohou":
			meta.Tables[i].Count = countTiaohouRules()
		}
	}
}

func fallbackRuleMeta() RuleMeta {
	return RuleMeta{
		RuleVersion:  RuleVersion,
		School:       RuleSchool,
		BodyStrength: defaultBodyStrengthRuleConfig(),
		Tables: []RuleTableMeta{
			{Key: "ten_god_matrix", Name: "十神矩阵", Version: "2026-06-16.1", School: "子平十神", Source: "五行生克与阴阳同异", Description: "以日干为主，按五行生克与阴阳同异推导十神。"},
			{Key: "hidden_stems", Name: "地支藏干", Version: "2026-06-16.1", School: "子平藏干", Source: "tyme4go MAIN/MIDDLE/RESIDUAL", Description: "按本气、中气、余气返回藏干。"},
			{Key: "nayin", Name: "六十甲子纳音", Version: "2026-06-16.1", School: "纳音五行", Source: "内置三十纳音知识库", Description: "六十甲子映射到三十纳音。"},
			{Key: "twelve_stage", Name: "十二长生", Version: "2026-06-16.1", School: "长生十二宫", Source: "阳顺阴逆长生表", Description: "用于日课地支状态和身强评分。"},
			{Key: "branch_relations", Name: "刑冲合害破", Version: "2026-06-16.1", School: "地支作用", Source: "三命通会、协纪辨方书常用关系表", Description: "覆盖六合、六冲、六害、三刑、六破、三合、三会。"},
			{Key: "shensha", Name: "神煞规则", Version: "2026-06-16.1", School: "子平神煞", Source: "内置神煞表", Description: "按柱位和全局条件触发神煞。"},
			{Key: "tiaohou", Name: "调候用神", Version: "2026-06-16.1", School: "穷通宝鉴", Source: "日干十二月调候表", Description: "以日干、月令取调候喜忌。"},
			{Key: "body_strength", Name: "身强身弱评分", Version: "2026-06-16.1", School: "扶抑法", Source: "得令/得地/得势/得生权重表", Description: "输出分项分数、证据和后验修正。"},
			{Key: "fortune_layers", Name: "运势分层", Version: "2026-06-16.1", School: "大运流年流月小运叠加", Source: "命局与查询日期干支作用", Description: "按大运、流年、流月、小运分层输出。"},
		},
	}
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
