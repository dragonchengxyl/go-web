package assistant

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// MessageRole identifies who sent an assistant conversation message.
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// Card is a structured recommendation attached to an assistant reply.
type Card struct {
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Href    string `json:"href"`
	Meta    string `json:"meta,omitempty"`
}

// Conversation stores a user's AI assistant thread.
type Conversation struct {
	ID                 uuid.UUID `json:"id"`
	UserID             uuid.UUID `json:"user_id"`
	Title              string    `json:"title"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	LastMessagePreview string    `json:"last_message_preview,omitempty"`
	LastRole           string    `json:"last_role,omitempty"`
}

// Message is a single assistant conversation message.
type Message struct {
	ID             uuid.UUID   `json:"id"`
	ConversationID uuid.UUID   `json:"conversation_id"`
	Role           MessageRole `json:"role"`
	Content        string      `json:"content"`
	Cards          []Card      `json:"cards,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

var (
	ErrConversationNotFound = errors.New("assistant conversation not found")
	ErrForbidden            = errors.New("assistant conversation access denied")
)
