package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-web/internal/domain/order"
	"go-web/internal/pkg/response"
	"go-web/internal/transport/http/middleware"
	"go-web/internal/usecase"
)

type OrderHandler struct {
	orderService *usecase.OrderService
}

func NewOrderHandler(orderService *usecase.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder 创建订单
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Items          []usecase.OrderItemRequest `json:"items" binding:"required"`
		CouponCode     *string                    `json:"coupon_code"`
		IdempotencyKey string                     `json:"idempotency_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	createReq := usecase.CreateOrderRequest{
		UserID:         userID,
		Items:          req.Items,
		CouponCode:     req.CouponCode,
		IdempotencyKey: req.IdempotencyKey,
	}

	o, err := h.orderService.CreateOrder(c.Request.Context(), createReq)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create order", err)
		return
	}

	response.Success(c, o)
}

// GetOrder 获取订单详情
func (h *OrderHandler) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id", err)
		return
	}

	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found", err)
		return
	}

	// 检查订单是否属于当前用户
	if o.UserID != userID {
		response.Error(c, http.StatusForbidden, "access denied", nil)
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
		response.Error(c, http.StatusInternalServerError, "failed to list orders", err)
		return
	}

	response.SuccessWithPagination(c, orders, total, page, pageSize)
}

// PayOrder 支付订单
func (h *OrderHandler) PayOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id", err)
		return
	}

	var req struct {
		PaymentMethod order.PaymentMethod `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	// 检查订单是否属于当前用户
	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found", err)
		return
	}

	if o.UserID != userID {
		response.Error(c, http.StatusForbidden, "access denied", nil)
		return
	}

	if err := h.orderService.PayOrder(c.Request.Context(), id, req.PaymentMethod); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to pay order", err)
		return
	}

	response.Success(c, gin.H{"message": "order paid successfully"})
}

// CancelOrder 取消订单
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id", err)
		return
	}

	// 检查订单是否属于当前用户
	o, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "order not found", err)
		return
	}

	if o.UserID != userID {
		response.Error(c, http.StatusForbidden, "access denied", nil)
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to cancel order", err)
		return
	}

	response.Success(c, gin.H{"message": "order cancelled successfully"})
}
