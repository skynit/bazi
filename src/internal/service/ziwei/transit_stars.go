package ziwei

var transitChangQuBranch = [10][2]int{
	{5, 9},  // 甲: 昌巳、曲酉
	{6, 8},  // 乙: 昌午、曲申
	{8, 6},  // 丙: 昌申、曲午
	{9, 5},  // 丁: 昌酉、曲巳
	{8, 6},  // 戊: 昌申、曲午
	{9, 5},  // 己: 昌酉、曲巳
	{11, 3}, // 庚: 昌亥、曲卯
	{0, 2},  // 辛: 昌子、曲寅
	{2, 0},  // 壬: 昌寅、曲子
	{3, 11}, // 癸: 昌卯、曲亥
}

var transitStarLabels = map[string][10]string{
	"liunian": {"流魁", "流钺", "流昌", "流曲", "流禄", "流羊", "流陀", "流马", "流鸾", "流喜"},
	"liuyue":  {"月魁", "月钺", "月昌", "月曲", "月禄", "月羊", "月陀", "月马", "月鸾", "月喜"},
	"liuri":   {"日魁", "日钺", "日昌", "日曲", "日禄", "日羊", "日陀", "日马", "日鸾", "日喜"},
}

var dayunStarLabels = [10]string{"运魁", "运钺", "运昌", "运曲", "运禄", "运羊", "运陀", "运马", "运鸾", "运喜"}

func buildTransitStarDistribution(chart *ZiWeiChart, stem, branch int, scope string) [12][]string {
	var result [12][]string
	if chart == nil || stem < 0 || stem >= 10 || branch < 0 || branch >= 12 {
		return result
	}
	labels, ok := transitStarLabels[scope]
	if scope == "dayun" {
		labels, ok = dayunStarLabels, true
	}
	if !ok {
		return result
	}

	var byBranch [12][]string
	if scope == "liunian" {
		index := fixIndex(10 - branch)
		byBranch[index] = append(byBranch[index], "年解")
	}
	positions := [10]int{
		KuiYueTable[stem][0],
		KuiYueTable[stem][1],
		transitChangQuBranch[stem][0],
		transitChangQuBranch[stem][1],
		LucunBranchIdx[stem],
		QingyangIndex(stem),
		TuoluoIndex(stem),
		TianmaBranchIdx[branch],
		HongLuanIndex(branch),
		TianXiIndex(branch),
	}
	for i, position := range positions {
		byBranch[position] = append(byBranch[position], labels[i])
	}
	for i, palace := range chart.Palaces {
		branchIndex, exists := BranchIndex[palace.Branch]
		if !exists {
			result[i] = []string{}
			continue
		}
		result[i] = append([]string(nil), byBranch[branchIndex]...)
	}
	return result
}

func buildTransitFourHua(chart *ZiWeiChart, stem int) [12][]string {
	var result [12][]string
	if chart == nil || stem < 0 || stem >= 10 {
		return result
	}
	hua := calcFourHua(stem)
	for i, palace := range chart.Palaces {
		result[i] = []string{}
		for _, star := range palaceAllStarNames(palace) {
			if label, ok := hua[star]; ok {
				result[i] = append(result[i], star+label)
			}
		}
	}
	return result
}
