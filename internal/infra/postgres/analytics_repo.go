package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/analytics"
)

type analyticsRepo struct {
	db *pgxpool.Pool
}

// NewAnalyticsRepository creates a new analytics repository
func NewAnalyticsRepository(db *pgxpool.Pool) analytics.Repository {
	return &analyticsRepo{db: db}
}

const trackEventSQL = `
	INSERT INTO analytics_events (id, event_type, user_id, session_id, properties, ip_address, user_agent, referrer, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

func (r *analyticsRepo) Track(event *analytics.Event) error {
	propertiesJSON, err := json.Marshal(event.Properties)
	if err != nil {
		return fmt.Errorf("failed to marshal properties: %w", err)
	}

	_, err = r.db.Exec(context.Background(), trackEventSQL,
		event.ID, event.EventType, event.UserID, event.SessionID,
		propertiesJSON, event.IPAddress, event.UserAgent, event.Referrer, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to track event: %w", err)
	}

	return nil
}

func (r *analyticsRepo) GetEvents(filter analytics.EventFilter) ([]*analytics.Event, error) {
	query := `
		SELECT id, event_type, user_id, session_id, properties, ip_address, user_agent, referrer, created_at
		FROM analytics_events WHERE 1=1
	`
	args := []any{}
	argIndex := 1

	if filter.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, *filter.EventType)
		argIndex++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	events := make([]*analytics.Event, 0)
	for rows.Next() {
		var event analytics.Event
		var propertiesJSON []byte

		err := rows.Scan(
			&event.ID, &event.EventType, &event.UserID, &event.SessionID,
			&propertiesJSON, &event.IPAddress, &event.UserAgent, &event.Referrer, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if len(propertiesJSON) > 0 {
			if err := json.Unmarshal(propertiesJSON, &event.Properties); err != nil {
				return nil, fmt.Errorf("failed to unmarshal properties: %w", err)
			}
		}

		events = append(events, &event)
	}

	return events, nil
}

func (r *analyticsRepo) GetEventCount(filter analytics.EventFilter) (int64, error) {
	query := `SELECT COUNT(*) FROM analytics_events WHERE 1=1`
	args := []any{}
	argIndex := 1

	if filter.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argIndex)
		args = append(args, *filter.EventType)
		argIndex++
	}

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndTime)
	}

	var count int64
	err := r.db.QueryRow(context.Background(), query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}

	return count, nil
}
