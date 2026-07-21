package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

type patternBehaviorTruthRow struct {
	Input  string `json:"input"`
	Output string `json:"output,omitempty"`
}

func simplePatternDetectorByRuleID(t *testing.T, ruleID string) patternDetectorDefinition {
	t.Helper()
	for _, detector := range patternDetectorRegistry() {
		if detector.ruleID == ruleID {
			return detector
		}
	}
	t.Fatalf("simple detector %s not found", ruleID)
	return patternDetectorDefinition{}
}

func canonicalPatternStems() []string {
	targets := patternStemElementProfile()
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		result = append(result, target.Symbol)
	}
	return result
}

func canonicalPatternBranches() []string {
	targets := patternBranchElementProfile()
	result := make([]string, 0, len(targets))
	for _, target := range targets {
		result = append(result, target.Symbol)
	}
	return result
}

func canonicalSixtyPatternPillars(t *testing.T) []model.Pillar {
	t.Helper()
	stems, branches := canonicalPatternStems(), canonicalPatternBranches()
	result := make([]model.Pillar, 0, 60)
	for index := 0; index < 60; index++ {
		pillar := model.Pillar{Gan: stems[index%len(stems)], Zhi: branches[index%len(branches)]}
		if _, err := (tyme.SixtyCycle{}).FromName(pillar.Gan + pillar.Zhi); err != nil {
			t.Fatalf("canonical sixty-cycle pillar %d %s%s is invalid: %v", index, pillar.Gan, pillar.Zhi, err)
		}
		result = append(result, pillar)
	}
	return result
}

func canonicalBranchForPatternStem(t *testing.T, stem string) string {
	t.Helper()
	for _, branch := range canonicalPatternBranches() {
		if _, err := (tyme.SixtyCycle{}).FromName(stem + branch); err == nil {
			return branch
		}
	}
	t.Fatalf("no canonical branch for stem %s", stem)
	return ""
}

func patternBehaviorOutput(result *patternDetection) string {
	if result == nil {
		return ""
	}
	return result.PatternName
}

func simplePatternBehaviorRows(t *testing.T, ruleID string) []patternBehaviorTruthRow {
	t.Helper()
	detector := simplePatternDetectorByRuleID(t, ruleID)
	rows := make([]patternBehaviorTruthRow, 0)
	switch ruleID {
	case "pattern.lu.jianlu", "pattern.lu.yueren":
		for _, stem := range canonicalPatternStems() {
			for _, branch := range canonicalPatternBranches() {
				rows = append(rows, patternBehaviorTruthRow{
					Input: stem + "/" + branch,
					Output: patternBehaviorOutput(detector.detect(patternDetectorContext{
						dayGan: stem, monthBranch: branch,
					})),
				})
			}
		}
	case "pattern.lu.zhuanlu", "pattern.lu.riren", "pattern.aux.kuigang", "pattern.aux.ride":
		for _, pillar := range canonicalSixtyPatternPillars(t) {
			rows = append(rows, patternBehaviorTruthRow{
				Input: pillar.Gan + pillar.Zhi,
				Output: patternBehaviorOutput(detector.detect(patternDetectorContext{
					dayGan: pillar.Gan, dayZhi: pillar.Zhi,
				})),
			})
		}
	case "pattern.aux.jinshen":
		for _, hour := range canonicalSixtyPatternPillars(t) {
			pillars := []model.Pillar{{Gan: "甲", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"}, hour}
			rows = append(rows, patternBehaviorTruthRow{
				Input:  hour.Gan + hour.Zhi,
				Output: patternBehaviorOutput(detector.detect(patternDetectorContext{pillars: pillars})),
			})
		}
	case "pattern.aux.sanqi":
		stems := canonicalPatternStems()
		for _, year := range stems {
			for _, month := range stems {
				for _, day := range stems {
					for _, hour := range stems {
						sequence := []string{year, month, day, hour}
						pillars := make([]model.Pillar, 0, 4)
						for _, stem := range sequence {
							pillars = append(pillars, model.Pillar{Gan: stem, Zhi: canonicalBranchForPatternStem(t, stem)})
						}
						rows = append(rows, patternBehaviorTruthRow{
							Input:  strings.Join(sequence, "/"),
							Output: patternBehaviorOutput(detector.detect(patternDetectorContext{pillars: pillars})),
						})
					}
				}
			}
		}
	default:
		t.Fatalf("rule %s has no simple behavior domain", ruleID)
	}
	return rows
}

func patternBehaviorRowsSHA256(t *testing.T, rows []patternBehaviorTruthRow) string {
	t.Helper()
	payload, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func TestSimplePatternBehaviorManifestsBindExhaustiveTruthTables(t *testing.T) {
	ruleIDs := []string{
		"pattern.lu.jianlu", "pattern.lu.yueren", "pattern.lu.zhuanlu", "pattern.lu.riren",
		"pattern.aux.kuigang", "pattern.aux.ride", "pattern.aux.jinshen", "pattern.aux.sanqi",
	}
	seen := make(map[string]string, len(ruleIDs))
	for _, ruleID := range ruleIDs {
		manifest := patternDetectorBehaviorManifestProfile(ruleID)
		rows := simplePatternBehaviorRows(t, ruleID)
		matches := 0
		for _, row := range rows {
			if row.Output != "" {
				matches++
			}
		}
		digest := patternBehaviorRowsSHA256(t, rows)
		if manifest == nil || manifest.Scheme != "canonical_truth_table_v1" ||
			manifest.CaseCount != len(rows) || manifest.MatchCount != matches || manifest.SHA256 != digest {
			t.Errorf("%s behavior manifest:\n domain=%v\n cases=%d matches=%d sha256=%s", ruleID, manifest, len(rows), matches, digest)
		}
		if former, duplicate := seen[digest]; duplicate {
			t.Errorf("%s and %s share behavior SHA-256 %s", former, ruleID, digest)
		}
		seen[digest] = ruleID
	}
	if len(seen) != 8 {
		t.Fatalf("simple behavior manifest count = %d, want 8", len(seen))
	}
}

func TestSimplePatternBehaviorManifestProfilesAreIndependentAndBound(t *testing.T) {
	manifest := patternDetectorBehaviorManifestProfile("pattern.aux.sanqi")
	manifest.Domain, manifest.SHA256 = "mutated", "mutated"
	fresh := patternDetectorBehaviorManifestProfile("pattern.aux.sanqi")
	if fresh.Domain == "mutated" || fresh.SHA256 == "mutated" {
		t.Fatalf("fresh behavior manifest inherited mutation: %+v", fresh)
	}
	profile, ok := patternDetectorSemanticProfile("pattern.aux.sanqi")
	envelope, envelopeOK := profile.(patternDetectorSemanticEnvelope)
	if !ok || !envelopeOK || !reflect.DeepEqual(envelope.BehaviorManifest, fresh) {
		t.Fatalf("san-qi semantic envelope omits behavior manifest: %#v", profile)
	}
	for _, ruleID := range []string{"pattern.special.zhuanwang", "pattern.special.liangqi", "pattern.unknown"} {
		if got := patternDetectorBehaviorManifestProfile(ruleID); got != nil {
			t.Errorf("complex/unknown rule %s received simple behavior manifest: %+v", ruleID, got)
		}
	}
}

func TestSimplePatternBehaviorManifestMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"八个简单表型检测器的穷举结果仍分散在测试中",
			"canonical_truth_table_v1",
			"建禄与月刃各120例",
			"四个日柱表及金神各60例",
			"三奇10000例",
			"行为清单SHA-256",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
