package fortune

import (
	"fmt"
	"strings"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

var layerElementOrder = []string{"木", "火", "土", "金", "水"}

// BuildFortuneLayers returns 大运、流年、流月、小运 as explicit layers.
func BuildFortuneLayers(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayerSet {
	return model.FortuneLayerSet{
		RuleVersion: bazipkg.RuleVersion,
		School:      bazipkg.RuleSchool,
		DaYun:       buildDaYunLayer(bazi, queryDate, birthYear),
		LiuNian:     buildLiuNianLayer(bazi, queryDate),
		LiuYue:      buildLiuYueLayer(bazi, queryDate, birthYear),
		XiaoYun:     buildXiaoYunLayer(bazi, queryDate, birthYear),
	}
}

func buildDaYunLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	influence := calcDaYunInfluence(bazi, queryDate, birthYear)
	layer := baseLayer("dayun", "大运", influence.CurrentPillar, bazi, queryDate)
	layer.StartAge = influence.StartAge
	layer.EndAge = influence.EndAge
	layer.Age = queryDate.Year() - birthYear
	layer.TenGod = influence.TenGod
	layer.Favorable = influence.Favorable
	layer.Score += influence.Score
	layer.Description = influence.Description
	if influence.Relation != "" {
		layer.Evidence = append(layer.Evidence, influence.Relation)
	}
	return layer
}

func buildLiuNianLayer(bazi *bazipkg.BaziResult, queryDate time.Time) model.FortuneLayer {
	pillar := getYearGanZhi(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
	layer := baseLayer("liunian", "流年", pillar, bazi, queryDate)
	influence := calcLiuNianInfluence(bazi, queryDate)
	layer.Year = queryDate.Year()
	layer.TenGod = influence.TenGod
	layer.Favorable = influence.Favorable
	layer.Score += influence.Score
	layer.Description = influence.Description
	if influence.TaiSuiRelation != "" {
		layer.Evidence = append(layer.Evidence, influence.TaiSuiRelation)
	}
	return layer
}

func buildLiuYueLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	pillar := ""
	if ec, err := getDayEightChar(queryDate.Year(), int(queryDate.Month()), queryDate.Day()); err == nil {
		pillar = ec.GetMonth().GetName()
	}
	layer := baseLayer("liuyue", "流月", pillar, bazi, queryDate)
	layer.Year = queryDate.Year()
	layer.Month = int(queryDate.Month())
	layer.Age = queryDate.Year() - birthYear
	layer.Description = fmt.Sprintf("%d年%d月流月%s，对日主为%s。", layer.Year, layer.Month, layer.Pillar, layer.TenGod)
	return layer
}

func buildXiaoYunLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	age := queryDate.Year() - birthYear
	offset := age
	if strings.Contains(bazi.DaYunInfo.Direction, "逆") {
		offset = -age
	}
	pillar := offsetPillar(bazi.HourPillar, offset)
	layer := baseLayer("xiaoyun", "小运", pillar, bazi, queryDate)
	layer.Age = age
	layer.Year = queryDate.Year()
	layer.Description = fmt.Sprintf("小运以时柱%s为起点，按%s推至%d岁为%s。", bazi.HourPillar.Gan+bazi.HourPillar.Zhi, bazi.DaYunInfo.Direction, age, layer.Pillar)
	layer.Evidence = append(layer.Evidence, "小运用于补充流年以下的年龄层影响，权重低于大运与流年。")
	return layer
}

func baseLayer(key, name, pillar string, bazi *bazipkg.BaziResult, queryDate time.Time) model.FortuneLayer {
	gan, zhi := splitPillar(pillar)
	layer := model.FortuneLayer{
		Key:              key,
		Name:             name,
		Pillar:           pillar,
		Gan:              gan,
		Zhi:              zhi,
		ElementChange:    pillarElementChange(gan, zhi),
		Relations:        layerRelations(gan, zhi, bazi, queryDate),
		ActivatedShenSha: layerShenShaNames(gan, zhi, bazi),
	}
	if gan != "" {
		layer.TenGod = bazipkg.ClassifyTenGod(gan, bazi.DayPillar.Gan, false)
		like, _, _ := getEffectiveFavor(bazi)
		layer.Favorable = isFavorableTenGodByFavor(layer.TenGod, like, bazi.DayPillar.Gan)
		if layer.Favorable {
			layer.Score += 8
			layer.Evidence = append(layer.Evidence, fmt.Sprintf("%s天干%s为%s，落入喜用。", name, gan, layer.TenGod))
		} else if layer.TenGod != "" {
			layer.Score -= 8
			layer.Evidence = append(layer.Evidence, fmt.Sprintf("%s天干%s为%s，非当前喜用。", name, gan, layer.TenGod))
		}
	}
	for _, rel := range layer.Relations {
		layer.Score += rel.Score
	}
	if len(layer.ActivatedShenSha) > 0 {
		layer.Evidence = append(layer.Evidence, fmt.Sprintf("引动神煞：%s。", strings.Join(layer.ActivatedShenSha, "、")))
	}
	return layer
}

func splitPillar(pillar string) (string, string) {
	runes := []rune(strings.TrimSpace(pillar))
	if len(runes) < 2 {
		return "", ""
	}
	return string(runes[0]), string(runes[1])
}

func pillarElementChange(gan, zhi string) map[string]int {
	out := make(map[string]int, len(layerElementOrder))
	for _, e := range layerElementOrder {
		out[e] = 0
	}
	if elem := data.GanElement[gan]; elem != "" {
		out[elem] += 5
	}
	if elem := data.ZhiElement[zhi]; elem != "" {
		out[elem] += 3
	}
	return out
}

func layerRelations(gan, zhi string, bazi *bazipkg.BaziResult, queryDate time.Time) []model.FortuneLayerRelation {
	var rels []model.FortuneLayerRelation
	if gan != "" {
		rel := stemRelation(bazi.DayPillar.Gan, gan)
		if rel != "unknown" {
			score := relationScore(rel)
			rels = append(rels, model.FortuneLayerRelation{
				Target: "日干",
				Type:   stemRelLabel(rel, bazi.DayPillar.Gan, gan),
				Detail: fmt.Sprintf("%s与日干%s形成%s。", gan, bazi.DayPillar.Gan, stemRelLabel(rel, bazi.DayPillar.Gan, gan)),
				Score:  score,
			})
		}
		if todayGan := queryDateGan(queryDate); todayGan != "" {
			rel := stemRelation(todayGan, gan)
			if rel != "unknown" {
				rels = append(rels, model.FortuneLayerRelation{
					Target: "流日天干",
					Type:   stemRelLabel(rel, todayGan, gan),
					Detail: fmt.Sprintf("%s与流日天干%s形成%s。", gan, todayGan, stemRelLabel(rel, todayGan, gan)),
					Score:  relationScore(rel) / 2,
				})
			}
		}
	}
	if zhi != "" {
		targets := []struct {
			name string
			zhi  string
		}{
			{"年支", bazi.YearPillar.Zhi},
			{"月支", bazi.MonthPillar.Zhi},
			{"日支", bazi.DayPillar.Zhi},
			{"时支", bazi.HourPillar.Zhi},
		}
		for _, target := range targets {
			rel := branchRelation(target.zhi, zhi)
			if rel == "neutral" || rel == "unknown" {
				continue
			}
			rels = append(rels, model.FortuneLayerRelation{
				Target: target.name,
				Type:   branchRelLabel(rel),
				Detail: fmt.Sprintf("%s与%s%s形成%s。", zhi, target.name, target.zhi, branchRelLabel(rel)),
				Score:  relationScore(rel),
			})
		}
	}
	return rels
}

func relationScore(rel string) int {
	switch rel {
	case "same":
		return 4
	case "shengWo":
		return 8
	case "woSheng":
		return 3
	case "woKe":
		return 4
	case "keWo":
		return -8
	case "combine":
		return 6
	case "sanHe":
		return 8
	case "sanHui":
		return 10
	case "clash":
		return -12
	case "punish":
		return -9
	case "harm":
		return -7
	case "break":
		return -5
	default:
		return 0
	}
}

func layerShenShaNames(gan, zhi string, bazi *bazipkg.BaziResult) []string {
	if gan == "" || zhi == "" {
		return nil
	}
	acts := calcShenShaActivation(gan, zhi, bazi)
	names := make([]string, 0, len(acts))
	seen := map[string]bool{}
	for _, act := range acts {
		if act.Name != "" && !seen[act.Name] {
			names = append(names, act.Name)
			seen[act.Name] = true
		}
	}
	return names
}

func offsetPillar(p model.Pillar, offset int) string {
	start := sixtyIndex(p)
	if start < 0 {
		return p.Gan + p.Zhi
	}
	idx := (start + offset) % 60
	if idx < 0 {
		idx += 60
	}
	return data.Gans[idx%10] + data.Zhis[idx%12]
}

func sixtyIndex(p model.Pillar) int {
	for i := 0; i < 60; i++ {
		if data.Gans[i%10] == p.Gan && data.Zhis[i%12] == p.Zhi {
			return i
		}
	}
	return -1
}
