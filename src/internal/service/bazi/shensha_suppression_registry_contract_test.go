package bazi

import (
	"os"
	"strings"
	"testing"
)

func TestSuppressedHighRiskShenShaContainsOnlyAuditedSiFu(t *testing.T) {
	if len(suppressedHighRiskShenSha) != 1 {
		t.Fatalf("suppressedHighRiskShenSha = %v, want only 死符", suppressedHighRiskShenSha)
	}
	if _, ok := suppressedHighRiskShenSha["死符"]; !ok {
		t.Fatalf("suppressedHighRiskShenSha = %v, want 死符", suppressedHighRiskShenSha)
	}

	meta := LookupShenShaMeta("死符")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("死符 metadata = %+v, want observed/not_adjudicated", meta)
	}

	items := []string{"existing"}
	appendShenSha(&items, "死符", "巳")
	if len(items) != 1 || items[0] != "existing" {
		t.Fatalf("appendShenSha published suppressed 死符: %v", items)
	}
}

func TestRetiredHighRiskNamesAreNotProductionRules(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"天杀", "自缢煞", "产厄"} {
		if strings.Contains(string(source), "\""+name+"\"") {
			t.Errorf("production source still contains retired exact name %s", name)
		}
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}
