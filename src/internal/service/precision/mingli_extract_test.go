package precision

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractMingLiCandidates(t *testing.T) {
	root := t.TempDir()
	bookDir := filepath.Join(root, "bazi", "五行精纪")
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `# 五行精纪

六丁日（丁丑、丁卯、丁巳、丁未、丁酉、丁亥）只是枚举，不应抽样。

【财库发金格】
例如毕状元的命造：己亥年、癸酉月、庚辰日、甲申时。（出自《紫虚局》）

戊午 乙丑 乙巳 庚辰
徐乐吾曰：副总统冯国璋命，咸丰八年十二月初四日辰时。取财生官旺。
`
	if err := os.WriteFile(filepath.Join(bookDir, "032.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	fixture, err := ExtractMingLiCandidates(MingLiExtractOptions{
		SourceDir:     root,
		System:        "bazi",
		MinConfidence: 0.50,
	})
	if err != nil {
		t.Fatalf("ExtractMingLiCandidates: %v", err)
	}
	if got, want := len(fixture.Cases), 2; got != want {
		t.Fatalf("cases=%d, want %d: %#v", got, want, fixture.Cases)
	}

	first := fixture.Cases[0]
	if first.Expected.YearPillar == "" || first.Expected.MonthPillar == "" || first.Expected.DayPillar == "" || first.Expected.HourPillar == "" {
		t.Fatalf("empty pillar in first candidate: %#v", first.Expected)
	}
	if first.Review.Status != "pending" {
		t.Fatalf("review status=%q, want pending", first.Review.Status)
	}

	foundLabelled := false
	foundRawDate := false
	for _, tc := range fixture.Cases {
		if tc.Evidence.Kind == "labelled_pillars" && tc.Expected.YearPillar == "己亥" {
			foundLabelled = true
		}
		if tc.Input.RawBirthDate == "咸丰八年十二月初四日辰时" {
			foundRawDate = true
		}
	}
	if !foundLabelled {
		t.Fatal("labelled pillar candidate was not extracted")
	}
	if !foundRawDate {
		t.Fatal("raw birth date was not captured from adjacent context")
	}
}
