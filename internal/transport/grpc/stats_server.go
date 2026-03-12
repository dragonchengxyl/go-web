package grpc

import (
	"context"

	"github.com/studio/platform/internal/usecase"
	commonv1 "github.com/studio/platform/proto/common/v1"
	statsv1 "github.com/studio/platform/proto/stats/v1"
)

// StatsServer implements statsv1.StatsServiceServer using the local StatsService.
type StatsServer struct {
	statsv1.UnimplementedStatsServiceServer
	svc *usecase.StatsService
}

// NewStatsServer creates a new StatsServer.
func NewStatsServer(svc *usecase.StatsService) *StatsServer {
	return &StatsServer{svc: svc}
}

// GetDashboardStats returns dashboard metrics.
func (s *StatsServer) GetDashboardStats(ctx context.Context, _ *commonv1.Empty) (*statsv1.DashboardStats, error) {
	stats, err := s.svc.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &statsv1.DashboardStats{
		TotalUsers:    stats.TotalUsers,
		NewUsersToday: stats.NewUsersToday,
		TotalPosts:    stats.TotalPosts,
		TotalReports:  stats.TotalReports,
	}, nil
}

// StreamUserGrowth streams user growth data points.
func (s *StatsServer) StreamUserGrowth(req *statsv1.ChartRequest, stream statsv1.StatsService_StreamUserGrowthServer) error {
	points, err := s.svc.GetUserGrowthChart(stream.Context(), int(req.Days))
	if err != nil {
		return err
	}
	for _, p := range points {
		if err := stream.Send(&statsv1.ChartPoint{Date: p.Date, Value: p.Value}); err != nil {
			return err
		}
	}
	return nil
}
