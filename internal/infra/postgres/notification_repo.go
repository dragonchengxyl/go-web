package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/notification"
)

// NotificationRepository implements notification.Repository using PostgreSQL
type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

const createNotificationSQL = `
	INSERT INTO notifications (id, user_id, actor_id, type, target_id, target_type, is_read, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func (r *NotificationRepository) Create(ctx context.Context, n *notification.Notification) error {
	_, err := r.pool.Exec(ctx, createNotificationSQL,
		n.ID, n.UserID, n.ActorID, string(n.Type), n.TargetID, n.TargetType, n.IsRead, n.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

const listNotificationsSQL = `
	SELECT n.id, n.user_id, n.actor_id, n.type, n.target_id, n.target_type, n.is_read, n.created_at,
	       u.username, u.avatar_key
	FROM notifications n
	LEFT JOIN users u ON u.id = n.actor_id
	WHERE n.user_id = $1
	ORDER BY n.created_at DESC
	LIMIT $2 OFFSET $3
`

const countNotificationsSQL = `SELECT COUNT(*) FROM notifications WHERE user_id = $1`

func (r *NotificationRepository) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*notification.Notification, int64, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, countNotificationsSQL, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	offset := (page - 1) * pageSize
	rows, err := r.pool.Query(ctx, listNotificationsSQL, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*notification.Notification, 0)
	for rows.Next() {
		var n notification.Notification
		var typeStr string
		err := rows.Scan(
			&n.ID, &n.UserID, &n.ActorID, &typeStr, &n.TargetID, &n.TargetType, &n.IsRead, &n.CreatedAt,
			&n.ActorUsername, &n.ActorAvatarKey,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan notification: %w", err)
		}
		n.Type = notification.Type(typeStr)
		notifications = append(notifications, &n)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}
	return notifications, total, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND id = ANY($2)`,
		userID, ids,
	)
	return err
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE notifications SET is_read = TRUE WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`,
		userID,
	).Scan(&count)
	return count, err
}
