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

// Settings controls runtime behaviour of the assistant.
type Settings struct {
	Enabled         bool       `json:"enabled"`
	PersonaName     string     `json:"persona_name"`
	SystemPrompt    string     `json:"system_prompt"`
	MaxContextItems int        `json:"max_context_items"`
	IncludePages    bool       `json:"include_pages"`
	IncludePosts    bool       `json:"include_posts"`
	IncludeUsers    bool       `json:"include_users"`
	IncludeTags     bool       `json:"include_tags"`
	IncludeGroups   bool       `json:"include_groups"`
	IncludeEvents   bool       `json:"include_events"`
	UpdatedAt       time.Time  `json:"updated_at"`
	UpdatedBy       *uuid.UUID `json:"updated_by,omitempty"`
}

var (
	ErrConversationNotFound = errors.New("assistant conversation not found")
	ErrForbidden            = errors.New("assistant conversation access denied")
)
