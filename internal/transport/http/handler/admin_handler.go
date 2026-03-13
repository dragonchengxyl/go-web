package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/domain/report"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	statsService   usecase.StatsProvider
	userService    *usecase.UserService
	commentService *usecase.CommentService
	postService    *usecase.PostService
	reportRepo     report.Repository
	notifyService  *usecase.NotificationService
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(
	statsService usecase.StatsProvider,
	userService *usecase.UserService,
	_ any, // was gameService - no longer needed
	commentService *usecase.CommentService,
	postService *usecase.PostService,
	reportRepo report.Repository,
	notifyService *usecase.NotificationService,
) *AdminHandler {
	return &AdminHandler{
		statsService:   statsService,
		userService:    userService,
		commentService: commentService,
		postService:    postService,
		reportRepo:     reportRepo,
		notifyService:  notifyService,
	}
}

// GetDashboardStats returns main dashboard metrics
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.statsService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, stats)
}

// GetUserGrowthChart returns daily user registration chart data
func (h *AdminHandler) GetUserGrowthChart(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	data, err := h.statsService.GetUserGrowthChart(c.Request.Context(), days)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, data)
}

// ListUsers returns paginated user list with filters
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	input := usecase.ListUsersInput{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		s := user.Status(statusStr)
		input.Status = &s
	}

	result, err := h.userService.ListUsers(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// UpdateUserRole changes a user's role
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Role user.Role `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	u, err := h.userService.UpdateUserRole(c.Request.Context(), userID, input.Role)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// BanUser bans a user
func (h *AdminHandler) BanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	u, err := h.userService.UpdateUserStatus(c.Request.Context(), userID, user.StatusBanned)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// UnbanUser unbans a user
func (h *AdminHandler) UnbanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	u, err := h.userService.UpdateUserStatus(c.Request.Context(), userID, user.StatusActive)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, u)
}

// DeleteUser deletes a user
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), userID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListComments returns paginated comments for moderation
func (h *AdminHandler) ListComments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.commentService.AdminListComments(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// DeleteComment deletes a comment (admin)
func (h *AdminHandler) DeleteComment(c *gin.Context) {
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	if err := h.commentService.AdminDeleteComment(c.Request.Context(), commentID); err != nil {
		response.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPosts returns paginated posts filtered by moderation_status (admin)
func (h *AdminHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status") // pending | approved | blocked | ""

	posts, total, err := h.postService.AdminListPosts(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"posts": posts,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdatePostModeration updates a post's moderation_status (admin)
func (h *AdminHandler) UpdatePostModeration(c *gin.Context) {
	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	ms := post.ModerationStatus(input.Status)
	if ms != post.ModerationApproved && ms != post.ModerationBlocked && ms != post.ModerationPending {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的审核状态"))
		return
	}

	if err := h.postService.AdminUpdateModerationStatus(c.Request.Context(), postID, ms); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"status": input.Status})
}

// ListReports returns paginated reports (admin)
func (h *AdminHandler) ListReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status") // pending | reviewed | dismissed | ""

	reports, total, err := h.reportRepo.List(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"reports": reports,
		"total":   total,
		"page":    page,
		"size":    pageSize,
	})
}

// UpdateReport updates a report's status (admin)
func (h *AdminHandler) UpdateReport(c *gin.Context) {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
		Action string `json:"action"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	rs := report.Status(input.Status)
	if rs != report.StatusReviewed && rs != report.StatusDismissed {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的举报状态"))
		return
	}

	reviewerID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	rep, err := h.reportRepo.GetByID(c.Request.Context(), reportID)
	if err != nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var actionTaken *report.Action
	if rs == report.StatusReviewed {
		action, err := h.applyReportAction(c, rep, input.Action)
		if err != nil {
			response.Error(c, err)
			return
		}
		actionTaken = action
	}

	if err := h.reportRepo.UpdateStatus(c.Request.Context(), reportID, rs, reviewerID, actionTaken); err != nil {
		response.Error(c, err)
		return
	}

	if h.notifyService != nil {
		targetType := "report_dismissed"
		if rs == report.StatusReviewed {
			targetType = "report_reviewed"
			if actionTaken != nil {
				switch *actionTaken {
				case report.ActionBlockPost:
					targetType = "report_post_blocked"
				case report.ActionDeleteComment:
					targetType = "report_comment_deleted"
				case report.ActionBanUser:
					targetType = "report_user_banned"
				}
			}
		}
		_ = h.notifyService.Notify(c.Request.Context(), &notification.Notification{
			UserID:     rep.ReporterID,
			Type:       notification.TypeSystem,
			TargetID:   &rep.TargetID,
			TargetType: targetType,
		})
	}

	response.Success(c, gin.H{"status": input.Status, "action": input.Action})
}

func (h *AdminHandler) applyReportAction(c *gin.Context, rep *report.Report, rawAction string) (*report.Action, error) {
	if rawAction == "" {
		return nil, nil
	}

	action := report.Action(rawAction)
	switch rep.TargetType {
	case report.TargetTypePost:
		if action != report.ActionBlockPost {
			return nil, apperr.New(apperr.CodeInvalidParam, "该举报类型不支持此动作")
		}
		if err := h.postService.AdminUpdateModerationStatus(c.Request.Context(), rep.TargetID, post.ModerationBlocked); err != nil {
			return nil, err
		}
	case report.TargetTypeComment:
		if action != report.ActionDeleteComment {
			return nil, apperr.New(apperr.CodeInvalidParam, "该举报类型不支持此动作")
		}
		if err := h.commentService.AdminDeleteComment(c.Request.Context(), rep.TargetID); err != nil {
			return nil, err
		}
	case report.TargetTypeUser:
		if action != report.ActionBanUser {
			return nil, apperr.New(apperr.CodeInvalidParam, "该举报类型不支持此动作")
		}
		if _, err := h.userService.UpdateUserStatus(c.Request.Context(), rep.TargetID, user.StatusBanned); err != nil {
			return nil, err
		}
	default:
		return nil, apperr.New(apperr.CodeInvalidParam, "未知举报目标类型")
	}

	return &action, nil
}
