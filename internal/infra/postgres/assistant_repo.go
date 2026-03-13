package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
