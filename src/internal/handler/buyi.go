package handler

import (
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service/buyi"

	"github.com/gin-gonic/gin"
)

type BuyiStore interface {
	Create(record *model.BuyiRecord) error
	FindByUserDate(userID uint, date time.Time) (*model.BuyiRecord, error)
}

type BuyiHandler struct {
	Service *buyi.Service
	Store   BuyiStore
	Now     func() time.Time
}

func RegisterBuyiRoutes(router gin.IRoutes, svc *buyi.Service, store BuyiStore) {
	h := &BuyiHandler{Service: svc, Store: store}
	router.GET("/buyi/today", h.Today)
	router.POST("/buyi/today", h.DrawToday)
}

func (h *BuyiHandler) Today(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	date := h.todayDate()
	record, err := h.Store.FindByUserDate(userID, date)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to query buyi record")
		return
	}

	respondJSON(c, http.StatusOK, h.todayResponse(date, record, false))
}

func (h *BuyiHandler) DrawToday(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	date := h.todayDate()
	existing, err := h.Store.FindByUserDate(userID, date)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to query buyi record")
		return
	}
	if existing != nil {
		respondJSON(c, http.StatusOK, h.todayResponse(date, existing, true))
		return
	}

	reading := h.buyiService().Draw()
	record := &model.BuyiRecord{
		UserID:         userID,
		DivinationDate: date,
		HexagramNumber: reading.Hexagram.Number,
		HexagramName:   reading.Hexagram.Name,
		Score:          reading.Score,
		Level:          reading.Level,
		Summary:        reading.Summary,
		HumanWay:       reading.Hexagram.HumanWay,
		ImageReading:   reading.Hexagram.ImageReading,
		Advice:         reading.Advice,
	}
	if err := h.Store.Create(record); err != nil {
		record, readErr := h.Store.FindByUserDate(userID, date)
		if readErr == nil && record != nil {
			respondJSON(c, http.StatusOK, h.todayResponse(date, record, true))
			return
		}
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to create buyi record")
		return
	}

	respondJSON(c, http.StatusOK, h.todayResponse(date, record, false))
}

func (h *BuyiHandler) buyiService() *buyi.Service {
	if h.Service != nil {
		return h.Service
	}
	return buyi.NewService()
}

func (h *BuyiHandler) todayDate() time.Time {
	loc := shanghaiLocation()
	now := time.Now()
	if h.Now != nil {
		now = h.Now()
	}
	now = now.In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

func (h *BuyiHandler) todayResponse(date time.Time, record *model.BuyiRecord, alreadyDrawn bool) model.BuyiTodayResponse {
	return model.BuyiTodayResponse{
		Date:         date.In(shanghaiLocation()).Format("2006-01-02"),
		HasRecord:    record != nil,
		AlreadyDrawn: alreadyDrawn,
		Record:       buyiRecordResponse(record),
	}
}

func buyiRecordResponse(record *model.BuyiRecord) *model.BuyiRecordResponse {
	if record == nil {
		return nil
	}
	return &model.BuyiRecordResponse{
		ID:             record.ID,
		HexagramNumber: record.HexagramNumber,
		HexagramName:   record.HexagramName,
		Score:          record.Score,
		Level:          record.Level,
		Summary:        record.Summary,
		HumanWay:       record.HumanWay,
		ImageReading:   record.ImageReading,
		Advice:         record.Advice,
		Source:         buyi.Source,
		CreatedAt:      formatAPITime(record.CreatedAt),
	}
}

func currentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

func shanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*3600)
	}
	return loc
}
