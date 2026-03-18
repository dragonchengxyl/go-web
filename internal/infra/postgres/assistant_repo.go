package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	assistantdomain "github.com/studio/platform/internal/domain/assistant"
)

// AssistantRepository implements assistant.Repository using PostgreSQL.
type AssistantRepository struct {
	pool *pgxpool.Pool
}

func NewAssistantRepository(pool *pgxpool.Pool) *AssistantRepository {
	return &AssistantRepository{pool: pool}
}

func (r *AssistantRepository) CreateConversation(ctx context.Context, c *assistantdomain.Conversation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO assistant_conversations (id, user_id, title, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, c.ID, c.UserID, c.Title, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *AssistantRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*assistantdomain.Conversation, error) {
	var c assistantdomain.Conversation
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.user_id, c.title, c.created_at, c.updated_at,
		       COALESCE(m.content, ''), COALESCE(m.role, '')
		FROM assistant_conversations c
		LEFT JOIN LATERAL (
			SELECT content, role
			FROM assistant_messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		WHERE c.id = $1
	`, id).Scan(
		&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt,
		&c.LastMessagePreview, &c.LastRole,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, assistantdomain.ErrConversationNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *AssistantRepository) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*assistantdomain.Conversation, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM assistant_conversations WHERE user_id = $1
	`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.user_id, c.title, c.created_at, c.updated_at,
		       COALESCE(m.content, ''), COALESCE(m.role, '')
		FROM assistant_conversations c
		LEFT JOIN LATERAL (
			SELECT content, role
			FROM assistant_messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		WHERE c.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*assistantdomain.Conversation, 0, pageSize)
	for rows.Next() {
		var c assistantdomain.Conversation
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Title, &c.CreatedAt, &c.UpdatedAt,
			&c.LastMessagePreview, &c.LastRole,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, &c)
	}
	return items, total, rows.Err()
}

func (r *AssistantRepository) CreateMessage(ctx context.Context, m *assistantdomain.Message) error {
	cardsJSON, err := json.Marshal(m.Cards)
	if err != nil {
		return fmt.Errorf("marshal assistant cards: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx, `
		INSERT INTO assistant_messages (id, conversation_id, role, content, cards, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, m.ID, m.ConversationID, string(m.Role), m.Content, cardsJSON, m.CreatedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE assistant_conversations
		SET updated_at = $2
		WHERE id = $1
	`, m.ConversationID, m.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *AssistantRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) ([]*assistantdomain.Message, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM assistant_messages WHERE conversation_id = $1
	`, conversationID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, role, content, cards, created_at
		FROM assistant_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, conversationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]*assistantdomain.Message, 0, pageSize)
	for rows.Next() {
		msg, err := scanAssistantMessage(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, msg)
	}
	return items, total, rows.Err()
}

func (r *AssistantRepository) ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*assistantdomain.Message, error) {
	if limit <= 0 {
		limit = 12
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, role, content, cards, created_at
		FROM (
			SELECT id, conversation_id, role, content, cards, created_at
			FROM assistant_messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) t
		ORDER BY created_at ASC
	`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*assistantdomain.Message, 0, limit)
	for rows.Next() {
		msg, err := scanAssistantMessage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, msg)
	}
	return items, rows.Err()
}

func (r *AssistantRepository) GetSettings(ctx context.Context) (*assistantdomain.Settings, error) {
	var settings assistantdomain.Settings
	err := r.pool.QueryRow(ctx, `
		SELECT enabled, persona_name, system_prompt, max_context_items,
		       include_pages, include_posts, include_users, include_tags, include_groups, include_events,
		       updated_at, updated_by
		FROM assistant_settings
		WHERE id = 1
	`).Scan(
		&settings.Enabled,
		&settings.PersonaName,
		&settings.SystemPrompt,
		&settings.MaxContextItems,
		&settings.IncludePages,
		&settings.IncludePosts,
		&settings.IncludeUsers,
		&settings.IncludeTags,
		&settings.IncludeGroups,
		&settings.IncludeEvents,
		&settings.UpdatedAt,
		&settings.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *AssistantRepository) UpsertSettings(ctx context.Context, settings *assistantdomain.Settings) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO assistant_settings (
			id, enabled, persona_name, system_prompt, max_context_items,
			include_pages, include_posts, include_users, include_tags, include_groups, include_events,
			updated_at, updated_by
		)
		VALUES (
			1, $1, $2, $3, $4,
			$5, $6, $7, $8, $9, $10,
			$11, $12
		)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			persona_name = EXCLUDED.persona_name,
			system_prompt = EXCLUDED.system_prompt,
			max_context_items = EXCLUDED.max_context_items,
			include_pages = EXCLUDED.include_pages,
			include_posts = EXCLUDED.include_posts,
			include_users = EXCLUDED.include_users,
			include_tags = EXCLUDED.include_tags,
			include_groups = EXCLUDED.include_groups,
			include_events = EXCLUDED.include_events,
			updated_at = EXCLUDED.updated_at,
			updated_by = EXCLUDED.updated_by
	`,
		settings.Enabled,
		settings.PersonaName,
		settings.SystemPrompt,
		settings.MaxContextItems,
		settings.IncludePages,
		settings.IncludePosts,
		settings.IncludeUsers,
		settings.IncludeTags,
		settings.IncludeGroups,
		settings.IncludeEvents,
		settings.UpdatedAt,
		settings.UpdatedBy,
	)
	return err
}

func (r *AssistantRepository) ReplaceKnowledgeDocuments(ctx context.Context, sourceType string, docs []*assistantdomain.KnowledgeDocument) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, `
		DELETE FROM assistant_knowledge_documents
		WHERE source_type = $1
	`, sourceType); err != nil {
		return err
	}

	for _, doc := range docs {
		if doc == nil {
			continue
		}
		tagsJSON, err := json.Marshal(doc.Tags)
		if err != nil {
			return fmt.Errorf("marshal knowledge tags: %w", err)
		}
		embeddingJSON, err := json.Marshal(doc.Embedding)
		if err != nil {
			return fmt.Errorf("marshal knowledge embedding: %w", err)
		}

		indexedAt := doc.IndexedAt
		if indexedAt.IsZero() {
			indexedAt = time.Now()
		}
		sourceUpdatedAt := doc.SourceUpdatedAt
		if sourceUpdatedAt.IsZero() {
			sourceUpdatedAt = indexedAt
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO assistant_knowledge_documents (
				id, source_type, source_key, chunk_index, title, summary, content, href, meta,
				source_label, tags, search_text, search_vector, embedding, indexed_at, source_updated_at, is_active
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9,
				$10, $11, $12, to_tsvector('simple', $12), $13, $14, $15, true
			)
		`,
			doc.ID,
			doc.SourceType,
			doc.SourceKey,
			doc.ChunkIndex,
			doc.Title,
			doc.Summary,
			doc.Content,
			doc.Href,
			doc.Meta,
			doc.SourceLabel,
			tagsJSON,
			doc.SearchText,
			embeddingJSON,
			indexedAt,
			sourceUpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *AssistantRepository) SearchKnowledgeDocuments(ctx context.Context, query string, sourceTypes []string, limit int) ([]*assistantdomain.KnowledgeDocument, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(sourceTypes) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 12
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, source_type, source_key, chunk_index, title, summary, content, href, meta,
		       source_label, tags, embedding, indexed_at, source_updated_at,
		       (
		           CASE WHEN title ILIKE '%' || $1 || '%' THEN 2.4 ELSE 0 END +
		           CASE WHEN summary ILIKE '%' || $1 || '%' THEN 1.6 ELSE 0 END +
		           CASE WHEN content ILIKE '%' || $1 || '%' THEN 1.2 ELSE 0 END +
		           CASE WHEN search_text ILIKE '%' || $1 || '%' THEN 0.8 ELSE 0 END +
		           CASE
		               WHEN search_vector @@ websearch_to_tsquery('simple', $1)
		               THEN ts_rank_cd(search_vector, websearch_to_tsquery('simple', $1))
		               ELSE 0
		           END
		       ) AS keyword_score
		FROM assistant_knowledge_documents
		WHERE is_active = true
		  AND source_type = ANY($2)
		  AND (
		      title ILIKE '%' || $1 || '%'
		      OR summary ILIKE '%' || $1 || '%'
		      OR content ILIKE '%' || $1 || '%'
		      OR search_text ILIKE '%' || $1 || '%'
		      OR search_vector @@ websearch_to_tsquery('simple', $1)
		  )
		ORDER BY keyword_score DESC, source_updated_at DESC
		LIMIT $3
	`, query, sourceTypes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*assistantdomain.KnowledgeDocument, 0, limit)
	for rows.Next() {
		item, err := scanKnowledgeDocument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssistantRepository) ListKnowledgeDocumentsForScan(ctx context.Context, sourceTypes []string, limit int) ([]*assistantdomain.KnowledgeDocument, error) {
	if len(sourceTypes) == 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 2000
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, source_type, source_key, chunk_index, title, summary, content, href, meta,
		       source_label, tags, embedding, indexed_at, source_updated_at, 0::float8 AS keyword_score
		FROM assistant_knowledge_documents
		WHERE is_active = true
		  AND source_type = ANY($1)
		ORDER BY source_updated_at DESC
		LIMIT $2
	`, sourceTypes, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*assistantdomain.KnowledgeDocument, 0, limit)
	for rows.Next() {
		item, err := scanKnowledgeDocument(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *AssistantRepository) GetKnowledgeOverview(ctx context.Context) (*assistantdomain.Overview, error) {
	overview := &assistantdomain.Overview{
		DocumentsBySource: make(map[string]int64),
	}

	var lastIndexedAt *time.Time
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), MAX(indexed_at)
		FROM assistant_knowledge_documents
		WHERE is_active = true
	`).Scan(&overview.IndexedDocuments, &lastIndexedAt); err != nil {
		return nil, err
	}
	overview.LastIndexedAt = lastIndexedAt

	rows, err := r.pool.Query(ctx, `
		SELECT source_type, COUNT(*)
		FROM assistant_knowledge_documents
		WHERE is_active = true
		GROUP BY source_type
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sourceType string
		var count int64
		if err := rows.Scan(&sourceType, &count); err != nil {
			return nil, err
		}
		overview.DocumentsBySource[sourceType] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE value = 'helpful'),
			COUNT(*) FILTER (WHERE value = 'unhelpful')
		FROM assistant_feedback
	`).Scan(&overview.FeedbackHelpful, &overview.FeedbackUnhelpful); err != nil {
		return nil, err
	}

	return overview, nil
}

func (r *AssistantRepository) UpsertFeedback(ctx context.Context, item *assistantdomain.Feedback) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	sourceCountsJSON, err := json.Marshal(item.SourceCounts)
	if err != nil {
		return fmt.Errorf("marshal assistant feedback source_counts: %w", err)
	}
	cardsJSON, err := json.Marshal(item.Cards)
	if err != nil {
		return fmt.Errorf("marshal assistant feedback cards: %w", err)
	}

	createdAt := item.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO assistant_feedback (
			id, response_id, conversation_id, user_id, value, query_text, reply_excerpt,
			provider, intent, fallback, page_path, source_counts, cards, created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)
		ON CONFLICT (response_id) DO UPDATE SET
			conversation_id = EXCLUDED.conversation_id,
			user_id = EXCLUDED.user_id,
			value = EXCLUDED.value,
			query_text = EXCLUDED.query_text,
			reply_excerpt = EXCLUDED.reply_excerpt,
			provider = EXCLUDED.provider,
			intent = EXCLUDED.intent,
			fallback = EXCLUDED.fallback,
			page_path = EXCLUDED.page_path,
			source_counts = EXCLUDED.source_counts,
			cards = EXCLUDED.cards,
			created_at = EXCLUDED.created_at
	`,
		item.ID,
		item.ResponseID,
		item.ConversationID,
		item.UserID,
		string(item.Value),
		item.QueryText,
		item.ReplyExcerpt,
		item.Provider,
		item.Intent,
		item.Fallback,
		item.PagePath,
		sourceCountsJSON,
		cardsJSON,
		createdAt,
	)
	return err
}

func scanAssistantMessage(row pgx.Row) (*assistantdomain.Message, error) {
	var msg assistantdomain.Message
	var role string
	var cardsJSON []byte
	if err := row.Scan(&msg.ID, &msg.ConversationID, &role, &msg.Content, &cardsJSON, &msg.CreatedAt); err != nil {
		return nil, err
	}
	msg.Role = assistantdomain.MessageRole(role)
	if len(cardsJSON) > 0 {
		_ = json.Unmarshal(cardsJSON, &msg.Cards)
	}
	return &msg, nil
}

func scanKnowledgeDocument(row pgx.Row) (*assistantdomain.KnowledgeDocument, error) {
	var item assistantdomain.KnowledgeDocument
	var tagsJSON []byte
	var embeddingJSON []byte
	if err := row.Scan(
		&item.ID,
		&item.SourceType,
		&item.SourceKey,
		&item.ChunkIndex,
		&item.Title,
		&item.Summary,
		&item.Content,
		&item.Href,
		&item.Meta,
		&item.SourceLabel,
		&tagsJSON,
		&embeddingJSON,
		&item.IndexedAt,
		&item.SourceUpdatedAt,
		&item.KeywordScore,
	); err != nil {
		return nil, err
	}
	if len(tagsJSON) > 0 {
		_ = json.Unmarshal(tagsJSON, &item.Tags)
	}
	if len(embeddingJSON) > 0 {
		_ = json.Unmarshal(embeddingJSON, &item.Embedding)
	}
	return &item, nil
}
