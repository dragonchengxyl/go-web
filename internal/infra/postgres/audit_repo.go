package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/audit"
)

type auditRepo struct {
	db *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *pgxpool.Pool) audit.Repository {
	return &auditRepo{db: db}
}

const createAuditLogSQL = `
	INSERT INTO audit_logs (id, user_id, username, action, resource, resource_id, ip_address, user_agent, before_data, after_data, error_message, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`

func (r *auditRepo) Create(ctx context.Context, log *audit.Log) error {
	_, err := r.db.Exec(ctx, createAuditLogSQL,
		log.ID, log.UserID, log.Username, log.Action, log.Resource,
		log.ResourceID, log.IPAddress, log.UserAgent, log.BeforeData,
		log.AfterData, log.ErrorMessage, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}
	return nil
}

const getAuditLogByIDSQL = `
	SELECT id, user_id, username, action, resource, resource_id, ip_address, user_agent, before_data, after_data, error_message, created_at
	FROM audit_logs WHERE id = $1
`

func (r *auditRepo) GetByID(ctx context.Context, id uuid.UUID) (*audit.Log, error) {
	var log audit.Log
	err := r.db.QueryRow(ctx, getAuditLogByIDSQL, id).Scan(
		&log.ID, &log.UserID, &log.Username, &log.Action, &log.Resource,
		&log.ResourceID, &log.IPAddress, &log.UserAgent, &log.BeforeData,
		&log.AfterData, &log.ErrorMessage, &log.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}
	return &log, nil
}

const getAuditLogsByUserIDSQL = `
	SELECT id, user_id, username, action, resource, resource_id, ip_address, user_agent, before_data, after_data, error_message, created_at
	FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2
`

func (r *auditRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*audit.Log, error) {
	rows, err := r.db.Query(ctx, getAuditLogsByUserIDSQL, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user: %w", err)
	}
	defer rows.Close()

	logs := make([]*audit.Log, 0)
	for rows.Next() {
		var log audit.Log
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Action, &log.Resource,
			&log.ResourceID, &log.IPAddress, &log.UserAgent, &log.BeforeData,
			&log.AfterData, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

const getAuditLogsByResourceSQL = `
	SELECT id, user_id, username, action, resource, resource_id, ip_address, user_agent, before_data, after_data, error_message, created_at
	FROM audit_logs WHERE resource = $1 AND resource_id = $2 ORDER BY created_at DESC LIMIT $3
`

func (r *auditRepo) GetByResource(ctx context.Context, resource audit.Resource, resourceID uuid.UUID, limit int) ([]*audit.Log, error) {
	rows, err := r.db.Query(ctx, getAuditLogsByResourceSQL, resource, resourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by resource: %w", err)
	}
	defer rows.Close()

	logs := make([]*audit.Log, 0)
	for rows.Next() {
		var log audit.Log
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Action, &log.Resource,
			&log.ResourceID, &log.IPAddress, &log.UserAgent, &log.BeforeData,
			&log.AfterData, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

func (r *auditRepo) List(ctx context.Context, filter audit.ListFilter) ([]*audit.Log, int64, error) {
	query := `
		SELECT id, user_id, username, action, resource, resource_id, ip_address, user_agent, before_data, after_data, error_message, created_at
		FROM audit_logs WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []any{}
	argIndex := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, *filter.Action)
		argIndex++
	}

	if filter.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND resource = $%d", argIndex)
		args = append(args, *filter.Resource)
		argIndex++
	}

	if filter.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *filter.ResourceID)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*audit.Log, 0)
	for rows.Next() {
		var log audit.Log
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Username, &log.Action, &log.Resource,
			&log.ResourceID, &log.IPAddress, &log.UserAgent, &log.BeforeData,
			&log.AfterData, &log.ErrorMessage, &log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, total, nil
}