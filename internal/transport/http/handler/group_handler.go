package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// GroupHandler handles group HTTP endpoints.
type GroupHandler struct {
	groupSvc    *usecase.GroupService
	userService *usecase.UserService
}

// NewGroupHandler creates a new GroupHandler.
func NewGroupHandler(groupSvc *usecase.GroupService, userService *usecase.UserService) *GroupHandler {
	return &GroupHandler{groupSvc: groupSvc, userService: userService}
}

// ListGroups handles GET /api/v1/groups
func (h *GroupHandler) ListGroups(c *gin.Context) {
	page, pageSize := getPageParams(c)
	search := c.Query("search")

	var privacy *group.GroupPrivacy
	if p := c.Query("privacy"); p != "" {
		gp := group.GroupPrivacy(p)
		privacy = &gp
	}

	groups, total, err := h.groupSvc.ListGroups(c.Request.Context(), usecase.ListGroupsInput{
		Privacy:  privacy,
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"groups": groups, "total": total, "page": page, "page_size": pageSize})
}

// GetGroup handles GET /api/v1/groups/:id
func (h *GroupHandler) GetGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	g, err := h.groupSvc.GetGroup(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, g)
}

type createGroupRequest struct {
	Name        string             `json:"name" binding:"required"`
	Description string             `json:"description"`
	Tags        []string           `json:"tags"`
	Privacy     group.GroupPrivacy `json:"privacy"`
}

// CreateGroup handles POST /api/v1/groups
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	if h.userService != nil {
		u, err := h.userService.GetProfile(c.Request.Context(), userID)
		if err != nil {
			response.Error(c, err)
			return
		}
		if u.EmailVerifiedAt == nil {
			response.Error(c, apperr.New(apperr.CodeForbidden, "请先验证邮箱后再创建圈子"))
			return
		}
	}

	var req createGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest(err.Error()))
		return
	}

	g, err := h.groupSvc.CreateGroup(c.Request.Context(), usecase.CreateGroupInput{
		OwnerID:     userID,
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Privacy:     req.Privacy,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, g)
}

type updateGroupRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Tags        []string           `json:"tags"`
	Privacy     group.GroupPrivacy `json:"privacy"`
}

// UpdateGroup handles PUT /api/v1/groups/:id
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}

	var req updateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest(err.Error()))
		return
	}

	g, err := h.groupSvc.UpdateGroup(c.Request.Context(), userID, id, usecase.UpdateGroupInput{
		Name:        req.Name,
		Description: req.Description,
		Tags:        req.Tags,
		Privacy:     req.Privacy,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, g)
}

// JoinGroup handles POST /api/v1/groups/:id/join
func (h *GroupHandler) JoinGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	if err := h.groupSvc.JoinGroup(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// LeaveGroup handles DELETE /api/v1/groups/:id/leave
func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	if err := h.groupSvc.LeaveGroup(c.Request.Context(), id, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// ListMembers handles GET /api/v1/groups/:id/members
func (h *GroupHandler) ListMembers(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	page, pageSize := getPageParams(c)
	members, total, err := h.groupSvc.ListMembers(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"members": members, "total": total})
}

type updateMemberRoleRequest struct {
	Role group.GroupRole `json:"role" binding:"required"`
}

// UpdateMemberRole handles PUT /api/v1/groups/:id/members/:uid
func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	targetUID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的用户ID"))
		return
	}
	var req updateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest(err.Error()))
		return
	}
	if err := h.groupSvc.UpdateMemberRole(c.Request.Context(), userID, groupID, targetUID, req.Role); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// KickMember handles DELETE /api/v1/groups/:id/members/:uid
func (h *GroupHandler) KickMember(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的圈子ID"))
		return
	}
	targetUID, err := uuid.Parse(c.Param("uid"))
	if err != nil {
		response.Error(c, apperr.BadRequest("无效的用户ID"))
		return
	}
	if err := h.groupSvc.KickMember(c.Request.Context(), userID, groupID, targetUID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}

// MyGroups handles GET /api/v1/users/me/groups
func (h *GroupHandler) MyGroups(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.New(apperr.CodeUnauthorized, "未登录"))
		return
	}
	page, pageSize := getPageParams(c)
	groups, total, err := h.groupSvc.MyGroups(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"groups": groups, "total": total})
}
