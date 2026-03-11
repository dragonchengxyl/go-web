package usecase

import "context"

// StatsProvider is the interface for stats operations.
// *StatsService satisfies this interface (local mode).
// *StatsGRPCClient also satisfies this interface (gRPC mode).
type StatsProvider interface {
	GetDashboardStats(ctx context.Context) (*DashboardStats, error)
	GetRevenueChart(ctx context.Context, days int) ([]ChartPoint, error)
	GetUserGrowthChart(ctx context.Context, days int) ([]ChartPoint, error)
}
