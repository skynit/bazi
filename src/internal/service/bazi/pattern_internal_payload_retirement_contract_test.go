package bazi

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPatternDetectionCarriesOnlyCandidateIdentity(t *testing.T) {
	typeOfDetection := reflect.TypeOf(patternDetection{})
	wantFields := []string{"PatternName"}
	if typeOfDetection.NumField() != len(wantFields) {
		t.Fatalf("patternDetection field count = %d, want %d", typeOfDetection.NumField(), len(wantFields))
	}
	for i, want := range wantFields {
		if got := typeOfDetection.Field(i).Name; got != want {
			t.Errorf("patternDetection field %d = %s, want %s", i, got, want)
		}
	}
}

func TestPatternInternalInterpretationPayloadIsAbsent(t *testing.T) {
	source, err := os.ReadFile("pattern.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"Description", "FavorableElements", "UnfavorableElements", "SubType",
		"twoQiRelation", "喜食伤泄秀", "财官流通", "忌印比重叠",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("pattern production source still contains retired internal payload %q", forbidden)
		}
	}
}

func TestPatternInternalPayloadRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v6 的私有patternDetection",
			"Description、SubType、FavorableElements和UnfavorableElements",
			"专禄残留统一喜忌",
			"pattern-candidate-set-v7只保留PatternName和PatternType",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
