package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/pkg/response"
)

// SponsorHandler serves the sponsor/donation dashboard data.
type SponsorHandler struct {
	cfg *configs.Config
}

func NewSponsorHandler(cfg *configs.Config) *SponsorHandler {
	return &SponsorHandler{cfg: cfg}
}

// GetSponsorInfo GET /api/v1/sponsor
// Returns monthly server cost goal, current raised amount, and QR code URLs.
func (h *SponsorHandler) GetSponsorInfo(c *gin.Context) {
	response.Success(c, gin.H{
		"monthly_goal":   h.cfg.Sponsor.MonthlyGoal,
		"current_raised": h.cfg.Sponsor.CurrentRaised,
		"alipay_qr_url":  h.cfg.Sponsor.AlipayQRURL,
		"wechat_qr_url":  h.cfg.Sponsor.WechatQRURL,
		"message":        h.cfg.Sponsor.Message,
	})
}
