package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
	"github.com/studio/platform/internal/pkg/apperr"
	"github.com/studio/platform/internal/pkg/response"
	"github.com/studio/platform/internal/usecase"
)

// AssistantHandler serves the lightweight site AI assistant.
type AssistantHandler struct {
	service *usecase.AssistantService
	timeout time.Duration
}

// NewAssistantHandler creates a handler for SSE chat responses.
func NewAssistantHandler(service *usecase.AssistantService, timeout time.Duration) *AssistantHandler {
	if timeout <= 0 {
		timeout = 90 * time.Second
	}
	return &AssistantHandler{service: service, timeout: timeout}
}

// StreamChat handles POST /api/v1/assistant/chat/stream.
func (h *AssistantHandler) StreamChat(c *gin.Context) {
	if h.service == nil {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "AI 助手未初始化", nil))
		return
	}

	var req struct {
		ConversationID string                         `json:"conversation_id"`
		Messages       []usecase.AssistantChatMessage `json:"messages" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, apperr.Wrap(apperr.CodeInternalError, "当前环境不支持流式响应", nil))
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	var persistedConversationID uuid.UUID
	streamMessages := req.Messages
	if userID, ok := getUserID(c); ok && h.service.HistoryEnabled() {
		latestUser := latestAssistantUserMessage(req.Messages)
		var conversationID *uuid.UUID
		if strings.TrimSpace(req.ConversationID) != "" {
			parsed, err := uuid.Parse(req.ConversationID)
			if err != nil {
				response.Error(c, apperr.New(apperr.CodeInvalidParam, "无效的会话ID"))
				return
			}
			conversationID = &parsed
		}

		conv, historyMessages, err := h.service.PrepareConversation(ctx, userID, conversationID, latestUser)
		if err != nil {
			response.Error(c, err)
			return
		}
		persistedConversationID = conv.ID
		streamMessages = historyMessages
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	flusher.Flush()

	writeEvent := func(event string, payload any) error {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	var reply strings.Builder
	var cards []usecase.AssistantCard
	if err := h.service.StreamReply(
		ctx,
		streamMessages,
		func(meta usecase.AssistantMeta) error {
			if persistedConversationID != uuid.Nil {
				meta.ConversationID = persistedConversationID.String()
			}
			cards = meta.Cards
			return writeEvent("meta", meta)
		},
		func(token string) error {
			reply.WriteString(token)
			return writeEvent("token", gin.H{"content": token})
		},
	); err != nil {
		_ = writeEvent("error", gin.H{"message": err.Error()})
		return
	}

	if persistedConversationID != uuid.Nil {
		if err := h.service.SaveAssistantReply(ctx, persistedConversationID, reply.String(), cards); err != nil {
			_ = writeEvent("error", gin.H{"message": err.Error()})
			return
		}
	}

	_ = writeEvent("done", gin.H{"ok": true})
}

// ListConversations handles GET /api/v1/assistant/conversations.
func (h *AssistantHandler) ListConversations(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	page, pageSize := getPageParams(c)
	items, total, err := h.service.ListConversations(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, gin.H{
		"conversations": items,
		"total":         total,
		"page":          page,
		"size":          pageSize,
	})
}

// GetConversation handles GET /api/v1/assistant/conversations/:id.
func (h *AssistantHandler) GetConversation(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	conversationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperr.ErrInvalidParam)
		return
	}

	page, pageSize := getPageParams(c)
	conv, messages, total, err := h.service.GetConversation(c.Request.Context(), userID, conversationID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"conversation": conv,
		"messages":     messages,
		"total":        total,
		"page":         page,
		"size":         pageSize,
	})
}

// GetSettings handles GET /api/v1/admin/assistant/settings.
func (h *AssistantHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, settings)
}

// UpdateSettings handles PUT /api/v1/admin/assistant/settings.
func (h *AssistantHandler) UpdateSettings(c *gin.Context) {
	adminID, ok := getUserID(c)
	if !ok {
		response.Error(c, apperr.ErrUnauthorized)
		return
	}

	var req struct {
		Enabled         bool   `json:"enabled"`
		PersonaName     string `json:"persona_name"`
		SystemPrompt    string `json:"system_prompt"`
		MaxContextItems int    `json:"max_context_items"`
		IncludePages    bool   `json:"include_pages"`
		IncludePosts    bool   `json:"include_posts"`
		IncludeUsers    bool   `json:"include_users"`
		IncludeTags     bool   `json:"include_tags"`
		IncludeGroups   bool   `json:"include_groups"`
		IncludeEvents   bool   `json:"include_events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.New(apperr.CodeInvalidParam, "请求参数错误"))
		return
	}

	settings, err := h.service.UpdateSettings(c.Request.Context(), adminID, assistantdomain.Settings{
		Enabled:         req.Enabled,
		PersonaName:     req.PersonaName,
		SystemPrompt:    req.SystemPrompt,
		MaxContextItems: req.MaxContextItems,
		IncludePages:    req.IncludePages,
		IncludePosts:    req.IncludePosts,
		IncludeUsers:    req.IncludeUsers,
		IncludeTags:     req.IncludeTags,
		IncludeGroups:   req.IncludeGroups,
		IncludeEvents:   req.IncludeEvents,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, settings)
}

func latestAssistantUserMessage(messages []usecase.AssistantChatMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return strings.TrimSpace(messages[i].Content)
		}
	}
	return ""
}
