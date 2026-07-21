package bazi

import (
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

func TestPillarDetailsExhaustAllSixtyCycles(t *testing.T) {
	for index := 0; index < 60; index++ {
		cycle := tyme.SixtyCycle{}.FromIndex(index)
		stem := cycle.GetHeavenStem().GetName()
		branch := cycle.GetEarthBranch().GetName()
		details := buildPillarDetails([]model.Pillar{{Gan: stem, Zhi: branch}})
		if len(details) != 1 {
			t.Fatalf("pillar detail count for %s = %d", cycle.GetName(), len(details))
		}
		got := details[0]
		empties := cycle.GetExtraEarthBranches()
		wantEmpties := [2]string{empties[0].GetName(), empties[1].GetName()}
		if got.Stem != stem || got.Branch != branch || got.Empties != wantEmpties {
			t.Errorf("pillar detail %s = %+v, want empties %v", cycle.GetName(), got, wantEmpties)
		}
		if got.ShengXiao != data.ShengXiao[cycle.GetEarthBranch().GetIndex()] {
			t.Errorf("pillar detail %s zodiac = %s", cycle.GetName(), got.ShengXiao)
		}
		if got.Nayin.Status != "observed" || got.Nayin.GanZhi != cycle.GetName() ||
			canonicalNaYinAlias(got.Nayin.Name) != canonicalNaYinAlias(cycle.GetSound().GetName()) {
			t.Errorf("pillar detail %s na-yin = %+v, want %s", cycle.GetName(), got.Nayin, cycle.GetSound().GetName())
		}
	}
}

func TestPillarDetailsIgnoreMutableLegacyLookupArrays(t *testing.T) {
	originalZodiac := data.ShengXiao[0]
	originalEmpties := data.Empties[0][0]
	defer func() {
		data.ShengXiao[0] = originalZodiac
		data.Empties[0][0] = originalEmpties
	}()
	data.ShengXiao[0] = "篡改"
	data.Empties[0][0] = [2]string{"篡", "改"}

	detail := buildPillarDetails([]model.Pillar{{Gan: "甲", Zhi: "子"}})[0]
	if detail.ShengXiao != "鼠" || detail.Empties != ([2]string{"戌", "亥"}) {
		t.Fatalf("mutable legacy arrays changed pillar detail: %+v", detail)
	}
}

func TestPillarDetailsFailClosedForInvalidCycle(t *testing.T) {
	detail := buildPillarDetails([]model.Pillar{{Gan: "甲", Zhi: "丑"}})[0]
	if detail.ShengXiao != "" || detail.Empties != ([2]string{}) || detail.Nayin.Status != "unavailable" {
		t.Fatalf("invalid cycle produced derived pillar facts: %+v", detail)
	}
}
