package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/achievement"
)

// AchievementRepository implements achievement.Repository
type AchievementRepository struct {
	pool *pgxpool.Pool
}

// NewAchievementRepository creates a new AchievementRepository
func NewAchievementRepository(pool *pgxpool.Pool) *AchievementRepository {
	return &AchievementRepository{pool: pool}
}

// ListAchievements returns all non-secret achievements
func (r *AchievementRepository) ListAchievements(ctx context.Context) ([]*achievement.Achievement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, description, icon_key, rarity, points,
		       condition_type, condition_value, is_secret, created_at
		FROM achievements
		WHERE is_secret = FALSE
		ORDER BY rarity DESC, points DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	defer rows.Close()

	var list []*achievement.Achievement
	for rows.Next() {
		a := &achievement.Achievement{}
		if err := rows.Scan(
			&a.ID, &a.Slug, &a.Name, &a.Description, &a.IconKey,
			&a.Rarity, &a.Points, &a.ConditionType, &a.ConditionValue, &a.IsSecret, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		list = append(list, a)
	}
	return list, nil
}

// GetAchievementByID retrieves an achievement by id
func (r *AchievementRepository) GetAchievementByID(ctx context.Context, id int) (*achievement.Achievement, error) {
	a := &achievement.Achievement{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, icon_key, rarity, points,
		       condition_type, condition_value, is_secret, created_at
		FROM achievements WHERE id = $1
	`, id).Scan(
		&a.ID, &a.Slug, &a.Name, &a.Description, &a.IconKey,
		&a.Rarity, &a.Points, &a.ConditionType, &a.ConditionValue, &a.IsSecret, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, achievement.ErrNotFound
		}
		return nil, fmt.Errorf("get achievement: %w", err)
	}
	return a, nil
}

// GetAchievementBySlug retrieves an achievement by slug
func (r *AchievementRepository) GetAchievementBySlug(ctx context.Context, slug string) (*achievement.Achievement, error) {
	a := &achievement.Achievement{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, icon_key, rarity, points,
		       condition_type, condition_value, is_secret, created_at
		FROM achievements WHERE slug = $1
	`, slug).Scan(
		&a.ID, &a.Slug, &a.Name, &a.Description, &a.IconKey,
		&a.Rarity, &a.Points, &a.ConditionType, &a.ConditionValue, &a.IsSecret, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, achievement.ErrNotFound
		}
		return nil, fmt.Errorf("get achievement by slug: %w", err)
	}
	return a, nil
}

// UnlockAchievement grants an achievement to a user
func (r *AchievementRepository) UnlockAchievement(ctx context.Context, ua *achievement.UserAchievement) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_achievements (user_id, achievement_id, obtained_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, achievement_id) DO NOTHING
	`, ua.UserID, ua.AchievementID, ua.ObtainedAt)
	if err != nil {
		return fmt.Errorf("unlock achievement: %w", err)
	}
	return nil
}

// ListUserAchievements returns all achievements unlocked by a user
func (r *AchievementRepository) ListUserAchievements(ctx context.Context, userID uuid.UUID) ([]*achievement.UserAchievement, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ua.id, ua.user_id, ua.achievement_id, ua.obtained_at,
		       a.slug, a.name, a.description, a.icon_key, a.rarity, a.points, a.is_secret
		FROM user_achievements ua
		JOIN achievements a ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
		ORDER BY ua.obtained_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	defer rows.Close()

	var list []*achievement.UserAchievement
	for rows.Next() {
		ua := &achievement.UserAchievement{Achievement: &achievement.Achievement{}}
		if err := rows.Scan(
			&ua.ID, &ua.UserID, &ua.AchievementID, &ua.ObtainedAt,
			&ua.Achievement.Slug, &ua.Achievement.Name, &ua.Achievement.Description,
			&ua.Achievement.IconKey, &ua.Achievement.Rarity, &ua.Achievement.Points, &ua.Achievement.IsSecret,
		); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		ua.Achievement.ID = ua.AchievementID
		list = append(list, ua)
	}
	return list, nil
}

// HasAchievement checks if a user has unlocked an achievement
func (r *AchievementRepository) HasAchievement(ctx context.Context, userID uuid.UUID, achievementID int) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_achievements WHERE user_id=$1 AND achievement_id=$2)
	`, userID, achievementID).Scan(&exists)
	return exists, err
}

// GetUserPoints returns the point balance for a user
func (r *AchievementRepository) GetUserPoints(ctx context.Context, userID uuid.UUID) (*achievement.UserPoints, error) {
	up := &achievement.UserPoints{UserID: userID}
	err := r.pool.QueryRow(ctx, `
		SELECT balance, total_earned, updated_at FROM user_points WHERE user_id = $1
	`, userID).Scan(&up.Balance, &up.TotalEarned, &up.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return zero balance if no row yet
			up.UpdatedAt = time.Now()
			return up, nil
		}
		return nil, fmt.Errorf("get user points: %w", err)
	}
	return up, nil
}

// AddPoints records a point transaction and updates the balance
func (r *AchievementRepository) AddPoints(ctx context.Context, tx *achievement.PointTransaction) error {
	// Insert transaction
	_, err := r.pool.Exec(ctx, `
		INSERT INTO point_transactions (user_id, amount, source, ref_id, note, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tx.UserID, tx.Amount, tx.Source, tx.RefID, tx.Note, tx.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert point transaction: %w", err)
	}

	// Upsert balance
	earned := 0
	if tx.Amount > 0 {
		earned = tx.Amount
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO user_points (user_id, balance, total_earned, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
		    balance      = user_points.balance + EXCLUDED.balance,
		    total_earned = user_points.total_earned + $3,
		    updated_at   = NOW()
	`, tx.UserID, tx.Amount, earned)
	if err != nil {
		return fmt.Errorf("update user points: %w", err)
	}
	return nil
}

// ListPointTransactions returns recent point transactions for a user
func (r *AchievementRepository) ListPointTransactions(ctx context.Context, userID uuid.UUID, limit int) ([]*achievement.PointTransaction, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, amount, source, ref_id, note, created_at
		FROM point_transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list point transactions: %w", err)
	}
	defer rows.Close()

	var list []*achievement.PointTransaction
	for rows.Next() {
		t := &achievement.PointTransaction{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Amount, &t.Source, &t.RefID, &t.Note, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan point transaction: %w", err)
		}
		list = append(list, t)
	}
	return list, nil
}
