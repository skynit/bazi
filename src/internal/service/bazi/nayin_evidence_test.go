package bazi

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

func TestNaYinEvidenceUsesOnlyPillarGanZhi(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		key     string
		ganZhi  string
		name    string
		element string
	}{
		{key: "year", ganZhi: "甲子", name: "海中金", element: "金"},
		{key: "month", ganZhi: "丙寅", name: "炉中火", element: "火"},
		{key: "day", ganZhi: "戊辰", name: "大林木", element: "木"},
		{key: "hour", ganZhi: "庚申", name: "石榴木", element: "木"},
	}
	for _, tc := range tests {
		got := result.NaYin[tc.key]
		if got.GanZhi != tc.ganZhi || got.Name != tc.name || got.Element != tc.element {
			t.Fatalf("%s na-yin = %+v, want %s/%s/%s", tc.key, got, tc.ganZhi, tc.name, tc.element)
		}
		if !ValidNaYinEvidence(got, string([]rune(tc.ganZhi)[0]), string([]rune(tc.ganZhi)[1])) {
			t.Fatalf("%s na-yin evidence is not valid: %+v", tc.key, got)
		}
	}
}

func TestNaYinEvidenceJSONHasOnlyFactualFields(t *testing.T) {
	evidence := observeNaYin("甲", "子")
	payload, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(payload, &fields); err != nil {
		t.Fatal(err)
	}
	gotKeys := make([]string, 0, len(fields))
	for key := range fields {
		gotKeys = append(gotKeys, key)
	}
	sort.Strings(gotKeys)
	wantKeys := []string{"basis", "element", "gan_zhi", "name", "rule_id", "status"}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("na-yin JSON fields = %v, want %v: %s", gotKeys, wantKeys, payload)
	}
	if evidence.Status != "observed" || evidence.Basis != "pillar_gan_zhi" {
		t.Fatalf("na-yin evidence metadata = %+v", evidence)
	}
}

func TestInvalidPillarDoesNotProduceObservedNaYin(t *testing.T) {
	got := observeNaYin("甲", "无")
	if got.Status != "unavailable" || ValidNaYinEvidence(got, "甲", "无") {
		t.Fatalf("invalid pillar produced valid na-yin evidence: %+v", got)
	}
}

func TestNaYinRuntimeIgnoresMutableLegacyTables(t *testing.T) {
	firePillars := ShenShaPillars{
		Year:   model.Pillar{Gan: "丙", Zhi: "寅"},
		Month:  model.Pillar{Gan: "甲", Zhi: "戌"},
		Day:    model.Pillar{Gan: "乙", Zhi: "亥"},
		Hour:   model.Pillar{Gan: "乙", Zhi: "巳"},
		Gender: model.GenderMale,
	}
	waterPillars := ShenShaPillars{
		Year:   model.Pillar{Gan: "丙", Zhi: "子"},
		Month:  model.Pillar{Gan: "丙", Zhi: "辰"},
		Day:    model.Pillar{Gan: "甲", Zhi: "子"},
		Hour:   model.Pillar{Gan: "丁", Zhi: "巳"},
		Gender: model.GenderMale,
	}
	fireBefore, err := CalcShenShaByPillars(firePillars)
	if err != nil {
		t.Fatal(err)
	}
	waterBefore, err := CalcShenShaByPillars(waterPillars)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"学堂", "词馆", "天罗"} {
		if !shenShaResultHasName(fireBefore, name) {
			t.Fatalf("fire fixture does not exercise %s: %+v", name, fireBefore)
		}
	}
	for _, name := range []string{"地网", "四大空亡"} {
		if !shenShaResultHasName(waterBefore, name) {
			t.Fatalf("water fixture does not exercise %s: %+v", name, waterBefore)
		}
	}

	wantEvidence := observeNaYin("甲", "申")
	wantMingGong := buildMingGongDetail("甲午")
	legacyNayin := data.Nayin
	legacyNaYinMap := data.NaYinMap
	legacyMingGongShenSha := data.MingGongShenShaByZhi
	t.Cleanup(func() {
		data.Nayin = legacyNayin
		data.NaYinMap = legacyNaYinMap
		data.MingGongShenShaByZhi = legacyMingGongShenSha
	})
	data.Nayin = [10][12]string{}
	data.NaYinMap = map[string]data.NaYinEntry{}
	data.MingGongShenShaByZhi = [12]string{}

	if got := observeNaYin("甲", "申"); got != wantEvidence || got.Name != "井泉水" || got.Element != "水" {
		t.Fatalf("legacy-table pollution changed na-yin evidence: got %+v, want %+v", got, wantEvidence)
	}
	if got := buildMingGongDetail("甲午"); got != wantMingGong || got.Nayin != "砂中金" || got.ShenSha != "天福" {
		t.Fatalf("legacy-table pollution changed ming-gong detail: got %+v, want %+v", got, wantMingGong)
	}
	for index := 0; index < 60; index++ {
		cycle := tyme.SixtyCycle{}.FromIndex(index)
		got := observeNaYin(cycle.GetHeavenStem().GetName(), cycle.GetEarthBranch().GetName())
		wantName := projectNaYinNameForTest(cycle.GetSound().GetName())
		wantRunes := []rune(wantName)
		wantElement := string(wantRunes[len(wantRunes)-1])
		if got.Status != "observed" || got.Name != wantName || got.Element != wantElement {
			t.Errorf("sixty-cycle %s na-yin = %+v, want %s/%s", cycle.GetName(), got, wantName, wantElement)
		}
	}

	fireAfter, err := CalcShenShaByPillars(firePillars)
	if err != nil {
		t.Fatal(err)
	}
	waterAfter, err := CalcShenShaByPillars(waterPillars)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(fireAfter, fireBefore) || !reflect.DeepEqual(waterAfter, waterBefore) {
		t.Fatalf("legacy-table pollution changed na-yin shen-sha results:\nfire before=%+v\nfire after=%+v\nwater before=%+v\nwater after=%+v",
			fireBefore, fireAfter, waterBefore, waterAfter)
	}
}

func projectNaYinNameForTest(name string) string {
	switch name {
	case "沙中金":
		return "砂中金"
	case "沙中土":
		return "砂中土"
	case "泉中水":
		return "井泉水"
	default:
		return name
	}
}

func shenShaResultHasName(result ShenShaCalcResult, name string) bool {
	for _, bucket := range [][]string{result.Year, result.Month, result.Day, result.Hour, result.Global} {
		for _, item := range bucket {
			if item == name || strings.HasPrefix(item, name+"：") {
				return true
			}
		}
	}
	return false
}
