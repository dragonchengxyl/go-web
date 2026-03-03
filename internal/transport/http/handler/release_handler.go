package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// ReleaseHandler handles game release HTTP requests
type ReleaseHandler struct {
	releaseService *usecase.ReleaseService
}

// NewReleaseHandler creates a new ReleaseHandler
func NewReleaseHandler(releaseService *usecase.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{
		releaseService: releaseService,
	}
}

// CreateRelease creates a new game release (Admin only)
// @Summary Create game release
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path string true "Branch ID"
// @Param input body usecase.CreateReleaseInput true "Release creation input"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/branches/{branch_id}/releases [post]
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Get creator ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	creatorID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	// Parse input
	var input usecase.CreateReleaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Create release
	release, err := h.releaseService.CreateRelease(c.Request.Context(), branchID, creatorID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}

// GetRelease retrieves a release by ID
// @Summary Get release by ID
// @Tags game
// @Param release_id path string true "Release ID"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/releases/{release_id} [get]
func (h *ReleaseHandler) GetRelease(c *gin.Context) {
	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Get release
	release, err := h.releaseService.GetRelease(c.Request.Context(), releaseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}

// ListReleases retrieves releases for a branch
// @Summary List branch releases
// @Tags game
// @Param branch_id path string true "Branch ID"
// @Param published_only query bool false "Show only published releases"
// @Success 200 {object} response.Response{data=[]game.Release}
// @Router /api/v1/branches/{branch_id}/releases [get]
func (h *ReleaseHandler) ListReleases(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Parse published_only parameter
	publishedOnly := false
	if po := c.Query("published_only"); po != "" {
		publishedOnly, _ = strconv.ParseBool(po)
	}

	// List releases
	releases, err := h.releaseService.ListReleases(c.Request.Context(), branchID, publishedOnly)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, releases)
}

// UpdateRelease updates a release (Admin only)
// @Summary Update release
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param release_id path string true "Release ID"
// @Param input body usecase.UpdateReleaseInput true "Release update input"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/releases/{release_id} [put]
func (h *ReleaseHandler) UpdateRelease(c *gin.Context) {
	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Parse input
	var input usecase.UpdateReleaseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update release
	release, err := h.releaseService.UpdateRelease(c.Request.Context(), releaseID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}

// PublishRelease publishes a release (Admin only)
// @Summary Publish release
// @Tags game
// @Security BearerAuth
// @Param release_id path string true "Release ID"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/releases/{release_id}/publish [post]
func (h *ReleaseHandler) PublishRelease(c *gin.Context) {
	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Publish release
	release, err := h.releaseService.PublishRelease(c.Request.Context(), releaseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}

// UnpublishRelease unpublishes a release (Admin only)
// @Summary Unpublish release
// @Tags game
// @Security BearerAuth
// @Param release_id path string true "Release ID"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/releases/{release_id}/unpublish [post]
func (h *ReleaseHandler) UnpublishRelease(c *gin.Context) {
	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Unpublish release
	release, err := h.releaseService.UnpublishRelease(c.Request.Context(), releaseID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}

// DeleteRelease deletes a release (Admin only)
// @Summary Delete release
// @Tags game
// @Security BearerAuth
// @Param release_id path string true "Release ID"
// @Success 200 {object} response.Response
// @Router /api/v1/releases/{release_id} [delete]
func (h *ReleaseHandler) DeleteRelease(c *gin.Context) {
	// Parse release ID
	releaseID, err := uuid.Parse(c.Param("release_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的版本ID"))
		return
	}

	// Delete release
	if err := h.releaseService.DeleteRelease(c.Request.Context(), releaseID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "版本已删除"})
}

// GetLatestRelease retrieves the latest published release for a branch
// @Summary Get latest release
// @Tags game
// @Param branch_id path string true "Branch ID"
// @Success 200 {object} response.Response{data=game.Release}
// @Router /api/v1/branches/{branch_id}/releases/latest [get]
func (h *ReleaseHandler) GetLatestRelease(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Get latest release
	release, err := h.releaseService.GetLatestRelease(c.Request.Context(), branchID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, release)
}
