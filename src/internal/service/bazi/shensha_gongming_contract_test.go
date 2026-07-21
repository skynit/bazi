package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestUnsupportedGongMingShortcutsAreAbsentFromProductionSource(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"addLongHuGongMing", "addFengHuangGongMing", "龙虎拱命", "凤凰拱命"} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production shen-sha source still contains unsupported shortcut %q", forbidden)
		}
	}
}

func TestUnorderedBranchPairsDoNotCreateGongMing(t *testing.T) {
	for _, pair := range []struct {
		name      string
		first     string
		second    string
		forbidden string
	}{
		{name: "dragon-tiger", first: "辰", second: "寅", forbidden: "龙虎拱命"},
		{name: "phoenix", first: "酉", second: "巳", forbidden: "凤凰拱命"},
	} {
		t.Run(pair.name, func(t *testing.T) {
			for firstIndex := 0; firstIndex < 4; firstIndex++ {
				for secondIndex := 0; secondIndex < 4; secondIndex++ {
					if firstIndex == secondIndex {
						continue
					}
					name := pillarBucketName(firstIndex) + "-" + pillarBucketName(secondIndex)
					t.Run(name, func(t *testing.T) {
						pillars := []model.Pillar{
							{Gan: "甲", Zhi: "子"}, {Gan: "甲", Zhi: "子"},
							{Gan: "甲", Zhi: "子"}, {Gan: "甲", Zhi: "子"},
						}
						pillars[firstIndex] = gongMingTestPillarForBranch(t, pair.first)
						pillars[secondIndex] = gongMingTestPillarForBranch(t, pair.second)
						got, err := CalcShenShaByPillars(ShenShaPillars{
							Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: "MALE",
						})
						if err != nil {
							t.Fatal(err)
						}
						for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
							if hasShenShaName(bucket, pair.forbidden) {
								t.Fatalf("unordered branches %s/%s produced %s: %+v", pair.first, pair.second, pair.forbidden, got)
							}
						}
					})
				}
			}
		})
	}
}

func TestRemovedGongMingNamesRemainUnregistered(t *testing.T) {
	for _, name := range []string{"龙虎拱命", "凤凰拱命"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
			t.Errorf("removed %s metadata = %+v", name, meta)
		}
	}
}

func gongMingTestPillarForBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi == zhi {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %q", zhi)
	return model.Pillar{}
}

func pillarBucketName(index int) string {
	return []string{"year", "month", "day", "hour"}[index]
}
