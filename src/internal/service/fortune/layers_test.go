package fortune

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"github.com/6tail/tyme4go/tyme"
)

func TestBuildFortuneLayersWithDefaultCalculators(t *testing.T) {
	svc := &bazipkg.BaziService{}
	chart, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars failed: %v", err)
	}
	chart.DaYunInfo.Pillars = []model.Pillar{
		{Gan: "乙", Zhi: "卯"}, {Gan: "丙", Zhi: "辰"},
		{Gan: "丁", Zhi: "巳"}, {Gan: "戊", Zhi: "午"},
		{Gan: "己", Zhi: "未"}, {Gan: "庚", Zhi: "申"},
		{Gan: "辛", Zhi: "酉"}, {Gan: "壬", Zhi: "戌"},
	}

	queryDate := time.Date(2026, 6, 16, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	layers := BuildFortuneLayers(chart, queryDate, 1990)
	if layers.RuleVersion != bazipkg.RuleVersion || layers.School != bazipkg.RuleSchool {
		t.Fatalf("unexpected layer rule meta: %+v", layers)
	}
	checks := map[string]string{
		"dayun":   layers.DaYun.Key,
		"liunian": layers.LiuNian.Key,
		"liuyue":  layers.LiuYue.Key,
		"xiaoyun": layers.XiaoYun.Key,
	}
	for want, got := range checks {
		if got != want {
			t.Fatalf("layer key = %q, want %q", got, want)
		}
	}
	if layers.LiuNian.Year != 2026 {
		t.Fatalf("liunian year = %d, want 2026", layers.LiuNian.Year)
	}
	if len(layers.InterLayerRelations) < 3 {
		t.Fatalf("大运、流年、流月核心层间关系缺失: %+v", layers.InterLayerRelations)
	}
	wantPairs := map[string]bool{
		"流年天干->大运天干": false,
		"流月天干->流年天干": false,
		"流月天干->大运天干": false,
	}
	for _, relation := range layers.InterLayerRelations {
		key := relation.Source + "->" + relation.Target
		if _, ok := wantPairs[key]; ok {
			wantPairs[key] = true
		}
		if relation.Basis != "period_layer_stem_pair" && relation.Basis != "period_layer_branch_pair" {
			t.Fatalf("unexpected inter-layer relation basis: %+v", relation)
		}
		if relation.Status != "observed" || relation.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("inter-layer relation must remain structural evidence: %+v", relation)
		}
	}
	for pair, observed := range wantPairs {
		if !observed {
			t.Errorf("missing core inter-layer stem relation %s: %+v", pair, layers.InterLayerRelations)
		}
	}
	for _, layer := range []model.FortuneLayer{layers.DaYun, layers.LiuNian, layers.LiuYue, layers.XiaoYun} {
		if layer.RuleID == "" || layer.Status != "observed" || layer.InterpretationStatus != "not_adjudicated" || layer.Basis == "" {
			t.Fatalf("layer evidence metadata is incomplete: %+v", layer)
		}
		if layer.TenGod.RuleID == "" || layer.TenGod.Status != "observed" || layer.TenGod.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("layer ten-god evidence metadata is incomplete: %+v", layer)
		}
		for _, relation := range layer.Relations {
			if relation.RuleID == "" || relation.SourceValue == "" || relation.TargetValue == "" || relation.Basis == "" || relation.Status != "observed" || relation.InterpretationStatus != "not_adjudicated" {
				t.Fatalf("layer relation evidence metadata is incomplete: %+v", relation)
			}
		}
		for _, activation := range layer.ShenShaDetails {
			if activation.RuleID == "" || activation.Basis == "" || activation.Status != "observed" || activation.InterpretationStatus != "not_adjudicated" {
				t.Fatalf("layer shen-sha evidence metadata is incomplete: %+v", activation)
			}
		}
	}
}

func TestXiaoYunMatchesTyme4GoNominalAge(t *testing.T) {
	queryDate := time.Date(2026, 6, 16, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	for _, tc := range []struct {
		name       string
		gender     string
		timeGender tyme.Gender
	}{
		{name: "阳年男命顺推", gender: "MALE", timeGender: tyme.MAN},
		{name: "阳年女命逆推", gender: "FEMALE", timeGender: tyme.WOMAN},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := (&bazipkg.BaziService{}).Calculate(1990, 6, 15, 8, 0, tc.gender)
			if err != nil {
				t.Fatal(err)
			}
			birth, err := tyme.SolarTime{}.FromYmdHms(1990, 6, 15, 8, 0, 0)
			if err != nil {
				t.Fatal(err)
			}
			start := tyme.ChildLimit{}.FromSolarTime(*birth, tc.timeGender).GetStartFortune()
			want := start.Next(queryDate.Year() - start.GetSixtyCycleYear().GetYear()).GetName()

			got := buildXiaoYunLayer(chart, queryDate, 1990)
			if got.Pillar != want {
				t.Fatalf("小运 = %s, want tyme4go %s", got.Pillar, want)
			}
			if got.Age != 37 || got.RuleID != "fortune.layer.xiaoyun-v3" ||
				got.Basis != "tyme4go_fortune_hour_pillar_direction_and_nominal_age" {
				t.Fatalf("小运虚岁证据不完整: %+v", got)
			}
		})
	}
}

func TestLayerRelationsIncludeQueryDayBranch(t *testing.T) {
	chart := &bazipkg.BaziResult{
		YearPillar:  model.Pillar{Gan: "庚", Zhi: "午"},
		MonthPillar: model.Pillar{Gan: "壬", Zhi: "申"},
		DayPillar:   model.Pillar{Gan: "甲", Zhi: "子"},
		HourPillar:  model.Pillar{Gan: "戊", Zhi: "辰"},
	}
	queryDate := time.Date(2026, 6, 16, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	today, err := getDayPillar(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
	if err != nil {
		t.Fatal(err)
	}
	relations := layerRelations(today.Gan, today.Zhi, chart, queryDate)
	found := false
	for _, relation := range relations {
		if relation.Target == "查询日支" && relation.SourceValue == today.Zhi &&
			relation.TargetValue == today.Zhi && relation.Type == "same" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("周期地支与查询日支的同支事实缺失: today=%+v relations=%+v", today, relations)
	}
}

func TestLayerBranchRelationsPreserveConcurrentStructures(t *testing.T) {
	relations := layerBranchRelations("周期地支", "巳", "日支", "申")
	want := map[string]bool{"punish": false, "break": false, "combine": false}
	for _, relation := range relations {
		if _, ok := want[relation.Type]; ok {
			want[relation.Type] = true
		}
		if !strings.HasPrefix(relation.RuleID, "fortune.layer-relation.branch-v3.") ||
			relation.Basis != "period_branch_and_target_branch_all_structures" {
			t.Fatalf("复合地支关系缺少可复核依据: %+v", relation)
		}
	}
	for relationType, found := range want {
		if !found {
			t.Errorf("巳申关系缺少 %s: %+v", relationType, relations)
		}
	}
}

func TestLayerBranchRelationsKeepReversePunishmentAndSelfPunishment(t *testing.T) {
	for _, tc := range []struct {
		source, target string
		want           []string
	}{
		{source: "巳", target: "寅", want: []string{"punish", "harm"}},
		{source: "戌", target: "丑", want: []string{"punish"}},
		{source: "辰", target: "辰", want: []string{"same", "punish"}},
	} {
		relations := layerBranchRelations("周期地支", tc.source, "日支", tc.target)
		for _, wantType := range tc.want {
			found := false
			for _, relation := range relations {
				if relation.Type == wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s%s缺少%s: %+v", tc.source, tc.target, wantType, relations)
			}
		}
	}
}

func TestLayerStemRelationsPreserveSpecialAndElementRelations(t *testing.T) {
	for _, tc := range []struct {
		source, target string
		want           []string
	}{
		{source: "己", target: "甲", want: []string{"five_combine", "woKe"}},
		{source: "庚", target: "甲", want: []string{"clash", "keWo"}},
	} {
		relations := layerStemRelations("周期天干", tc.source, "日干", tc.target)
		if len(relations) != len(tc.want) {
			t.Fatalf("%s%s天干关系数量 = %d, want %d: %+v", tc.source, tc.target, len(relations), len(tc.want), relations)
		}
		counts := make(map[string]int, len(relations))
		for _, relation := range relations {
			counts[relation.Type]++
		}
		for _, wantType := range tc.want {
			if counts[wantType] != 1 {
				t.Errorf("%s%s的%s关系数量 = %d, want 1: %+v", tc.source, tc.target, wantType, counts[wantType], relations)
				continue
			}
			for _, relation := range relations {
				if relation.Type != wantType {
					continue
				}
				if !strings.HasPrefix(relation.RuleID, "fortune.layer-relation.stem-v3.") ||
					relation.Basis != "period_stem_and_target_stem_all_structures" {
					t.Fatalf("复合天干关系缺少可复核依据: %+v", relation)
				}
			}
		}
	}
}

func TestFortuneLayersJSONDoesNotExposeJudgmentFields(t *testing.T) {
	layers := model.FortuneLayerSet{
		DaYun: model.FortuneLayer{
			RuleID:               "fortune.layer.dayun-v2",
			Key:                  "dayun",
			InterpretationStatus: "not_adjudicated",
		},
		InterLayerRelations: []model.FortuneLayerRelation{},
	}
	payload, err := json.Marshal(layers)
	if err != nil {
		t.Fatalf("marshal fortune layers: %v", err)
	}
	text := string(payload)
	for _, forbidden := range []string{
		`"favorable"`, `"is_favorable"`, `"score"`, `"description"`,
		`"evidence"`, `"element_change"`, `"activated_shen_sha"`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("fortune layer JSON contains prohibited field %q: %s", forbidden, text)
		}
	}
	for _, required := range []string{`"rule_id"`, `"interpretation_status":"not_adjudicated"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("fortune layer JSON missing evidence marker %q: %s", required, text)
		}
	}
}
