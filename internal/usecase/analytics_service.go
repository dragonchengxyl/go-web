package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/analytics"
	"github.com/studio/platform/internal/pkg/apperr"
)

// AnalyticsService handles analytics operations
type AnalyticsService struct {
	analyticsRepo analytics.Repository
	db            *pgxpool.Pool
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepo analytics.Repository, db *pgxpool.Pool) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		db:            db,
	}
}

// TrackEvent tracks an analytics event
func (s *AnalyticsService) TrackEvent(ctx context.Context, event *analytics.Event) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	return s.analyticsRepo.Track(event)
}

// GetDashboardMetrics retrieves dashboard metrics
func (s *AnalyticsService) GetDashboardMetrics(ctx context.Context, startDate, endDate time.Time) (*analytics.Metrics, error) {
	metrics := &analytics.Metrics{}

	// Get DAU
	dauQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'user_login'
		AND created_at >= $1 AND created_at <= $2
		AND DATE(created_at) = CURRENT_DATE
	`
	if err := s.db.QueryRow(ctx, dauQuery, startDate, endDate).Scan(&metrics.DAU); err != nil {
		return nil, fmt.Errorf("failed to get DAU: %w", err)
	}

	// Get MAU
	mauQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'user_login'
		AND created_at >= NOW() - INTERVAL '30 days'
	`
	if err := s.db.QueryRow(ctx, mauQuery).Scan(&metrics.MAU); err != nil {
		return nil, fmt.Errorf("failed to get MAU: %w", err)
	}

	// Get new users
	newUsersQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'user_register'
		AND created_at >= $1 AND created_at <= $2
	`
	if err := s.db.QueryRow(ctx, newUsersQuery, startDate, endDate).Scan(&metrics.NewUsers); err != nil {
		return nil, fmt.Errorf("failed to get new users: %w", err)
	}

	// Get total revenue
	revenueQuery := `
		SELECT COALESCE(SUM(total_cents), 0) / 100.0
		FROM orders
		WHERE status = 'paid'
		AND created_at >= $1 AND created_at <= $2
	`
	if err := s.db.QueryRow(ctx, revenueQuery, startDate, endDate).Scan(&metrics.TotalRevenue); err != nil {
		return nil, fmt.Errorf("failed to get revenue: %w", err)
	}

	// Get top games
	topGamesQuery := `
		SELECT
			g.id,
			g.title,
			COUNT(DISTINCT ae.session_id) FILTER (WHERE ae.event_type = 'game_view') as views,
			COUNT(*) FILTER (WHERE ae.event_type = 'game_download') as downloads,
			COALESCE(SUM(o.total_cents) FILTER (WHERE o.status = 'paid'), 0) / 100.0 as revenue
		FROM games g
		LEFT JOIN analytics_events ae ON ae.properties->>'game_id' = g.id::text
			AND ae.created_at >= $1 AND ae.created_at <= $2
		LEFT JOIN orders o ON o.user_id = ae.user_id
			AND o.created_at >= $1 AND o.created_at <= $2
		GROUP BY g.id, g.title
		ORDER BY views DESC
		LIMIT 10
	`
	rows, err := s.db.Query(ctx, topGamesQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get top games: %w", err)
	}
	defer rows.Close()

	metrics.TopGames = make([]analytics.GameMetric, 0)
	for rows.Next() {
		var gm analytics.GameMetric
		if err := rows.Scan(&gm.GameID, &gm.GameName, &gm.Views, &gm.Downloads, &gm.Revenue); err != nil {
			return nil, fmt.Errorf("failed to scan game metric: %w", err)
		}
		metrics.TopGames = append(metrics.TopGames, gm)
	}

	// Get user growth
	userGrowthQuery := `
		SELECT
			DATE(created_at) as date,
			COUNT(DISTINCT user_id) as value
		FROM analytics_events
		WHERE event_type = 'user_register'
		AND created_at >= $1 AND created_at <= $2
		GROUP BY DATE(created_at)
		ORDER BY date
	`
	rows, err = s.db.Query(ctx, userGrowthQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user growth: %w", err)
	}
	defer rows.Close()

	metrics.UserGrowth = make([]analytics.DailyMetric, 0)
	for rows.Next() {
		var dm analytics.DailyMetric
		if err := rows.Scan(&dm.Date, &dm.Value); err != nil {
			return nil, fmt.Errorf("failed to scan daily metric: %w", err)
		}
		metrics.UserGrowth = append(metrics.UserGrowth, dm)
	}

	return metrics, nil
}

// GetConversionFunnel retrieves conversion funnel data
func (s *AnalyticsService) GetConversionFunnel(ctx context.Context, startDate, endDate time.Time) (*analytics.ConversionFunnel, error) {
	funnel := &analytics.ConversionFunnel{
		Name:  "User Conversion",
		Steps: make([]analytics.FunnelStep, 0),
	}

	// Step 1: Page views
	step1Query := `
		SELECT COUNT(DISTINCT session_id)
		FROM analytics_events
		WHERE event_type = 'page_view'
		AND created_at >= $1 AND created_at <= $2
	`
	var step1Count int64
	if err := s.db.QueryRow(ctx, step1Query, startDate, endDate).Scan(&step1Count); err != nil {
		return nil, err
	}
	funnel.Total = step1Count

	// Step 2: Game views
	step2Query := `
		SELECT COUNT(DISTINCT session_id)
		FROM analytics_events
		WHERE event_type = 'game_view'
		AND created_at >= $1 AND created_at <= $2
	`
	var step2Count int64
	if err := s.db.QueryRow(ctx, step2Query, startDate, endDate).Scan(&step2Count); err != nil {
		return nil, err
	}

	// Step 3: Registrations
	step3Query := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'user_register'
		AND created_at >= $1 AND created_at <= $2
	`
	var step3Count int64
	if err := s.db.QueryRow(ctx, step3Query, startDate, endDate).Scan(&step3Count); err != nil {
		return nil, err
	}

	// Step 4: Purchases
	step4Query := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'purchase_complete'
		AND created_at >= $1 AND created_at <= $2
	`
	var step4Count int64
	if err := s.db.QueryRow(ctx, step4Query, startDate, endDate).Scan(&step4Count); err != nil {
		return nil, err
	}

	funnel.Steps = []analytics.FunnelStep{
		{Step: "Page View", Users: step1Count, Conversion: 100.0},
		{Step: "Game View", Users: step2Count, Conversion: float64(step2Count) / float64(step1Count) * 100},
		{Step: "Register", Users: step3Count, Conversion: float64(step3Count) / float64(step1Count) * 100},
		{Step: "Purchase", Users: step4Count, Conversion: float64(step4Count) / float64(step1Count) * 100},
	}

	return funnel, nil
}

// GetUserRetention retrieves user retention data
func (s *AnalyticsService) GetUserRetention(ctx context.Context, cohortDate time.Time, days int) (*analytics.CohortAnalysis, error) {
	cohort := &analytics.CohortAnalysis{
		Cohort:         cohortDate,
		RetentionRates: make(map[int]float64),
	}

	// Get initial users
	initialQuery := `
		SELECT COUNT(DISTINCT user_id)
		FROM analytics_events
		WHERE event_type = 'user_register'
		AND DATE(created_at) = $1
	`
	if err := s.db.QueryRow(ctx, initialQuery, cohortDate).Scan(&cohort.InitialUsers); err != nil {
		return nil, err
	}

	// Calculate retention for each day
	for day := 1; day <= days; day++ {
		retentionQuery := `
			SELECT COUNT(DISTINCT ae.user_id)
			FROM analytics_events ae
			WHERE ae.user_id IN (
				SELECT DISTINCT user_id
				FROM analytics_events
				WHERE event_type = 'user_register'
				AND DATE(created_at) = $1
			)
			AND ae.event_type = 'user_login'
			AND DATE(ae.created_at) = $2
		`
		targetDate := cohortDate.AddDate(0, 0, day)
		var retainedUsers int64
		if err := s.db.QueryRow(ctx, retentionQuery, cohortDate, targetDate).Scan(&retainedUsers); err != nil {
			return nil, err
		}

		if cohort.InitialUsers > 0 {
			cohort.RetentionRates[day] = float64(retainedUsers) / float64(cohort.InitialUsers) * 100
		}
	}

	return cohort, nil
}

// RefreshDailyMetrics refreshes the daily metrics materialized view
func (s *AnalyticsService) RefreshDailyMetrics(ctx context.Context) error {
	_, err := s.db.Exec(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY daily_metrics")
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "刷新每日指标失败", err)
	}
	return nil
}
