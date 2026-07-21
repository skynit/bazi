package ziwei_test

import "bazi/internal/service/ziwei"

func publishedMainStarNames(palace ziwei.PalaceInfo) []string {
	names := make([]string, 0, len(palace.Stars))
	for _, star := range palace.Stars {
		if star.Type == "major" {
			names = append(names, star.Name)
		}
	}
	return names
}

func publishedAuxStarNames(palace ziwei.PalaceInfo) []string {
	names := make([]string, 0, len(palace.Stars))
	for _, star := range palace.Stars {
		if star.Type != "major" {
			names = append(names, star.Name)
		}
	}
	return names
}

func publishedStarOutputs(mainStars, auxStars []string) []ziwei.StarOutput {
	stars := make([]ziwei.StarOutput, 0, len(mainStars)+len(auxStars))
	for _, name := range mainStars {
		stars = append(stars, ziwei.StarOutput{Name: name, Type: "major", Scope: "origin"})
	}
	for _, name := range auxStars {
		stars = append(stars, ziwei.StarOutput{Name: name, Type: "soft", Scope: "origin"})
	}
	return stars
}
