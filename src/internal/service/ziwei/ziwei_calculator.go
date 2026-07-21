package ziwei

import (
	"fmt"
)

// StarInfo holds a star's name, brightness, and four-hua transformation tag.
type StarInfo struct {
	Name       string
	Brightness string
	Mutagen    string // 化禄/化权/化科/化忌, or empty
}

// ziwei_calculator.go — Core calculation algorithms for ZiWei Dou Shu.
// All algorithms reference iztro (https://github.com/SylarLong/iztro) and classical texts.
// No dependency on ziwei-zenith engine.

// ──────────── Birth Info ────────────

// BirthData holds the essential birth parameters for chart calculation.
type BirthData struct {
	SolarYear         int
	SolarMonth        int
	SolarDay          int
	Hour              int
	Minute            int
	Gender            string // "男"/"女" or "MALE"/"FEMALE"
	LunarYear         int
	LunarMonth        int
	LunarDay          int
	YearStem          int // 0=甲...9=癸
	YearBranch        int // 0=子...11=亥
	MonthPillarStem   int
	MonthPillarBranch int
	DayStem           int
	DayBranch         int
	HourStem          int
	HourBranch        int // 0=子时...11=亥时
	IsLeapMonth       bool
}

// ──────────── Chart Calculation Entry Point ────────────

// calculateZiWeiChart computes the core chart after the service has normalized
// and validated all public birth input.
func calculateZiWeiChart(birth *BirthData) (*ZiWeiChart, error) {
	if birth == nil {
		return nil, fmt.Errorf("birth data is nil")
	}

	// 1. Compute soul (命宫) and body (身宫) positions
	soulBranch, bodyBranch, soulStem := calcSoulAndBody(birth)

	// 2. Compute five element bureau (五行局)
	juValue := calcFiveBureau(soulStem, soulBranch)

	// 3. Compute ZiWei and TianFu star positions
	ziweiIdx, tianfuIdx := calcZiweiTianfuPosition(juValue, birth.LunarDay, birth.HourBranch, birth.LunarMonth, birth.IsLeapMonth)

	// 4. Place main stars
	majorStars := placeMajorStars(ziweiIdx, tianfuIdx)

	// 5. Place auxiliary stars
	auxStars := placeAuxStars(birth)

	// 6. Compute four transformations (四化)
	fourHua := calcFourHua(birth.YearStem)

	// 7. Compute adjective stars
	adjectiveStars := placeAdjectiveStars(birth)

	// 8. Compute twelve shen systems
	changSheng12 := placeChangSheng12(juValue, birth.YearStem, birth.Gender)
	boShi12 := placeBoShi12(birth.YearStem, birth.Gender)
	suiQian12, jiangQian12 := placeYearly12(birth.YearBranch)

	// 9. Assemble chart
	chart := assembleChart(birth, soulBranch, bodyBranch, juValue,
		majorStars, auxStars, fourHua, adjectiveStars,
		changSheng12, boShi12, suiQian12, jiangQian12)

	return chart, nil
}

// ──────────── 1. Soul & Body Palace ────────────

// calcSoulAndBody computes 命宫 and 身宫 branch indices and the 命宫天干.
// Algorithm: "寅起正月，顺数至生月" → then "逆数生时为命宫，顺数生时为身宫"
// Reference: iztro palace.ts getSoulAndBody()
func calcSoulAndBody(birth *BirthData) (soulBranch, bodyBranch, soulStem int) {
	// 命宫: 从寅宫起正月，顺数至生月，得到月支位置；然后从该位置逆数时辰
	// In iztro's coordinate system, 寅=0, 卯=1, ..., 丑=11
	// But we use standard: 子=0, 丑=1, ..., 亥=11
	// Convert: 寅-based index = lunarMonth - 1 (正月=0, 二月=1, ...)
	// Then: soul_branch_standard = (寅位 + lunar_month_offset - time_offset) → to standard

	// iztro formula: soulIndex = (lunarMonth - 1 + timeBranch) % 12
	// but this gives index from 寅. We need to convert:
	// In iztro: 寅=0, so soulIndex is relative to 寅.
	// Our standard: 子=0, 寅=2, so: soulBranch = (2 + (lunarMonth-1) - timeBranch + 12) % 12
	// Wait, let me reconsider.
	//
	// The formula from iztro: soulIndex = fixIndex(monthIndex - timeBranchIndex)
	// where monthIndex is 0-based from 寅 (正月=0, 二月=1...)
	// and timeBranchIndex is the standard branch index (子=0...亥=11)
	//
	// But iztro uses 寅=0 indexing for palace positions, while we use 子=0.
	// The correct formula for standard 子=0 indexing:
	//   命宫: start from 寅(2), count forward lunarMonth-1 steps (to 生月地支)
	//          then count backward from that point timeBranch steps
	//   = (2 + (lunarMonth - 1) - birth.HourBranch + 12) % 12
	//
	// Actually, the classic formula is simpler:
	//   命宫地支 = (月支序号 - 时支序号 + 寅支序号) % 12
	//   where 月支 = (lunarMonth - 1 + 寅) % 12 = (lunarMonth + 1) % 12
	//   but more correctly: 寅起正月 means 正月=寅(2), 二月=卯(3), ..., 十二月=丑(1)
	//   so: 月支 = (2 + lunarMonth - 1) % 12 = (lunarMonth + 1) % 12
	//
	// Simplified (iztro logic):
	//   monthBranch (in 寅=0 system) = lunarMonth - 1
	//   soul (in 寅=0 system) = fixIndex(monthBranch - timeBranch)
	//   convert to 子=0 system: soul_std = (soul + 2) % 12
	//
	// But let's verify with the user's example:
	//   癸未年三月十四日未时
	//   lunarMonth = 3, HourBranch = 未 = 7
	//   monthBranch = 3 - 1 = 2 (which is 辰 in 寅=0 system)
	//   Wait, that doesn't match. Let me re-derive.
	//
	// Classic algorithm:
	//   1. 寅(2)起正月: 正月=寅(2), 二月=卯(3), 三月=辰(4)
	//   2. 从辰(4)起子时，逆数至未时:
	//      辰←子, 卯←丑, 寅←寅, 丑←卯, 子←辰, 亥←巳, 戌←午, 酉←未
	//   3. 命宫在酉(9) ✓

	monthBranch := (2 + effectiveLunarMonth(birth) - 1) % 12 // 寅(2)起正月

	// 命宫: 从月支位置逆数时辰
	// 时间子时从月支开始，0=子, 1=丑, ..., 7=未
	soulBranch = fixIndex(monthBranch - birth.HourBranch)

	// 身宫: 从月支位置顺数时辰
	bodyBranch = fixIndex(monthBranch + birth.HourBranch)

	// 命宫天干: 用五虎遁从寅宫起始天干推算
	soulStem = GetPalaceStem(birth.YearStem, soulBranch)

	return soulBranch, bodyBranch, soulStem
}

// effectiveLunarMonth implements this profile's fixLeap convention: the first
// 15 days of a leap month use that month, and day 16 onward uses the next one.
func effectiveLunarMonth(birth *BirthData) int {
	month := birth.LunarMonth
	if birth.IsLeapMonth && birth.LunarDay > 15 {
		month++
	}
	return fixIndex(month-1) + 1
}

// ──────────── 2. Five Element Bureau (五行局) ────────────

// calcFiveBureau computes the 五行局 value for the soul palace.
// Uses 纳音 lookup: find the 干支 pair index for the soul palace's
// 天干地支, then look up the bureau value from NaYinBureauTable.
func calcFiveBureau(soulStem, soulBranch int) int {
	pairIdx := ganzhiPairIndex(soulStem, soulBranch)
	if pairIdx < 0 || pairIdx >= 30 {
		return 5
	}
	return NaYinBureauTable[pairIdx]
}

// ──────────── 3. ZiWei & TianFu Position ────────────

// calcZiweiTianfuPosition computes the ZiWei and TianFu star positions.
// Algorithm: find smallest offset such that (day + offset) % juValue == 0,
// then determine position based on quotient and offset parity.
// Reference: iztro location.ts getStartIndex()
func calcZiweiTianfuPosition(juValue, lunarDay, hourBranch, lunarMonth int, isLeapMonth bool) (ziweiIdx, tianfuIdx int) {
	day := lunarDay

	offset := -1
	quotient := 0
	remainder := 1

	for remainder != 0 {
		offset++
		divisor := day + offset
		quotient = divisor / juValue
		remainder = divisor % juValue
	}
	quotient %= 12

	var ziweiPos int
	if offset%2 == 0 {
		ziweiPos = quotient + offset - 1
	} else {
		ziweiPos = quotient - offset - 1
	}
	ziweiYinBasedIdx := fixIndex(ziweiPos)
	tianfuYinBasedIdx := fixIndex(12 - ziweiYinBasedIdx)
	ziweiIdx = fixIndex(ziweiYinBasedIdx + 2)
	tianfuIdx = fixIndex(tianfuYinBasedIdx + 2)

	return ziweiIdx, tianfuIdx
}

// ──────────── 4. Main Star Placement ────────────

func placeMajorStars(ziweiIdx, tianfuIdx int) [12][]StarInfo {
	var stars [12][]StarInfo

	// 紫微星系: counterclockwise from 紫微 position
	// Order: 紫微, 天机, (空), 太阳, 武曲, 天同, (空), (空), 廉贞
	for i, name := range ZiweiStarOrder {
		if name == "" {
			continue
		}
		idx := fixIndex(ziweiIdx - i)
		brightness := getStarBrightness(name, idx)
		stars[idx] = append(stars[idx], StarInfo{
			Name:       name,
			Brightness: brightness,
		})
	}

	// 天府星系: clockwise from 天府 position
	// Order: 天府, 太阴, 贪狼, 巨门, 天相, 天梁, 七杀, (空*3), 破军
	for i, name := range TianfuStarOrder {
		if name == "" {
			continue
		}
		idx := fixIndex(tianfuIdx + i)
		brightness := getStarBrightness(name, idx)
		stars[idx] = append(stars[idx], StarInfo{
			Name:       name,
			Brightness: brightness,
		})
	}

	return stars
}

func getStarBrightness(starName string, branchIdx int) string {
	if brightness, ok := StarBrightnessMap[starName]; ok {
		return brightness[branchIdx]
	}
	if brightness, ok := AuxStarBrightnessMap[starName]; ok {
		return brightness[branchIdx]
	}
	return ""
}

// ──────────── 5. Auxiliary Star Placement ────────────

func placeAuxStars(birth *BirthData) [12][]StarInfo {
	var stars [12][]StarInfo
	month := effectiveLunarMonth(birth)

	// 左辅: 辰上顺正寻左辅
	zuofuIdx := ZuofuIndex(month)
	stars[zuofuIdx] = append(stars[zuofuIdx], StarInfo{Name: "左辅", Brightness: getStarBrightness("左辅", zuofuIdx)})

	// 右弼: 戌上逆正右弼当
	youbiIdx := YoubiIndex(month)
	stars[youbiIdx] = append(stars[youbiIdx], StarInfo{Name: "右弼", Brightness: getStarBrightness("右弼", youbiIdx)})

	// 文昌: from 戌(10) 逆数时辰
	wenchangIdx := WenchangIndex(birth.HourBranch)
	stars[wenchangIdx] = append(stars[wenchangIdx], StarInfo{Name: "文昌", Brightness: getStarBrightness("文昌", wenchangIdx)})

	// 文曲: from 辰(4) 顺数时辰
	wenquIdx := WenquIndex(birth.HourBranch)
	stars[wenquIdx] = append(stars[wenquIdx], StarInfo{Name: "文曲", Brightness: getStarBrightness("文曲", wenquIdx)})

	// 天魁天钺: 按年干
	kuiYue := KuiYueTable[birth.YearStem]
	stars[kuiYue[0]] = append(stars[kuiYue[0]], StarInfo{Name: "天魁", Brightness: getStarBrightness("天魁", kuiYue[0])})
	stars[kuiYue[1]] = append(stars[kuiYue[1]], StarInfo{Name: "天钺", Brightness: getStarBrightness("天钺", kuiYue[1])})

	// 禄存擎羊陀罗天马: 按年干+年支
	lucunIdx := LucunBranchIdx[birth.YearStem]
	stars[lucunIdx] = append(stars[lucunIdx], StarInfo{Name: "禄存", Brightness: getStarBrightness("禄存", lucunIdx)})

	qingyangIdx := fixIndex(lucunIdx + 1)
	stars[qingyangIdx] = append(stars[qingyangIdx], StarInfo{Name: "擎羊", Brightness: getStarBrightness("擎羊", qingyangIdx)})

	tuoluoIdx := fixIndex(lucunIdx - 1)
	stars[tuoluoIdx] = append(stars[tuoluoIdx], StarInfo{Name: "陀罗", Brightness: getStarBrightness("陀罗", tuoluoIdx)})

	tianmaIdx := TianmaBranchIdx[birth.YearBranch]
	stars[tianmaIdx] = append(stars[tianmaIdx], StarInfo{Name: "天马"})

	// 地空地劫: from 亥(11) ± time
	dikongIdx := fixIndex(11 - birth.HourBranch)
	dijieIdx := fixIndex(11 + birth.HourBranch)
	stars[dikongIdx] = append(stars[dikongIdx], StarInfo{Name: "地空", Brightness: getStarBrightness("地空", dikongIdx)})
	stars[dijieIdx] = append(stars[dijieIdx], StarInfo{Name: "地劫", Brightness: getStarBrightness("地劫", dijieIdx)})

	// 火星铃星: 按年支+时辰
	huoIdx, lingIdx := HuolingIndex(birth.YearBranch, birth.HourBranch)
	stars[huoIdx] = append(stars[huoIdx], StarInfo{Name: "火星", Brightness: getStarBrightness("火星", huoIdx)})
	stars[lingIdx] = append(stars[lingIdx], StarInfo{Name: "铃星", Brightness: getStarBrightness("铃星", lingIdx)})

	return stars
}

// ──────────── 6. Four Transformations (四化) ────────────

func calcFourHua(yearStem int) map[string]string {
	hua := SiHuaTable[yearStem]
	return map[string]string{
		hua[0]: "化禄",
		hua[1]: "化权",
		hua[2]: "化科",
		hua[3]: "化忌",
	}
}

// applyFourHua marks stars with their 四化 and returns the four hua star names per palace.
func applyFourHua(majorStars [12][]StarInfo, auxStars [12][]StarInfo, yearStem int) [12][]string {
	huaMap := calcFourHua(yearStem)
	var result [12][]string

	for i := 0; i < 12; i++ {
		// Check major stars
		for j := range majorStars[i] {
			if label, ok := huaMap[majorStars[i][j].Name]; ok {
				majorStars[i][j].Mutagen = label
				result[i] = append(result[i], majorStars[i][j].Name+label)
			}
		}
		// Check aux stars (左辅右弼文昌文曲 can have 四化)
		for j := range auxStars[i] {
			if label, ok := huaMap[auxStars[i][j].Name]; ok {
				auxStars[i][j].Mutagen = label
				result[i] = append(result[i], auxStars[i][j].Name+label)
			}
		}
	}
	return result
}

// ──────────── 7. Adjective Star Placement ────────────

func placeAdjectiveStars(birth *BirthData) [12][]string {
	var stars [12][]string
	month := effectiveLunarMonth(birth)
	monthIndex := month - 1
	dayIndex := birth.LunarDay - 1
	soulBranch, bodyBranch, _ := calcSoulAndBody(birth)

	addAdjectiveStar(&stars, HongLuanIndex(birth.YearBranch), "红鸾")
	addAdjectiveStar(&stars, TianXiIndex(birth.YearBranch), "天喜")
	addAdjectiveStar(&stars, TianYaoIndex(month), "天姚")
	addAdjectiveStar(&stars, XianChiBranch[birth.YearBranch], "咸池")
	addAdjectiveStar(&stars, YueJieBranchByMonth[monthIndex], "解神")

	addAdjectiveStar(&stars, ZuofuIndex(month)+dayIndex, "三台")
	addAdjectiveStar(&stars, YoubiIndex(month)-dayIndex, "八座")
	addAdjectiveStar(&stars, WenchangIndex(birth.HourBranch)+dayIndex-1, "恩光")
	addAdjectiveStar(&stars, WenquIndex(birth.HourBranch)+dayIndex-1, "天贵")

	addAdjectiveStar(&stars, 4+birth.YearBranch, "龙池")
	addAdjectiveStar(&stars, 10-birth.YearBranch, "凤阁")
	addAdjectiveStar(&stars, soulBranch+birth.YearBranch, "天才")
	addAdjectiveStar(&stars, bodyBranch+birth.YearBranch, "天寿")
	addAdjectiveStar(&stars, 6+birth.HourBranch, "台辅")
	addAdjectiveStar(&stars, 2+birth.HourBranch, "封诰")
	addAdjectiveStar(&stars, TianWuBranchByMonth[monthIndex], "天巫")

	addAdjectiveStar(&stars, HuaGaiBranch[birth.YearBranch], "华盖")
	addAdjectiveStar(&stars, TianGuanBranchByStem[birth.YearStem], "天官")
	addAdjectiveStar(&stars, TianFuAdjectiveBranchByStem[birth.YearStem], "天福")
	addAdjectiveStar(&stars, TianChuBranchByStem[birth.YearStem], "天厨")
	addAdjectiveStar(&stars, TianYueBranchByMonth[monthIndex], "天月")
	addAdjectiveStar(&stars, 9+birth.YearBranch, "天德")
	addAdjectiveStar(&stars, 5+birth.YearBranch, "月德")
	addAdjectiveStar(&stars, birth.YearBranch+1, "天空")

	xunkongBranch := fixIndex(birth.YearBranch + 10 - birth.YearStem)
	if birth.YearBranch%2 != xunkongBranch%2 {
		xunkongBranch = fixIndex(xunkongBranch + 1)
	}
	addAdjectiveStar(&stars, xunkongBranch, "旬空")
	addAdjectiveStar(&stars, JieLuBranchByStem[birth.YearStem], "截路")
	addAdjectiveStar(&stars, KongWangBranchByStem[birth.YearStem], "空亡")

	addAdjectiveStar(&stars, GuChenBranch[birth.YearBranch], "孤辰")
	addAdjectiveStar(&stars, GuaSuBranch[birth.YearBranch], "寡宿")
	addAdjectiveStar(&stars, FeiLianBranch[birth.YearBranch], "蜚廉")
	addAdjectiveStar(&stars, PoSuiBranch[birth.YearBranch], "破碎")
	addAdjectiveStar(&stars, TianXingIndex(month), "天刑")
	addAdjectiveStar(&stars, YinShaBranch[monthIndex], "阴煞")
	addAdjectiveStar(&stars, 6-birth.YearBranch, "天哭")
	addAdjectiveStar(&stars, 6+birth.YearBranch, "天虚")
	addAdjectiveStar(&stars, soulBranch+7, "天使")
	addAdjectiveStar(&stars, soulBranch+5, "天伤")
	addAdjectiveStar(&stars, NianJieBranch[birth.YearBranch], "年解")

	return stars
}

func addAdjectiveStar(stars *[12][]string, branch int, name string) {
	branch = fixIndex(branch)
	stars[branch] = append(stars[branch], name)
}

// ──────────── 8. Twelve Shen ────────────

func placeChangSheng12(juValue, yearStem int, gender string) [12]string {
	var result [12]string
	startBranch := ChangshengStartBranch[juValue]

	forward := isForwardByYearStem(yearStem, gender)

	for i, name := range ChangSheng12 {
		var idx int
		if forward {
			idx = fixIndex(startBranch + i)
		} else {
			idx = fixIndex(startBranch - i)
		}
		result[idx] = name
	}
	return result
}

func placeBoShi12(yearStem int, gender string) [12]string {
	var result [12]string
	// 博士12神 from 禄存 position, same direction as 长生
	lucunIdx := LucunBranchIdx[yearStem]

	forward := isForwardByYearStem(yearStem, gender)

	for i, name := range BoShi12 {
		var idx int
		if forward {
			idx = fixIndex(lucunIdx + i)
		} else {
			idx = fixIndex(lucunIdx - i)
		}
		result[idx] = name
	}
	return result
}

func isForwardByYearStem(yearStem int, gender string) bool {
	// 大限、长生、博士方向取生年天干阴阳：阳男阴女顺，阴男阳女逆。
	// 合法干支的天干地支阴阳一致，但用年干表达规则更直接。
	isMale := gender == "男" || gender == "MALE" || gender == "M"
	yearStemIsYang := StemIsYang(yearStem)
	return (yearStemIsYang && isMale) || (!yearStemIsYang && !isMale)
}

func placeYearly12(yearBranch int) (suiQian [12]string, jiangQian [12]string) {
	// 岁前12神: starting from year branch, clockwise
	for i, name := range SuiQian12 {
		idx := fixIndex(yearBranch + i)
		suiQian[idx] = name
	}

	// 将前12神: starting from 将星 position determined by year branch three-combination group
	jiangXingBranch := getJiangXingStartBranch(yearBranch)
	for i, name := range JiangQian12 {
		idx := fixIndex(jiangXingBranch + i)
		jiangQian[idx] = name
	}
	return suiQian, jiangQian
}

func getJiangXingStartBranch(yearBranch int) int {
	// Determine three-combination group
	switch {
	case yearBranch == 2 || yearBranch == 6 || yearBranch == 10: // 寅午戌
		return 6 // 午
	case yearBranch == 0 || yearBranch == 4 || yearBranch == 8: // 申子辰
		return 0 // 子
	case yearBranch == 5 || yearBranch == 9 || yearBranch == 1: // 巳酉丑
		return 9 // 酉
	case yearBranch == 11 || yearBranch == 3 || yearBranch == 7: // 亥卯未
		return 3 // 卯
	default:
		return 0
	}
}

// ──────────── 9. Chart Assembly ────────────

func assembleChart(
	birth *BirthData,
	soulBranch, bodyBranch, juValue int,
	majorStars [12][]StarInfo,
	auxStars [12][]StarInfo,
	fourHua map[string]string,
	adjectiveStars [12][]string,
	changSheng12, boShi12 [12]string,
	suiQian12, jiangQian12 [12]string,
) *ZiWeiChart {
	chart := &ZiWeiChart{
		BodyPalace:                BranchNames[bodyBranch],
		LifeMaster:                LifeMasterTable[soulBranch],
		BodyMaster:                BodyMasterTable[birth.YearBranch],
		FiveBureau:                FiveBureauName[juValue],
		LunarMonth:                birth.LunarMonth,
		EarthlyBranchOfSoulPalace: BranchNames[soulBranch],
		EarthlyBranchOfBodyPalace: BranchNames[bodyBranch],
	}

	// Apply four hua to stars
	_ = applyFourHua(majorStars, auxStars, birth.YearStem)

	// Build 12 palaces
	for i := 0; i < 12; i++ {
		branchIdx := fixIndex(soulBranch - i)
		palaceStem := GetPalaceStem(birth.YearStem, branchIdx)

		palace := PalaceInfo{
			Name:           ZIWEI_PALACE_NAMES[i],
			Branch:         BranchNames[branchIdx],
			HeavenlyStem:   StemNames[palaceStem],
			IsBodyPalace:   branchIdx == bodyBranch,
			Stars:          append(starInfosToOutput(majorStars[branchIdx], "major"), starInfosToOutput(auxStars[branchIdx], "aux")...),
			FourHua:        getFourHuaInPalace(majorStars[branchIdx], auxStars[branchIdx], fourHua),
			AdjectiveStars: adjectiveStars[branchIdx],
			Changsheng12:   changSheng12[branchIdx],
			Boshi12:        boShi12[branchIdx],
			JiangQian12:    jiangQian12[branchIdx],
			SuiQian12:      suiQian12[branchIdx],
		}

		chart.Palaces[i] = palace
	}

	for i := 0; i < 12; i++ {
		chart.SanfangSizheng[i] = *getChartPalaceSanfang(chart, i)
	}

	// Detect patterns
	chart.Patterns = DetectLocalPatterns(chart)

	return chart
}

func starInfosToOutput(infos []StarInfo, starType string) []StarOutput {
	out := make([]StarOutput, 0, len(infos))
	for _, info := range infos {
		t := starType
		if t == "aux" {
			t = starAuxType(info.Name)
		}
		out = append(out, StarOutput{
			Name:       info.Name,
			Type:       t,
			Scope:      "origin",
			Brightness: info.Brightness,
		})
	}
	return out
}

func starAuxType(name string) string {
	switch name {
	case "左辅", "右弼", "文昌", "文曲", "天魁", "天钺":
		return "soft"
	case "擎羊", "陀罗", "火星", "铃星", "地空", "地劫":
		return "tough"
	case "天马":
		return "tianma"
	case "禄存":
		return "lucun"
	default:
		return "soft"
	}
}

func getFourHuaInPalace(major, aux []StarInfo, huaMap map[string]string) []string {
	var result []string
	for _, info := range major {
		if label, ok := huaMap[info.Name]; ok {
			result = append(result, info.Name+label)
		}
	}
	for _, info := range aux {
		if label, ok := huaMap[info.Name]; ok {
			result = append(result, info.Name+label)
		}
	}
	return result
}
