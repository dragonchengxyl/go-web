package usecase

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StatsService provides analytics and dashboard data
type StatsService struct {
	pool *pgxpool.Pool
}

// NewStatsService creates a new StatsService
func NewStatsService(pool *pgxpool.Pool) *StatsService {
	return &StatsService{pool: pool}
}

// DashboardStats holds the main dashboard metrics
type DashboardStats struct {
	TotalUsers    int64 `json:"total_users"`
	NewUsersToday int64 `json:"new_users_today"`
	TotalPosts    int64 `json:"total_posts"`
	TotalReports  int64 `json:"total_reports"` // pending status only
}

// ChartPoint represents a single data point on a chart
type ChartPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// GetDashboardStats returns the main dashboard metrics
func (s *StatsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	today := time.Now().Truncate(24 * time.Hour)

	queries := []struct {
		query string
		args  []any
		dest  *int64
	}{
		{`SELECT COUNT(*) FROM users`, nil, &stats.TotalUsers},
		{`SELECT COUNT(*) FROM users WHERE created_at >= $1`, []any{today}, &stats.NewUsersToday},
		{`SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL`, nil, &stats.TotalPosts},
		{`SELECT COUNT(*) FROM reports WHERE status = 'pending'`, nil, &stats.TotalReports},
	}

	for _, q := range queries {
		if q.dest != nil {
			if err := s.pool.QueryRow(ctx, q.query, q.args...).Scan(q.dest); err != nil {
				*q.dest = 0
			}
		}
	}

	return stats, nil
}

// GetUserGrowthChart returns daily new user registrations for the past N days
func (s *StatsService) GetUserGrowthChart(ctx context.Context, days int) ([]ChartPoint, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			DATE(created_at) AS date,
			COUNT(*) AS count
		FROM users
		WHERE created_at >= NOW() - ($1 || ' days')::INTERVAL
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, days)
	if err != nil {
		return []ChartPoint{}, nil
	}
	defer rows.Close()

	points := make([]ChartPoint, 0)
	for rows.Next() {
		var p ChartPoint
		var date time.Time
		if err := rows.Scan(&date, &p.Value); err != nil {
			continue
		}
		p.Date = date.Format("2006-01-02")
		points = append(points, p)
	}
	return points, nil
}
