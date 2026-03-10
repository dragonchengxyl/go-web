package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
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

func (r *ReportRepository) List(ctx context.Context, status string, page, size int) ([]*report.Report, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	args := []any{}
	where := ""
	if status != "" {
		args = append(args, status)
		where = fmt.Sprintf("WHERE rp.status = $%d", len(args))
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM reports rp %s`, where)
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	args = append(args, size, offset)
	dataQuery := fmt.Sprintf(`
		SELECT rp.id, rp.reporter_id, COALESCE(u.username, ''), rp.target_type, rp.target_id,
		       rp.reason, COALESCE(rp.description, ''), rp.status,
		       rp.reviewed_by, rp.reviewed_at, rp.created_at
		FROM reports rp
		LEFT JOIN users u ON u.id = rp.reporter_id
		%s
		ORDER BY rp.created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, len(args)-1, len(args))

	rows, err := r.pool.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}
	defer rows.Close()

	reports := make([]*report.Report, 0)
	for rows.Next() {
		rep := &report.Report{}
		if err := rows.Scan(
			&rep.ID, &rep.ReporterID, &rep.ReporterUsername,
			&rep.TargetType, &rep.TargetID,
			&rep.Reason, &rep.Description, &rep.Status,
			&rep.ReviewedBy, &rep.ReviewedAt, &rep.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, rep)
	}
	return reports, total, nil
}

func (r *ReportRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status report.Status, reviewedBy uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE reports SET status = $1, reviewed_by = $2, reviewed_at = NOW() WHERE id = $3`,
		string(status), reviewedBy, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update report status: %w", err)
	}
	return nil
}

