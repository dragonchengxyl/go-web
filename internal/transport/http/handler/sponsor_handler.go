package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/pkg/response"
)

// SponsorHandler serves the sponsor/donation dashboard data.
type SponsorHandler struct {
	cfg configs.SponsorConfig
}

func NewSponsorHandler(cfg configs.SponsorConfig) *SponsorHandler {
	return &SponsorHandler{cfg: cfg}
}

// GetSponsorInfo GET /api/v1/sponsor
// Returns monthly server cost goal, current raised amount, and QR code URLs.
func (h *SponsorHandler) GetSponsorInfo(c *gin.Context) {
	response.Success(c, gin.H{
		"monthly_goal":   h.cfg.MonthlyGoal,
		"current_raised": h.cfg.CurrentRaised,
		"alipay_qr_url":  h.cfg.AlipayQRURL,
		"wechat_qr_url":  h.cfg.WechatQRURL,
		"message":        h.cfg.Message,
	})
}
