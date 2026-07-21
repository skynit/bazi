package bazi

import (
	"os"
	"strings"
	"testing"
)

func TestShenShaProductionHasNoDayOnlyCandidateEntrypoint(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if strings.Contains(text, "CalcDayShenShaOnly") {
		t.Fatal("day-only candidate table is still exposed as a calculation entrypoint")
	}
	if !strings.Contains(text, "func CalcShenShaByPillars(") {
		t.Fatal("authoritative four-pillar shen-sha entrypoint is missing")
	}
}
