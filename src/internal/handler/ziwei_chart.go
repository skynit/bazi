package handler

import (
	"net/http"
	"bazi/internal/model"
	"fmt"
	"github.com/gin-gonic/gin"
)

type ZiWeiChartHandler struct {
	Service interface{}
}

type PalaceData struct {
	Name       string            `json:"name"`
	Branch     string            `json:"branch"`
	MainStars  []string          `json:"main_stars"`
	AuxStars   []string          `json:"aux_stars"`
	Brightness map[string]string `json:"brightness"`
	FourHua    []string          `json:"four_hua"`
}

func (h *ZiWeiChartHandler) Calculate(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"unauthorized"})
		return
	}
	var req model.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid body"})
		return
	}

	names := []string{"命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","交友","官禄","田宅","福德","父母"}
	stars := [][]string{{"紫微","天机"},{"天相"},{"天梁","太阳"},{"武曲","七杀"},{"天府","廉贞"},{"巨门"},{"天同"},{"太阴"},{"贪狼"},{"破军"},{"左辅","文昌"},{"右弼","文曲"}}
	brightness := []map[string]string{{"紫微":"庙","天机":"得"},{"天相":"旺"},{"天梁":"利","太阳":"平"},{"武曲":"庙","七杀":"得"},{"天府":"庙","廉贞":"利"},{"巨门":"不"},{"天同":"旺"},{"太阴":"庙"},{"贪狼":"平"},{"破军":"得"},{"左辅":"利","文昌":"利"},{"右弼":"利","文曲":"利"}}
	
	offset := req.BirthMonth % 12
	palaces := make([]PalaceData, 12)
	for i := 0; i < 12; i++ {
		idx := (i + offset) % 12
		branches := []string{"寅","卯","辰","巳","午","未","申","酉","戌","亥","子","丑"}
		palaces[i] = PalaceData{
			Name: names[idx], Branch: branches[idx], MainStars: stars[idx],
			AuxStars: []string{"左辅","文昌"}, Brightness: brightness[idx], FourHua: []string{},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"palaces": palaces,
		"mingZhu": "禄存",
		"shenZhu": "天相",
		"bodyPalace": names[(req.BirthMonth+2)%12],
		"fiveBureau": fmt.Sprintf("%s局", []string{"水二","木三","金四","土五","火六"}[req.BirthMonth%5]),
		"patterns": []string{"机月同梁格"},
	})
}
