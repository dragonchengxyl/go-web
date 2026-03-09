package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/transport/http/middleware"
	"github.com/studio/platform/internal/usecase"
)

type OrderHandler struct {
	orderService *usecase.OrderService
}

func NewOrderHandler(orderService *usecase.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid order id"))
		return
	}

	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if o.UserID != userID {
		response.Error(c, apperr.ErrForbidden)
		return
	}

	response.Success(c, o)
}

// ListMyOrders 获取我的订单列表
func (h *OrderHandler) ListMyOrders(c *gin.Context) {
	userID := middleware.GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	orders, total, err := h.orderService.ListUserOrders(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, orders, total, page, pageSize)
}

// PayOrder 支付订单
func (h *OrderHandler) PayOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid order id"))
		return
	}

	var req struct {
		PaymentMethod order.PaymentMethod `json:"payment_method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInvalidParam, "invalid request", err))
		return
	}

	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if o.UserID != userID {
		response.Error(c, apperr.ErrForbidden)
		return
	}

	if err := h.orderService.PayOrder(c.Request.Context(), id, req.PaymentMethod); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "order paid successfully"})
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid order id"))
		return
	}

	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	if o.UserID != userID {
		response.Error(c, apperr.ErrForbidden)
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "order cancelled successfully"})
}
