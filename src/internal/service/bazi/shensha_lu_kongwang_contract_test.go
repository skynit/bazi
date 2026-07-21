package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestSiDaKongWangUsesWholeSixtyCycleXun(t *testing.T) {
	for i := 0; i < 60; i++ {
		want := ""
		switch i / 10 {
		case 0, 3:
			want = "水"
		case 2, 5:
			want = "金"
		}
		gan, zhi := data.Gans[i%10], data.Zhis[i%12]
		if got := siDaKongWangElement(gan, zhi); got != want {
			t.Errorf("cycle %d %s%s empty element = %q, want %q", i, gan, zhi, got, want)
		}
	}
	if got := siDaKongWangElement("甲", "丑"); got != "" {
		t.Fatalf("invalid sixty-cycle pillar returned %q", got)
	}
}

func TestSiDaKongWangAttachesEveryMatchingNayinPillar(t *testing.T) {
	tests := []struct {
		name     string
		pillars  ShenShaPillars
		element  string
		hasEmpty bool
	}{
		{
			name: "non-jia day in jia-zi xun lacks water",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "丙", Zhi: "子"}, Month: model.Pillar{Gan: "丁", Zhi: "丑"},
				Day: model.Pillar{Gan: "乙", Zhi: "丑"}, Hour: model.Pillar{Gan: "戊", Zhi: "寅"},
			},
			element: "水", hasEmpty: true,
		},
		{
			name: "non-jia day in jia-shen xun lacks metal",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "乙", Zhi: "丑"},
				Day: model.Pillar{Gan: "乙", Zhi: "酉"}, Hour: model.Pillar{Gan: "丙", Zhi: "寅"},
			},
			element: "金", hasEmpty: true,
		},
		{
			name: "jia-xu xun has all five elements",
			pillars: ShenShaPillars{
				Year: model.Pillar{Gan: "甲", Zhi: "子"}, Month: model.Pillar{Gan: "乙", Zhi: "丑"},
				Day: model.Pillar{Gan: "乙", Zhi: "亥"}, Hour: model.Pillar{Gan: "丙", Zhi: "子"},
			},
			hasEmpty: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got ShenShaCalcResult
			branches := []string{tc.pillars.Year.Zhi, tc.pillars.Month.Zhi, tc.pillars.Day.Zhi, tc.pillars.Hour.Zhi}
			addDayExtra(tc.pillars, branches, &got)
			if !tc.hasEmpty {
				for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
					if hasShenShaName(bucket, "四大空亡") {
						t.Fatalf("unexpected four-great-void hit: %+v", got)
					}
				}
				return
			}
			assertExactShenShaInBucket(t, "year", got.Year, "四大空亡："+tc.element)
			assertExactShenShaInBucket(t, "month", got.Month, "四大空亡："+tc.element)
			if hasShenShaName(got.Day, "四大空亡") || hasShenShaName(got.Hour, "四大空亡") {
				t.Errorf("four-great-void leaked into nonmatching na-yin pillar: %+v", got)
			}
		})
	}
}

func TestAnLuRequiresNoVisibleLuAndAttachesToTargetPillars(t *testing.T) {
	withoutVisibleLu := ShenShaPillars{
		Year: model.Pillar{Gan: "乙", Zhi: "亥"}, Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day: model.Pillar{Gan: "甲", Zhi: "子"}, Hour: model.Pillar{Gan: "丁", Zhi: "亥"},
	}
	var got ShenShaCalcResult
	addDayExtra(withoutVisibleLu, []string{"亥", "辰", "子", "亥"}, &got)
	assertExactShenShaInBucket(t, "year", got.Year, "暗禄：亥")
	assertExactShenShaInBucket(t, "hour", got.Hour, "暗禄：亥")
	if hasShenShaName(got.Day, "暗禄") || hasShenShaName(got.Month, "暗禄") {
		t.Errorf("dark lu attached outside target pillars: %+v", got)
	}

	withVisibleLu := withoutVisibleLu
	withVisibleLu.Month.Zhi = "寅"
	got = ShenShaCalcResult{}
	addDayExtra(withVisibleLu, []string{"亥", "寅", "子", "亥"}, &got)
	for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
		if hasShenShaName(bucket, "暗禄") {
			t.Fatalf("visible lu did not suppress dark lu: %+v", got)
		}
	}
}

func TestJiaoLuRequiresMutualLuExchange(t *testing.T) {
	pillars := ShenShaPillars{
		Year: model.Pillar{Gan: "庚", Zhi: "寅"}, Month: model.Pillar{Gan: "丙", Zhi: "子"},
		Day: model.Pillar{Gan: "甲", Zhi: "申"}, Hour: model.Pillar{Gan: "丁", Zhi: "卯"},
	}
	var got ShenShaCalcResult
	addJiaoLu(pillars, &got)
	assertExactShenShaInBucket(t, "year", got.Year, "交禄：庚寅/甲申")
	assertExactShenShaInBucket(t, "day", got.Day, "交禄：庚寅/甲申")
	if hasShenShaName(got.Month, "交禄") || hasShenShaName(got.Hour, "交禄") {
		t.Errorf("mutual lu exchange leaked outside participating pillars: %+v", got)
	}

	oldGanHeFalsePositive := ShenShaPillars{
		Year: model.Pillar{Gan: "己", Zhi: "午"}, Month: model.Pillar{Gan: "丙", Zhi: "辰"},
		Day: model.Pillar{Gan: "甲", Zhi: "子"}, Hour: model.Pillar{Gan: "丁", Zhi: "卯"},
	}
	got = ShenShaCalcResult{}
	addJiaoLu(oldGanHeFalsePositive, &got)
	for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
		if hasShenShaName(bucket, "交禄") {
			t.Fatalf("stem-combination false positive survived: %+v", got)
		}
	}

	withoutDayPillar := ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "申"}, Month: model.Pillar{Gan: "庚", Zhi: "寅"},
		Day: model.Pillar{Gan: "乙", Zhi: "丑"}, Hour: model.Pillar{Gan: "丁", Zhi: "卯"},
	}
	got = ShenShaCalcResult{}
	addJiaoLu(withoutDayPillar, &got)
	assertExactShenShaInBucket(t, "year", got.Year, "交禄：甲申/庚寅")
	assertExactShenShaInBucket(t, "month", got.Month, "交禄：甲申/庚寅")
	if hasShenShaName(got.Day, "交禄") || hasShenShaName(got.Hour, "交禄") {
		t.Errorf("non-day mutual exchange leaked outside participating pillars: %+v", got)
	}
}

func TestUnsupportedLuAliasesAndGlobalDuplicatePathAreAbsent(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"addGlobalLuDerived", "天元暗禄", "飞禄"} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains unsupported lu path %q", forbidden)
		}
	}
}

func TestClassicalLuAndFourGreatVoidEvidenceIsRegistered(t *testing.T) {
	wants := map[string]string{
		"暗禄":   "第97页",
		"交禄":   "第96页",
		"四大空亡": "第109页",
	}
	for name, citation := range wants {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" || !strings.Contains(meta.Basis, citation) {
			t.Errorf("%s metadata = %+v, want located classical evidence", name, meta)
		}
	}
}

func TestShenShaClassicalSourceFingerprint(t *testing.T) {
	meta := DefaultRuleMeta()
	for _, table := range meta.Tables {
		if table.Key != "shensha" {
			continue
		}
		if len(table.Sources) != 2 {
			t.Fatalf("shen-sha sources = %+v, want two classical sources", table.Sources)
		}
		wants := map[string]map[string]string{
			"yuan_hai_zi_ping_local_pdf":  {"library/渊海子平.pdf": YuanHaiZiPingPDFSHA256},
			"san_ming_tong_hui_local_pdf": {"library/三命通会.pdf": SanMingTongHuiPDFSHA256},
		}
		for _, source := range table.Sources {
			files, ok := wants[source.ID]
			if !ok || source.Repository != "workspace://library" || source.Commit != "not_applicable" ||
				source.License != "not_recorded" || source.SourceTier != "classical_text_local" ||
				source.ValidationStatus != "text_located_not_expert_gold" || !reflect.DeepEqual(source.Files, files) {
				t.Fatalf("shen-sha classical source = %+v", source)
			}
			delete(wants, source.ID)
		}
		if len(wants) != 0 {
			t.Fatalf("missing shen-sha classical sources = %+v", wants)
		}
		return
	}
	t.Fatal("shen-sha rule table not found")
}

func TestShenShaLocalClassicalFilesMatchManifestFingerprints(t *testing.T) {
	for path, want := range map[string]string{
		"library/渊海子平.pdf": YuanHaiZiPingPDFSHA256,
		"library/三命通会.pdf": SanMingTongHuiPDFSHA256,
	} {
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

func hasShenShaName(items []string, want string) bool {
	for _, item := range items {
		if shenShaName(item) == want {
			return true
		}
	}
	return false
}
