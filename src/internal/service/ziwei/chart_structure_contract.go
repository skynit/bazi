package ziwei

// validPublishedZiWeiChartStructure verifies deterministic facts that can be
// reconstructed cheaply from the authenticated calculation input. A content
// hash alone only proves that bytes are self-consistent; it does not prove that
// those bytes describe a valid chart under the selected profile.
func validPublishedZiWeiChartStructure(chart *ZiWeiChart, profile CalculationProfile) bool {
	if chart == nil || len(profile.PluginManifest) != 0 {
		// Plugin mutations require a profile-specific validator. The current
		// production profile is intentionally plugin-free.
		return false
	}
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok || chart.DerivationType != "" || chart.DerivationInput != nil ||
		chart.DerivationFingerprint != "" || chart.BaseContentHash != "" || chart.DerivedContentHash != "" ||
		!emptyNatalTransitLayers(chart) {
		return false
	}

	soulBranch, bodyBranch, soulStem := calcSoulAndBody(birth)
	juValue := calcFiveBureau(soulStem, soulBranch)
	if chart.LunarMonth != birth.LunarMonth ||
		chart.EarthlyBranchOfSoulPalace != BranchNames[soulBranch] ||
		chart.EarthlyBranchOfBodyPalace != BranchNames[bodyBranch] ||
		chart.BodyPalace != BranchNames[bodyBranch] ||
		chart.LifeMaster != LifeMasterTable[soulBranch] ||
		chart.BodyMaster != BodyMasterTable[birth.YearBranch] ||
		chart.FiveBureau != FiveBureauName[juValue] {
		return false
	}

	ziweiIdx, tianfuIdx := calcZiweiTianfuPosition(
		juValue,
		birth.LunarDay,
		birth.HourBranch,
		birth.LunarMonth,
		birth.IsLeapMonth,
	)
	majorStars := placeMajorStars(ziweiIdx, tianfuIdx)
	auxStars := placeAuxStars(birth)
	fourHua := calcFourHua(birth.YearStem)
	adjectiveStars := placeAdjectiveStars(birth)
	changSheng12 := placeChangSheng12(juValue, birth.YearStem, birth.Gender)
	boShi12 := placeBoShi12(birth.YearStem, birth.Gender)
	suiQian12, jiangQian12 := placeYearly12(birth.YearBranch)

	bodyPalaces := 0
	seenStars := make(map[string]bool)
	for palaceIdx, palace := range chart.Palaces {
		branchIdx := fixIndex(soulBranch - palaceIdx)
		if palace.Name != ZIWEI_PALACE_NAMES[palaceIdx] ||
			palace.Branch != BranchNames[branchIdx] ||
			palace.HeavenlyStem != StemNames[GetPalaceStem(birth.YearStem, branchIdx)] ||
			palace.IsBodyPalace != (branchIdx == bodyBranch) {
			return false
		}
		if palace.IsBodyPalace {
			bodyPalaces++
		}

		expectedStars := append(
			starInfosToOutput(majorStars[branchIdx], "major"),
			starInfosToOutput(auxStars[branchIdx], "aux")...,
		)
		if !equalPublishedStars(palace.Stars, expectedStars, seenStars) ||
			!equalPublishedStrings(palace.FourHua, getFourHuaInPalace(majorStars[branchIdx], auxStars[branchIdx], fourHua)) ||
			!equalPublishedStrings(palace.AdjectiveStars, adjectiveStars[branchIdx]) ||
			palace.Changsheng12 != changSheng12[branchIdx] ||
			palace.Boshi12 != boShi12[branchIdx] ||
			palace.JiangQian12 != jiangQian12[branchIdx] ||
			palace.SuiQian12 != suiQian12[branchIdx] {
			return false
		}

		wantSanfang := getChartPalaceSanfang(chart, palaceIdx)
		if wantSanfang == nil || chart.SanfangSizheng[palaceIdx] != *wantSanfang {
			return false
		}
	}
	if bodyPalaces != 1 || !equalPublishedStrings(chart.Patterns, DetectLocalPatterns(chart)) {
		return false
	}
	return true
}

func emptyNatalTransitLayers(chart *ZiWeiChart) bool {
	if chart == nil {
		return false
	}
	for i := range chart.Palaces {
		if len(chart.LiuNianStars[i]) != 0 || len(chart.LiuYueStars[i]) != 0 || len(chart.LiuRiStars[i]) != 0 ||
			len(chart.LiuNianFourHua[i]) != 0 || len(chart.LiuYueFourHua[i]) != 0 || len(chart.LiuRiFourHua[i]) != 0 ||
			chart.LiuNianPalaces[i] != "" || chart.LiuYuePalaces[i] != "" || chart.LiuRiPalaces[i] != "" {
			return false
		}
	}
	return true
}

func equalPublishedStars(got, want []StarOutput, seen map[string]bool) bool {
	if got == nil || len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] || got[i].Name == "" || seen[got[i].Name] {
			return false
		}
		seen[got[i].Name] = true
	}
	return true
}

func equalPublishedStrings(got, want []string) bool {
	if (got == nil) != (want == nil) || len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
