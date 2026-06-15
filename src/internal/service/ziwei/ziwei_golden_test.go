package ziwei

import "testing"

type goldenStarPlacement struct {
	branch string
	name   string
	main   []string
	aux    []string
}

type goldenDayun struct {
	start  int
	end    int
	palace string
}

type goldenCase struct {
	name       string
	year       int
	month      int
	day        int
	hour       int
	minute     int
	gender     string
	soul       string
	body       string
	fiveBureau string
	lifeMaster string
	bodyMaster string
	dayun      []goldenDayun
	placements map[string]goldenStarPlacement
}

func TestZiWeiGoldenMatchesIztro(t *testing.T) {
	svc := NewZiWeiService()

	cases := []goldenCase{
		{
			name: "2003-04-15 male Wei",
			year: 2003, month: 4, day: 15, hour: 14, minute: 0, gender: "男",
			soul:       "酉",
			body:       "亥",
			fiveBureau: "木三局",
			lifeMaster: "文曲",
			bodyMaster: "天相",
			dayun: []goldenDayun{
				{3, 12, "命宫"}, {13, 22, "兄弟"}, {23, 32, "夫妻"}, {33, 42, "子女"},
				{43, 52, "财帛"}, {53, 62, "疾厄"}, {63, 72, "迁移"}, {73, 82, "交友"},
				{83, 92, "事业"}, {93, 102, "田宅"}, {103, 112, "福德"}, {113, 122, "父母"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", main: []string{"太阳", "巨门"}},
				"卯": {branch: "卯", main: []string{"天相"}, aux: []string{"文昌", "天魁"}},
				"辰": {branch: "辰", main: []string{"天机", "天梁"}, aux: []string{"地空", "火星"}},
				"巳": {branch: "巳", main: []string{"紫微", "七杀"}, aux: []string{"天钺", "天马", "铃星"}},
				"午": {branch: "午", aux: []string{"左辅", "地劫"}},
				"申": {branch: "申", aux: []string{"右弼"}},
				"亥": {branch: "亥", main: []string{"天府"}, aux: []string{"文曲", "陀罗"}},
				"子": {branch: "子", main: []string{"天同", "太阴"}, aux: []string{"禄存"}},
				"丑": {branch: "丑", main: []string{"武曲", "贪狼"}, aux: []string{"擎羊"}},
			},
		},
		{
			name: "1984-01-01 male Zi",
			year: 1984, month: 1, day: 1, hour: 0, minute: 0, gender: "男",
			soul:       "子",
			body:       "子",
			fiveBureau: "金四局",
			lifeMaster: "贪狼",
			bodyMaster: "天机",
			dayun: []goldenDayun{
				{4, 13, "命宫"}, {14, 23, "兄弟"}, {24, 33, "夫妻"}, {34, 43, "子女"},
				{44, 53, "财帛"}, {54, 63, "疾厄"}, {64, 73, "迁移"}, {74, 83, "交友"},
				{84, 93, "事业"}, {94, 103, "田宅"}, {104, 113, "福德"}, {114, 123, "父母"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", name: "福德", main: []string{"武曲", "天相"}, aux: []string{"左辅"}},
				"卯": {branch: "卯", name: "田宅", main: []string{"太阳", "天梁"}, aux: []string{"天魁"}},
				"辰": {branch: "辰", name: "事业", main: []string{"七杀"}, aux: []string{"文曲"}},
				"巳": {branch: "巳", name: "交友", main: []string{"天机"}, aux: []string{"天钺", "天马"}},
				"午": {branch: "午", name: "迁移", main: []string{"紫微"}},
				"戌": {branch: "戌", name: "夫妻", main: []string{"廉贞", "天府"}, aux: []string{"文昌", "铃星"}},
				"亥": {branch: "亥", name: "兄弟", main: []string{"太阴"}, aux: []string{"地空", "地劫", "陀罗"}},
				"子": {branch: "子", name: "命宫", main: []string{"贪狼"}, aux: []string{"右弼", "禄存"}},
				"丑": {branch: "丑", name: "父母", main: []string{"天同", "巨门"}, aux: []string{"擎羊"}},
			},
		},
		{
			name: "2020-05-24 leap male",
			year: 2020, month: 5, day: 24, hour: 0, minute: 0, gender: "男",
			soul:       "巳",
			body:       "巳",
			fiveBureau: "金四局",
			lifeMaster: "武曲",
			bodyMaster: "火星",
			dayun: []goldenDayun{
				{4, 13, "命宫"}, {14, 23, "父母"}, {24, 33, "福德"}, {34, 43, "田宅"},
				{44, 53, "事业"}, {54, 63, "交友"}, {64, 73, "迁移"}, {74, 83, "疾厄"},
				{84, 93, "财帛"}, {94, 103, "子女"}, {104, 113, "夫妻"}, {114, 123, "兄弟"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", name: "子女", main: []string{"贪狼"}, aux: []string{"天马", "火星"}},
				"卯": {branch: "卯", name: "夫妻", main: []string{"天机", "巨门"}},
				"辰": {branch: "辰", name: "兄弟", main: []string{"紫微", "天相"}, aux: []string{"文曲"}},
				"巳": {branch: "巳", name: "命宫", main: []string{"天梁"}},
				"未": {branch: "未", name: "福德", aux: []string{"左辅", "右弼", "天钺", "陀罗"}},
				"申": {branch: "申", name: "田宅", main: []string{"廉贞"}, aux: []string{"禄存"}},
				"戌": {branch: "戌", name: "交友", main: []string{"破军"}, aux: []string{"文昌", "铃星"}},
				"亥": {branch: "亥", name: "迁移", main: []string{"天同"}, aux: []string{"地空", "地劫"}},
				"子": {branch: "子", name: "疾厄", main: []string{"武曲", "天府"}},
				"丑": {branch: "丑", name: "财帛", main: []string{"太阳", "太阴"}, aux: []string{"天魁"}},
			},
		},
		{
			name: "1984-02-04 female Zi",
			year: 1984, month: 2, day: 4, hour: 0, minute: 0, gender: "女",
			soul:       "寅",
			body:       "寅",
			fiveBureau: "火六局",
			lifeMaster: "禄存",
			bodyMaster: "火星",
			dayun: []goldenDayun{
				{6, 15, "命宫"}, {16, 25, "兄弟"}, {26, 35, "夫妻"}, {36, 45, "子女"},
				{46, 55, "财帛"}, {56, 65, "疾厄"}, {66, 75, "迁移"}, {76, 85, "交友"},
				{86, 95, "事业"}, {96, 105, "田宅"}, {106, 115, "福德"}, {116, 125, "父母"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", name: "命宫", aux: []string{"禄存", "天马", "火星"}},
				"卯": {branch: "卯", name: "父母", main: []string{"廉贞", "破军"}, aux: []string{"擎羊"}},
				"辰": {branch: "辰", name: "福德", aux: []string{"左辅", "文曲"}},
				"巳": {branch: "巳", name: "田宅", main: []string{"天府"}},
				"午": {branch: "午", name: "事业", main: []string{"天同", "太阴"}},
				"未": {branch: "未", name: "交友", main: []string{"武曲", "贪狼"}, aux: []string{"天钺"}},
				"申": {branch: "申", name: "迁移", main: []string{"太阳", "巨门"}},
				"酉": {branch: "酉", name: "疾厄", main: []string{"天相"}},
				"戌": {branch: "戌", name: "财帛", main: []string{"天机", "天梁"}, aux: []string{"右弼", "文昌", "铃星"}},
				"亥": {branch: "亥", name: "子女", main: []string{"紫微", "七杀"}, aux: []string{"地空", "地劫"}},
				"丑": {branch: "丑", name: "兄弟", aux: []string{"天魁", "陀罗"}},
			},
		},
		{
			name: "1985-06-15 male Wu",
			year: 1985, month: 6, day: 15, hour: 12, minute: 0, gender: "男",
			soul:       "亥",
			body:       "亥",
			fiveBureau: "土五局",
			lifeMaster: "巨门",
			bodyMaster: "天相",
			dayun: []goldenDayun{
				{5, 14, "命宫"}, {15, 24, "兄弟"}, {25, 34, "夫妻"},
			},
			placements: map[string]goldenStarPlacement{
				"亥": {branch: "亥", name: "命宫", main: []string{"天同"}, aux: []string{"天马"}},
			},
		},
		{
			name: "1985-06-15 female Wu",
			year: 1985, month: 6, day: 15, hour: 12, minute: 0, gender: "女",
			soul:       "亥",
			body:       "亥",
			fiveBureau: "土五局",
			lifeMaster: "巨门",
			bodyMaster: "天相",
			dayun: []goldenDayun{
				{5, 14, "命宫"}, {15, 24, "父母"}, {25, 34, "福德"},
			},
			placements: map[string]goldenStarPlacement{
				"亥": {branch: "亥", name: "命宫", main: []string{"天同"}, aux: []string{"天马"}},
			},
		},
		{
			name: "1990-06-15 male Wu",
			year: 1990, month: 6, day: 15, hour: 12, minute: 0, gender: "男",
			soul:       "子",
			body:       "子",
			fiveBureau: "火六局",
			lifeMaster: "贪狼",
			bodyMaster: "火星",
			dayun: []goldenDayun{
				{6, 15, "命宫"}, {16, 25, "父母"}, {26, 35, "福德"},
			},
			placements: map[string]goldenStarPlacement{
				"子": {branch: "子", name: "命宫", main: []string{"武曲", "天府"}},
			},
		},
		{
			name: "1991-02-15 female Hai",
			year: 1991, month: 2, day: 15, hour: 22, minute: 0, gender: "女",
			soul:       "卯",
			body:       "丑",
			fiveBureau: "木三局",
			lifeMaster: "文曲",
			bodyMaster: "天相",
			dayun: []goldenDayun{
				{3, 12, "命宫"}, {13, 22, "父母"}, {23, 32, "福德"},
			},
			placements: map[string]goldenStarPlacement{
				"卯": {branch: "卯", name: "命宫", main: []string{"天机", "巨门"}, aux: []string{"文曲"}},
				"丑": {branch: "丑", name: "夫妻", main: []string{"太阳", "太阴"}},
			},
		},
		{
			name: "1992-01-25 male Hai",
			year: 1992, month: 1, day: 25, hour: 22, minute: 0, gender: "男",
			soul:       "寅",
			body:       "子",
			fiveBureau: "木三局",
			lifeMaster: "禄存",
			bodyMaster: "天相",
			dayun: []goldenDayun{
				{3, 12, "命宫"}, {13, 22, "兄弟"}, {23, 32, "夫妻"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", name: "命宫", main: []string{"七杀"}, aux: []string{"天钺"}},
				"子": {branch: "子", name: "夫妻", main: []string{"廉贞", "天相"}, aux: []string{"地空"}},
			},
		},
		{
			name: "1993-03-01 female Zi",
			year: 1993, month: 3, day: 1, hour: 0, minute: 0, gender: "女",
			soul:       "卯",
			body:       "卯",
			fiveBureau: "水二局",
			lifeMaster: "文曲",
			bodyMaster: "天同",
			dayun: []goldenDayun{
				{2, 11, "命宫"}, {12, 21, "父母"}, {22, 31, "福德"},
			},
			placements: map[string]goldenStarPlacement{
				"卯": {branch: "卯", name: "命宫", main: []string{"天相"}, aux: []string{"天魁", "火星"}},
			},
		},
		{
			name: "1976-09-09 male Hai",
			year: 1976, month: 9, day: 9, hour: 22, minute: 0, gender: "男",
			soul:       "戌",
			body:       "申",
			fiveBureau: "木三局",
			lifeMaster: "禄存",
			bodyMaster: "文昌",
			dayun: []goldenDayun{
				{3, 12, "命宫"}, {13, 22, "父母"}, {23, 32, "福德"},
			},
			placements: map[string]goldenStarPlacement{
				"戌": {branch: "戌", name: "命宫", main: []string{"巨门"}, aux: []string{"地劫"}},
				"申": {branch: "申", name: "夫妻", main: []string{"天机", "太阴"}},
			},
		},
		{
			name: "2024-02-10 female Zi",
			year: 2024, month: 2, day: 10, hour: 0, minute: 0, gender: "女",
			soul:       "寅",
			body:       "寅",
			fiveBureau: "火六局",
			lifeMaster: "禄存",
			bodyMaster: "文昌",
			dayun: []goldenDayun{
				{6, 15, "命宫"}, {16, 25, "兄弟"}, {26, 35, "夫妻"},
			},
			placements: map[string]goldenStarPlacement{
				"寅": {branch: "寅", name: "命宫", aux: []string{"禄存", "天马", "火星"}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			if chart.EarthlyBranchOfSoulPalace != tc.soul {
				t.Fatalf("soul = %s, want %s", chart.EarthlyBranchOfSoulPalace, tc.soul)
			}
			if chart.EarthlyBranchOfBodyPalace != tc.body {
				t.Fatalf("body = %s, want %s", chart.EarthlyBranchOfBodyPalace, tc.body)
			}
			if chart.FiveBureau != tc.fiveBureau {
				t.Fatalf("five bureau = %s, want %s", chart.FiveBureau, tc.fiveBureau)
			}
			if chart.LifeMaster != tc.lifeMaster {
				t.Fatalf("life master = %s, want %s", chart.LifeMaster, tc.lifeMaster)
			}
			if chart.BodyMaster != tc.bodyMaster {
				t.Fatalf("body master = %s, want %s", chart.BodyMaster, tc.bodyMaster)
			}
			dayun := svc.CalculateDayun(chart)
			if len(dayun) < len(tc.dayun) {
				t.Fatalf("dayun length = %d, want at least %d", len(dayun), len(tc.dayun))
			}
			for i, want := range tc.dayun {
				got := dayun[i]
				if got.StartAge != want.start || got.EndAge != want.end || got.Palace != want.palace {
					t.Errorf("dayun[%d] = %d-%d %s, want %d-%d %s",
						i, got.StartAge, got.EndAge, got.Palace, want.start, want.end, want.palace)
				}
			}

			for branch, want := range tc.placements {
				p := findPalaceByBranchForTest(chart, branch)
				if p == nil {
					t.Fatalf("missing palace for branch %s", branch)
				}
				if want.name != "" && p.Name != want.name {
					t.Errorf("%s palace name = %s, want %s", branch, p.Name, want.name)
				}
				if !sameStringSet(p.MainStars, want.main) {
					t.Errorf("%s main stars = %v, want %v", branch, p.MainStars, want.main)
				}
				if !sameStringSet(p.AuxStars, want.aux) {
					t.Errorf("%s aux stars = %v, want %v", branch, p.AuxStars, want.aux)
				}
			}
		})
	}
}
