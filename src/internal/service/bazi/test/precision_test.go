package bazi_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	. "bazi/internal/service/bazi"
)

// TestCase is the legacy fixture shape. Expected values are deliberately not
// compared here because the file is quarantined Bronze, not an adjudicated
// oracle.
type TestCase struct {
	ID          string `json:"id"`
	LegacyID    string `json:"legacy_id"`
	SourceIndex int    `json:"source_index"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	Day         int    `json:"day"`
	Hour        int    `json:"hour"`
	Minute      int    `json:"minute"`
	Gender      string `json:"gender"`
	Expected    struct {
		YearPillar   string `json:"year_pillar"`
		MonthPillar  string `json:"month_pillar"`
		DayPillar    string `json:"day_pillar"`
		HourPillar   string `json:"hour_pillar"`
		DayMaster    string `json:"day_master"`
		BodyStrength string `json:"body_strength"`
		Pattern      string `json:"pattern"`
		YongShen     string `json:"yong_shen"`
		XiShen       string `json:"xi_shen"`
		JiShen       string `json:"ji_shen"`
		TiaoHou      string `json:"tiao_hou"`
	} `json:"expected"`
	LegacyAnnotations struct {
		Description string `json:"description"`
	} `json:"legacy_annotations"`
}

type legacyFixtureMetadata struct {
	Tier             string `json:"tier"`
	ReviewStatus     string `json:"review_status"`
	QuarantineReason string `json:"quarantine_reason"`
}

type TestData struct {
	Version  string                `json:"version"`
	Metadata legacyFixtureMetadata `json:"metadata"`
	Cases    []TestCase            `json:"cases"`
}

func TestLegacyBaziDateFixtureRemainsQuarantined(t *testing.T) {
	data, err := loadTestData("../../testdata/bazi_date_gold_candidates.json")
	if err != nil {
		t.Fatal(err)
	}
	assertLegacyFixtureQuarantined(t, data.Metadata)
	if len(data.Cases) == 0 {
		t.Fatal("legacy date fixture unexpectedly contains no cases")
	}

	service := &BaziService{}
	ids := make(map[string]struct{}, len(data.Cases))
	for _, tc := range data.Cases {
		if tc.ID == "" || tc.Year <= 0 || tc.Month <= 0 || tc.Day <= 0 {
			t.Fatalf("invalid Bronze smoke input: %+v", tc)
		}
		if _, duplicate := ids[tc.ID]; duplicate {
			t.Fatalf("duplicate date-fixture ID %q", tc.ID)
		}
		ids[tc.ID] = struct{}{}
		result, err := service.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		if err != nil {
			t.Fatalf("Bronze smoke calculation %s failed: %v", tc.ID, err)
		}
		if result.Tiaohou == nil || result.Tiaohou.DepthEvidence.Status != "observed" {
			t.Fatalf("date calculation %s lacks factual Tiaohou depth evidence", tc.ID)
		}
	}
}

func loadTestData(path string) (*TestData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var data TestData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func assertLegacyFixtureQuarantined(t *testing.T, metadata legacyFixtureMetadata) {
	t.Helper()
	if strings.ToLower(metadata.Tier) != "bronze" || strings.ToLower(metadata.ReviewStatus) != "quarantined" || strings.TrimSpace(metadata.QuarantineReason) == "" {
		t.Fatalf("legacy fixture must remain quarantined Bronze: %+v", metadata)
	}
}

func pillarStr(p model.Pillar) string {
	return p.Gan + p.Zhi
}
