package bazi

import (
	"testing"

	"bazi/internal/model"
)

func TestValidPillarDerivedEvidenceRejectsInternallyConsistentTampering(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", model.GenderMale)
	if err != nil {
		t.Fatal(err)
	}
	if !ValidPillarDerivedEvidence(result, model.GenderMale) {
		t.Fatal("freshly calculated pillar-derived evidence must validate")
	}

	tamperedTenGod := *result
	tamperedTenGod.TenGodProportion = append([]TenGodRatio(nil), result.TenGodProportion...)
	tamperedTenGod.TenGodProportion[0].Count++
	tamperedTenGod.TenGodAnalysis = ObserveTenGodDistribution(tamperedTenGod.TenGodProportion)
	if ValidPillarDerivedEvidence(&tamperedTenGod, model.GenderMale) {
		t.Fatal("internally consistent but pillar-inconsistent ten-god counts must be rejected")
	}

	tamperedHiddenStems := *result
	tamperedHiddenStems.HiddenStems = cloneHiddenStemMap(result.HiddenStems)
	tamperedHiddenStems.HiddenStems["year"] = append(tamperedHiddenStems.HiddenStems["year"], "篡改")
	if ValidPillarDerivedEvidence(&tamperedHiddenStems, model.GenderMale) {
		t.Fatal("tampered hidden stems must be rejected")
	}

	tamperedMingGong := *result
	tamperedMingGong.MingGong.ShenSha = "篡改"
	if ValidPillarDerivedEvidence(&tamperedMingGong, model.GenderMale) {
		t.Fatal("tampered ming-gong detail must be rejected")
	}

	tamperedShenSha := *result
	tamperedShenSha.DayShenSha = append(append([]string(nil), result.DayShenSha...), "天乙贵人：子")
	tamperedShenSha.DayShenShaDetails = BuildShenShaDetails(tamperedShenSha.DayShenSha)
	tamperedShenSha.ShenShaByPillar = append([]PillarShenSha(nil), result.ShenShaByPillar...)
	tamperedShenSha.ShenShaByPillar[2].Items = append([]string(nil), tamperedShenSha.DayShenSha...)
	tamperedShenSha.ShenShaByPillar[2].Details = BuildShenShaDetails(tamperedShenSha.DayShenSha)
	if ValidPillarDerivedEvidence(&tamperedShenSha, model.GenderMale) {
		t.Fatal("internally consistent but pillar-inconsistent shen-sha evidence must be rejected")
	}

	if ValidPillarDerivedEvidence(result, "unknown") {
		t.Fatal("invalid normalized gender must be rejected")
	}
}

func cloneHiddenStemMap(source map[string][]string) map[string][]string {
	cloned := make(map[string][]string, len(source))
	for key, values := range source {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
