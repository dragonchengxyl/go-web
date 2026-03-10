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
	TotalUsers     int64   `json:"total_users"`
	NewUsersToday  int64   `json:"new_users_today"`
	TotalDownloads int64   `json:"total_downloads"`
	DownloadsToday int64   `json:"downloads_today"`
	TotalRevenue   float64 `json:"total_revenue"`
	RevenueToday   float64 `json:"revenue_today"`
	TotalGames     int64   `json:"total_games"`
	TotalOrders    int64   `json:"total_orders"`
	TotalPosts     int64   `json:"total_posts"`
	TotalReports   int64   `json:"total_reports"` // pending status only
}

// ChartPoint represents a single data point on a chart
type ChartPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

// PopularGame holds popular game stats
type PopularGame struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Downloads int64  `json:"downloads"`
}

// GetDashboardStats returns the main dashboard metrics
func (s *StatsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	today := time.Now().Truncate(24 * time.Hour)

	queries := []struct {
		query  string
		args   []any
		dest   *int64
		destF  *float64
	}{
		{`SELECT COUNT(*) FROM users`, nil, &stats.TotalUsers, nil},
		{`SELECT COUNT(*) FROM users WHERE created_at >= $1`, []any{today}, &stats.NewUsersToday, nil},
		{`SELECT COUNT(*) FROM download_logs`, nil, &stats.TotalDownloads, nil},
		{`SELECT COUNT(*) FROM download_logs WHERE downloaded_at >= $1`, []any{today}, &stats.DownloadsToday, nil},
		{`SELECT COUNT(*) FROM games`, nil, &stats.TotalGames, nil},
		{`SELECT COUNT(*) FROM orders`, nil, &stats.TotalOrders, nil},
		{`SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL`, nil, &stats.TotalPosts, nil},
		{`SELECT COUNT(*) FROM reports WHERE status = 'pending'`, nil, &stats.TotalReports, nil},
	}

	for _, q := range queries {
		if q.dest != nil {
			if err := s.pool.QueryRow(ctx, q.query, q.args...).Scan(q.dest); err != nil {
				*q.dest = 0
			}
		}
	}

	// Revenue queries (amount stored in cents)
	var totalRevenueCents, todayRevenueCents int64
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed'`,
	).Scan(&totalRevenueCents); err != nil {
		totalRevenueCents = 0
	}
	if err := s.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed' AND created_at >= $1`,
		today,
	).Scan(&todayRevenueCents); err != nil {
		todayRevenueCents = 0
	}
	stats.TotalRevenue = float64(totalRevenueCents) / 100
	stats.RevenueToday = float64(todayRevenueCents) / 100

	return stats, nil
}

// GetRevenueChart returns daily revenue for the past N days
func (s *StatsService) GetRevenueChart(ctx context.Context, days int) ([]ChartPoint, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			DATE(created_at) AS date,
			COALESCE(SUM(total_amount), 0) / 100.0 AS revenue
		FROM orders
		WHERE status = 'completed'
		  AND created_at >= NOW() - ($1 || ' days')::INTERVAL
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

// GetPopularGames returns the top games by download count
func (s *StatsService) GetPopularGames(ctx context.Context, limit int) ([]PopularGame, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	rows, err := s.pool.Query(ctx, `
		SELECT g.id, g.title, g.slug, COUNT(dl.id) AS downloads
		FROM games g
		LEFT JOIN releases r ON r.branch_id IN (
			SELECT id FROM game_branches WHERE game_id = g.id
		)
		LEFT JOIN download_logs dl ON dl.release_id = r.id
		GROUP BY g.id, g.title, g.slug
		ORDER BY downloads DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return []PopularGame{}, nil
	}
	defer rows.Close()

	games := make([]PopularGame, 0)
	for rows.Next() {
		var g PopularGame
		if err := rows.Scan(&g.ID, &g.Title, &g.Slug, &g.Downloads); err != nil {
			continue
		}
		games = append(games, g)
	}
	return games, nil
}
