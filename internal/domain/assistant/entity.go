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
	Ref     string `json:"ref,omitempty"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
	Href    string `json:"href"`
	Meta    string `json:"meta,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Source  string `json:"source,omitempty"`
}

// Insight is a structured copilot output for page-specific assistance.
type Insight struct {
	Kind    string   `json:"kind"`
	Title   string   `json:"title"`
	Summary string   `json:"summary,omitempty"`
	Bullets []string `json:"bullets,omitempty"`
}

// MediaAnalysis stores image understanding results for UI and moderation.
type MediaAnalysis struct {
	ID                uuid.UUID `json:"id"`
	MediaURL          string    `json:"media_url"`
	AltText           string    `json:"alt_text"`
	Tags              []string  `json:"tags,omitempty"`
	ImageSummary      string    `json:"image_summary,omitempty"`
	ModerationSummary string    `json:"moderation_summary,omitempty"`
	RiskLevel         string    `json:"risk_level,omitempty"`
	SafetyNotes       []string  `json:"safety_notes,omitempty"`
	Provider          string    `json:"provider,omitempty"`
	Model             string    `json:"model,omitempty"`
	Fallback          bool      `json:"fallback"`
	CachedAt          time.Time `json:"cached_at"`
	ExpiresAt         time.Time `json:"expires_at"`
}

const (
	KnowledgeSourcePage  = "page"
	KnowledgeSourcePost  = "post"
	KnowledgeSourceGroup = "group"
	KnowledgeSourceEvent = "event"
)

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
	Insights       []Insight   `json:"insights,omitempty"`
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

// KnowledgeDocument stores an indexed assistant retrieval chunk.
type KnowledgeDocument struct {
	ID              uuid.UUID `json:"id"`
	SourceType      string    `json:"source_type"`
	SourceKey       string    `json:"source_key"`
	ChunkIndex      int       `json:"chunk_index"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	Content         string    `json:"content"`
	Href            string    `json:"href"`
	Meta            string    `json:"meta,omitempty"`
	SourceLabel     string    `json:"source_label,omitempty"`
	Tags            []string  `json:"tags,omitempty"`
	SearchText      string    `json:"search_text,omitempty"`
	Embedding       []float64 `json:"embedding,omitempty"`
	IndexedAt       time.Time `json:"indexed_at"`
	SourceUpdatedAt time.Time `json:"source_updated_at"`
	KeywordScore    float64   `json:"keyword_score,omitempty"`
	VectorScore     float64   `json:"vector_score,omitempty"`
}

type FeedbackValue string

const (
	FeedbackHelpful   FeedbackValue = "helpful"
	FeedbackUnhelpful FeedbackValue = "unhelpful"
)

// Feedback stores a user's judgement for a single assistant response.
type Feedback struct {
	ID             uuid.UUID      `json:"id"`
	ResponseID     uuid.UUID      `json:"response_id"`
	ConversationID *uuid.UUID     `json:"conversation_id,omitempty"`
	UserID         *uuid.UUID     `json:"user_id,omitempty"`
	Value          FeedbackValue  `json:"value"`
	QueryText      string         `json:"query_text"`
	ReplyExcerpt   string         `json:"reply_excerpt"`
	Provider       string         `json:"provider,omitempty"`
	Intent         string         `json:"intent,omitempty"`
	Fallback       bool           `json:"fallback"`
	PagePath       string         `json:"page_path,omitempty"`
	SourceCounts   map[string]int `json:"source_counts,omitempty"`
	Cards          []Card         `json:"cards,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// Overview summarises assistant retrieval/index/feedback status for admin use.
type Overview struct {
	IndexedDocuments    int64            `json:"indexed_documents"`
	DocumentsBySource   map[string]int64 `json:"documents_by_source"`
	LastIndexedAt       *time.Time       `json:"last_indexed_at,omitempty"`
	MediaCacheEntries   int64            `json:"media_cache_entries"`
	FeedbackHelpful     int64            `json:"feedback_helpful"`
	FeedbackUnhelpful   int64            `json:"feedback_unhelpful"`
	EmbeddingConfigured bool             `json:"embedding_configured"`
	EmbeddingModel      string           `json:"embedding_model,omitempty"`
	VisionConfigured    bool             `json:"vision_configured"`
	VisionModel         string           `json:"vision_model,omitempty"`
	RetrievalLimit      int              `json:"retrieval_limit"`
	VectorScanLimit     int              `json:"vector_scan_limit"`
	SyncIntervalSec     int              `json:"sync_interval_sec"`
}

var (
	ErrConversationNotFound = errors.New("assistant conversation not found")
	ErrForbidden            = errors.New("assistant conversation access denied")
)
