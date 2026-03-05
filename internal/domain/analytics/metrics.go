package analytics

import (
	"time"

	"github.com/google/uuid"
)

// Metrics represents analytics metrics
type Metrics struct {
	DAU              int64              `json:"dau"`
	MAU              int64              `json:"mau"`
	NewUsers         int64              `json:"new_users"`
	ActiveUsers      int64              `json:"active_users"`
	TotalRevenue     float64            `json:"total_revenue"`
	AverageRevenue   float64            `json:"average_revenue"`
	ConversionRate   float64            `json:"conversion_rate"`
	RetentionRate    float64            `json:"retention_rate"`
	TopGames         []GameMetric       `json:"top_games"`
	TopProducts      []ProductMetric    `json:"top_products"`
	UserGrowth       []DailyMetric      `json:"user_growth"`
	RevenueGrowth    []DailyMetric      `json:"revenue_growth"`
}

// GameMetric represents game-specific metrics
type GameMetric struct {
	GameID       uuid.UUID `json:"game_id"`
	GameName     string    `json:"game_name"`
	Views        int64     `json:"views"`
	Downloads    int64     `json:"downloads"`
	Revenue      float64   `json:"revenue"`
	Rating       float64   `json:"rating"`
}

// ProductMetric represents product-specific metrics
type ProductMetric struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Sales       int64     `json:"sales"`
	Revenue     float64   `json:"revenue"`
}

// DailyMetric represents daily time-series metrics
type DailyMetric struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// FunnelStep represents a step in a conversion funnel
type FunnelStep struct {
	Step       string  `json:"step"`
	Users      int64   `json:"users"`
	Conversion float64 `json:"conversion"`
}

// ConversionFunnel represents a complete conversion funnel
type ConversionFunnel struct {
	Name  string        `json:"name"`
	Steps []FunnelStep  `json:"steps"`
	Total int64         `json:"total"`
}

// CohortAnalysis represents cohort retention analysis
type CohortAnalysis struct {
	Cohort         time.Time       `json:"cohort"`
	InitialUsers   int64           `json:"initial_users"`
	RetentionRates map[int]float64 `json:"retention_rates"` // day -> rate
}

// MetricsService defines the interface for analytics metrics
type MetricsService interface {
	// GetDashboardMetrics retrieves dashboard metrics
	GetDashboardMetrics(startDate, endDate time.Time) (*Metrics, error)

	// GetConversionFunnel retrieves conversion funnel data
	GetConversionFunnel(funnelName string, startDate, endDate time.Time) (*ConversionFunnel, error)

	// GetCohortAnalysis retrieves cohort retention analysis
	GetCohortAnalysis(startDate time.Time, days int) ([]*CohortAnalysis, error)

	// GetUserGrowth retrieves user growth over time
	GetUserGrowth(startDate, endDate time.Time) ([]DailyMetric, error)

	// GetRevenueGrowth retrieves revenue growth over time
	GetRevenueGrowth(startDate, endDate time.Time) ([]DailyMetric, error)
}
