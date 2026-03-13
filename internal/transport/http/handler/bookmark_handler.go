package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/bookmark"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

type BookmarkHandler struct {
	service *usecase.BookmarkService
}

func NewBookmarkHandler(service *usecase.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{service: service}
}

func (h *BookmarkHandler) BookmarkPost(c *gin.Context) {
	h.add(c, bookmark.TargetPost, c.Param("id"))
}

func (h *BookmarkHandler) UnbookmarkPost(c *gin.Context) {
	h.remove(c, bookmark.TargetPost, c.Param("id"))
}

func (h *BookmarkHandler) BookmarkGroup(c *gin.Context) {
	h.add(c, bookmark.TargetGroup, c.Param("id"))
}

func (h *BookmarkHandler) UnbookmarkGroup(c *gin.Context) {
	h.remove(c, bookmark.TargetGroup, c.Param("id"))
}

func (h *BookmarkHandler) BookmarkEvent(c *gin.Context) {
	h.add(c, bookmark.TargetEvent, c.Param("id"))
}

func (h *BookmarkHandler) UnbookmarkEvent(c *gin.Context) {
	h.remove(c, bookmark.TargetEvent, c.Param("id"))
}

func (h *BookmarkHandler) Check(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	targetID, err := uuid.Parse(c.Query("target_id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的目标ID"))
		return
	}
	targetType := bookmark.TargetType(c.Query("target_type"))
	if !isValidBookmarkType(targetType) {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的收藏类型"))
		return
	}

	exists, err := h.service.Exists(c.Request.Context(), userID, targetType, targetID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"bookmarked": exists})
}

func (h *BookmarkHandler) ListPosts(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	sort := c.DefaultQuery("sort", "latest")
	posts, total, err := h.service.ListPosts(c.Request.Context(), userID, page, pageSize, sort)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"posts": posts, "total": total, "page": page, "size": len(posts)})
}

func (h *BookmarkHandler) ListGroups(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	sort := c.DefaultQuery("sort", "latest")
	groups, total, err := h.service.ListGroups(c.Request.Context(), userID, page, pageSize, sort)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"groups": groups, "total": total, "page": page, "size": len(groups)})
}

func (h *BookmarkHandler) ListEvents(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	sort := c.DefaultQuery("sort", "latest")
	events, total, err := h.service.ListEvents(c.Request.Context(), userID, page, pageSize, sort)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"events": events, "total": total, "page": page, "size": len(events)})
}

func (h *BookmarkHandler) add(c *gin.Context, targetType bookmark.TargetType, rawID string) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	targetID, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的目标ID"))
		return
	}
	if err := h.service.Add(c.Request.Context(), userID, targetType, targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "收藏成功"})
}

func (h *BookmarkHandler) remove(c *gin.Context, targetType bookmark.TargetType, rawID string) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	targetID, err := uuid.Parse(rawID)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的目标ID"))
		return
	}
	if err := h.service.Remove(c.Request.Context(), userID, targetType, targetID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已取消收藏"})
}

func (h *BookmarkHandler) RemoveBatch(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		TargetType string   `json:"target_type" binding:"required"`
		TargetIDs  []string `json:"target_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	targetType := bookmark.TargetType(req.TargetType)
	if !isValidBookmarkType(targetType) {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的收藏类型"))
		return
	}

	ids := make([]uuid.UUID, 0, len(req.TargetIDs))
	for _, rawID := range req.TargetIDs {
		id, err := uuid.Parse(rawID)
		if err != nil {
			response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的目标ID"))
			return
		}
		ids = append(ids, id)
	}

	if err := h.service.RemoveBatch(c.Request.Context(), userID, targetType, ids); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "批量移除完成"})
}

func isValidBookmarkType(t bookmark.TargetType) bool {
	return t == bookmark.TargetPost || t == bookmark.TargetGroup || t == bookmark.TargetEvent
}
