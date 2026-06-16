package bazi

import "bazi/internal/service/data"

const (
	// RuleVersion identifies the deterministic rule tables and scoring
	// weights used by the BaZi and fortune APIs.
	RuleVersion = "bazi-rules-2026-06-16"
	// RuleSchool names the primary school of interpretation for the current
	// deterministic rule set.
	RuleSchool = "子平八字-扶抑调候-v1"
)

// RuleMeta describes which rule tables were used for a calculation.
type RuleMeta struct {
	RuleVersion string          `json:"rule_version"`
	School      string          `json:"school"`
	Tables      []RuleTableMeta `json:"tables"`
}

// RuleTableMeta is a lightweight manifest entry for one rule table.
type RuleTableMeta struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	School      string `json:"school"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Count       int    `json:"count,omitempty"`
}

// DefaultRuleMeta returns the public rule manifest for API responses.
func DefaultRuleMeta() RuleMeta {
	return RuleMeta{
		RuleVersion: RuleVersion,
		School:      RuleSchool,
		Tables: []RuleTableMeta{
			{
				Key:         "ten_god_matrix",
				Name:        "十神矩阵",
				Version:     "2026-06-16.1",
				School:      "子平十神",
				Source:      "五行生克与阴阳同异",
				Description: "以日干为主，按同我、我生、生我、我克、克我与阴阳同异推导十神。",
				Count:       len(tenGodNames),
			},
			{
				Key:         "hidden_stems",
				Name:        "地支藏干",
				Version:     "2026-06-16.1",
				School:      "子平藏干",
				Source:      "tyme4go MAIN/MIDDLE/RESIDUAL",
				Description: "按本气、中气、余气返回藏干，并在强弱评分中采用 0.6/0.3/0.1 权重。",
				Count:       len(data.Zhis),
			},
			{
				Key:         "nayin",
				Name:        "六十甲子纳音",
				Version:     "2026-06-16.1",
				School:      "纳音五行",
				Source:      "内置三十纳音知识库",
				Description: "六十甲子映射到三十纳音，并附纳音五行、取象与断语。",
				Count:       len(data.NaYinMap),
			},
			{
				Key:         "twelve_stage",
				Name:        "十二长生",
				Version:     "2026-06-16.1",
				School:      "长生十二宫",
				Source:      "阳顺阴逆长生表",
				Description: "用于日课地支状态和身强评分的长生、临官、帝旺等阶段判断。",
				Count:       12,
			},
			{
				Key:         "branch_relations",
				Name:        "刑冲合害破",
				Version:     "2026-06-16.1",
				School:      "地支作用",
				Source:      "三命通会、协纪辨方书常用关系表",
				Description: "覆盖六合、六冲、六害、三刑、六破、三合、三会。",
			},
			{
				Key:         "shensha",
				Name:        "神煞规则",
				Version:     "2026-06-16.1",
				School:      "子平神煞",
				Source:      "内置年干、年支、月支、日干、日柱及全局神煞表",
				Description: "按柱位和全局条件触发天乙、太极、禄神、华盖、驿马等神煞。",
			},
			{
				Key:         "tiaohou",
				Name:        "调候用神",
				Version:     "2026-06-16.1",
				School:      "穷通宝鉴",
				Source:      "日干十二月调候表",
				Description: "以日干、月令取调候喜忌，并参与中和命局的喜用排序。",
				Count:       countTiaohouRules(),
			},
			{
				Key:         "body_strength",
				Name:        "身强身弱评分",
				Version:     "2026-06-16.1",
				School:      "扶抑法",
				Source:      "得令40%、得地30%、得势20%、得生10% + 禄刃加成",
				Description: "输出分项分数、归一化权重、后验修正与喜忌证据。",
			},
			{
				Key:         "fortune_layers",
				Name:        "运势分层",
				Version:     "2026-06-16.1",
				School:      "大运流年流月小运叠加",
				Source:      "命局与查询日期干支作用",
				Description: "按大运、流年、流月、小运分层输出干支、年龄/年份、关系、神煞与五行变化。",
			},
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
