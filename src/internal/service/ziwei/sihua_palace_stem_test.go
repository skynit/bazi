package ziwei

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPalaceStemSihuaDistinguishesSelfMutagenFromNatalFourHua(t *testing.T) {
	chart := palaceStemSihuaFixture()

	chain := analyzeSihuaChain(chart)
	if chain == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}

	if len(chain.HuaLu) != 1 {
		t.Fatalf("hua_lu = %+v, want one palace-stem flight", chain.HuaLu)
	}
	assertSihuaFlight(t, chain.HuaLu[0], "廉贞", "化禄", "命宫", "甲", "命宫", true)

	if len(chain.HuaQuan) != 1 {
		t.Fatalf("hua_quan = %+v, want one palace-stem flight", chain.HuaQuan)
	}
	assertSihuaFlight(t, chain.HuaQuan[0], "破军", "化权", "命宫", "甲", "兄弟", false)

	if len(chain.HuaJi) != 1 {
		t.Fatalf("hua_ji = %+v, want one palace-stem flight", chain.HuaJi)
	}
	assertSihuaFlight(t, chain.HuaJi[0], "太阳", "化忌", "命宫", "甲", "夫妻", false)

	selfMutagens := detectSelfMutagens(chart)
	wantSelfMutagens := []SelfMutagenResult{{
		SihuaProjectionSemantics: sihuaProjectionSemantics(),
		Palace:                   "命宫",
		PalaceStem:               "甲",
		TransformedStar:          "廉贞",
		HuaType:                  "化禄",
		StructureStatus:          "same_palace_transformation",
		IsSelfMutagen:            true,
	}}
	if !reflect.DeepEqual(selfMutagens, wantSelfMutagens) {
		t.Fatalf("self mutagens = %+v, want %+v", selfMutagens, wantSelfMutagens)
	}

	for _, result := range selfMutagens {
		if result.TransformedStar == "太阳" {
			t.Fatalf("natal four_hua was misreported as a self mutagen: %+v", result)
		}
	}
}

func TestSihuaProjectionContractDoesNotExposePseudoMetricsOrEffects(t *testing.T) {
	chart := palaceStemSihuaFixture()
	payload, err := json.Marshal(map[string]any{
		"natal":  buildFlyingStarAnalysisFromChart(chart),
		"flight": analyzeSihuaChain(chart),
		"self":   detectSelfMutagens(chart),
	})
	if err != nil {
		t.Fatalf("marshal sihua projection contract: %v", err)
	}

	serialized := string(payload)
	for _, forbidden := range []string{
		`"effect"`, `"chain_depth"`, `"total_chain_depth"`,
		`"star_affinity"`, `"key_mutagens"`, `"mutagen_type"`,
	} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("legacy sihua pseudo metric or interpretation leaked into JSON: %s", forbidden)
		}
	}
	for _, required := range []string{
		`"placement_basis":"deterministic_rule_projection"`,
		`"validation_status":"cross_checked_not_gold"`,
		`"is_outcome_conclusion":false`,
		`"analysis_kind":"direct_palace_stem_four_hua_flights"`,
	} {
		if !strings.Contains(serialized, required) {
			t.Fatalf("sihua projection semantics missing from JSON: %s", required)
		}
	}
}

func TestPalaceStemSihuaSurvivesJSONRoundTrip(t *testing.T) {
	chart := palaceStemSihuaFixture()
	wantChain := analyzeSihuaChain(chart)
	wantSelfMutagens := detectSelfMutagens(chart)

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("marshal chart: %v", err)
	}
	var replayed ZiWeiChart
	if err := json.Unmarshal(data, &replayed); err != nil {
		t.Fatalf("unmarshal chart: %v", err)
	}

	if got := analyzeSihuaChain(&replayed); !reflect.DeepEqual(got, wantChain) {
		t.Fatalf("JSON-replayed sihua chain = %+v, want %+v", got, wantChain)
	}
	if got := detectSelfMutagens(&replayed); !reflect.DeepEqual(got, wantSelfMutagens) {
		t.Fatalf("JSON-replayed self mutagens = %+v, want %+v", got, wantSelfMutagens)
	}
}

func TestPalaceStemSihuaFullChartUsesAllTwelvePalaceStems(t *testing.T) {
	chart, err := NewZiWeiService().CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart: %v", err)
	}

	want := analyzeSihuaChain(chart)
	assertFullPalaceStemFlights(t, chart, want)

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("marshal chart: %v", err)
	}
	var replayed ZiWeiChart
	if err := json.Unmarshal(data, &replayed); err != nil {
		t.Fatalf("unmarshal chart: %v", err)
	}
	got := analyzeSihuaChain(&replayed)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON-replayed full-chart sihua chain differs: got %+v, want %+v", got, want)
	}
	assertFullPalaceStemFlights(t, &replayed, got)
}

func palaceStemSihuaFixture() *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i := range chart.Palaces {
		chart.Palaces[i].Name = PALACE_NAMES[i]
		chart.Palaces[i].Branch = BranchNames[i]
	}

	// 甲干命宫飞出：廉贞化禄留在命宫，破军化权飞入兄弟，
	// 太阳化忌飞入夫妻。只有第一条属于宫干自化。
	chart.Palaces[0].HeavenlyStem = "甲"
	chart.Palaces[0].Stars = []StarOutput{{Name: "廉贞", Type: "major"}}
	chart.Palaces[1].Stars = []StarOutput{{Name: "破军", Type: "major"}}
	chart.Palaces[2].Stars = []StarOutput{{Name: "太阳", Type: "major"}}

	// four_hua 是本命年干四化的落宫投影，不能作为宫干自化依据。
	chart.Palaces[2].FourHua = []string{"太阳化忌"}
	return chart
}

func assertSihuaFlight(t *testing.T, got SihuaChainItem, star, huaType, sourcePalace, sourceStem, targetPalace string, self bool) {
	t.Helper()
	if got.TransformedStar != star || got.HuaType != huaType ||
		got.SourcePalace != sourcePalace || got.SourcePalaceStem != sourceStem ||
		got.TargetPalace != targetPalace || got.IsSelfMutagen != self {
		t.Errorf("flight = %+v, want star=%s hua=%s source=%s(%s) target=%s self=%t", got, star, huaType, sourcePalace, sourceStem, targetPalace, self)
	}
	wantScope := "cross_palace"
	if self {
		wantScope = "same_palace"
	}
	if got.FlightScope != wantScope {
		t.Errorf("flight metadata = %+v, want scope=%s", got, wantScope)
	}
	assertSihuaProjectionSemantics(t, got.SihuaProjectionSemantics)
}

func assertSihuaProjectionSemantics(t *testing.T, got SihuaProjectionSemantics) {
	t.Helper()
	if got != sihuaProjectionSemantics() {
		t.Errorf("sihua projection semantics = %+v, want %+v", got, sihuaProjectionSemantics())
	}
}

func assertFullPalaceStemFlights(t *testing.T, chart *ZiWeiChart, chain *SihuaChainResult) {
	t.Helper()
	if chain == nil {
		t.Fatal("AnalyzeSihuaChain returned nil")
	}
	groups := [4][]SihuaChainItem{chain.HuaLu, chain.HuaQuan, chain.HuaKe, chain.HuaJi}
	starToPalace := buildStarPalaceIndex(chart)
	for huaIdx, items := range groups {
		if len(items) != len(chart.Palaces) {
			t.Fatalf("%s flights = %d, want %d", SiHuaLabels[huaIdx], len(items), len(chart.Palaces))
		}
		for palaceIdx, palace := range chart.Palaces {
			stemIdx, ok := StemIndex[palace.HeavenlyStem]
			if !ok {
				t.Fatalf("palace %s has invalid heavenly stem %q", palace.Name, palace.HeavenlyStem)
			}
			star := SiHuaTable[stemIdx][huaIdx]
			targetIdx, ok := starToPalace[star]
			if !ok {
				t.Fatalf("transformed star %s is absent from chart", star)
			}
			assertSihuaFlight(t, items[palaceIdx], star, SiHuaLabels[huaIdx], palace.Name, palace.HeavenlyStem, chart.Palaces[targetIdx].Name, palaceIdx == targetIdx)
		}
	}
}
