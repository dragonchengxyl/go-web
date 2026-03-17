package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/sponsor"
)

type SponsorRepository struct {
	pool *pgxpool.Pool
}

func NewSponsorRepository(pool *pgxpool.Pool) sponsor.Repository {
	return &SponsorRepository{pool: pool}
}

func (r *SponsorRepository) Get(ctx context.Context) (*configs.SponsorConfig, error) {
	var cfg configs.SponsorConfig
	err := r.pool.QueryRow(ctx, `
		SELECT monthly_goal, current_raised, alipay_qr_url, wechat_qr_url, message
		FROM sponsor_settings
		WHERE id = 1
	`).Scan(
		&cfg.MonthlyGoal,
		&cfg.CurrentRaised,
		&cfg.AlipayQRURL,
		&cfg.WechatQRURL,
		&cfg.Message,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *SponsorRepository) Upsert(ctx context.Context, cfg configs.SponsorConfig, updatedBy *uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sponsor_settings (
			id, monthly_goal, current_raised, alipay_qr_url, wechat_qr_url, message, updated_at, updated_by
		)
		VALUES (1, $1, $2, $3, $4, $5, NOW(), $6)
		ON CONFLICT (id) DO UPDATE SET
			monthly_goal = EXCLUDED.monthly_goal,
			current_raised = EXCLUDED.current_raised,
			alipay_qr_url = EXCLUDED.alipay_qr_url,
			wechat_qr_url = EXCLUDED.wechat_qr_url,
			message = EXCLUDED.message,
			updated_at = NOW(),
			updated_by = EXCLUDED.updated_by
	`, cfg.MonthlyGoal, cfg.CurrentRaised, cfg.AlipayQRURL, cfg.WechatQRURL, cfg.Message, updatedBy)
	return err
}
