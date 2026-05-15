package service

import (
	"testing"
)

func TestInterpretPalace(t *testing.T) {
	interp := NewZiWeiInterpreter()

	palace := &PalaceInfo{
		Name:      "命宮",
		MainStars: []string{"紫微"},
		AuxStars:  []string{"左輔", "文昌"},
		Brightness: map[string]string{
			"紫微": "廟",
		},
		FourHua: []string{"紫微化祿"},
	}

	chart := &ZiWeiChart{}
	chart.Palaces[0] = *palace

	reading := interp.InterpretPalace(palace, chart)

	if reading.PalaceName == "" {
		t.Error("PalaceName is empty")
	}
	if reading.MainStarDesc == "" {
		t.Error("MainStarDesc is empty")
	}
	if reading.AuxStarsDesc == "" {
		t.Error("AuxStarsDesc is empty")
	}
	if reading.FourHuaDesc == "" {
		t.Error("FourHuaDesc is empty")
	}
	if reading.TrineOppDesc == "" {
		t.Error("TrineOppDesc is empty")
	}
	if reading.PatternDesc == "" {
		t.Error("PatternDesc is empty")
	}
	if reading.OverallDesc == "" {
		t.Error("OverallDesc is empty")
	}
}
