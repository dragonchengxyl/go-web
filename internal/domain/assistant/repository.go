package assistant

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence operations for assistant conversations.
type Repository interface {
	CreateConversation(ctx context.Context, c *Conversation) error
	GetConversationByID(ctx context.Context, id uuid.UUID) (*Conversation, error)
	ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*Conversation, int64, error)

	CreateMessage(ctx context.Context, m *Message) error
	ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) ([]*Message, int64, error)
	ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*Message, error)

	GetSettings(ctx context.Context) (*Settings, error)
	UpsertSettings(ctx context.Context, settings *Settings) error
}
