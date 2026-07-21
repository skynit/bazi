package bazi

import (
	"fmt"
	"sort"
	"strings"
)

type ganRelationPillar struct {
	key, label, stem, branch string
}

func buildGanRelationGraph(pillars []ganRelationPillar) []GanRelation {
	relations := make([]GanRelation, 0, 12)
	for i := 0; i < len(pillars); i++ {
		for j := i + 1; j < len(pillars); j++ {
			a, b := pillars[i], pillars[j]
			if ganHe[a.stem] == b.stem {
				relations = append(relations, newGanCombineRelation(pillars, i, j))
			}
			if isGanFourClash(a.stem, b.stem) {
				relations = append(relations, newGanClashRelation(a, b))
			}
			if relation, ok := newGanElementRelation(a, b); ok {
				relations = append(relations, relation)
			}
		}
	}
	markGanCombineDisputes(relations)
	sort.SliceStable(relations, func(i, j int) bool {
		if relations[i].Priority != relations[j].Priority {
			return relations[i].Priority > relations[j].Priority
		}
		return relations[i].ID < relations[j].ID
	})
	return relations
}

func isGanFourClash(a, b string) bool {
	switch canonicalGanPair(a, b) {
	case "甲庚", "乙辛", "丙壬", "丁癸":
		return true
	default:
		return false
	}
}

func newGanClashRelation(a, b ganRelationPillar) GanRelation {
	pair := canonicalGanPair(a.stem, b.stem)
	return GanRelation{
		ID:      "stem.clash." + pair + "." + a.key + "-" + b.key,
		RuleID:  "stem.clash." + pair,
		Pillar1: a.label, Pillar2: b.label, Pillars: []string{a.label, b.label}, Stems: []string{a.stem, b.stem},
		Type: "天干相冲", Subtype: "四冲", Status: "observed",
		StructureStatus: "complete_structure", TransformationStatus: "not_applicable",
		Direction: "mutual", Proximity: ganRelationProximity(ganPillarOrder(a.key), ganPillarOrder(b.key)), Priority: 55,
		ConflictsWith: []string{}, DisputeReasons: []string{},
		Evidence: []string{
			"《三命通会》卷十一列甲庚、乙辛、丙壬、丁癸为阴阳不合、交差相畏",
			"lunar-java 0194eb4574f33ab056fe7cac62a9d8bf24272478 LunarUtil.CHONG_GAN_4 固定同四组（相关 Silver）",
		},
		Detail: fmt.Sprintf("%s%s与%s%s构成天干四冲结构；相克方向另由五行关系记录，本项不单独推断具体事件。", a.label, a.stem, b.label, b.stem),
	}
}

func newGanCombineRelation(pillars []ganRelationPillar, i, j int) GanRelation {
	a, b := pillars[i], pillars[j]
	pair := canonicalGanPair(a.stem, b.stem)
	target := ganHeHua[a.stem+b.stem]
	proximity := ganRelationProximity(i, j)
	evidence := ganTransformationEvidence(pillars, i, j, target)
	return GanRelation{
		ID:      "stem.combine." + pair + "." + a.key + "-" + b.key,
		RuleID:  "stem.combine." + pair,
		Pillar1: a.label, Pillar2: b.label, Pillars: []string{a.label, b.label}, Stems: []string{a.stem, b.stem},
		Type: "五合", Subtype: ganHeName[a.stem+b.stem], Status: "observed",
		StructureStatus: "complete_structure", TransformationStatus: "unadjudicated",
		TargetElement: target, Proximity: proximity, Priority: 60,
		ConflictsWith: []string{}, DisputeReasons: []string{},
		Evidence:               []string{fmt.Sprintf("%s%s与%s%s构成%s", a.label, a.stem, b.label, b.stem, ganHeName[a.stem+b.stem]), fmt.Sprintf("柱位关系为%s", proximity)},
		TransformationEvidence: &evidence,
		Detail:                 fmt.Sprintf("%s与%s构成天干五合结构，目标五行为%s；成化状态未裁决，仍需月令、透干、根气及克制条件的独立 Gold 规则。", a.stem, b.stem, target),
	}
}

func newGanElementRelation(a, b ganRelationPillar) (GanRelation, bool) {
	elementA, elementB := GanInfoOf(a.stem).elem, GanInfoOf(b.stem).elem
	if elementA == "" || elementB == "" {
		return GanRelation{}, false
	}
	relationType, direction, priority := "", "", 0
	if elementA == elementB {
		relationType, direction, priority = "比和", "mutual", 30
	} else if generatesElement(elementA, elementB) {
		relationType, direction, priority = "相生", a.key+"_to_"+b.key, 40
	} else if generatesElement(elementB, elementA) {
		relationType, direction, priority = "相生", b.key+"_to_"+a.key, 40
	} else if controlsElement(elementA, elementB) {
		relationType, direction, priority = "相克", a.key+"_to_"+b.key, 50
	} else if controlsElement(elementB, elementA) {
		relationType, direction, priority = "相克", b.key+"_to_"+a.key, 50
	}
	if relationType == "" {
		return GanRelation{}, false
	}
	proximity := ganRelationProximity(ganPillarOrder(a.key), ganPillarOrder(b.key))
	ruleID := "stem." + map[string]string{"比和": "peer", "相生": "generate", "相克": "control"}[relationType] + "." + elementA + elementB
	return GanRelation{
		ID: ruleID + "." + a.key + "-" + b.key, RuleID: ruleID,
		Pillar1: a.label, Pillar2: b.label, Pillars: []string{a.label, b.label}, Stems: []string{a.stem, b.stem},
		Type: relationType, Status: "observed", StructureStatus: "observed_relation", TransformationStatus: "not_applicable",
		Direction: direction, Proximity: proximity, Priority: priority,
		ConflictsWith: []string{}, DisputeReasons: []string{},
		Evidence: []string{fmt.Sprintf("%s属%s，%s属%s", a.stem, elementA, b.stem, elementB)},
		Detail:   fmt.Sprintf("%s%s与%s%s形成五行%s关系（方向：%s）；仅记录确定性生克结构，不单独推断具体事件。", a.label, a.stem, b.label, b.stem, relationType, direction),
	}, true
}

func ganTransformationEvidence(pillars []ganRelationPillar, combineA, combineB int, target string) GanTransformationEvidence {
	monthBranch, monthElement := "", ""
	for _, pillar := range pillars {
		if pillar.key == "month" {
			monthBranch = pillar.branch
			monthElement = zhiWuXing[pillar.branch]
			break
		}
	}
	targetExposed := false
	rootBranches := make([]string, 0, 4)
	for i, pillar := range pillars {
		if i != combineA && i != combineB && GanInfoOf(pillar.stem).elem == target {
			targetExposed = true
		}
		if containsBranch(zhiAllElements[pillar.branch], target) {
			rootBranches = append(rootBranches, pillar.label+pillar.branch)
		}
	}
	return GanTransformationEvidence{
		MonthBranch: monthBranch, MonthElement: monthElement, MonthSupportsTarget: monthElement == target,
		TargetStemExposed: targetExposed, TargetRootBranches: rootBranches,
		Note: "这些字段仅记录常见成化条件证据；当前 Profile 未用它们自动裁决成化。",
	}
}

func markGanCombineDisputes(relations []GanRelation) {
	degrees := make(map[string]int)
	for _, relation := range relations {
		if relation.Type != "五合" {
			continue
		}
		for _, pillar := range relation.Pillars {
			degrees[pillar]++
		}
	}
	for i := range relations {
		if relations[i].Type != "五合" {
			continue
		}
		for _, pillar := range relations[i].Pillars {
			if degrees[pillar] <= 1 {
				continue
			}
			relations[i].Status = "disputed"
			relations[i].TransformationStatus = "disputed"
			relations[i].DisputeReasons = appendUnique(relations[i].DisputeReasons, pillar+"同时参与多组五合，构成争合/妒合")
			for j := range relations {
				if i != j && relations[j].Type == "五合" && containsBranch(relations[j].Pillars, pillar) {
					relations[i].ConflictsWith = appendUnique(relations[i].ConflictsWith, relations[j].ID)
				}
			}
		}
		if relations[i].Status == "disputed" {
			relations[i].Detail += " 同盘存在争合/妒合，不作单一成化结论。"
			sort.Strings(relations[i].ConflictsWith)
			sort.Strings(relations[i].DisputeReasons)
		}
	}
}

func canonicalGanPair(a, b string) string {
	order := "甲乙丙丁戊己庚辛壬癸"
	if strings.Index(order, a) <= strings.Index(order, b) {
		return a + b
	}
	return b + a
}

func ganRelationProximity(a, b int) string {
	if a > b {
		a, b = b, a
	}
	if b-a == 1 {
		return "adjacent"
	}
	return "remote"
}

func ganPillarOrder(key string) int {
	return map[string]int{"year": 0, "month": 1, "day": 2, "hour": 3}[key]
}

func generatesElement(a, b string) bool {
	return map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}[a] == b
}

func controlsElement(a, b string) bool {
	return map[string]string{"木": "土", "土": "水", "水": "火", "火": "金", "金": "木"}[a] == b
}
