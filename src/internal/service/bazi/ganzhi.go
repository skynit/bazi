package bazi

import (
	"fmt"
	"reflect"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

// ganHeHua maps each ordered heavenly-stem combination to its traditional
// target element. A complete pair does not by itself prove transformation.
var ganHeHua = map[string]string{
	"甲己": "土", "己甲": "土",
	"乙庚": "金", "庚乙": "金",
	"丙辛": "水", "辛丙": "水",
	"丁壬": "木", "壬丁": "木",
	"戊癸": "火", "癸戊": "火",
}

// zhiWuXing maps each earthly branch to its fixed five-element category.
var zhiWuXing = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木", "辰": "土", "巳": "火",
	"午": "火", "未": "土", "申": "金", "酉": "金", "戌": "土", "亥": "水",
}

var zhiLiuHai = map[string]string{
	"子": "未", "未": "子", "丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅", "卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申", "酉": "戌", "戌": "酉",
}

// zhiLiuPo follows the traditional six-break pairing table.
var zhiLiuPo = map[string]string{
	"子": "酉", "酉": "子",
	"丑": "辰", "辰": "丑",
	"寅": "亥", "亥": "寅",
	"卯": "午", "午": "卯",
	"巳": "申", "申": "巳",
	"未": "戌", "戌": "未",
}

var ganHeName = map[string]string{
	"甲己": "中正之合", "己甲": "中正之合",
	"乙庚": "仁义之合", "庚乙": "仁义之合",
	"丙辛": "威制之合", "辛丙": "威制之合",
	"丁壬": "淫慝之合", "壬丁": "淫慝之合",
	"戊癸": "无情之合", "癸戊": "无情之合",
}

type GanRelation struct {
	ID                     string                     `json:"id"`
	RuleID                 string                     `json:"rule_id"`
	Pillar1                string                     `json:"pillar1"`
	Pillar2                string                     `json:"pillar2"`
	Pillars                []string                   `json:"pillars"`
	Stems                  []string                   `json:"stems"`
	Type                   string                     `json:"type"`
	Subtype                string                     `json:"subtype,omitempty"`
	Status                 string                     `json:"status"`
	StructureStatus        string                     `json:"structure_status"`
	TransformationStatus   string                     `json:"transformation_status"`
	TargetElement          string                     `json:"target_element,omitempty"`
	Direction              string                     `json:"direction,omitempty"`
	Proximity              string                     `json:"proximity"`
	Priority               int                        `json:"priority"`
	ConflictsWith          []string                   `json:"conflicts_with"`
	DisputeReasons         []string                   `json:"dispute_reasons"`
	Evidence               []string                   `json:"evidence"`
	TransformationEvidence *GanTransformationEvidence `json:"transformation_evidence,omitempty"`
	Detail                 string                     `json:"detail"`
}

type GanTransformationEvidence struct {
	MonthBranch         string   `json:"month_branch"`
	MonthElement        string   `json:"month_element"`
	MonthSupportsTarget bool     `json:"month_supports_target"`
	TargetStemExposed   bool     `json:"target_stem_exposed"`
	TargetRootBranches  []string `json:"target_root_branches"`
	Note                string   `json:"note"`
}

type ZhiRelation struct {
	ID                   string   `json:"id"`
	RuleID               string   `json:"rule_id"`
	Pillar1              string   `json:"pillar1"`
	Pillar2              string   `json:"pillar2"`
	Pillars              []string `json:"pillars"`
	Branches             []string `json:"branches"`
	Type                 string   `json:"type"`
	Subtype              string   `json:"subtype,omitempty"`
	Status               string   `json:"status"`
	StructureStatus      string   `json:"structure_status"`
	TransformationStatus string   `json:"transformation_status"`
	TargetElement        string   `json:"target_element,omitempty"`
	Priority             int      `json:"priority"`
	ConflictsWith        []string `json:"conflicts_with"`
	DisputeReasons       []string `json:"dispute_reasons"`
	Evidence             []string `json:"evidence"`
	Detail               string   `json:"detail"`
}

type GanZhiAnalysis struct {
	GanRelations []GanRelation `json:"gan_relations"`
	ZhiRelations []ZhiRelation `json:"zhi_relations"`
}

const (
	labelYear  = "年柱"
	labelMonth = "月柱"
	labelDay   = "日柱"
	labelHour  = "时柱"
)

func CalcGanZhiAnalysis(year, month, day, hour model.Pillar) (GanZhiAnalysis, error) {
	for _, item := range []struct {
		name   string
		pillar model.Pillar
	}{
		{"year", year}, {"month", month}, {"day", day}, {"hour", hour},
	} {
		if _, err := (tyme.SixtyCycle{}).FromName(item.pillar.Gan + item.pillar.Zhi); err != nil {
			return GanZhiAnalysis{}, fmt.Errorf("invalid %s pillar %q: must be one of the sixty-cycle pairs", item.name, item.pillar.Gan+item.pillar.Zhi)
		}
	}
	pillars := []ganRelationPillar{
		{key: "year", label: labelYear, stem: year.Gan, branch: year.Zhi},
		{key: "month", label: labelMonth, stem: month.Gan, branch: month.Zhi},
		{key: "day", label: labelDay, stem: day.Gan, branch: day.Zhi},
		{key: "hour", label: labelHour, stem: hour.Gan, branch: hour.Zhi},
	}
	return GanZhiAnalysis{
		GanRelations: buildGanRelationGraph(pillars),
		ZhiRelations: buildZhiRelationGraph([]branchRelationPillar{
			{key: "year", label: labelYear, branch: year.Zhi},
			{key: "month", label: labelMonth, branch: month.Zhi},
			{key: "day", label: labelDay, branch: day.Zhi},
			{key: "hour", label: labelHour, branch: hour.Zhi},
		}),
	}, nil
}

// ValidGanZhiAnalysis requires the complete public relation graph to be exactly
// reproducible from the four stored pillars.
func ValidGanZhiAnalysis(actual GanZhiAnalysis, year, month, day, hour model.Pillar) bool {
	expected, err := CalcGanZhiAnalysis(year, month, day, hour)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(actual, expected)
}
