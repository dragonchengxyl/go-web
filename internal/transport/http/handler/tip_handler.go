package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// TipHandler handles tip (donation) related HTTP requests
type TipHandler struct {
	tipService     *usecase.TipService
	paymentService *usecase.PaymentService
}

func NewTipHandler(tipService *usecase.TipService, paymentService *usecase.PaymentService) *TipHandler {
	return &TipHandler{tipService: tipService, paymentService: paymentService}
}

// CreateTip POST /api/v1/tips
func (h *TipHandler) CreateTip(c *gin.Context) {
	fromUserID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		ToUserID  string  `json:"to_user_id" binding:"required"`
		Amount    float64 `json:"amount" binding:"required,gt=0"`
		Message   string  `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	toUserID, err := uuid.Parse(req.ToUserID)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	order, err := h.tipService.CreateTip(c.Request.Context(), usecase.CreateTipInput{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		AmountCNY:  req.Amount,
		Message:    req.Message,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, order)
}

// ListReceivedTips GET /api/v1/users/:id/tips/received
func (h *TipHandler) ListReceivedTips(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	page, pageSize := getPageParams(c)
	tips, total, err := h.tipService.ListReceivedTips(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"tips": tips, "total": total, "page": page, "size": len(tips)})
}
