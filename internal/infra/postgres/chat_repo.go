package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/chat"
)

// ChatRepository implements chat.Repository using PostgreSQL
type ChatRepository struct {
	pool *pgxpool.Pool
}

func NewChatRepository(pool *pgxpool.Pool) *ChatRepository {
	return &ChatRepository{pool: pool}
}

func (r *ChatRepository) CreateConversation(ctx context.Context, c *chat.Conversation) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO conversations (id, type, name, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`,
		c.ID, string(c.Type), c.Name, c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	for _, memberID := range c.Members {
		_, err = tx.Exec(ctx,
			`INSERT INTO conversation_members (conversation_id, user_id, joined_at, last_read_at) VALUES ($1, $2, $3, $4)`,
			c.ID, memberID, c.CreatedAt, c.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to add member: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *ChatRepository) GetConversationByID(ctx context.Context, id uuid.UUID) (*chat.Conversation, error) {
	var c chat.Conversation
	var typeStr string
	err := r.pool.QueryRow(ctx,
		`SELECT id, type, name, created_at, updated_at FROM conversations WHERE id = $1`,
		id,
	).Scan(&c.ID, &typeStr, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, chat.ErrNotFound
		}
		return nil, err
	}
	c.Type = chat.ConversationType(typeStr)

	// Load members
	rows, err := r.pool.Query(ctx, `SELECT user_id FROM conversation_members WHERE conversation_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid uuid.UUID
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		c.Members = append(c.Members, uid)
	}
	return &c, rows.Err()
}

func (r *ChatRepository) GetDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*chat.Conversation, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT c.id FROM conversations c
		JOIN conversation_members m1 ON m1.conversation_id = c.id AND m1.user_id = $1
		JOIN conversation_members m2 ON m2.conversation_id = c.id AND m2.user_id = $2
		WHERE c.type = 'direct'
		LIMIT 1
	`, userA, userB).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.GetConversationByID(ctx, id)
}

func (r *ChatRepository) ListConversations(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*chat.Conversation, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM conversation_members WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.type, c.name, c.created_at, c.updated_at
		FROM conversations c
		JOIN conversation_members cm ON cm.conversation_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
		LIMIT $2 OFFSET $3
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	convs := make([]*chat.Conversation, 0)
	for rows.Next() {
		var c chat.Conversation
		var typeStr string
		if err := rows.Scan(&c.ID, &typeStr, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		c.Type = chat.ConversationType(typeStr)
		convs = append(convs, &c)
	}
	return convs, total, rows.Err()
}

func (r *ChatRepository) IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM conversation_members WHERE conversation_id = $1 AND user_id = $2)`,
		conversationID, userID,
	).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) AddMember(ctx context.Context, member *chat.ConversationMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO conversation_members (conversation_id, user_id, joined_at, last_read_at) VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		member.ConversationID, member.UserID, member.JoinedAt, member.LastReadAt,
	)
	return err
}

func (r *ChatRepository) CreateMessage(ctx context.Context, m *chat.Message) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.Exec(ctx,
		`INSERT INTO messages (id, conversation_id, sender_id, content, media_url, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.ConversationID, m.SenderID, m.Content, m.MediaURL, m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE conversations SET updated_at = $2 WHERE id = $1`,
		m.ConversationID, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *ChatRepository) GetMessageByID(ctx context.Context, id uuid.UUID) (*chat.Message, error) {
	var m chat.Message
	err := r.pool.QueryRow(ctx,
		`SELECT m.id, m.conversation_id, m.sender_id, m.content, m.media_url, m.created_at,
		        u.username, u.avatar_key
		 FROM messages m JOIN users u ON u.id = m.sender_id
		 WHERE m.id = $1 AND m.deleted_at IS NULL`,
		id,
	).Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MediaURL, &m.CreatedAt,
		&m.SenderUsername, &m.SenderAvatarKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, chat.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *ChatRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, page, pageSize int) ([]*chat.Message, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM messages WHERE conversation_id = $1 AND deleted_at IS NULL`,
		conversationID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, `
		SELECT m.id, m.conversation_id, m.sender_id, m.content, m.media_url, m.created_at,
		       u.username, u.avatar_key
		FROM messages m JOIN users u ON u.id = m.sender_id
		WHERE m.conversation_id = $1 AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`, conversationID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	msgs := make([]*chat.Message, 0)
	for rows.Next() {
		var m chat.Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.MediaURL, &m.CreatedAt,
			&m.SenderUsername, &m.SenderAvatarKey); err != nil {
			return nil, 0, err
		}
		msgs = append(msgs, &m)
	}
	return msgs, total, rows.Err()
}

func (r *ChatRepository) MarkRead(ctx context.Context, conversationID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE conversation_members SET last_read_at = NOW() WHERE conversation_id = $1 AND user_id = $2`,
		conversationID, userID,
	)
	return err
}
