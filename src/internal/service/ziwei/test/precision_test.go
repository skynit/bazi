package ziwei_test

import (
	"bazi/internal/service/ziwei"
	"encoding/json"
	"os"
	"testing"
)

// ZiWeiTestCase is the legacy Bronze shape used by structural smoke tests.
// Its expected text is not a full-chart oracle.
type ZiWeiTestCase struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Gender   string `json:"gender"`
	Expected struct {
		Pattern            string            `json:"pattern"`
		FiveBureau         string            `json:"five_bureau"`
		MainStar           string            `json:"main_star"`
		KeyPalaces         map[string]string `json:"key_palaces"`
		OccupationTendency string            `json:"occupation_tendency"`
		LifeTendency       string            `json:"life_tendency"`
		RankTendency       string            `json:"rank_tendency"`
		WealthTendency     string            `json:"wealth_tendency"`
		Tendency           string            `json:"tendency"`
	} `json:"expected"`
}

type ZiWeiTestData struct {
	Version  string `json:"version"`
	Metadata struct {
		Tier             string `json:"tier"`
		ReviewStatus     string `json:"review_status"`
		QuarantineReason string `json:"quarantine_reason"`
	} `json:"metadata"`
	Cases []ZiWeiTestCase `json:"cases"`
}

// TestLegacyZiWeiFixtureRemainsQuarantined verifies governance and calculation
// coverage only. Fuzzy text agreement from this file must never be reported as
// accuracy because dates and full charts lack page-level adjudication.
func TestLegacyZiWeiFixtureRemainsQuarantined(t *testing.T) {
	data, err := loadZiWeiTestData("../../testdata/ziwei_cases.json")
	if err != nil {
		t.Fatalf("load legacy Ziwei fixtures: %v", err)
	}
	if data.Metadata.Tier != "bronze" || data.Metadata.ReviewStatus != "quarantined" || data.Metadata.QuarantineReason == "" {
		t.Fatalf("legacy fixture lost quarantine metadata: %+v", data.Metadata)
	}
	if len(data.Cases) == 0 {
		t.Fatal("legacy fixture unexpectedly contains no smoke cases")
	}

	service := ziwei.NewZiWeiService()
	ids := make(map[string]struct{}, len(data.Cases))
	for _, tc := range data.Cases {
		if tc.ID == "" || tc.Year <= 0 || tc.Month <= 0 || tc.Day <= 0 {
			t.Fatalf("invalid Bronze smoke input: %+v", tc)
		}
		if _, duplicate := ids[tc.ID]; duplicate {
			t.Fatalf("duplicate Bronze case ID %q", tc.ID)
		}
		ids[tc.ID] = struct{}{}
		if _, err := service.CalculateChart(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender); err != nil {
			t.Fatalf("Bronze smoke calculation %s failed: %v", tc.ID, err)
		}
	}
}

func loadZiWeiTestData(path string) (*ZiWeiTestData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ZiWeiTestData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}
