package bazi

import (
	"fmt"
	"sort"
	"strings"
)

type branchRelationPillar struct {
	key, label, branch string
}

type branchGroupRule struct {
	ruleID, relationType, subtype, targetElement string
	branches                                     []string
	priority                                     int
}

var branchSanHeRules = []branchGroupRule{
	{"branch.sanhe.water", "三合局", "申子辰", "水", []string{"申", "子", "辰"}, 65},
	{"branch.sanhe.wood", "三合局", "亥卯未", "木", []string{"亥", "卯", "未"}, 65},
	{"branch.sanhe.fire", "三合局", "寅午戌", "火", []string{"寅", "午", "戌"}, 65},
	{"branch.sanhe.metal", "三合局", "巳酉丑", "金", []string{"巳", "酉", "丑"}, 65},
}

var branchSanHuiRules = []branchGroupRule{
	{"branch.sanhui.wood", "三会局", "寅卯辰", "木", []string{"寅", "卯", "辰"}, 66},
	{"branch.sanhui.fire", "三会局", "巳午未", "火", []string{"巳", "午", "未"}, 66},
	{"branch.sanhui.metal", "三会局", "申酉戌", "金", []string{"申", "酉", "戌"}, 66},
	{"branch.sanhui.water", "三会局", "亥子丑", "水", []string{"亥", "子", "丑"}, 66},
}

var branchSanXingRules = []branchGroupRule{
	{"branch.sanxing.wuen", "三刑", "无恩之刑", "", []string{"寅", "巳", "申"}, 95},
	{"branch.sanxing.shishi", "三刑", "恃势之刑", "", []string{"丑", "戌", "未"}, 95},
}

func buildZhiRelationGraph(pillars []branchRelationPillar) []ZhiRelation {
	relations := make([]ZhiRelation, 0, 20)
	completeGroups := make(map[string]bool)
	for _, group := range append(append([]branchGroupRule{}, branchSanHeRules...), branchSanHuiRules...) {
		if relation, ok := completeBranchGroupRelation(pillars, group); ok {
			relations = append(relations, relation)
			completeGroups[group.ruleID] = true
		}
	}
	completeXing := make(map[string]bool)
	for _, group := range branchSanXingRules {
		if relation, ok := completeBranchGroupRelation(pillars, group); ok {
			relation.TransformationStatus = "not_applicable"
			relation.Evidence = append(relation.Evidence, threePunishmentNamingEvidence(group.ruleID)...)
			relation.Detail = fmt.Sprintf("%s三支齐备，构成%s完整结构；名称采用《三命通会》口径，《渊海子平》对两组三刑的无恩/恃势名称互换；同时出现的冲、合、害、破仍分别保留。", strings.Join(group.branches, ""), group.subtype)
			relations = append(relations, relation)
			completeXing[group.ruleID] = true
		}
	}

	for i := 0; i < len(pillars); i++ {
		for j := i + 1; j < len(pillars); j++ {
			a, b := pillars[i], pillars[j]
			if a.branch == b.branch {
				relations = append(relations, newPairZhiRelation(a, b, "伏吟", "", "branch.fuyin."+a.branch, "", 50, "observed", "not_applicable"))
			}
			if zhiLiuChong[a.branch] == b.branch {
				relations = append(relations, newPairZhiRelation(a, b, "六冲", "", "branch.clash."+canonicalBranchPair(a.branch, b.branch), "", 100, "observed", "not_applicable"))
			}
			if zhiLiuHe[a.branch] == b.branch {
				target := liuHeTargetElement(a.branch, b.branch)
				relations = append(relations, newPairZhiRelation(a, b, "六合", "", "branch.liuhe."+canonicalBranchPair(a.branch, b.branch), target, 60, "complete", "unadjudicated"))
			}
			if zhiLiuHai[a.branch] == b.branch {
				relations = append(relations, newPairZhiRelation(a, b, "六害", "", "branch.harm."+canonicalBranchPair(a.branch, b.branch), "", 80, "observed", "not_applicable"))
			}
			if subtype, groupID := branchPunishment(a.branch, b.branch); subtype != "" && !completeXing[groupID] {
				relation := newPairZhiRelation(a, b, "相刑", subtype, "branch.punish."+canonicalBranchPair(a.branch, b.branch), "", 90, "observed", "not_applicable")
				if groupID != "" {
					relation.Evidence = append(relation.Evidence, threePunishmentNamingEvidence(groupID)...)
					relation.Detail += " 无恩/恃势名称采用《三命通会》口径，《渊海子平》使用相反异称。"
				}
				relations = append(relations, relation)
			}
			if zhiLiuPo[a.branch] == b.branch {
				relations = append(relations, newPairZhiRelation(a, b, "六破", "", "branch.break."+canonicalBranchPair(a.branch, b.branch), "", 70, "observed", "not_applicable"))
			}
			if relation, ok := partialBranchGroupRelation(a, b, branchSanHeRules, completeGroups); ok {
				relations = append(relations, relation)
			}
			if relation, ok := partialBranchGroupRelation(a, b, branchSanHuiRules, completeGroups); ok {
				relations = append(relations, relation)
			}
		}
	}

	markBranchRelationDisputes(relations)
	sort.SliceStable(relations, func(i, j int) bool {
		if relations[i].Priority != relations[j].Priority {
			return relations[i].Priority > relations[j].Priority
		}
		return relations[i].ID < relations[j].ID
	})
	return relations
}

func completeBranchGroupRelation(pillars []branchRelationPillar, group branchGroupRule) (ZhiRelation, bool) {
	present := make(map[string]bool, 3)
	members := make([]branchRelationPillar, 0, 4)
	for _, pillar := range pillars {
		if containsBranch(group.branches, pillar.branch) {
			present[pillar.branch] = true
			members = append(members, pillar)
		}
	}
	for _, branch := range group.branches {
		if !present[branch] {
			return ZhiRelation{}, false
		}
	}
	labels, branches, keys := branchMemberFields(members)
	return ZhiRelation{
		ID: group.ruleID + "." + strings.Join(keys, "-"), RuleID: group.ruleID,
		Pillar1: labels[0], Pillar2: labels[1], Pillars: labels, Branches: branches,
		Type: group.relationType, Subtype: group.subtype, Status: "complete",
		StructureStatus: "complete_structure", TransformationStatus: "unadjudicated",
		TargetElement: group.targetElement, Priority: group.priority,
		ConflictsWith: []string{}, DisputeReasons: []string{},
		Evidence: []string{fmt.Sprintf("%s三支齐备", strings.Join(group.branches, ""))},
		Detail:   fmt.Sprintf("%s三支齐备，形成%s完整结构；是否成化为%s仍需月令、透干、强弱及冲刑条件裁决。", strings.Join(group.branches, ""), group.relationType, group.targetElement),
	}, true
}

func partialBranchGroupRelation(a, b branchRelationPillar, groups []branchGroupRule, complete map[string]bool) (ZhiRelation, bool) {
	for _, group := range groups {
		if complete[group.ruleID] || !containsBranch(group.branches, a.branch) || !containsBranch(group.branches, b.branch) || a.branch == b.branch {
			continue
		}
		relationType, priority := "半会", 40
		if group.relationType == "三合局" {
			relationType, priority = "拱合", 30
			if a.branch == group.branches[1] || b.branch == group.branches[1] {
				relationType, priority = "半合", 35
			}
		}
		relation := newPairZhiRelation(a, b, relationType, group.subtype, group.ruleID+".partial", group.targetElement, priority, "partial", "not_applicable")
		relation.StructureStatus = "partial_structure"
		relation.Detail = fmt.Sprintf("%s与%s同属%s%s，但当前只见两支，仅记%s，不宣称完整成局或成化。", a.branch, b.branch, group.subtype, group.relationType, relationType)
		return relation, true
	}
	return ZhiRelation{}, false
}

func newPairZhiRelation(a, b branchRelationPillar, relationType, subtype, ruleID, target string, priority int, status, transformation string) ZhiRelation {
	id := ruleID + "." + a.key + "-" + b.key
	return ZhiRelation{
		ID: id, RuleID: ruleID, Pillar1: a.label, Pillar2: b.label,
		Pillars: []string{a.label, b.label}, Branches: []string{a.branch, b.branch},
		Type: relationType, Subtype: subtype, Status: status, StructureStatus: status + "_structure",
		TransformationStatus: transformation, TargetElement: target, Priority: priority,
		ConflictsWith: []string{}, DisputeReasons: []string{},
		Evidence: []string{fmt.Sprintf("%s%s与%s%s命中%s", a.label, a.branch, b.label, b.branch, relationType)},
		Detail:   fmt.Sprintf("%s%s与%s%s构成%s结构；仅记录规则命中，不据此单独推断具体事件。", a.label, a.branch, b.label, b.branch, relationType),
	}
}

func markBranchRelationDisputes(relations []ZhiRelation) {
	for i := range relations {
		if !isCombinationRelation(relations[i].Type) {
			continue
		}
		for j := range relations {
			if i == j || !isNegativeRelation(relations[j].Type) || !stringSlicesIntersect(relations[i].Pillars, relations[j].Pillars) {
				continue
			}
			relations[i].ConflictsWith = appendUnique(relations[i].ConflictsWith, relations[j].ID)
			relations[j].ConflictsWith = appendUnique(relations[j].ConflictsWith, relations[i].ID)
			reason := fmt.Sprintf("%s同时受%s结构影响", relations[i].Type, relations[j].Type)
			relations[i].DisputeReasons = appendUnique(relations[i].DisputeReasons, reason)
		}
		if len(relations[i].ConflictsWith) > 0 {
			relations[i].Status = "disputed"
			if relations[i].TransformationStatus == "unadjudicated" {
				relations[i].TransformationStatus = "disputed"
			}
			relations[i].Detail += " 同盘存在冲、刑、害或破，关系不作静默取舍。"
		}
		sort.Strings(relations[i].ConflictsWith)
		sort.Strings(relations[i].DisputeReasons)
	}
}

func branchPunishment(a, b string) (string, string) {
	if a == b && containsBranch([]string{"辰", "午", "酉", "亥"}, a) {
		return "自刑", ""
	}
	if sameBranchSetPair(a, b, []string{"子", "卯"}) {
		return "无礼之刑", ""
	}
	if pairInGroup(a, b, []string{"寅", "巳", "申"}) {
		return "无恩之刑", "branch.sanxing.wuen"
	}
	if pairInGroup(a, b, []string{"丑", "戌", "未"}) {
		return "恃势之刑", "branch.sanxing.shishi"
	}
	return "", ""
}

func threePunishmentNamingEvidence(ruleID string) []string {
	switch ruleID {
	case "branch.sanxing.wuen":
		return []string{
			"《三命通会》PDF第83-84页：寅巳申为无恩之刑",
			"《渊海子平》PDF第32页异称：寅巳申为恃势之刑",
		}
	case "branch.sanxing.shishi":
		return []string{
			"《三命通会》PDF第84页：丑戌未为恃势之刑",
			"《渊海子平》PDF第32页异称：丑戌未为无恩之刑",
		}
	default:
		return nil
	}
}

func liuHeTargetElement(a, b string) string {
	pair := canonicalBranchPair(a, b)
	return map[string]string{"子丑": "土", "寅亥": "木", "卯戌": "火", "辰酉": "金", "巳申": "水", "午未": "土"}[pair]
}

func canonicalBranchPair(a, b string) string {
	order := "子丑寅卯辰巳午未申酉戌亥"
	if strings.Index(order, a) <= strings.Index(order, b) {
		return a + b
	}
	return b + a
}

func branchMemberFields(members []branchRelationPillar) ([]string, []string, []string) {
	labels, branches, keys := make([]string, 0, len(members)), make([]string, 0, len(members)), make([]string, 0, len(members))
	for _, member := range members {
		labels = append(labels, member.label)
		branches = append(branches, member.branch)
		keys = append(keys, member.key)
	}
	return labels, branches, keys
}

func containsBranch(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func pairInGroup(a, b string, group []string) bool {
	return a != b && containsBranch(group, a) && containsBranch(group, b)
}
func sameBranchSetPair(a, b string, pair []string) bool { return pairInGroup(a, b, pair) }

func isCombinationRelation(relationType string) bool {
	return containsBranch([]string{"六合", "半合", "拱合", "半会", "三合局", "三会局"}, relationType)
}

func isNegativeRelation(relationType string) bool {
	return containsBranch([]string{"六冲", "相刑", "三刑", "六害", "六破"}, relationType)
}

func stringSlicesIntersect(a, b []string) bool {
	for _, left := range a {
		for _, right := range b {
			if left == right {
				return true
			}
		}
	}
	return false
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
