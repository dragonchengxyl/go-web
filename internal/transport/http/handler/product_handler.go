package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/studio/platform/internal/domain/product"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type ProductHandler struct {
	productService *usecase.ProductService
}

func NewProductHandler(productService *usecase.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProduct 创建商品
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req product.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInvalidParam, "invalid request", err))
		return
	}

	if err := h.productService.CreateProduct(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, req)
}

// GetProduct 获取商品详情
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid product id"))
		return
	}

	p, err := h.productService.GetProduct(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, p)
}

// ListProducts 获取商品列表
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := product.ListFilter{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if productType := c.Query("product_type"); productType != "" {
		pt := product.ProductType(productType)
		filter.ProductType = &pt
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active := isActive == "true"
		filter.IsActive = &active
	}

	products, total, err := h.productService.ListProducts(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, products, total, page, pageSize)
}

// UpdateProduct 更新商品
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid product id"))
		return
	}

	var req product.Product
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInvalidParam, "invalid request", err))
		return
	}

	req.ID = id
	if err := h.productService.UpdateProduct(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, req)
}

// DeleteProduct 删除商品
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.Error(c, apperr.BadRequest("invalid product id"))
		return
	}

	if err := h.productService.DeleteProduct(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "product deleted"})
}

// CreateDiscount 创建折扣规则
func (h *ProductHandler) CreateDiscount(c *gin.Context) {
	var req product.DiscountRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInvalidParam, "invalid request", err))
		return
	}

	if err := h.productService.CreateDiscount(c.Request.Context(), &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, req)
}
