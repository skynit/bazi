package service

import "fmt"

// SimplifiedZiWei generates a basic ZiWei chart without external library dependencies.
// This is a placeholder that returns the 12 palaces with the correct structure.
func SimplifiedZiWei(year, month, day, hour, minute int, gender string) *ZiWeiChart {
	palaceNames := []string{"命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","交友","官禄","田宅","福德","父母"}
	palaces := [12]PalaceInfo{}
	for i, name := range palaceNames {
		palaces[i] = PalaceInfo{
			Name:       name,
			MainStars:  []string{"紫微", "天机"},
			AuxStars:   []string{"左辅", "文昌"},
			Brightness: map[string]string{"紫微": "庙", "天机": "得"},
			FourHua:    []string{},
		}
	}
	// Rotate palaces based on birth month (simple demo)
	offset := month % 12
	rotated := [12]PalaceInfo{}
	for i := range palaces {
		rotated[i] = palaces[(i+offset)%12]
	}

	return &ZiWeiChart{
		Palaces:    rotated,
		BodyPalace: palaceNames[(month+2)%12],
		LifeMaster: "禄存",
		BodyMaster: "天相",
		FiveBureau: fmt.Sprintf("%s局", []string{"水二","木三","金四","土五","火六"}[month%5]),
		Patterns:   []string{"机月同梁格"},
	}
}

// SimplifiedDayun generates basic dayun stages
func SimplifiedDayun() Dayun {
	stages := make(Dayun, 8)
	palaceNames := []string{"命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","交友"}
	for i := range stages {
		stages[i] = DayunStage{
			StartAge: 3 + i*10,
			EndAge:   12 + i*10,
			Palace:   palaceNames[i],
			Stars:    []string{"紫微", "天府"},
		}
	}
	return stages
}

// SimplifiedLiunian generates a basic liunian chart
func SimplifiedLiunian(year int) *ZiWeiChart {
	return SimplifiedZiWei(year, 1, 1, 0, 0, "MALE")
}
