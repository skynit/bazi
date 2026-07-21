package precision

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const comparatorVersion = "precision-comparator-v1"

type checkRuleID string

const (
	checkBaziYearPillar       checkRuleID = "bazi.calendar.year_pillar"
	checkBaziMonthPillar      checkRuleID = "bazi.calendar.month_pillar"
	checkBaziDayPillar        checkRuleID = "bazi.calendar.day_pillar"
	checkBaziHourPillar       checkRuleID = "bazi.calendar.hour_pillar"
	checkBaziDayMaster        checkRuleID = "bazi.calendar.day_master"
	checkBaziBodyStrength     checkRuleID = "bazi.body_strength.score_band_candidate"
	checkBaziPatternCandidate checkRuleID = "bazi.pattern.candidate_membership"

	checkZiweiLegacyPattern    checkRuleID = "ziwei.legacy.pattern_candidate_membership"
	checkZiweiLegacyFiveBureau checkRuleID = "ziwei.legacy.five_bureau"

	checkRikuyoTwelveStage checkRuleID = "rikuyo.twelve_stage_name"
	checkRikuyoJianChu     checkRuleID = "rikuyo.jian_chu_name"
	checkRikuyoHuangDao    checkRuleID = "rikuyo.huang_dao_name"
	checkRikuyoMonthBranch checkRuleID = "rikuyo.month_branch"
	checkRikuyoQueryBranch checkRuleID = "rikuyo.query_branch"
	checkRikuyoQueryStem   checkRuleID = "rikuyo.query_stem"

	checkZiweiGoldScalar       checkRuleID = "ziwei.gold.scalar"
	checkZiweiGoldPalaceScalar checkRuleID = "ziwei.gold.palace_scalar"
	checkZiweiGoldPalaceSet    checkRuleID = "ziwei.gold.palace_set"
	checkZiweiGoldDayun        checkRuleID = "ziwei.gold.dayun"
)

type comparisonMode string

const (
	compareExact     comparisonMode = "exact"
	compareSetMember comparisonMode = "set_member"
	compareSetEqual  comparisonMode = "set_equal"
	compareTolerance comparisonMode = "tolerance"
	compareRubric    comparisonMode = "rubric"
)

type checkRule struct {
	Mode        comparisonMode
	Publishable bool
	Tolerance   float64
}

var checkRegistry = map[checkRuleID]checkRule{
	checkBaziYearPillar:       {Mode: compareExact, Publishable: true},
	checkBaziMonthPillar:      {Mode: compareExact, Publishable: true},
	checkBaziDayPillar:        {Mode: compareExact, Publishable: true},
	checkBaziHourPillar:       {Mode: compareExact, Publishable: true},
	checkBaziDayMaster:        {Mode: compareExact, Publishable: true},
	checkBaziBodyStrength:     {Mode: compareSetMember},
	checkBaziPatternCandidate: {Mode: compareSetMember},

	checkZiweiLegacyPattern:    {Mode: compareSetMember},
	checkZiweiLegacyFiveBureau: {Mode: compareExact},

	checkRikuyoTwelveStage: {Mode: compareExact},
	checkRikuyoJianChu:     {Mode: compareExact},
	checkRikuyoHuangDao:    {Mode: compareExact},
	checkRikuyoMonthBranch: {Mode: compareExact},
	checkRikuyoQueryBranch: {Mode: compareExact},
	checkRikuyoQueryStem:   {Mode: compareExact},

	checkZiweiGoldScalar:       {Mode: compareExact, Publishable: true},
	checkZiweiGoldPalaceScalar: {Mode: compareExact, Publishable: true},
	checkZiweiGoldPalaceSet:    {Mode: compareSetEqual, Publishable: true},
	checkZiweiGoldDayun:        {Mode: compareExact, Publishable: true},
}

func checkRegistryHash() string {
	ids := make([]string, 0, len(checkRegistry))
	for id := range checkRegistry {
		ids = append(ids, string(id))
	}
	sort.Strings(ids)
	hash := sha256.New()
	for _, id := range ids {
		rule := checkRegistry[checkRuleID(id)]
		_, _ = fmt.Fprintf(hash, "%s|%s|%t|%.17g\n", id, rule.Mode, rule.Publishable, rule.Tolerance)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil))
}

func compareFieldCheck(rule checkRule, check fieldCheck) (bool, bool) {
	switch rule.Mode {
	case compareExact:
		return check.got == check.want, true
	case compareSetMember:
		for _, item := range check.gotSet {
			if item == check.want {
				return true, true
			}
		}
		return false, true
	case compareSetEqual:
		return equalStringLists(check.gotSet, check.wantSet), true
	case compareTolerance:
		want, wantErr := strconv.ParseFloat(strings.TrimSpace(check.want), 64)
		got, gotErr := strconv.ParseFloat(strings.TrimSpace(check.got), 64)
		if wantErr != nil || gotErr != nil || rule.Tolerance < 0 {
			return false, false
		}
		return math.Abs(got-want) <= rule.Tolerance, true
	case compareRubric:
		return false, false
	default:
		return false, false
	}
}

func equalStringLists(left, right []string) bool {
	left = sortedStringList(left)
	right = sortedStringList(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func sortedStringList(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
