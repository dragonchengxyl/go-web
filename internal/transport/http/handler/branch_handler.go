package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// BranchHandler handles game branch HTTP requests
type BranchHandler struct {
	branchService *usecase.BranchService
}

// NewBranchHandler creates a new BranchHandler
func NewBranchHandler(branchService *usecase.BranchService) *BranchHandler {
	return &BranchHandler{
		branchService: branchService,
	}
}

// CreateBranch creates a new game branch (Admin only)
// @Summary Create game branch
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param game_id path string true "Game ID"
// @Param input body usecase.CreateBranchInput true "Branch creation input"
// @Success 200 {object} response.Response{data=game.Branch}
// @Router /api/v1/games/{game_id}/branches [post]
func (h *BranchHandler) CreateBranch(c *gin.Context) {
	// Parse game ID
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// Parse input
	var input usecase.CreateBranchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Create branch
	branch, err := h.branchService.CreateBranch(c.Request.Context(), gameID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, branch)
}

// GetBranch retrieves a branch by ID
// @Summary Get branch by ID
// @Tags game
// @Param branch_id path string true "Branch ID"
// @Success 200 {object} response.Response{data=game.Branch}
// @Router /api/v1/branches/{branch_id} [get]
func (h *BranchHandler) GetBranch(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Get branch
	branch, err := h.branchService.GetBranch(c.Request.Context(), branchID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, branch)
}

// ListBranches retrieves branches for a game
// @Summary List game branches
// @Tags game
// @Param game_id path string true "Game ID"
// @Success 200 {object} response.Response{data=[]game.Branch}
// @Router /api/v1/games/{game_id}/branches [get]
func (h *BranchHandler) ListBranches(c *gin.Context) {
	// Parse game ID
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的游戏ID"))
		return
	}

	// List branches
	branches, err := h.branchService.ListBranches(c.Request.Context(), gameID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, branches)
}

// UpdateBranch updates a branch (Admin only)
// @Summary Update branch
// @Tags game
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path string true "Branch ID"
// @Param input body usecase.UpdateBranchInput true "Branch update input"
// @Success 200 {object} response.Response{data=game.Branch}
// @Router /api/v1/branches/{branch_id} [put]
func (h *BranchHandler) UpdateBranch(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Parse input
	var input usecase.UpdateBranchInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	// Update branch
	branch, err := h.branchService.UpdateBranch(c.Request.Context(), branchID, input)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, branch)
}

// DeleteBranch deletes a branch (Admin only)
// @Summary Delete branch
// @Tags game
// @Security BearerAuth
// @Param branch_id path string true "Branch ID"
// @Success 200 {object} response.Response
// @Router /api/v1/branches/{branch_id} [delete]
func (h *BranchHandler) DeleteBranch(c *gin.Context) {
	// Parse branch ID
	branchID, err := uuid.Parse(c.Param("branch_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的分支ID"))
		return
	}

	// Delete branch
	if err := h.branchService.DeleteBranch(c.Request.Context(), branchID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "分支已删除"})
}
