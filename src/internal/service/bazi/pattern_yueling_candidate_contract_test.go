package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestJianLuAndYueRenExactProfilesAcrossTenByTwelve(t *testing.T) {
	jianLuMatches := 0
	yueRenMatches := 0
	for _, dayGan := range data.Gans {
		for _, monthZhi := range data.Zhis {
			jianLu := checkJianLuGe(dayGan, monthZhi)
			luBranch, _ := luBranchForStem(dayGan)
			wantJianLu := monthZhi == luBranch
			if (jianLu != nil) != wantJianLu {
				t.Errorf("%s day %s month jian-lu = %+v, want %v", dayGan, monthZhi, jianLu, wantJianLu)
			}
			if jianLu != nil {
				jianLuMatches++
				if jianLu.PatternName != "建禄格" {
					t.Errorf("%s day %s month jian-lu metadata = %+v", dayGan, monthZhi, jianLu)
				}
			}

			yueRen := checkYueRenGe(dayGan, monthZhi)
			wantYueRen := monthZhi == yangRenZhi(dayGan) && yangRenZhi(dayGan) != ""
			if (yueRen != nil) != wantYueRen {
				t.Errorf("%s day %s month yue-ren = %+v, want %v", dayGan, monthZhi, yueRen, wantYueRen)
			}
			if yueRen != nil {
				yueRenMatches++
				if yueRen.PatternName != "月刃格" {
					t.Errorf("%s day %s month yue-ren metadata = %+v", dayGan, monthZhi, yueRen)
				}
			}
		}
	}
	if jianLuMatches != 10 || yueRenMatches != 5 {
		t.Fatalf("month-profile matches = jian-lu %d/yue-ren %d, want 10/5", jianLuMatches, yueRenMatches)
	}
}

func TestJianLuAndYueRenUseIndependentFormalRuleIDs(t *testing.T) {
	jianLuPillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	jianLu := AnalyzePatternExtended(jianLuPillars, "寅")
	assertMonthPatternCandidate(t, jianLu.Candidates, "pattern.lu.jianlu", "建禄格", "PDF第230-232页")
	if hasLuPatternRuleID(jianLu.Candidates, "pattern.lu.yueren") {
		t.Errorf("jian-lu chart also published yue-ren: %+v", jianLu.Candidates)
	}

	yueRenPillars := []model.Pillar{
		{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "午"},
		{Gan: "丙", Zhi: "午"}, {Gan: "庚", Zhi: "辰"},
	}
	yueRen := AnalyzePatternExtended(yueRenPillars, "午")
	assertMonthPatternCandidate(t, yueRen.Candidates, "pattern.lu.yueren", "月刃格", "PDF第226、228-230页")
	if hasLuPatternRuleID(yueRen.Candidates, "pattern.lu.jianlu") {
		t.Errorf("yue-ren chart also published jian-lu: %+v", yueRen.Candidates)
	}

	for _, candidates := range [][]PatternCandidate{jianLu.Candidates, yueRen.Candidates} {
		if hasLuPatternRuleID(candidates, "pattern.lu.yueling") {
			t.Errorf("retired combined month rule ID survived: %+v", candidates)
		}
	}
}

func TestCombinedMonthPatternShortcutIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"checkJianLuYueRen", "pattern.lu.yueling", "月刃用官须透出有根", "建禄格用财须身旺方能任财"} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains combined shortcut %q", path, forbidden)
			}
		}
	}
}

func TestMonthPatternManifestRecordsSplitProfiles(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"pattern.lu.jianlu 与 pattern.lu.yueren",
			"建禄固定十干禄位落月支",
			"甲卯、丙午、戊午、庚酉、壬子五阳干Profile",
			"不再共用模糊 pattern.lu.yueling",
			"PDF第226及228-230页阳刃月柱",
			"PDF第230-232页建禄",
		} {
			if !strings.Contains(table.Description+table.Source, fragment) {
				t.Errorf("pattern description/source missing %q: %+v", fragment, table)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func assertMonthPatternCandidate(t *testing.T, candidates []PatternCandidate, ruleID, name, sourceFragment string) {
	t.Helper()
	candidate, ok := luPatternCandidateByID(candidates, ruleID)
	if !ok || candidate.PatternName != name || !strings.Contains(candidate.Source, sourceFragment) {
		t.Errorf("month pattern %s = %+v", ruleID, candidate)
	}
}

func yueLingPatternScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}
