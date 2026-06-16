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
	FindByIDForUser(id uint, userID uint) (*model.BirthChart, error)
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
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
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
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to query charts")
		return
	}

	items := make([]model.ChartSummaryResponse, 0, len(charts))
	for _, chart := range charts {
		items = append(items, chartSummaryResponse(chart))
	}

	respondJSON(c, http.StatusOK, model.ChartListResponse{
		Charts:   items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// GetChart handles GET /api/charts/:id.
func (h *HistoryHandler) GetChart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid chart id")
		return
	}

	chart, err := h.Charts.FindByIDForUser(uint(id), userID.(uint))
	if err != nil || chart == nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "chart not found")
		return
	}

	respondJSON(c, http.StatusOK, chartDetailResponse(*chart))
}

// FortuneHistoryList handles GET /api/fortune/history?chart_id=X.
func (h *HistoryHandler) FortuneHistoryList(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	chartID, err := strconv.ParseUint(c.Query("chart_id"), 10, 64)
	if err != nil || chartID == 0 {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "chart_id is required")
		return
	}

	chart, err := h.Charts.FindByIDForUser(uint(chartID), userID.(uint))
	if err != nil || chart == nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "chart not found")
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
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to query fortune history")
		return
	}

	respondJSON(c, http.StatusOK, gin.H{
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
