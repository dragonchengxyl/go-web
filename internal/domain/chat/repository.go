package chat

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for chat data access
type Repository interface {
	// Conversations
	CreateConversation(ctx context.Context, c *Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	GetDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*Conversation, error)
	ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Conversation, int64, error)
	IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	AddMember(ctx context.Context, member *ConversationMember) error

	// Messages
	CreateMessage(ctx context.Context, m *Message) error
	GetMessageByID(ctx context.Context, id uuid.UUID) (*Message, error)
	ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) ([]*Message, int64, error)
	MarkRead(ctx context.Context, conversationID, userID uuid.UUID) error
}
