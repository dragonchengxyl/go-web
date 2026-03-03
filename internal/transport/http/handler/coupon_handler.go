package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-web/internal/domain/coupon"
	"go-web/internal/pkg/response"
	"go-web/internal/transport/http/middleware"
	"go-web/internal/usecase"
)

type CouponHandler struct {
	couponService *usecase.CouponService
}

func NewCouponHandler(couponService *usecase.CouponService) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

// CreateCoupon 创建优惠券（管理员）
func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req coupon.Coupon
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if err := h.couponService.CreateCoupon(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create coupon", err)
		return
	}

	response.Success(c, req)
}

// ValidateCoupon 验证优惠券
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	userID := middleware.GetUserID(c)
	code := c.Query("code")

	if code == "" {
		response.Error(c, http.StatusBadRequest, "coupon code is required", nil)
		return
	}

	couponEntity, err := h.couponService.ValidateCoupon(c.Request.Context(), code, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid coupon", err)
		return
	}

	response.Success(c, couponEntity)
}

// CreateRedeemCode 创建兑换码（管理员）
func (h *CouponHandler) CreateRedeemCode(c *gin.Context) {
	var req coupon.RedeemCode
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	if err := h.couponService.CreateRedeemCode(c.Request.Context(), &req); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to create redeem code", err)
		return
	}

	response.Success(c, req)
}

// BatchCreateRedeemCodes 批量创建兑换码（管理员）
func (h *CouponHandler) BatchCreateRedeemCodes(c *gin.Context) {
	var req struct {
		ProductID   *uuid.UUID `json:"product_id"`
		Description string     `json:"description"`
		Count       int        `json:"count" binding:"required"`
		ExpiresAt   *string    `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	codes, err := h.couponService.BatchCreateRedeemCodes(c.Request.Context(), req.ProductID, req.Description, req.Count, req.ExpiresAt)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to batch create redeem codes", err)
		return
	}

	response.Success(c, codes)
}

// RedeemCode 使用兑换码
func (h *CouponHandler) RedeemCode(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request", err)
		return
	}

	rc, err := h.couponService.RedeemCode(c.Request.Context(), req.Code, userID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to redeem code", err)
		return
	}

	response.Success(c, rc)
}
