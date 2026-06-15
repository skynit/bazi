package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

const maxFeedbackCommentRunes = 1000

type FeedbackChartStore interface {
	FindByIDForUser(id uint, userID uint) (*model.BirthChart, error)
}

type FeedbackStore interface {
	Create(feedback *model.FortuneFeedback) error
	SummaryByChartID(userID, chartID uint) ([]model.FeedbackSummaryItem, int64, error)
}

type FeedbackHandler struct {
	Charts   FeedbackChartStore
	Feedback FeedbackStore
}

func (h *FeedbackHandler) Create(c *gin.Context) {
	uid, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.FeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.ChartID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id is required"})
		return
	}
	rating := strings.TrimSpace(req.Rating)
	if !validFeedbackRating(rating) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rating"})
		return
	}
	comment := strings.TrimSpace(req.Comment)
	if utf8.RuneCountInString(comment) > maxFeedbackCommentRunes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "comment is too long"})
		return
	}
	if h == nil || h.Charts == nil || h.Feedback == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}
	chart, err := h.Charts.FindByIDForUser(req.ChartID, uid)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	tagsJSON, err := json.Marshal(cleanFeedbackTags(req.Tags))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tags"})
		return
	}
	targetType := strings.TrimSpace(req.TargetType)
	if targetType == "" {
		targetType = "section"
	}
	feedback := &model.FortuneFeedback{
		UserID:          uid,
		ChartID:         req.ChartID,
		TargetType:      targetType,
		TargetID:        strings.TrimSpace(req.TargetID),
		Rating:          rating,
		Tags:            string(tagsJSON),
		Comment:         comment,
		EventYear:       req.EventYear,
		EventCategory:   strings.TrimSpace(req.EventCategory),
		ConsentResearch: req.ConsentResearch,
		ConsentTraining: req.ConsentTraining,
	}
	if err := h.Feedback.Create(feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save feedback"})
		return
	}

	c.JSON(http.StatusOK, model.FeedbackResponse{ID: feedback.ID, Status: "ok"})
}

func (h *FeedbackHandler) Summary(c *gin.Context) {
	uid, ok := authUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	chartID, ok := parseUintQuery(c, "chart_id")
	if !ok || chartID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id is required"})
		return
	}
	if h == nil || h.Charts == nil || h.Feedback == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}
	chart, err := h.Charts.FindByIDForUser(uint(chartID), uid)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}
	items, total, err := h.Feedback.SummaryByChartID(uid, uint(chartID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query feedback"})
		return
	}
	c.JSON(http.StatusOK, model.FeedbackSummaryResponse{
		ChartID: uint(chartID),
		Total:   total,
		Items:   items,
	})
}

func RegisterFeedbackRoutes(r gin.IRouter, charts FeedbackChartStore, feedback FeedbackStore) {
	h := &FeedbackHandler{Charts: charts, Feedback: feedback}
	r.POST("/feedback", h.Create)
	r.GET("/feedback/summary", h.Summary)
}

func authUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	uid, ok := userID.(uint)
	return uid, ok
}

func validFeedbackRating(rating string) bool {
	switch rating {
	case model.FeedbackRatingAccurate,
		model.FeedbackRatingInaccurate,
		model.FeedbackRatingTooGeneric,
		model.FeedbackRatingConfusing,
		model.FeedbackRatingHelpful:
		return true
	default:
		return false
	}
}

func cleanFeedbackTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	seen := map[string]struct{}{}
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || utf8.RuneCountInString(tag) > 32 {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
		if len(out) >= 10 {
			break
		}
	}
	return out
}

func parseUintQuery(c *gin.Context, key string) (uint64, bool) {
	raw := strings.TrimSpace(c.Query(key))
	if raw == "" {
		return 0, false
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return value, true
}
