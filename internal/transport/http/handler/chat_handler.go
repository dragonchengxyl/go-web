package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/transport/ws"
	"github.com/studio/platform/internal/usecase"
	"go.uber.org/zap"
)

// ChatHandler handles chat-related HTTP requests and WebSocket connections
type ChatHandler struct {
	chatService *usecase.ChatService
	hub         *ws.Hub
	logger      *zap.Logger
}

func NewChatHandler(chatService *usecase.ChatService, hub *ws.Hub, logger *zap.Logger) *ChatHandler {
	return &ChatHandler{chatService: chatService, hub: hub, logger: logger}
}

// CreateConversation POST /api/v1/conversations
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		OtherUserID string `json:"other_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	otherID, err := uuid.Parse(req.OtherUserID)
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的用户ID"))
		return
	}

	conv, err := h.chatService.CreateDirectConversation(c.Request.Context(), userID, otherID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, conv)
}

// ListConversations GET /api/v1/conversations
func (h *ChatHandler) ListConversations(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	page, pageSize := getPageParams(c)
	convs, total, err := h.chatService.ListConversations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"conversations": convs, "total": total, "page": page, "size": len(convs)})
}

// ListMessages GET /api/v1/conversations/:id/messages
func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的会话ID"))
		return
	}
	page, pageSize := getPageParams(c)
	msgs, total, err := h.chatService.ListMessages(c.Request.Context(), convID, userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"messages": msgs, "total": total, "page": page, "size": len(msgs)})
}

// SendMessage POST /api/v1/conversations/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的会话ID"))
		return
	}

	var req struct {
		Content  string  `json:"content"`
		MediaURL *string `json:"media_url,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	msg, err := h.chatService.SendMessage(c.Request.Context(), usecase.SendMessageInput{
		ConversationID: convID,
		SenderID:       userID,
		Content:        req.Content,
		MediaURL:       req.MediaURL,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	// Push to WebSocket clients
	conv, _ := h.chatService.GetConversation(c.Request.Context(), convID, userID)
	if conv != nil {
		wsMsg := ws.WSMessage{
			Type:           ws.MessageTypeChat,
			ConversationID: &convID,
			Payload:        msg,
		}
		for _, memberID := range conv.Members {
			if memberID != userID {
				h.hub.SendToUser(memberID, wsMsg)
			}
		}
	}

	response.Success(c, msg)
}

// MarkRead PUT /api/v1/conversations/:id/read
func (h *ChatHandler) MarkRead(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}
	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的会话ID"))
		return
	}
	if err := h.chatService.MarkRead(c.Request.Context(), convID, userID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{"message": "已标记为已读"})
}

// ServeWS GET /ws/chat - WebSocket endpoint
func (h *ChatHandler) ServeWS(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := ws.NewClient(h.hub, c.Writer, c.Request, userID, h.logger); err != nil {
		h.logger.Error("failed to upgrade websocket", zap.Error(err))
	}
}
