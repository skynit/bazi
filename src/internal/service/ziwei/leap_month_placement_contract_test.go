package ziwei

import "testing"

func TestLeapMonthEffectivePlacementBoundary(t *testing.T) {
	tests := []struct {
		name           string
		lunarDay       int
		effectiveMonth int
		soulBody       string
		zuofu          string
		youbi          string
		tianyao        string
		tianxing       string
		yinsha         string
	}{
		{
			name: "day 15 stays in leap month", lunarDay: 15, effectiveMonth: 4,
			soulBody: "巳", zuofu: "未", youbi: "未", tianyao: "辰", tianxing: "子", yinsha: "申",
		},
		{
			name: "day 16 advances to next month", lunarDay: 16, effectiveMonth: 5,
			soulBody: "午", zuofu: "申", youbi: "午", tianyao: "巳", tianxing: "丑", yinsha: "午",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			birth := &BirthData{
				LunarMonth:  4,
				LunarDay:    tt.lunarDay,
				IsLeapMonth: true,
				YearStem:    6,
				YearBranch:  0,
				Gender:      "男",
			}
			if got := effectiveLunarMonth(birth); got != tt.effectiveMonth {
				t.Fatalf("effective lunar month = %d, want %d", got, tt.effectiveMonth)
			}

			soul, body, _ := calcSoulAndBody(birth)
			if BranchNames[soul] != tt.soulBody || BranchNames[body] != tt.soulBody {
				t.Fatalf("soul/body = %s/%s, want %s/%s", BranchNames[soul], BranchNames[body], tt.soulBody, tt.soulBody)
			}

			aux := placeAuxStars(birth)
			assertStarBranch(t, aux, "左辅", tt.zuofu)
			assertStarBranch(t, aux, "右弼", tt.youbi)

			adjective := placeAdjectiveStars(birth)
			assertAdjectiveStarBranch(t, adjective, "天姚", tt.tianyao)
			assertAdjectiveStarBranch(t, adjective, "天刑", tt.tianxing)
			assertAdjectiveStarBranch(t, adjective, "阴煞", tt.yinsha)
		})
	}
}

func assertStarBranch(t testing.TB, placements [12][]StarInfo, star, wantBranch string) {
	t.Helper()
	for branch, stars := range placements {
		for _, placed := range stars {
			if placed.Name == star {
				if got := BranchNames[branch]; got != wantBranch {
					t.Fatalf("%s branch = %s, want %s", star, got, wantBranch)
				}
				return
			}
		}
	}
	t.Fatalf("%s was not placed", star)
}

func assertAdjectiveStarBranch(t testing.TB, placements [12][]string, star, wantBranch string) {
	t.Helper()
	for branch, stars := range placements {
		for _, placed := range stars {
			if placed == star {
				if got := BranchNames[branch]; got != wantBranch {
					t.Fatalf("%s branch = %s, want %s", star, got, wantBranch)
				}
				return
			}
		}
	}
	t.Fatalf("%s was not placed", star)
}
