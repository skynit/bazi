package service

import (
	"fmt"

	"github.com/6tail/tyme4go/tyme"
	"bazi/internal/model"
)

// BaziService calculates BaZi (八字) birth charts using tyme4go.
type BaziService struct{}

// BaziResult holds the complete BaZi calculation output.
type BaziResult struct {
	YearPillar   model.Pillar        `json:"year_pillar"`
	MonthPillar  model.Pillar        `json:"month_pillar"`
	DayPillar    model.Pillar        `json:"day_pillar"`
	HourPillar   model.Pillar        `json:"hour_pillar"`
	FiveElements map[string]int      `json:"five_elements"`
	TenGods      map[string]string   `json:"ten_gods"`
	NaYin        map[string]string   `json:"na_yin"`
	HiddenStems  map[string][]string `json:"hidden_stems"`
	DaYunInfo    DaYunInfo           `json:"da_yun_info"`
	ClashHarmony []ClashRelation     `json:"clash_harmony"`
}

// DaYunInfo describes the major fortune cycle (大运).
type DaYunInfo struct {
	StartAge  int            `json:"start_age"`
	Direction string         `json:"direction"`
	Pillars   []model.Pillar `json:"pillars"`
}

// ClashRelation describes a clash/harmony relation between two pillars.
type ClashRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
}

// Calculate computes a full BaZi chart.
func (s *BaziService) Calculate(year, month, day, hour, minute int, gender string) (*BaziResult, error) {
	tymeGender, err := toTymeGender(gender)
	if err != nil {
		return nil, err
	}

	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, hour, minute, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid birth time: %w", err)
	}

	ec := st.GetLunarHour().GetEightChar()

	result := &BaziResult{}

	// --- four pillars ---
	result.YearPillar = pillarFromSixtyCycle(ec.GetYear())
	result.MonthPillar = pillarFromSixtyCycle(ec.GetMonth())
	result.DayPillar = pillarFromSixtyCycle(ec.GetDay())
	result.HourPillar = pillarFromSixtyCycle(ec.GetHour())

	// --- five elements scores ---
	result.FiveElements = calcFiveElements(&ec)

	// --- ten gods ---
	result.TenGods = calcTenGods(&ec)

	// --- na yin ---
	result.NaYin = calcNaYin(&ec)

	// --- hidden stems ---
	result.HiddenStems = calcHiddenStems(&ec)

	// --- da yun ---
	result.DaYunInfo = calcDaYun(st, tymeGender)

	// --- clash / harmony ---
	result.ClashHarmony = calcClashHarmony(&ec)

	return result, nil
}

// --- helpers ---------------------------------------------------------------

func pillarFromSixtyCycle(sc tyme.SixtyCycle) model.Pillar {
	return model.Pillar{
		Gan: sc.GetHeavenStem().GetName(),
		Zhi: sc.GetEarthBranch().GetName(),
	}
}

func calcFiveElements(ec *tyme.EightChar) map[string]int {
	scores := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}

	pillars := [](func() tyme.SixtyCycle){
		ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour,
	}
	for _, fn := range pillars {
		sc := fn()
		// heavenly stem: 5 points
		elem := sc.GetHeavenStem().GetElement().GetName()
		scores[elem] += 5

		// earthly branch hidden stems: main=3, middle=2, residual=1
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			weight := 1
			switch hhs.GetType() {
			case tyme.MAIN:
				weight = 3
			case tyme.MIDDLE:
				weight = 2
			}
			elem := hhs.GetHeavenStem().GetElement().GetName()
			scores[elem] += weight
		}
	}
	return scores
}

func calcTenGods(ec *tyme.EightChar) map[string]string {
	dayStem := ec.GetDay().GetHeavenStem()
	return map[string]string{
		"year":  dayStem.GetTenStar(ec.GetYear().GetHeavenStem()).GetName(),
		"month": dayStem.GetTenStar(ec.GetMonth().GetHeavenStem()).GetName(),
		"day":   "日主",
		"hour":  dayStem.GetTenStar(ec.GetHour().GetHeavenStem()).GetName(),
	}
}

func calcNaYin(ec *tyme.EightChar) map[string]string {
	return map[string]string{
		"year":  ec.GetYear().GetSound().GetName(),
		"month": ec.GetMonth().GetSound().GetName(),
		"day":   ec.GetDay().GetSound().GetName(),
		"hour":  ec.GetHour().GetSound().GetName(),
	}
}

func calcHiddenStems(ec *tyme.EightChar) map[string][]string {
	result := make(map[string][]string, 4)
	pillars := map[string]func() tyme.SixtyCycle{
		"year": ec.GetYear, "month": ec.GetMonth,
		"day": ec.GetDay, "hour": ec.GetHour,
	}
	for name, fn := range pillars {
		sc := fn()
		var stems []string
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			label := hhs.GetHeavenStem().GetName()
			switch hhs.GetType() {
			case tyme.MAIN:
				// keep as-is
			case tyme.MIDDLE:
				label += "(中)"
			case tyme.RESIDUAL:
				label += "(余)"
			}
			stems = append(stems, label)
		}
		result[name] = stems
	}
	return result
}

func calcDaYun(st *tyme.SolarTime, gender tyme.Gender) DaYunInfo {
	cl := tyme.ChildLimit{}.FromSolarTime(*st, gender)

	dir := "逆排"
	if cl.IsForward() {
		dir = "顺排"
	}

	daYun := DaYunInfo{
		StartAge:  cl.GetYearCount(),
		Direction: dir,
	}

	df := cl.GetStartDecadeFortune()
	for i := 0; i < 8; i++ {
		cur := df.Next(i)
		sx := cur.GetSixtyCycle()
		daYun.Pillars = append(daYun.Pillars, model.Pillar{
			Gan: sx.GetHeavenStem().GetName(),
			Zhi: sx.GetEarthBranch().GetName(),
		})
	}
	return daYun
}

// --- clash / harmony detection ---------------------------------------------

type pillarPair struct {
	name   string
	branch tyme.EarthBranch
}

func calcClashHarmony(ec *tyme.EightChar) []ClashRelation {
	pairs := []pillarPair{
		{"年柱", ec.GetYear().GetEarthBranch()},
		{"月柱", ec.GetMonth().GetEarthBranch()},
		{"日柱", ec.GetDay().GetEarthBranch()},
		{"时柱", ec.GetHour().GetEarthBranch()},
	}

	var relations []ClashRelation

	// pairwise
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			a, b := pairs[i], pairs[j]

			if a.branch.GetOpposite().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六冲"})
			}
			if a.branch.GetCombine().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六合"})
			}
			if a.branch.GetHarm().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六害"})
			}

			// 三刑
			if t := tortureType(a.branch, b.branch); t != "" {
				relations = append(relations, ClashRelation{a.name, b.name, t})
			}
		}
	}

	// 三会: check all 4 branches for 3-branch groups
	relations = append(relations, detectTripleMeetings(pairs)...)

	return relations
}

// tortureType checks if two branches form a 三刑 relation.
func tortureType(a, b tyme.EarthBranch) string {
	defer func() { recover() }() // guard invalid index panics

	aName := a.GetName()
	bName := b.GetName()

	// 无礼之刑: 子-卯
	if (aName == "子" && bName == "卯") || (aName == "卯" && bName == "子") {
		return "无礼之刑"
	}
	// 无恩之刑: 丑-戌, 戌-未, 未-丑
	wuEn := [][]string{{"丑", "戌"}, {"戌", "未"}, {"未", "丑"}}
	for _, pair := range wuEn {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "无恩之刑"
		}
	}
	// 恃势之刑: 寅-巳, 巳-申, 申-寅
	shiShi := [][]string{{"寅", "巳"}, {"巳", "申"}, {"申", "寅"}}
	for _, pair := range shiShi {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "恃势之刑"
		}
	}
	// 自刑: 辰-辰, 午-午, 酉-酉, 亥-亥
	selfTorture := map[string]bool{"辰": true, "午": true, "酉": true, "亥": true}
	if aName == bName && selfTorture[aName] {
		return "自刑"
	}
	return ""
}

// detectTripleMeetings detects 三会 (three-branch meeting of same direction element).
func detectTripleMeetings(pairs []pillarPair) []ClashRelation {
	// 三会局: 寅卯辰(木), 巳午未(火), 申酉戌(金), 亥子丑(水)
	tripleGroups := [][]string{
		{"寅", "卯", "辰"}, // 东方木
		{"巳", "午", "未"}, // 南方火
		{"申", "酉", "戌"}, // 西方金
		{"亥", "子", "丑"}, // 北方水
	}

	var relations []ClashRelation
	branchIndex := make(map[string]int)
	for i, p := range pairs {
		n := p.branch.GetName()
		branchIndex[n] = i
	}

	seen := make(map[string]bool) // deduplicate

	for _, group := range tripleGroups {
		var matched []int
		for _, b := range group {
			if idx, ok := branchIndex[b]; ok {
				matched = append(matched, idx)
			}
		}
		if len(matched) >= 3 {
			// generate pairwise relations among the three
			for i := 0; i < len(matched); i++ {
				for j := i + 1; j < len(matched); j++ {
					pi, pj := pairs[matched[i]], pairs[matched[j]]
					key := pi.name + "<>" + pj.name + "<>三会"
					if key2 := pj.name + "<>" + pi.name + "<>三会"; seen[key2] {
						continue
					}
					if !seen[key] {
						seen[key] = true
						relations = append(relations, ClashRelation{pi.name, pj.name, "三会"})
					}
				}
			}
		}
	}
	return relations
}

func toTymeGender(gender string) (tyme.Gender, error) {
	switch gender {
	case "MALE":
		return tyme.MAN, nil
	case "FEMALE":
		return tyme.WOMAN, nil
	default:
		return 0, fmt.Errorf("invalid gender %q: must be MALE or FEMALE", gender)
	}
}


