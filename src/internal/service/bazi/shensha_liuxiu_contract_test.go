package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestUnsupportedLiuXiuRiIsAbsentAcrossSixtyDays(t *testing.T) {
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		if specialDayContainsName(day, "六秀日") {
			t.Errorf("special day table still contains 六秀日 for %s", day)
		}
		result := calcSpecialDayFixture(t, day)
		assertShenShaNameAbsentEverywhere(t, result, "六秀日")
	}
}

func TestFormerNineLiuXiuDaysRemainNegative(t *testing.T) {
	former := []string{"戊子", "戊午", "辛酉", "丙午", "丙子", "丁丑", "丁未", "己未", "己丑"}
	if len(former) != 9 {
		t.Fatalf("former 六秀日 fixture count = %d, want 9", len(former))
	}
	for _, day := range former {
		result := calcSpecialDayFixture(t, day)
		assertShenShaNameAbsentEverywhere(t, result, "六秀日")
	}
}

func TestUnsupportedLiuXiuRiIsAbsentFromProductionSourceAndMetadata(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), `"六秀日"`) {
		t.Fatal("production shen-sha source still publishes 六秀日")
	}
	meta := LookupShenShaMeta("六秀日")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("removed 六秀日 metadata = %+v", meta)
	}
}

func TestUnlocatedSpecialDayResearchPDFHashes(t *testing.T) {
	wants := map[string]string{
		"library/三命通会.pdf":    "63eb2a85036ebbd360b815a58780edd242bda0fd1a5faaac7413e00c5f726d47",
		"library/三命通会_白话.pdf": "290bf0a2ea897e020a91ac8be6a38a92deed34cf6f9337676b9204a6cfd0468d",
		"library/渊海子平.pdf":    "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5",
		"library/滴天髓阐微.pdf":   "65c67d88421319fccbba23bce88d61d4ace288a7913edd6a10ebf3143e72a48b",
		"library/穷通宝鉴.pdf":    "1c7ef872809bd17ce55607343680489832d58a39b8e2e3584429484d2fd02219",
	}
	for path, want := range wants {
		raw, err := os.ReadFile("../../../../" + path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		if got := hex.EncodeToString(sum[:]); got != want {
			t.Errorf("%s SHA-256 = %s, want %s", path, got, want)
		}
	}
}
