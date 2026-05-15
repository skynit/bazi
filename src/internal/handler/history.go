package handler

import (
	"net/http"
	"strconv"

	"bazi/internal/middleware"
	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

// ChartListStore defines the interface for querying birth charts.
type ChartListStore interface {
	FindByID(id uint) (*model.BirthChart, error)
	ListByUser(userID uint, page, pageSize int) ([]model.BirthChart, int64, error)
}

// FortuneHistoryStore defines the interface for querying fortune history.
type FortuneHistoryStore interface {
	ListByChartID(chartID uint, page, pageSize int) ([]model.HistoryResponse, int64, error)
}

// HistoryHandler handles chart listing and fortune history endpoints.
type HistoryHandler struct {
	Charts         ChartListStore
	FortuneHistory FortuneHistoryStore
}

// ListCharts handles GET /api/charts.
func (h *HistoryHandler) ListCharts(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	charts, total, err := h.Charts.ListByUser(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query charts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"charts":    charts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetChart handles GET /api/charts/:id.
func (h *HistoryHandler) GetChart(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chart id"})
		return
	}

	chart, err := h.Charts.FindByID(uint(id))
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	c.JSON(http.StatusOK, chart)
}

// FortuneHistoryList handles GET /api/fortune/history?chart_id=X.
func (h *HistoryHandler) FortuneHistoryList(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	chartID, err := strconv.ParseUint(c.Query("chart_id"), 10, 64)
	if err != nil || chartID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	items, total, err := h.FortuneHistory.ListByChartID(uint(chartID), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query fortune history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RegisterHistoryRoutes registers history routes on the given router.
func RegisterHistoryRoutes(router *gin.Engine, charts ChartListStore, fortuneHistory FortuneHistoryStore) {
	h := &HistoryHandler{Charts: charts, FortuneHistory: fortuneHistory}

	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/charts", h.ListCharts)
		api.GET("/charts/:id", h.GetChart)
		api.GET("/fortune/history", h.FortuneHistoryList)
	}
}
