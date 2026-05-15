package handler

import (
	"net/http"

	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

// AIStubHandler handles AI analysis stub endpoint.
type AIStubHandler struct{}

// AnalyzeFortune handles POST /api/fortune/ai.
// Returns a coming-soon stub. JWT required. Ignores request body.
func (h *AIStubHandler) AnalyzeFortune(c *gin.Context) {
	c.JSON(http.StatusOK, model.AIFortuneStubResponse{
		Status:  "coming_soon",
		Message: "AI分析功能即将上线",
	})
}
