package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/report"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

const createReportSQL = `
	INSERT INTO reports (id, reporter_id, target_type, target_id, reason, description, status, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func (r *ReportRepository) Create(ctx context.Context, rep *report.Report) error {
	_, err := r.pool.Exec(ctx, createReportSQL,
		rep.ID, rep.ReporterID, string(rep.TargetType), rep.TargetID,
		rep.Reason, rep.Description, string(rep.Status), rep.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return report.ErrAlreadyReported
		}
		return fmt.Errorf("failed to create report: %w", err)
	}
	return nil
}
