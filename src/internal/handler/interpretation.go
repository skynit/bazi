package handler

import (
	"errors"
	"net/http"

	"bazi/internal/model"
	"bazi/internal/service/interpretation"

	"github.com/gin-gonic/gin"
)

type InterpretationHandler struct {
	Service *interpretation.Service
}

func (h *InterpretationHandler) Bazi(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.BaziInterpretationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.ChartID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id is required"})
		return
	}
	if h.Service == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	resp, err := h.Service.InterpretBazi(c.Request.Context(), interpretation.Request{
		ChartID: req.ChartID,
		UserID:  uid,
		Focus:   req.Focus,
	})
	if err != nil {
		switch {
		case errors.Is(err, interpretation.ErrChartIDRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id is required"})
		case errors.Is(err, interpretation.ErrChartNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		case errors.Is(err, interpretation.ErrChartStore):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func RegisterInterpretationRoutes(r gin.IRouter, svc *interpretation.Service) {
	h := &InterpretationHandler{Service: svc}
	r.POST("/interpretation/bazi", h.Bazi)
}
