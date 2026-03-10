package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// OSSHandler provides OSS upload credential endpoints.
type OSSHandler struct {
	ossService *usecase.OSSService
}

func NewOSSHandler(ossService *usecase.OSSService) *OSSHandler {
	return &OSSHandler{ossService: ossService}
}

// GetUploadToken POST /api/v1/upload/oss-policy
// Returns a signed OSS Post Policy so the frontend can upload directly to OSS.
func (h *OSSHandler) GetUploadToken(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		Purpose string `json:"purpose"`
	}
	_ = c.ShouldBindJSON(&req) // purpose is optional

	policy, err := h.ossService.GenerateUploadToken(c.Request.Context(), usecase.UploadTokenInput{
		UserID:  userID,
		Purpose: req.Purpose,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, policy)
}
