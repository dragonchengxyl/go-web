package grpc

import (
	"context"

	commonv1 "github.com/studio/platform/proto/common/v1"
	statsv1 "github.com/studio/platform/proto/stats/v1"
	"github.com/studio/platform/internal/usecase"
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
func (s *StatsServer) GetDashboardStats(ctx context.Context, _ *commonv1.Empty) (*statsv1.DashboardStatsProto, error) {
	stats, err := s.svc.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}
	return &statsv1.DashboardStatsProto{
		TotalUsers:     stats.TotalUsers,
		NewUsersToday:  stats.NewUsersToday,
		TotalDownloads: stats.TotalDownloads,
		DownloadsToday: stats.DownloadsToday,
		TotalRevenue:   stats.TotalRevenue,
		RevenueToday:   stats.RevenueToday,
		TotalGames:     stats.TotalGames,
		TotalOrders:    stats.TotalOrders,
		TotalPosts:     stats.TotalPosts,
		TotalReports:   stats.TotalReports,
	}, nil
}

// GetRevenueChart returns revenue chart data.
func (s *StatsServer) GetRevenueChart(ctx context.Context, req *statsv1.ChartRequest) (*statsv1.ChartResponse, error) {
	points, err := s.svc.GetRevenueChart(ctx, int(req.Days))
	if err != nil {
		return nil, err
	}
	resp := &statsv1.ChartResponse{Points: make([]*statsv1.ChartPointProto, len(points))}
	for i, p := range points {
		resp.Points[i] = &statsv1.ChartPointProto{Date: p.Date, Value: p.Value}
	}
	return resp, nil
}

// StreamUserGrowth streams user growth data points.
func (s *StatsServer) StreamUserGrowth(req *statsv1.ChartRequest, stream statsv1.StatsService_StreamUserGrowthServer) error {
	points, err := s.svc.GetUserGrowthChart(stream.Context(), int(req.Days))
	if err != nil {
		return err
	}
	for _, p := range points {
		if err := stream.Send(&statsv1.ChartPointProto{Date: p.Date, Value: p.Value}); err != nil {
			return err
		}
	}
	return nil
}
