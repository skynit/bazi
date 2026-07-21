package bazi

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBaziOutputOmitsAutomaticHealthDiagnosis(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbiddenField := range []string{
		`"health_note"`, `"jin_bu_huan"`, `"shen_sha_desc"`, `"zhi_detail"`,
		`"season_text"`, `"season_text_month"`, `"wuxing_season_note"`,
		`"ri_zhu_desc"`, `"ri_zhu_poem"`, `"ri_zhu_source"`,
		`"ri_zhu_comment"`, `"ri_zhu_hour_detail"`, `"jia_zi_detail"`,
		`"shen_sha_summary"`,
	} {
		if strings.Contains(string(payload), forbiddenField) {
			t.Fatalf("unsafe interpretation field %s leaked into chart JSON: %s", forbiddenField, payload)
		}
	}
	for _, forbiddenField := range []string{`"flow_pattern_desc"`, `"dayun_flow"`, `"flow_change"`, `"clash_harmony"`} {
		if strings.Contains(string(payload), forbiddenField) {
			t.Fatalf("Bazi result leaked removed flow-result field %s: %s", forbiddenField, payload)
		}
	}
	for _, forbiddenText := range []string{
		"长寿健康", "福寿绵长", "一生少灾", "女命妨夫", "不宜早婚",
		"克子刑妻", "死无棺椁", "短夭", "聋哑", "主法死",
		"女不靠子女", "多子多福", "早年运势辛苦", "与父母中一人缘薄",
	} {
		if strings.Contains(string(payload), forbiddenText) {
			t.Fatalf("unadjudicated ming-gong outcome %q leaked into chart JSON: %s", forbiddenText, payload)
		}
	}
}

func TestTenGodAnalysisIsStructuralEvidenceOnly(t *testing.T) {
	counts := map[string]int{"比肩": 2, "劫财": 2, "食神": 1}
	proportions := make([]TenGodRatio, 0, len(tenGodNames))
	for _, god := range tenGodNames {
		proportions = append(proportions, TenGodRatio{Name: god, Count: counts[god], Percent: 99})
	}

	analysis := ObserveTenGodDistribution(proportions)
	if analysis.Status != "observed" || analysis.ValidationStatus != "not_validated" || analysis.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("ten-god governance metadata is incomplete: %+v", analysis)
	}
	if analysis.TotalOccurrences != 5 || analysis.DominantPercent != 40 {
		t.Fatalf("ten-god totals must be derived from counts: %+v", analysis)
	}
	if len(analysis.DominantGods) != 2 || analysis.DominantGods[0] != "比肩" || analysis.DominantGods[1] != "劫财" {
		t.Fatalf("tied dominant gods must remain explicit: %+v", analysis.DominantGods)
	}
	if len(analysis.RankedGods) != len(tenGodNames) || analysis.RankedGods[0].Rank != 1 || analysis.RankedGods[1].Rank != 1 || analysis.RankedGods[2].Rank != 2 {
		t.Fatalf("ten-god dense ranking is invalid: %+v", analysis.RankedGods)
	}

	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, forbiddenField := range []string{
		`"personality"`, `"interpersonal"`, `"career_fortune"`, `"emotion_relation"`,
		`"taboo"`, `"summary"`, `"meaning"`, `"advice"`, `"health_note"`,
	} {
		if strings.Contains(text, forbiddenField) {
			t.Fatalf("ten-god analysis leaked outcome field %s: %s", forbiddenField, payload)
		}
	}
	for _, forbiddenText := range []string{
		"事业", "财运", "投资", "婚姻", "感情", "性格", "疾病", "健康", "必有所成",
	} {
		if strings.Contains(text, forbiddenText) {
			t.Fatalf("ten-god analysis inferred %q from occurrences: %s", forbiddenText, payload)
		}
	}
}

func TestTenGodAnalysisRejectsIncompleteOrDuplicateInput(t *testing.T) {
	incomplete := ObserveTenGodDistribution([]TenGodRatio{{Name: "比肩", Count: 1}})
	if incomplete.Status != "unavailable" || len(incomplete.RankedGods) != 0 {
		t.Fatalf("incomplete distribution must be unavailable: %+v", incomplete)
	}

	duplicate := make([]TenGodRatio, len(tenGodNames))
	for index, god := range tenGodNames {
		duplicate[index] = TenGodRatio{Name: god, Count: 1}
	}
	duplicate[len(duplicate)-1].Name = duplicate[0].Name
	got := ObserveTenGodDistribution(duplicate)
	if got.Status != "unavailable" || len(got.RankedGods) != 0 {
		t.Fatalf("duplicate distribution must be unavailable: %+v", got)
	}
}

func TestShenShaOutputDoesNotEmbedMedicalFortuneText(t *testing.T) {
	var items []string
	for _, name := range []string{"病符", "白虎", "流霞", "血刃", "血忌", "死符"} {
		appendShenSha(&items, name, "子")
	}
	for _, item := range items {
		if strings.Contains(item, "｜") {
			t.Fatalf("raw shen-sha item embedded an interpretation: %q", item)
		}
	}

	for name := range suppressedHighRiskShenSha {
		var suppressed []string
		appendShenSha(&suppressed, name, "子")
		if len(suppressed) != 0 {
			t.Fatalf("high-risk shen-sha %q was emitted: %v", name, suppressed)
		}
	}

	for _, name := range []string{"病符"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" || meta.Basis == "" {
			t.Fatalf("medical shen-sha name was not constrained to a rule hit: %+v", meta)
		}
	}
	for _, name := range []string{"白虎", "流霞", "血刃", "血忌"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Fatalf("retired medicalized shen-sha %q metadata = %+v", name, meta)
		}
	}
}

func TestRetiredBloodShenShaDoNotLeakNamesOrDiagnosisText(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲申", "丙寅", "甲子", "癸酉", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	shenShaPayload, err := json.Marshal(struct {
		Day           []string        `json:"day"`
		DayDetails    []ShenShaMeta   `json:"day_details"`
		ByPillar      []PillarShenSha `json:"by_pillar"`
		Global        []string        `json:"global"`
		GlobalDetails []ShenShaMeta   `json:"global_details"`
	}{result.DayShenSha, result.DayShenShaDetails, result.ShenShaByPillar, result.GlobalShenSha, result.GlobalShenShaDetails})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"流霞", "血刃", "血忌"} {
		if strings.Contains(string(shenShaPayload), forbidden) {
			t.Fatalf("retired blood shen-sha leaked %q into result fields: %s", forbidden, shenShaPayload)
		}
	}
	for _, forbidden := range []string{"产厄", "血光", "开刀", "创伤", "出血", "手术", "穿刺"} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("retired blood shen-sha leaked diagnostic text %q into chart JSON: %s", forbidden, payload)
		}
	}
}
