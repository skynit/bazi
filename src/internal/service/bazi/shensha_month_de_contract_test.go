package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestMonthDeTablesAndCombinationTargetsAreExact(t *testing.T) {
	wantTianDe := [12]string{"巳", "庚", "丁", "申", "壬", "辛", "亥", "甲", "癸", "寅", "丙", "乙"}
	wantYueDe := [12]string{"壬", "庚", "丙", "甲", "壬", "庚", "丙", "甲", "壬", "庚", "丙", "甲"}
	wantTianDeHe := [12]string{"申", "乙", "壬", "巳", "丁", "丙", "寅", "己", "戊", "亥", "辛", "庚"}
	wantYueDeHe := [12]string{"丁", "乙", "辛", "己", "丁", "乙", "辛", "己", "丁", "乙", "辛", "己"}

	if got := data.MonthShenMap[data.TianDe]; got != wantTianDe {
		t.Fatalf("天德十二月表 = %v, want %v", got, wantTianDe)
	}
	if got := data.MonthShenMap[data.YueDe]; got != wantYueDe {
		t.Fatalf("月德十二月表 = %v, want %v", got, wantYueDe)
	}
	for i := range wantTianDe {
		if got := monthDeHeTarget(wantTianDe[i]); got != wantTianDeHe[i] {
			t.Errorf("month %s 天德合目标 = %q, want %q", data.Zhis[i], got, wantTianDeHe[i])
		}
		if got := monthDeHeTarget(wantYueDe[i]); got != wantYueDeHe[i] {
			t.Errorf("month %s 月德合目标 = %q, want %q", data.Zhis[i], got, wantYueDeHe[i])
		}
	}
	if got := monthDeHeTarget("unknown"); got != "" {
		t.Fatalf("unknown 德合 target = %q, want empty", got)
	}
}

func TestTianDeStemTargetCanAttachToEveryPillarBucket(t *testing.T) {
	base := ShenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "子"},
		Month: model.Pillar{Gan: "戊", Zhi: "寅"},
		Day:   model.Pillar{Gan: "丙", Zhi: "辰"},
		Hour:  model.Pillar{Gan: "庚", Zhi: "午"},
	}
	for _, tc := range []struct {
		name   string
		index  int
		mutate func(*ShenShaPillars)
	}{
		{name: "year", index: 0, mutate: func(p *ShenShaPillars) { p.Year.Gan = "丁" }},
		{name: "month", index: 1, mutate: func(p *ShenShaPillars) { p.Month.Gan = "丁" }},
		{name: "day", index: 2, mutate: func(p *ShenShaPillars) { p.Day.Gan = "丁" }},
		{name: "hour", index: 3, mutate: func(p *ShenShaPillars) { p.Hour.Gan = "丁" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pillars := base
			tc.mutate(&pillars)
			var got ShenShaCalcResult
			addMonthGanRules(pillars, &got)
			assertOnlyPillarBucketHas(t, got, tc.index, "天德贵人：丁")
		})
	}
}

func TestTianDeBranchTargetUsesBranchesIncludingYearAndHour(t *testing.T) {
	pillars := ShenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "申"},
		Month: model.Pillar{Gan: "乙", Zhi: "卯"},
		Day:   model.Pillar{Gan: "丙", Zhi: "子"},
		Hour:  model.Pillar{Gan: "丁", Zhi: "申"},
	}
	var got ShenShaCalcResult
	addMonthGanRules(pillars, &got)

	assertExactShenShaInBucket(t, "year", got.Year, "天德贵人：申")
	assertExactShenShaInBucket(t, "hour", got.Hour, "天德贵人：申")
	if containsExactShenSha(got.Month, "天德贵人：申") || containsExactShenSha(got.Day, "天德贵人：申") {
		t.Fatalf("天德地支目标 leaked to nonmatching buckets: %+v", got)
	}
}

func TestTianDeCombinationMatchesStemAndBranchTargets(t *testing.T) {
	stemPillars := ShenShaPillars{
		Year:  model.Pillar{Gan: "壬", Zhi: "子"},
		Month: model.Pillar{Gan: "戊", Zhi: "寅"},
		Day:   model.Pillar{Gan: "丙", Zhi: "辰"},
		Hour:  model.Pillar{Gan: "壬", Zhi: "午"},
	}
	var stemGot ShenShaCalcResult
	addMonthGanRules(stemPillars, &stemGot)
	assertExactShenShaInBucket(t, "year", stemGot.Year, "天德合：壬")
	assertExactShenShaInBucket(t, "hour", stemGot.Hour, "天德合：壬")

	branchPillars := ShenShaPillars{
		Year:  model.Pillar{Gan: "甲", Zhi: "子"},
		Month: model.Pillar{Gan: "乙", Zhi: "卯"},
		Day:   model.Pillar{Gan: "丙", Zhi: "巳"},
		Hour:  model.Pillar{Gan: "丁", Zhi: "巳"},
	}
	var branchGot ShenShaCalcResult
	addMonthGanRules(branchPillars, &branchGot)
	assertExactShenShaInBucket(t, "day", branchGot.Day, "天德合：巳")
	assertExactShenShaInBucket(t, "hour", branchGot.Hour, "天德合：巳")
}

func TestYueDeAndYueDeCombinationAreDayStemOnly(t *testing.T) {
	for _, tc := range []struct {
		name   string
		target string
	}{
		{name: "月德贵人", target: "壬"},
		{name: "月德合", target: "丁"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pillars := ShenShaPillars{
				Year:  model.Pillar{Gan: tc.target, Zhi: "寅"},
				Month: model.Pillar{Gan: tc.target, Zhi: "子"},
				Day:   model.Pillar{Gan: "甲", Zhi: "辰"},
				Hour:  model.Pillar{Gan: tc.target, Zhi: "午"},
			}
			var negative ShenShaCalcResult
			addMonthGanRules(pillars, &negative)
			assertNoPillarBucketHas(t, negative, tc.name+"："+tc.target)

			pillars.Day.Gan = tc.target
			var positive ShenShaCalcResult
			addMonthGanRules(pillars, &positive)
			assertOnlyPillarBucketHas(t, positive, 2, tc.name+"："+tc.target)
		})
	}
}

func TestMonthDeRulesFlowThroughAuthoritativeEntrypoint(t *testing.T) {
	t.Run("stem tian-de on year pillar", func(t *testing.T) {
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year:   model.Pillar{Gan: "丁", Zhi: "丑"},
			Month:  model.Pillar{Gan: "戊", Zhi: "寅"},
			Day:    model.Pillar{Gan: "甲", Zhi: "子"},
			Hour:   model.Pillar{Gan: "甲", Zhi: "子"},
			Gender: "MALE",
		})
		if err != nil {
			t.Fatal(err)
		}
		assertExactShenShaInBucket(t, "year", got.Year, "天德贵人：丁")
	})

	t.Run("branch tian-de on repeated year and hour pillars", func(t *testing.T) {
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year:   model.Pillar{Gan: "壬", Zhi: "申"},
			Month:  model.Pillar{Gan: "己", Zhi: "卯"},
			Day:    model.Pillar{Gan: "甲", Zhi: "子"},
			Hour:   model.Pillar{Gan: "壬", Zhi: "申"},
			Gender: "FEMALE",
		})
		if err != nil {
			t.Fatal(err)
		}
		assertExactShenShaInBucket(t, "year", got.Year, "天德贵人：申")
		assertExactShenShaInBucket(t, "hour", got.Hour, "天德贵人：申")
	})

	t.Run("yue-de remains day-only", func(t *testing.T) {
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year:   model.Pillar{Gan: "壬", Zhi: "申"},
			Month:  model.Pillar{Gan: "丙", Zhi: "子"},
			Day:    model.Pillar{Gan: "甲", Zhi: "子"},
			Hour:   model.Pillar{Gan: "壬", Zhi: "申"},
			Gender: "MALE",
		})
		if err != nil {
			t.Fatal(err)
		}
		assertNoPillarBucketHas(t, got, "月德贵人：壬")

		got, err = CalcShenShaByPillars(ShenShaPillars{
			Year:   model.Pillar{Gan: "甲", Zhi: "子"},
			Month:  model.Pillar{Gan: "丙", Zhi: "子"},
			Day:    model.Pillar{Gan: "壬", Zhi: "申"},
			Hour:   model.Pillar{Gan: "甲", Zhi: "子"},
			Gender: "MALE",
		})
		if err != nil {
			t.Fatal(err)
		}
		assertOnlyPillarBucketHas(t, got, 2, "月德贵人：壬")
	})
}

func TestMonthDeMetadataHasLocatedClassicalEvidence(t *testing.T) {
	wants := map[string][]string{
		"天德贵人": {"《渊海子平》第75页", "四柱"},
		"天德合":  {"《三命通会》第103页", "四柱"},
		"月德贵人": {"《渊海子平》第71页", "日干"},
		"月德合":  {"《渊海子平》第73页", "日干"},
	}
	for name, fragments := range wants {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata status = %+v", name, meta)
		}
		for _, fragment := range fragments {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s metadata basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}
}

func assertOnlyPillarBucketHas(t testing.TB, got ShenShaCalcResult, wantIndex int, item string) {
	t.Helper()
	buckets := [][]string{got.Year, got.Month, got.Day, got.Hour}
	for i, bucket := range buckets {
		if containsExactShenSha(bucket, item) != (i == wantIndex) {
			t.Errorf("bucket %d shen-sha = %v for %s, want only bucket %d", i, bucket, item, wantIndex)
		}
	}
}

func assertNoPillarBucketHas(t testing.TB, got ShenShaCalcResult, item string) {
	t.Helper()
	for i, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
		if containsExactShenSha(bucket, item) {
			t.Errorf("bucket %d shen-sha = %v, must not contain %s", i, bucket, item)
		}
	}
}
