package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bazi/internal/service/precision"
)

func TestRunProducesBlockedAuditableReport(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-root", filepath.Clean("../..")}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr=%s", exitCode, stderr.String())
	}
	var report precision.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Version != "4.0" || report.ComparatorVersion != "precision-comparator-v1" ||
		!strings.HasPrefix(report.ComparatorHash, "sha256:") {
		t.Fatalf("report comparator identity is incomplete: %+v", report)
	}
	if report.PublicationStatus != "blocked" || report.PublishableChecks != 0 || len(report.ReleaseBlockers) == 0 {
		t.Fatalf("report lost its publication gate: %+v", report)
	}
}

func TestRunWritesRequestedOutputFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "precision-report.json")
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-root", filepath.Clean("../.."), "-out", out}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, stderr=%s", exitCode, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout is not empty when -out is used: %q", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report precision.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.ComparatorHash == "" || report.PublicationStatus != "blocked" {
		t.Fatalf("written report is incomplete: %+v", report)
	}
}
