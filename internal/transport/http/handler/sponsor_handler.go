package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// SponsorHandler serves the sponsor/donation dashboard data.
type SponsorHandler struct {
	service *usecase.SponsorSettingsService
}

func NewSponsorHandler(service *usecase.SponsorSettingsService) *SponsorHandler {
	return &SponsorHandler{service: service}
}

// GetSponsorInfo GET /api/v1/sponsor
// Returns monthly server cost goal, current raised amount, and QR code URLs.
func (h *SponsorHandler) GetSponsorInfo(c *gin.Context) {
	if h.service == nil {
		response.Error(c, apperr.ErrInternalError)
		return
	}
	cfg, err := h.service.Get(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"monthly_goal":   cfg.MonthlyGoal,
		"current_raised": cfg.CurrentRaised,
		"alipay_qr_url":  cfg.AlipayQRURL,
		"wechat_qr_url":  cfg.WechatQRURL,
		"message":        cfg.Message,
	})
}
