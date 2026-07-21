package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEnhancedSanfangProjectsAuthoritativeFourHua(t *testing.T) {
	chart := calculateProjectionFixture(t)
	assertEnhancedSanfangFourHua(t, chart)

	raw, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	var cached ZiWeiChart
	if err := json.Unmarshal(raw, &cached); err != nil {
		t.Fatal(err)
	}
	assertEnhancedSanfangFourHua(t, &cached)
}

func assertEnhancedSanfangFourHua(t *testing.T, chart *ZiWeiChart) {
	t.Helper()
	for palaceIdx := range chart.Palaces {
		oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
		palaceName := chart.Palaces[palaceIdx].Name

		wantOpposite := suffixFourHua(chart.Palaces[oppositeIdx].FourHua, "照"+palaceName)
		wantTrine := append(
			suffixFourHua(chart.Palaces[trine1Idx].FourHua, "拱"+palaceName),
			suffixFourHua(chart.Palaces[trine2Idx].FourHua, "拱"+palaceName)...,
		)
		got := getEnhancedSanfang(chart, palaceIdx)
		if got.OppositeSihua != strings.Join(wantOpposite, "、") {
			t.Errorf("%s opposite four-hua = %q, want %q", palaceName, got.OppositeSihua, strings.Join(wantOpposite, "、"))
		}
		if got.TrineSihua != strings.Join(wantTrine, "、") {
			t.Errorf("%s trine four-hua = %q, want %q", palaceName, got.TrineSihua, strings.Join(wantTrine, "、"))
		}
	}
}

func suffixFourHua(fourHua []string, suffix string) []string {
	result := make([]string, 0, len(fourHua))
	for _, hua := range fourHua {
		result = append(result, hua+suffix)
	}
	return result
}
