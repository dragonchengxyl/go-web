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

	ReplaceKnowledgeDocuments(ctx context.Context, sourceType string, docs []*KnowledgeDocument) error
	SearchKnowledgeDocuments(ctx context.Context, query string, sourceTypes []string, limit int) ([]*KnowledgeDocument, error)
	ListKnowledgeDocumentsForScan(ctx context.Context, sourceTypes []string, limit int) ([]*KnowledgeDocument, error)
	GetKnowledgeOverview(ctx context.Context) (*Overview, error)
	GetMediaAnalysis(ctx context.Context, mediaURL string) (*MediaAnalysis, error)
	UpsertMediaAnalysis(ctx context.Context, item *MediaAnalysis) error

	UpsertFeedback(ctx context.Context, item *Feedback) error
}
