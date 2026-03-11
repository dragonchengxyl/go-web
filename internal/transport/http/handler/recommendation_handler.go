package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// RecommendationHandler handles content recommendation endpoints.
type RecommendationHandler struct {
	svc *usecase.RecommendationService
}

// NewRecommendationHandler creates a new RecommendationHandler.
func NewRecommendationHandler(svc *usecase.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{svc: svc}
}

// GetRecommended handles GET /api/v1/posts/recommended
func (h *RecommendationHandler) GetRecommended(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	_, pageSize := getPageParams(c)
	posts, err := h.svc.GetRecommended(c.Request.Context(), userID, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"posts": posts})
}
