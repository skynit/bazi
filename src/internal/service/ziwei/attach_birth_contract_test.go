package ziwei

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestAttachBirthDataRequiresExactPublishedChartAndInput(t *testing.T) {
	svc := NewZiWeiService()
	fresh, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(fresh)
	if err != nil {
		t.Fatal(err)
	}
	var replayed ZiWeiChart
	if err := json.Unmarshal(payload, &replayed); err != nil {
		t.Fatal(err)
	}

	before, err := json.Marshal(&replayed)
	if err != nil {
		t.Fatal(err)
	}
	if err = svc.AttachBirthData(&replayed, 2003, 4, 15, 14, 1, "男"); err == nil {
		t.Fatal("attachment accepted mismatched birth minute")
	}
	after, err := json.Marshal(&replayed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("failed attachment mutated the published chart")
	}
	if err := svc.AttachBirthData(&replayed, 2003, 4, 15, 14, 0, "男"); err != nil {
		t.Fatalf("valid replay attachment failed: %v", err)
	}
	afterSuccess, err := json.Marshal(&replayed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterSuccess, before) {
		t.Fatal("successful cache authentication mutated the published chart")
	}
	if !svc.ChartMatchesInputProfile(&replayed, DefaultProfileID, 2003, 4, 15, 14, 0, "男") {
		t.Fatal("cache authentication changed the published chart")
	}
}

func TestAttachBirthDataRejectsInvalidPublishedBranchesWithoutFallback(t *testing.T) {
	svc := NewZiWeiService()
	fresh, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	for _, mutate := range []func(*ZiWeiChart){
		func(chart *ZiWeiChart) { chart.EarthlyBranchOfSoulPalace = "未知命宫支" },
		func(chart *ZiWeiChart) { chart.EarthlyBranchOfBodyPalace = "未知身宫支" },
	} {
		invalid := *fresh
		mutate(&invalid)
		restampZiWeiChartContentHash(t, &invalid)
		before, err := json.Marshal(&invalid)
		if err != nil {
			t.Fatal(err)
		}
		if err := svc.AttachBirthData(&invalid, 2003, 4, 15, 14, 0, "男"); err == nil {
			t.Fatalf("attachment accepted invalid published branches: %+v", invalid)
		}
		after, err := json.Marshal(&invalid)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(after, before) {
			t.Fatal("failed attachment mutated invalid published chart")
		}
	}
}

func mustPublishedBirthData(t testing.TB, chart *ZiWeiChart) *BirthData {
	t.Helper()
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok {
		t.Fatal("published chart did not restore birth data")
	}
	return birth
}
