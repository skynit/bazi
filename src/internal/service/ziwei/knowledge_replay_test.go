package ziwei

import (
	"reflect"
	"testing"
)

func TestKnowledgeConsumers_JSONReplayMatchesFreshChart(t *testing.T) {
	chartA, chartB := calculateKnowledgeFixtures(t)

	replayedA := roundTripProjectionFixture(t, chartA)
	replayedB := roundTripProjectionFixture(t, chartB)

	if want, got := DetectLocalPatterns(chartA), DetectLocalPatterns(replayedA); !reflect.DeepEqual(got, want) {
		t.Fatalf("cached pattern detection depends on omitted runtime fields:\ngot=%v\nwant=%v", got, want)
	}
	if want, got := analyzeHeming(chartA, chartB), analyzeHeming(replayedA, replayedB); !reflect.DeepEqual(got, want) {
		t.Fatalf("cached heming analysis depends on omitted runtime fields:\ngot=%+v\nwant=%+v", got, want)
	}
}

func TestAnalyzeHeming_RejectsInvalidPublishedContracts(t *testing.T) {
	chartA, chartB := calculateKnowledgeFixtures(t)

	missingHash := *chartA
	missingHash.ContentHash = ""
	if got := analyzeHeming(&missingHash, chartB); got != nil {
		t.Fatalf("missing content hash must be rejected: %+v", got)
	}

	invalidFingerprint := *chartA
	invalidFingerprint.InputFingerprint = "invalid"
	restampZiWeiChartContentHash(t, &invalidFingerprint)
	if got := analyzeHeming(&invalidFingerprint, chartB); got != nil {
		t.Fatalf("invalid input fingerprint must be rejected: %+v", got)
	}

	nonCanonicalInput := *chartA
	nonCanonicalInput.CalculationInput.Gender = "invalid"
	nonCanonicalInput.InputFingerprint = ziweiInputFingerprint(nonCanonicalInput.CalculationInput)
	restampZiWeiChartContentHash(t, &nonCanonicalInput)
	if got := analyzeHeming(&nonCanonicalInput, chartB); got != nil {
		t.Fatalf("non-canonical calculation input must be rejected: %+v", got)
	}
}

func calculateKnowledgeFixtures(t *testing.T) (*ZiWeiChart, *ZiWeiChart) {
	t.Helper()
	svc := NewZiWeiService()
	chartA, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	chartB, err := svc.CalculateChart(1992, 9, 8, 9, 0, "女")
	if err != nil {
		t.Fatal(err)
	}
	return chartA, chartB
}
