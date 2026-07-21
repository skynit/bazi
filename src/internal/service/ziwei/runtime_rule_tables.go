package ziwei

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	ZiWeiRuntimeRuleTablesSchema = "ziwei-runtime-rule-tables-v1"
	ZiWeiRuntimeRuleTablesSHA256 = "75feeb8b9c835bad275a4bce3f7364f67fb33d872965d2787802bd97d8cbc6d4"
)

func runtimeRuleTablesHash() (string, error) {
	patternNames := make([]string, len(patternCheckers))
	for i, checker := range patternCheckers {
		patternNames[i] = checker.name
	}
	payload := map[string]any{
		"schema":                      ZiWeiRuntimeRuleTablesSchema,
		"stem_index":                  StemIndex,
		"stem_names":                  StemNames,
		"branch_index":                BranchIndex,
		"branch_names":                BranchNames,
		"branch_five_element":         BranchFiveElement,
		"palace_names":                ZIWEI_PALACE_NAMES,
		"tiger_rule":                  TigerRule,
		"five_bureau_name":            FiveBureauName,
		"five_bureau_value":           FiveBureauValue,
		"nayin_bureau":                NaYinBureauTable,
		"ziwei_star_order":            ZiweiStarOrder,
		"tianfu_star_order":           TianfuStarOrder,
		"star_brightness":             StarBrightnessMap,
		"aux_star_brightness":         AuxStarBrightnessMap,
		"sihua":                       SiHuaTable,
		"sihua_labels":                SiHuaLabels,
		"kui_yue":                     KuiYueTable,
		"lucun_branch":                LucunBranchIdx,
		"legacy_lucun_branch":         LuCunTable,
		"tianma_branch":               TianmaBranchIdx,
		"xianchi_branch":              XianChiBranch,
		"huagai_branch":               HuaGaiBranch,
		"yuejie_branch_by_month":      YueJieBranchByMonth,
		"tianyue_branch_by_month":     TianYueBranchByMonth,
		"tianwu_branch_by_month":      TianWuBranchByMonth,
		"guchen_branch":               GuChenBranch,
		"guasu_branch":                GuaSuBranch,
		"posui_branch":                PoSuiBranch,
		"feilian_branch":              FeiLianBranch,
		"yinsha_branch":               YinShaBranch,
		"tianchu_branch_by_stem":      TianChuBranchByStem,
		"tianguan_branch_by_stem":     TianGuanBranchByStem,
		"tianfu_adj_branch_by_stem":   TianFuAdjectiveBranchByStem,
		"jielu_branch_by_stem":        JieLuBranchByStem,
		"kongwang_branch_by_stem":     KongWangBranchByStem,
		"nianjie_branch":              NianJieBranch,
		"life_master":                 LifeMasterTable,
		"body_master":                 BodyMasterTable,
		"changsheng_start":            ChangshengStartBranch,
		"changsheng_12":               ChangSheng12,
		"boshi_12":                    BoShi12,
		"suiqian_12":                  SuiQian12,
		"jiangqian_12":                JiangQian12,
		"jiangqian_start":             JiangQianStartBranch,
		"transit_chang_qu":            transitChangQuBranch,
		"transit_star_labels":         transitStarLabels,
		"brightness_levels":           brightnessLevels,
		"brightness_descriptions":     STAR_BRIGHTNESS,
		"aux_brightness_descriptions": STAR_BRIGHTNESS_AUX,
		"pattern_lucun":               LUCUN_TABLE,
		"pattern_tianma":              TIANMA_TABLE,
		"pattern_palace_names":        PALACE_NAMES,
		"pattern_checker_names":       patternNames,
		"template_brightness_levels":  templateBrightnessLevels,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal ziwei runtime rule tables: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func validateRuntimeRuleTables(profile CalculationProfile) error {
	if profile.RuntimeRuleTablesSchema != ZiWeiRuntimeRuleTablesSchema ||
		profile.RuntimeRuleTablesHash != ZiWeiRuntimeRuleTablesSHA256 {
		return fmt.Errorf("profile runtime rule-table contract is invalid")
	}
	actual, err := runtimeRuleTablesHash()
	if err != nil {
		return err
	}
	if actual != profile.RuntimeRuleTablesHash {
		return fmt.Errorf("runtime rule-table hash mismatch: got %s want %s", actual, profile.RuntimeRuleTablesHash)
	}
	return nil
}
