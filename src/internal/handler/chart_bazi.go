package handler

import (
	"encoding/json"
	"fmt"
	"reflect"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
)

const (
	chartBaziSourceSnapshot   = "bazi_snapshot"
	chartBaziSourceNormalized = bazi.StoredBirthSourceNormalized
	chartBaziSourceRaw        = bazi.StoredBirthSourceRaw
)

type resolvedChartBazi struct {
	Result        *bazi.BaziResult
	BirthYear     int
	Source        string
	EngineVersion string
}

// resolveChartBazi reconstructs the exact factual birth input first, then
// accepts a stored result snapshot only when its versions and four pillars are
// consistent with that input. This keeps saved rule output reproducible without
// allowing a corrupt or unrelated snapshot to drive later fortune queries.
func resolveChartBazi(service *bazi.BaziService, chart *model.BirthChart) (*resolvedChartBazi, error) {
	if service == nil {
		return nil, fmt.Errorf("bazi service is not available")
	}
	if chart == nil {
		return nil, fmt.Errorf("birth chart is nil")
	}

	normalized, calculated, source, err := calculateStoredChartBirth(service, chart)
	if err != nil {
		return nil, err
	}

	if snapshot, ok := validatedBaziSnapshot(chart, calculated, normalized.Gender); ok {
		return &resolvedChartBazi{
			Result:        snapshot,
			BirthYear:     normalized.Year,
			Source:        chartBaziSourceSnapshot,
			EngineVersion: chart.EngineVersion,
		}, nil
	}

	return &resolvedChartBazi{
		Result:        calculated,
		BirthYear:     normalized.Year,
		Source:        source,
		EngineVersion: bazi.EngineVersion,
	}, nil
}

func calculateStoredChartBirth(service *bazi.BaziService, chart *model.BirthChart) (*bazi.NormalizedBirth, *bazi.BaziResult, string, error) {
	resolved, err := bazi.ResolveStoredBirth(service, chart)
	if err != nil {
		return nil, nil, "", err
	}
	return resolved.Normalized, resolved.Result, resolved.Source, nil
}

func validatedBaziSnapshot(chart *model.BirthChart, calculated *bazi.BaziResult, normalizedGender string) (*bazi.BaziResult, bool) {
	if len(chart.BaziSnapshot) == 0 || !json.Valid(chart.BaziSnapshot) {
		return nil, false
	}

	var snapshot bazi.BaziResult
	if err := json.Unmarshal(chart.BaziSnapshot, &snapshot); err != nil {
		return nil, false
	}
	if !isCompleteBaziSnapshot(chart, &snapshot, normalizedGender) ||
		snapshot.ZiHourPolicy != calculated.ZiHourPolicy ||
		!sameFourPillars(&snapshot, calculated) ||
		!reflect.DeepEqual(snapshot.DaYunInfo, calculated.DaYunInfo) ||
		!reflect.DeepEqual(snapshot.Tiaohou, calculated.Tiaohou) {
		return nil, false
	}
	return &snapshot, true
}

func isCompleteBaziSnapshot(chart *model.BirthChart, snapshot *bazi.BaziResult, normalizedGender string) bool {
	if chart.EngineVersion == "" || chart.RuleVersion == "" || snapshot.ZiHourPolicy == "" ||
		snapshot.CalendarVersion != bazi.CalendarEngineVersion ||
		snapshot.RuleVersion != chart.RuleVersion || snapshot.School == "" ||
		snapshot.RuleMeta.RuleVersion != snapshot.RuleVersion ||
		snapshot.RuleMeta.School != snapshot.School ||
		!bazi.ValidRuleMeta(snapshot.RuleMeta) {
		return false
	}
	if snapshot.BodyStrength.RuleVersion != snapshot.RuleVersion ||
		snapshot.BodyStrength.School != snapshot.School ||
		len(snapshot.FiveElements) == 0 ||
		!snapshot.DaYunInfo.Calculated || len(snapshot.DaYunInfo.Pillars) == 0 {
		return false
	}
	pillars := []model.Pillar{snapshot.YearPillar, snapshot.MonthPillar, snapshot.DayPillar, snapshot.HourPillar}
	if !bazi.ValidFiveElementEvidence(
		snapshot.FiveElements,
		snapshot.ElementDetail,
		snapshot.MissingElements,
		pillars,
	) {
		return false
	}
	if !bazi.ValidGanZhiAnalysis(
		snapshot.GanZhiAnalysis,
		snapshot.YearPillar,
		snapshot.MonthPillar,
		snapshot.DayPillar,
		snapshot.HourPillar,
	) {
		return false
	}
	if !bazi.ValidPillarDerivedEvidence(snapshot, normalizedGender) {
		return false
	}
	if !bazi.ValidBodyStrengthEvidence(snapshot.BodyStrength, pillars) {
		return false
	}
	if !bazi.ValidMonthSeasonEvidence(snapshot.MonthSeason, snapshot.MonthPillar.Zhi) {
		return false
	}
	if !bazi.ValidTiaohouEvidenceForPillars(snapshot.Tiaohou, pillars) {
		return false
	}
	if !bazi.ValidPatternAnalysis(
		snapshot.PatternAnalysis,
		pillars,
		snapshot.MonthPillar.Zhi,
	) {
		return false
	}
	for _, item := range []struct {
		key    string
		pillar model.Pillar
	}{
		{key: "year", pillar: snapshot.YearPillar},
		{key: "month", pillar: snapshot.MonthPillar},
		{key: "day", pillar: snapshot.DayPillar},
		{key: "hour", pillar: snapshot.HourPillar},
	} {
		if !bazi.ValidNaYinEvidence(snapshot.NaYin[item.key], item.pillar.Gan, item.pillar.Zhi) {
			return false
		}
	}
	return validPillar(snapshot.YearPillar) &&
		validPillar(snapshot.MonthPillar) &&
		validPillar(snapshot.DayPillar) &&
		validPillar(snapshot.HourPillar)
}

func sameFourPillars(left, right *bazi.BaziResult) bool {
	return left != nil && right != nil &&
		left.YearPillar == right.YearPillar &&
		left.MonthPillar == right.MonthPillar &&
		left.DayPillar == right.DayPillar &&
		left.HourPillar == right.HourPillar
}

func validPillar(pillar model.Pillar) bool {
	_, ganOK := validGans[pillar.Gan]
	_, zhiOK := validZhis[pillar.Zhi]
	return ganOK && zhiOK
}

var validGans = map[string]struct{}{
	"甲": {}, "乙": {}, "丙": {}, "丁": {}, "戊": {},
	"己": {}, "庚": {}, "辛": {}, "壬": {}, "癸": {},
}

var validZhis = map[string]struct{}{
	"子": {}, "丑": {}, "寅": {}, "卯": {}, "辰": {}, "巳": {},
	"午": {}, "未": {}, "申": {}, "酉": {}, "戌": {}, "亥": {},
}
