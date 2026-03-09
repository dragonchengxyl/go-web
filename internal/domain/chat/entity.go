package chat

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ConversationType represents the type of conversation
type ConversationType string

const (
	ConversationTypeDirect ConversationType = "direct"
	ConversationTypeGroup  ConversationType = "group"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID        `json:"id"`
	Type      ConversationType `json:"type"`
	Name      *string          `json:"name,omitempty"` // for group chats
	Members   []uuid.UUID      `json:"members"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`

	// Joined fields
	LastMessage     *Message  `json:"last_message,omitempty"`
	UnreadCount     int       `json:"unread_count,omitempty"`
}

// Message represents a chat message
type Message struct {
	ID             uuid.UUID  `json:"id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Content        string     `json:"content"`
	MediaURL       *string    `json:"media_url,omitempty"`
	IsRead         bool       `json:"is_read"`
	CreatedAt      time.Time  `json:"created_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`

	// Joined fields
	SenderUsername  string  `json:"sender_username,omitempty"`
	SenderAvatarKey *string `json:"sender_avatar_key,omitempty"`
}

// ConversationMember represents membership in a conversation
type ConversationMember struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	UserID         uuid.UUID `json:"user_id"`
	JoinedAt       time.Time `json:"joined_at"`
	LastReadAt     time.Time `json:"last_read_at"`
}

var (
	ErrNotFound          = errors.New("conversation not found")
	ErrNotMember         = errors.New("user is not a member of this conversation")
	ErrConversationExists = errors.New("direct conversation already exists")
)
