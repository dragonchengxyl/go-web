package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/asset"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AssetHandler handles user asset HTTP requests
type AssetHandler struct {
	assetService *usecase.AssetService
}

// NewAssetHandler creates a new AssetHandler
func NewAssetHandler(assetService *usecase.AssetService) *AssetHandler {
	return &AssetHandler{
		assetService: assetService,
	}
}

// GrantAsset grants an asset to a user (Admin only)
// @Summary Grant asset to user
// @Tags asset
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param input body usecase.GrantAssetInput true "Grant asset input"
// @Success 200 {object} response.Response{data=asset.UserGameAsset}
// @Router /api/v1/admin/users/{user_id}/assets [post]
func (h *AssetHandler) GrantAsset(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Parse input
	var input usecase.GrantAssetInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Grant asset
	userAsset, err := h.assetService.GrantAsset(c.Request.Context(), userID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, userAsset)
}

// GetMyAssets retrieves current user's assets
// @Summary Get my assets
// @Tags asset
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]asset.UserGameAsset}
// @Router /api/v1/users/me/assets [get]
func (h *AssetHandler) GetMyAssets(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Get assets
	assets, err := h.assetService.GetUserAssets(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, assets)
}

// GetUserAssets retrieves a user's assets (Admin only)
// @Summary Get user assets
// @Tags asset
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=[]asset.UserGameAsset}
// @Router /api/v1/admin/users/{user_id}/assets [get]
func (h *AssetHandler) GetUserAssets(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Get assets
	assets, err := h.assetService.GetUserAssets(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, assets)
}

// GetMyGameAssets retrieves current user's assets for a specific game
// @Summary Get my game assets
// @Tags asset
// @Security BearerAuth
// @Param game_id path string true "Game ID"
// @Success 200 {object} response.Response{data=[]asset.UserGameAsset}
// @Router /api/v1/users/me/games/{game_id}/assets [get]
func (h *AssetHandler) GetMyGameAssets(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Parse game ID
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// Get game assets
	assets, err := h.assetService.GetUserGameAssets(c.Request.Context(), userID, gameID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, assets)
}

// RevokeAsset revokes an asset from a user (Admin only)
// @Summary Revoke asset from user
// @Tags asset
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param asset_type query string true "Asset type"
// @Param asset_id query string true "Asset ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/users/{user_id}/assets [delete]
func (h *AssetHandler) RevokeAsset(c *gin.Context) {
	// Parse user ID
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	// Parse asset type
	assetTypeStr := c.Query("asset_type")
	if assetTypeStr == "" {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "资产类型不能为空"))
		return
	}
	assetType := asset.AssetType(assetTypeStr)

	// Parse asset ID
	assetID, err := uuid.Parse(c.Query("asset_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的资产ID"))
		return
	}

	// Revoke asset
	if err := h.assetService.RevokeAsset(c.Request.Context(), userID, assetType, assetID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "资产已撤销"})
}

// RequestDownload requests a download URL for a release
// @Summary Request download URL
// @Tags asset
// @Security BearerAuth
// @Param release_id path string true "Release ID"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/v1/releases/{release_id}/download [post]
func (h *AssetHandler) RequestDownload(c *gin.Context) {
	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Check download permission
	if err := h.assetService.CheckDownloadPermission(c.Request.Context(), userID, releaseID); err != nil {
		response.Error(c, err)
		return
	}

	// Log download
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	if err := h.assetService.LogDownload(c.Request.Context(), userID, releaseID, clientIP, userAgent); err != nil {
		// Log error but don't fail the request
		c.Error(err)
	}

	// TODO: Generate pre-signed URL from OSS
	// For now, return a placeholder
	response.Success(c, gin.H{
		"download_url": "https://example.com/download/placeholder",
		"expires_in":   900, // 15 minutes
		"message":      "OSS integration pending",
	})
}
