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
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req model.BaziInterpretationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}
	if req.ChartID == 0 {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "chart_id is required")
		return
	}
	if h.Service == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceDisabled, "service not available")
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
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "chart_id is required")
		case errors.Is(err, interpretation.ErrChartNotFound):
			respondError(c, http.StatusNotFound, ErrCodeNotFound, "chart not found")
		case errors.Is(err, interpretation.ErrChartStore):
			respondError(c, http.StatusInternalServerError, ErrCodeServiceDisabled, "service not available")
		default:
			respondError(c, http.StatusInternalServerError, ErrCodeServiceError, err.Error())
		}
		return
	}

	respondJSON(c, http.StatusOK, resp)
}

func RegisterInterpretationRoutes(r gin.IRouter, svc *interpretation.Service) {
	h := &InterpretationHandler{Service: svc}
	r.POST("/interpretation/bazi", h.Bazi)
}
