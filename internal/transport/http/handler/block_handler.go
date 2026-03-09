package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/block"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
)

type BlockHandler struct {
	repo block.Repository
}

func NewBlockHandler(repo block.Repository) *BlockHandler {
	return &BlockHandler{repo: repo}
}

// Block POST /api/v1/users/:id/block
func (h *BlockHandler) Block(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	if uid == targetID {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "不能屏蔽自己"))
		return
	}
	if err := h.repo.Block(c.Request.Context(), uid, targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已屏蔽"})
}

// Unblock DELETE /api/v1/users/:id/block
func (h *BlockHandler) Unblock(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}
	if err := h.repo.Unblock(c.Request.Context(), uid, targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已取消屏蔽"})
}

// ListBlocked GET /api/v1/users/me/blocked
func (h *BlockHandler) ListBlocked(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	ids, err := h.repo.ListBlockedIDs(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, ids)
}
