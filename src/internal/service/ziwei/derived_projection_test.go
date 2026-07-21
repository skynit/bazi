package ziwei

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPublishedAdjectiveStarsSurviveJSONReplay(t *testing.T) {
	chart := calculateProjectionFixture(t)
	want := adjectiveStarsByBranch(chart)
	cached := roundTripProjectionFixture(t, chart)
	if got := adjectiveStarsByBranch(cached); !reflect.DeepEqual(got, want) {
		t.Fatalf("cached chart adjective stars diverged from published palace fields:\ngot=%v\nwant=%v", got, want)
	}
}

func TestPublishedTwelveShenSurviveJSONReplay(t *testing.T) {
	chart := calculateProjectionFixture(t)
	want := twelveShenByBranch(chart)
	cached := roundTripProjectionFixture(t, chart)
	if got := twelveShenByBranch(cached); !reflect.DeepEqual(got, want) {
		t.Fatalf("cached chart twelve-shen diverged from published palace fields:\ngot=%v\nwant=%v", got, want)
	}
}

func calculateProjectionFixture(t *testing.T) *ZiWeiChart {
	t.Helper()
	chart, err := NewZiWeiService().CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	return chart
}

func roundTripProjectionFixture(t *testing.T, chart *ZiWeiChart) *ZiWeiChart {
	t.Helper()
	raw, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	var cached ZiWeiChart
	if err := json.Unmarshal(raw, &cached); err != nil {
		t.Fatal(err)
	}
	return &cached
}

func adjectiveStarsByBranch(chart *ZiWeiChart) map[int][]string {
	result := make(map[int][]string, len(chart.Palaces))
	for i := 0; i < 12; i++ {
		result[i] = []string{}
	}
	for _, palace := range chart.Palaces {
		idx := BranchIndex[palace.Branch]
		result[idx] = append([]string{}, palace.AdjectiveStars...)
	}
	return result
}

func twelveShenByBranch(chart *ZiWeiChart) [12]struct {
	Changsheng, Boshi, Jiangqian, Suiqian string
} {
	var result [12]struct {
		Changsheng, Boshi, Jiangqian, Suiqian string
	}
	for _, palace := range chart.Palaces {
		result[BranchIndex[palace.Branch]] = struct {
			Changsheng, Boshi, Jiangqian, Suiqian string
		}{palace.Changsheng12, palace.Boshi12, palace.JiangQian12, palace.SuiQian12}
	}
	return result
}
