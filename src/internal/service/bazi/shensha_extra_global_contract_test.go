package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestUnsupportedTianHuoShaFourPillarShortcutIsRemoved(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), `"天火煞"`) {
		t.Fatal("production source still publishes the incomplete four-pillar 天火煞 shortcut")
	}

	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: model.Pillar{Gan: "丙", Zhi: "寅"}, Month: model.Pillar{Gan: "庚", Zhi: "午"},
		Day: model.Pillar{Gan: "戊", Zhi: "戌"}, Hour: model.Pillar{Gan: "戊", Zhi: "戌"}, Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertShenShaNameAbsentEverywhere(t, got, "天火煞")
	meta := LookupShenShaMeta("天火煞")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("removed 天火煞 metadata = %+v", meta)
	}
}

func TestUnsupportedMonthLookupTianHuoIsAlsoRemoved(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
		Day: model.Pillar{Gan: "甲", Zhi: "子"}, Hour: model.Pillar{Gan: "甲", Zhi: "子"}, Gender: "FEMALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertShenShaNameAbsentEverywhere(t, got, "天火")
	assertShenShaNameAbsentEverywhere(t, got, "天火煞")
}

func TestGuaJianAcceptsOnlyTheTwoLocatedClassicalStructures(t *testing.T) {
	positive := []struct {
		branches []string
		evidence string
	}{
		{branches: []string{"巳", "酉", "丑", "申"}, evidence: "巳酉丑申"},
		{branches: []string{"巳", "酉", "丑", "巳"}, evidence: "巳酉丑重"},
		{branches: []string{"巳", "酉", "丑", "酉"}, evidence: "巳酉丑重"},
		{branches: []string{"巳", "酉", "丑", "丑"}, evidence: "巳酉丑重"},
	}
	for _, tc := range positive {
		permutations := uniqueBranchPermutations(tc.branches)
		wantPermutationCount := 12
		if tc.evidence == "巳酉丑申" {
			wantPermutationCount = 24
		}
		if len(permutations) != wantPermutationCount {
			t.Fatalf("%v unique permutation count = %d, want %d", tc.branches, len(permutations), wantPermutationCount)
		}
		for _, branches := range permutations {
			evidence, ok := guaJianEvidence(branches)
			if !ok || evidence != tc.evidence {
				t.Errorf("guaJianEvidence(%v) = %q, %v, want %q, true", branches, evidence, ok, tc.evidence)
			}
			got := calcShenShaForBranches(t, branches)
			assertExactShenShaInBucket(t, "global", got.Global, "挂剑煞："+tc.evidence)
		}
	}

	for _, branches := range [][]string{
		{"巳", "酉", "丑", "子"},
		{"巳", "酉", "申", "申"},
		{"巳", "丑", "申", "申"},
		{"酉", "丑", "申", "申"},
		{"巳", "酉", "丑"},
	} {
		if evidence, ok := guaJianEvidence(branches); ok {
			t.Errorf("guaJianEvidence(%v) = %q, true, want false", branches, evidence)
		}
		if len(branches) == 4 {
			got := calcShenShaForBranches(t, branches)
			if hasShenShaName(got.Global, "挂剑煞") {
				t.Errorf("non-classical branches %v produced 挂剑煞: %+v", branches, got)
			}
		}
	}
}

func TestLeiTingUsesExactTwelveMonthTableAndTargetPillars(t *testing.T) {
	wants := map[string]string{
		"寅": "子", "申": "子", "卯": "寅", "酉": "寅", "辰": "辰", "戌": "辰",
		"巳": "午", "亥": "午", "午": "申", "子": "申", "未": "戌", "丑": "戌",
	}
	if len(leiTingShaByMonthZhi) != 12 {
		t.Fatalf("雷霆煞 month table size = %d, want 12", len(leiTingShaByMonthZhi))
	}
	for _, month := range data.Zhis {
		target := wants[month]
		if got := leiTingShaByMonthZhi[month]; got != target {
			t.Errorf("雷霆煞 month %s target = %s, want %s", month, got, target)
		}

		filler := firstBranchExcept(target, month)
		branches := []string{target, month, filler, target}
		result := calcShenShaForBranches(t, branches)
		buckets := [][]string{result.Year, result.Month, result.Day, result.Hour}
		for i, bucket := range buckets {
			wantHit := branches[i] == target
			if gotHit := hasShenShaName(bucket, "雷霆煞"); gotHit != wantHit {
				t.Errorf("month %s target %s %s hit = %v, want %v: %+v", month, target, pillarBucketName(i), gotHit, wantHit, result)
			}
		}
		if hasShenShaName(result.Global, "雷霆煞") {
			t.Errorf("month %s target %s leaked 雷霆煞 into global: %+v", month, target, result)
		}
	}
}

func TestLeiTingRejectsFormerSpringFourCardinalShortcut(t *testing.T) {
	for _, branches := range [][]string{
		{"午", "寅", "午", "午"},
		{"酉", "卯", "酉", "酉"},
	} {
		got := calcShenShaForBranches(t, branches)
		assertShenShaNameAbsentEverywhere(t, got, "雷霆煞")
	}
}

func TestLocatedGuaJianAndLeiTingMetadata(t *testing.T) {
	for _, tc := range []struct {
		name      string
		citations []string
	}{
		{name: "挂剑煞", citations: []string{"PDF第121页", "书内第118页", "巳酉丑申", "重见"}},
		{name: "雷霆煞", citations: []string{"PDF第122页", "书内第119页", "正七子", "六十二戌", "逐柱"}},
	} {
		meta := LookupShenShaMeta(tc.name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v", tc.name, meta)
		}
		for _, citation := range tc.citations {
			if !strings.Contains(meta.Basis, citation) {
				t.Errorf("%s basis = %q, want %q", tc.name, meta.Basis, citation)
			}
		}
	}
}

func calcShenShaForBranches(t testing.TB, branches []string) ShenShaCalcResult {
	t.Helper()
	if len(branches) != 4 {
		t.Fatalf("branch fixture length = %d, want 4", len(branches))
	}
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: shenShaTestPillarForBranch(t, branches[0]), Month: shenShaTestPillarForBranch(t, branches[1]),
		Day: shenShaTestPillarForBranch(t, branches[2]), Hour: shenShaTestPillarForBranch(t, branches[3]), Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func shenShaTestPillarForBranch(t testing.TB, branch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		if data.Zhis[i%12] == branch {
			return model.Pillar{Gan: data.Gans[i%10], Zhi: branch}
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %q", branch)
	return model.Pillar{}
}

func uniqueBranchPermutations(branches []string) [][]string {
	var result [][]string
	var walk func(int)
	walk = func(index int) {
		if index == len(branches) {
			result = append(result, append([]string(nil), branches...))
			return
		}
		seen := map[string]bool{}
		for i := index; i < len(branches); i++ {
			if seen[branches[i]] {
				continue
			}
			seen[branches[i]] = true
			branches[index], branches[i] = branches[i], branches[index]
			walk(index + 1)
			branches[index], branches[i] = branches[i], branches[index]
		}
	}
	walk(0)
	return result
}

func firstBranchExcept(excluded ...string) string {
	for _, candidate := range data.Zhis {
		match := false
		for _, value := range excluded {
			if candidate == value {
				match = true
				break
			}
		}
		if !match {
			return candidate
		}
	}
	return ""
}

func assertShenShaNameAbsentEverywhere(t testing.TB, got ShenShaCalcResult, name string) {
	t.Helper()
	for bucketName, bucket := range map[string][]string{
		"year": got.Year, "month": got.Month, "day": got.Day, "hour": got.Hour, "global": got.Global,
	} {
		if hasShenShaName(bucket, name) {
			t.Errorf("%s unexpectedly appears in %s bucket: %+v", name, bucketName, got)
		}
	}
}
