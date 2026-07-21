package bazi

import (
	"encoding/json"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestRelationGraphDoesNotInferLifeEvents(t *testing.T) {
	tests := []struct {
		name    string
		pillars [4]model.Pillar
	}{
		{
			name: "punishment_and_harm",
			pillars: [4]model.Pillar{
				{Gan: "甲", Zhi: "寅"}, {Gan: "己", Zhi: "巳"},
				{Gan: "庚", Zhi: "申"}, {Gan: "乙", Zhi: "亥"},
			},
		},
		{
			name: "combine_clash_break",
			pillars: [4]model.Pillar{
				{Gan: "甲", Zhi: "子"}, {Gan: "庚", Zhi: "午"},
				{Gan: "乙", Zhi: "酉"}, {Gan: "辛", Zhi: "丑"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis, err := CalcGanZhiAnalysis(tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3])
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(analysis)
			if err != nil {
				t.Fatal(err)
			}
			text := string(payload)
			for _, forbidden := range []string{
				"牢狱", "第三者", "婚姻质量", "晚景安稳", "子女缘佳",
				"忘恩负义", "自我折磨", "祖上庇荫", "事业转型不顺利",
			} {
				if strings.Contains(text, forbidden) {
					t.Fatalf("relation graph inferred %q: %s", forbidden, payload)
				}
			}
		})
	}
}
