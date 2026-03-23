package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/audit"
	"github.com/studio/platform/internal/domain/event"
	"github.com/studio/platform/internal/domain/gameplay"
	"github.com/studio/platform/internal/domain/group"
	"github.com/studio/platform/internal/domain/notification"
	"github.com/studio/platform/internal/domain/order"
	"github.com/studio/platform/internal/domain/permission"
	"github.com/studio/platform/internal/domain/post"
	"github.com/studio/platform/internal/domain/report"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/observability/audiometrics"
	"github.com/studio/platform/internal/observability/gamemetrics"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AdminHandler handles admin HTTP requests
type AdminHandler struct {
	statsService     usecase.StatsProvider
	userService      *usecase.UserService
	gameService      *usecase.HexBlitzRoomService
	commentService   *usecase.CommentService
	postService      *usecase.PostService
	audioWorkService *usecase.AudioWorkService
	groupService     *usecase.GroupService
	eventService     *usecase.EventService
	auditService     *usecase.AuditService
	aiToolService    *usecase.AdminAIToolService
	config           *configs.Config
	sponsorService   *usecase.SponsorSettingsService
	orderRepo        order.Repository
	reportRepo       report.Repository
	notifyService    *usecase.NotificationService
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(
	statsService usecase.StatsProvider,
	userService *usecase.UserService,
	gameService *usecase.HexBlitzRoomService,
	commentService *usecase.CommentService,
	postService *usecase.PostService,
	audioWorkService *usecase.AudioWorkService,
	groupService *usecase.GroupService,
	eventService *usecase.EventService,
	auditService *usecase.AuditService,
	aiToolService *usecase.AdminAIToolService,
	config *configs.Config,
	sponsorService *usecase.SponsorSettingsService,
	orderRepo order.Repository,
	reportRepo report.Repository,
	notifyService *usecase.NotificationService,
) *AdminHandler {
	return &AdminHandler{
		statsService:     statsService,
		userService:      userService,
		gameService:      gameService,
		commentService:   commentService,
		postService:      postService,
		audioWorkService: audioWorkService,
		groupService:     groupService,
		eventService:     eventService,
		auditService:     auditService,
		aiToolService:    aiToolService,
		config:           config,
		sponsorService:   sponsorService,
		orderRepo:        orderRepo,
		reportRepo:       reportRepo,
		notifyService:    notifyService,
	}
}

// GetGameOverview returns Hex Blitz runtime metrics, active rooms, leaderboard and recent matches.
func (h *AdminHandler) GetGameOverview(c *gin.Context) {
	gameMetrics := gamemetrics.GetSnapshot()

	var rooms []*gameplay.Room
	var leaderboard []*gameplay.LeaderboardEntry
	var recentMatches []*gameplay.MatchSummary
	if h.gameService != nil {
		rooms = h.gameService.ListRooms()
		entries, err := h.gameService.ListLeaderboard(c.Request.Context(), 10)
		if err != nil {
			response.Error(c, err)
			return
		}
		leaderboard = entries
		matches, err := h.gameService.ListRecentMatches(c.Request.Context(), 8)
		if err != nil {
			response.Error(c, err)
			return
		}
		recentMatches = matches
	}

	response.Success(c, gin.H{
		"metrics":        gameMetrics,
		"rooms":          rooms,
		"leaderboard":    leaderboard,
		"recent_matches": recentMatches,
	})
}

// ListAudioWorks returns paginated audio works filtered by moderation status (admin).
func (h *AdminHandler) ListAudioWorks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	works, total, err := h.audioWorkService.AdminListWorks(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"works": works,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// UpdateAudioWorkModeration updates an audio work moderation status (admin).
func (h *AdminHandler) UpdateAudioWorkModeration(c *gin.Context) {
	workID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	ms := post.ModerationStatus(input.Status)
	if ms != post.ModerationPending && ms != post.ModerationApproved && ms != post.ModerationBlocked {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的审核状态"))
		return
	}

	if err := h.audioWorkService.AdminUpdateModerationStatus(c.Request.Context(), workID, ms, input.Note); err != nil {
		response.Error(c, err)
		return
	}
	h.logAudit(c, audit.ActionUpdate, audit.ResourceAudioWork, &workID, gin.H{"status": input.Status, "note": input.Note})
	response.Success(c, gin.H{"status": input.Status, "note": input.Note})
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
	if roleStr := c.Query("role"); roleStr != "" {
		r := user.Role(roleStr)
		input.Role = &r
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
	h.logAudit(c, audit.ActionUpdate, audit.ResourceUser, &userID, gin.H{"role": input.Role})
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
	h.logAudit(c, audit.ActionUpdate, audit.ResourceUser, &userID, gin.H{"status": user.StatusBanned})
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
	h.logAudit(c, audit.ActionUpdate, audit.ResourceUser, &userID, gin.H{"status": user.StatusActive})
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
	h.logAudit(c, audit.ActionDelete, audit.ResourceUser, &userID, nil)
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
	h.logAudit(c, audit.ActionDelete, audit.ResourceComment, &commentID, nil)
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
	h.logAudit(c, audit.ActionUpdate, audit.ResourcePost, &postID, gin.H{"status": input.Status})
	response.Success(c, gin.H{"status": input.Status})
}

// ListGroups returns paginated groups for admin operations.
func (h *AdminHandler) ListGroups(c *gin.Context) {
	if h.groupService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")

	var privacy *group.GroupPrivacy
	if raw := c.Query("privacy"); raw != "" {
		gp := group.GroupPrivacy(raw)
		privacy = &gp
	}

	items, total, err := h.groupService.ListGroups(c.Request.Context(), usecase.ListGroupsInput{
		Privacy:  privacy,
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	type adminGroupItem struct {
		*group.Group
		OwnerUsername string `json:"owner_username,omitempty"`
		OwnerEmail    string `json:"owner_email,omitempty"`
	}

	cache := make(map[uuid.UUID]*user.User)
	result := make([]adminGroupItem, 0, len(items))
	for _, item := range items {
		enriched := adminGroupItem{Group: item}
		if owner := h.resolveUser(c, cache, item.OwnerID); owner != nil {
			enriched.OwnerUsername = owner.Username
			enriched.OwnerEmail = owner.Email
		}
		result = append(result, enriched)
	}

	response.Success(c, gin.H{"groups": result, "total": total, "page": page, "size": pageSize})
}

// UpdateGroup updates admin-managed group settings.
func (h *AdminHandler) UpdateGroup(c *gin.Context) {
	if h.groupService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	groupID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var req struct {
		Privacy group.GroupPrivacy `json:"privacy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	item, err := h.groupService.AdminUpdatePrivacy(c.Request.Context(), groupID, req.Privacy)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.logAudit(c, audit.ActionUpdate, audit.ResourceGroup, &groupID, gin.H{"privacy": req.Privacy})
	response.Success(c, item)
}

// ListEvents returns paginated events for admin operations.
func (h *AdminHandler) ListEvents(c *gin.Context) {
	if h.eventService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var status *event.EventStatus
	if raw := c.Query("status"); raw != "" {
		es := event.EventStatus(raw)
		status = &es
	}

	items, total, err := h.eventService.ListEvents(c.Request.Context(), usecase.ListEventsInput{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	type adminEventItem struct {
		*event.Event
		OrganizerUsername string `json:"organizer_username,omitempty"`
		OrganizerEmail    string `json:"organizer_email,omitempty"`
	}

	cache := make(map[uuid.UUID]*user.User)
	result := make([]adminEventItem, 0, len(items))
	for _, item := range items {
		enriched := adminEventItem{Event: item}
		if organizer := h.resolveUser(c, cache, item.OrganizerID); organizer != nil {
			enriched.OrganizerUsername = organizer.Username
			enriched.OrganizerEmail = organizer.Email
		}
		result = append(result, enriched)
	}

	response.Success(c, gin.H{"events": result, "total": total, "page": page, "size": pageSize})
}

// UpdateEventStatus updates an event status from the admin console.
func (h *AdminHandler) UpdateEventStatus(c *gin.Context) {
	if h.eventService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var req struct {
		Status event.EventStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	item, err := h.eventService.AdminUpdateStatus(c.Request.Context(), eventID, req.Status)
	if err != nil {
		response.Error(c, err)
		return
	}
	h.logAudit(c, audit.ActionUpdate, audit.ResourceEvent, &eventID, gin.H{"status": req.Status})
	response.Success(c, item)
}

// ListOrders returns paginated orders for admin operations.
func (h *AdminHandler) ListOrders(c *gin.Context) {
	if h.orderRepo == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	search := c.Query("search")
	orderType := c.Query("type")

	var status *order.OrderStatus
	if raw := c.Query("status"); raw != "" {
		os := order.OrderStatus(raw)
		status = &os
	}

	var method *order.PaymentMethod
	if raw := c.Query("payment_method"); raw != "" {
		pm := order.PaymentMethod(raw)
		method = &pm
	}

	items, total, err := h.orderRepo.List(c.Request.Context(), order.ListFilter{
		Status:        status,
		PaymentMethod: method,
		Type:          orderType,
		Search:        search,
		Page:          page,
		PageSize:      pageSize,
	})
	if err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "查询订单失败", err))
		return
	}

	type adminOrderItem struct {
		*order.Order
		OrderType         string `json:"order_type,omitempty"`
		RecipientUserID   string `json:"recipient_user_id,omitempty"`
		PayerUsername     string `json:"payer_username,omitempty"`
		PayerEmail        string `json:"payer_email,omitempty"`
		RecipientUsername string `json:"recipient_username,omitempty"`
		RecipientEmail    string `json:"recipient_email,omitempty"`
	}

	cache := make(map[uuid.UUID]*user.User)
	result := make([]adminOrderItem, 0, len(items))
	for _, item := range items {
		enriched := adminOrderItem{Order: item}
		if payer := h.resolveUser(c, cache, item.UserID); payer != nil {
			enriched.PayerUsername = payer.Username
			enriched.PayerEmail = payer.Email
		}
		if orderTypeValue, ok := item.Metadata["type"].(string); ok {
			enriched.OrderType = orderTypeValue
		}
		if recipientID, ok := item.Metadata["to_user_id"].(string); ok && recipientID != "" {
			enriched.RecipientUserID = recipientID
			if uid, parseErr := uuid.Parse(recipientID); parseErr == nil {
				if recipient := h.resolveUser(c, cache, uid); recipient != nil {
					enriched.RecipientUsername = recipient.Username
					enriched.RecipientEmail = recipient.Email
				}
			}
		}
		result = append(result, enriched)
	}

	response.Success(c, gin.H{"orders": result, "total": total, "page": page, "size": pageSize})
}

// UpdateOrderStatus updates an order status from the admin console.
func (h *AdminHandler) UpdateOrderStatus(c *gin.Context) {
	if h.orderRepo == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	var req struct {
		Status order.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	switch req.Status {
	case order.OrderStatusPendingPayment, order.OrderStatusPaid, order.OrderStatusFulfilled, order.OrderStatusCancelled, order.OrderStatusFailed, order.OrderStatusRefunded:
	default:
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的订单状态"))
		return
	}

	if err := h.orderRepo.UpdateStatus(c.Request.Context(), orderID, req.Status); err != nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "更新订单状态失败", err))
		return
	}
	h.logAudit(c, audit.ActionUpdate, audit.ResourceOrder, &orderID, gin.H{"status": req.Status})
	response.Success(c, gin.H{"status": req.Status})
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
	h.logAudit(c, audit.ActionUpdate, audit.ResourceReport, &reportID, gin.H{"status": input.Status, "action": input.Action})

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
				case report.ActionBlockAudioWork:
					targetType = "report_audio_work_blocked"
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

// ListAuditLogs returns paginated audit logs (admin).
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	if h.auditService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	input, err := h.parseAuditLogFilters(c, true)
	if err != nil {
		response.Error(c, err)
		return
	}
	result, err := h.auditService.ListAuditLogs(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, result)
}

// ExportAuditLogs returns CSV export for audit logs (admin).
func (h *AdminHandler) ExportAuditLogs(c *gin.Context) {
	if h.auditService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	input, err := h.parseAuditLogFilters(c, false)
	if err != nil {
		response.Error(c, err)
		return
	}
	input.Page = 1
	input.PageSize = 1000

	result, err := h.auditService.ListAuditLogs(c.Request.Context(), input)
	if err != nil {
		response.Error(c, err)
		return
	}

	filename := "audit_logs_" + time.Now().Format("20060102_150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{
		"created_at",
		"username",
		"user_id",
		"action",
		"resource",
		"resource_id",
		"ip_address",
		"user_agent",
		"before_data",
		"after_data",
		"error_message",
	})

	for _, logItem := range result.Logs {
		var userID, resourceID, beforeData, afterData, errorMessage string
		if logItem.UserID != nil {
			userID = logItem.UserID.String()
		}
		if logItem.ResourceID != nil {
			resourceID = logItem.ResourceID.String()
		}
		if logItem.BeforeData != nil {
			beforeData = *logItem.BeforeData
		}
		if logItem.AfterData != nil {
			afterData = *logItem.AfterData
		}
		if logItem.ErrorMessage != nil {
			errorMessage = *logItem.ErrorMessage
		}

		_ = writer.Write([]string{
			logItem.CreatedAt.Format(time.RFC3339),
			logItem.Username,
			userID,
			string(logItem.Action),
			string(logItem.Resource),
			resourceID,
			logItem.IPAddress,
			logItem.UserAgent,
			beforeData,
			afterData,
			errorMessage,
		})
	}
}

// GetPermissionMatrix returns role-permission mappings (admin).
func (h *AdminHandler) GetPermissionMatrix(c *gin.Context) {
	matrix := []gin.H{}
	roles := []user.Role{
		user.RoleSuperAdmin,
		user.RoleAdmin,
		user.RoleModerator,
		user.RoleCreator,
		user.RoleSupporter,
		user.RoleMember,
		user.RoleGuest,
	}

	for _, role := range roles {
		perms := permission.GetPermissionStrings(role)
		matrix = append(matrix, gin.H{
			"role":        role,
			"permissions": perms,
			"count":       len(perms),
		})
	}

	catalog := gin.H{
		"dashboard": []string{string(permission.DashboardView)},
		"user": []string{
			string(permission.UserView),
			string(permission.UserUpdate),
			string(permission.UserDelete),
			string(permission.UserManage),
		},
		"comment": []string{
			string(permission.CommentCreate),
			string(permission.CommentDeleteOwn),
			string(permission.CommentDeleteAny),
			string(permission.CommentUpdate),
		},
		"game": []string{
			string(permission.GameView),
			string(permission.GameManage),
			string(permission.GameReleaseCreate),
			string(permission.GameReleaseUpdate),
			string(permission.GameReleaseDelete),
		},
		"ost": []string{
			string(permission.OSTView),
			string(permission.OSTDownloadHiFi),
			string(permission.OSTManage),
		},
	}

	response.Success(c, gin.H{
		"roles":   matrix,
		"catalog": catalog,
	})
}

// GetSystemConfig returns a safe runtime config snapshot for the admin console.
func (h *AdminHandler) GetSystemConfig(c *gin.Context) {
	if h.config == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	sponsorCfg := h.config.Sponsor
	if h.sponsorService != nil {
		if cfg, err := h.sponsorService.Get(c.Request.Context()); err == nil {
			sponsorCfg = cfg
		}
	}
	audioMetrics := audiometrics.GetSnapshot()

	response.Success(c, gin.H{
		"server": gin.H{
			"mode":          h.config.Server.Mode,
			"port":          h.config.Server.Port,
			"frontend_url":  h.config.Server.FrontendURL,
			"allow_origins": h.config.Server.AllowOrigins,
		},
		"ratelimit": gin.H{
			"unauthenticated": h.config.RateLimit.Unauthenticated,
			"authenticated":   h.config.RateLimit.Authenticated,
			"admin":           h.config.RateLimit.Admin,
		},
		"sponsor": gin.H{
			"monthly_goal":   sponsorCfg.MonthlyGoal,
			"current_raised": sponsorCfg.CurrentRaised,
			"alipay_qr_url":  sponsorCfg.AlipayQRURL,
			"wechat_qr_url":  sponsorCfg.WechatQRURL,
			"message":        sponsorCfg.Message,
		},
		"assistant": gin.H{
			"provider":             h.config.Assistant.Provider,
			"base_url":             h.config.Assistant.BaseURL,
			"model":                h.config.Assistant.Model,
			"embedding_base_url":   h.config.Assistant.EmbeddingBaseURL,
			"embedding_model":      h.config.Assistant.EmbeddingModel,
			"embedding_dims":       h.config.Assistant.EmbeddingDims,
			"vision_base_url":      h.config.Assistant.VisionBaseURL,
			"vision_model":         h.config.Assistant.VisionModel,
			"timeout_sec":          h.config.Assistant.TimeoutSec,
			"vision_timeout_sec":   h.config.Assistant.VisionTimeoutSec,
			"max_context_items":    h.config.Assistant.MaxContextItems,
			"persona_name":         h.config.Assistant.PersonaName,
			"retrieval_limit":      h.config.Assistant.RetrievalLimit,
			"vector_scan_limit":    h.config.Assistant.VectorScanLimit,
			"sync_interval_sec":    h.config.Assistant.SyncIntervalSec,
			"configured":           h.config.Assistant.APIKey != "",
			"embedding_configured": h.config.Assistant.EmbeddingAPIKey != "",
			"vision_configured":    h.config.Assistant.VisionAPIKey != "",
		},
		"oss": gin.H{
			"provider":      h.config.OSS.Provider,
			"bucket":        h.config.OSS.Bucket,
			"endpoint":      h.config.OSS.Endpoint,
			"region":        h.config.OSS.Region,
			"allowed_hosts": h.config.OSS.AllowedHosts,
		},
		"email": gin.H{
			"configured": h.config.Email.Host != "",
			"host":       h.config.Email.Host,
			"from":       h.config.Email.From,
		},
		"payment": gin.H{
			"alipay_configured": h.config.Payment.Alipay.AppID != "",
			"wechat_configured": h.config.Payment.Wechat.AppID != "",
		},
		"grpc": gin.H{
			"stats_addr":        h.config.GRPC.StatsAddr,
			"notification_addr": h.config.GRPC.NotificationAddr,
			"moderation_addr":   h.config.GRPC.ModerationAddr,
			"stats_port":        h.config.GRPC.StatsPort,
			"notification_port": h.config.GRPC.NotificationPort,
			"moderation_port":   h.config.GRPC.ModerationPort,
		},
		"audio_metrics": audioMetrics,
	})
}

// UpdateSponsorConfig updates runtime sponsor display config.
func (h *AdminHandler) UpdateSponsorConfig(c *gin.Context) {
	if h.config == nil || h.sponsorService == nil {
		response.Error(c, apperr.ErrNotFound)
		return
	}

	var req struct {
		MonthlyGoal   float64 `json:"monthly_goal"`
		CurrentRaised float64 `json:"current_raised"`
		AlipayQRURL   string  `json:"alipay_qr_url"`
		WechatQRURL   string  `json:"wechat_qr_url"`
		Message       string  `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}
	if req.MonthlyGoal < 0 || req.CurrentRaised < 0 {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "金额不能为负数"))
		return
	}

	var updatedBy *uuid.UUID
	if uid, ok := getUserID(c); ok {
		updatedBy = &uid
	}

	cfg, err := h.sponsorService.Update(c.Request.Context(), configs.SponsorConfig{
		MonthlyGoal:   req.MonthlyGoal,
		CurrentRaised: req.CurrentRaised,
		AlipayQRURL:   req.AlipayQRURL,
		WechatQRURL:   req.WechatQRURL,
		Message:       req.Message,
	}, updatedBy)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.logAudit(c, audit.ActionUpdate, audit.ResourceSystem, nil, gin.H{
		"monthly_goal":   req.MonthlyGoal,
		"current_raised": req.CurrentRaised,
		"message":        req.Message,
	})

	response.Success(c, gin.H{
		"monthly_goal":   cfg.MonthlyGoal,
		"current_raised": cfg.CurrentRaised,
		"alipay_qr_url":  cfg.AlipayQRURL,
		"wechat_qr_url":  cfg.WechatQRURL,
		"message":        cfg.Message,
	})
}

func (h *AdminHandler) resolveUser(c *gin.Context, cache map[uuid.UUID]*user.User, userID uuid.UUID) *user.User {
	if userID == uuid.Nil || h.userService == nil {
		return nil
	}
	if cached, ok := cache[userID]; ok {
		return cached
	}
	item, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		cache[userID] = nil
		return nil
	}
	cache[userID] = item
	return item
}

func (h *AdminHandler) parseAuditLogFilters(c *gin.Context, usePagination bool) (usecase.ListAuditLogsInput, error) {
	input := usecase.ListAuditLogsInput{}
	if usePagination {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		input.Page = page
		input.PageSize = pageSize
	}

	if raw := c.Query("user_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return input, apperr.New(apperr.CodeInvalidParam, "无效的用户ID")
		}
		input.UserID = &parsed
	}

	if raw := c.Query("resource_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return input, apperr.New(apperr.CodeInvalidParam, "无效的资源ID")
		}
		input.ResourceID = &parsed
	}

	if raw := c.Query("action"); raw != "" {
		next := audit.Action(raw)
		input.Action = &next
	}

	if raw := c.Query("resource"); raw != "" {
		next := audit.Resource(raw)
		input.Resource = &next
	}

	input.StartTime = optionalStringPtr(c.Query("start_time"))
	input.EndTime = optionalStringPtr(c.Query("end_time"))
	return input, nil
}

func (h *AdminHandler) logAudit(c *gin.Context, action audit.Action, resource audit.Resource, resourceID *uuid.UUID, afterData any) {
	if h.auditService == nil {
		return
	}

	var userID *uuid.UUID
	var username string
	if value, ok := getUserID(c); ok {
		userID = &value
		if userItem, err := h.userService.GetUserByID(c.Request.Context(), value); err == nil {
			username = userItem.Username
		}
	}
	if username == "" {
		username = "admin"
	}

	_ = h.auditService.Log(c.Request.Context(), usecase.LogInput{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
		AfterData:  afterData,
	})
}

func optionalStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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
	case report.TargetTypeAudioWork:
		if action != report.ActionBlockAudioWork {
			return nil, apperr.New(apperr.CodeInvalidParam, "该举报类型不支持此动作")
		}
		if err := h.audioWorkService.AdminUpdateModerationStatus(c.Request.Context(), rep.TargetID, post.ModerationBlocked, "由举报处理触发封禁"); err != nil {
			return nil, err
		}
	default:
		return nil, apperr.New(apperr.CodeInvalidParam, "未知举报目标类型")
	}

	return &action, nil
}
