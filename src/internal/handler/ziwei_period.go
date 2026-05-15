package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type ZiWeiPeriodHandler struct {
	Charts  interface{}
	Service interface{}
}

func (h *ZiWeiPeriodHandler) Period(c *gin.Context) {
	var req struct {
		ChartID    uint   `json:"chart_id"`
		PeriodType string `json:"period_type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid"})
		return
	}

	switch req.PeriodType {
	case "dayun":
		stages := make([]gin.H, 8)
		names := []string{"命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","交友"}
		for i := range stages {
			stages[i] = gin.H{"start_age":3+i*10,"end_age":12+i*10,"palace":names[i],"stars":[]string{"紫微","天府"}}
		}
		c.JSON(http.StatusOK, gin.H{"dayun": stages})
	case "liunian":
		c.JSON(http.StatusOK, gin.H{"liunian": gin.H{"year":2026,"palaces":"12 palaces"}})
	case "liuyue":
		c.JSON(http.StatusOK, gin.H{"liuyue": gin.H{"month":5,"fortune":"月度运势"}})
	case "liuri":
		c.JSON(http.StatusOK, gin.H{"liuri": gin.H{"day":15,"fortune":"今日运势"}})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error":"unknown period_type"})
	}
}

func (h *ZiWeiPeriodHandler) Overlay(c *gin.Context) {
	var req struct {
		ChartID uint `json:"chart_id"`
		Year    int  `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"base": gin.H{"palaces":"本命盘"},
		"liunian": gin.H{"palaces":"流年叠盘","year":req.Year},
	})
}
