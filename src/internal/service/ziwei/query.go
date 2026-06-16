package ziwei

import "sort"

const (
	QueryRuleVersion = "ziwei-query-2026-06-16"
	QuerySchool      = "紫微斗数-三方四正查询-v1"
)

// QueryView is a precomputed, JSON-friendly query API for a ZiWei chart.
type QueryView struct {
	RuleVersion string              `json:"rule_version"`
	School      string              `json:"school"`
	Palaces     []PalaceQuery       `json:"palaces"`
	StarIndex   map[string][]string `json:"star_index"`
	Patterns    []string            `json:"patterns"`
}

// PalaceQuery mirrors common iztro-style lookups without requiring the
// frontend to scan every palace manually.
type PalaceQuery struct {
	Name              string             `json:"name"`
	Branch            string             `json:"branch"`
	Index             int                `json:"index"`
	IsBodyPalace      bool               `json:"is_body_palace"`
	MainStars         []string           `json:"main_stars"`
	AuxStars          []string           `json:"aux_stars"`
	AdjectiveStars    []string           `json:"adjective_stars"`
	AllStars          []string           `json:"all_stars"`
	HasStar           map[string]bool    `json:"has_star"`
	FourHua           []string           `json:"four_hua"`
	SanfangSizheng    SanfangQuery       `json:"sanfang_sizheng"`
	SurroundedPalaces []SurroundedPalace `json:"surrounded_palaces"`
}

// SanfangQuery names the opposite/trine palaces and the stars available there.
type SanfangQuery struct {
	Opposite      string   `json:"opposite"`
	Trine1        string   `json:"trine1"`
	Trine2        string   `json:"trine2"`
	OppositeStars []string `json:"opposite_stars"`
	Trine1Stars   []string `json:"trine1_stars"`
	Trine2Stars   []string `json:"trine2_stars"`
	AllStars      []string `json:"all_stars"`
}

// SurroundedPalace is a palace participating in the current palace's 三方四正.
type SurroundedPalace struct {
	Name   string   `json:"name"`
	Branch string   `json:"branch"`
	Role   string   `json:"role"`
	Stars  []string `json:"stars"`
}

// BuildQueryView builds a precomputed palace/star lookup surface.
func BuildQueryView(chart *ZiWeiChart) *QueryView {
	if chart == nil {
		return nil
	}
	view := &QueryView{
		RuleVersion: QueryRuleVersion,
		School:      QuerySchool,
		Palaces:     make([]PalaceQuery, 0, len(chart.Palaces)),
		StarIndex:   map[string][]string{},
		Patterns:    append([]string(nil), chart.Patterns...),
	}
	for i := range chart.Palaces {
		q := buildPalaceQuery(chart, i)
		view.Palaces = append(view.Palaces, q)
		for _, star := range q.AllStars {
			view.StarIndex[star] = appendUnique(view.StarIndex[star], q.Name)
		}
	}
	for star := range view.StarIndex {
		sort.Strings(view.StarIndex[star])
	}
	return view
}

func buildPalaceQuery(chart *ZiWeiChart, idx int) PalaceQuery {
	p := chart.Palaces[idx]
	allStars := palaceAllStarNames(p)
	hasStar := make(map[string]bool, len(allStars))
	for _, star := range allStars {
		hasStar[star] = true
	}
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, idx)
	opposite := chart.Palaces[oppositeIdx]
	trine1 := chart.Palaces[trine1Idx]
	trine2 := chart.Palaces[trine2Idx]
	oppositeStars := palaceAllStarNames(opposite)
	trine1Stars := palaceAllStarNames(trine1)
	trine2Stars := palaceAllStarNames(trine2)
	return PalaceQuery{
		Name:           p.Name,
		Branch:         p.Branch,
		Index:          idx,
		IsBodyPalace:   p.IsBodyPalace,
		MainStars:      palaceMainStars(p),
		AuxStars:       palaceAuxStars(p),
		AdjectiveStars: append([]string(nil), p.AdjectiveStars...),
		AllStars:       allStars,
		HasStar:        hasStar,
		FourHua:        append([]string(nil), p.FourHua...),
		SanfangSizheng: SanfangQuery{
			Opposite:      opposite.Name,
			Trine1:        trine1.Name,
			Trine2:        trine2.Name,
			OppositeStars: oppositeStars,
			Trine1Stars:   trine1Stars,
			Trine2Stars:   trine2Stars,
			AllStars:      mergeUniqueStrings(oppositeStars, trine1Stars, trine2Stars),
		},
		SurroundedPalaces: []SurroundedPalace{
			{Name: opposite.Name, Branch: opposite.Branch, Role: "opposite", Stars: oppositeStars},
			{Name: trine1.Name, Branch: trine1.Branch, Role: "trine", Stars: trine1Stars},
			{Name: trine2.Name, Branch: trine2.Branch, Role: "trine", Stars: trine2Stars},
		},
	}
}

func appendUnique(values []string, value string) []string {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}
