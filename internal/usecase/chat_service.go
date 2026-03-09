package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/chat"
	"github.com/studio/platform/internal/pkg/apperr"
)

// ChatService handles chat-related business logic
type ChatService struct {
	chatRepo chat.Repository
}

func NewChatService(chatRepo chat.Repository) *ChatService {
	return &ChatService{chatRepo: chatRepo}
}

// CreateDirectConversation creates or returns an existing direct conversation
func (s *ChatService) CreateDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*chat.Conversation, error) {
	if userA == userB {
		return nil, apperr.BadRequest("不能与自己建立会话")
	}

	// Check if conversation already exists
	existing, err := s.chatRepo.GetDirectConversation(ctx, userA, userB)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	now := time.Now()
	c := &chat.Conversation{
		ID:        uuid.New(),
		Type:      chat.ConversationTypeDirect,
		Members:   []uuid.UUID{userA, userB},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.chatRepo.CreateConversation(ctx, c); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "创建会话失败", err)
	}
	return c, nil
}

// SendMessage sends a message to a conversation
type SendMessageInput struct {
	ConversationID uuid.UUID
	SenderID       uuid.UUID
	Content        string
	MediaURL       *string
}

func (s *ChatService) SendMessage(ctx context.Context, input SendMessageInput) (*chat.Message, error) {
	// Verify sender is a member
	isMember, err := s.chatRepo.IsMember(ctx, input.ConversationID, input.SenderID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperr.New(apperr.CodeForbidden, "您不是该会话的成员")
	}

	if input.Content == "" && input.MediaURL == nil {
		return nil, apperr.BadRequest("消息内容不能为空")
	}

	now := time.Now()
	m := &chat.Message{
		ID:             uuid.New(),
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		Content:        input.Content,
		MediaURL:       input.MediaURL,
		CreatedAt:      now,
	}
	if err := s.chatRepo.CreateMessage(ctx, m); err != nil {
		return nil, apperr.Wrap(apperr.CodeInternalError, "发送消息失败", err)
	}
	return m, nil
}

// ListMessages retrieves messages for a conversation
func (s *ChatService) ListMessages(ctx context.Context, conversationID, userID uuid.UUID, page, pageSize int) ([]*chat.Message, int64, error) {
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return nil, 0, err
	}
	if !isMember {
		return nil, 0, apperr.New(apperr.CodeForbidden, "您不是该会话的成员")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	return s.chatRepo.ListMessages(ctx, conversationID, page, pageSize)
}

// ListConversations retrieves conversations for a user
func (s *ChatService) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*chat.Conversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.chatRepo.ListConversations(ctx, userID, page, pageSize)
}

// MarkRead marks all messages in a conversation as read for the user
func (s *ChatService) MarkRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return apperr.New(apperr.CodeForbidden, "您不是该会话的成员")
	}
	return s.chatRepo.MarkRead(ctx, conversationID, userID)
}

// GetConversation retrieves a conversation if the user is a member
func (s *ChatService) GetConversation(ctx context.Context, conversationID, userID uuid.UUID) (*chat.Conversation, error) {
	c, err := s.chatRepo.GetConversationByID(ctx, conversationID)
	if err != nil {
		return nil, apperr.ErrNotFound
	}
	isMember, err := s.chatRepo.IsMember(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperr.New(apperr.CodeForbidden, "您不是该会话的成员")
	}
	return c, nil
}
